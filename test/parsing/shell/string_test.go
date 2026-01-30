package shell_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/shell"
)

func TestParseShellStringDquote(t *testing.T) {
	type Case struct {
		Name     string
		Line     parsing.LineInfo
		Rest     string
		Want     string
		WantRest string
		Err      string
	}
	cases := []Case{
		{Name: "empty", Err: "no match"},
		{
			Name: "missing quotes",
			Line: parsing.LineInfo{
				Line:     "noquotes",
				Number:   1,
				FileName: "missing.sh",
			},
			Rest:     "noquotes",
			WantRest: "noquotes",
			Err:      "no match",
		},
		{
			Name: "ok",
			Line: parsing.LineInfo{
				Line:     `"ok"`,
				Number:   1,
				FileName: "ok.sh",
			},
			Rest: `"ok"`,
			Want: "ok",
		},
		{
			Name: "missing end quote",
			Line: parsing.LineInfo{
				Line:     `"missing end`,
				Number:   1,
				FileName: "missing-end.sh",
			},
			Rest:     `"missing end`,
			WantRest: `"missing end`,
			Err: `syntax at missing-end.sh:1:2-13: expected ending '"' in this span

	"missing end
	 ^^^^^^^^^^^`,
		},
		{
			Name: "ok escaped quote",
			Line: parsing.LineInfo{
				Line:     `"ok \""`,
				Number:   1,
				FileName: "ok-quote.sh",
			},
			Rest: `"ok \""`,
			Want: "ok \"",
		},
		{
			Name: "ok surrounded escaped quote",
			Line: parsing.LineInfo{
				Line:     `"ok \" quotes"`,
				Number:   1,
				FileName: "ok-quotes.sh",
			},
			Rest: `"ok \" quotes"`,
			Want: "ok \" quotes",
		},
		{
			Name: "missing escaped char",
			Line: parsing.LineInfo{
				Line:     `"missing escape \`,
				Number:   1,
				FileName: "missing-escape.sh",
			},
			Rest:     `"missing escape \`,
			WantRest: `"missing escape \`,
			Err: `syntax at missing-escape.sh:1:18: unexpected end after escape char

	"missing escape \
	                 ^`,
		},
		{
			Name: "multibyte unicode",
			Line: parsing.LineInfo{
				Line:     `"🤩"`,
				Number:   1,
				FileName: "multibyte-unicode.sh",
			},
			Rest: `"🤩"`,
			Want: "🤩",
		},
		{
			Name: "invalid 1-byte unicode",
			Line: parsing.LineInfo{
				Line:     "\"\x80 is invalid unicode\"",
				Number:   1,
				FileName: "invalid-unicode.sh",
			},
			Rest:     "\"\x80 is invalid unicode\"",
			WantRest: "\"\x80 is invalid unicode\"",
			Err: `syntax at invalid-unicode.sh:1:2: invalid unicode

	"� is invalid unicode"
	 ^                    `,
		},
		{
			Name: "invalid 2-byte unicode",
			Line: parsing.LineInfo{
				Line:     "\"\xc0\x80 is invalid unicode\"",
				Number:   1,
				FileName: "invalid-unicode.sh",
			},
			Rest:     "\"\xc0\x80 is invalid unicode\"",
			WantRest: "\"\xc0\x80 is invalid unicode\"",
			Err: `syntax at invalid-unicode.sh:1:2: invalid unicode

	"�� is invalid unicode"
	 ^                     `,
		},
		{
			Name: "invalid escaped 1-byte unicode",
			Line: parsing.LineInfo{
				Line:     "\"\\\x80 is invalid unicode\"",
				Number:   1,
				FileName: "invalid-unicode-escape.sh",
			},
			Rest:     "\"\\\x80 is invalid unicode\"",
			WantRest: "\"\\\x80 is invalid unicode\"",
			Err: `syntax at invalid-unicode-escape.sh:1:3: invalid unicode

	"\� is invalid unicode"
	  ^                    `,
		},
		{
			Name: "escape nl",
			Line: parsing.LineInfo{
				Line:     `"escape \n"`,
				Number:   1,
				FileName: "escape-nl.sh",
			},
			Rest: `"escape \n"`,
			Want: "escape \n",
		},
		{
			Name: "escape cr",
			Line: parsing.LineInfo{
				Line:     `"escape \r"`,
				Number:   1,
				FileName: "escape-cr.sh",
			},
			Rest: `"escape \r"`,
			Want: "escape \r",
		},
		{
			Name: "escape tab",
			Line: parsing.LineInfo{
				Line:     `"escape \t"`,
				Number:   1,
				FileName: "escape-tab.sh",
			},
			Rest: `"escape \t"`,
			Want: "escape \t",
		},
		{
			Name: "missing quote after escaped quote",
			Line: parsing.LineInfo{
				Line:     `"missing close\"`,
				Number:   1,
				FileName: "missing-close.sh",
			},
			Rest:     `"missing close\"`,
			WantRest: `"missing close\"`,
			Err: `syntax at missing-close.sh:1:17: expected ending '"' for double-quoted shell string

	"missing close\"
	                ^`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, rest, err := shell.ParseShellStringDquote(
				c.Line,
				[]byte(c.Rest),
			)
			test.CheckComparable(t, "value", c.Want, got)
			test.CheckComparable(t, "rest", string(c.WantRest), string(rest))
			test.CheckErr(t, err, c.Err)
		})
	}
}

