#!/bin/bash

set -euo pipefail

cd "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ ! -e assets/heroicons/src ]
then
    wget -O- https://github.com/tailwindlabs/heroicons/archive/refs/tags/v2.2.0.tar.gz \
        | tar -xzf- -C assets/heroicons --strip-components=1 heroicons-2.2.0/src heroicons-2.2.0/LICENSE
fi

rm -fr static/assets/styles
mkdir -p static/assets/styles
tailwindcss -i assets/tailwindcss/main.css -o ./static/assets/styles/main.development.css
tailwindcss -i assets/tailwindcss/main.css -o ./static/assets/styles/main.css --minify --optimize

cat <<EOF > assets/tailwindcss.build.json
{
    "fingerprint": "$( find ./static/assets/styles -type f -exec cat {} \; | sha256sum | cut -d' ' -f1 | cut -c1-10 )",
    "fileIntegrity": {
        "main.development.css": "sha384-$(sha384sum ./static/assets/styles/main.development.css | cut -d' ' -f1 | xxd -r -p | base64 )",
        "main.css": "sha384-$(sha384sum ./static/assets/styles/main.css | cut -d' ' -f1 | xxd -r -p | base64 )"
    }
}
EOF

# ./assets/openseadragon/build.sh

pushd assets/svelte > /dev/null

npm run build

popd > /dev/null
