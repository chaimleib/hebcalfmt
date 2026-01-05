package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/chaimleib/hebcalfmt/config"
	"github.com/chaimleib/hebcalfmt/daterange"
)

var (
	// ErrDone indicates that the program should end early with a 0 exit code.
	// This can happen if the help message, version message,
	// or a list operation was requested.
	ErrDone = errors.New("done")

	// ErrUnreachable means that there is a coding defect if returned.
	ErrUnreachable = errors.New("unreachable code")

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
	files fs.FS,
	flagSet *pflag.FlagSet,
	args []string,
) (*config.Config, error) {
	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	// pflag would return an error if
	// - the flag was never defined (i.e. in [NewFlags]),
	// - the flag type and the requested type don't match,
	// - the string value could not be decoded.
	// All of these would indicate coding defects,
	// in which case I want to know the source file and line number.
	// slog is set up for this purpose.
	help, err := flagSet.GetBool("help")
	if err != nil {
		slog.Error("failed to get --help flag", "error", err)
		return nil, fmt.
			Errorf("%w: get --help: %w", ErrUnreachable, err)
	}
	if help {
		fmt.Println(usage(flagSet))
		return nil, ErrDone
	}

	version, err := flagSet.GetBool("version")
	if err != nil {
		slog.Error("failed to get --version flag", "error", err)
		return nil, fmt.Errorf("%w: get --version: %w", ErrUnreachable, err)
	}
	if version {
		fmt.Println(versionMessage())
		return nil, ErrDone
	}

	key, err := flagSet.GetString("info")
	if err != nil {
		slog.Error("failed to get --info option", "error", err)
		return nil, fmt.Errorf("%w: get --info: %w", ErrUnreachable, err)
	}
	if key != "" {
		info, err := infoString(key, flagSet)
		if err != nil {
			return nil, err
		}
		fmt.Println(info)
		return nil, ErrDone
	}

	return loadConfigFromFlags(files, flagSet)
}

// loadConfigFromFlags reads the --config flag option
// and loads the config file specified.
// Otherwise, it loads the default config.
// Then it calls Normalize on the result.
func loadConfigFromFlags(
	files fs.FS,
	flagSet *pflag.FlagSet,
) (*config.Config, error) {
	fpath, err := flagSet.GetString("config")
	if err != nil {
		slog.Error("failed to get --config option", "error", err)
		return nil, fmt.Errorf("%w: get --config: %w", ErrUnreachable, err)
	}

	cfg, err := loadConfigOrDefault(files, fpath)
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
func loadConfigOrDefault(files fs.FS, fpath string) (*config.Config, error) {
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

	cfg, err := config.FromFile(files, fpath)
	if suppressMissingConfigErr && errors.Is(err, os.ErrNotExist) {
		defaultCfg := config.Default
		return &defaultCfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

// processArgs takes the `args` remaining after flags have been removed,
// reads certain fields in `cfg`,
// and returns the resulting template path.
// It sets cfg.DateRange.
//
// If the is a problem with arguments provided,
// a wrapped [ErrUsage] will be returned.
//
// The following fields are read from `cfg`:
//   - IsHebrewYear
//   - Now
//   - Today
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

	if cfg.Today && cfg.DateRange.Source.Defaulted() {
		cfg.DateRange = daterange.FromTime(cfg.Now)
	}

	return tmplPath, nil
}
