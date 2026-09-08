#!/bin/bash

set -euxo pipefail

# configure

pkg_name='hlsjs'
pkg_source_url='https://registry.npmjs.org/hls.js/-/hls.js-1.7.1.tgz'
pkg_source_digest='700a82512a9d90e9d18ba6725708679cea20a7e25368602df1a68a686b100750'

theme_dir="$( cd ../.. && pwd)"
output_dir="${theme_dir}/static/assets/hlsjs"

rm -fr "${output_dir}"

compiledir="${TMPDIR:-/tmp}/${pkg_name}-compile-$$"
tmptarball="${compiledir}/tarball.tar.gz"
tmpworkdir="${compiledir}/workdir"

# download

mkdir -p "${tmpworkdir}"

curl -Lo "${tmptarball}" "${pkg_source_url}"
echo "${pkg_source_digest} ${tmptarball}" | sha256sum -c

tar -xzf "${tmptarball}" --strip-components=1 -C "${tmpworkdir}"

# install

mkdir -p "${output_dir}/${pkg_name}"

tar -cf- -C "${tmpworkdir}" \
  LICENSE \
  dist/hls.min.js \
  | tee >( sha256sum - | cut -c-8 > "${compiledir}/digest.txt" ) \
  | tar -xf- -C "${output_dir}/${pkg_name}"

mv "${output_dir}/${pkg_name}" "${output_dir}/${pkg_name}-$( cat "${compiledir}/digest.txt" )"
ln -s "${pkg_name}-$( cat "${compiledir}/digest.txt" )" "${output_dir}/${pkg_name}-current"
