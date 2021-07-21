package cli

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"
)

func List(c *cli.Context) error {
	config, err := getConfig(c)
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
