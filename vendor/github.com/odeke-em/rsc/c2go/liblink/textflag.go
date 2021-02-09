package liblink

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

const (
	fmtLong = 1 << iota
)

// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// This file defines flags attached to various functions
// and data objects.  The compilers, assemblers, and linker must
// all agree on these values.
// Don't profile the marked routine.  This flag is deprecated.
// It is ok for the linker to get multiple of these symbols.  It will
// pick one of the duplicates to use.
// Don't insert stack check preamble.
// Put this data in a read-only section.
// This data contains no pointers.
// This is a wrapper function and should not count as disabling 'recover'.
// This function uses its incoming context register.
const (
	NOPROF   = 1
	DUPOK    = 2
	NOSPLIT  = 4
	RODATA   = 8
	NOPTR    = 16
	WRAPPER  = 32
	NEEDCTXT = 64
)
