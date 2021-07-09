package gollery

import (
	"io/ioutil"
	"log"
	"path"
	"regexp"
	"strings"
	"time"
)

type FileType string

const (
	// A single video
	Video FileType = "video"
	// A single image
	Image FileType = "image"

	// Unknown filetype
	Unknown FileType = "unknown"
)

type FileDir struct {
	Path      string              `json:"path"`
	Tags      []string            `json:"tags"`
	Recursive bool                `json:"recursive,omitempty"`
	ChildTags map[string][]string `json:"child_tags,omitempty"`
	// regex that must find a match on a file's name for that file to be included
	Pattern string `json:"pattern,omitempty"`
}

type Config struct {
	TagConfig   map[string]TagConfig `json:"tag_config"`
	Directories []FileDir            `json:"directories"`
}

type File struct {
	Path  string    `json:"path"`
	Type  FileType  `json:"type"`
	Mtime time.Time `json:"mtime"`
	Tags  []string  `json:"tags"`
}

type TagConfig struct {
	Thumbnail string `json:"thumbnail"`
}

func (f FileDir) ListFiles() ([]File, error) {
	return f.listFiles(f.Path)
}

func (fd FileDir) listFiles(dir string) ([]File, error) {
	var res []File

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return res, err
	}

	regex, err := regexp.Compile(fd.Pattern)
	if err != nil {
		return res, err
	}

	for _, file := range files {
		if file.IsDir() {
			if fd.Recursive {
				children, err := fd.listFiles(path.Join(dir, file.Name()))
				if err != nil {
					return res, err
				}

				res = append(res, children...)
			}
		} else {
			relFilename := path.Join(dir, file.Name())

			if !regex.MatchString(relFilename) {
				continue
			}

			// Tags + childTags if we match a name in the childTags object
			tags := fd.Tags
			if childTags, ok := fd.ChildTags[relFilename]; ok {
				tags = append(tags, childTags...)
			}

			res = append(res, File{
				Path:  path.Join(dir, file.Name()),
				Type:  typeFromFilename(file.Name()),
				Mtime: file.ModTime(),
				Tags:  tags,
			})
		}
	}

	return res, nil
}

func typeFromFilename(fname string) FileType {
	ext := strings.ToLower(path.Ext(fname))

	switch ext {
	case ".mp4":
		return Video
	case ".jpeg", ".jpg", ".png", ".gif":
		return Image
	default:
		log.Printf("warn: could not determine filetype of extension \"%s\"\n", ext)
		return Unknown
	}
}
