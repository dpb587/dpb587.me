---
date: 2016-10-21
title: "Self-Upgrading Packages in BOSH Releases, Part 2"
description: "A strategy for upgrading and testing dependencies for self-sustaining packages."
primary_image: "/blog/2016-10-21-self-upgrading-packages-in-bosh-releases-part-2/pull-request.png"
tags:
- bosh
- package manager
- updates
- upgrades
- versions
aliases:
- /blog/2016/10/21/self-upgrading-packages-in-bosh-releases-part-2.html
---

Last year I wrote [a post][1] about how the process of updating BOSH release blobs could be better automated. The post relied on some [scripts][12] which could be executed to check and download new versions of blobs. The scripts were useful, but they still required manual execution and then testing to verify compatibility. My latest evolution of the idea further automates this with [Concourse][2] to check for new versions, download new blobs, build test releases, and then send pull requests for successful upgrades.

<!--more-->

![Screenshot: blobs-pipeline](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2016-10-21-self-upgrading-packages-in-bosh-releases-part-2/blobs-pipeline.png)


## Existing Scripts as Concourse Resources

In the previous post, I relied on one script to monitor versions and another script to download a specific version. This is very similar to Concourse resources which utilize a `check` and `get` script, but it would be a lot of work to manually maintain a full Concourse-friendly resource for every blob package. Instead, I created a [task][3] which wraps the simple version/download scripts into a [Docker][4] image that Concourse can execute as a regular resource. For each self-upgrading blob, I have a job which looks [like][5]...

```json
{ "plan": [
  { "get": "repo",
    "trigger": true },
  { "task": "prepare-buildroot",
    "file": "repo/ci/images/release-blob/prepare/task.yml",
    "params": {
      "blob": "{{blob}}" } },
  { "put": "ci-release-blob-{{blob}}-image",
    "params": {
      "build": "buildroot" } } ] }
```

Once the [naïve][13] [scripts][14] are wrapped by [resource-friendly][7] [scripts][15], I can then use the Docker images as a [custom resource type][6] in my pipelines. By `get`ing them as a `trigger`, tasks will have access to whatever files were downloaded. For example, I can have a job which adds the new package blobs to my release's blobstore before continuing to push somewhere...

```json
{ "plan": [
  { "get": "blob",
    "resource": "release-blob-{{blob}}",
    "trigger": true },
  { "get": "repo" },
  { "task": "bump-release-blob",
    "file": "repo/ci/tasks/bump-release-blob/task.yml",
    "params": {
      "blob": "{{blob}}" } },
  ... ] }
```


## Running Tests

Whenever something changes in a release, you typically want to run all of the tests. Historically I maintained this logic in my main pipeline which meant I either needed to immediately push the new (potentially broken) blobs to my main branch, or I needed to duplicate my test plans. For a third option, I utilized [`jq`][8] to extract my integration tests into a [reusable function][9]...

```
def run_integration_tests:
  [ { "aggregate": [
        { "get": "bosh-lite-stemcell" },
        { "put": "bosh-lite" } ] },
    { "put": "bosh-lite-integration-deployment",
      "params": {
        "target_file": "bosh-lite/target",
        "manifest": "repo/ci/tasks/integration-test/deployment.yml",
        "stemcells": [
          "bosh-lite-stemcell/*.tgz" ],
        "releases": [
          "release/*.tgz" ] } },
    { "task": "integration-test",
      "file": "repo/ci/tasks/integration-test/task.yml" } ] ;
```

With my tests refactored into a separate function I can add this `run_integration_tests` function after I execute `bump-release-blob` from above...

```json
{ "plan": [
  ...,
  { "task": "bump-release-blob",
    "file": "repo/ci/tasks/bump-release-blob/task.yml",
    "params": {
      "blob": "{{blob}}" } },
  create_dev_release,
  run_integration_tests[] ] }
```

This allows me to immediately test my release changes without introducing them on a public branch. I'm also able to continue using the exact same integration tests in my main pipeline.


## Making it Easy to Merge

Once tests are successful, it is more reasonable for someone to spend time reviewing and merging the new version. To that end, I configured the final blob-upgrade-tester job to push the changes to a blob-specific branch and [send a pull request][10]. The pull request includes a reminder about what exactly needs to happen.

![Screenshot: pull-request](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2016-10-21-self-upgrading-packages-in-bosh-releases-part-2/pull-request.png)

By making this [pull request][16], the upgrade becomes something I can easily finish, merge, and cleanup from the [GitHub][11] web UI, even without a terminal or laptop.


## Automate

By automating package upgrades, it is easier to stay up to date with security patches which affect my releases. Refactoring my test tasks out into reusable functions helps provide better confidence in the upgrades before they hit any main branches. Utilizing pull requests for applying the changes reduces friction and provides a reminder in my inbox that something needs to happen. Although this process requires a bit of development overhead, the ability to rely on bots for upgrading, testing, and reminding me about dependency changes lets me focus on more creative tasks in the releases I have been testing this approach in.


 [1]: /blog/2015/08/03/self-upgrading-packages-in-bosh-releases.html
 [2]: https://concourse.ci/
 [3]: https://github.com/dpb587/openvpn-bosh-release/tree/f6a46f923c364ca4bfbdd3da9de00d7fc5c155b6/ci/images/release-blob/prepare
 [4]: https://www.docker.com/
 [5]: https://github.com/dpb587/openvpn-bosh-release/blob/f6a46f923c364ca4bfbdd3da9de00d7fc5c155b6/ci/pipelines/release-blobs/pipeline.jq#L19-L41
 [6]: http://concourse.ci/implementing-resources.html
 [7]: https://github.com/dpb587/openvpn-bosh-release/blob/f6a46f923c364ca4bfbdd3da9de00d7fc5c155b6/ci/images/release-blob/assets/check
 [8]: https://stedolan.github.io/jq/
 [9]: https://github.com/dpb587/openvpn-bosh-release/blob/f6a46f923c364ca4bfbdd3da9de00d7fc5c155b6/ci/pipelines/shared.jq#L25-L51
 [10]: https://github.com/dpb587/openvpn-bosh-release/blob/f6a46f923c364ca4bfbdd3da9de00d7fc5c155b6/ci/tasks/send-release-blob-pr/run.sh#L29-L53
 [11]: https://github.com/
 [12]: {{< appendix-ref "2015-08-03-self-upgrading-packages-in-bosh-releases/" >}}
 [13]: https://github.com/dpb587/openvpn-bosh-release/blob/f6a46f923c364ca4bfbdd3da9de00d7fc5c155b6/src/blobs/openssl/check
 [14]: https://github.com/dpb587/openvpn-bosh-release/blob/f6a46f923c364ca4bfbdd3da9de00d7fc5c155b6/src/blobs/openssl/get
 [15]: https://github.com/dpb587/openvpn-bosh-release/blob/f6a46f923c364ca4bfbdd3da9de00d7fc5c155b6/ci/images/release-blob/assets/in
 [16]: https://github.com/dpb587/openvpn-bosh-release/pull/6
