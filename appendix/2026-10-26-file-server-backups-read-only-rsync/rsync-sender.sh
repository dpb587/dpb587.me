#!/bin/sh

set -eu

cmd="${SSH_ORIGINAL_COMMAND:-}"

case "${cmd}" in
  *[\;\&\|\`\$\(\)\<\>]*|*"
"*)
    ;;
  "rsync --server --sender "*)
    exec /bin/sh -c "${cmd}"
    ;;
esac

echo "error: only read-only rsync is permitted" >&2
exit 1
