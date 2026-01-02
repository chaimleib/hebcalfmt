package test

import (
	"testing"

	"github.com/hebcal/hdate"
)

// CheckHDate checks to see if the Day, Month and Year all match.
// It is needed, because the struct also caches the Rata Die date,
// and that field may or may not be populated.
func CheckHDate(t *testing.T, label string, want, got hdate.HDate) {
	ok := want.Day() == got.Day() &&
		want.Month() == got.Month() &&
		want.Year() == got.Year()
	if !ok {
		t.Errorf("%s did not match - want:\n%s\ngot:\n%s", label, want, got)
	}
}
