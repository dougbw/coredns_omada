package coredns_omada

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterSites(t *testing.T) {

	tests := []struct {
		pattern  string
		sites    []string
		expected []string
	}{
		{
			pattern: `home`,
			sites: []string{
				"home",
				"work",
			},
			expected: []string{
				"home",
			},
		},
		{
			pattern: `.*`,
			sites: []string{
				"home",
				"work",
			},
			expected: []string{
				"home",
				"work",
			},
		},
	}

	for _, test := range tests {
		actual := filterSites(test.pattern, test.sites)
		assert.Equal(t, test.expected, actual)
	}
}
