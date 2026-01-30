package shell

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/chaimleib/hebcalfmt/test/parsing"
)

var ErrNoMatch = errors.New("no match")

func ErrOut[T any](
	line parsing.LineInfo,
	rest *[]byte,
) func(span int, msg string, args ...any) (T, []byte, error) {
	orig := *rest
	var zero T
	return func(span int, msg string, args ...any) (T, []byte, error) {
		col := 1 + len(line.Line) - len(*rest)
		var colEnd int
		if span > 0 {
			colEnd = col + span - 1
		}
		return zero, orig, parsing.NewSyntaxError(
			line, col, colEnd,
			fmt.Errorf(msg, args...),
		)
	}
}

func BufferReturn(
	line parsing.LineInfo,
	rest *[]byte,
) (
	buf *bytes.Buffer,
	errOut func(span int, msg string, args ...any) (string, []byte, error),
	noMatchOut func() (string, []byte, error),
	okOut func() (string, []byte, error),
) {
	orig := *rest
	noMatchOut = func() (string, []byte, error) {
		return "", orig, ErrNoMatch
	}

	buf = new(bytes.Buffer)
	okOut = func() (string, []byte, error) {
		return buf.String(), *rest, nil
	}

	errOut = ErrOut[string](line, rest)

	return buf, errOut, noMatchOut, okOut
}
