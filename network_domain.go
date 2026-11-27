package coredns_omada

import (
	"fmt"
	"net"
	"strings"

	omada "github.com/dougbw/go-omada"
	"github.com/miekg/dns"
)

// networkDomain assigns a DNS domain to Omada networks selected by name or by
// subnet. It is configured with the `network_domain` Corefile property and is
// used for networks on which the controller has no domain set (for example
// Omada 6.3, which has no per-network Domain Name setting).
type networkDomain struct {
	name   string     // match the Omada network name (case-insensitive), or
	subnet *net.IPNet // match the Omada network subnet
	domain string     // domain without trailing dot
}

// parseNetworkDomain parses the arguments of `network_domain <network> <domain>`.
// <network> is treated as a CIDR when it contains a "/", otherwise as a network name.
func parseNetworkDomain(selector, domain string) (networkDomain, error) {
	d, err := parseDomain(domain)
	if err != nil {
		return networkDomain{}, err
	}
	nd := networkDomain{domain: d}
	if strings.Contains(selector, "/") {
		_, subnet, err := net.ParseCIDR(selector)
		if err != nil {
			return networkDomain{}, fmt.Errorf("invalid network CIDR %q: %w", selector, err)
		}
		nd.subnet = subnet
	} else {
		nd.name = selector
	}
	return nd, nil
}

// parseDomain validates a domain name and strips any trailing dot.
func parseDomain(domain string) (string, error) {
	d := strings.TrimSuffix(domain, ".")
	if d == "" {
		return "", fmt.Errorf("invalid domain %q", domain)
	}
	if _, ok := dns.IsDomainName(d); !ok {
		return "", fmt.Errorf("invalid domain %q", domain)
	}
	return d, nil
}

func (nd networkDomain) matches(network omada.OmadaNetwork) bool {
	if nd.subnet != nil {
		_, subnet, err := net.ParseCIDR(network.Subnet)
		if err != nil {
			return false
		}
		return subnet.String() == nd.subnet.String()
	}
	return strings.EqualFold(nd.name, network.Name)
}

// getNetworkDomain returns the DNS domain for a network, or "" to skip it.
// A domain set on the controller always wins. Otherwise the first matching
// `network_domain` applies, then `default_domain` (if configured).
func (c config) getNetworkDomain(network omada.OmadaNetwork) string {
	if network.Domain != "" {
		return network.Domain
	}
	for _, nd := range c.network_domains {
		if nd.matches(network) {
			return nd.domain
		}
	}
	return c.default_domain
}
