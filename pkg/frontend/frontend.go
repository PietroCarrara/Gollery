package frontend

import (
	"embed"
	"io/fs"
	"path"
)

//go:embed frontend/dist
var frontend embed.FS

type FrontendFS struct {
	root string
	base fs.FS
}

var Frontend fs.FS = &FrontendFS{
	root: "frontend/dist",
	base: frontend,
}

func (f *FrontendFS) Open(name string) (fs.File, error) {
	filepath := path.Join(f.root, name)
	return f.base.Open(filepath)
}
