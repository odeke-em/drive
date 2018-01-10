package liblink

/*
 * The authors of this software are Rob Pike and Ken Thompson.
 *              Copyright (c) 2002 by Lucent Technologies.
 * Permission to use, copy, modify, and distribute this software for any
 * purpose without fee is hereby granted, provided that this entire notice
 * is included in all copies of any software which is or includes a copy
 * or modification of this software and in all copies of the supporting
 * documentation for such software.
 * THIS SOFTWARE IS BEING PROVIDED "AS IS", WITHOUT ANY EXPRESS OR IMPLIED
 * WARRANTY.  IN PARTICULAR, NEITHER THE AUTHORS NOR LUCENT TECHNOLOGIES MAKE ANY
 * REPRESENTATION OR WARRANTY OF ANY KIND CONCERNING THE MERCHANTABILITY
 * OF THIS SOFTWARE OR ITS FITNESS FOR ANY PARTICULAR PURPOSE.
 */
/*
 * The authors of this software are Rob Pike and Ken Thompson.
 *              Copyright (c) 1998-2002 by Lucent Technologies.
 *              Portions Copyright (c) 2009 The Go Authors.  All rights reserved.
 * Permission to use, copy, modify, and distribute this software for any
 * purpose without fee is hereby granted, provided that this entire notice
 * is included in all copies of any software which is or includes a copy
 * or modification of this software and in all copies of the supporting
 * documentation for such software.
 * THIS SOFTWARE IS BEING PROVIDED "AS IS", WITHOUT ANY EXPRESS OR IMPLIED
 * WARRANTY.  IN PARTICULAR, NEITHER THE AUTHORS NOR LUCENT TECHNOLOGIES MAKE ANY
 * REPRESENTATION OR WARRANTY OF ANY KIND CONCERNING THE MERCHANTABILITY
 * OF THIS SOFTWARE OR ITS FITNESS FOR ANY PARTICULAR PURPOSE.
 */
type Rune uint /* Code-point values in Unicode 4.0 are 21 bits wide.*/

const (
	UTFmax    = 4
	Runesync  = 0x80
	Runeself  = 0x80
	Runeerror = 0xFFFD
	Runemax   = 0x10FFFF
)

/*
 * rune routines
 */
/*
 * These routines were written by Rob Pike and Ken Thompson
 * and first appeared in Plan 9.
 * SEE ALSO
 * utf (7)
 * tcs (1)
 */
// runetochar copies (encodes) one rune, pointed to by r, to at most
// UTFmax bytes starting at s and returns the number of bytes generated.

// chartorune copies (decodes) at most UTFmax bytes starting at s to
// one rune, pointed to by r, and returns the number of bytes consumed.
// If the input is not exactly in UTF format, chartorune will set *r
// to Runeerror and return 1.
//
// Note: There is no special case for a "null-terminated" string. A
// string whose first byte has the value 0 is the UTF8 encoding of the
// Unicode value 0 (i.e., ASCII NULL). A byte value of 0 is illegal
// anywhere else in a UTF sequence.

// charntorune is like chartorune, except that it will access at most
// n bytes of s.  If the UTF sequence is incomplete within n bytes,
// charntorune will set *r to Runeerror and return 0. If it is complete
// but not in UTF format, it will set *r to Runeerror and return 1.
//
// Added 2004-09-24 by Wei-Hwa Huang

// isvalidcharntorune(str, n, r, consumed)
// is a convenience function that calls "*consumed = charntorune(r, str, n)"
// and returns an int (logically boolean) indicating whether the first
// n bytes of str was a valid and complete UTF sequence.

// runelen returns the number of bytes required to convert r into UTF.

// runenlen returns the number of bytes required to convert the n
// runes pointed to by r into UTF.

// fullrune returns 1 if the string s of length n is long enough to be
// decoded by chartorune, and 0 otherwise. This does not guarantee
// that the string contains a legal UTF encoding. This routine is used
// by programs that obtain input one byte at a time and need to know
// when a full rune has arrived.

// The following routines are analogous to the corresponding string
// routines with "utf" substituted for "str", and "rune" substituted
// for "chr".
// utflen returns the number of runes that are represented by the UTF
// string s. (cf. strlen)

// utfnlen returns the number of complete runes that are represented
// by the first n bytes of the UTF string s. If the last few bytes of
// the string contain an incompletely coded rune, utfnlen will not
// count them; in this way, it differs from utflen, which includes
// every byte of the string. (cf. strnlen)

// utfrune returns a pointer to the first occurrence of rune r in the
// UTF string s, or 0 if r does not occur in the string.  The NULL
// byte terminating a string is considered to be part of the string s.
// (cf. strchr)
/*const*/

// utfrrune returns a pointer to the last occurrence of rune r in the
// UTF string s, or 0 if r does not occur in the string.  The NULL
// byte terminating a string is considered to be part of the string s.
// (cf. strrchr)
/*const*/

// utfutf returns a pointer to the first occurrence of the UTF string
// s2 as a UTF substring of s1, or 0 if there is none. If s2 is the
// null string, utfutf returns s1. (cf. strstr)

// utfecpy copies UTF sequences until a null sequence has been copied,
// but writes no sequences beyond es1.  If any sequences are copied,
// s1 is terminated by a null sequence, and a pointer to that sequence
// is returned.  Otherwise, the original s1 is returned. (cf. strecpy)

// These functions are rune-string analogues of the corresponding
// functions in strcat (3).
//
// These routines first appeared in Plan 9.
// SEE ALSO
// memmove (3)
// rune (3)
// strcat (2)
//
// BUGS: The outcome of overlapping moves varies among implementations.

// The following routines test types and modify cases for Unicode
// characters.  Unicode defines some characters as letters and
// specifies three cases: upper, lower, and title.  Mappings among the
// cases are also defined, although they are not exhaustive: some
// upper case letters have no lower case mapping, and so on.  Unicode
// also defines several character properties, a subset of which are
// checked by these routines.  These routines are based on Unicode
// version 3.0.0.
//
// NOTE: The routines are implemented in C, so the boolean functions
// (e.g., isupperrune) return 0 for false and 1 for true.
//
//
// toupperrune, tolowerrune, and totitlerune are the Unicode case
// mappings. These routines return the character unchanged if it has
// no defined mapping.

// isupperrune tests for upper case characters, including Unicode
// upper case letters and targets of the toupper mapping. islowerrune
// and istitlerune are defined analogously.

// isalpharune tests for Unicode letters; this includes ideographs in
// addition to alphabetic characters.

// isdigitrune tests for digits. Non-digit numbers, such as Roman
// numerals, are not included.

// isspacerune tests for whitespace characters, including "C" locale
// whitespace, Unicode defined whitespace, and the "zero-width
// non-break space" character.
