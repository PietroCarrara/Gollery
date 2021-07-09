package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

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
					&cli.BoolFlag{
						Name:  "recursive",
						Usage: "Does this group look for files in subdirectories?",
					},
					&cli.StringFlag{
						Name:  "pattern",
						Usage: "The regex to match on files",
					},
					&cli.StringSliceFlag{
						Name:    "extension",
						Usage:   "Only look for files containing these extensions (-e mp4 -e jpg)",
						Aliases: []string{"e"},
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
			{
				Name:   "list",
				Usage:  "lists all of the files that are part of the gallery",
				Action: list,
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

	if c.String("pattern") != "" && len(c.StringSlice("extension")) > 0 {
		return fmt.Errorf("--pattern and --extension are not usable together")
	}

	pattern := c.String("pattern")
	if len(c.StringSlice("extension")) > 0 {
		pattern = "(?i)" + strings.Join(c.StringSlice("extension"), "|") + "$"
	}

	directory := gollery.FileDir{
		Path:      path.Clean(c.Args().Get(0)),
		Tags:      c.StringSlice("tags"),
		Recursive: c.Bool("recursive"),
		Pattern:   pattern,
	}

	for _, d := range config.Directories {
		if path.Join(d.Path) == path.Join(directory.Path) {
			return fmt.Errorf("group is already added")
		}
	}

	config.Directories = append(config.Directories, directory)

	return saveConfig(c, config)
}

func list(c *cli.Context) error {
	config := getConfig(c)

	for _, dir := range config.Directories {
		files, err := dir.ListFiles()
		if err != nil {
			log.Println(err)
		}
		for _, file := range files {
			fmt.Println(file.Path)
		}
	}

	return nil
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
