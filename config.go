package coredns_omada

import (
	"strconv"
	"time"

	"github.com/coredns/caddy"
	"github.com/go-playground/validator/v10"
)

type config struct {
	Controller_url string `validate:"required,url"`
	Site           string `validate:"required"`
	Username       string `validate:"required"`
	Password       string `validate:"required"`

	refresh_minutes           int           // update dns zones every x minutes
	refresh_login_hours       int           // login and get a new session token every x hours
	resolve_clients           bool          // resolve 'client' addresses
	resolve_devices           bool          // resolve 'device' addresses
	resolve_dhcp_reservations bool          // resolve static 'dhcp reservations'
	stale_record_duration     time.Duration // resolve static 'dhcp reservations'
}

func parse(c *caddy.Controller) (config config, err error) {

	// defaults
	config.refresh_minutes = 1
	config.refresh_login_hours = 24
	config.resolve_clients = true
	config.resolve_devices = true
	config.resolve_dhcp_reservations = true
	config.stale_record_duration, _ = time.ParseDuration("5m")

	for c.Next() {

		for c.NextBlock() {
			switch c.Val() {

			case "controller_url":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.Controller_url = c.Val()

			case "site":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.Site = c.Val()

			case "username":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.Username = c.Val()

			case "password":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.Password = c.Val()

			case "refresh_minutes":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.refresh_minutes, err = strconv.Atoi(c.Val())
				if err != nil {
					return config, c.ArgErr()
				}

			case "refresh_login_hours":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.refresh_login_hours, err = strconv.Atoi(c.Val())
				if err != nil {
					return config, c.ArgErr()
				}

			case "resolve_clients":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.resolve_clients, err = strconv.ParseBool(c.Val())
				if err != nil {
					return config, c.ArgErr()
				}

			case "resolve_devices":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.resolve_devices, err = strconv.ParseBool(c.Val())
				if err != nil {
					return config, c.ArgErr()
				}

			case "resolve_dhcp_reservations":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.resolve_dhcp_reservations, err = strconv.ParseBool(c.Val())
				if err != nil {
					return config, c.ArgErr()
				}

			case "stale_record_duration":
				if !c.NextArg() {
					return config, c.ArgErr()
				}
				config.stale_record_duration, err = time.ParseDuration(c.Val())
				if err != nil {
					return config, c.ArgErr()
				}

			default:
				return config, c.Errf("unknown property: %q", c.Val())
			}

		}

	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		log.Info("There is a Corefile configuration error:")
		return config, err
	}

	return config, nil

}
