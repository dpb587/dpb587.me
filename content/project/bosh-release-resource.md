---
title: "Concourse Resource Type: BOSH Release"
description: A Concourse resource which helps users access and publish BOSH releases using best practices.
github: dpb587/bosh-release-resource
tags:
- bosh
- concourse-resource
- golang
---

I created this Concourse resource type because I was managing many BOSH releases and wanted to codify the best practices for myself and others. I'm particularly proud of:

 * coming up with a [semver](https://semver.org/)-compatible versioning scheme that works with both BOSH, Concourse, and `git`.
 * the resource being adopted by Concourse itself for managing [their own BOSH release](https://github.com/concourse/ci/blob/b9d39fe9ab616d37bdd98c9dc87088f36df2bee1/pipelines/concourse.yml#L9-L11) in their pipelines.
