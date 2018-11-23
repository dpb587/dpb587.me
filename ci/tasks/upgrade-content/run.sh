#!/bin/bash

set -eu

export GIT_COMMITTER_NAME="Concourse"
export GIT_COMMITTER_EMAIL="concourse.ci@localhost"

git config --global user.email "${git_user_email:-ci@localhost}"
git config --global user.name "${git_user_name:-CI Bot}"

git clone file://$PWD/repo-input repo

cd content-repo

content_commit=$( git rev-parse HEAD )

cd ../repo

git submodule update --init

cd src/content

git checkout $content_commit

cd ../..

git add src/content

if git diff --staged --exit-code --quiet ; then
  # no changes pending
  exit
fi

git commit -m "src: Update content to $( echo "$content_commit" | cut -c-7 )"
