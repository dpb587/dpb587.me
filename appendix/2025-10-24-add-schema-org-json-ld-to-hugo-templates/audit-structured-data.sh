#!/bin/bash
# this file may be outdated; see scripts/audit-structured-data.sh

set -euo pipefail

cd hugo

rm -fr public
hugo build

cd public/

dovalidate() {
  set -euo pipefail

  file="${1}"

  curl --fail -sS -o- \
    'https://api.namedgraph.com/toolkit.v0/structuredData.process' \
    --header "Authorization: Bearer ${NG_API_TOKEN}" \
    --form sourceFile=@"${file}" \
    --form experimental=web.googlesearch.validator=true \
    | jq -r --arg file "${file}" '
        "\($file)\n\([
            .. | .messages? | select(.)[]
          ] | map("+++ \(.info.severity) [\(.info.title)] \(.info.message)")[]
        )"
      '
}

export -f dovalidate

grep -lsnr '<script type="application/ld+json">' . \
  | sort \
  | xargs -P8 -I{} -- bash -c 'dovalidate "{}"'
