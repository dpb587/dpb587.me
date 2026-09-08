#!/bin/bash

set -euxo pipefail

# configure

pkg_name='plyr'
pkg_source_url='https://registry.npmjs.org/plyr/-/plyr-3.8.4.tgz'
pkg_source_digest='9ec5b3ec3573adbbc173b00c55f3791b983722823f11784a7433b35acd51d2d6'

theme_dir="$( cd ../.. && pwd)"
output_dir="${theme_dir}/static/assets/plyr"

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
  LICENSE.md \
  dist/plyr.css \
  dist/plyr.min.js \
  dist/plyr.svg \
  | tee >( sha256sum - | cut -c-8 > "${compiledir}/digest.txt" ) \
  | tar -xf- -C "${output_dir}/${pkg_name}"

mv "${output_dir}/${pkg_name}" "${output_dir}/${pkg_name}-$( cat "${compiledir}/digest.txt" )"
ln -s "${pkg_name}-$( cat "${compiledir}/digest.txt" )" "${output_dir}/${pkg_name}-current"
