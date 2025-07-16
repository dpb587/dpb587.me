#!/bin/bash

set -e

curl -s -l \
  ftp://xmlsoft.org/libxml2/ \
  | grep -E '^libxml2-.+.tar.gz$' \
  | sed -E 's/^libxml2-(.+)\.tar.gz$/\1/' \
  | grep -E '^\d+\.\d+\.\d+\w*$' \
  | gsort -rV \
  | head -n1
