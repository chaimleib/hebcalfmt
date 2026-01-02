package hcfiles_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/hebcal"

	"github.com/chaimleib/hebcalfmt/hcfiles"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestParseEvents(t *testing.T) {
	const fileName = "testEvents.txt"
	cases := []struct {
		Name    string
		Content string
		WantErr string
		Want    []hebcal.UserEvent
	}{
		{Name: "empty", Content: "", WantErr: "", Want: nil},
		{
			Name:    "basic",
			Content: "Cheshvan 03 Joe Shmo",
			WantErr: "",
			Want: []hebcal.UserEvent{
				{
					Month: hdate.Cheshvan,
					Day:   3,
					Desc:  "Joe Shmo",
				},
			},
		},
		{
			Name:    "multiple entries",
			Content: "Cheshvan 03 Joe Shmo\nKislev 6 Jane Doe",
			WantErr: "",
			Want: []hebcal.UserEvent{
				{
					Month: hdate.Cheshvan,
					Day:   3,
					Desc:  "Joe Shmo",
				},
				{
					Month: hdate.Kislev,
					Day:   6,
					Desc:  "Jane Doe",
				},
			},
		},
		{
			Name:    "invalid line",
			Content: "INVALID",
			WantErr: "ParseEvents: " + hcfiles.SyntaxError{
				Err: fmt.Errorf(
					"%w: expected 4 capture fields, got 0",
					hcfiles.ErrInvalidFormat,
				),
				FileName:   fileName,
				LineNumber: 1,
			}.Error(),
			Want: nil,
		},
		{
			Name:    "invalid lines",
			Content: "INVALID\nWRONG",
			WantErr: "ParseEvents: " + errors.Join(
				hcfiles.SyntaxError{
					Err: fmt.Errorf(
						"%w: expected 4 capture fields, got 0",
						hcfiles.ErrInvalidFormat,
					),
					FileName:   fileName,
					LineNumber: 1,
				},
				hcfiles.SyntaxError{
					Err: fmt.Errorf(
						"%w: expected 4 capture fields, got 0",
						hcfiles.ErrInvalidFormat,
					),
					FileName:   fileName,
					LineNumber: 2,
				},
			).Error(),
			Want: nil,
		},
		{
			Name:    "invalid month",
			Content: "13 03 Joe Shmo",
			WantErr: "ParseEvents: " + hcfiles.SyntaxError{
				Err:        hcfiles.ErrInvalidMonth,
				FileName:   fileName,
				LineNumber: 1,
			}.Error(),
			Want: nil,
		},
		{
			Name:    "invalid day",
			Content: "Cheshvan 32 Joe Shmo",
			WantErr: "ParseEvents: " + hcfiles.SyntaxError{
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
			got, err := hcfiles.ParseEvents(f, fileName)
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
