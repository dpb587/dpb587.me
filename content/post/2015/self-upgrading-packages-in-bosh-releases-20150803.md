---
description: A strategy for monitoring upstream dependencies for self-sustaining packages.
params:
    nav:
        tag:
            bosh: true
            package manager: true
            updates: true
            upgrades: true
            versions: true
publishDate: "2015-08-03"
title: Self-Upgrading Packages in BOSH Releases
---

Outside of [BOSH][1] world, package management is often handled by tools like [yum][2] and [apt][3]. With those tools, you're able to run trivial commands like `yum info apache2` to check the available versions or `yum update apache2` to upgrade to the latest version. It's even possible to automatically apply updates via cron job. With BOSH, it's not nearly so easy since you must monitor upstream releases, manually downloading the sources before moving on to testing and deploying. Personally, this repetitive sort of maintenance is one of my least favorite tasks; so, to avoid it, I started automating.


# Automating {#automating}

There are two critical steps involved with sort of thing. First is being able to `check` when new versions are available. For this post, I'll use my [OpenVPN BOSH Release][9] which has a single package with three dependencies. For each dependency, I can use commands to check for the latest version...

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    # lzo
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    wget -q -O- http://www.oberhumer.com/opensource/lzo/download/ | grep -E 'href="lzo-[^"]+.tar.gz"' | sed -E 's/^.+href="lzo-([^"]+).tar.gz".+$/\1/' | gsort -rV | head -n1
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    2.09
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    # openssl
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    git ls-remote --tags https://github.com/openssl/openssl.git | cut -f2 | grep -Ev '\^{}' | grep -E '^refs/tags/OpenSSL_.+$' | sed -E 's/^refs\/tags\/OpenSSL_(.+)$/\1/' | tr '_' '.' | grep -E '^\d+\.\d+\.\d+\w*$' | gsort -rV | head -n1
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    1.0.2d
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    # openvpn
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    git ls-remote --tags https://github.com/OpenVPN/openvpn.git | cut -f2 | grep -Ev '\^{}' | grep -E '^refs/tags/v.+$' | sed -E 's/^refs\/tags\/v(.+)$/\1/' | tr '_' '.' | grep -E '^\d+\.\d+\.\d+$' | gsort -rV | head -n1
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    2.3.7
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

The location to download the source for a dependency is typically predictable, once the pattern is known...

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    wget -O lzo.tar.gz "http://www.oberhumer.com/opensource/lzo/download/lzo-${VERSION}.tar.gz"
    ```

  {{< /terminal-input >}}

{{< /terminal >}}

Within the release, files become structured like:

```
./blobs/openvpn-blobs/
  ./lzo/
    lzo.tar.gz
  ./openssl/
    openssl.tar.gz
  ./openvpn/
    openvpn.tar.gz
./packages/openvpn/
  ./deps/
    ./lzo/
      ./check
      ./get
      ./VERSION
    ./openssl/
      ./check
      ./get
      ./VERSION
    ./openvpn/
      ./check
      ./get
      ./VERSION
  ./packaging
  ./spec
```

Each dependency has its own blob directory, allowing old versions to be fully removed before replacing it with the new version's file(s). Inside the package directory, `VERSION` is a committed state file used for comparison in version checks. It can also be used to quickly reference and document what versions are being used...

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    find packages -name VERSION | xargs -I {} -- /bin/bash -c 'A={} ; printf "%12s %s/%s\n" $( cat $A ) $( basename $( dirname $( dirname $( dirname $A ) ) ) ) $( basename $( dirname $A ))'
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
            2.09 openvpn/lzo
          1.0.2d openvpn/openssl
            2.3.7 openvpn/openvpn
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

One side effect of this structure is that the `packaging` script and `spec` manifest should be version agnostic. Otherwise you still end up needing to tweak them every time a version changes, defeating the automation. In `packaging`, references such as `openssl-1.0.2d` would typically become `openssl-*`. In `spec`, the `files` property is minimal...

```yaml
---
title: "openvpn"
files:
  - "openvpn-blobs/**/*"
