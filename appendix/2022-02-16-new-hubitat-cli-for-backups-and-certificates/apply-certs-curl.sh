#!/bin/bash

set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

apt update
apt install -y curl openssl

comm -13 \
  <( openssl x509 -noout -serial < /mnt/config/tls.crt ) \
  <( openssl s_client -showcerts -connect ${HUBITAT_HOST}:443 </dev/null 2>/dev/null | openssl x509 -noout -serial || echo invalid ) \
  > comm.txt \
  || true
if [ ! -s comm.txt ]; then
  echo certificates are up to date
  exit
fi

# TODO non-insecure
# --form toggleSSLEnableFlag=1 \
curl -vkL --fail \
  --cookie-jar cookies.txt \
  --data username="${HUBITAT_USERNAME}" \
  --data password="${HUBITAT_PASSWORD}" \
  --data submit=Login \
  "https://${HUBITAT_HOST}/login"
curl -vk --fail \
  --cookie ./cookies.txt \
  --cookie-jar cookies.txt \
  --data-urlencode certificate@/mnt/config/tls.crt \
  --data-urlencode privateKey@/mnt/config/tls.key \
  --data _action_update='Save Certificate and Key' \
  "https://${HUBITAT_HOST}/hub/advanced/certificate/save"
curl -vk --fail \
  --cookie ./cookies.txt \
  --cookie-jar cookies.txt \
  -X POST \
  "https://${HUBITAT_HOST}/hub/reboot"