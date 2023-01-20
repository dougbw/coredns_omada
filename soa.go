package coredns_omada

import (
	"github.com/coredns/coredns/plugin/file"
	"github.com/miekg/dns"
)

func addSoaRecord(zone *file.Zone, domain string) {

	soa := &dns.SOA{Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 300},
		Minttl:  uint32(3600),
		Expire:  uint32(3600),
		Retry:   uint32(3600),
		Refresh: uint32(3600),
		Serial:  uint32(3600),
		Mbox:    dns.Fqdn(domain),
		Ns:      "127.0.0.1"}
	zone.Insert(soa)

}
