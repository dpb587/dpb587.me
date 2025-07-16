---
description: Using metalink repositories to track upstream artifacts.
params:
    nav:
        tag:
            automation: true
            blobs: true
            metalink: true
            metalink-repository: true
publishDate: "2018-12-30"
title: "Metalink Repositories: Mirroring Third-Party Dependencies"
---

When managing project dependencies which are outside of your control, it is often best practice to assume those artifacts may disappear (e.g. they may move, disappear, or become corrupt). For this reason, you may want to be mirroring your assets which, with [metalink repositories](@/src/content/entry/2018/metalink-repositories-background-and-motivation-20181228), provides the functionality of:

 * Multiple URLs can be configured for where to find an artifact. This allows for documenting where files were originally discovered, but also supports retrying download from mirrors if one location fails.
 * Download locations can be prioritized and configured for locations which helps ensure you always use a local mirror which may be optimized for your environment.
 * Checksums in the metalink continue to verify data integrity between download locations.

Building on top of metalinks is the idea of having a shared repository to document them. When you separate the processes of mirroring files and consuming them, it allows your product workflows to remain simpler and focused on a single responsibility. For example, if you have a large team with many shared dependencies, perhaps the mirroring process is your own "internal product" with a single configuration/pipeline for mirroring dependencies into metalink repositories. Then, individual products do not need to worry about 1) knowing how to download a dependency (e.g. [Go](https://golang.org/)); 2) dealing with the overhead of mirroring; and 3) can reuse artifacts from a local, faster mirror.


# Example {#example}

For a concrete example of mirroring, here is a Concourse pipeline which mirrors Go to a custom S3 bucket. It is based on the [`dynamic-metalink` resource](https://github.com/dpb587/dynamic-metalink-resource) ([learn more](@/src/content/entry/2018/watching-upstream-binaries-with-concourse-20181202)) and the [`metalink-repository` resource](https://github.com/dpb587/metalink-repository-resource). The inline comments provide more insight about what it is doing.

```yaml
resources:

# This is a simple check which watches the go download endpoint for versions
# and provides the download locations and checksums in a metalink JSON format.
- name: golang
  type: dynamic-metalink
  source:
    version_check: |
      curl -s https://golang.org/dl/?mode=json | jq -r '.[].version[2:]'
    metalink_get: |
      curl -s https://golang.org/dl/?mode=json | jq '
        map(select(.version[2:] == env.version)) | map({
          "files": (.files | map({
            "name": .filename,
            "version": env.version,
            "size": .size,
            "urls": [ { "url": "https://dl.google.com/go/\(.filename)" } ],
            "hashes": [ { "type": "sha-256", "hash": .sha256 } ] } ) ) } )[]'

# This configures a GitHub repository to store our metalink and mirror data. It
# will put each golang version into its own file within the golang.org directory
# and upload them to an S3 bucket, using a checksum as the object key.
- name: golang-mirror
  type: metalink-repository
  source:
    uri: git+ssh://git@github.com:acme/mirrors.git//golang.org
    options:
      private_key: ((git_private_key))
    # Optionally, these next two settings will mirror the artifacts to a custom
    # S3 bucket to ensure continued access.
    mirror_files:
    - destination: s3://s3-external-1.amazonaws.com/acme-mirror-us-east-1/golang.org/{{.SHA256}}
      location: US
      priority: 10
    url_handlers:
    - type: s3
      options:
        access_key: ((s3_access_key))
        secret_key: ((s3_secret_key))

jobs:

# Whenever there is a new version of golang, get the generated metalink which
# refers to the official download URLs and checksums. Then, put that metalink
# into our own repository and, because the golang-mirror configures the
# mirror_files option, it will take care of reuploading them to our bucket and
# adding it to the list of download locations.
- name: mirror-golang
  plan:
  - get: golang
    trigger: true
    params:
      skip_download: true
  - put: golang-mirror
    params:
      metalink: golang/.resource/metalink.meta4
    get_params:
      skip_download: true

# These configure the custom resource types to support this pipeline.
resource_types:
- name: dynamic-metalink
  type: docker-image
  source:
    repository: dpb587/dynamic-metalink-resource
- name: metalink-repository
  type: docker-image
  source:
    repository: dpb587/metalink-repository-resource
```
