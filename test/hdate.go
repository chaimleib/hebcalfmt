package test

import (
	"github.com/hebcal/hdate"

	"github.com/chaimleib/hebcalfmt/xhdate"
)

// CheckHDate checks to see if the Day, Month and Year all match.
// It is needed, because the struct also caches the Rata Die date,
// and that field may or may not be populated.
func CheckHDate(t Test, label string, want, got hdate.HDate) {
	t.Helper()
	if !xhdate.Equal(want, got) {
		t.Errorf("%s did not match - want:\n%s\ngot:\n%s", label, want, got)
	}
}
