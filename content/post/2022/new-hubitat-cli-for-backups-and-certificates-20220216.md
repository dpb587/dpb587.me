---
description: Converting shell scripts and curl to make local management easier.
publishDate: "2022-02-16"
title: New Hubitat CLI for Backups and Certificates
---

One of the devices I use in my home is a [Hubitat Elevation hub](https://hubitat.com/). Similar to [SmartThings](https://www.smartthings.com/) or even [Google Home](https://developers.google.com/home), it integrates a wide range of devices, protocols, and apps to help make a "smart" home. I originally chose the Hubitat because of its good support for local and private automation (rather than expecting the internet to always be working and fast).

Keeping with the theme of local operations, I automate some of its management tasks. Historically I downloaded backups and rotated certificates with shell scripts, but Hubitat does not support a formal API and the scripts have always felt a little fragile. This week I decided to convert them into a standalone [`hubitat-cli` binary](https://github.com/dpb587/hubitat-cli) (written in [Go](https://golang.org/)) which is much easier to [test](https://github.com/dpb587/hubitat-cli/blob/57571c4ffe460cd3be6c92a3537452863e67a5a1/hub/login_test.go), [maintain](https://github.com/dpb587/hubitat-cli/tree/57571c4ffe460cd3be6c92a3537452863e67a5a1/.github/workflows), and [use](https://github.com/dpb587/hubitat-cli#usage). Now, instead of [~20 lines](https://github.com/dpb587/dpb587.me/tree/master/appendix/2022-02-16-new-hubitat-cli-for-backups-and-certificates/download-backup-curl.sh) for my backup task, I can just run:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    hubitat-cli backup download
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    2022/02/16 08:47:16 hubitat-cli: "level"=0 "msg"="downloaded backup file to Hubitat_2022-02-16~2.3.0.124.lzf"
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

And, instead of [~40 lines](https://github.com/dpb587/dpb587.me/tree/master/appendix/2022-02-16-new-hubitat-cli-for-backups-and-certificates/apply-certs-curl.sh) to update certificates and reboot the hub, I can run:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    hubitat-cli advanced certificate update --certificate-path=tls.crt --private-key-path=tls.key
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    2022/02/16 08:48:37 hubitat-cli: "level"=0 "msg"="updated certificate"
    2022/02/16 08:48:37 hubitat-cli: "level"=0 "msg"="requested reboot"
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

I use [Kubernetes](https://kubernetes.io/) to coordinate these tasks, so it's also available as a small (~8MB) image:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    docker run ghcr.io/dpb587/hubitat-cli
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    For interacting with Hubitat

    Usage:
      hubitat-cli [command]

    Available Commands:
      advanced    For advanced features
      backup      For managing backups
      reboot      For restarting the hub

    Flags:
      -h, --help                  help for hubitat-cli
          --hub-ca-path string    custom certificate authorities to trust ($HUBITAT_CA_PATH)
          --hub-insecure          disable TLS verifications ($HUBITAT_INSECURE)
          --hub-password string   password for login ($HUBITAT_PASSWORD)
          --hub-url string        URL for accessing the hub (e.g. http://192.0.2.100; $HUBITAT_URL)
          --hub-username string   username for login ($HUBITAT_USERNAME)
      -v, --verbose int           verbosity level

    Use "hubitat-cli [command] --help" for more information about a command.
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

You can learn more from [github.com/dpb587/hubitat-cli](https://github.com/dpb587/hubitat-cli), and test it out yourself by downloading the binary for Linux, macOS, or Windows from the [releases page](https://github.com/dpb587/hubitat-cli/releases).
