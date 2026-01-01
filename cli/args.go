package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/pflag"

	"github.com/chaimleib/hebcalfmt/config"
	"github.com/chaimleib/hebcalfmt/daterange"
)

var (
	// ErrDone indicates that the program should end early with a 0 exit code.
	// This can happen if the help message, version message,
	// or a list operation was requested.
	ErrDone = errors.New("done")

	// ErrUnreachable means that there is a coding defect.
	ErrUnreachable = errors.New("unreachable")

	// ErrUsage means that CLI arguments were invalid,
	// and the usage message should be displayed.
	ErrUsage = errors.New("usage error")
)

// NewFlags returns a [pflag.FlagSet] configured with the flags
// used by hebcalfmt.
func NewFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet(ProgName, pflag.ContinueOnError)
	// opt.SetParameters("[[ month [ day ]] year]")

	fs.BoolP("help", "h", false,
		"print this help text")
	fs.Bool("version", false,
		"show version number")
	fs.StringP("config", "c", "",
		"select a JSON config file (default $HOME/.config/hebcalfmt/config.json)")
	fs.StringP(
		"info",
		"i",
		"",
		"show data from the internal databases or compiled values. Available options: cities, default-city, languages",
	)

	return fs
}

// processFlags produces a [config.Config],
// using just the hyphenated options and flags in args.
// NOTE: Other args like the template file path and the date range spec
// are processed by processArgs.
func processFlags(
	fs *pflag.FlagSet,
	args []string,
) (*config.Config, error) {
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// pflag would return an error if
	// - the flag was never defined (i.e. in [NewFlags]),
	// - the flag type and the requested type don't match,
	// - the string value could not be decoded.
	// All of these would indicate coding defects,
	// in which case I want to know the source file and line number.
	// slog is set up for this purpose.
	help, err := fs.GetBool("help")
	if err != nil {
		slog.Error("failed to get --help flag", "error", err)
		return nil, fmt.
			Errorf("%w: get --help: %w", ErrUnreachable, err)
	}
	if help {
		fmt.Println(usage(fs))
		return nil, ErrDone
	}

	version, err := fs.GetBool("version")
	if err != nil {
		slog.Error("failed to get --version flag", "error", err)
		return nil, fmt.Errorf("%w: get --version: %w", ErrUnreachable, err)
	}
	if version {
		fmt.Println(versionMessage())
		return nil, ErrDone
	}

	key, err := fs.GetString("info")
	if err != nil {
		slog.Error("failed to get --info option", "error", err)
		return nil, fmt.Errorf("%w: get --info: %w", ErrUnreachable, err)
	}
	if key != "" {
		info, err := infoString(key, fs)
		if err != nil {
			return nil, err
		}
		fmt.Println(info)
		return nil, ErrDone
	}

	return loadConfigFromFlags(fs)
}

// loadConfigFromFlags reads the --config flag option
// and loads the config file specified.
// Otherwise, it loads the default config.
// Then it calls Normalize on the result.
func loadConfigFromFlags(fs *pflag.FlagSet) (*config.Config, error) {
	fpath, err := fs.GetString("config")
	if err != nil {
		slog.Error("failed to get --config option", "error", err)
		return nil, fmt.Errorf("%w: get --config: %w", ErrUnreachable, err)
	}

	cfg, err := loadConfigOrDefault(fpath)
	if err != nil {
		return nil, err
	}

	cfg, err = cfg.Normalize()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadConfigOrDefault reads fpath if not empty.
// It will error if fpath does not exist.
//
// If fpath is empty, default to loading from ~/.config/hebcalfmt/config.json,
// but if that file is not present, return config.Default with no error.
func loadConfigOrDefault(fpath string) (*config.Config, error) {
	var suppressMissingConfigErr bool

	// Try to configure a default configPath if one was not provided
	if fpath == "" {
		suppressMissingConfigErr = true
		home := os.Getenv("HOME")
		if home == "" {
			return new(config.Config), nil
		}
		fpath = filepath.Join(home, ".config", ProgName, "config.json")
	}

	cfg, err := config.FromFile(fpath)
	if suppressMissingConfigErr && errors.Is(err, os.ErrNotExist) {
		defaultCfg := config.Default
		return &defaultCfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s", fpath)
	}

	return cfg, nil
}

// processArgs takes the args remaining after flags have been removed,
// and returns the resulting template path and date range.
//
// If the is a problem with arguments provided,
// a wrapped [ErrUsage] will be returned.
func processArgs(
	args []string,
	cfg *config.Config,
) (tmplPath string, err error) {
	if len(args) == 0 {
		return "", fmt.Errorf("%w: missing a template file argument", ErrUsage)
	}
	tmplPath = args[0]

	dr, err := daterange.FromArgs(args[1:], cfg.IsHebrewYear, cfg.Now)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUsage, err)
	}
	cfg.DateRange = dr

	// This will be the idea of now for the entire program run.
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

	if cfg.Today && cfg.DateRange.Source.IsZero() {
		cfg.DateRange = daterange.FromTime(cfg.Now)
	}

	return tmplPath, nil
}
