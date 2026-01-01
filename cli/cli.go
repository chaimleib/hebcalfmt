package cli

import (
	"errors"
	"log"
	"log/slog"
	"os"
	"text/template"

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
