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
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.config)
		if _, err := parse(c); (err == nil) == test.expectedError {
			t.Fatalf("Unexpected errors: %v in test: %d\n\t%s", err, i, test.config)
		}
	}
}