func TestParseShellStringSquote(t *testing.T) {
	type Case struct {
		Name     string
		Line     parsing.LineInfo
		Rest     string
		Want     string
		WantRest string
		Err      string
	}
	cases := []Case{
		{Name: "empty", Err: "no match"},
		{
			Name: "missing quotes",
			Line: parsing.LineInfo{
				Line:     "noquotes",
				Number:   1,
				FileName: "missing.sh",
			},
			Rest:     "noquotes",
			WantRest: "noquotes",
			Err:      "no match",
		},
		{
			Name: "ok",
			Line: parsing.LineInfo{
				Line:     `'ok'`,
				Number:   1,
				FileName: "ok.sh",
			},
			Rest: `'ok'`,
			Want: "ok",
		},
		{
			Name: "missing end quote",
			Line: parsing.LineInfo{
				Line:     `'missing end`,
				Number:   1,
				FileName: "missing-end.sh",
			},
			Rest:     `'missing end`,
			WantRest: `'missing end`,
			Err: `syntax at missing-end.sh:1:2-13: expected ending ` + "`'`" + ` in this span

	'missing end
	 ^^^^^^^^^^^`,
		},
		{
			Name: "ok literal backslash",
			Line: parsing.LineInfo{
				Line:     `'ok \'`,
				Number:   1,
				FileName: "ok-literal.sh",
			},
			Rest: `'ok \'`,
			Want: `ok \`,
		},
		{
			Name: "ok surrounded literal backslash",
			Line: parsing.LineInfo{
				Line:     `'ok \ literal'`,
				Number:   1,
				FileName: "ok-surrounded-literal.sh",
			},
			Rest: `'ok \ literal'`,
			Want: `ok \ literal`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, rest, err := shell.ParseShellStringSquote(
				c.Line,
				[]byte(c.Rest),
			)
			test.CheckComparable(t, "value", c.Want, got)
			test.CheckComparable(t, "rest", string(c.WantRest), string(rest))
			test.CheckErr(t, err, c.Err)
		})
	}
}

func TestParseShellStringRaw(t *testing.T) {
	type Case struct {
		Name     string
		Line     parsing.LineInfo
		Rest     string
		Want     string
		WantRest string
		Err      string
	}
	cases := []Case{
		{Name: "empty", Err: "no match"},
		{
			Name: "unexpected quotes",
			Line: parsing.LineInfo{
				Line:     "'unexpected quote",
				Number:   1,
				FileName: "unexpected-quote.sh",
			},
			Rest:     "'unexpected quote",
			WantRest: "'unexpected quote",
			Err:      "no match",
		},
		{
			Name: "ok",
			Line: parsing.LineInfo{
				Line:     `ok`,
				Number:   1,
				FileName: "ok.sh",
			},
			Rest: `ok`,
			Want: "ok",
		},
		{
			Name: "ok before space",
			Line: parsing.LineInfo{
				Line:     `ok ignored`,
				Number:   1,
				FileName: "ok-space.sh",
			},
			Rest:     `ok ignored`,
			Want:     `ok`,
			WantRest: ` ignored`,
		},
		{
			Name: "ok escaped space",
			Line: parsing.LineInfo{
				Line:     `ok\ good`,
				Number:   1,
				FileName: "ok-esc-space.sh",
			},
			Rest: `ok\ good`,
			Want: `ok good`,
		},
		{
			Name: "ok trailing escape",
			Line: parsing.LineInfo{
				Line:     `ok\<`,
				Number:   1,
				FileName: "ok-esc-literal.sh",
			},
			Rest: `ok\<`,
			Want: `ok<`,
		},
		{
			Name: "ok surrounded literal backslash",
			Line: parsing.LineInfo{
				Line:     `ok\\literal`,
				Number:   1,
				FileName: "ok-surrounded-literal.sh",
			},
			Rest: `ok\\literal`,
			Want: `ok\literal`,
		},
		{
			Name: "ok flag",
			Line: parsing.LineInfo{
				Line:     `-v`,
				Number:   1,
				FileName: "ok-flag.sh",
			},
			Rest: `-v`,
			Want: "-v",
		},
		{
			Name: "ok option",
			Line: parsing.LineInfo{
				Line:     `--color=auto`,
				Number:   1,
				FileName: "ok-option.sh",
			},
			Rest: `--color=auto`,
			Want: "--color=auto",
		},
		{
			Name: "missing escape",
			Line: parsing.LineInfo{
				Line:     `missingEscape\`,
				Number:   1,
				FileName: "missing-esc.sh",
			},
			Rest:     `missingEscape\`,
			WantRest: `missingEscape\`,
			Err: `syntax at missing-esc.sh:1:15: unexpected end of raw string after escape

	missingEscape\
	              ^`,
		},
		{
			Name: "multibyte unicode",
			Line: parsing.LineInfo{
				Line:     `🤩`,
				Number:   1,
				FileName: "multibyte-unicode.sh",
			},
			Rest: `🤩`,
			Want: "🤩",
		},
		{
			Name: "invalid 1-byte unicode",
			Line: parsing.LineInfo{
				Line:     "\x80 is invalid unicode",
				Number:   1,
				FileName: "invalid-unicode.sh",
			},
			Rest:     "\x80 is invalid unicode",
			WantRest: "\x80 is invalid unicode",
			Err: `syntax at invalid-unicode.sh:1:1: invalid unicode in raw string

	� is invalid unicode
	^                   `,
		},
		{
			Name: "invalid 2-byte unicode",
			Line: parsing.LineInfo{
				Line:     "\xc0\x80 is invalid unicode",
				Number:   1,
				FileName: "invalid-unicode.sh",
			},
			Rest:     "\xc0\x80 is invalid unicode",
			WantRest: "\xc0\x80 is invalid unicode",
			Err: `syntax at invalid-unicode.sh:1:1: invalid unicode in raw string

	�� is invalid unicode
	^                    `,
		},
		{
			Name: "invalid escaped 1-byte unicode",
			Line: parsing.LineInfo{
				Line:     "\\\x80 is invalid unicode",
				Number:   1,
				FileName: "invalid-unicode-escape.sh",
			},
			Rest:     "\\\x80 is invalid unicode",
			WantRest: "\\\x80 is invalid unicode",
			Err: `syntax at invalid-unicode-escape.sh:1:2: invalid unicode after escape in raw string

	\� is invalid unicode
	 ^                   `,
		},
		{
			Name: "todo backtick",
			Line: parsing.LineInfo{
				Line:     "`",
				Number:   1,
				FileName: "todo-backtick.sh",
			},
			Rest:     "`",
			WantRest: "`",
			Err:      "no match",
		},
		{
			Name: "todo history",
			Line: parsing.LineInfo{
				Line:     "!-1",
				Number:   1,
				FileName: "todo-history.sh",
			},
			Rest:     "!-1",
			WantRest: "!-1",
			Err:      "no match",
		},
		{
			Name: "todo subshell",
			Line: parsing.LineInfo{
				Line:     "$(echo a)",
				Number:   1,
				FileName: "todo-subshell.sh",
			},
			Rest:     "$(echo a)",
			WantRest: "$(echo a)",
			Err:      "no match",
		},
		{
			Name: "todo glob",
			Line: parsing.LineInfo{
				Line:     "*glob",
				Number:   1,
				FileName: "todo-glob.sh",
			},
			Rest:     "*glob",
			WantRest: "*glob",
			Err:      "no match",
		},
		{
			Name: "var",
			Line: parsing.LineInfo{
				Line:     "var$name",
				Number:   1,
				FileName: "todo-var.sh",
			},
			Rest:     "var$name",
			Want:     "var",
			WantRest: "$name",
		},
		{
			Name: "unexpected close bracket",
			Line: parsing.LineInfo{
				Line:     "unpaired]",
				Number:   1,
				FileName: "unexpected-close.sh",
			},
			Rest:     "unpaired]",
			WantRest: "]",
			Want:     "unpaired",
		},
		{
			Name: "redirection",
			Line: parsing.LineInfo{
				Line:     "input>output",
				Number:   1,
				FileName: "redirection.sh",
			},
			Rest:     "input>output",
			Want:     "input",
			WantRest: ">output",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, rest, err := shell.ParseShellStringRaw(
				c.Line,
				[]byte(c.Rest),
			)
			test.CheckComparable(t, "value", c.Want, got)
			test.CheckComparable(t, "rest", string(c.WantRest), string(rest))
			test.CheckErr(t, err, c.Err)
		})
	}
}

