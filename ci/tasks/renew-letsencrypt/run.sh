#!/bin/bash

set -eu -o pipefail

REM() { /bin/echo $( date -u +"%Y-%m-%dT%H:%M:%SZ" ) "$@"; }

REM 'starting'

REM 'renewing certificate'

bosh ssh -r -c 'sudo /var/vcap/jobs/http/bin/letsencrypt --cron' \
  | cut -f2 \
  | tee /tmp/cron

if grep -q ' + Valid till' /tmp/cron ; then
  REM 'finished'

  exit
fi

REM 'downloading certificate'

bosh ssh -r -c 'cat /var/vcap/jobs/http/etc/tls.crt' \
  | cut -f2 > /tmp/tls.crt

REM 'downloading private key'

bosh ssh -r -c 'sudo cat /var/vcap/jobs/http/etc/tls.key' \
  | cut -f2 > /tmp/tls.key

REM 'patching deployment'

jq -cn \
  --arg path "$secrets_path" \
  --arg certificate "$( cat /tmp/tls.crt )" \
  --arg private_key "$( cat /tmp/tls.key )" \
  '
    [
      {
        "type": "replace",
        "path": "\($path)/certificate",
        "value": $certificate
      },
      {
        "type": "replace",
        "path": "\($path)/private_key",
        "value": $private_key
      }
    ]
  ' \
  | bosh interpolate \
    --ops-file /dev/stdin \
    $secrets_file \
    > $secrets_file.new

mv $secrets_file.new $secrets_file

REM 'finished'
