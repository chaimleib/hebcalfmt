package cli

import (
	"fmt"
	"log"
	"maps"
	"os"
	"text/template"
	"time"

	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/templating"
)

func usage() string {
	prog := os.Args[0]
	return fmt.Sprintf("usage: %s template-path", prog)
}

func Run() int {
	log.SetFlags(0)

	opts, tmpl, err := handleArgs()
	if err != nil {
		log.Println(err)
		return 1
	}

	now := time.Now()
	z := zmanim.New(opts.Location, now)

	tz, err := time.LoadLocation(opts.Location.TimeZoneId)
	if err != nil {
		log.Println(err)
		return 1
	}

	err = tmpl.Execute(os.Stdout, map[string]any{
		"now":      now,
		"tz":       tz,
		"location": opts.Location,
		"z":        &z,
		"time":     templating.TimeConsts,
		"event":    templating.EventConsts,
	})
	if err != nil {
		log.Println(err)
		return 1
	}

	return 0
}

func handleArgs() (opts *hebcal.CalOptions, tmpl *template.Template, err error) {
	if len(os.Args) != 2 {
		log.Println(usage())
		return nil, nil, fmt.Errorf("expected path to a template")
	}

	loc := zmanim.LookupCity("Phoenix")
	now := time.Now()
	opts = &hebcal.CalOptions{
		Year:           now.Year(),
		Month:          now.Month(),
		Sedrot:         true,
		CandleLighting: true,
		DailyZmanim:    true,
		Location:       loc,
		HavdalahDeg:    zmanim.Tzeit3SmallStars,
		NoModern:       true,
	}

	// Set up the Template's FuncMap.
	// This must be done before parsing the file.
	tmpl = new(template.Template)
	tmpl = setFuncMap(tmpl, opts)

	tmpl, err = parseFile(tmpl, os.Args[1])
	if err != nil {
		return nil, nil, err
	}

	return opts, tmpl, nil
}

func setFuncMap(
	tmpl interface {
		Funcs(template.FuncMap) *template.Template
	},
	opts *hebcal.CalOptions,
) *template.Template {
	funcs := make(map[string]any)
	maps.Insert(funcs, maps.All(templating.HebcalFuncs(opts)))
	maps.Insert(funcs, maps.All(templating.StringFuncs))
	maps.Insert(funcs, maps.All(templating.TimeFuncs))
	maps.Insert(funcs, maps.All(templating.CastFuncs))
	maps.Insert(funcs, maps.All(templating.EnvFuncs))
	return tmpl.Funcs(funcs)
}

func parseFile(tmpl *template.Template, fpath string) (*template.Template, error) {
	buf, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.Parse(string(buf))
	return tmpl, err
}
