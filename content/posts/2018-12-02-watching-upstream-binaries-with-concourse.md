---
'@context': http://schema.org
'@type': BlogPosting
datePublished: "2018-12-02"
description: Unifying how pipelines monitor third-party assets and versions.
keywords:
- bosh
- ci
- concourse
- concourse-resource
- dynamic-metalink
- metalink
- updates
- upgrades
- versions
name: Watching Upstream Binaries with Concourse
---

When building software packages, it's easy to accumulate dependencies on dozens of other, upstream software components. When building the first version of something, it's easy to blindly download the source of the latest version off the packages' website. However, once you're past prototypes and need to deal with auditing or maintenance, it becomes important to have some [automated] processes in place.

<!--more-->

I have written [several](https://dpb587.me/blog/2016/10/21/self-upgrading-packages-in-bosh-releases-part-2.html) [posts](https://dpb587.me/blog/2015/08/03/self-upgrading-packages-in-bosh-releases.html) over the years around experiments for automatically upgrading components to avoid repetitive work. Over time, and across projects, I noticed that a lot of the version and asset management responsibilities could be handled by more native Concourse concepts.


## Difficulties Worth Solving

In my [most recent iteration](https://dpb587.me/blog/2016/10/21/self-upgrading-packages-in-bosh-releases-part-2.html), `check` and `get` scripts are committed to the repository which encode the logic for finding the latest version and downloading the blobs. Then, a custom resource type was built for the repository so that each blob could be used as a native resource in the pipeline.

This approach had a few limitations:

 * it requires repository-specific Docker images to be maintained;
 * the Docker image may have different `get`/`check` logic from what is committed to the repository at a given time;
 * version constraints are difficult since they need to be encoded into the `check` script (e.g. `grep -v 0.*`); and
 * each `get` script was responsible for correctly downloading and performing its own checksum or signature verifications.

Since then, I have standardized how I track blobs by using [metalinks](https://dpb587.me/blog/2017/10/09/documenting-blobs-with-metalink-files.html) because they support tracking download origins, checksum verification, signature verification, and lightweight version annotations. By reusing a metalink-like interface, I realized I could simplify many of these issues.


## Planning an Interface

Concourse provides the `check`/`get` interface already, so I only needed to figure out how the resource should be configured. After reviewing the `check` behaviors I've used across projects, I settled on a few assumptions on how the resource should be configured:

 * simple command line tools would typically be sufficient for querying upstream versions (e.g. `curl`, `git`, `grep`, `jq`, `sed`);
 * instead of encoding version constraints in BASH commands:
    * return all known, upstream versions;
    * assume versions are semantically versioned; and
    * provide a resource configuration option for the user to specify a semver constraint.

The `get` behaviors were a bit more complicated, but eventually I settled on the user managing a script which receives a `version` environment variable and then the output is a metalink (in JSON format) of the blob details.

The basic functionality led to three main resource configuration options:

 * `version_check` - a bash script for returning a list of known versions
 * `metalink_get` a bash script for generating a metalink with the given `version`
 * `version` - an optional, semantic version constraint in the case that latest is not good enough

It seems a bit counter-intuitive that a Concourse resource would be configured with such dynamic commands (typically resources are much more deterministic in their configuration). But this seemed like a more practical way to start (when compared to managing a large number of Docker images, one per upstream, third-party software and a large number of resource types).

Eventually I settled on calling it `dynamic-metalink` and created the repository at [dpb587/dynamic-metalink-resource](https://github.com/dpb587/dynamic-metalink-resource).


## Testing the Configuration

Using Go as an example, golang.org provides a simple API with the most recent patch of their most recent minor versions. The following `curl` is sufficient for getting the version list:

```bash
$ curl -s https://golang.org/dl/?mode=json | jq -r '.[].version[2:]'
1.11.2
1.10.5
```

Once the version is known, it can build the metalink-friendly structure of download and hash verification data:

```bash
$ version="1.11.2" curl -s https://golang.org/dl/?mode=json | jq '
  map(select(.version[2:] == env.version)) | map({
    "files": (.files | map({
      "name": .filename,
      "size": .size,
      "urls": [ { "url": "https://dl.google.com/go/\(.filename)" } ],
      "hashes": [ { "type": "sha-256", "hash": .sha256 } ] } ) ) } )[]'
{ "files": [
  { "name": "go1.11.2.src.tar.gz",
    "size": 21100145,
    "urls": [
      { "url": "https://dl.google.com/go/go1.11.2.src.tar.gz" } ],
    "hashes": [
      { "type": "sha-256",
        "hash": "042fba357210816160341f1002440550e952eb12678f7c9e7e9d389437942550" } ] },
  { "name": "go1.11.2.darwin-amd64.tar.gz",
...snip...
```


## Testing the Pipeline

For a more specific example, I use the following for this website and automatically updating hugo. It watches for new versions, imports the new binary into the release, builds the site and deploys it, and then finally pushes it to the repository. When it succeeds or fails, it sends a notification to a Slack channel just to let me know.

```yaml
jobs:
- name: upgrade-hugo-blob
  plan:
  - aggregate:
    - get: blob
      trigger: true
      resource: hugo-blob
    - get: repo
    - get: bosh-release-blobs-upgrader-pipeline
  - task: sync-blobs
    file: bosh-release-blobs-upgrader-pipeline/tasks/sync-blobs.yml
    params:
      blob: hugo
      track_files: .resource/metalink.meta4
  - task: test-deploy
    privileged: true
    file: repo/ci/tasks/test-deploy/config.yml
  - task: upload-blob
    file: bosh-release-blobs-upgrader-pipeline/tasks/upload-blobs.yml
    params:
      release_private_yml: |
        blobstore:
          options:
            access_key_id: ((access_key))
            secret_access_key: ((secret_key))
  - put: repo
    params:
      rebase: true
      repository: repo
  on_failure: *slack-notify-blob-failure
  on_success: *slack-notify-blob-success
resources:
- name: hugo-blob
  type: dynamic-metalink
  source:
    metalink_get: |
      jq -n '
        env.version | {
          "files": [
            { "name": "hugo_\(.)_Linux-64bit.tar.gz",
              "urls": [ { "url": "https://github.com/gohugoio/hugo/releases/download/v\(.)/hugo_\(.)_Linux-64bit.tar.gz" } ] } ] }'
    version_check: |
      git ls-remote --tags https://github.com/gohugoio/hugo.git \
        | cut -f2 \
        | grep -v '\^{}' \
        | grep -E '^refs/tags/v.+$' \
        | sed -E 's/^refs\/tags\/v(.+)$/\1/' \
        | grep -v '-' \
        | grep -E '^\d+\.\d+(\.\d+)?$'
- name: repo
  type: git
  source: ...snip...
- name: bosh-release-blobs-upgrader-pipeline
  type: git
  source: ...snip...
```

For example, Concourse automatically upgraded to [v0.52](https://github.com/dpb587/dpb587.me/commit/db6a898c1bcb3ebbeff33d1cd161b115f42e7658) without me needing to worry about anything (you'll also see I'm tracking the original download source information in those commits as well). To see more examples in pipeline-form, see the [`examples` directory](https://github.com/dpb587/dynamic-metalink-resource/tree/master/examples) in the repository.


## Usefulness

By having this `dynamic-metalink` resource available it has been even easier for me to implement automatic updating of upstream dependencies for my projects, especially BOSH releases. This sort of version management is a little different than package manager-based tools since it is intended to handle downloads that someone would otherwise handle manually. For example, if you are looking for something to keep your Ruby `Gemfile` or PHP `composer.json` up to date, tools like [dependabot](https://dependabot.com/) are much more useful since they can take advantage of the existing package and versioning systems. If there was a large, consolidated repository of package/binary/version assets for arbitrary projects, it would be a bit easier to have something like `dependabot`. But, in the meantime, the extra sections of pipeline configuration are helping me keep dependencies up to date.

If you are looking for more examples of this resource being used or to learn more, visit the repository at [dpb587/dynamic-metalink-resource](https://github.com/dpb587/dynamic-metalink-resource).
