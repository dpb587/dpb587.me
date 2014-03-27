#!/bin/bash

# args: s3cmd-config

set -e

cd "$ARTIFACT_PATH"

cd static/

s3cmd sync \
  --config "$1" \
  --acl-public \
  --no-delete-removed \
  --no-preserve \
  --add-header 'cache-control:max-age=604800' \
  --verbose \
  . \
  s3://dpb587-us-west-2/static/

cd ../

s3cmd sync \
  --config "$1" \
  --acl-public \
  --no-delete-removed \
  --no-preserve \
  --exclude 'private/*' \
  --exclude 'asset/*' \
  --exclude 'static/*' \
  --add-header 'cache-control:max-age=3600' \
  --verbose \
  . \
  s3://dpb587-us-west-2
