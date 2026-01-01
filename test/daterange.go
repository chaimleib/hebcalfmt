package test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/daterange"
)

func CheckDateRangeSource(t *testing.T, want, got daterange.Source) {
	// Check the simple fields.
	if want.IsHebrewDate != got.IsHebrewDate {
		t.Errorf("Source.IsHebrewDate's do not match - want: %v, got: %v",
			want.IsHebrewDate, got.IsHebrewDate)
	}
	if want.Now != got.Now {
		t.Errorf("Source.Now's do not match - want:\n%s\ngot:\n%s",
			want.Now, got.Now)
	}
	if (want.FromTime == nil) != (got.FromTime == nil) {
		t.Errorf("Source.FromTime's nilness do not match - want:\n%s\ngot:\n%s",
			want.FromTime, got.FromTime)
	} else if want.FromTime != nil && *want.FromTime != *got.FromTime {
		// either != nil implies the other != nil b/c of prev test
		t.Errorf("Source.FromTime's values do not match - want:\n%s\ngot:\n%s",
			want.FromTime, got.FromTime)
	}

	// Diff the Args
	wantArgsNil := want.Args == nil
	gotArgsNil := got.Args == nil
	if len(want.Args) != len(got.Args) {
		t.Errorf("length of Source.Args does not match - want: %d, got: %d",
			len(want.Args), len(got.Args))
	} else if wantArgsNil != gotArgsNil {
		t.Errorf("nil-ness of Source.Args does not match - want.Source.Args==nil: %v, got.Source.Args==nil: %v",
			wantArgsNil, gotArgsNil)
	}
	var i, j int
	for j = range got.Args {
		if i >= len(want.Args) {
			t.Errorf(
				"unexpected extra Source.Args at index %d, skipping rest - got: %q",
				j,
				got.Args[j],
			)
			break
		}
		if want.Args[i] != got.Args[j] {
			t.Errorf(
				"unexpected Source.Args[%d] - want: %q, got: %q",
				j,
				want.Args[i],
				got.Args[j],
			)
			continue
		}
		i++
	}
}

func CheckDateRange(t *testing.T, want, got daterange.DateRange) {
	CheckDateRangeSource(t, want.Source, got.Source)
	for _, field := range []struct {
		Name      string
		Want, Got any
	}{
		{"RangeType", want.RangeType, got.RangeType},
		{"Day", want.Day, got.Day},
		{"GregMonth", want.GregMonth, got.GregMonth},
		{"HebMonth", want.HebMonth, got.HebMonth},
		{"Year", want.Year, got.Year},
		{"IsHebrewDate", want.IsHebrewDate, got.IsHebrewDate},
	} {
		if field.Want != field.Got {
			t.Errorf("%s's did not match - want: %v, got: %v",
				field.Name, field.Want, field.Got)
		}
	}
}
