package hcfiles

import (
	"errors"
	"fmt"
)

type SyntaxError struct {
	Err        error
	FileName   string
	LineNumber int
}

var _ error = SyntaxError{}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("error at %s:%d: %v", e.FileName, e.LineNumber, e.Err)
}

func (e SyntaxError) Unwrap() error { return e.Err }

var (
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidMonth  = errors.New("invalid month")
	ErrInvalidDays   = errors.New("invalid days")
)
