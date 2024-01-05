# Configuration

CoreDNS is configured using a configuration file called a [Corefile](https://coredns.io/2017/07/23/corefile-explained/) which supports [variable substitution](https://coredns.io/manual/configuration/#environment-variables) so values can be provided using environment variables.

## omada plugin configuration

| Name                | Required | Type   | Notes                                                                    |
| ------------------- | -------- | ------ | ------------------------------------------------------------------------ |
| controller_url      | ✅       | string | address of the Omada controller. Include `https://` prefix               |
| site                | ✅       | string | name of the site from the Omada controller (note this is a regex pattern) |
| username            | ✅       | string | Omada controller username                                                |
| password            | ✅       | string | Omada controller password                                                |
| refresh_minutes     | ❌       | int    | how often to refresh the zones (default 1 minute)                        |
| refresh_login_hours | ❌       | int    | how often to refresh the login token (default 24 hours)                  |

## Credentials

For this service you should create a new user in the `Admin` page of the controller with a `Viewer` role.

## Omada Site

A single Omada controller can support multiple network sites. This plugin can be configured to use multiple sites via the `site` configuration property (regex). Multiple sites can be specified using the `|` separator like this `SiteA|SiteB|SiteC` or all sites can be selected by setting it to `.*`

## HTTPS Verification

This will depend on your network and configuration, but due to the lack of a suitable internal DNS resolution you may need to disable HTTPS verification to the controller, as even if you have a valid certificate on your controller you need a valid DNS record pointing to your controller where coredns is running.

HTTPS verification can be disabled by setting environment variable `OMADA_DISABLE_HTTPS_VERIFICATION` to `true`

An option to keep HTTPS verification enabled is to create a public DNS A record pointing to your controllers private IP address.

## Corefile example

See [Corefile](../Corefile)

```
. {
    health :8080
    omada {
        controller_url {$OMADA_URL}
        site {$OMADA_SITE}
        username {$OMADA_USERNAME}
        password {$OMADA_PASSWORD}
        refresh_minutes 1
    }
    forward . {$UPSTREAM_DNS}
}
```

### Enable debug logging

- `debug` will enable debug logging which will include debug logs from the omada plugin
- `log` will enable query/response logging for queries which are forwarded to the upstream dns server

```
. {
    log
    debug
    health :8080
    omada {
        controller_url {$OMADA_URL}
        site {$OMADA_SITE}
        username {$OMADA_USERNAME}
        password {$OMADA_PASSWORD}
        refresh_minutes 1
    }
    forward . {$UPSTREAM_DNS}
}
```
