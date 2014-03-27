#!/bin/bash

# args: follow-script follow-args...

set -e

[ -e _build/target ] && rm -fr _build/target

mkdir _build/target

export ARTIFACT_COMMIT=`git rev-parse HEAD`
export ARTIFACT_BRANCH=`git rev-parse --abbrev-ref HEAD`

echo "--> building $ARTIFACT_BRANCH/$ARTIFACT_COMMIT..."

(
  echo 'artifact_commit:' "$ARTIFACT_COMMIT" ;
  echo 'artifact_branch:' "$ARTIFACT_BRANCH" ;
  echo 'destination: _build/target/artifact' ;
  echo 'asset_prefix: //assets.dpb587.me/asset' ;
  echo 'environment: prod'
) > _build/target/_config.yml

jekyll build --config _config.yml,_build/target/_config.yml

mv _build/target/artifact/static/dev _build/target/artifact/static/`echo $ARTIFACT_COMMIT | cut -c-10`

export ARTIFACT_PATH="$PWD/_build/target/artifact"

if [[ "" != "$1" ]] ; then
    exec $@
else
    env | grep '^ARTIFACT_' | sed 's/^/export /'
fi
 