package templating

import (
	"os"
	"text/template"
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
