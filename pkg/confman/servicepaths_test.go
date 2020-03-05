package confman_test

import (
	"fmt"
	"testing"

	"github.com/micvbang/go-helpy/stringy"
	"github.com/stretchr/testify/require"
	"gitlab.com/micvbang/confman-go/pkg/confman"
)

func TestServicePaths(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected stringy.Set
	}{
		"single service name": {
			input:    "/service/path/environment",
			expected: stringy.MakeSet("/service/path/environment"),
		},
		"service name with +": {
			input:    "/service/dev+prod",
			expected: stringy.MakeSet("/service/dev", "/service/prod"),
		},
		"service name with , and +": {
			input:    "/service/dev+prod,/service2/",
			expected: stringy.MakeSet("/service/dev", "/service/prod", "/service2"),
		},
		"service name with multiple , and +": {
			input:    "/service/dev+prod,/service2/abc+def,/haps",
			expected: stringy.MakeSet("/service/dev", "/service/prod", "/service2/abc", "/service2/def", "/haps"),
		},
		"service fuzzy names with multiple , and +": {
			input:    "/service/dev/+prod/,service2/abc+def,haps",
			expected: stringy.MakeSet("/service/dev", "/service/prod", "/service2/abc", "/service2/def", "/haps"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			servicePaths := confman.ParseServicePaths(test.input)
			require.Equal(t, len(test.expected), len(servicePaths))

			for _, got := range servicePaths {
				require.True(t, test.expected.Contains(got), fmt.Sprintf("Failed to find %s", got))
			}
		})
	}
}
