package coredns_omada

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/coredns/coredns/plugin/file"
	omada "github.com/dougbw/go-omada"
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

	err := o.controller.Login(u, p)
	if err != nil {
		return err
	}

	return nil
}

type ARecord struct {
	record    *dns.A
	timestamp time.Time
}

type PtrRecord struct {
	record    *dns.PTR
	timestamp time.Time
}

type DnsRecords struct {
	ARecords   map[string]ARecord
	PtrRecords map[string]PtrRecord
}

func (d *DnsRecords) purgeStaleRecords(maxAge int) {
	now := time.Now()
	for k, v := range d.ARecords {
		diff := now.Sub(v.timestamp)
		if diff.Seconds() > float64(maxAge) {
			delete(d.ARecords, k)
		}
	}
}

// update dns zones
func (o *Omada) updateZones(ctx context.Context) error {

	log.Info("update: updating zones...")
	zones := make(map[string]*file.Zone)
	records := make(map[string]DnsRecords)
	timestamp := time.Now()

	var networks []omada.OmadaNetwork
	var clients []omada.Client
	var devices []omada.Device
	var reservations []omada.DhcpReservation

	for _, s := range o.sites {
		o.controller.SetSite(s)

		log.Debugf("update: getting networks for site: %s", s)
		o.controller.SetSite(s)
		n, err := o.controller.GetNetworks()
		if err != nil {
			return fmt.Errorf("error getting networks from omada controller: %w", err)
		}
		interfaces := getInterfaces(n)
		networks = append(networks, interfaces...)

		log.Debugf("update: getting clients for site: %s", s)
		c, err := o.controller.GetClients()
		if err != nil {
			return fmt.Errorf("error getting clients from omada controller: %w", err)
		}
		clients = append(clients, c...)

		log.Debugf("update: getting devices for site: %s", s)
		d, err := o.controller.GetDevices()
		if err != nil {
			return fmt.Errorf("error getting devices from omada controller: %w", err)
		}
		devices = append(devices, d...)

		log.Debugf("update: getting dhcp reservations for site: %s", s)
		r, err := o.controller.GetDhcpReservations()
		if err != nil {
			return fmt.Errorf("error getting dhcp reservations from omada controller: %w", err)
		}
		reservations = append(reservations, r...)

	}
	log.Debugf("update: found '%d' omada clients\n", len(clients))
	log.Debugf("update: found '%d' omada devices\n", len(devices))
	log.Debugf("update: found '%d' omada reservations\n", len(reservations))

	// reverse zones
	for _, network := range networks {
		dnsDomain := network.Domain
		_, subnet, err := net.ParseCIDR(network.Subnet)
		if err != nil {
			log.Debugf("failed to parse network cidr: %v", err)
			continue
		}
		for _, client := range clients {
			// get PTR zone
			ptrZone := getParentPtrZoneFromIp(client.Ip)
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
			if ip == nil {
				continue
			}
			if subnet.Contains(ip) {
				ptrName := getPtrZoneFromIp(client.Ip)
				ptrRecord := fmt.Sprintf("%s.%s", client.DnsName, dnsDomain)
				ptr := &dns.PTR{Hdr: dns.RR_Header{Name: ptrName, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 60},
					Ptr: dns.Fqdn(ptrRecord)}
				log.Debugf("update ptr: -- adding record to zone: %s, %s", ptrRecord, ptrZone)
				zones[ptrZone].Insert(ptr)
			}
		}
		for _, device := range devices {
			// get PTR zone
			ptrZone := getParentPtrZoneFromIp(device.IP)
			// create PTR zone
			_, ok := zones[ptrZone]
			if !ok {
				log.Debugf("update ptr: creating PTR zone: %s", ptrZone)
				zones[ptrZone] = file.NewZone(ptrZone, "")
				addSoaRecord(zones[ptrZone], ptrZone)
			}

			// if device is in this networks subnet then we can determine the fqdn
			// and create ptr record
			ip := net.ParseIP(device.IP)
			if ip == nil {
				continue
			}
			if subnet.Contains(ip) {
				ptrName := getPtrZoneFromIp(device.IP)
				ptrRecord := fmt.Sprintf("%s.%s", device.DnsName, dnsDomain)
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

		// create record map
		_, ok := records[dnsDomain]
		if !ok {
			log.Debugf("update: creating record map: %s", dnsDomain)
			records[dnsDomain] = DnsRecords{
				ARecords:   make(map[string]ARecord),
				PtrRecords: make(map[string]PtrRecord),
			}
		}

		// process client records
		log.Debugf("update: adding records to zone: %s\n", dnsDomain)
		_, subnet, _ := net.ParseCIDR(network.Subnet)
		for _, client := range clients {

			// if client is in this networks subnet then add record to zone
			ip := net.ParseIP(client.Ip)
			if ip == nil {
				continue
			}
			if subnet.Contains(ip) {
				clientFqdn := fmt.Sprintf("%s.%s", client.DnsName, dnsDomain)
				a := &dns.A{Hdr: dns.RR_Header{Name: clientFqdn, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A: net.ParseIP(client.Ip)}
				records[dnsDomain].ARecords[clientFqdn] = ARecord{
					record:    a,
					timestamp: timestamp,
				}
			}
		}

		// process device records
		for _, device := range devices {
			ip := net.ParseIP(device.IP)
			if subnet.Contains(ip) {
				deviceFqdn := fmt.Sprintf("%s.%s", device.DnsName, dnsDomain)
				a := &dns.A{Hdr: dns.RR_Header{Name: deviceFqdn, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A: net.ParseIP(device.IP)}
				records[dnsDomain].ARecords[deviceFqdn] = ARecord{
					record:    a,
					timestamp: timestamp,
				}
			}
		}

		// process dhcp reservation records
		for _, reservation := range reservations {
			ip := net.ParseIP(reservation.IP)
			if ip == nil {
				continue
			}
			if subnet.Contains(ip) {
				dnsName := reservation.ClientName
				if reservation.ClientName == reservation.Mac && reservation.Description != "" {
					dnsName = reservation.Description
				}
				reservationFqdn := fmt.Sprintf("%s.%s", dnsName, dnsDomain)
				a := &dns.A{Hdr: dns.RR_Header{Name: reservationFqdn, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A: net.ParseIP(reservation.IP)}
				records[dnsDomain].ARecords[reservationFqdn] = ARecord{
					record:    a,
					timestamp: timestamp,
				}
			}
		}

	}

	// add records to zone
	for dnsDomain, domainRecords := range records {
		_, ok := zones[dnsDomain]
		if !ok {
			log.Debugf("update: creating zone: %s", dnsDomain)
			zones[dnsDomain] = file.NewZone(dnsDomain, "")
			addSoaRecord(zones[dnsDomain], dnsDomain)
		}
		domainRecords.purgeStaleRecords(300)
		for _, v := range domainRecords.ARecords {
			zones[dnsDomain].Insert(v.record)
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

func getInterfaces(networks []omada.OmadaNetwork) (ret []omada.OmadaNetwork) {
	for _, network := range networks {
		match, _ := regexp.MatchString("interface", network.Purpose)
		if match {
			ret = append(ret, network)
		}
	}
	return
}
