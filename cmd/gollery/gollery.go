package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/PietroCarrara/Gollery/pkg/frontend"
	"github.com/PietroCarrara/Gollery/pkg/gollery"
	"github.com/trhura/simplecli"
)

var context = Context{}

type Context struct {
	config gollery.Config
}

type Gollery struct {
	Port   int
	Config string
}

func main() {
	simplecli.Handle(&Gollery{
		Port:   8080,
		Config: "./gollery.json",
	})
}

func (g Gollery) setup() {
	config, _ := os.Open(g.Config)
	defer config.Close()

	bytes, _ := ioutil.ReadAll(config)

	json.Unmarshal(bytes, &context.config)
}

func (g Gollery) Serve() {
	g.setup()

	http.Handle("/", http.FileServer(http.FS(frontend.Frontend)))

	log.Printf("Listening on port http://localhost:%d\n", g.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", g.Port), nil))
}
