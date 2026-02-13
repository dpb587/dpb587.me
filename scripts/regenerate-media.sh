#!/bin/bash

set -euo pipefail

mediadir="${1}"

while read -r file
do
    filedirname="$( dirname "${file}" )"
    filesuffixless="$( basename "${file}" | sed 's/\.[^.]*$//' )"
    fileext="$( echo "${file}" | sed 's/.*\(\.[^.]*\)$/\1/' )"

    mkdir -p "${filedirname}"
    vipsthumbnail \
        "private/${file}" \
        -o "$PWD/${filedirname}/${filesuffixless}${fileext}" \
        -s 1720x1720
done < <(
    cd private
    find "${mediadir}" -type f
)
