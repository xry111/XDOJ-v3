// getSyscall.go - Wrap sys_ptrace to get syscall ID
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

// Wrap sys_ptrace to get syscall ID.
func ptraceGetSyscall(pid int) (ID int, err error) {
	// value from sys/reg.h, sys/ptrace.h
	const PTRACE_PEEKUSER = 3
	var id uintptr

	// In kernel the API of PTRACE_PEEKUSER is strange.
	// We have to wrap it like Glibc.  See ptrace(2).
	r1, _, e1 := unix.RawSyscall6(unix.SYS_PTRACE,
		PTRACE_PEEKUSER, uintptr(pid),
		_REG_SYSCALL_ID*unsafe.Sizeof(id),
		uintptr(unsafe.Pointer(&id)), 0, 0)

	if r1 == ^uintptr(0) {
		err = syscall.Errno(e1)
	}
	return int(id), err
}
