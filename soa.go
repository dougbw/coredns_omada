package coredns_omada

import (
	"github.com/coredns/coredns/plugin/file"
	"github.com/miekg/dns"
)

func addSoaRecord(zone *file.Zone, domain string) {

	soa := &dns.SOA{Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 300},
		Minttl:  uint32(300),
		Expire:  uint32(86400),
		Retry:   uint32(3600),
		Refresh: uint32(7200),
		Serial:  uint32(1),
		Mbox:    dns.Fqdn("hostmaster." + domain),
		Ns:      dns.Fqdn("ns." + domain)}
	zone.Insert(soa)

}
