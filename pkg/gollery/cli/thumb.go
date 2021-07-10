package cli

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/PietroCarrara/Gollery/pkg/gollery"
	"github.com/urfave/cli/v2"
)

func Thumb(c *cli.Context) error {
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
		fmt.Printf("\rProgress: %.2f%%", float64(i*100)/float64(len(files)))

		err := file.GenThumbnails(thumbDir, c.Bool("force-regen"))
		if err != nil {
			return err
		}
	}
	fmt.Print("\rProgress: 100%  ")

	return nil
}
