package shell

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type OverlayFS []fs.FS

var _ fs.FS = OverlayFS(nil)

func (o OverlayFS) Open(fpath string) (fs.File, error) {
	for i, files := range o {
		f, err := files.Open(fpath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return f, fmt.Errorf("OverlayFS[%d].Open %s: %w", i, fpath, err)
		}
		return f, nil
	}
	return nil, fmt.Errorf("OverlayFS.Open %s: %w", fpath, os.ErrNotExist)
}
