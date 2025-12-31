package hcfiles_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hebcal/hebcal-go/hebcal"

	"github.com/chaimleib/hebcalfmt/hcfiles"
	"github.com/chaimleib/hebcalfmt/test"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestParseYahrzeits(t *testing.T) {
	const fileName = "testYahrzeit.txt"
	cases := []struct {
		Name    string
		Content string
		WantErr string
		Want    []hebcal.UserYahrzeit
	}{
		{Name: "empty", Content: "", WantErr: "", Want: nil},
		{
			Name:    "basic",
			Content: "02 03 2004 Joe Shmo",
			WantErr: "",
			Want: []hebcal.UserYahrzeit{
				{
					Date: date(2004, time.February, 3),
					Name: "Joe Shmo",
				},
			},
		},
		{
			Name:    "multiple entries",
			Content: "02 03 2004 Joe Shmo\n5 6 2001 Jane Doe",
			WantErr: "",
			Want: []hebcal.UserYahrzeit{
				{
					Date: date(2004, time.February, 3),
					Name: "Joe Shmo",
				},
				{
					Date: date(2001, time.May, 6),
					Name: "Jane Doe",
				},
			},
		},
		{
			Name:    "invalid line",
			Content: "INVALID",
			WantErr: hcfiles.SyntaxError{
				Err:        hcfiles.ErrInvalidFormat,
				FileName:   fileName,
				LineNumber: 1,
			}.Error(),
			Want: nil,
		},
		{
			Name:    "invalid lines",
			Content: "INVALID\nWRONG",
			WantErr: errors.Join(
				hcfiles.SyntaxError{
					Err:        hcfiles.ErrInvalidFormat,
					FileName:   fileName,
					LineNumber: 1,
				},
				hcfiles.SyntaxError{
					Err:        hcfiles.ErrInvalidFormat,
					FileName:   fileName,
					LineNumber: 2,
				},
			).Error(),
			Want: nil,
		},
		{
			Name:    "invalid month",
			Content: "13 03 2004 Joe Shmo",
			WantErr: hcfiles.SyntaxError{
				Err:        hcfiles.ErrInvalidMonth,
				FileName:   fileName,
				LineNumber: 1,
			}.Error(),
			Want: nil,
		},
		{
			Name:    "invalid day",
			Content: "2 29 2001 Joe Shmo",
			WantErr: hcfiles.SyntaxError{
				Err:        hcfiles.ErrInvalidDays,
				FileName:   fileName,
				LineNumber: 1,
			}.Error(),
			Want: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			f := strings.NewReader(c.Content)
			got, err := hcfiles.ParseYahrzeits(f, fileName)
			test.CheckErr(t, err, c.WantErr)
			var i, j int
			for j = range c.Want {
				if i >= len(got) {
					t.Errorf(
						"unexpected extra item at index %d, skipping rest:\n%v",
						i,
						got[i],
					)
					break
				}
				if c.Want[j] != got[i] {
					t.Errorf("unexpected item at index %d:\n%v\nwant:\n%v",
						i, got[i], c.Want[j])
					// assume other lines will match, so allow line to increment
				}
				i++
			}
			if len(c.Want) != len(got) {
				t.Errorf("expected %d results, got %d", len(c.Want), len(got))
			}
		})
	}
}
