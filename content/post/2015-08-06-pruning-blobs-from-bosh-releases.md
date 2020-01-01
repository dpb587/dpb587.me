---
date: 2015-08-06
title: "Pruning Blobs from BOSH Releases"
description: "Avoiding unnecessary disk usage for old, unneeded package files."
tags:
- blobs
- bosh
- cleanup
- packages
- pruning
aliases:
- /blog/2015/08/06/pruning-blobs-from-bosh-releases.html
---

Over time, as blobs are continually added to [BOSH][1] releases, the files can start consuming lots of disk space. Blobs are frequently abandoned because newer versions replace them, or sometimes the original packages referencing them are removed. Unfortunately, freeing the disk space isn't as simple as `rm blobs/elasticsearch-1.5.2.tar.gz` because BOSH keeps track of blobs in the `config/blobs.yml` file and uses symlinks to cached copies.

To help keep a lean workspace, I remove references to blobs which are no longer needed in my release. The blobs remain untouched in the blobstore/S3, but as far as my local `bosh` command cares about, it doesn't need to keep local copies. One option for pruning is to manually edit `config/blobs.yml` and remove the old references (and then run `bosh sync blobs` to update `blobs/`). However, I tend to go the other direction - interactively or with shell scripts - removing files from `blobs/` and then updating `blobs.yml` with this command...

    for FILE in $( grep -E '^[^ ].+:$' config/blobs.yml | tr -d ':' ) ; do
      [ -e "blobs/${FILE}" ] || sed -i '' -E -e "\\#^${FILE}:\$#{N;N;N;d;}" config/blobs.yml
    done

Once they're gone from `blobs.yml` I can commit the changes and know that the next time I need to clone/sync into a new workspace it'll be faster.

    git commit -om 'Prune old blob references' config/blobs.yml

But... while those blobs are no longer listed in `config/blobs.yml` and they are no longer in `blobs/`, the blob still exists in the `.blobs` directory where `bosh` keeps an original copy. I can remove unreferenced blobs from `.blobs` with this command...

    for BLOBSHA in $( find .blobs -type f ) ; do
      grep -qE "^  sha:\s+$( basename $BLOBSHA )" config/blobs.yml || rm "$BLOBSHA"
    done

Even though the blobs are now effectively gone, their references still exist in repository history. For example, if you wanted to rebuild your `.blobs` cache directory you could loop through changes to `blobs.yml` and rerun `bosh sync blobs` to restore local copies...

    for COMMIT in $( git rev-list --parents HEAD -- config/blobs.yml | cut -d" " -f1 ; git rev-parse HEAD ) ; do
      git checkout "$COMMIT" config/blobs.yml
      bosh sync blobs
    done

As an example, here's a before and after of cleaning up blobs in my long-running [logsearch-boshrelease][2] workspace...

    $ du -sh .blobs/ | cut -f1
    904M
    ...cleanup...
    $ du -sh .blobs/ | cut -f1
    168M


 [1]: http://bosh.io/
 [2]: https://github.com/logsearch/logsearch-boshrelease
