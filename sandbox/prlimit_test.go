package sandbox

import (
	"golang.org/x/sys/unix"
	"testing"
)

// Test prlimit64
func Test_prlimit64(t *testing.T) {
	rlimit1 := &unix.Rlimit{Cur:131072, Max:1048576}
	myself := unix.Getpid()
	err := prlimit64(myself, unix.RLIMIT_FSIZE, rlimit1, nil)
	if err != nil {
		t.Errorf("Test #1 failed. Error: %s", err.Error())
	}

	// Use unix.Getrlimit to check result
	rlimit1Copy1 := &unix.Rlimit{}
	err = unix.Getrlimit(unix.RLIMIT_FSIZE, rlimit1Copy1);
	if err != nil {
		t.Fatalf("Getrlimit failed: %s", err.Error());
	}
	if (*rlimit1 != *rlimit1Copy1) {
		t.Errorf("Test #1 failed. NewLimit has not been set.")
	}

	rlimit2 := &unix.Rlimit{Cur:262144, Max:524288}
	rlimit1Copy2 := &unix.Rlimit{}
	err = prlimit64(myself, unix.RLIMIT_FSIZE, rlimit2, rlimit1Copy2)
	if err != nil {
		t.Errorf("Test #2 failed. Error: %s", err.Error());
	} else if *rlimit1Copy2 != *rlimit1 {
		t.Errorf("Test #2 failed. OldLimit contains wrong value.");
	}

	uid := unix.Getuid()
	if uid != 0 {
		rlimit3 := &unix.Rlimit{Cur:262144, Max:1048576}
		err = prlimit64(myself, unix.RLIMIT_FSIZE, rlimit3, nil)
		if err == nil {
			t.Errorf("Test #3 failed. Error expected.")
		} else {
			t.Logf("Test #3 produced error: %s", err.Error());
		}
	} else {
		t.Logf("Test #3 must be run without root permittion.")
	}

	// Test with a strange argument
	err = prlimit64(myself, 233333, rlimit1, nil)
	if err == nil {
		t.Errorf("Test #4 failed. Error expected.")
	} else {
		t.Logf("Test #4 produced error: %s", err.Error());
	}
}
