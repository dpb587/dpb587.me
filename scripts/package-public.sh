#!/bin/bash

set -euxo pipefail

: "${1:?missing TARGET-DIR}"

target="$( cd "${1}" && pwd )"

cd "${target}"

if grep -Esnr 'x0pqdrbrnr4h(/|%2F)' .
then
    echo "error: found non-canonical page link"

    exit 1
elif grep -Esnr "black-canyon-of-the-gunnison-park(/|%2F)(<|'|\")" .
then
    echo "error: found non-canonical page link"

    exit 1
fi

# sync with server/main.go
find . -type f \( -name '*.css' -o -name '*.html' -o -name '*.js' -o -name '*.svg' -o -name '*.xml' \) \
    -exec gzip --keep --best {} \; \
    -exec brotli --best {} \; \
    -exec zstd -19 --quiet {} \;
