package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

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
			{
				Name:   "thumb",
				Usage:  "generates/updates the thumbnails for all the files",
				Action: thumb,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "force-regen",
						Usage: "Forces the thumbnails to be regenerated",
						Value: false,
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
	config, _, err := getConfig(c)
	if err != nil {
		return err
	}

	if c.String("pattern") != "" && len(c.StringSlice("extension")) > 0 {
		return fmt.Errorf("--pattern and --extension are not usable together")
	}

	pattern := c.String("pattern")
	if len(c.StringSlice("extension")) > 0 {
		pattern = "(?i)" + strings.Join(c.StringSlice("extension"), "|") + "$"
	}

	dirPath := path.Clean(c.Args().Get(0))

	// Try to find a group to put the new tags inside
	for _, d := range config.Directories {
		cleanPath := path.Clean(d.Path)

		if cleanPath == dirPath || strings.HasPrefix(dirPath, cleanPath) {
			if cleanPath == dirPath {
				// Add tags directly to this dir
				for _, tag := range c.StringSlice("tags") {
					if !contains(d.Tags, tag) {
						d.Tags = append(d.Tags, tag)
					}
				}
			} else {
				// Add tags as children of this dir
				name := strings.Trim(strings.TrimPrefix(dirPath, cleanPath), "/")

				for _, tag := range c.StringSlice("tags") {
					if !contains(d.ChildTags[name], tag) {
						d.ChildTags[name] = append(d.ChildTags[name], tag)
					}
				}
			}

			// Warn about unused flags
			for _, flag := range []string{"recursive", "pattern", "extension"} {
				if c.IsSet(flag) {
					log.Printf("warn: flag '--%s' is not supported when grouping files that already belong in a group. Ignoring...\n", flag)
				}
			}

			return saveConfig(c, config)
		}
	}

	// If we did not find a previous directory to put this info inside,
	// create a new one
	directory := gollery.FileDir{
		Path:      dirPath,
		Tags:      c.StringSlice("tags"),
		Recursive: c.Bool("recursive"),
		Pattern:   pattern,
	}

	config.Directories = append(config.Directories, directory)

	return saveConfig(c, config)
}

func thumb(c *cli.Context) error {
	var files []gollery.File

	config, pwd, err := getConfig(c)
	if err != nil {
		return err
	}

	thumbDir := path.Join(pwd, ".gollery/thumbs")

	err = os.MkdirAll(thumbDir, os.FileMode(0777))
	if err != nil || (false && thumbDir == "") {
		return err
	}

	for _, d := range config.Directories {
		dirFiles, err := d.ListFiles()
		if err != nil {
			return err
		}

		files = append(files, dirFiles...)
	}

	// Don't let the program die at a Ctrl+C, so we can delete
	// the thumbnails that we were generating
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sig {
			fmt.Println("\nStopping...")
		}
	}()

	fmt.Println()
	defer fmt.Println()
	for i, file := range files {
		fmt.Printf("\rProgress: %.2f%%", float64(i)/float64(len(files)))

		err := file.GenThumbnails(thumbDir, c.Bool("force-regen"))
		if err != nil {
			return err
		}
	}
	fmt.Print("\rProgress: 100%  ")

	return nil
}

func list(c *cli.Context) error {
	config, _, err := getConfig(c)
	if err != nil {
		return err
	}

	for _, dir := range config.Directories {
		files, err := dir.ListFiles()
		if err != nil {
			log.Println(err)
		}
		for _, file := range files {
			fmt.Printf("%s: %v\n", file.Path, file.Tags)
		}
	}

	return nil
}

func serve(c *cli.Context) error {
	http.Handle("/", http.FileServer(http.FS(frontend.Frontend)))

	log.Printf("Listening on port http://localhost:%d\n", c.Int("port"))
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", c.Int("port")), nil)
}

func getConfig(c *cli.Context) (config gollery.Config, pwd string, err error) {
	config = gollery.Config{}

	pwd, err = os.Getwd()
	if err != nil {
		return config, pwd, err
	}

	if contents, err := os.Open(c.String("config-file")); err == nil {
		defer contents.Close()
		bytes, _ := ioutil.ReadAll(contents)
		json.Unmarshal(bytes, &config)
	}

	return
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

func contains(haystack []string, needle string) bool {
	for _, str := range haystack {
		if needle == str {
			return true
		}
	}

	return false
}
