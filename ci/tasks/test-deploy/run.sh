#!/bin/bash

set -eu

fail () { echo "FAILURE: $1" >&2 ; exit 1 ; }

cd repo

start-bosh

source /tmp/local-bosh/director/env

if [ -e stemcell/*.tgz ]; then
  bosh upload-stemcell stemcell/*.tgz
  os=$( tar -Oxzf stemcell/*.tgz stemcell.MF | grep '^operating_system: ' | awk '{ print $2 }' )
else
  bosh upload-stemcell \
    --name=bosh-warden-boshlite-ubuntu-trusty-go_agent \
    --version=3586.57 \
    --sha1=9aca8b9484e9ca7095077d51d6af129698c9fab1 \
    https://s3.amazonaws.com/bosh-core-stemcells/warden/bosh-stemcell-3586.57-warden-boshlite-ubuntu-trusty-go_agent.tgz
  os=ubuntu-trusty
fi

export BOSH_DEPLOYMENT=test-deploy

bosh -n deploy \
  --var os="$os" \
  --var repo_dir="$PWD" \
  --vars-store=/tmp/deployment-vars.yml \
  ci/tasks/test-deploy/deployment.yml


#
# simple page test
#

bosh ssh role1/0 '
  set -e
  curl --fail http://localhost:8080/about/ | grep "Danny Berger"
'


#
# teardown
#

bosh -n delete-deployment


#
# stop-bosh
#

bosh -n clean-up --all

bosh delete-env "/tmp/local-bosh/director/bosh-director.yml" \
  --vars-store="/tmp/local-bosh/director/creds.yml" \
  --state="/tmp/local-bosh/director/state.json"
