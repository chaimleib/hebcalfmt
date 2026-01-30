package shell

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/chaimleib/hebcalfmt/test/parsing"
)

const (
	// DquoteSpecialChars are characters that could interrupt
	// a double-quoted shell string unless escaped.
	DquoteSpecialChars = "$\r\n\t\"`"

	// DquoteNonliteralChars are characters whose escaped form
	// does not not end with their literal value.
	// For example, "\$" has a literal value of "$", so is not included.
	// But "\n" does not have a literal value of "n", so it is included.
	DquoteNonliteralChars = "\r\n\t"

	// SpecialChars are characters that could interrupt a raw shell string
	// unless quoted or escaped.
	SpecialChars = DquoteSpecialChars + `!()[]{}<>|!&*? ;#'"\`
)

func ParseShellStringDquote(
	li parsing.LineInfo,
	rest []byte,
) (value string, newRest []byte, err error) {
	buf, errOut, noMatchOut, okOut := BufferReturn(li, &rest)

	var ok bool
	if rest, ok = bytes.CutPrefix(rest, []byte(`"`)); !ok {
		return noMatchOut()
	}

	for len(rest) > 0 {
		r, size := utf8.DecodeRune(rest)
		if r == utf8.RuneError {
			return errOut(size, "invalid unicode")
		}
		if size > 1 {
			buf.WriteRune(r)
			rest = rest[size:]
			continue
		}

		i := bytes.IndexAny(rest, `"\`)
		if i < 0 {
			return errOut(len(rest)+1, "expected ending '\"' in this span")
		}

		buf.Write(rest[:i])
		rest = rest[i:]
		if rest[0] == '"' { // end of string?
			break
		}

		// parse escapes
		rest = rest[1:] // slice off \
		if len(rest) == 0 {
			return errOut(0, "unexpected end after escape char")
		}
		size = 1
		switch rest[0] {
		case 'n':
			buf.WriteRune('\n')
		case 'r':
			buf.WriteRune('\r')
		case 't':
			buf.WriteRune('\t')
		default:
			r, size = utf8.DecodeRune(rest)
			if r == utf8.RuneError {
				return errOut(size, "invalid unicode")
			}
			buf.WriteRune(r)
		}
		rest = rest[size:] // slice off escaped char
	} // get rest of string

	if rest, ok = bytes.CutPrefix(rest, []byte(`"`)); !ok {
		return errOut(0, "expected ending '\"' for double-quoted shell string")
	}

	return okOut()
}

func ParseShellStringSquote(
	li parsing.LineInfo,
	rest []byte,
) (value string, newRest []byte, err error) {
	buf, errOut, noMatchOut, okOut := BufferReturn(li, &rest)

	var ok bool
	if rest, ok = bytes.CutPrefix(rest, []byte(`'`)); !ok {
		return noMatchOut()
	}

	i := bytes.IndexAny(rest, `'`)
	if i < 0 {
		return errOut(len(rest)+1, "expected ending `'` in this span")
	}

	buf.Write(rest[:i])
	rest = rest[i:]

	rest = bytes.TrimPrefix(rest, []byte(`'`))

	return okOut()
}

func ParseShellStringRaw(
	li parsing.LineInfo,
	rest []byte,
) (value string, newRest []byte, err error) {
	buf, errOut, noMatchOut, _ := BufferReturn(li, &rest)
	okOut := func() (string, []byte, error) {
		if buf.Len() == 0 {
			return noMatchOut()
		}
		return buf.String(), rest, nil
	}

	if len(rest) == 0 {
		return noMatchOut()
	}

	for len(rest) > 0 {
		r, size := utf8.DecodeRune(rest)
		if r == utf8.RuneError {
			return errOut(0, "invalid unicode in raw string")
		}
		// multibyte unicode is not special
		if size != 1 || !bytes.ContainsAny(rest[:1], SpecialChars) {
			buf.WriteRune(r)
			rest = rest[size:]
			continue
		}
		// some sort of special

		var ok bool
		rest, ok = bytes.CutPrefix(rest, []byte("\\"))
		if !ok {
			return okOut()
		}

		// parse escaped char
		if len(rest) == 0 {
			return errOut(0, "unexpected end of raw string after escape")
		}
		r, size = utf8.DecodeRune(rest)
		if r == utf8.RuneError {
			return errOut(0, "invalid unicode after escape in raw string")
		}
		buf.WriteRune(r)
		rest = rest[size:]
	} // get rest of string

	return okOut()
}

func ParseShellString(
	li parsing.LineInfo,
	rest []byte,
) (value string, newRest []byte, err error) {
	var buf bytes.Buffer
	orig := rest
	var couldBeExplicitEmpty bool
	okOut := func() (string, []byte, error) {
		if buf.Len() == 0 && !couldBeExplicitEmpty {
			return "", orig, ErrNoMatch
		}
		return buf.String(), rest, nil
	}

concat:
	for len(rest) > 0 {
		var s string
		var err error
		switch rest[0] {
		case '"':
			s, rest, err = ParseShellStringDquote(li, rest)
			couldBeExplicitEmpty = true
		case '\'':
			s, rest, err = ParseShellStringSquote(li, rest)
			couldBeExplicitEmpty = true
		case ' ':
			break concat
		default:
			s, rest, err = ParseShellStringRaw(li, rest)
			if errors.Is(err, ErrNoMatch) {
				break concat
			}
		}
		if err != nil && !errors.Is(err, ErrNoMatch) {
			return "", orig, err
		}
		buf.WriteString(s)
	}

	return okOut()
}

// FormatString returns value if it is not empty
// and it equals its raw shell string rendering.
// Otherwise, we escape as needed and wrap it in double-quotes.
func FormatString(value string) string {
	if value == "" {
		return `""`
	}
	if !strings.ContainsAny(value, SpecialChars) {
		return value
	}
	return DquoteString(value)
}

func DquoteString(value string) string {
	var count int
	for _, r := range value {
		if strings.ContainsRune(DquoteSpecialChars, r) {
			count++
		}
	}
	if count == 0 {
		return fmt.Sprintf("%q", value)
	}

	var buf bytes.Buffer
	buf.Grow(2 + len(value) + count)
	buf.WriteRune('"')

	for _, r := range value {
		if strings.ContainsRune(DquoteSpecialChars, r) {
			buf.WriteRune('\\')
			switch r {
			case '\n':
				buf.WriteRune('n')
				continue
			case '\r':
				buf.WriteRune('r')
				continue
			case '\t':
				buf.WriteRune('t')
				continue
			}
		}
		buf.WriteRune(r)
	}

	buf.WriteRune('"')
	return buf.String()
}
