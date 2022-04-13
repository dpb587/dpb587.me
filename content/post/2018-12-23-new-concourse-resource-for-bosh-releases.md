---
date: 2018-12-23
title: New Concourse Resource for BOSH Releases
description: Automating tarball creation and publishing of new versions.
projects:
- bosh-release-resource
tags:
- automation
- bosh
- bosh-release
- concourse
- concourse-resource
aliases:
- /blog/2018/12/23/new-concourse-resource-for-bosh-releases.html
---

As a "continuous thing-doer" [Concourse](https://concourse-ci.org/) is great for documenting workflows and making sure they run. One of the workflows I frequently automate is consuming and publishing BOSH releases. Existing resources had some shortcomings for my needs, so I created the [`bosh-release` resource](https://github.com/dpb587/bosh-release-resource) to support those workflows. This post discusses more of the background and decisions that went into the resource.

<!--more-->


## Prior Workflows

For public BOSH releases which are listed on [bosh.io](https://bosh.io/), the [`bosh-io-release` resource](https://github.com/concourse/bosh-io-release-resource) works very well. It allows pipelines to easily download and use finalized versions of the releases by referring to a repository name (e.g. `github.com/dpb587/openvpn-bosh-release`). Unfortunately, because it requires releases to be listed on bosh.io, it does not support private releases nor unlisted, public releases. As a workaround, pipelines could consume repositories via a regular [`git` resource](https://github.com/concourse/git-resource), but each pipeline ends up needing to implement their own scripts for building tarballs or managing version constraints.

On the release-publishing side of things, no resource types natively supported creating new versions of a BOSH release. Instead, teams implement their own scripts to run through the `bosh finalize-release` or `bosh create-release` workflows, making sure to handle private credentials, versioning, committing, and tagging results. With different teams (and even releases within teams) using different variations of the process, different releases tend to have slightly different versioning practices. The lack of standardization makes it difficult to know what to expect of BOSH release repositories.

Another workflow that I often saw was pipelines creating dev releases from repositories. The scripts to perform this step were fairly short and easily copied between repositories, but they usually require some extra steps for ensuring version uniqueness across the pipeline. Typically developers would rely on an external [`semver` resource](https://github.com/concourse/semver-resource) or, less frequently, use the `--timestamp-version` option of `create-release`. Both of these worked, but it makes it difficult to easily see which semver or timestamp corresponds to a particular commit in the repository.


## Desired Functionality

Having dealt with those issues across quite a few repositories, I thought there was room to consolidate a lot of those behaviors. Specifically, I wanted a resource that supported:

 * public *and* private repositories (and did not need to rely on external servers like bosh.io);
 * building final releases *and* dev releases;
 * versioning which provides more context about the commit it is based on;
 * repositories with multiple feature or release branches;
 * using version constraints to restrict versions of a release that would be found; and
 * creating new final releases (and handles uploading, committing, tagging).

Eventually, I made [`dpb587/bosh-release-resource`](https://github.com/dpb587/bosh-release-resource) to meet those goals and have been using it in my public and private releases.


## Advantages

One of the really nice things about this resource is that it helps remove a lot of the duplicated scripts around finalizing releases. For example, when my OpenVPN release [switched to `bosh-release`](https://github.com/dpb587/openvpn-bosh-release/commit/9e20263fe49f02e25f4a4b056814defb72ae8a77) I was able to remove several CI tasks and significantly simplify the pipeline configuration. The `put` operation in the updated pipeline has the nice side-effect of Concourse performing the `get` step right after which, in this case, builds the release tarball that can then automatically be uploaded to S3 or the GitHub release right after.

When it comes to deploying releases, the resource has also been helpful for testing releases and pulling in private releases. For example, I have a couple pipelines which use [`bosh-deployment` resource](https://github.com/cloudfoundry/bosh-deployment-resource) and deploy whichever versions of releases are passed in. With `bosh-release`, I have been able to switch from using generic `git` resources and no longer need preliminary steps to create the custom dev release tarball before continuing to the `bosh-deployment` step.

On the publishing side of things, `bosh-release` takes care of tagging when it pushes a new version of the release. I applied a couple opinions on how to tag a release (since BOSH release repositories are structured a bit weirdly). First, it always tags the "commit hash" of the release (not the commit where the release is finalized). This ensures that tags refer to the underlying code of the release which enables easier code diffs and reviews. Second, instead of using lightweight tags, it uses annotated tags to correctly record the date when a version was released (not just when the last commit was).


## Try It Out

If you maintain a pipeline for a BOSH releases, consider giving the [`bosh-release` resource](https://github.com/dpb587/bosh-release-resource) a try and see if it can help simplify your repository. For information to get started and more details about its behavior, see the [README](https://github.com/dpb587/bosh-release-resource#readme). For bugs or ideas about how the resource can be improved, feel free to create a [GitHub issue](https://github.com/dpb587/bosh-release-resource/issues).
