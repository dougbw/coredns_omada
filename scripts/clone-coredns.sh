#!/bin/bash
set -euxo pipefail

release="latest"

repo_url="https://github.com/coredns/coredns.git"
release_url="https://api.github.com/repos/coredns/coredns/releases/${release}"

tag=$(curl $release_url | jq -r '.tag_name')
git clone "${repo_url}" --branch "${tag}" --depth 1