package shell_test

import (
	"testing"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/shell"
)

func TestParseInlineFile(t *testing.T) {
	type Case struct {
		Name         string
		Rest         string
		Line         parsing.LineInfo
		FileID       int
		WantFileName string
		WantContent  string
		WantRest     string
		Err          string
	}
	cases := []Case{
		{Name: "empty", Err: "no match"},
		{
			Name: "empty cmd",
			Rest: "<()",
			Line: parsing.LineInfo{
				Line:     "<()",
				Number:   1,
				FileName: "empty-cmd.sh",
			},
			WantRest: "<()",
			Err: `syntax at empty-cmd.sh:1:1-3: inline file with no commands

	<()
	^^^`,
		},
		{
			Name: "unknown cmd",
			Rest: "<(bc calc.bc)",
			Line: parsing.LineInfo{
				Line:     "<(bc calc.bc)",
				Number:   1,
				FileName: "unknown-cmd.sh",
			},
			WantRest: "<(bc calc.bc)",
			Err: `syntax at unknown-cmd.sh:1:3: unknown command: "bc" - exited with code 127

	<(bc calc.bc)
	  ^          `,
		},
		{
			Name: "naked echo",
			Rest: "<(echo)",
			Line: parsing.LineInfo{
				Line:     "<(echo)",
				Number:   1,
				FileName: "naked-echo.sh",
			},
			WantFileName: "tmp/inlineFile00",
			WantContent:  "\n",
		},
		{
			Name:   "echo trailing space",
			Rest:   "<(echo )",
			FileID: 10,
			Line: parsing.LineInfo{
				Line:     "<(echo )",
				Number:   1,
				FileName: "echo-trail-space.sh",
			},
			WantFileName: "tmp/inlineFile10",
			WantContent:  "\n",
		},
		{
			Name: "echo",
			Rest: "<(echo ok)",
			Line: parsing.LineInfo{
				Line:     "<(echo ok)",
				Number:   1,
				FileName: "echo.sh",
			},
			WantFileName: "tmp/inlineFile00",
			WantContent:  "ok\n",
		},
		{
			Name: "echo 2 args",
			Rest: "<(echo 1  2)",
			Line: parsing.LineInfo{
				Line:     "<(echo 1  2)",
				Number:   1,
				FileName: "echo-2.sh",
			},
			WantFileName: "tmp/inlineFile00",
			WantContent:  "1 2\n",
		},
		{
			Name: "echo json",
			Rest: `<(echo '{"status": "ok"}')`,
			Line: parsing.LineInfo{
				Line:     `<(echo '{"status": "ok"}')`,
				Number:   1,
				FileName: "echo.sh",
			},
			WantFileName: "tmp/inlineFile00",
			WantContent:  `{"status": "ok"}` + "\n",
		},
		{
			Name: "echo twice",
			Rest: "<(echo ab; \techo cd)",
			Line: parsing.LineInfo{
				Line:     "<(echo ab; \techo cd)",
				Number:   1,
				FileName: "echo-twice.sh",
			},
			WantFileName: "tmp/inlineFile00",
			WantContent:  "ab\ncd\n",
		},
		{
			Name: "optional semicolon",
			Rest: "<(echo ab;)",
			Line: parsing.LineInfo{
				Line:     "<(echo ab;)",
				Number:   1,
				FileName: "echo-optional-semi.sh",
			},
			WantContent:  "ab\n",
			WantFileName: "tmp/inlineFile00",
		},
		{
			Name: "unclosed paren",
			Rest: "<(echo ab",
			Line: parsing.LineInfo{
				Line:     "<(echo ab",
				Number:   1,
				FileName: "echo-unclosed.sh",
			},
			WantRest: "<(echo ab",
			Err: `syntax at echo-unclosed.sh:1:10: unexpected end when building inline file, expected ')'

	<(echo ab
	         ^`,
		},
		{
			Name: "unclosed paren after semicolon",
			Rest: "<(echo ab;",
			Line: parsing.LineInfo{
				Line:     "<(echo ab;",
				Number:   1,
				FileName: "echo-unclosed-semi.sh",
			},
			WantRest: "<(echo ab;",
			Err: `syntax at echo-unclosed-semi.sh:1:11: unexpected end when building inline file, expected ')'

	<(echo ab;
	          ^`,
		},
		{
			Name: "extra close bracket",
			Rest: "<(echo]",
			Line: parsing.LineInfo{
				Line:     "<(echo]",
				Number:   1,
				FileName: "extra-close.sh",
			},
			WantRest: "<(echo]",
			Err: `syntax at extra-close.sh:1:7: expected command termination (e.g. ';') between commands when building inline file

	<(echo]
	      ^`,
		},
		{
			Name: "invalid command",
			Rest: "<( ] )",
			Line: parsing.LineInfo{
				Line:     "<( ] )",
				Number:   1,
				FileName: "invalid-command.sh",
			},
			WantRest: "<( ] )",
			Err: `syntax at invalid-command.sh:1:4: expected a command

	<( ] )
	   ^  `,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			files := make(fstest.MapFS)
			fname, rest, err := shell.ParseInlineFile(
				c.Line, []byte(c.Rest), c.FileID, files)
			test.CheckErr(t, err, c.Err)
			test.CheckComparable(t, "fname", c.WantFileName, fname)
			test.CheckComparable(t, "rest", c.WantRest, string(rest))

			if err == nil {
				fileContent, err := files.ReadFile(fname)
				test.CheckErr(t, err, "")
				test.CheckComparable(t, "fileData", c.WantContent, string(fileContent))
			}
		})
	}
}
