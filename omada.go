package coredns_omada

import (
	"context"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/file"
	"github.com/coredns/coredns/request"
	omada "github.com/dougbw/go-omada"

	"github.com/miekg/dns"
)

// Omada is the core struct of the omada plugin.
type Omada struct {
	config     config
	controller omada.Controller
	sites      []string
	zoneNames  []string
	zones      map[string]*file.Zone
	zMu        sync.RWMutex
	records    map[string]DnsRecords
	Next       plugin.Handler
}

func NewOmada(ctx context.Context, url string, u string, p string) (*Omada, error) {

	omada := omada.New(url)

	zones := make(map[string]*file.Zone)
	records := make(map[string]DnsRecords)

	return &Omada{
		controller: omada,
		zones:      zones,
		records:    records,
	}, nil
}

const ptrZone string = "in-addr.arpa."

// ServeDNS implements the plugin.Handler interface.
func (o *Omada) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()
	qtype := state.QType()
	log.Debugf("query; type: %d, name: %s\n", qtype, qname)

	// this plugin can only handle 'A' and 'PTR' queries
	var qzone string
	switch qtype {
	case 1: // A
		qzone = qname
	case 12: // PTR
		qzone = ptrZone
	default:
		return plugin.NextOrFailure(o.Name(), o.Next, ctx, w, r)
	}

	// check zone
	log.Debugf("checking if zone is managed: %s", qzone)
	zoneName := plugin.Zones(o.zoneNames).Matches(qzone)
	if zoneName == "" {
		log.Debugf("-- ❌ query is not in managed zones: %s\n", qname)
		return plugin.NextOrFailure(o.Name(), o.Next, ctx, w, r)
	}
	log.Debugf("-- ✅ zone name: %s\n", zoneName)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	var result file.Result

	// lookup record in zones
	o.zMu.RLock()
	m.Answer, m.Ns, m.Extra, result = o.zones[zoneName].Lookup(ctx, state, qname)
	o.zMu.RUnlock()

	// no answer
	if len(m.Answer) == 0 && result != file.NoData {
		log.Debugf("-- ❌ answer len: %d, result: %v\n", len(m.Answer), result)
		return plugin.NextOrFailure(o.Name(), o.Next, ctx, w, r)
	}
	log.Debugf("-- ✅ answer len: %d, result: %v\n", len(m.Answer), result)

	switch result {
	case file.Success:
	case file.NoData:
	case file.NameError:
		m.Rcode = dns.RcodeNameError
	case file.Delegation:
		m.Authoritative = false
	case file.ServerFailure:
		log.Debugf("RcodeServerFailure")
		return dns.RcodeServerFailure, nil
	}

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

func (o *Omada) Name() string { return "omada" }
