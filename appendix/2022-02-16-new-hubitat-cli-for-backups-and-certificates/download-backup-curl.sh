#!/bin/bash

set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

apt update
apt install -y curl openssl

# TODO non-insecure
curl -vkL --fail \
  --cookie-jar cookies.txt \
  --data username="${HUBITAT_USERNAME}" \
  --data password="${HUBITAT_PASSWORD}" \
  --data submit=Login \
  "https://${HUBITAT_HOST}/login"
curl -vk --fail \
  --cookie ./cookies.txt \
  --cookie-jar cookies.txt \
  --output /mnt/backup/latest.lzf \
  "https://${HUBITAT_HOST}/hub/backupDB?fileName=latest"
