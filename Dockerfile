# build command: 
# docker buildx build --platform linux/amd64,linux/arm64 -t coredns-omada --load
#
# push command:
# docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/dougbw/coredns_omada:1.4.3 -t ghcr.io/dougbw/coredns_omada:latest --push .
#
# How to setup multi platform builder:
# docker buildx create --name multiplatform
# docker buildx use multiplatform
# docker buildx inspect --bootstrap

FROM --platform=$BUILDPLATFORM golang:1.21-bookworm as builder
ARG TARGETOS TARGETARCH
RUN apt update
RUN apt install git curl jq -y
COPY . /coredns_omada
WORKDIR /

# clone latest coredns release
RUN /coredns_omada/scripts/clone-coredns.sh

# insert plugin config
RUN sed -i '/^forward:forward$/i omada:github.com/dougbw/coredns_omada' /coredns/plugin.cfg
RUN echo "replace github.com/dougbw/coredns_omada => /coredns_omada" >> /coredns/go.mod

# compile coredns
WORKDIR /coredns
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make

FROM --platform=$BUILDPLATFORM debian:stable-slim as certificates
RUN export DEBCONF_NONINTERACTIVE_SEEN=true \
           DEBIAN_FRONTEND=noninteractive \
           DEBIAN_PRIORITY=critical \
           TERM=linux ; \
    apt-get -qq update ; \
    apt-get -yyqq upgrade ; \
    apt-get -yyqq install ca-certificates ; \
    apt-get clean

FROM --platform=$TARGETPLATFORM scratch
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /coredns/coredns /coredns
COPY Corefile /Corefile
EXPOSE 53 53/udp
ENTRYPOINT ["/coredns"]