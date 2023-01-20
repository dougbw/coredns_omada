package coredns_omada

import (
	"context"
	"fmt"
	"time"

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

	url := config.controller_url
	u := config.username
	p := config.password

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

	// initial zone update
	if err := o.updateZones(ctx); err != nil {
		cancel()
		return plugin.Error("omada", err)
	}

	delay := 5 * time.Minute
	fmt.Printf("delay: %d\n", delay)

	// start update loop
	refresh := time.Duration(o.config.refresh_minutes) * time.Minute
	fmt.Printf("refresh: %d\n", refresh)
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
