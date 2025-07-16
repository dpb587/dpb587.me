#!/bin/bash

set -euo pipefail

target="$( cd "${1}" && pwd )"

mkdir -p tmp

cd tools

go run ./publish/cmd/noslash/main.go "${target}" 2>&1 | tee ../tmp/noslash.log

cd "${target}"

# sync with server/main.go
find . -type f \( -name '*.css' -o -name '*.html' -o -name '*.js' -o -name '*.svg' -o -name '*.xml' \) \
    -exec gzip --keep --best {} \; \
    -exec brotli --best {} \; \
    -exec zstd -19 --quiet {} \;
