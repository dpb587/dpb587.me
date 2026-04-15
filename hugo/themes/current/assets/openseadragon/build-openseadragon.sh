#!/bin/bash

set -euxo pipefail

# configure

pkg_name='openseadragon'
pkg_source_url='https://github.com/openseadragon/openseadragon/releases/download/v6.0.2/openseadragon-bin-6.0.2.tar.gz'
pkg_source_digest='3103459c626f7ae41741e8ae603136a5f852b7625fa15a93b46f0579edacbd13'

resultdir="/result"
compiledir="${TMPDIR:-/tmp}/${pkg_name}-compile-$$"
tmptarball="${compiledir}/tarball.tar.gz"
tmpworkdir="${compiledir}/workdir"

# download

mkdir -p "${tmpworkdir}"

curl -Lo "${tmptarball}" "${pkg_source_url}"
echo "${pkg_source_digest} ${tmptarball}" | sha256sum -c

tar -xzf "${tmptarball}" --strip-components=1 -C "${tmpworkdir}"

curl -fo "${tmpworkdir}/favicon.png" 'https://avatars.githubusercontent.com/u/3392630?size=64'

# install

mkdir -p "${resultdir}/${pkg_name}"

tar -cf- -C "${tmpworkdir}" \
  LICENSE.txt \
  openseadragon.min.js \
  openseadragon.min.js.map \
  favicon.png \
  images \
  | tee >( sha256sum - | cut -c-8 > "${compiledir}/digest.txt" ) \
  | tar -xf- -C "${resultdir}/${pkg_name}"

mv "${resultdir}/${pkg_name}" "${resultdir}/${pkg_name}-$( cat "${compiledir}/digest.txt" )"
ln -s "${pkg_name}-$( cat "${compiledir}/digest.txt" )" "${resultdir}/${pkg_name}-current"
