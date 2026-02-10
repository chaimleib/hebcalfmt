package markdown

func TrimRepeating(s []byte) (rest []byte, trimLen int) {
	if len(s) == 0 {
		return nil, 0
	}

	c := s[0]
	rest = s[1:]
	for len(rest) > 0 {
		if c != rest[0] {
			break
		}
		rest = rest[1:]
	}

	if len(s)-len(rest) == 1 {
		return s, 0
	}

	return rest, len(s) - len(rest)
}

func TrimSpace(s []byte) []byte {
	if len(s) == 0 {
		return nil
	}

	for len(s) > 0 {
		if c := s[0]; c != ' ' && c != '\t' {
			break
		}
		s = s[1:]
	}
	return s
}
