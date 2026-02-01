package cli_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/chaimleib/hebcalfmt/cli"
	"github.com/chaimleib/hebcalfmt/config"
	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestRunInEnvironment(t *testing.T) {
	fdata := func(s string) *fstest.MapFile {
		return &fstest.MapFile{Data: []byte(s)}
	}
	files := fstest.MapFS{
		"date.tmpl":            fdata(`{{$.dateRange.StartOrToday false}}`),
		"executeError.tmpl":    fdata(`{{printf $.tz "INVALID FORMAT"}}`),
		"invalid.json":         fdata(`{INVALID JSON`),
		"invalid.tmpl":         fdata(`{{INVALID`),
		"invalidCity.json":     fdata(`{"city": "Invalid City"}`),
		"invalidLanguage.json": fdata(`{"language": "Invalid Language"}`),
		"stub.tmpl":            fdata(`ok`),
		"today.json":           fdata(`{"today": true}`),
	}

	now := time.Date(2025, 12, 21, 0, 0, 0, 0, time.UTC)
	usagePrefix := fmt.Sprintf("usage:\n  %s ", cli.ProgName)

	cases := []struct {
		Args        string
		FS          fs.FS
		Want        string
		WantMode    test.WantMode
		WantLog     string
		WantLogMode test.WantMode
		Err         string
	}{
		{
			Args:        "",
			WantLog:     usagePrefix,
			WantLogMode: test.WantPrefix,
			Err:         "usage error: missing a template file argument",
		},
		{
			Args:     "-h",
			Want:     usagePrefix,
			WantMode: test.WantPrefix,
		},
		{
			Args:     "--help",
			Want:     usagePrefix,
			WantMode: test.WantPrefix,
		},
		{
			Args:        "--invalid-flag",
			WantLog:     usagePrefix,
			WantLogMode: test.WantPrefix,
			Err:         "usage error: unknown flag: --invalid-flag",
		},
		{
			Args: "--version",
			Want: fmt.Sprintf("%s %s\n", cli.ProgName, cli.Version),
		},
		{
			Args:        "--info",
			WantLog:     usagePrefix,
			WantLogMode: test.WantPrefix,
			Err:         "usage error: flag needs an argument: --info",
		},
		{
			Args:        "-i",
			WantLog:     usagePrefix,
			WantLogMode: test.WantPrefix,
			Err:         "usage error: flag needs an argument: 'i' in -i",
		},
		{
			Args: "--info default-city",
			Want: config.DefaultCity + "\n",
		},
		{
			Args: "--info=default-city",
			Want: config.DefaultCity + "\n",
		},
		{
			Args: "-i default-city",
			Want: config.DefaultCity + "\n",
		},
		{
			Args:     "--info cities",
			Want:     "\n" + config.DefaultCity + "\n",
			WantMode: test.WantContains,
		},
		{
			Args:     "--info languages",
			Want:     "\nashkenazi_standard\n",
			WantMode: test.WantContains,
		},
		{
			Args: "--info INVALID_INFO",
			WantLog: fmt.Sprintf(
				`unrecognized key for --info flag: "INVALID_INFO"
Available options: %q
%s`,
				cli.InfoKeys,
				usagePrefix,
			),
			WantLogMode: test.WantPrefix,
			Err:         `unrecognized key for --info flag: "INVALID_INFO"`,
		},
		{
			Args: "--config invalid.json stub.tmpl",
			Err:  `failed to load config: failed to parse config from "invalid.json": invalid character 'I' looking for beginning of object key string`,
		},
		{
			Args: "--config invalidCity.json stub.tmpl",
			WantLog: `unknown city: "Invalid City"
Use a nearby city; or add geo.lat, geo.lon, and timezone.
To show available cities, run:
  hebcalfmt --info cities
`,
			Err: `failed to build hebcal options from invalidCity.json: failed to resolve place configs: unknown city: "Invalid City"`,
		},
		{
			Args: "--config invalidLanguage.json stub.tmpl",
			WantLog: fmt.Sprintf(
				`unknown language: "Invalid Language"
To show the available languages, run
  %s --info languages
`,
				cli.ProgName,
			),
			Err: `unknown language: "Invalid Language"`,
		},
		{Args: "date.tmpl", Want: "1 Tevet 5786"},
		{Args: "date.tmpl 2024", Want: "20 Tevet 5784"},
		{Args: "date.tmpl 3 2024", Want: "21 Adar I 5784"},
		{Args: "date.tmpl 3 2 2024", Want: "22 Adar I 5784"},
		{
			Args:        "stub.tmpl 3 2 INVALIDYEAR",
			WantLog:     usagePrefix,
			WantLogMode: test.WantPrefix,
			Err:         `usage error: invalid year: strconv.Atoi: parsing "INVALIDYEAR": invalid syntax`,
		},
		{
			Args: "invalid.tmpl",
			Err:  "template: invalid.tmpl:1: function \"INVALID\" not defined",
		},
		{
			Args: "executeError.tmpl",
			Err:  `template: executeError.tmpl:1:10: executing "executeError.tmpl" at <$.tz>: wrong type for value; expected string; got *time.Location`,
		},
		{
			Args: "--config today.json date.tmpl",
			Want: "1 Tevet 5786",
		},
	}
	for _, c := range cases {
		t.Run(c.Args, func(t *testing.T) {
			var args []string
			if c.Args != "" {
				args = strings.Fields(c.Args)
			}
			var buf bytes.Buffer
			logBuf := test.Logger(t)
			err := cli.RunInEnvironment(args, files, now, templating.BuildData, &buf)
			test.CheckErr(t, err, c.Err)

			test.CheckStringMode(t, "output", c.Want, buf.String(), c.WantMode)
			test.CheckStringMode(t, "logs", c.WantLog, logBuf.String(), c.WantLogMode)
		})
	}
}
