package main

import (
	"log"
	"os"

	gollery "github.com/PietroCarrara/Gollery/pkg/gollery/cli"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "group",
				Usage:  "creates a new group using the provided argument as the root",
				Action: gollery.Group,
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
				Action: gollery.Serve,
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
				Action: gollery.List,
			},
			{
				Name:   "thumb",
				Usage:  "generates/updates the thumbnails for all the files",
				Action: gollery.Thumb,
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
