package cli

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/PietroCarrara/Gollery/pkg/gollery"
	"github.com/urfave/cli/v2"
)

func Group(c *cli.Context) error {
	config, err := getConfig(c)
	if err != nil {
		return err
	}

	if c.String("pattern") != "" && len(c.StringSlice("extension")) > 0 {
		return fmt.Errorf("--pattern and --extension are not usable together")
	}

	pattern := c.String("pattern")
	if len(c.StringSlice("extension")) > 0 {
		pattern = "(?i)(" + strings.Join(c.StringSlice("extension"), "|") + ")$"
	}

args:
	for _, arg := range c.Args().Slice() {
		dirPath := path.Clean(arg)
		if !path.IsAbs(dirPath) {
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}
			dirPath = path.Join(pwd, dirPath)
		}

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
				for _, flag := range []string{"no-recursion", "pattern", "extension"} {
					if c.IsSet(flag) {
						log.Printf("warn: flag '--%s' is not supported when grouping files that already belong in a group. Ignoring...\n", flag)
					}
				}

				continue args
			}
		}

		// If we did not find a previous directory to put this info inside,
		// create a new one
		directory := gollery.FileDir{
			Path:      dirPath,
			Tags:      c.StringSlice("tags"),
			Recursive: !c.Bool("no-recursion"),
			Pattern:   pattern,
		}

		config.Directories = append(config.Directories, directory)
	}

	return saveConfig(c, config)
}
