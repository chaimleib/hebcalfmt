package templating_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestMap(t *testing.T) {
	cases := []struct {
		Name string
		Args []any
		Want map[string]any
		Err  string
	}{
		{Name: "empty"},
		{
			Name: "bad: odd args 1",
			Args: []any{"key"},
			Err:  "must provide an even number of arguments to make key-value pairs",
		},
		{
			Name: "bad: odd args 3",
			Args: []any{"key", "value", "unmatched"},
			Err:  "must provide an even number of arguments to make key-value pairs",
		},
		{
			Name: "bad: non-string key",
			Args: []any{2, "value"},
			Err:  "arg index 0 should have been a string, got a int: 2",
		},
		{
			Name: "1 string",
			Args: []any{"key", "value"},
			Want: map[string]any{"key": "value"},
		},
		{
			Name: "2 strings",
			Args: []any{"key1", "a", "key2", "b"},
			Want: map[string]any{"key1": "a", "key2": "b"},
		},
		{
			Name: "1 int",
			Args: []any{"key", 1},
			Want: map[string]any{"key": 1},
		},
		{
			Name: "2 ints",
			Args: []any{"key1", 1, "key2", 2},
			Want: map[string]any{"key1": 1, "key2": 2},
		},
		{
			Name: "1 string 1 int",
			Args: []any{"string", "value", "int", 2},
			Want: map[string]any{"string": "value", "int": 2},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := templating.MakeMap(c.Args...)
			test.CheckErr(t, err, c.Err)
			test.CheckMap(t, "map", c.Want, got)
		})
	}
}

func TestList(t *testing.T) {
	cases := []struct {
		Name string
		Args []any
	}{
		{Name: "empty"},
		{Name: "1 string", Args: []any{"hello"}},
		{Name: "2 strings", Args: []any{"Hello", "world!"}},
		{Name: "1 string 1 int", Args: []any{"hello", 42}},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := templating.MakeList(c.Args...)
			test.CheckSlice(t, "list", c.Args, got)
		})
	}
}

func TestReversed(t *testing.T) {
	cases := []struct {
		Name string
		Args []any
		Want []any
	}{
		{Name: "empty"},
		{Name: "1 string", Args: []any{"hello"}, Want: []any{"hello"}},
		{
			Name: "2 strings",
			Args: []any{"Hello", "world!"},
			Want: []any{"world!", "Hello"},
		},
		{
			Name: "1 string 1 int",
			Args: []any{"hello", 42},
			Want: []any{42, "hello"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := templating.Reversed(c.Args)
			test.CheckSlice(t, "list", c.Want, got)
		})
	}
}

func TestAppendList(t *testing.T) {
	cases := []struct {
		Name string
		List []any
		Args []any
		Want []any
	}{
		{Name: "empty"},
		{
			Name: "nil + 1 string",
			Args: []any{"hello"},
			Want: []any{"hello"},
		},
		{
			Name: "nil + 2 strings",
			Args: []any{"Hello", "world!"},
			Want: []any{"Hello", "world!"},
		},
		{
			Name: "nil + 1 string 1 int",
			Args: []any{"hello", 42},
			Want: []any{"hello", 42},
		},
		{
			Name: "string + 1 string",
			List: []any{"oh"},
			Args: []any{"hello"},
			Want: []any{"oh", "hello"},
		},
		{
			Name: "string + 2 strings",
			List: []any{"oh"},
			Args: []any{"Hello", "world!"},
			Want: []any{"oh", "Hello", "world!"},
		},
		{
			Name: "string + 1 string 1 int",
			List: []any{"oh"},
			Args: []any{"hello", 42},
			Want: []any{"oh", "hello", 42},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := templating.AppendList(c.List, c.Args...)
			test.CheckSlice(t, "append", c.Want, got)
		})
	}
}
