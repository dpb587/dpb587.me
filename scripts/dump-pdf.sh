#!/bin/bash

set -euo pipefail

/usr/local/bin/server "${PWD}" &

sleep 1

# wkhtmltopdf \
#     --grayscale \
#     --print-media-type \
#     http://localhost:8080/cv/ \
#     cv.pdf

kill %1
