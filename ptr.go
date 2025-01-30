package coredns_omada

import (
	"fmt"
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
	return fmt.Sprintf("%s.%s", reverse, ptrZone)
}

func reverseSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
