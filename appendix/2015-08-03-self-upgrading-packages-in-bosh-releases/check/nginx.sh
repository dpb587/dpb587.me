#!/bin/bash

set -e

wget \
  -q \
  -O- \
  http://hg.nginx.org/nginx/tags?style=raw \
  | cut -f1 \
  | grep '^release-' \
  | sed -E 's/^release-(.+)$/\1/' \
  | gsort -rV \
  | head -n1
