package shell

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
