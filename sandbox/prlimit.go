// prlimit.go - Wrap sys_prlimit64 calling
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
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

// Wrap sys_prlimit64.
//
// In Go, unix.Rlimit always contains 64-bit values. So we don't
// need to process 32-bit rlimits like Glibc.
func prlimit64(pid int, resource int, newLimit *unix.Rlimit,
	oldLimit *unix.Rlimit) (err error) {
	_, _, e1 := unix.RawSyscall6(unix.SYS_PRLIMIT64,
		uintptr(pid), uintptr(resource), uintptr(unsafe.Pointer(newLimit)),
		uintptr(unsafe.Pointer(oldLimit)), 0, 0)
	if e1 != 0 {
		err = syscall.Errno(e1)
	}
	return
}
