package parsing

import "unicode/utf8"

// ASCIISet enables faster checks on byte slices
// against a set of ASCII characters.
// It is a 32-byte value, where each bit represents the presence of a
// given ASCII character in the set. The 128-bits of the lower 16 bytes,
// starting with the least-significant bit of the lowest word to the
// most-significant bit of the highest word, map to the full range of all
// 128 ASCII characters. The 128-bits of the upper 16 bytes will be zeroed,
// ensuring that any non-ASCII character will be reported as not in the set.
// This allocates a total of 32 bytes even though the upper half
// is unused to avoid bounds checks.
//
// The implementation is mostly copied from the internals of the bytes package.
type ASCIISet [8]uint32

// MakeASCIISet creates a set of ASCII characters and reports whether all
// characters in s are ASCII.
func MakeASCIISet(s []byte) (as ASCIISet, ok bool) {
	for len(s) > 0 {
		c, l := utf8.DecodeRune(s)
		if c >= utf8.RuneSelf {
			return as, false
		}
		as[c/32] |= 1 << (c % 32)
		s = s[l:]
	}
	return as, true
}

// ContainsRune returns whether r is ASCII and is in the set.
func (as ASCIISet) ContainsRune(r rune) bool {
	if r >= utf8.RuneSelf {
		return false
	}
	return as.ContainsByte(byte(r))
}

// ContainsByte returns whether b is in the set.
func (as ASCIISet) ContainsByte(b byte) bool {
	return (as[b/32] & (1 << (b % 32))) != 0
}

// TrimLeft treats s as a string and removes ASCII characters in the set,
// repeatedly, from the beginning of s.
func (as ASCIISet) TrimLeft(s []byte) []byte {
	for len(s) > 0 {
		c := s[0]
		if c >= utf8.RuneSelf {
			break
		}
		if !as.ContainsByte(c) {
			break
		}
		s = s[1:]
	}
	if len(s) == 0 {
		return nil
	}
	return s
}
