#!/bin/bash

set -euxo pipefail

# configure

pkg_name='openseadragon-flat-toolbar-icons'
pkg_source_url='https://github.com/peterthomet/openseadragon-flat-toolbar-icons/tarball/1d32812051ab512f5b05f2b8a179f2cdc88c6db0'
pkg_source_digest='e078d9dfdfc5ecbe192181dd78fd57be180cbba575e169511bd4d9a5e01f895b'

resultdir="/result"
compiledir="${TMPDIR:-/tmp}/${pkg_name}-compile-$$"
tmptarball="${compiledir}/tarball.tar.gz"
tmpworkdir="${compiledir}/workdir"

# download

mkdir -p "${tmpworkdir}"

curl -Lo "${tmptarball}" "${pkg_source_url}"
echo "${pkg_source_digest} ${tmptarball}" | sha256sum -c

tar -xzf "${tmptarball}" --strip-components=1 -C "${tmpworkdir}"

# install

mkdir -p "${resultdir}/${pkg_name}"

tar -cf- -C "${tmpworkdir}" \
  images \
  LICENSE.txt \
  | tee >( sha256sum - | cut -c-8 > "${compiledir}/digest.txt" ) \
  | tar -xf- -C "${resultdir}/${pkg_name}"

mv "${resultdir}/${pkg_name}" "${resultdir}/${pkg_name}-$( cat "${compiledir}/digest.txt" )"
ln -s "${pkg_name}-$( cat "${compiledir}/digest.txt" )" "${resultdir}/${pkg_name}-current"
