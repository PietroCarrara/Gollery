package gollery

import (
	"github.com/trhura/simplecli"
)

type Gollery struct {
	ConfigPath string
}

func main() {
	simplecli.Handle(&Gollery{
		ConfigPath: "./gollery.json",
	})
}

func (g *Gollery) Serve() {
	// TODO
}
