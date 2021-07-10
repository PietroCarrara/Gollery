package gollery

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	"gopkg.in/vansante/go-ffprobe.v2"
)

type Thumbnail struct {
	Path string   `json:"path"`
	Type FileType `json:"type"`
}

func (f File) GetThumbnails(dir string) []Thumbnail {
	name := path.Join(dir, f.Path)

	files, err := ioutil.ReadDir(name)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("warn: error while fetching thumbnails for '%s': %e\n", f.Path, err)
		}
		return make([]Thumbnail, 0)
	}

	thumbs := make([]Thumbnail, 0, len(files))
	for _, file := range files {
		thumbs = append(thumbs, Thumbnail{
			Path: path.Join(f.Path, file.Name()),
			Type: typeFromFilename(file.Name()),
		})
	}

	return thumbs
}

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
		length := float64(totalLength) - start // Cut the start

		// Make small video segments
		for i := 0; i < thumbCount; i++ {
			makeThumb := exec.Command(
				"ffmpeg",
				"-ss",
				fmt.Sprintf("%f", start+length*(float64(i)/thumbCount)),
				"-i",
				f.Path,
				"-map",
				"0:v:0",
				"-t",
				fmt.Sprintf("%f", videoDuration),
				"-vf",
				"scale=-2:240",
				"-preset",
				"veryfast",
				path.Join(name, fmt.Sprintf("thumb-%03d.mp4", i+1)),
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
