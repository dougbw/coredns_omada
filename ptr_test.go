package coredns_omada

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPtrZoneFromIp(t *testing.T) {

	tests := []struct {
		ip       string
		expected string
	}{
		{
			"10.0.0.100",
			"100.0.0.10.in-addr.arpa.",
		},
		{
			"192.168.0.10",
			"10.0.168.192.in-addr.arpa.",
		},
	}

	for _, test := range tests {
		result := getPtrZoneFromIp(test.ip)
		assert.Equal(t, test.expected, result)
	}

}
