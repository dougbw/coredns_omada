package coredns_omada

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/coredns/coredns/plugin/file"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

func testZones() map[string]*file.Zone {

	dnsDomain := "omada.test."
	testClients := map[string]string{
		"client1": "192.168.0.101",
		"client2": "192.168.0.102",
		"client3": "192.168.0.103",
	}

	zones := make(map[string]*file.Zone)
	zones[dnsDomain] = file.NewZone(dnsDomain, "")
	addSoaRecord(zones[dnsDomain], dnsDomain)

	zones[ptrZone] = file.NewZone(ptrZone, "")
	addSoaRecord(zones[ptrZone], ptrZone)

	for name, ip := range testClients {
		fqdn := fmt.Sprintf("%s.%s", name, dnsDomain)
		a := &dns.A{Hdr: dns.RR_Header{Name: fqdn, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
			A: net.ParseIP(ip)}
		zones[dnsDomain].Insert(a)

		ptrName := getPtrZoneFromIp(ip)
		ptr := &dns.PTR{Hdr: dns.RR_Header{Name: ptrName, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 60},
			Ptr: dns.Fqdn(fqdn)}
		zones[ptrZone].Insert(ptr)

	}

	return zones
}

func testHandler() test.HandlerFunc {
	return func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
		state := request.Request{W: w, Req: r}
		qname := state.Name()
		m := new(dns.Msg)
		rcode := dns.RcodeServerFailure
		if qname == "example.gov." { // No records match, test fallthrough.
			m.SetReply(r)
			rr := test.A("example.gov.  300 IN  A   2.4.6.8")
			m.Answer = []dns.RR{rr}
			m.Authoritative = true
			rcode = dns.RcodeSuccess
		}
		m.SetRcode(r, rcode)
		w.WriteMsg(m)
		return rcode, nil
	}
}

type testCases struct {
	qname        string
	qtype        uint16
	wantRetCode  int
	wantAnswer   []string
	wantMsgRCode int
	wantNS       []string
	expectedErr  error
}

func TestOmada(t *testing.T) {

	var testOmada = &Omada{
		Next:      testHandler(),
		zoneNames: []string{"omada.test.", ptrZone},
		zones:     testZones(),
	}

	tests := []testCases{
		{
			qname:      "client1.omada.test.",
			qtype:      dns.TypeA,
			wantAnswer: []string{"client1.omada.test.	60	IN	A	192.168.0.101"},
		},
		{
			qname:        "client4.omada.test.",
			qtype:        dns.TypeA,
			wantRetCode:  dns.RcodeServerFailure,
			wantMsgRCode: dns.RcodeServerFailure,
		},
		{
			qname:        "client.example.com.",
			qtype:        dns.TypeA,
			wantRetCode:  dns.RcodeServerFailure,
			wantMsgRCode: dns.RcodeServerFailure,
		},
		{
			qname:      "101.0.168.192.in-addr.arpa.",
			qtype:      dns.TypePTR,
			wantAnswer: []string{"101.0.168.192.in-addr.arpa.	60	IN	PTR	client1.omada.test."},
		},
	}
	executeTestCases(t, testOmada, tests)

}

func executeTestCases(t *testing.T, omada *Omada, testCases []testCases) {
	for ti, tc := range testCases {
		req := new(dns.Msg)
		req.SetQuestion(dns.Fqdn(tc.qname), tc.qtype)

		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		code, err := omada.ServeDNS(context.Background(), rec, req)

		if err != tc.expectedErr {
			t.Fatalf("Test %d: Expected error %v, but got %v", ti, tc.expectedErr, err)
		}

		if code != tc.wantRetCode {
			fmt.Println(tc)
			t.Fatalf("Test %d: Expected returned status code %s, but got %s", ti, dns.RcodeToString[tc.wantRetCode], dns.RcodeToString[code])
		}

		if tc.wantMsgRCode != rec.Msg.Rcode {
			t.Errorf("Test %d: Unexpected msg status code. Want: %s, got: %s", ti, dns.RcodeToString[tc.wantMsgRCode], dns.RcodeToString[rec.Msg.Rcode])
		}

		if len(tc.wantAnswer) != len(rec.Msg.Answer) {
			t.Errorf("Test %d: Unexpected number of Answers. Want: %d, got: %d", ti, len(tc.wantAnswer), len(rec.Msg.Answer))
		} else {
			for i, gotAnswer := range rec.Msg.Answer {
				if gotAnswer.String() != tc.wantAnswer[i] {
					t.Errorf("Test %d: Unexpected answer.\nWant:\n\t%s\nGot:\n\t%s", ti, tc.wantAnswer[i], gotAnswer)
				}
			}
		}

		if len(tc.wantNS) != len(rec.Msg.Ns) {
			t.Errorf("Test %d: Unexpected NS number. Want: %d, got: %d", ti, len(tc.wantNS), len(rec.Msg.Ns))
		} else {
			for i, ns := range rec.Msg.Ns {
				got, ok := ns.(*dns.SOA)
				if !ok {
					t.Errorf("Test %d: Unexpected NS type. Want: SOA, got: %v", ti, reflect.TypeOf(got))
				}
				if got.String() != tc.wantNS[i] {
					t.Errorf("Test %d: Unexpected NS.\nWant: %v\nGot: %v", ti, tc.wantNS[i], got)
				}
			}
		}
	}
}
