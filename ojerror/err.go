// err.go - Package XDOJ-v3/ojerror.
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

// Package XDOJ-v3/err provides struct Err, which implements
// builtin.error. It tell us the position (function, file and line)
// where the error occurs, so we can identify bugs easily.
package ojerror

import (
	"strings"
	"fmt"
	"runtime"
)

type Err struct {
	ActualErr error
	Pos string
}

func (err *Err) Error() (msg string) {
	msg = "At " + err.Pos + " :\n" + err.ActualErr.Error()
	return
}

func New(err1 error) (err error) {
	if err1 == nil {
		return
	}
	pc, fn, ln, ok := runtime.Caller(1)
	pos := "?????"

	if ok {
		funcObj := runtime.FuncForPC(pc)
		funcName := "???"
		path := strings.Split(fn, "/")
		fn = path[len(path)-1]
		if funcObj != nil {
			funcName = funcObj.Name()
		}
		pos = fmt.Sprintf("%x (%s)\n in %s:%d", pc, funcName, fn, ln)
	}

	err = &Err {
		ActualErr: err1,
		Pos: pos,
	}
	return
}
