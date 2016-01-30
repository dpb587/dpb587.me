#!/bin/bash

set -e
set -u

cd repository/

bundle install --jobs=$( grep -c ^processor /proc/cpuinfo ) --path=vendor

ARTIFACT_COMMIT=$( git rev-parse HEAD )
VERSION_DATE=1.$( date +%Y%m%d ).$( date +%H%M%S | sed -E 's/^0+//' )

STATIC_NAME=$( git rev-list -1 HEAD -- 'static/dev' | cut -c-10 )

echo "--> building $ARTIFACT_COMMIT..."

(
  echo 'artifact_commit:' "$ARTIFACT_COMMIT" ;
  echo "static_prefix: /static/$STATIC_NAME" ;
  echo "environment: $SITE_ENVIRONMENT" ;
  echo "url: \"$SITE_URL\""
) > _config.patch.yml

bundle exec jekyll build --config _config.yml,_config.patch.yml

mv _site/static/dev _site/static/$STATIC_NAME

mv _site ../docroot/docroot

cd ../docroot
tar -czf $ARTIFACT_COMMIT-$VERSION_DATE.tgz docroot
