package shell_test

import (
	"bytes"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/shell"
)

func TestParseInlineFile(t *testing.T) {
	type Case struct {
		Name        string
		Rest        string
		WantRest    string
		Want        []shell.Closure
		Err         string
		WantCode    shell.Code
		WantContent string
		WantRunErr  string
	}
	cases := []Case{
		{Name: "empty", Err: "no match"},
		{
			Name:     "empty cmd",
			Rest:     "<()",
			WantRest: "<()",
			Err: `syntax at inline-file.sh:1:1-3: inline file with no commands

	<()
	^^^`,
		},
		{
			Name:     "unknown cmd",
			Rest:     "<(bc calc.bc)",
			WantCode: shell.CodeCommandNotFound,
			WantRunErr: `syntax at inline-file.sh:1:3: unknown command: "bc" - exited with code 127

	<(bc calc.bc)
	  ^          `,
		},
		{
			Name:        "naked echo",
			Rest:        "<(echo)",
			WantContent: "\n",
		},
		{
			Name:        "echo trailing space",
			Rest:        "<(echo )",
			WantContent: "\n",
		},
		{
			Name:        "echo",
			Rest:        "<(echo ok)",
			WantContent: "ok\n",
		},
		{
			Name:        "echo 2 args",
			Rest:        "<(echo 1  2)",
			WantContent: "1 2\n",
		},
		{
			Name:        "echo json",
			Rest:        `<(echo '{"status": "ok"}')`,
			WantContent: `{"status": "ok"}` + "\n",
		},
		{
			Name:        "echo twice",
			Rest:        "<(echo ab; \techo cd)",
			WantContent: "ab\ncd\n",
		},
		{
			Name:        "optional semicolon",
			Rest:        "<(echo ab;)",
			WantContent: "ab\n",
		},
		{
			Name:     "unclosed paren",
			Rest:     "<(echo ab",
			WantRest: "<(echo ab",
			Err: `syntax at inline-file.sh:1:10: unexpected end when building inline file, expected ')'

	<(echo ab
	         ^`,
		},
		{
			Name:     "unclosed paren after semicolon",
			Rest:     "<(echo ab;",
			WantRest: "<(echo ab;",
			Err: `syntax at inline-file.sh:1:11: unexpected end when building inline file, expected ')'

	<(echo ab;
	          ^`,
		},
		{
			Name:     "extra close bracket",
			Rest:     "<(echo]",
			WantRest: "<(echo]",
			Err: `syntax at inline-file.sh:1:7: expected command termination (e.g. ';') between commands when building inline file

	<(echo]
	      ^`,
		},
		{
			Name:     "invalid command",
			Rest:     "<( ] )",
			WantRest: "<( ] )",
			Err: `syntax at inline-file.sh:1:4: expected a command

	<( ] )
	   ^  `,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			li := parsing.LineInfo{
				Line:     []byte(c.Rest),
				FileName: "inline-file.sh",
				Number:   1,
			}
			subProg, rest, err := shell.ParseInlineFile(li, []byte(c.Rest))
			test.CheckErr(t, err, c.Err)
			test.CheckString(t, "rest", c.WantRest, string(rest))
			if len(subProg) == 0 {
				if len(c.WantContent) > 0 {
					t.Errorf(
						"wanted file content, but there are no commands to write it:\n%q",
						c.WantContent,
					)
				}
			} else {
				var buf bytes.Buffer
				inlineFile := &shell.InlineFile{
					SubProg: subProg,
				}
				code, err := inlineFile.Run(&buf, &buf)
				test.CheckComparable(t, "return code", c.WantCode, code)
				test.CheckErr(t, err, c.WantRunErr)
			}
		})
	}
}
