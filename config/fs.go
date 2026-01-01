package config

import (
	"fmt"
	"io/fs"
	"os"
)

func DefaultFS() (fs.FS, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to Getwd: %w", err)
	}
	return os.DirFS(cwd), nil
}
