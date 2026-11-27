package coredns_omada

import (
	"testing"

	omada "github.com/dougbw/go-omada"
	"github.com/stretchr/testify/assert"
)

func TestGetNetworkDomain(t *testing.T) {
	byName, _ := parseNetworkDomain("Default", "home.arpa")
	bySubnet, _ := parseNetworkDomain("10.0.20.0/24", "office.home.arpa")
	c := config{network_domains: []networkDomain{byName, bySubnet}}

	tests := []struct {
		network omada.OmadaNetwork
		want    string
	}{
		// domain set on the controller takes precedence
		{omada.OmadaNetwork{Name: "Default", Subnet: "10.0.0.1/24", Domain: "omada.home"}, "omada.home"},
		// name match is case-insensitive
		{omada.OmadaNetwork{Name: "default", Subnet: "10.0.0.1/24"}, "home.arpa"},
		// subnet match uses network address and prefix length
		{omada.OmadaNetwork{Name: "Office", Subnet: "10.0.20.1/24"}, "office.home.arpa"},
		{omada.OmadaNetwork{Name: "Office", Subnet: "10.0.20.1/16"}, ""},
		// no match
		{omada.OmadaNetwork{Name: "Guest", Subnet: "10.0.200.1/24"}, ""},
	}
	for i, test := range tests {
		assert.Equal(t, test.want, c.getNetworkDomain(test.network), "test %d", i)
	}

	c.default_domain = "other.home.arpa"
	assert.Equal(t, "other.home.arpa", c.getNetworkDomain(omada.OmadaNetwork{Name: "Guest", Subnet: "10.0.200.1/24"}))
}
