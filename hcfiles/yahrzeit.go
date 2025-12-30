package hcfiles

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hebcal/greg"
	"github.com/hebcal/hebcal-go/hebcal"
)

var gregRe = regexp.MustCompile(`^(\d+)\s+(\d+)\s+(\d+)\s+(.+)$`)

// ParseYahrzeits parses an [io.Reader] of event lines
// and returns a slice of [hebcal.UserYahrzeit] entries.
// In case of an error, fileName helps with debugging.
//
// The lines are in the following format, using Gregorian dates:
//
//	MM DD YYYY Description
//
// Description is a newline-terminated string.
func ParseYahrzeits(
	f io.Reader,
	fileName string,
) ([]hebcal.UserYahrzeit, error) {
	lineNumber := 0
	var errs []error
	lineErr := func(err error) {
		errs = append(errs, SyntaxError{
			Err:        err,
			FileName:   fileName,
			LineNumber: lineNumber,
		})
	}

	entries := make([]hebcal.UserYahrzeit, 0, 10)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line0 := scanner.Text()
		line := strings.TrimSpace(line0)
		lineNumber++

		fields := gregRe.FindStringSubmatch(line)
		if len(fields) != 5 {
			lineErr(errInvalidFormat)
			continue
		}

		month0, _ := strconv.Atoi(fields[1])
		if month0 < 1 || month0 > 12 {
			lineErr(errInvalidMonth)
		}

		day, _ := strconv.Atoi(fields[2])
		year, _ := strconv.Atoi(fields[3])
		month := time.Month(month0)
		if day < 1 || day > greg.DaysIn(month, year) {
			lineErr(errInvalidDays)
		}

		gregDate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		entries = append(
			entries,
			hebcal.UserYahrzeit{Date: gregDate, Name: fields[4]},
		)
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}
	return entries, nil
}
