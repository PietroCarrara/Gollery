package gollery

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"gopkg.in/vansante/go-ffprobe.v2"
)

func (f File) GenThumbnails(dir string, force bool) error {
	name := path.Join(dir, f.Path)

	stat, err := os.Stat(name)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If we are not being forced to regen and the file exists,
	// check if we have to regen them
	if !force && !os.IsNotExist(err) {
		if stat.ModTime().After(f.Mtime) {
			// Thumbnail modified after the file, no need to regen
			return nil
		}
	}

	// Remove the old thumbnails
	err = os.RemoveAll(name)
	if err != nil {
		return err
	}

	os.MkdirAll(name, os.FileMode(0777))

	switch f.Type {
	case Video:
		data, err := ffprobe.ProbeURL(context.TODO(), f.Path)
		if err != nil {
			return err
		}

		totalLength := data.Format.DurationSeconds

		const thumbCount = 10
		const videoDuration float64 = 5

		margins := .1 // 10%
		start := margins * float64(totalLength)
		length := float64(totalLength) - 2*start // Cut start and end margins

		// Make small video segments
		for i := 0; i < thumbCount; i++ {
			makeThumb := exec.Command(
				"ffmpeg",
				"-ss",
				fmt.Sprintf("%f", start+length*(float64(i)/thumbCount)),
				"-i",
				f.Path,
				"-map",
				"0:v",
				"-t",
				fmt.Sprintf("%f", videoDuration),
				"-vf",
				"scale=-2:240",
				"-preset",
				"veryfast",
				path.Join(name, fmt.Sprintf("thumb-%d.mp4", i+1)),
			)

			err := makeThumb.Run()
			if err != nil {
				os.RemoveAll(name)
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("could not make thumbnail for filetype '%s'", f.Type)
}
