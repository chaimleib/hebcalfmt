package shell_test

import (
	"fmt"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/shell"
)

func TestIsAlphaOrUnderscore(t *testing.T) {
	cases := []struct {
		R    rune
		Want bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'/', false},
		{'-', false},
		{'_', true},
		{'0', false},
		{'9', false},
		{'!', false},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q %v", string(c.R), c.Want)
		t.Run(name, func(t *testing.T) {
			got := shell.IsAlphaOrUnderscore(c.R)
			test.CheckComparable(t, "got", c.Want, got)
		})
	}
}

func TestIsDigit(t *testing.T) {
	cases := []struct {
		R    rune
		Want bool
	}{
		{'a', false},
		{'z', false},
		{'A', false},
		{'Z', false},
		{'/', false},
		{'-', false},
		{'_', false},
		{'0', true},
		{'9', true},
		{'!', false},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q %v", string(c.R), c.Want)
		t.Run(name, func(t *testing.T) {
			got := shell.IsDigit(c.R)
			test.CheckComparable(t, "got", c.Want, got)
		})
	}
}

func TestParseIdentifier(t *testing.T) {
	cases := []struct {
		Name     string
		Rest     string
		Line     parsing.LineInfo
		Want     string
		WantRest string
		Err      string
	}{
		{Name: "empty", Err: "no match"},
		{
			Name: "invalid initial unicode",
			Rest: "\x80 invalid unicode",
			Line: parsing.LineInfo{
				Line:     "\x80 invalid unicode",
				Number:   1,
				FileName: "invalid-unicode.sh",
			},
			WantRest: "\x80 invalid unicode",
			Err: `syntax at invalid-unicode.sh:1:1: invalid unicode

	� invalid unicode
	^                `,
		},
		{
			Name: "invalid internal unicode",
			Rest: "hi\x80 invalid unicode",
			Line: parsing.LineInfo{
				Line:     "hi\x80 invalid unicode",
				Number:   1,
				FileName: "invalid-unicode.sh",
			},
			WantRest: "hi\x80 invalid unicode",
			Err: `syntax at invalid-unicode.sh:1:3: invalid unicode

	hi� invalid unicode
	  ^                `,
		},
		{
			Name: "bad initial digit",
			Rest: "80invalid",
			Line: parsing.LineInfo{
				Line:     "80invalid",
				Number:   1,
				FileName: "invalid-initial-digit.sh",
			},
			WantRest: "80invalid",
			Err:      "no match",
		},
		{
			Name: "ok",
			Rest: "ok",
			Line: parsing.LineInfo{
				Line:     "ok",
				Number:   1,
				FileName: "ok.sh",
			},
			Want: "ok",
		},
		{
			Name: "ok mixed",
			Rest: "_ok_2",
			Line: parsing.LineInfo{
				Line:     "_ok_2",
				Number:   1,
				FileName: "ok.sh",
			},
			Want: "_ok_2",
		},
		{
			Name: "ok after $",
			Rest: "ok",
			Line: parsing.LineInfo{
				Line:     "echo $ok",
				Number:   1,
				FileName: "ok.sh",
			},
			Want: "ok",
		},
		{
			Name: "ok in assignment",
			Rest: "ok=true",
			Line: parsing.LineInfo{
				Line:     "ok=true",
				Number:   1,
				FileName: "ok.sh",
			},
			Want:     "ok",
			WantRest: "=true",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, rest, err := shell.ParseIdentifier(c.Line, []byte(c.Rest))
			test.CheckErr(t, err, c.Err)
			test.CheckComparable(t, "got", c.Want, got)
			test.CheckComparable(t, "rest", c.WantRest, string(rest))
		})
	}
}

