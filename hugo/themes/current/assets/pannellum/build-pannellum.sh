#!/bin/bash

set -euxo pipefail

# configure

pkg_name='pannellum'
pkg_source_url='https://github.com/mpetroff/pannellum/releases/download/2.5.7/pannellum-2.5.7.zip'
pkg_source_digest='de768a7999b65a890a889c85b923677da77df7c06900ae1a1493a24671f314c3'

resultdir="/result"
compiledir="${TMPDIR:-/tmp}/${pkg_name}-compile-$$"
tmptarball="${compiledir}/tarball.zip"
tmpworkdir="${compiledir}/workdir"

# download

mkdir -p "${tmpworkdir}"

curl -Lo "${tmptarball}" "${pkg_source_url}"
echo "${pkg_source_digest} ${tmptarball}" | sha256sum -c

unzip "${tmptarball}" -d "${compiledir}/extracted"
mv "${compiledir}/extracted/${pkg_name}/"* "${tmpworkdir}/"

# install

mkdir -p "${resultdir}/${pkg_name}"

tar -cf- -C "${tmpworkdir}" \
  COPYING \
  pannellum.js \
  pannellum.css \
  | tee >( sha256sum - | cut -c-8 > "${compiledir}/digest.txt" ) \
  | tar -xf- -C "${resultdir}/${pkg_name}"

mv "${resultdir}/${pkg_name}" "${resultdir}/${pkg_name}-$( cat "${compiledir}/digest.txt" )"
ln -s "${pkg_name}-$( cat "${compiledir}/digest.txt" )" "${resultdir}/${pkg_name}-current"
