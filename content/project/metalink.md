---
title: Metalink
description: CLI tools and library for using the XML-based specification for assets.
github: dpb587/metalink
tags:
- metalink
---

I created this after researching ways to record checksums and mirrors of files and finding an RFC about it. I used this in several private projects, and it also ended up being adopted by the [Cloud Foundry BOSH website](https://bosh.io/) for its `git`-backed database of [releases](https://github.com/bosh-io/releases-index) and [stemcells](https://github.com/bosh-io/stemcells-core-index). I also ended up creating a Concourse resource type for it.
