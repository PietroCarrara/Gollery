package cli

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PietroCarrara/Gollery/pkg/frontend"
	"github.com/urfave/cli/v2"
)

func Serve(c *cli.Context) error {
	http.Handle("/", http.FileServer(http.FS(frontend.Frontend)))

	log.Printf("Listening on port http://localhost:%d\n", c.Int("port"))
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", c.Int("port")), nil)
}
