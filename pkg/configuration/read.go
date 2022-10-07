package configuration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/micvbang/go-helpy/filepathy"
	"github.com/micvbang/go-helpy/mapy"
)

var supportedConfigExtensions = []string{".yml"}

func Read(path string) ([]ServiceConfig, error) {
	// TODO: consider globbing
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	configPaths := getConfigPaths(path)

	serviceConfigs := make([]ServiceConfig, 0, len(configPaths))
	for _, configPath := range configPaths {
		config, err := readConfiguration(configPath)
		if err != nil {
			return nil, err
		}

		serviceConfigs = append(serviceConfigs, ServiceConfig{
			Path:   configPath,
			Config: config,
		})
	}

	return serviceConfigs, nil
}

var extensionReader map[string]func(string) (map[string]string, error) = map[string]func(string) (map[string]string, error){
	".yml": readYML,
}

func readConfiguration(path string) (map[string]string, error) {
	ext := filepath.Ext(path)
	f, exists := extensionReader[ext]
	if !exists {
		supportedFormats := mapy.Keys(extensionReader)
		return nil, ConfigError{msg: fmt.Sprintf("failed to parse file \"%s\". Configuration in \"%s\" format not supported. Supported formats are: %v", path, ext, supportedFormats)}
	}

	return f(path)
}

func getConfigPaths(path string) []string {
	walkConfig := filepathy.WalkConfig{
		Dirs:       false,
		Files:      true,
		Recursive:  true,
		Root:       true,
		Extensions: supportedConfigExtensions,
	}

	filePaths := []string{}
	filepathy.Walk(path, walkConfig, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		filePaths = append(filePaths, path)
		return nil
	})

	return filePaths
}

type ServiceConfig struct {
	Path   string
	Config map[string]string
}

type ConfigError struct {
	msg string
}

func (c ConfigError) Error() string {
	return fmt.Sprintf("ConfigError: %s", c.msg)
}
