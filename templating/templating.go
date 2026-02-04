package templating

import (
	"fmt"
	"io"
	"io/fs"
	"maps"
	"text/template"

	"github.com/chaimleib/hebcalfmt/config"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/zmanim"
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
	return tmpl.Funcs(funcs)
}

// BuildData loads tmplPath from the files and configures it.
// It returns the tmpl and the data on which it should be executed.
//
// In detail:
//
//  1. Builds variables and adds them to a data map,
//     which the template will run on:
//     - `$.now`
//     - `$.nowInLocation`
//     - `$.calOptions`
//     - `$.language`
//     - `$.dateRange`
//     - `$.tz`
//     - `$.location`
//     - `$.z`
//     - `$.hdate.*`
//     - `$.event.*`
//     - `$.sedra.*`
//     - `$.time.*`
//
//  2. Builds the FuncMap and adds it to the template.
//
//  3. Parses the template.
//
//  4. Sets the template ParseName (for runtime debugging messages).
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
