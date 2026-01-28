package warning

import (
	"errors"
	"fmt"
)

var ErrWarn = errors.New("warn")

type Warnings []error

func (w Warnings) Build() error {
	switch len(w) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("%w: %w", ErrWarn, w[0])
	default:
		return fmt.Errorf("%d %wings:\n%w",
			len(w), ErrWarn, errors.Join(w...))
	}
}

func (w Warnings) Join(err error) error {
	warning := w.Build()
	if warning == nil {
		return err
	}
	if err != nil {
		return fmt.Errorf("%w\n%w", warning, err)
	}
	return warning
}

func (w *Warnings) Append(warn error) { *w = append(*w, warn) }
