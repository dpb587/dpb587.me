#!/bin/bash

set -e

git ls-remote --tags https://github.com/openssl/openssl.git \
  | cut -f2 \
  | grep -Ev '\^{}' \
  | grep -E '^refs/tags/OpenSSL_.+$' \
  | sed -E 's/^refs\/tags\/OpenSSL_(.+)$/\1/' \
  | tr '_' '.' \
  | grep -E '^\d+\.\d+\.\d+\w*$' \
  | gsort -rV \
  | head -n1
