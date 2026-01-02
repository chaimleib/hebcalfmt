// Package xhdate contains functions relevant to hebcal/hdate
// which are not yet upstreamed.
package xhdate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hebcal/hdate"
)

// Equal returns whether the Day, Month and Year all match.
// It is needed, because the struct also caches the Rata Die date,
// and that field may or may not be populated.
func Equal(a, b hdate.HDate) bool {
	return a.Day() == b.Day() &&
		a.Month() == b.Month() &&
		a.Year() == b.Year()
}

// Parse parses a string in DD MMMM YYYY or MMMM DD YYYY format into an HDate.
// MMMM is the name of the Hebrew month, like "Adar II", "Adar 2", or "Tishrei".
func Parse(s string) (hdate.HDate, error) {
	var rv hdate.HDate

	parts := strings.Fields(s)
	length := len(parts)
	if length > 4 { // 4 in case of Adar I/II/1/2
		return rv, fmt.Errorf("too many words in a Hebrew date: %q", s)
	}
	if length < 3 {
		return rv, fmt.Errorf("too few words in a Hebrew date: %q", s)
	}

	year, err := strconv.Atoi(parts[length-1])
	if err != nil {
		return rv, fmt.
			Errorf("could not parse last word of Hebrew date as year: %q", s)
	}

	var monthParts []string
	day, err := strconv.Atoi(parts[0])
	if err != nil {
		day, err = strconv.Atoi(parts[length-2])
		if err != nil {
			return rv, fmt.Errorf("could not parse day from Hebrew date: %q", s)
		}
		monthParts = parts[0 : length-2]
	} else {
		monthParts = parts[1 : length-1]
	}

	monthStr := strings.Join(monthParts, " ")
	month, err := hdate.MonthFromName(monthStr)
	if err != nil {
		return rv, fmt.Errorf("failed to parse month from Hebrew date: %q", s)
	}

	if day < 1 || day > hdate.DaysInMonth(month, year) {
		return rv, fmt.Errorf("invalid day of month for %s %d: %d",
			month, year, day)
	}

	return hdate.New(year, month, day), nil
}
