#!/bin/bash

set -e


#
# Run through all package deps
#

for PACKAGE_NAME in $( cd packages ; find . -type d -depth 1 | cut -c3- ) ; do
  [ -e "packages/${PACKAGE_NAME}/deps" ] || continue

  for DEP_NAME in $( cd "packages/${PACKAGE_NAME}/deps" ; find . -type d -depth 1 | cut -c3- ) ; do
    ./bin/deps-upgrade "${PACKAGE_NAME}" "${DEP_NAME}"
  done
done


#
# Check if we actually updated anything
#

DIFFS=$( echo $( git diff HEAD --name-only | wc -l ) )

if [ 0 -eq $DIFFS ] ; then
  # no changes; nothing else to do
  exit
elif [ 1 -eq $DIFFS ] ; then
  COMMIT_HEADER="Upgraded 1 package dependency"
else
  COMMIT_HEADER="Upgraded ${DIFFS} package dependencies"
fi


#
# Upload new blobs
#

bosh -n upload blobs

git add config/blobs.yml


#
# Generate the commit message
#

(
  echo "${COMMIT_HEADER}"

  PACKAGE_CURR=""

  for VERSION_PATH in $(
    git diff HEAD --name-only \
      | grep -E "^packages/([^/]+)/deps/([^/]+)/VERSION$" \
      | sort
  ) ; do
    git add "${VERSION_PATH}"

    DEP_NAME=$( basename $( dirname "${VERSION_PATH}" ) )
    PACKAGE_NAME=$( basename $( dirname $( dirname $( dirname "${VERSION_PATH}" ) ) ) )

    if [[ "$PACKAGE_CURR" != "$PACKAGE_NAME" ]] ; then
      echo ""
      echo "${PACKAGE_NAME}"
      echo ""
    
      PACKAGE_CURR="${PACKAGE_NAME}"
    fi

    VERSION_OLD=$( git show HEAD:$VERSION_PATH )
    VERSION_NEW=$( cat "${VERSION_PATH}" )

    echo " * ${DEP_NAME} now ${VERSION_NEW} (was ${VERSION_OLD})"
  done
) > commit.msg


#
# Create the commit
#

git commit -F commit.msg


#
# Cleanup after ourselves
#

rm commit.msg
