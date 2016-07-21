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
	"syscall"
	"unsafe"
	"golang.org/x/sys/unix"
	"sync"
)

var syscallIdAddr uintptr
var once sync.Once

// Wrap sys_ptrace to get syscall ID.
func ptraceGetSyscall(pid int) (ID int, err error){
	var id uintptr

	// We'll peek tracee's _kernel_ data structure.
	// Get kernel information
	getkerninfo := func() {
		buf := &unix.Utsname{}
		err := unix.Uname(buf)
		if err != nil {
			// I can't do anything meaningful when such a
			// simple syscall has failed.
			panic("Cannot determine kernel architechture.")
		}

		// seems slow, but it's only executed on init.
		// nothing serious.
		arch := []byte{}
		for i := 0; i < len(buf.Machine); i++ {
			if buf.Machine[i] == 0 {
				break
			}
			arch = append(arch, byte(buf.Machine[i]))
		}

		switch string(arch) {
			case "x86_64" :
				syscallIdAddr = 15 * 8 // ORIG_RAX
			case "i386", "i486", "i586", "i686" :
				syscallIdAddr = 11 * 4 // ORIG_EAX
			default:
				syscallIdAddr = 0
		}

		if syscallIdAddr == 0 {
			// I am too young, too simple to know this
			// architechture. I should increase my knowledge
			// level.
			panic("What is this machine?")
		}
	}

	once.Do(getkerninfo)

	// In kernel the API of PTRACE_PEEKUSER is strange.
	// We have to wrap it like Glibc.  See ptrace(2).
	r1, _, e1 := unix.RawSyscall6(unix.SYS_PTRACE,
	  unix.PTRACE_PEEKUSR, uintptr(pid),
	  syscallIdAddr, uintptr(unsafe.Pointer(&id)), 0, 0)

	if r1 < 0 {
		err = syscall.Errno(e1)
	}

	// When kernel and tracer are both 64-bit, but tracee
	// is 32-bit, we have to throw additional high bits in rax.
	return int(id & 0xff), err
}
