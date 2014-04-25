#!/bin/bash

# args: follow-script follow-args...

set -e

RPWD=$PWD

[ -e _build/target ] && rm -fr _build/target

git clone file://$PWD _build/target

cd _build/target

export ARTIFACT_COMMIT=`git rev-parse HEAD`
export ARTIFACT_BRANCH=`git rev-parse --abbrev-ref HEAD`

STATIC_NAME=`git rev-list -1 HEAD -- 'static/dev' | cut -c-10`

echo "--> building $ARTIFACT_BRANCH/$ARTIFACT_COMMIT..."

(
  echo 'artifact_commit:' "$ARTIFACT_COMMIT" ;
  echo 'artifact_branch:' "$ARTIFACT_BRANCH" ;
  echo 'asset_prefix: http://assets.dpb587.me/asset' ;
  echo "static_prefix: /static/$STATIC_NAME" ;
  echo 'environment: prod'
) > _config.patch.yml

jekyll build --config _config.yml,_config.patch.yml

mv _site/static/dev _site/static/$STATIC_NAME

export ARTIFACT_PATH="$RPWD/_build/target/_site"

if [[ "" != "$1" ]] ; then
    exec $@
else
    env | grep '^ARTIFACT_' | sed 's/^/export /'
fi
