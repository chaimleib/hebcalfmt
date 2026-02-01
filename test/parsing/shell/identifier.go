package shell

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/chaimleib/hebcalfmt/test/parsing"
)

func IsAlphaOrUnderscore(c rune) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		c == '_'
}

func IsDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

func ParseIdentifier(
	line parsing.LineInfo,
	rest []byte,
) (identifier string, newRest []byte, err error) {
	orig := rest
	errOut := ErrOut[string](line, &rest)

	if len(rest) == 0 {
		return "", orig, ErrNoMatch
	}

	r, size := utf8.DecodeRune(rest)
	if r == utf8.RuneError {
		return errOut(size, "invalid unicode")
	}

	if !IsAlphaOrUnderscore(r) {
		return "", orig, ErrNoMatch
	}
	rest = rest[size:]

	for len(rest) > 0 {
		r, size = utf8.DecodeRune(rest)
		if r == utf8.RuneError {
			return errOut(size, "invalid unicode")
		}

		if !IsAlphaOrUnderscore(r) && !IsDigit(r) {
			break
		}
		rest = rest[size:]
	}

	identifier = string(orig[:len(orig)-len(rest)])
	return identifier, rest, nil
}

func ParseAssignment(
	line parsing.LineInfo,
	rest []byte,
) (key, value string, newRest []byte, err error) {
	orig := rest
	noMatchOut := func() (string, string, []byte, error) {
		return "", "", orig, ErrNoMatch
	}

	key, rest, err = ParseIdentifier(line, rest)
	if err != nil {
		return "", "", orig, err
	}

	rest, ok := bytes.CutPrefix(rest, []byte("="))
	if !ok {
		return noMatchOut()
	}

	value, rest, err = ParseShellString(line, rest)
	if errors.Is(err, ErrNoMatch) {
		return key, "", rest, nil
	}
	if err != nil {
		return "", "", orig, err
	}

	return key, value, rest, nil
}

func FormatAssignment(key, value string) string {
	return fmt.Sprintf("%s=%s", key, FormatString(value))
}

type Vars map[string]string

func (v Vars) PairStrings() []string {
	keys := slices.Sorted(maps.Keys(v))
	pairs := make([]string, 0, len(v))
	for _, key := range keys {
		pairs = append(pairs, FormatAssignment(key, v[key]))
	}
	return pairs
}

func (v Vars) String() string {
	return strings.Join(v.PairStrings(), " ")
}

func (v Vars) Lines() string {
	return strings.Join(v.PairStrings(), "\n")
}
