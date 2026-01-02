package hcfiles

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/hebcal"
)

var hebRe = regexp.MustCompile(`^(\S+)\s+(\d+)\s+(.+)$`)

// ParseEvents parses an [io.Reader] of event lines
// and returns a slice of [hebcal.UserEvent] entries.
// In case of an error, fileName helps with debugging.
//
// The lines are in the following format, using Hebrew dates:
//
//	MMMM DD Description
//
// MMMM is a string identifying the Hebrew month.
// Description is a newline-terminated string.
func ParseEvents(f io.Reader, fileName string) ([]hebcal.UserEvent, error) {
	lineNumber := 0
	var errs []error
	lineErr := func(err error) {
		errs = append(errs, SyntaxError{
			Err:        err,
			FileName:   fileName,
			LineNumber: lineNumber,
		})
	}

	entries := make([]hebcal.UserEvent, 0, 10)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line0 := scanner.Text()
		line := strings.TrimSpace(line0)
		lineNumber++

		fields := hebRe.FindStringSubmatch(line)
		if len(fields) != 4 {
			lineErr(
				fmt.Errorf(
					"%w: expected 4 capture fields, got %d",
					ErrInvalidFormat,
					len(fields),
				),
			)
			continue
		}

		month, err := hdate.MonthFromName(fields[1])
		if err != nil {
			lineErr(ErrInvalidMonth)
			continue
		}

		day, _ := strconv.Atoi(fields[2])
		if day < 1 || day > 30 {
			lineErr(ErrInvalidDays)
			continue
		}

		entries = append(
			entries,
			hebcal.UserEvent{Month: month, Day: day, Desc: fields[3]},
		)
	}

	if len(errs) != 0 {
		return nil, fmt.Errorf("ParseEvents: %w", errors.Join(errs...))
	}
	return entries, nil
}
