#!/bin/bash

set -euo pipefail

asset_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "${asset_dir}"

theme_dir="$( cd ../.. && pwd)"
output_dir="${theme_dir}/static/assets/pannellum"

podman build -t tmp-pkg-pannellum .

rm -fr "${output_dir}"
mkdir -p "${output_dir}"
cd "${output_dir}"

podman rm -f tmp-pkg-pannellum || true
podman create --name tmp-pkg-pannellum tmp-pkg-pannellum
podman container cp tmp-pkg-pannellum:/result.tar - | tar -Oxf- | tar -xf-
