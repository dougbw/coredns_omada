package coredns_omada

import (
	"context"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin("omada")

func init() { plugin.Register("omada", setup) }

func setup(c *caddy.Controller) error {
	config, err := parse(c)
	if err != nil {
		return plugin.Error("omada", err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	url := config.Controller_url
	u := config.Username
	p := config.Password

	o, err := NewOmada(ctx, url, u, p)
	if err != nil {
		cancel()
		return plugin.Error("omada", err)
	}
	o.config = config

	// initial login
	if err := o.login(ctx); err != nil {
		cancel()
		return plugin.Error("omada", err)
	}

	// setup site list
	var sites []string
	for s := range o.controller.Sites {
		sites = append(sites, s)
	}
	sites = filterSites(config.Site, sites)
	log.Infof("found '%d' sites: %v", len(sites), sites)
	o.sites = sites

	// initial zone update
	if err := o.updateZones(ctx); err != nil {
		cancel()
		return plugin.Error("omada", err)
	}

	// start update loop
	if err := o.updateLoop(ctx); err != nil {
		cancel()
		return plugin.Error("omada", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		o.Next = next
		return o
	})

	c.OnShutdown(func() error { cancel(); return nil })
	return nil
}
