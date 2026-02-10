package parsing

import "unicode/utf8"

// ContainsByte interprets s as a unicode string,
// and returns whether the single-byte character appears anywhere in s.
func ContainsByte(s []byte, b byte) bool {
	for len(s) > 0 {
		_, l := utf8.DecodeRune(s)
		// since len(s) > 0, l will never be 0, but it may be greater than 1.
		if l == 1 && s[0] == b {
			return true
		}
		s = s[l:]
	}
	return false
}
