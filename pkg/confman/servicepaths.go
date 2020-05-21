package confman

import (
	"fmt"
	"path"
	"strings"
)

func ParseServicePaths(servicePathsInput string) []string {
	servicePaths := make([]string, 0, 10)
	for _, servicePath := range strings.Split(servicePathsInput, ",") {
		servicePaths = append(servicePaths, parseServicePath(servicePath)...)
	}
	return servicePaths
}

func parseServicePath(servicePath string) []string {
	serviceEnvironments := strings.Split(servicePath, "+")

	servicePaths := make([]string, 0, len(serviceEnvironments))
	serviceDir, environment := path.Split(FormatServicePath(serviceEnvironments[0]))

	environments := append([]string{environment}, serviceEnvironments[1:]...)
	for _, environment := range environments {
		servicePath := FormatServicePath(path.Join(serviceDir, environment))
		servicePaths = append(servicePaths, servicePath)
	}

	return servicePaths
}

// FormatServicePath takes as input a short-hand service path and returns
// the full one, ensuring a prefixed "/", and no suffix "/", e.g.
// translating `service/name` to `/service/name`.
func FormatServicePath(servicePath string) string {
	// TODO: validate if valid service name
	if !strings.HasPrefix(servicePath, "/") {
		servicePath = fmt.Sprintf("/%s", servicePath)
	}

	return strings.TrimRight(servicePath, "/")
}