func TestParseShellString(t *testing.T) {
	type Case struct {
		Name     string
		Line     parsing.LineInfo
		Rest     string
		Want     string
		WantRest string
		Err      string
	}
	cases := []Case{
		{Name: "empty", Err: "no match"},
		{
			Name: "raw",
			Line: parsing.LineInfo{
				Line:     "raw",
				Number:   1,
				FileName: "raw.sh",
			},
			Rest: "raw",
			Want: "raw",
		},
		{
			Name: "raw trailing paren",
			Line: parsing.LineInfo{
				Line:     "raw)",
				Number:   1,
				FileName: "raw-trail-paren.sh",
			},
			Rest:     "raw)",
			Want:     "raw",
			WantRest: ")",
		},
		{
			Name: "squote",
			Line: parsing.LineInfo{
				Line:     "'squote'",
				Number:   1,
				FileName: "squote.sh",
			},
			Rest: "'squote'",
			Want: "squote",
		},
		{
			Name: "dquote",
			Line: parsing.LineInfo{
				Line:     `"'dquote'"`,
				Number:   1,
				FileName: "dquote.sh",
			},
			Rest: `"'dquote'"`,
			Want: `'dquote'`,
		},
		{
			Name: "combo",
			Line: parsing.LineInfo{
				Line:     `"'dquote'"'squote'raw`,
				Number:   1,
				FileName: "combo.sh",
			},
			Rest: `"'dquote'"'squote'raw`,
			Want: `'dquote'squoteraw`,
		},
		{
			Name: "ok flag",
			Line: parsing.LineInfo{
				Line:     `-v`,
				Number:   1,
				FileName: "ok-flag.sh",
			},
			Rest: `-v`,
			Want: "-v",
		},
		{
			Name: "ok option",
			Line: parsing.LineInfo{
				Line:     `--color="auto"`,
				Number:   1,
				FileName: "ok-option.sh",
			},
			Rest: `--color="auto"`,
			Want: "--color=auto",
		},
		{
			Name: "raw with close curly",
			Line: parsing.LineInfo{
				Line:     `raw}curly`,
				Number:   1,
				FileName: "raw-curly.sh",
			},
			Rest:     `raw}curly`,
			Want:     "raw",
			WantRest: `}curly`,
		},
		{
			Name: "raw then space",
			Line: parsing.LineInfo{
				Line:     "raw space",
				Number:   1,
				FileName: "raw-space.sh",
			},
			Rest:     "raw space",
			Want:     "raw",
			WantRest: " space",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, rest, err := shell.ParseShellString(
				c.Line,
				[]byte(c.Rest),
			)
			test.CheckComparable(t, "value", c.Want, got)
			test.CheckComparable(t, "rest", string(c.WantRest), string(rest))
			test.CheckErr(t, err, c.Err)
		})
	}
}

