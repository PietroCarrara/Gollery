package gollery

import "time"

type FileType string

const (
	// A single video
	Video FileType = "video"
	// A single image
	Image FileType = "image"
)

type FileDir struct {
	Path      string   `json:"path"`
	Tags      []string `json:"tags"`
	Recursive bool     `json:"recursive,omitempty"`
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
