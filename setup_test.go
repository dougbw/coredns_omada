package coredns_omada

import (
	"fmt"
	"testing"

	"github.com/coredns/caddy"
)

func TestSetup(t *testing.T) {

	testServer := setupTestServer()
	defer testServer.Close()
	url := testServer.URL

	tests := []struct {
		config        string
		expectedError bool
	}{
		// valid config with mock server
		{fmt.Sprintf(`omada {
			controller_url %s
			username test
			password test
			site .*
		}`, url), false},

		// there will be no response from the controller url
		{`omada {
			controller_url http://localhost
			username test
			password test
			site .*
		}`, true},

		// invalid config: missing username
		{`omada {
			controller_url http://localhost
			password test
			site .*
		}`, true},

		// ignore connection errors to omada controller on startup
		{`omada {
			controller_url http://localhost:8888
			username test
			password test
			site .*
			ignore_startup_errors true
		}`, false},

		// do not ignore connection errors to omada controller on startup
		{`omada {
			controller_url http://localhost:8888
			username test
			password test
			site .*
			ignore_startup_errors false
		}`, true},
	}

	for i, test := range tests {
		caddy := caddy.NewTestController("dns", test.config)
		if err := setup(caddy); (err == nil) == test.expectedError {
			t.Fatalf("Unexpected errors: %v in test: %d\n\t%s", err, i, test.config)
		}
	}
}
