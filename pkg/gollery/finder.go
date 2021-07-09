package gollery

import (
	"sort"
)

type Finder struct {
	config Config

	// all of the files found
	files []File
	// map[tag] = ids of files that have that tag
	tagsToFiles map[string][]int
}

type FinderFile struct {
	File
	ID int `json:"id"`
}

func (c Config) Finder() (Finder, error) {
	finder := Finder{
		config:      c,
		tagsToFiles: make(map[string][]int),
	}

	for _, dir := range c.Directories {
		list, err := dir.ListFiles()
		if err != nil {
			return finder, nil
		}

		// Link the files to their respective tags
		for i, file := range list {
			for _, tag := range file.Tags {
				finder.tagsToFiles[tag] = append(finder.tagsToFiles[tag], i+len(finder.files))
			}
		}

		finder.files = append(finder.files, list...)
	}

	return finder, nil
}

func (f Finder) FindTags() []string {
	keys := make([]string, 0, len(f.tagsToFiles))

	for k := range f.tagsToFiles {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		// Sort by the file count in each tag, reversed
		return !(len(f.tagsToFiles[keys[i]]) < len(f.tagsToFiles[keys[j]]))
	})

	return keys
}

func (f Finder) FindByTag(tag string) []FinderFile {
	list := f.tagsToFiles[tag]

	res := make([]FinderFile, 0, len(list))

	for _, id := range list {
		res = append(res, FinderFile{
			File: f.files[id],
			ID:   id,
		})
	}

	// Sort by mtime, reversed
	sort.Slice(res, func(i, j int) bool {
		return !res[i].Mtime.Before(res[j].Mtime)
	})

	return res
}

func (f Finder) FindByID(id int) FinderFile {
	return FinderFile{
		File: f.files[id],
		ID:   id,
	}
}
