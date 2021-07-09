package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/PietroCarrara/Gollery/pkg/frontend"
	"github.com/PietroCarrara/Gollery/pkg/gollery"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "group",
				Usage:  "creates a new group using the provided argument as the root",
				Action: group,
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "tags",
						Usage:   "The tags to apply to the group",
						Aliases: []string{"t"},
					},
				},
			},
			{
				Name:   "serve",
				Usage:  "starts the server",
				Action: serve,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "port",
						Usage: "Sets the port the server will listen on",
						Value: 8080,
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config-file",
				Usage: "Indicates the config file to use",
				Value: "./gollery.json",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func group(c *cli.Context) error {
	config := getConfig(c)

	base := path.Dir(c.String("config-path"))

	pwd, _ := os.Getwd()

	dir := path.Join(pwd, base, c.Args().Get(0))
	directory := gollery.FileDir{
		Path:      dir,
		Tags:      c.StringSlice("tags"),
		Recursive: c.Bool("recursive"),
	}

	for _, d := range config.Directories {
		if path.Join(d.Path) == path.Join(directory.Path) {
			return fmt.Errorf("group is already added")
		}
	}

	config.Directories = append(config.Directories, directory)

	return saveConfig(c, config)
}

func serve(c *cli.Context) error {
	http.Handle("/", http.FileServer(http.FS(frontend.Frontend)))

	log.Printf("Listening on port http://localhost:%d\n", c.Int("port"))
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", c.Int("port")), nil)
}

func getConfig(c *cli.Context) gollery.Config {
	config := gollery.Config{}

	if contents, err := os.Open(c.String("config-file")); err == nil {
		defer contents.Close()
		bytes, _ := ioutil.ReadAll(contents)
		json.Unmarshal(bytes, &config)
	}

	return config
}

func saveConfig(c *cli.Context, config gollery.Config) error {
	contents, err := os.Create(c.String("config-file"))
	if err != nil {
		return err
	}
	defer contents.Close()

	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	_, err = contents.Write(bytes)

	return err
}
