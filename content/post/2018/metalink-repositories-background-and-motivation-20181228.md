---
description: Some "whys" of this alternative way to track artifacts.
params:
    nav:
        tag:
            automation: true
            blobs: true
            metalink: true
            metalink-repository: true
publishDate: "2018-12-28"
title: "Metalink Repositories: Background and Motivation"
---

A while ago I started standardizing on using metalinks to record information about blobs. In my [original post](@/src/content/entry/2017/documenting-blobs-with-metalink-files-20171009) about metalinks, I briefly touched on the idea of [querying directories of metalinks](@/src/content/entry/2017/documenting-blobs-with-metalink-files-20171009#repository-querying), treating it as a lightweight database. This post starts a short series discussing what shortcomings motivated my interest in something like "metalink repositories" and how I have been building on the concept in several ways.


# Background {#background}

First, a brief introduction to how I have defined and used "metalink repositories". At its most basic level, a "repository" means something which, given a URI and optional parameters, can retrieve a list of metalinks. I have implemented several basic repository backends, but the primary one I use is a [`git`](https://git-scm.com)-based one. The Git approach builds on the standard advantages of a version control system, such as: private + public repositories, commit-based change history, and distributable information. For example, using the [`meta4-repo` helper](https://github.com/dpb587/metalink/releases) to list dev artifacts I have published in a public repository:

{{< terminal >}}

    {{< terminal-input >}}

        ```bash
        meta4-repo filter --format=version -n5 git+https://github.com/dpb587/ssoca.git//ssoca-dev#artifacts
        ```

    {{< /terminal-input >}}

    {{< terminal-output >}}

        ```
        0.13.0-dev.2
        0.13.0-dev.1
        0.12.0-dev.3
        0.12.0-dev.2
        0.12.0-dev.1
        ```

    {{< /terminal-output >}}

{{< /terminal >}}

Each version has its own file in the `ssoca-dev` directory of the `artifacts` branch. View the files in [dpb587/ssoca](https://github.com/dpb587/ssoca/tree/artifacts/ssoca-dev) if you want to take a closer look.


# Motivation {#motivation}

When managing a product that produces artifacts of some sort (e.g. binaries, tarballs), they often go through various stages of stability. For example, a common workflow I follow is producing an *alpha* after every change, *beta* after initial tests pass, *release candidate* after broader testing, and *stable* when officially tagging a release. In my experience, any implementation of this should prioritize automation (within the product and external) and ability to audit artifacts for security and provenance.

I have seen and experimented with several different strategies of tracking artifacts for products over time. However, drawbacks in each of them have caused me to look for something better -- especially when considered within the context of automation.


## S3 Bucket or Prefix-based {#s3-bucket-or-prefix-based}

Whenever an artifact passes a new stage it gets uploaded to a new bucket or directory. For example, *alphas* go into `s3://my-product-artifacts/alphas/product-v(semver).tgz` and then into `s3://my-product-artifacts/stable/product-v{semver}.tgz` once promoted. The advantage of this is that consumers are able to list those buckets to find versions they are interested in for specific stabilities. The disadvantages are that, on its own, this does not implicitly provide data guarantees or verification to ensure those objects are not tampered with and do not change over time; it also typically requires you to be storing multiple copies of the same artifact.


## S3 Object Versioning {#s3-object-versioning}

Whenever an artifact passes a new stage it gets uploaded to a specific stability-specific path. For example, *alphas* go into `s3://my-product-artifacts/product-alpha.tgz` and then into `s3://my-product-artifacts/product-stable.tgz`. Similar to the previous approach, consumers are able to list the versions of those objects to get older copies. The disadvantages are that this approach relies on a [primarily] Amazon S3-specific versioning feature; similar to the previous approach, it also does not track checksums for these artifacts (although versioning does help guarantee old objects can remain accessible); this approach also loses precision because it does not easily allow consumers to know semver-specific metadata and use it as constraints without downloading artifacts.


## GitHub Releases {#github-releases}

Whenever an artifact passes part or all of a CI pipeline, a new release is created on GitHub. Depending on what tests were passed, it might be marked as a pre-release or an official release. This does help external consumers know state for "pre-releases" instead of only stable releases, but it does rely on GitHub-specific conventions and GitHub does not implicitly provide checksum information about those artifacts. Depending on how frequently the product is creating pre-releases, they can add a lot of noise for regular users who are often just interested in latest, official releases. Additionally, some product development workflows may require more classifications than just pre-release and official releases (e.g. alpha or beta releases).


## Concourse `passed` constraints {#concourse-passed-constraints}

A popular approach in product pipelines, the [`passed`](https://concourse-ci.org/get-step.html#get-step-passed) constraint allows [Concourse](https://concourse-ci.org) to restrict how "stable" an artifact is depending on how far it has successfully been tested through a pipeline of tests. The advantage of this is that it is very popular and is simple to configure within a pipeline. The disadvantages start to appear when you have external consumers interested in the artifacts at earlier-than-published stabilities. Workarounds typically involve introducing new resources for each stability, typically using one of the earlier approaches (and their own disadvantages).


# Metalink [Repositories] {#metalink-repositories}

Most of the mentioned disadvantages can be handled by metalinks. In general, it comes down to metalinks separating two primary concerns: information about artifact bits; and where those artifacts are located.

 * Data Guarantees -- the metalink stores additional integrity information, such as checksums or signatures, which acts as an independent record of what an artifact should look like.
 * Single Object Storage -- the metalink file helps separate classification of artifacts from the actual storage. The lighter-weight metalink file can be duplicated for different contexts while still referring to the exact same objects in an artifact server. For example, for some artifacts I will store files in S3 by SHA1 content digest instead of a specific semver-based file name because an object name might imply a specific stability or version.
 * External Consumers -- by storing a metalink outside of a pipeline or workflow process, other consumers can more easily use it. For example, in development environments I can easily watch for release candidates instead of waiting for official, published versions.
 * Git repository-based -- by storing metalinks in a Git repository, any product or component can easily maintain their own index of artifacts in a decentralized manner without needing to maintain an independent artifact server or central registration.
 * Concourse Resource -- the [`s3` resource](https://github.com/concourse/s3-resource) is a popular choice due to availability and easiness of uploading/downloading artifacts. After missing that functionality, I implemented a [`metalink-repository` resource](https://github.com/dpb587/metalink-repository-resource) which helps provide equivalent functionality for artifact management, but with a metalink repository for storing those artifact references.


# To Be Continued {#to-be-continued}

Having summarized some of the alternatives approaches, concerns, and initial motivation for using metalink repository, several upcoming posts will go into more detail about how I have been adopting the metalink repositories in various use cases.
