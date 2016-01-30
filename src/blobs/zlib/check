#!/bin/bash

set -e

git ls-remote --tags https://github.com/madler/zlib.git \
  | cut -f2 \
  | grep -Ev '\^{}' \
  | grep -E '^refs/tags/v.+$' \
  | sed -E 's/^refs\/tags\/v(.+)$/\1/' \
  | tr '_' '.' \
  | grep -v '-' \
  | gsort -rV \
  | head -n1
