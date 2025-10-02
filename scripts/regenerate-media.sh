#!/bin/bash

set -euo pipefail

mediadir="${1}"

while read -r file
do
    filedirname="$( dirname "${file}" )"
    filesuffixless="$( basename "${file}" | sed 's/\.[^.]*$//' )"

    mkdir -p "${filedirname}"
    vipsthumbnail \
        "private/${file}" \
        -o "$PWD/${filedirname}/${filesuffixless}.png" \
        -s 1024x1024
done < <(
    cd private
    find "${mediadir}" -type f
)
