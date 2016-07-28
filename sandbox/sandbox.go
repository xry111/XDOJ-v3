// sandbox.go - Package sandbox.
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

// Package sandbox runs external commands in a limited environment.
// Most of limits have been implemented by SysProcAttr (thanks to the
// Go Authors), but we have to handle resources limits.
//
// Since forking a multithread process is unsafe, Go runtime would
// handle all things from fork to exec. So the resources limits are
// applied by syscall prlimit instead of setrlimit. We need
//   Linux Kernel Version > 2.6.36
package sandbox

import (
	"os"
)

// ResLimits contains resource limits.
type ResLimits struct {
	TimeLimit int // CPU time limit in milliseconds

	// The defination of "memory usage" and "memory limit"
	// is a little bit ambiguous.  Here we define "memory
	// usage" as "(minor page fault time) * (page size)".
	// The reasons are:
	//    Virtual memory size is hard to measure;
	//
	//    Some program such as JVM would allocate a lot of
	//    virtual memory, so VM size is much more than physical
	//    memory (including swap) usage;
	//
	//    Major page fault (with IO) which happens in ELF
	//    loading and swapping should not be considered.
	//
	// This is still a little buggy, for example, initialized
	// data segment is not considered, and someone may mmap
	// stdout to get some memory.  FIXME.
	MemoryLimit int // Memory usage limit in KB

	OutputLimit int // File size limit in KB
}

// Sandbox contains external command and all limits.
type Sandbox struct {
	Attr    *os.ProcAttr // Dir, Env, Files, SysProcAttr
	Path    string       // Path to executable file
	Args    []string     // Arguments
	RLimits *ResLimits   // Resource limits

	// DisableSyscall[x] = true means we permit Command to call
	// syscall with ID x.
	DisableSyscall []bool
}

// RunResult contains the results of a Sandbox run.
type RunResult struct {
	// If the command terminated normally (by call SYS_exit),
	// status contains the return value.
	// Otherwise, the command is terminated by signal x, and
	// status contains -x.
	Status int

	// If the command has called an illegal syscall
	IllegalSyscall int

	TimeCost int // CPU time cost in ms (rusage.ru_utime)

	// See note in ResLimits.MemoryLimit
	MemoryCost int // Maximum memory usage in KB
}

// Run a external command and wait until it terminate.
func (c *Sandbox) Run() (ret *RunResult, err error) {
	return c.run()
}
