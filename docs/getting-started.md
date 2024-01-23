# Getting started

CoreDNS plugins need to be compiled into CoreDNS, you can follow the [build](build.md) instructions to build the binaries or use the provided docker images.

1. Create Omada user
2. Run CoreDNS with omada plugin
3. Setup network

## 1 - Create Omada user

* Login to Omada controller
* Go to Admin > Add New Admin
* Choose "Local User"
* Role: "Viewer"
* Site Privileges: "All" or select the individual site
* Make a note of the username and password

## 2 - Setup network
1. From the Omada controller go to Settings -> Wired Networks -> LAN and choose your network(s):
2. `Domain Name` must be set (e.g `omada.home`)
3. `DNS Server` set to `Manual` and enter the IP address of your CoreDNS application.

## 3 - Run CoreDNS with omada plugin

This guide provides three alternatives on how to run CoreDNS:

- [CoreDNS binary](#coredns-binary)
- [Docker container](#docker)
- [Kubernetes](#kubernetes)

### CoreDNS binary

To build the CoreDNS binary with the `omada` plugin follow the steps provided [here](./build.md). Once built setup you `Corefile` and then run `coredns`.

Example:
```
./coredns -conf ./Corefile
```
Note: If you do not have a valid https certificate on your controller then set the `OMADA_DISABLE_HTTPS_VERIFICATION` environment variable to true

### Docker

- Use the pre-built images or build your own
- Pre-built images are published [here](https://github.com/dougbw/coredns_omada/pkgs/container/coredns_omada)
- The pre-built images contain a default corefile which *requires* the following environment variables to be set:
* `OMADA_URL`
* `OMADA_SITE`
* `OMADA_USERNAME`
* `OMADA_PASSWORD`
* `UPSTREAM_DNS`

Note: If you do not have a valid https certificate on your controller then set the `OMADA_DISABLE_HTTPS_VERIFICATION` environment variable to true

Example docker run command:
```
docker run \
--rm -it \
--expose=53 --expose=53/udp -p 53:53 -p 53:53/udp \
--env OMADA_URL="<OMADA_URL>" \
--env OMADA_SITE="<OMADA_SITE>" \
--env OMADA_USERNAME="<OMADA_USERNAME>" \
--env OMADA_PASSWORD="<OMADA_PASSWORD>" \
--env OMADA_DISABLE_HTTPS_VERIFICATION="false" \
--env UPSTREAM_DNS="8.8.8.8" \
ghcr.io/dougbw/coredns_omada:latest
```

To use a custom Corefile mount the file as a volume and add the `-conf /Corefile` command:
```
docker run \
--rm -it \
--expose=53 --expose=53/udp -p 53:53 -p 53:53/udp \
--env OMADA_URL="<OMADA_URL>" \
--env OMADA_SITE="<OMADA_SITE>" \
--env OMADA_USERNAME="<OMADA_USERNAME>" \
--env OMADA_PASSWORD="<OMADA_PASSWORD>" \
--env OMADA_DISABLE_HTTPS_VERIFICATION="false" \
--env UPSTREAM_DNS="8.8.8.8" \
-v "$PWD"/Corefile:/Corefile \
ghcr.io/dougbw/coredns_omada:latest -conf /Corefile
```

### Kubernetes
Some example manifest files to get started are in the [k8s](k8s) directory. Make sure you replace the following values:

* configmap.yaml
  * `omada-url`
  * `omada-site`
  * `omada-username`
  * `upstream-dns`
  * Note: if you do not have a valid https certification on your controller then set `omada-disable-https-verification` to `true`
* secret.yaml
  * `omada-password`



<img src="omada-network-setup.png"  width="75%">
