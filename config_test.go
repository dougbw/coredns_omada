package coredns_omada

import (
	"testing"

	"github.com/coredns/caddy"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		config        string
		expectedError bool
	}{
		// valid config
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			refresh_minutes 1
			resolve_clients true
			resolve_devices true
			resolve_dhcp_reservations true
			stale_record_duration 10m
			fallthrough
}`, false},

		// missing required property: controller url
		{`omada {
			username test
			password test
			site .*
}`, true},

		// missing required property: username
		{`omada {
			controller_url https://10.0.0.1
			password test
			site .*
}`, true},

		// missing required property: password
		{`omada {
			controller_url https://10.0.0.1
			username test
			site .*
}`, true},

		// missing required property: site
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
}`, true},

		// unexpected key
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			unexpected error
}`, true},

		// invalid value: refresh_minutes
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			refresh_minutes test
}`, true},

		// invalid value: refresh_login_hours
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			refresh_login_hours test
}`, true},

		// invalid value: resolve_clients
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			resolve_clients error
}`, true},

		// invalid value: resolve_devices
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			resolve_devices error
}`, true},

		// invalid value: resolve_dhcp_reservations
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			resolve_dhcp_reservations error
}`, true},

		// invalid value: stale_record_duration
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			stale_record_duration error
}`, true},

		// invalid value: ignore_startup_errors
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			ignore_startup_errors zzz
}`, true},

		// valid: network_domain and default_domain
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			network_domain Default home.arpa
			network_domain 10.0.20.0/24 office.home.arpa.
			default_domain other.home.arpa
}`, false},

		// invalid value: network_domain missing domain
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			network_domain Default
}`, true},

		// invalid value: network_domain bad subnet
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			network_domain 10.0.20.0/33 home.arpa
}`, true},

		// invalid value: network_domain bad domain
		{`omada {
			controller_url https://10.0.0.1
			username test
			password test
			site .*
			network_domain Default .
}`, true},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.config)
		if _, err := parse(c); (err == nil) == test.expectedError {
			t.Fatalf("Unexpected errors: %v in test: %d\n\t%s", err, i, test.config)
		}
	}
}
