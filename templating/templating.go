package templating

import (
	"fmt"
	"io"
	"io/fs"
	"maps"
	"text/template"

	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/config"
)

// ParseFile opens fpath from the files given and parses it into the tmpl.
func ParseFile(
	files fs.FS,
	tmpl *template.Template,
	fpath string,
) (*template.Template, error) {
	f, err := files.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.Parse(string(buf))
	return tmpl, err
}

// SetFuncMap loads the hebcalfmt templating functions into tmpl's FuncMap.
func SetFuncMap(
	tmpl interface {
		Funcs(template.FuncMap) *template.Template
	},
	opts *hebcal.CalOptions,
) *template.Template {
	funcs := make(map[string]any)
	maps.Insert(funcs, maps.All(CalOptionsFuncs(opts)))
	maps.Insert(funcs, maps.All(HebcalFuncs(opts)))
	maps.Insert(funcs, maps.All(ZmanimFuncs(opts)))
	maps.Insert(funcs, maps.All(HDateFuncs))
	maps.Insert(funcs, maps.All(SedraFuncs))
	maps.Insert(funcs, maps.All(StringFuncs))
	maps.Insert(funcs, maps.All(TimeFuncs))
	maps.Insert(funcs, maps.All(CastFuncs))
	maps.Insert(funcs, maps.All(EnvFuncs))
	maps.Insert(funcs, maps.All(CollectionFuncs))
	return tmpl.Funcs(funcs)
}

// BuildData loads tmplPath from the files and configures it.
// It returns the tmpl and the data on which it should be executed.
//
// In detail:
//
//  1. Builds variables and adds them to a data map,
//     which the template will run on.
//
//  2. Builds the FuncMap and adds it to the template.
//
//  3. Parses the template.
//
//  4. Sets the template ParseName (for runtime debugging messages).
//
// These are the variables provided to the template:
//
//   - `$.now` - the current time
//   - `$.nowInLocation` - the current time, localized
//     to the config file's timezone.
//   - `$.calOptions` - the options that will be passed through
//     to hebcal library functions.
//     This is controlled via the JSON config file, CLI arguments,
//     and certain config-altering functions which can be called
//     from the template itself.
//   - `$.configSource` - the name of the JSON config file,
//     or else the empty string if the compiled default config was used.
//   - `$.language` - the name of the language to be used.
//   - `$.dateRange` - the [daterange.DateRange] implied or specified
//     by the command line arguments.
//   - `$.tz` - the [time.Location] of the configured city.
//   - `$.location` - a [zmanim.Location] configuring which place
//     to calculate zmanim and holidays for.
//     This can be customized in the JSON config via `city`,
//     `geo.lat`, `geo.lon`, `timezone`,
//     and `il` (whether the place is in Israel).
//   - `$.z` - a [zmanim.Zmanim] object for calculating zmanim for a location.
//   - `$.hdate.*` - [HDateConsts], a map of constants for Hebrew dates.
//   - `$.event.*` - [EventConsts], a map of constants for categorizing events.
//   - `$.sedra.*` - [SedraConsts], a map of constants for parshas.
//   - `$.time.*` - [TimeConsts], a map of constants for the [time] package.
func BuildData(
	cfg *config.Config,
	files fs.FS,
	tmplPath string,
) (*template.Template, map[string]any, error) {
	opts, err := cfg.CalOptions()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build hebcal options from %s: %w",
			cfg.ConfigSource, err)
	}

	z := zmanim.New(opts.Location, cfg.Now)

	// Set up the Template's FuncMap.
	// This must be done before parsing the file.
	tmpl := template.New(tmplPath)
	tmpl = SetFuncMap(tmpl, opts)

	tmpl, err = ParseFile(files, tmpl, tmplPath)
	if err != nil {
		return nil, nil, err
	}
	tmpl.ParseName = tmplPath
	return tmpl, map[string]any{
		"now":           cfg.Now,
		"nowInLocation": cfg.Now.In(z.TimeZone),
		"calOptions":    opts,
		"configSource":  cfg.ConfigSource,
		"language":      cfg.Language,
		"dateRange":     cfg.DateRange,
		"tz":            z.TimeZone,
		"location":      opts.Location,
		"z":             &z,
		"hdate":         HDateConsts,
		"event":         EventConsts,
		"sedra":         SedraConsts,
		"time":          TimeConsts,
	}, nil
}
