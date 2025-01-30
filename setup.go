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

	if o.config.ignore_startup_errors {
		go o.controllerInit(ctx)
	} else {
		err = o.controllerInit(ctx)
		if err != nil {
			cancel()
			return plugin.Error("omada", err)
		}
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		o.Next = next
		return o
	})

	c.OnShutdown(func() error { cancel(); return nil })

	return nil
}

func (o *Omada) login() error {

	log.Info("logging in...")
	u := o.config.Username
	p := o.config.Password

	err := o.controller.Login(u, p)
	if err != nil {
		return err
	}

	return nil
}

func (o *Omada) controllerInit(ctx context.Context) error {

	log.Info("starting initial omada setup...")

	const retrySeconds = 15
	duration := time.Duration(retrySeconds) * time.Second

	for {

		err := o.controller.GetControllerInfo()
		if err != nil {
			if o.config.ignore_startup_errors {
				log.Warning(err)
				time.Sleep(duration)
				continue
			} else {
				return err
			}
		}

		err = o.login()
		if err != nil {
			if o.config.ignore_startup_errors {
				log.Warning(err)
				time.Sleep(duration)
				continue
			} else {
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
				time.Sleep(duration)
				continue
			} else {
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
				time.Sleep(duration)
				continue
			} else {
				return err
			}
		}

		log.Info("initial omada setup complete")
		break
	}

	go updateSessionLoop(ctx, o)
	go updateZoneLoop(ctx, o)

	return nil

}
