package coredns_omada

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/coredns/coredns/plugin/file"
	"github.com/miekg/dns"
)

// Run starts the update loops which:
// - refresh login session token
// - update the dns zones
func (o *Omada) updateLoop(ctx context.Context) error {

	// update zones
	go updateSessionLoop(ctx, o)

	// update zones
	go updateZoneLoop(ctx, o)

	return nil
}

func updateZoneLoop(ctx context.Context, o *Omada) {

	delay := time.Duration(o.config.refresh_minutes) * time.Minute
	timer := time.NewTimer(delay)
	defer timer.Stop()
	for {
		timer.Reset(delay)
		select {
		case <-ctx.Done():
			log.Debugf("Breaking out of zone update loop: %v", ctx.Err())
			return
		case <-timer.C:
			if err := o.updateZones(ctx); err != nil && ctx.Err() == nil {
				log.Errorf("Failed to update zones: %v", err)
			}
		}
	}
}

func updateSessionLoop(ctx context.Context, o *Omada) {

	// delay := 24 * time.Hour
	delay := time.Duration(o.config.refresh_login_hours) * time.Hour
	timer := time.NewTimer(delay)
	defer timer.Stop()
	for {
		timer.Reset(delay)
		select {
		case <-ctx.Done():
			log.Debugf("Breaking out of login update loop: %v", ctx.Err())
			return
		case <-timer.C:
			if err := o.login(ctx); err != nil && ctx.Err() == nil {
				log.Errorf("Failed to login to controller : %v", err)
			}
		}
	}
}

// login to the controller
func (o *Omada) login(ctx context.Context) error {

	log.Info("logging in...")
	u := o.config.Username
	p := o.config.Password
	s := o.config.Site

	err := o.controller.Login(u, p, s)
	if err != nil {
		return err
	}

	return nil
}

// update dns zones
func (o *Omada) updateZones(ctx context.Context) error {

	log.Info("update: updating zones...")
	zones := make(map[string]*file.Zone)

	networks, err := o.controller.GetAllNetworks()
	if err != nil {
		return fmt.Errorf("updateZones: error GetAllNetworks from omada controller: %w", err)
	}
	log.Debugf("updateZones: GetAllNetworks found '%d' omada networks\n", len(networks))

	clients, err := o.controller.GetAllClients()
	if err != nil {
		return fmt.Errorf("updateZones: error calling GetAllClients from omada controller: %w", err)
	}
	log.Debugf("updateZones: GetAllClients found '%d' omada clients\n", len(clients))

	devices, err := o.controller.GetAllDevices()
	if err != nil {
		return fmt.Errorf("updateZones: error calling GetAllDevices from omada controller: %w", err)
	}
	log.Debugf("updateZones: GetAllDevices found '%d' omada devices\n", len(devices))

	log.Debugf("******************** START CLIENT LIST **************************\n")
	for _, client := range clients {
		log.Debugf("Name: %s, DNS name: %s, Hostname: %s, IP: %s, MAC: %s\n", client.Name, client.DnsName, client.HostName, client.Ip, client.MAC)
	}
	log.Debugf("******************** END CLIENT LIST **************************\n")

	log.Debugf("******************** ================ **************************\n")

	log.Debugf("******************** START DEVICES LIST **************************\n")
	for _, device := range devices {
		log.Debugf("Name: %s, DNS name: %s, Model: %s, IP: %s, MAC: %s, Site: %s\n", device.Name, device.DnsName, device.Model, device.IP, device.Mac, device.Site)
	}
	log.Debugf("******************** END DEVICES LIST **************************\n")

	log.Debugf("******************** ================ **************************\n")

	log.Debugf("******************** START NETWORK LIST **************************\n")

	for _, network := range networks {
		log.Debugf("Domain: %s, ID: %s, Name: %s, Subnet: %s\n", network.Domain, network.Id, network.Name, network.Subnet)
	}
	log.Debugf("******************** END NETWORK LIST **************************\n")

	// reverse zones
	log.Debugf("******************** CREATE ZONES **************************\n")
	for _, network := range networks {
		dnsDomain := network.Domain
		_, subnet, _ := net.ParseCIDR(network.Subnet)
		log.Debugf("ParseCIDR subnet found: %s\n", subnet)
		for _, client := range clients {
			// get PTR zone
			ptrZone := getParentPtrZoneFromIp(client.Ip)
			log.Debugf("parentPtrZoneFromIp: %s for IP: %s\n", ptrZone, client.Ip)
			// create PTR zone
			_, ok := zones[ptrZone]
			if !ok {
				log.Debugf("update ptr: creating PTR zone: %s", ptrZone)
				zones[ptrZone] = file.NewZone(ptrZone, "")
				addSoaRecord(zones[ptrZone], ptrZone)
			}

			// if client is in this networks subnet then we can determine the fqdn
			// and create ptr record
			ip := net.ParseIP(client.Ip)
			if subnet.Contains(ip) {
				ptrName := getPtrZoneFromIp(client.Ip)
				ptrRecord := fmt.Sprintf("%s.%s", client.DnsName, dnsDomain)
				ptr := &dns.PTR{Hdr: dns.RR_Header{Name: ptrName, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 60},
					Ptr: dns.Fqdn(ptrRecord)}
				log.Debugf("update ptr: -- adding record to zone: %s, %s", ptrRecord, ptrZone)
				zones[ptrZone].Insert(ptr)
			}
		}

	}

	// forward zones
	for _, network := range networks {

		log.Debugf("update: -- processing network: %s", network.Name)

		// skip network if no dns search domain is set
		if network.Domain == "" {
			log.Debugf("update: skipping network: %s because not DNS search domain is set", network.Name)
			continue
		}
		dnsDomain := network.Domain + "."

		// create zone
		_, ok := zones[dnsDomain]
		if !ok {
			log.Debugf("update: creating zone: %s", dnsDomain)
			zones[dnsDomain] = file.NewZone(dnsDomain, "")
			addSoaRecord(zones[dnsDomain], dnsDomain)
		}

		// add client records to zone
		// todo: if o.config.resolve_client_names...
		log.Debugf("update: adding records to zone: %s\n", dnsDomain)
		_, subnet, _ := net.ParseCIDR(network.Subnet)
		for _, client := range clients {

			// if client is in this networks subnet then add record to zone
			ip := net.ParseIP(client.Ip)
			if subnet.Contains(ip) {
				clientFqdn := fmt.Sprintf("%s.%s", client.DnsName, dnsDomain)
				a := &dns.A{Hdr: dns.RR_Header{Name: clientFqdn, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A: net.ParseIP(client.Ip)}
				zones[dnsDomain].Insert(a)
			}
		}

		// add device records to zone
		for _, device := range devices {
			ip := net.ParseIP(device.IP)
			if subnet.Contains(ip) {
				deviceFqdn := fmt.Sprintf("%s.%s", device.DnsName, dnsDomain)
				a := &dns.A{Hdr: dns.RR_Header{Name: deviceFqdn, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A: net.ParseIP(device.IP)}
				zones[dnsDomain].Insert(a)
			}
		}

		log.Debugf("update: zone %s contains %d records", dnsDomain, zones[dnsDomain].Count)

	}

	// get list of zone names
	zoneNames := make([]string, 0, len(zones))
	for k := range zones {
		zoneNames = append(zoneNames, k)
	}

	o.zMu.Lock()
	o.zones = zones
	o.zoneNames = zoneNames
	o.zMu.Unlock()

	return nil
}
