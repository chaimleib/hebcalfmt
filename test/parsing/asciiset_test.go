package parsing_test

import (
	"testing"
	"unicode/utf8"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
)

func TestMakeASCIISet(t *testing.T) {
	cases := []struct {
		Name string
		Set  string
		Want bool
	}{
		{Name: "empty", Want: true},
		{Name: "hex", Set: "0123456789abcdefABCDEF", Want: true},
		{Name: "utf8", Set: "🤩", Want: false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, got := parsing.MakeASCIISet([]byte(c.Set))
			test.CheckComparable(t, "ok", c.Want, got)
		})
	}
}

func TestASCIISet_ContainsRune(t *testing.T) {
	hex, ok := parsing.MakeASCIISet([]byte("0123456789abcdefABCDEF"))
	if !ok {
		t.Fatal("failed to build hex ASCIISet")
	}

	cases := []struct {
		Name string
		Rune rune
		Want bool
	}{
		{Name: "zero"},
		{Rune: 'a', Want: true},
		{Rune: 'A', Want: true},
		{Rune: 'z', Want: false},
		{Rune: 'Z', Want: false},
		{Rune: '0', Want: true},
		{Rune: '9', Want: true},
		{Rune: '(', Want: false},
		{Rune: ' ', Want: false},
		{Rune: '🤩', Want: false},
		{Rune: utf8.RuneSelf, Want: false},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			name = string(c.Rune)
		}
		t.Run(c.Name, func(t *testing.T) {
			got := hex.ContainsRune(c.Rune)
			test.CheckComparable(t, "isIn", c.Want, got)
		})
	}
}

func TestASCIISet_TrimLeft(t *testing.T) {
	hex, ok := parsing.MakeASCIISet([]byte("0123456789abcdefABCDEF"))
	if !ok {
		t.Fatal("failed to build hex ASCIISet")
	}

	cases := []struct {
		Name  string
		Input string
		Want  string
	}{
		{Name: "empty"},
		{Input: "0xdeadbeef", Want: "xdeadbeef"},
		{Input: "dead beef", Want: " beef"},
		{Input: "not hex", Want: "not hex"},
		{Input: "123🤩 unicode", Want: "🤩 unicode"},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			name = c.Input
		}
		t.Run(name, func(t *testing.T) {
			got := hex.TrimLeft([]byte(c.Input))
			test.CheckString(t, "trimmed", c.Want, string(got))
		})
	}
}
