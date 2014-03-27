#!/bin/bash

# args: s3cmd-config asset-path

set -e

cd "./asset/$2"

s3cmd sync \
  --config "$1" \
  --acl-public \
  --no-preserve \
  --add-header 'cache-control:max-age=86400' \
  --verbose \
  . \
  "s3://assets.dpb587.me/asset/$2/"
