package sandbox

import (
	"golang.org/x/sys/unix"
	"os"
	"runtime"
	"syscall"
	"testing"
)

// Test ptraceGetSyscall
func do_Test_ptraceGetSyscall(t *testing.T, exe string, call int) {
	// Disable Go runtime preempt, or I may be switched
	// to another thread, who is not the tracer. Then we'll
	// receive ESRCH (no such tracee).
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Create a process to trace.
	// This is a small C program, call nanosleep(1000000) 100 times.
	sysAttr := &syscall.SysProcAttr{Ptrace: true}
	attr := &os.ProcAttr{Sys: sysAttr}
	process, err := os.StartProcess("./tracee/"+exe,
		[]string{exe}, attr)

	if err != nil {
		t.Fatal("Cannot start process to trace. Error:", err.Error())
	}

	pid := process.Pid
	cnt := 0

	for {
		var wstat unix.WaitStatus
		_, err := unix.Wait4(pid, &wstat, 0, nil)
		if err != nil {
			t.Fatal("Wait4 failed. Error:", err.Error())
		}

		if wstat.Stopped() && wstat.StopSignal() == unix.SIGTRAP {
			// Try to get syscall ID
			syscall, err := ptraceGetSyscall(pid)
			if err != nil {
				t.Error("ptraceGetSyscall failed. Error:", err.Error())
			} else {
				t.Logf("Get syscall with ID %d\n", syscall)
			}
			if syscall == call {
				cnt++
			}
		}

		if wstat.Signaled() {
			t.Log("Tracee killed by", wstat.Signal().String())
			break
		}
		if wstat.Exited() {
			t.Logf("Tracee exited with status %d", wstat.ExitStatus())
			break
		}

		err = unix.PtraceSyscall(pid, 0)
		if err != nil {
			t.Fatalf("PtraceSyscall failed. Error:", err)
		}
	}

	// check the result
	if cnt != 200 {
		t.Errorf("We have only traced %d syscall enter and leave, "+
			"expect 200.", cnt)
	}
}

func Test_ptraceGetSyscall64(t *testing.T) {
	do_Test_ptraceGetSyscall(t, "nsleep100.exe", 35)
}

func Test_ptraceGetSyscall32(t *testing.T) {
	do_Test_ptraceGetSyscall(t, "nsleep100_32.exe", 162)
}