```

When it comes time to upgrade dependencies I can run a [utility script]({{< appendix-ref "2015-08-03-self-upgrading-packages-in-bosh-releases/bin/deps-upgrade-auto.sh" >}})...

{{< terminal >}}

  {{< terminal-input >}}

    ```
    ./bin/deps-upgrade-auto
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    ==> openvpn/lzo
    --| local 2.09
    --| check 2.09
    ==> openvpn/openssl
    --| local 1.0.1m
    --| check 1.0.2d
    --> fetching new version
    --> 5.1M
    ==> openvpn/openvpn
    --| local 2.3.6
    --| check 2.3.7
    --> fetching new version
    --> 1.1M
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

The script runs through all the dependencies, uploads new blobs to the blobstore, and commits the changes with a nice summary...

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    git log --format=%B -n1
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    Upgraded 2 package dependencies

    openvpn

      * openssl now 1.0.2d (was 1.0.1m)
      * openvpn now 2.3.7 (was 2.3.6)
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

At this point, I have a single command that I can run to check and upgrade dependencies in all my packages. This openvpn example is fairly trivial, but some packages are much more complicated with many more dependencies from separate sites and using separate versioning and download strategies.


# Continuous Integration {#continuous-integration}

Of course, upgrades aren't always without issue, which is why it's important to integrate it with existing tests and Continuous Integration pipelines. Consider the following workflow:

 * weekly, CI runs `deps-upgrade-auto` off the `master` branch, pushing new versions to `master-autoupgrade`
 * CI monitors `master-autoupgrade` for new commits, and follows the typical development pipeline
    * it creates a new development release version (i.e. `bosh create release`)
    * it creates a new test deployment with the version and test data
    * it runs unit tests and errand tests against the deployment
 * based on what happens to this version-testing branch...
    * *on-success*: send a Pull Request for a human to review and merge (or, assuming you have quality tests, go ahead and merge it automatically)
    * *on-failure*: create an issue in the repo listing the dependency versions which changed and information about the failed step so that a human can intervene with a headstart on where they need to start investigating

This sort of pipeline results in...

 * best case scenario - a bot sends me a PR with upgraded dependencies which have been tested and confirmed to work in my release and I can click "Merge"
 * worst case scenario - a bot tells me I should upgrade OpenSSL but I need to investigate an issue where OpenVPN client connects are now failing a TLS handshake


# Conclusion {#conclusion}

These `check`/`get`-type scripts and the self-upgrading approach is something I've been using in my releases lately. The value for me comes from the inherent documentation it provides, but mainly it's from being able to offload some of the maintenance burdens I normally need to be concerned about. Although I have yet to fully implement the steps from the [CI section](#continuous-integration) into my [Concourse][8] pipelines, I hope to get there at some point soon.

If you're interested in experimenting with the scripts from this post, you can find them in [this gist]({{< appendix-ref "2015-08-03-self-upgrading-packages-in-bosh-releases/" >}}) along with a few other `check` scripts I've been using. You can also take a look at the commits in the OpenVPN BOSH Release where I [switched][10] to using `deps` and then subsequently [auto-upgraded][11] the dependencies.


 [1]: https://bosh.io/
 [2]: https://en.wikipedia.org/wiki/Yellowdog_Updater,_Modified
 [3]: https://wiki.debian.org/Apt
 [4]: https://openvpn.net/
 [6]: http://php.net/
 [8]: http://concourse.ci/
 [9]: https://github.com/dpb587/openvpn-boshrelease
 [10]: https://github.com/dpb587/openvpn-boshrelease/commit/26f115dfd5d80444fee543e17edf198e7d15b485
 [11]: https://github.com/dpb587/openvpn-boshrelease/commit/ac833f99cb361b0cb7fb39d70b70a0403ba87af8
