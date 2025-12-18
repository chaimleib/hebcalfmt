package hcfiles

import (
	"errors"
	"fmt"
)

type SyntaxError struct {
	err        error
	filename   string
	lineNumber int
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("error at %s:%d: %v", e.filename, e.lineNumber, e.err)
}

var (
	errInvalidFormat = errors.New("invalid format")
	errInvalidMonth  = errors.New("invalid month")
	errInvalidDays   = errors.New("invalid days")
)
