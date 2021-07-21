package gollery

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

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

		const segmentCount = 12
		const segmentDuration float64 = 4

		margins := .1 // 10%
		start := margins * float64(totalLength)
		length := float64(totalLength) - start // Cut the start

		var segments []string

		// Make small video segments
		for i := 0; i < segmentCount; i++ {
			fname := path.Join(name, fmt.Sprintf("thumb-segment-%03d.ts", i+1))
			segments = append(segments, fname)
			makeThumb := exec.Command(
				"ffmpeg",
				"-ss",
				fmt.Sprintf("%f", start+length*(float64(i)/segmentCount)),
				"-i",
				f.Path,
				"-map",
				"0:v:0",
				"-t",
				fmt.Sprintf("%f", segmentDuration),
				"-vf",
				"scale=-2:240",
				"-preset",
				"veryfast",
				"-c:v",
				"libx264",
				"-crf",
				"19",
				fname,
			)

			err := makeThumb.Run()
			if err != nil {
				os.RemoveAll(name)
				return err
			}
		}

		// Join the segments
		err = exec.Command(
			"ffmpeg",
			"-i",
			fmt.Sprintf("concat:%s", strings.Join(segments, "|")),
			"-c",
			"copy",
			path.Join(name, "thumb.mp4"),
		).Run()
		if err != nil {
			return err
		}

		// Delete the individual segments
		for _, fname := range segments {
			err := os.Remove(fname)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("could not make thumbnail for filetype '%s'", f.Type)
}
