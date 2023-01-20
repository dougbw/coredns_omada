# Build

CoreDNS plugins need to be compiled into the main CoreDNS binary, with some minor changes to include the plugin.

* clone the main [coredns](https://github.com/coredns/coredns) repo 
* clone [coredns_omada](https://github.com/dougbw/coredns_omada) repo
* both repos should share the same parent directory:
```bash
user@pc:/repos$ tree
.
├── coredns
└── coredns_omada
```
* in `coredns/plugin.cfg` add the following line to the START of the file:
    ```omada:github.com/dougbw/coredns_omada```
* (optional for local development only): in `coredns/go.mod` add the following line to the END of the file. The following example assumes you have the `coredns_omada` repo located at `/repos/coredns_omada` so this needs to match your system.
```
replace github.com/dougbw/coredns_omada => /repos/coredns_omada
```
* run `make` from the `coredns` repo
* grab the `coredns` binary and add your `Corefile` configuration file

# Build instructions (docker image)

## multi-platform:
See [setting up docker for multi-platform builds](docs/docker-multi-platform.md)
```
docker buildx build --platform linux/amd64,linux/arm64 -t coredns-omada .
```

## linux/amd64 only:
```
docker buildx build --platform linux/amd64 -t coredns-omada .
```

## linux/arm64 only:
```
docker buildx build --platform linux/arm64 -t coredns-omada .
```