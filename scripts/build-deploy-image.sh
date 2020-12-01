#!/bin/bash

set -euo pipefail

tagprefix=dpb587/dpb587.me/
repo_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."

cd "${repo_dir}"

export DOCKER_BUILDKIT=0

docker build \
  --tag dpb587/dpb587.me \
  --build-arg GITHUB_TOKEN="${GITHUB_TOKEN}" \
  --build-arg AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
  --build-arg AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
  .
