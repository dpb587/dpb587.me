---
'@context': http://schema.org
'@type': BlogPosting
datePublished: "2019-01-23"
description: Consistently representing both internal and external dependencies.
keywords:
- automation
- blobs
- metalink
- metalink-repository
name: 'Metalink Repositories: Stability Channels'
---

Continuing on the topic of [metalink repositories](/posts/2018/12/28/metalink-repositories-background-and-motivation), one of the biggest advantages I have found is that I no longer need to worry about complicated rules for when products should be published and what downstream components might be affected. Instead, I can focus on defining what "alpha" vs "rc" vs "stable" mean, and then downstream components consume however it's appropriate for them.

That may sound trivial and obvious. However, I have found implementing those workflows in a product to be difficult before using metalink repositories (see [Motivations](/posts/2018/12/28/metalink-repositories-background-and-motivation.md#motivations) for more details). Previously I might publish assets to S3, or use branches for source code, or prefixed tags in a git repository; but then you lose checksum options, or the ability to share pre-built assets, or noisy repositories. A [metalink file](/posts/2017/10/09/documenting-blobs-with-metalink-files) provides a method for separating the concerns of code from distributing assets.

Before getting into a few other benefits, let's start with an example.


## Beta vs RC vs Stable

I mentioned that metalink repositories are good for separating out stability channels. For example, my [openvpn-bosh-release](https://github.com/dpb587/openvpn-bosh-release) has an [`artifacts` branch](https://github.com/dpb587/openvpn-bosh-release/tree/artifacts) where I publish the metalinks for my builds. I can point consumers to the branch and they can use the [`README.md`](https://github.com/dpb587/openvpn-bosh-release/blob/artifacts/README.md) to understand which channel is best for them.

Production environments would typically consume the [`stable` directory](https://github.com/dpb587/openvpn-bosh-release/tree/artifacts/release/stable). For environments I manage though I point them to earlier, less stable channels for faster feedback and "dog fooding" changes. For example:

 * personal development environment is pointed to `beta` to run whatever version has passed CI tests;
 * client production environment is pointed to `rc` to run whatever version is about to be promoted for public release; and
 * critical infrastructure is pointed to `stable` to run only very confident versions.

Within [Concourse](https://concourse-ci.org/), environments are then able to consume releases with something like the following...

```yaml
jobs:
- name: "deploy-vpn"
  plan:
  - get: "openvpn-bosh-release"
    trigger: true
  - put: "vpn-deployment"
    params:
      releases:
      - "openvpn-bosh-release/*.tgz"
resources:
- name: "openvpn-bosh-release"
  type: "metalink-repository"
  source:
    uri: "git+https://github.com/dpb587/openvpn-bosh-release.git//release/stable#artifacts"
```

If a channel starts to become too stable or unstable, they only need to change a path (as opposed to switching from a different resource entirely, like `s3` or `bosh-io-release` for BOSH releases).


## Internal Stability Channels

Eventually, larger teams start implementing acceptance workflows as part of adopting upstream dependency changes. For example, an operations team likely has a separate staging and production environment, and they would not immediately roll out a new dependency straight to production. At that point, you need some way to track and coordinate across environments. Historically, I have seen teams handling this manually or by having some very large, complex pipelines which need to be aware of multiple environments.

I do not think environments and their pipelines should be aware of how their component dependencies may be related to other, downstream environments. Instead, they should declare what state their components are in, and then downstream environments can refer to that. By separating the concerns, you reduce the complexity and scope that a pipeline and environment is responsible for.

To make this more tangible, let's expand on the earlier example. After the `vpn-deployment` is successfully deployed, run some acceptance tests against it, and then publish it to an internal, environment and deployment-specific metalink repository.

```yaml
jobs:
- name: "deploy-vpn"
  plan:
  - { get: "openvpn-bosh-release", ... }
  - { put: "vpn-deployment", ... }
  - { task: "vpn-acceptance-tests", ... }
  - put: "openvpn-bosh-release-env"
    params:
      metalink: "openvpn-bosh-release/.resource/metalink.meta4"
resources:
- { name: "openvpn-bosh-release", ... }
- name: "openvpn-bosh-release-env"
  type: "metalink-repository"
  source:
    uri: "git+ssh://github.com/acme/ops-env-staging.git//vpn/openvpn-bosh-release#state"
    options:
      private_key: "((env_state_private_key))" # read+write access
```

Once the `deploy-vpn` job has finished and published to the env-specific repository, others can confidently know it has been accepted. This is, essentially, another example of mirroring metalinks discussed in the [previous post](/posts/2018/12/30/metalink-repositories-mirroring-third-party-dependencies), although this does not take the additional steps of mirroring the underlying data to a different storage service.

If we assume that was the staging environment, the production environment can have the exact same job automation and the only thing which needs to change is to declare that it wants to use the version that staging's `vpn` deployment has accepted.

```yaml
jobs:
- { name: "deploy-vpn", ... }
resources:
- name: "openvpn-bosh-release"
  type: "metalink-repository"
  source:
    uri: "git+ssh://github.com/acme/ops-env-staging.git//vpn/openvpn-bosh-release#state" # updated
    options:
      private_key: "((staging_env_state_private_key))" # read-only access
- name: "openvpn-bosh-release-env"
  type: "metalink-repository"
  source:
    uri: "git+ssh://github.com/acme/ops-env-production.git//vpn/openvpn-bosh-release#state"
    options:
      private_key: "((env_state_private_key))" # read+write access
```


### Stemcell Rollout Example

As a more specific example, my personal development environment automatically deploys new stemcells as they come out. To reduce potential impact of a breaking change, the stemcell starts deploying to low impact deployments and then works its way to potentially higher impact deployments. For example, the rollout looks something like the following:

```
infra-vpn -> bosh -> personal-vpn -> website -> infra-nat -> concourse
```

The Concourse pipeline managing this uses the "Internal Stability Channels" approach from the previous example. Instead of watching a separate environment, it's watching another deployment in its own environment to know when a stemcell is considered acceptable. This [development environment](https://dpb587.github.io/4491f193-db23-4e23-8b98-7fc8b3c826ef/) can then be the basis for other development or staging environments that I operate.
