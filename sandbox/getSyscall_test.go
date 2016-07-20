package sandbox

import (
	"os"
	"syscall"
	"golang.org/x/sys/unix"
	"testing"
	"runtime"
)

// Test ptraceGetSyscall
func Test_ptraceGetSyscall(t *testing.T) {
	// Disable Go runtime preempt, or I may be switched
	// to another thread, who is not the tracer. Then we'll
	// receive ESRCH (no such tracee).
	runtime.LockOSThread()

	// Create a process to trace.
	// This is a small C program, call nanosleep(1000000) 100 times.
	sysAttr := &syscall.SysProcAttr{Ptrace:true}
	attr := &os.ProcAttr{Sys:sysAttr}
	process, err := os.StartProcess("./nanosleep100.exe",
	  []string{"nanosleep100.exe"}, attr)

	if err != nil {
		t.Fatal("Cannot start process to trace. Error:", err.Error());
	}

	pid := process.Pid
	cnt := 0

	for {
		var wstat unix.WaitStatus
		_, err := unix.Wait4(pid, &wstat, 0, nil)
		if err != nil {
			t.Fatal("Wait4 failed. Error:", err.Error());
		}

		if wstat.Stopped() && wstat.StopSignal() == unix.SIGTRAP {
			// Try to get syscall ID
			syscall, err := ptraceGetSyscall(pid)
			if err != nil {
				t.Error("ptraceGetSyscall failed. Error:", err.Error());
			} else {
				t.Logf("Get syscall with ID %d\n", syscall)
			}
			if syscall == unix.SYS_NANOSLEEP {
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

	// enable preempt
	runtime.UnlockOSThread()

	// check the result
	if cnt != 200 {
		t.Errorf("We have only traced %d syscall enter and leave, " +
		  "expect 200.", cnt);
	}
}
