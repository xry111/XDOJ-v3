// trace.go - Trace a Sandbox run
// Copyright (C) 2016 Laboratory of ACM/ICPC, Xidian University

// This is part of XDOJ-v3.

// XDOJ-v3 is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of
// the License, or (at your option) any later version.

// XDOJ-v3 is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without eventhe implied warranty of MER-
// CHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see
// <http://www.gnu.org/licenses>.

// Author: Xi Ruoyao <xry111@outlook.com>

package sandbox

import (
	"errors"
	"github.com/xry111/XDOJ-v3/ojerror"
	"golang.org/x/sys/unix"
	"os"
	"runtime"
	"syscall"
)

var pageSizeKB int64

func init() {
	pageSizeKB = int64(os.Getpagesize() >> 10)
}

// Kill a process, but ignore errno 3 (ESRCH)
// since others (OOM for eg.) may kill our tracee.
func kill(pid int) (err error) {
	err1 := unix.Kill(pid, unix.SIGKILL)
	if err1 == syscall.Errno(3) {
		return nil
	}
	return err1
}

// Convert unix.Timeval to microseconds.
func (tv *unix.Timeval) micro() (usec int64) {
	return int64(tv.Usec) + int64(tv.Sec)*1000000
}

func (c *Sandbox) run() (ret *RunResult, err error) {
	// Always trace the process.
	if c.Attr.Sys == nil {
		c.Attr.Sys = &syscall.SysProcAttr{}
	}
	c.Attr.Sys.Ptrace = true

	// Go runtime may preempt our go routine, and then schedule
	// it on another Linux thread, which is not the tracer. Then
	// we can not trace the tracee any more. We disable preempt
	// here, and re-enable preempt at returning.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Start the process
	var process *os.Process
	process, err = os.StartProcess(c.Path, c.Args, c.Attr)
	if err != nil {
		err = ojerror.New(err)
		return
	}
	pid := process.Pid

	// We may exit this function when an error occurs.
	// Ensure no zombies wandering on the XDOJ server...
	alive := true
	// In fact this varible should be called "alive or zombie",
	// so we should use the last wait4 on the tracee to bury it.

	// We must be careful, if we wait a zombie twice, the thread
	// will block forever.
	defer func() {
		if alive {
			kill(pid)
			unix.Wait4(pid, nil, 0, nil)
		}
	}()

	// Now trace the process
	var rusage unix.Rusage
	var wstatus unix.WaitStatus
	var call int

	// First SIGTRAP, leaving sys_execve
	_, err = unix.Wait4(pid, &wstatus, 0, &rusage)
	if wstatus.Exited() {
		alive = false
		err = ojerror.New(errors.New(
			"Child exited early. Seems a Go Runtime BUG."))
		return
	} else if wstatus.Signaled() {
		alive = false
		err = ojerror.New(errors.New(
			"Child signaled early. Maybe an OOM."))
		return
	} else if wstatus.StopSignal() != unix.SIGTRAP {
		err = ojerror.New(errors.New(
			"Child stopped by an alien signal early."))
		return
	} else {
		call, err = ptraceGetSyscall(pid)
		if err != nil {
			err = ojerror.New(err)
			return
		}
		if call != unix.SYS_EXECVE {
			err = ojerror.New(errors.New(
				"Unexpected system call from child."))
			return
		}
	}

	// We can set rlimits here
	var rlimit unix.Rlimit

	// CPU time limit
	cpuLimitSec := uint64(c.RLimits.TimeLimit+999) / 1000
	rlimit.Cur = cpuLimitSec
	rlimit.Max = cpuLimitSec
	err = prlimit64(pid, unix.RLIMIT_CPU, &rlimit, nil)
	if err != nil {
		err = ojerror.New(err)
		return
	}

	// VM limit, 1GB should be enough
	// Prevent someone bypass the page fault limit and use all
	// the memory
	vmLimitByte := uint64(1 << 30)
	rlimit.Cur = vmLimitByte
	rlimit.Max = vmLimitByte
	err = prlimit64(pid, unix.RLIMIT_AS, &rlimit, nil)
	if err != nil {
		err = ojerror.New(err)
		return
	}

	// File size limit
	rlimit.Cur = uint64(c.RLimits.OutputLimit) << 10
	rlimit.Max = uint64(c.RLimits.OutputLimit) << 10
	err = prlimit64(pid, unix.RLIMIT_FSIZE, &rlimit, nil)
	if err != nil {
		err = ojerror.New(err)
		return
	}

	// Zero point usage
	time0 := rusage.Utime.micro() + rusage.Stime.micro()
	pf0 := rusage.Minflt

	// Let it go
	err = unix.PtraceSyscall(pid, 0)
	if err != nil {
		err = ojerror.New(err)
		return
	}

	result := RunResult{}

	// Loop...
	for ret == nil {
		_, err = unix.Wait4(pid, &wstatus, 0, &rusage)
		if err != nil {
			err = ojerror.New(err)
			return
		}

		result.TimeCost = int((rusage.Utime.micro() +
			rusage.Stime.micro() - time0 + 500) / 1000)
		result.MemoryCost = int((rusage.Minflt - pf0) * pageSizeKB)

		if wstatus.Exited() {
			alive = false
			result.Status = wstatus.ExitStatus()
			ret = &result
			return
		} else if wstatus.Signaled() {
			alive = false
			if result.Status == 0 {
				result.Status = -int(wstatus.Signal())
			}
			ret = &result
			return
		} else { // Stopped
			if result.TimeCost > c.RLimits.TimeLimit ||
				result.MemoryCost > c.RLimits.MemoryLimit {
				kill(pid)
				continue
			}
			signal := wstatus.StopSignal()
			if signal == unix.SIGTRAP {
				call, err1 := ptraceGetSyscall(pid)

				// Tracee may be killed by others now,
				// then we'll get ESRCH (errno 3). Simply
				// go to next loop and let wait4 clean
				// the mess.
				if err1 == syscall.Errno(3) {
					continue
				}

				// Other errnos is out of tolerance. GG.
				if err1 != nil {
					err = ojerror.New(err1)
					return
				}

				if c.DisableSyscall != nil && c.DisableSyscall[call] {
					result.IllegalSyscall = call
					err = kill(pid)
					if err != nil {
						err = ojerror.New(err)
						return
					}
				}

			} else {
				// other lethal signals
				result.Status = -int(signal)
				err = kill(pid)
				if err != nil {
					err = ojerror.New(err)
					return
				}
			}

			err1 := unix.PtraceSyscall(pid, 0)
			if err1 == syscall.Errno(3) { // Ignore ESRCH
				continue
			}

			if err1 != nil {
				err = ojerror.New(err1)
				return
			}
		}
	}

	return
}