func TestParseAssignment(t *testing.T) {
	cases := []struct {
		Name           string
		Rest           string
		Line           parsing.LineInfo
		WantIdentifier string
		WantValue      string
		WantRest       string
		Err            string
	}{
		{Name: "empty", Err: "no match"},
		{
			Name: "invalid initial unicode",
			Rest: "\x80invalid=unicode",
			Line: parsing.LineInfo{
				Line:     "\x80invalid=unicode",
				Number:   1,
				FileName: "invalid-unicode.sh",
			},
			WantRest: "\x80invalid=unicode",
			Err: `syntax at invalid-unicode.sh:1:1: invalid unicode

	�invalid=unicode
	^               `,
		},
		{
			Name: "invalid internal unicode",
			Rest: "hi\x80invalid=unicode",
			Line: parsing.LineInfo{
				Line:     "hi\x80invalid=unicode",
				Number:   1,
				FileName: "invalid-unicode.sh",
			},
			WantRest: "hi\x80invalid=unicode",
			Err: `syntax at invalid-unicode.sh:1:3: invalid unicode

	hi�invalid=unicode
	  ^               `,
		},
		{
			Name: "bad initial digit",
			Rest: "80invalid=value",
			Line: parsing.LineInfo{
				Line:     "80invalid=value",
				Number:   1,
				FileName: "invalid-initial-digit.sh",
			},
			WantRest: "80invalid=value",
			Err:      "no match",
		},
		{
			Name: "close paren in value",
			Rest: "key=)value",
			Line: parsing.LineInfo{
				Line:     "key=)value",
				Number:   1,
				FileName: "value-close-paren.sh",
			},
			WantIdentifier: "key",
			WantRest:       ")value",
		},
		{
			Name: "missing equal",
			Rest: "missing equal",
			Line: parsing.LineInfo{
				Line:     "missing equal",
				Number:   1,
				FileName: "missing-equal.sh",
			},
			WantRest: "missing equal",
			Err:      "no match",
		},
		{
			Name: "invalid string",
			Rest: "invalid='string",
			Line: parsing.LineInfo{
				Line:     "invalid='string",
				Number:   1,
				FileName: "invalid-string.sh",
			},
			WantRest: "invalid='string",
			Err: `syntax at invalid-string.sh:1:10-16: expected ending ` + "`'`" + ` in this span

	invalid='string
	         ^^^^^^`,
		},
		{
			Name: "ok raw",
			Rest: "ok=value",
			Line: parsing.LineInfo{
				Line:     "ok=value",
				Number:   1,
				FileName: "ok.sh",
			},
			WantIdentifier: "ok",
			WantValue:      "value",
		},
		{
			Name: "ok squote",
			Rest: "ok='value'",
			Line: parsing.LineInfo{
				Line:     "ok='value'",
				Number:   1,
				FileName: "ok.sh",
			},
			WantIdentifier: "ok",
			WantValue:      "value",
		},
		{
			Name: "ok dquote",
			Rest: `ok="value"`,
			Line: parsing.LineInfo{
				Line:     `ok="value"`,
				Number:   1,
				FileName: "ok.sh",
			},
			WantIdentifier: "ok",
			WantValue:      "value",
		},
		{
			Name: "ok mixed",
			Rest: `_ok_2=v'al'"ue"`,
			Line: parsing.LineInfo{
				Line:     `_ok_2=v'al'"ue"`,
				Number:   1,
				FileName: "ok.sh",
			},
			WantIdentifier: "_ok_2",
			WantValue:      "value",
		},
		{
			Name: "ok in series",
			Rest: "ok=value other=item",
			Line: parsing.LineInfo{
				Line:     "name=Joe ok=value other=item",
				Number:   1,
				FileName: "ok.sh",
			},
			WantIdentifier: "ok",
			WantValue:      "value",
			WantRest:       " other=item",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			key, value, rest, err := shell.ParseAssignment(c.Line, []byte(c.Rest))
			test.CheckErr(t, err, c.Err)
			test.CheckComparable(t, "key", c.WantIdentifier, key)
			test.CheckComparable(t, "value", c.WantValue, value)
			test.CheckComparable(t, "rest", c.WantRest, string(rest))
		})
	}
}

func TestFormatAssignment(t *testing.T) {
	cases := []struct {
		Key   string
		Value string
		Want  string
	}{
		// {"key", "value", "key=value"},
		// {"key", "spaced value", `key="spaced value"`},
		{"key", "special!", `key="special!"`},
		{"key", "\ttabbed", `key="\ttabbed"`},
		{"spaces", "   ", `spaces="   "`},
	}
	for _, c := range cases {
		t.Run(c.Want, func(t *testing.T) {
			got := shell.FormatAssignment(c.Key, c.Value)
			test.CheckComparable(t, "result", c.Want, got)
		})
	}
}

func TestVars_String(t *testing.T) {
	cases := []struct {
		Name string
		Vars shell.Vars
		Want string
	}{
		{Name: "empty"},
		{Name: "one pair", Vars: shell.Vars{"key": "value"}, Want: `key=value`},
		{
			Name: "two pairs",
			Vars: shell.Vars{"key": "value", "other": "item"},
			Want: `key=value other=item`,
		},
		{
			Name: "space in value",
			Vars: shell.Vars{"spaces": "spaced value"},
			Want: `spaces="spaced value"`,
		},
		{
			Name: "specials in value",
			Vars: shell.Vars{"specials": "^\t[value]$"},
			Want: `specials="^\t[value]\$"`,
		},
	}
	for _, c := range cases {
		t.Run(c.Want, func(t *testing.T) {
			got := c.Vars.String()
			test.CheckComparable(t, "result", c.Want, got)
		})
	}
}

func TestVars_Lines(t *testing.T) {
	cases := []struct {
		Name string
		Vars shell.Vars
		Want string
	}{
		{Name: "empty"},
		{Name: "one pair", Vars: shell.Vars{"key": "value"}, Want: `key=value`},
		{
			Name: "two pairs",
			Vars: shell.Vars{"key": "value", "aValue": "item"},
			Want: "aValue=item\nkey=value",
		},
		{
			Name: "space in value",
			Vars: shell.Vars{"spaces": "spaced value"},
			Want: `spaces="spaced value"`,
		},
		{
			Name: "specials in value",
			Vars: shell.Vars{"specials": "^\t[value]$"},
			Want: `specials="^\t[value]\$"`,
		},
	}
	for _, c := range cases {
		t.Run(c.Want, func(t *testing.T) {
			got := c.Vars.Lines()
			test.CheckComparable(t, "result", c.Want, got)
		})
	}
}
