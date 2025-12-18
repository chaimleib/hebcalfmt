package templating

import (
	"maps"
	"os"
	"text/template"

	"github.com/hebcal/hebcal-go/hebcal"
)

func ParseFile(
	tmpl *template.Template,
	fpath string,
) (*template.Template, error) {
	buf, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.Parse(string(buf))
	return tmpl, err
}

func SetFuncMap(
	tmpl interface {
		Funcs(template.FuncMap) *template.Template
	},
	opts *hebcal.CalOptions,
) *template.Template {
	funcs := make(map[string]any)
	maps.Insert(funcs, maps.All(HebcalFuncs(opts)))
	maps.Insert(funcs, maps.All(StringFuncs))
	maps.Insert(funcs, maps.All(TimeFuncs))
	maps.Insert(funcs, maps.All(CastFuncs))
	maps.Insert(funcs, maps.All(EnvFuncs))
	return tmpl.Funcs(funcs)
}
