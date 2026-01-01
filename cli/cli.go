package cli

import (
	"errors"
	"log"
	"log/slog"
	"os"
	"text/template"
	"time"

	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/templating"
)

// ProgName affects the messages we display and the default config file path.
var ProgName = "hebcalfmt"

// InitLogging sets up log and slog for program-wide use.
// By default, we set log to simply output to stderr with no special formatting.
// slog is configured to for JSON output with source line numbers.
//
// If you replace InitLogging, keep in mind that hebcalfmt uses
// log for ordinary errors, likely caused externally.
// slog is for internal errors, likely indicating a bug in hebcalfmt.
var InitLogging = func() {
	slogger := slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{
			AddSource: true,
		},
	))
	slog.SetDefault(slogger)

	// log is for general user errors.
	// This must be set after slog, since slog.SetDefault clobbers these settings.
	log.Default().SetFlags(0)
	log.Default().SetOutput(os.Stderr)
	log.Default().SetPrefix("")
}

// Run is the entry point for CLI program.
// It parses CLI flags and arguments and returns an exit code.
func Run() int {
	InitLogging()

	fs := NewFlags()
	cfg, err := processFlags(fs, os.Args[1:])
	if errors.Is(err, ErrDone) {
		return 0
	}
	if err != nil {
		log.Println(err)
		return 1
	}

	// cfg.Now will be the idea of now for the entire program run.
	// It uses the computer's timezone for our idea of "now",
	// rather than the city's timezone.
	// If a date/time in a different timezone is required,
	// that function should require a timezone argument,
	// rather than rely on the timezone embedded in this variable.
	//
	// NOTE: Even though this system is less consistent logically,
	// and, e.g., a computer in Phoenix will use the date in Phoenix
	// when calculating results for New York where it is already the next day,
	// this program is written for humans.
	// Humans would get confused if, e.g.,
	// results for Jan. 1 next year get generated
	// when for them it is still Dec. 31, and they didn't specify the date:
	//   hebcalfmt examples/hebcalClassic.tmpl
	// For those wanting full consistency, they should specify a timezone
	// in the template or on the CLI. For example:
	//   TZ=America/New_York hebcalfmt examples/hebcalClassic.tmpl
	cfg.Now = time.Now()

	tmplPath, err := processArgs(fs.Args(), cfg)
	if errors.Is(err, ErrUsage) {
		log.Println(usage(fs))
		log.Println(err)
		return 1
	}
	if err != nil {
		log.Println(err)
		return 1
	}

	opts, err := cfg.CalOptions()
	if err != nil {
		log.Printf("failed to build hebcal options from %s: %v",
			cfg.ConfigSource, err)
		return 1
	}

	z := zmanim.New(opts.Location, cfg.Now)

	// Set up the Template's FuncMap.
	// This must be done before parsing the file.
	tmpl := new(template.Template)
	tmpl = templating.SetFuncMap(tmpl, opts)

	tmpl, err = templating.ParseFile(tmpl, tmplPath)
	if err != nil {
		log.Println(err)
		return 1
	}

	err = tmpl.Execute(os.Stdout, map[string]any{
		"now":           cfg.Now,
		"nowInLocation": cfg.Now.In(z.TimeZone),
		"calOptions":    opts,
		"language":      cfg.Language,
		"dateRange":     cfg.DateRange,
		"tz":            z.TimeZone,
		"location":      opts.Location,
		"z":             &z,
		"hdate":         templating.HDateConsts,
		"event":         templating.EventConsts,
		"time":          templating.TimeConsts,
	})
	if err != nil {
		log.Println(err)
		return 1
	}

	return 0
}
