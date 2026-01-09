package test

import (
	"regexp"
	"strings"
	"testing"
)

type WantMode int

const (
	WantEqual WantMode = iota
	WantPrefix
	WantContains
	WantRegexp
)

func CheckString(t *testing.T, name, want, got string, mode WantMode) {
	t.Helper()
	switch mode {
	case WantPrefix:
		if !strings.HasPrefix(got, want) {
			t.Errorf("%s did not match prefix - want:\n%s\ngot:\n%s",
				name, want, got)
		}

	case WantContains:
		if !strings.Contains(got, want) {
			t.Errorf("%s did not contain string - want:\n%s\ngot:\n%s",
				name, want, got)
		}

	case WantRegexp:
		r := regexp.MustCompile(want)
		if !r.MatchString(got) {
			t.Errorf("%s did not match regexp - want:\n%s\ngot:\n%s",
				name, want, got)
		}

	default:
		if want != got {
			t.Errorf("%s did not match - want:\n%s\ngot:\n%s",
				name, want, got)
		}
	}
}
