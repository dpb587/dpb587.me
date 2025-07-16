#!/bin/bash

set -euo pipefail

asset_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "${asset_dir}"

theme_dir="$( cd ../.. && pwd)"
output_dir="${theme_dir}/static/assets/openseadragon"

podman build -t tmp-pkg-openseadragon .

rm -fr "${output_dir}"
mkdir -p "${output_dir}"
cd "${output_dir}"

podman rm -f tmp-pkg-openseadragon || true
podman create --name tmp-pkg-openseadragon tmp-pkg-openseadragon
podman container cp tmp-pkg-openseadragon:/result.tar - | tar -Oxf- | tar -xf-
