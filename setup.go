package coredns_omada

import (
	"context"
	"errors"
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

	url := config.Controller_url
	u := config.Username
	p := config.Password
	o, err := NewOmada(ctx, url, u, p)
	if err != nil {
		cancel()
		return plugin.Error("omada", err)
	}
	o.config = config

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		o.Next = next
		return o
	})

	c.OnShutdown(func() error { cancel(); return nil })

	err = o.startup(ctx)
	if err != nil {
		cancel()
		return plugin.Error("omada", err)
	}

	// start update loops
	go updateSessionLoop(ctx, o)
	go updateZoneLoop(ctx, o)

	return nil
}

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

func (o *Omada) startup(ctx context.Context) error {

	log.Info("starting initial omada setup...")

	retrySeconds := 15 * time.Second

	for {

		err := o.controller.GetControllerInfo()
		if err != nil {
			if o.config.ignore_startup_errors {
				log.Warning(err)
				time.Sleep(retrySeconds)
				continue
			} else {
				// cancel()
				return err
				// return plugin.Error("omada", err)
			}
		}

		err = o.login(ctx)
		if err != nil {
			if o.config.ignore_startup_errors {
				log.Warning(err)
				time.Sleep(retrySeconds)
				continue
			} else {
				// cancel()
				// return plugin.Error("omada", err)
				return err
			}
		}

		// setup site list
		var sites []string
		for s := range o.controller.Sites {
			sites = append(sites, s)
		}
		sites = filterSites(o.config.Site, sites)
		if len(sites) == 0 {
			if o.config.ignore_startup_errors {
				log.Warning(err)
				time.Sleep(retrySeconds)
				continue
			} else {
				// cancel()
				// return plugin.Error("omada", errors.New("no sites found"))
				return errors.New("no sites found")

			}
		}
		log.Infof("found '%d' sites: %v", len(sites), sites)
		o.sites = sites

		// initial zone update
		err = o.updateZones()
		if err != nil {
			if o.config.ignore_startup_errors {
				log.Warning(err)
				time.Sleep(retrySeconds)
				continue
			} else {
				// cancel()
				// return plugin.Error("omada", err)
				return err
			}
		}

		log.Info("initial omada setup complete")
		break
	}

	return nil

}
