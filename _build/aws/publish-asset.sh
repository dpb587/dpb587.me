#!/bin/bash

# args: s3cmd-config

set -e

s3cmd sync \
  --config "$1" \
  --acl-public \
  --no-preserve \
  --add-header 'cache-control:max-age=86400' \
  --verbose \
  "./asset/$2" \
  "s3://dpb587-assets-us-west-2/asset/$2"
