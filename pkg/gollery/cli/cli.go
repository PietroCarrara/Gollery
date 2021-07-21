package cli

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/PietroCarrara/Gollery/pkg/gollery"
	"github.com/urfave/cli/v2"
)

func getConfig(c *cli.Context) (config gollery.Config, err error) {
	config = gollery.Config{}

	if contents, err := os.Open(c.String("config-file")); err == nil {
		defer contents.Close()
		bytes, _ := ioutil.ReadAll(contents)
		json.Unmarshal(bytes, &config)
	}

	return
}

func saveConfig(c *cli.Context, config gollery.Config) error {
	contents, err := os.Create(c.String("config-file"))
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
