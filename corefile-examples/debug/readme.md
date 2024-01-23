# Corefile example - debug/log

This example enables query logging and debug logging (using the [log](https://coredns.io/plugins/log/) and [debug](https://coredns.io/plugins/debug/) plugins). Can be useful for troubleshooting but not reccomended for typical use as its very noisy.

- `debug` will enable debug logging which will include debug logs from the coredns_omada plugin
- `log` will enable query/response logging for queries which are forwarded to the upstream dns server