func TestFormatString(t *testing.T) {
	cases := []struct {
		Name  string // defaults to Input
		Input string
		Want  string
	}{
		{Name: "empty", Input: "", Want: `""`},
		{Name: "spaces", Input: "  ", Want: `"  "`},
		{Input: "value", Want: "value"},
		{Input: "spaced value", Want: `"spaced value"`},
		{Input: "special[chars]", Want: `"special[chars]"`},
		{Input: "special$dollar", Want: `"special\$dollar"`},
		{Input: "special\nnewline", Want: `"special\nnewline"`},
		{Input: "special\rcarriageReturn", Want: `"special\rcarriageReturn"`},
		{Input: "special\ttab", Want: `"special\ttab"`},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			name = c.Input
		}
		t.Run(name, func(t *testing.T) {
			got := shell.FormatString(c.Input)
			test.CheckComparable(t, "value", c.Want, got)
		})
	}
}

func TestDquoteString(t *testing.T) {
	cases := []struct {
		Name  string // defaults to Input
		Input string
		Want  string
	}{
		{Name: "empty", Input: "", Want: `""`},
		{Name: "spaces", Input: "  ", Want: `"  "`},
		{Input: "value", Want: `"value"`},
		{Input: "spaced value", Want: `"spaced value"`},
		{Input: "special[chars]", Want: `"special[chars]"`},
		{Input: "special$dollar", Want: `"special\$dollar"`},
		{Input: "special\nnewline", Want: `"special\nnewline"`},
		{Input: "special\rcarriageReturn", Want: `"special\rcarriageReturn"`},
		{Input: "special\ttab", Want: `"special\ttab"`},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			name = c.Input
		}
		t.Run(name, func(t *testing.T) {
			got := shell.DquoteString(c.Input)
			test.CheckComparable(t, "value", c.Want, got)
		})
	}
}
