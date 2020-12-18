package server

import (
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/afero"
)

func NewPackrFs(box *packr.Box) afero.Fs {
	fs := afero.NewMemMapFs()

	for _, file := range box.List() {
		b, err := box.Find(file)

		if err != nil {
			continue
		}

		afero.WriteFile(fs, file, b, 0664)
	}

	return fs
}
