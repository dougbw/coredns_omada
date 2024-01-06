# coredns-omada-plugin

coredns_omada is a [CoreDNS plugin](https://coredns.io/manual/plugins/) which resolves local DNS addresses for clients on TP-Link Omada SDN networks. It uses the omada API to periodically get a list of client addresses.

* [Getting started](docs/getting-started.md)
* [Configuration](docs/configuration.md)
* [Building coredns omada](docs/build.md)
* [Solution design](docs/solution-design.md)

# Pre-built docker images

Docker container images are now being published to the [GitHub Container Registry](https://github.com/dougbw/coredns_omada/pkgs/container/coredns_omada) under the following name:
 `ghcr.io/dougbw/coredns_omada`

# Version chart

| Omada Controller version  | Reccomended coredns_omada version |
| --------                  | -------               |
| `5.12.x`                  | `v1.5.0` or newer     |
| `5.9.x`                   | `v1.4.2`              |
