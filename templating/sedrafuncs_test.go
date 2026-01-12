package templating_test

import (
	"fmt"
	"testing"

	"github.com/hebcal/hdate"

	"github.com/chaimleib/hebcalfmt/templating"
)

func TestSedra(t *testing.T) {
	const year = 5786
	for _, il := range []bool{false, true} {
		t.Run(fmt.Sprintf("IL=%v", il), func(t *testing.T) {
			sedra := templating.Sedra(year, il)
			p := sedra.Lookup(hdate.New(year, hdate.Adar2, 11))
			if "Parashat Tetzaveh" != p.String() {
				t.Errorf("want: Parashat Tetzaveh, got: %s", p)
			}

			// Do it again to verify that the cache works.
			p2 := sedra.Lookup(hdate.New(year, hdate.Adar2, 11))
			if "Parashat Tetzaveh" != p2.String() {
				t.Errorf("cache - want: Parashat Tetzaveh, got: %s", p2)
			}
		})
	}
}
