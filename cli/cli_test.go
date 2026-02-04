package cli_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/chaimleib/hebcalfmt/cli"
	"github.com/chaimleib/hebcalfmt/config"
	"github.com/chaimleib/hebcalfmt/fsys"
	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestInitLogging(t *testing.T) {
	// exercise the function, make sure it doesn't panic.
	cli.InitLogging()
}

func setLogger(t *testing.T) fmt.Stringer {
	orig := cli.InitLogging
	t.Cleanup(func() {
		cli.InitLogging = orig
	})

	var buf bytes.Buffer
	cli.InitLogging = func() {
		slogger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(slogger)

		log.Default().SetFlags(0)
		log.Default().SetOutput(&buf)
		log.Default().SetPrefix("")
	}

	return &buf
}

func setArgs(t *testing.T, args string) {
	oldArgs := os.Args
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	argFields := strings.Fields(args)
	os.Args = append(os.Args[:1], argFields...)
}

func failFS() (fs.FS, error) { return nil, errors.New("test: failFS") }

func TestRun(t *testing.T) {
	cases := []struct {
		Name      string
		Args      string // split with strings.Fields()
		DefaultFS func() (fs.FS, error)
		Want      int
		Logs      string
	}{
		{Name: "empty input template", Args: "/dev/null"},
		{
			Name: "nonexistent template path",
			Args: "does-not-exist.tmpl",
			Want: 1,
			Logs: "open does-not-exist.tmpl: no such file or directory\n",
		},
		{
			Name:      "failFS",
			Want:      1,
			DefaultFS: failFS,
			Logs:      "time=... level=ERROR msg=\"failed to initialize DefaultFS\" error=\"test: failFS\"\n",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			logBuf := setLogger(t)
			setArgs(t, c.Args)
			if c.DefaultFS != nil {
				orig := fsys.DefaultFS
				fsys.DefaultFS = c.DefaultFS
				t.Cleanup(func() { fsys.DefaultFS = orig })
			}

			got := cli.Run()
			if c.Want != got {
				t.Errorf("Run exited with code %d, want %d", got, c.Want)
			}
			test.CheckEllipsis(t, "logs", c.Logs, logBuf.String())
		})
	}
}

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
