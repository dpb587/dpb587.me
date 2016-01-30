#!/bin/bash

set -e

curl -s -l \
  ftp://ftp.csx.cam.ac.uk/pub/software/programming/pcre/ \
  | grep -E '^pcre-.+.tar.gz$' \
  | sed -E 's/^pcre-(.+)\.tar.gz$/\1/' \
  | gsort -rV \
  | head -n1
