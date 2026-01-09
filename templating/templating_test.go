package templating_test

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
	"text/template"

	"github.com/hebcal/hebcal-go/hebcal"

	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

type ErrReadFile struct{}

var _ fs.File = ErrReadFile{}

func (f ErrReadFile) Read([]byte) (int, error) {
	return 0, errors.New("ErrReadFile")
}

func (f ErrReadFile) Stat() (fs.FileInfo, error) {
	return nil, nil
}

func (f ErrReadFile) Close() error { return nil }

type ErrReaderFS struct{}

func (e ErrReaderFS) Open(fpath string) (fs.File, error) {
	return ErrReadFile{}, nil
}

func TestParseFile(t *testing.T) {
	mapFiles := fstest.MapFS{
		"stub.tmpl":      &fstest.MapFile{Data: []byte("hi")},
		"invalid.tmpl":   &fstest.MapFile{Data: []byte("{{INVALID")},
		"readError.tmpl": &fstest.MapFile{},
	}
	errFiles := ErrReaderFS{}

	cases := []struct {
		Name  string
		Files fs.FS
		Err   string
	}{
		{Name: "stub.tmpl"},
		{
			Name: "invalid.tmpl",
			Err:  `template: invalid.tmpl:1: function "INVALID" not defined`,
		},
		{
			Name: "nonexistent.tmpl",
			Err:  `open nonexistent.tmpl: file does not exist`,
		},
		{
			Name:  "readError.tmpl",
			Files: errFiles,
			Err:   `ErrReadFile`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			files := c.Files
			if files == nil {
				files = mapFiles
			}

			tmpl := template.New(c.Name)
			tmpl, err := templating.ParseFile(files, tmpl, c.Name)
			test.CheckErr(t, err, c.Err)
		})
	}
}

type MockTemplate struct {
	funcs template.FuncMap
}

func (mt *MockTemplate) Funcs(fm template.FuncMap) *template.Template {
	mt.funcs = fm
	return template.New("").Funcs(fm)
}

func TestSetFuncMap(t *testing.T) {
	var mt MockTemplate
	var opts hebcal.CalOptions

	tmpl := templating.SetFuncMap(&mt, &opts)

	if tmpl == nil {
		t.Error("expected a non-nil tmpl")
	}
	if len(mt.funcs) == 0 {
		t.Error("expected at least one func in the resulting funcmap")
	}
}
