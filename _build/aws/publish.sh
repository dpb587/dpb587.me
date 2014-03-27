#!/bin/bash

# args: s3cmd-config

set -e

s3cmd sync \
  --config "$1" \
  --acl-public \
  --no-delete-removed \
  --no-preserve \
  --exclude 'private/*' \
  --exclude 'asset/*' \
  --add-header 'cache-control:max-age=3600' \
  --verbose \
  "$ARTIFACT_PATH" \
  s3://dpb587-us-west-2
