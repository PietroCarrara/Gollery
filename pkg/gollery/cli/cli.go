package cli

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path"

	"github.com/PietroCarrara/Gollery/pkg/gollery"
	"github.com/urfave/cli/v2"
)

var configPath = "."

func getConfig(c *cli.Context) (config gollery.Config, err error) {
	config = gollery.Config{}

	pwd, err := os.Getwd()
	if err != nil {
		return
	}

	// While not at root, keep going up to find the config file
	for pwd != "/" {

		var files []fs.FileInfo
		files, err = ioutil.ReadDir(pwd)
		if err != nil {
			return
		}

		// Try to find the config
		for _, file := range files {
			if file.Name() == "gollery.json" {
				var f *os.File
				f, err = os.Open(path.Join(pwd, file.Name()))
				if err != nil {
					return
				}
				defer f.Close()

				configPath = pwd

				dec := json.NewDecoder(f)
				err = dec.Decode(&config)

				return
			}
		}

		pwd = path.Join(pwd, "..")
	}

	return
}

func saveConfig(c *cli.Context, config gollery.Config) error {
	contents, err := os.Create(path.Join(configPath, "gollery.json"))
	if err != nil {
		return err
	}
	defer contents.Close()

	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	_, err = contents.Write(bytes)

	return err
}

func contains(haystack []string, needle string) bool {
	for _, str := range haystack {
		if needle == str {
			return true
		}
	}

	return false
}
