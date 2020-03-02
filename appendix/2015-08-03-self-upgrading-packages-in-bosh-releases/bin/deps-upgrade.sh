#!/bin/bash

# args: package-name dep-name

set -e
set -u

export PACKAGE_NAME="${1}"
export DEP_NAME="${2}"

DEP_DIR="${PWD}/packages/${PACKAGE_NAME}/deps/${DEP_NAME}"
DEP_BLOB_DIR="${PWD}/blobs/${PACKAGE_NAME}-blobs/${DEP_NAME}"


echo "==> ${PACKAGE_NAME}/${DEP_NAME}"


if [ -f "${DEP_DIR}/VERSION" ] ; then
  VERSION_LOCAL=$( cat "${DEP_DIR}/VERSION" )
else
  VERSION_LOCAL=missing
fi

echo "--| local ${VERSION_LOCAL}"


VERSION_CHECK=$( . "${DEP_DIR}/check" )

echo "--| check ${VERSION_CHECK}"


if [[ "${VERSION_CHECK}" == "${VERSION_LOCAL}" ]] ; then
  exit
fi


echo "--> fetching new version"

rm -fr "${DEP_BLOB_DIR}-new"
mkdir -p "${DEP_BLOB_DIR}-new"

cd "${DEP_BLOB_DIR}-new"

export VERSION="${VERSION_CHECK}"
"${DEP_DIR}/get"

rm -fr "${DEP_BLOB_DIR}"
mv "${DEP_BLOB_DIR}-new" "${DEP_BLOB_DIR}"

echo "${VERSION}" > "${DEP_DIR}/VERSION"


echo "-->" $( du -sh ${DEP_BLOB_DIR} | cut -f1 )
