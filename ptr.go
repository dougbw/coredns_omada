package coredns_omada

import (
	"net"
	"strings"
)

// takes an IPv4 address and returns PTR record:
// 10.0.0.1 -> 1.0.0.10.in-addr.arpa
func getPtrZoneFromIp(ip string) string {
	ipAddr := net.ParseIP(ip)
	ipAddr = ipAddr.To4()
	ipParts := strings.Split(ipAddr.String(), ".")
	reverseParts := reverseSlice(ipParts)
	reverse := strings.Join(reverseParts[:], ".")
	return reverse + ".in-addr.arpa."
}

// takes an IPv4 address and returns ptr zone:
// 10.0.0.1 -> 0.0.10.in-addr.arpa.
func getParentPtrZoneFromIp(ip string) string {
	ptr := getPtrZoneFromIp(ip)
	parent := getPtrParent(ptr)
	return parent
}

// takes PTR record and returns parent ptr zone:
// 1.0.0.10.in-addr.arpa -> 0.0.10.in-addr.arpa
func getPtrParent(ptr string) string {
	parts := strings.Split(ptr, ".")
	zone := strings.Join(parts[1:], ".")
	return zone
}

func reverseSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
