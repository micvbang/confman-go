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
	serviceDir, environment := path.Split(FormatServiceName(serviceEnvironments[0]))

	environments := append([]string{environment}, serviceEnvironments[1:]...)
	for _, environment := range environments {
		serviceName := FormatServiceName(path.Join(serviceDir, environment))
		servicePaths = append(servicePaths, serviceName)
	}

	return servicePaths
}

// FormatServiceName takes as input a short-hand service name and returns
// the correct one, e.g. converting `service/name` to `/service/name`.
func FormatServiceName(serviceName string) string {
	// TODO: validate if valid service name
	if !strings.HasPrefix(serviceName, "/") {
		serviceName = fmt.Sprintf("/%s", serviceName)
	}

	return strings.TrimRight(serviceName, "/")
}
