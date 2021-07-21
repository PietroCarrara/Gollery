package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PietroCarrara/Gollery/pkg/frontend"
	"github.com/PietroCarrara/Gollery/pkg/gollery"
	"github.com/urfave/cli/v2"
)

var finder gollery.Finder

type apiFileFS struct {
}

type frontendFS struct {
	fs http.FileSystem
}

type apiFile struct {
	gollery.FinderFile
	Thumbs []gollery.Thumbnail `json:"thumbs"`
}

func Serve(c *cli.Context) error {
	config, _, err := getConfig(c)
	if err != nil {
		return err
	}

	finder, err = config.Finder()
	if err != nil {
		return err
	}

	files := http.FileServer(&apiFileFS{})
	thumbs := http.FileServer(http.Dir(".gollery/thumbs"))

	http.HandleFunc("/api/tags", apiTags)
	http.HandleFunc("/api/tag", apiTag)

	http.Handle("/files/", http.StripPrefix("/files", files))
	http.Handle("/thumbs/", http.StripPrefix("/thumbs", thumbs))

	front := &frontendFS{
		fs: http.FS(frontend.Frontend),
	}
	http.Handle("/", http.FileServer(front))

	log.Printf("Listening on port http://localhost:%d\n", c.Int("port"))
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", c.Int("port")), nil)
}

// List all the tags
func apiTags(res http.ResponseWriter, req *http.Request) {
	jsonRes(res, finder.FindTags())
}

// List all the files that belong to a tag
func apiTag(res http.ResponseWriter, req *http.Request) {
	tag := req.URL.Query().Get("tag")

	f := finder.FindByTag(tag)
	files := make([]apiFile, 0, len(f))

	for _, file := range f {
		files = append(files, apiFile{
			FinderFile: file,
			Thumbs:     file.GetThumbnails(".gollery/thumbs"),
		})
	}

	jsonRes(res, files)
}

// Open a file by it's id
func (a *apiFileFS) Open(name string) (http.File, error) {
	id, err := strconv.Atoi(strings.TrimPrefix(name, "/"))
	if err != nil {
		return nil, err
	}

	f := finder.FindByID(id)

	return os.Open(f.Path)
}

// Access files from the frontend. If no file is found,
// open index.html
func (f *frontendFS) Open(name string) (http.File, error) {
	file, err := f.fs.Open(name)

	if os.IsNotExist(err) {
		return f.fs.Open("index.html")
	}

	return file, err
}

func jsonRes(res http.ResponseWriter, obj interface{}) error {
	res.Header().Add("Content-Type", "application/json")
	res.Header().Add("Access-Control-Allow-Origin", "*")

	enc := json.NewEncoder(res)
	return enc.Encode(obj)
}
