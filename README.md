# coredns-omada-plugin

coredns_omada is a [CoreDNS plugin](https://coredns.io/manual/plugins/) which resolves local DNS addresses for clients on TP-Link Omada SDN networks. It uses the omada API to periodically get a list of client addresses.

* [Getting started](docs/getting-started.md)
* [Configuration](docs/configuration.md)
* [Building coredns omada](docs/build.md)
* [Solution design](docs/solution-design.md)

# Pre-build docker images

Docker container images are now being published to (GitHub)[https://github.com/dougbw/coredns_omada/pkgs/container/corends_omada] under the following name:
 `ghcr.io/dougbw/corends_omada`

# Version chart

| Omada Controller version  | coredns_omada version |
| --------                  | -------               |
| `5.12.x`                  | `v1.4.3`              |
| `5.9.x`                   | `v1.4.2`              |