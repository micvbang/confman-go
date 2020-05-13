package configuration

import (
	"os"

	"gopkg.in/yaml.v2"
)

func readYML(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// This type definition enforces that we only parse non-nested yaml, i.e. only string values.
	config := map[string]string{}
	err = yaml.NewDecoder(f).Decode(&config)
	return config, err
}
