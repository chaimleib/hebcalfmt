package templating_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/templating"
)

func TestItof(t *testing.T) {
	got := templating.Itof(3)
	if got != 3.0 {
		t.Errorf("want: %f\ngot:  %f", 3.0, got)
	}
}

func TestFtoi(t *testing.T) {
	got := templating.Ftoi(4.0)
	if got != 4 {
		t.Errorf("want: %d\ngot:  %d", 4, got)
	}
}
