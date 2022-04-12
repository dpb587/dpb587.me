---
date: 2022-04-12
title: Synology NAS and Custom Certificate Authorities
description: Automating certificate management for private directories with acme.sh.
---

One of the easiest ways to get a trusted certificate for a [Synology NAS](https://www.synology.com/) is through its integrated [Let's Encrypt](https://letsencrypt.org/) support. While convenient, it requires the NAS to be accessible from the internet *and* the hostname ends up being part of public records through [certificate transparency](https://certificate.transparency.dev/). In my case, I have a NAS on an internal network with its own private [certificate authority](https://en.wikipedia.org/wiki/Certificate_authority) which supports [ACME](https://en.wikipedia.org/wiki/Automatic_Certificate_Management_Environment) (the same certificate provisioning protocol that Let's Encrypt uses). Since the NAS does not support custom ACME directories, I've been using the following steps to automate certificate management with the [`acme.sh` tool](https://github.com/acmesh-official/acme.sh).

## Service Account

To start, create a dedicated user which will take care of managing the certificates. I called it `certupdater` and configure it with:

* a password that can be used later
* administrator group membership (for permissions to update certificates and restart services)
* no default access to shared folders (aside from the [`homes` folder](https://kb.synology.com/en-af/DSM/help/DSM/AdminCenter/file_user_advanced?version=7) for its automation)
* no default access to applications (aside from DSM itself to manage the certificates)

{{< details summary="Detailed Steps (DSM 7)" >}}

1. Log in to Synology DSM as an administrator.
1. Open the **Control Panel** application.
1. Go to the **User & Group** component.
1. Click the **Create** button.
1. In the **User Creation Wizard** window...

   1. On the **Enter user information** page...

      1. For **Name**, enter **certupdater**.
      1. For **Password**, enter and confirm a secure value (or use **Generate Random Password**).
      1. Activate **Disallow the user to change account password**.
      1. Click the **Next** button.

   1. On the **Join groups** page...

      1. Activate the **administrators** group.
      1. Click the **Next** button.

   1. On the **Assign shared folders permissions** page, click the **Next** button.
   1. On the **Assign application permissions** page...
      1. For **User Permissions**, activate **Deny** for all but **DSM**.
      1. Click the **Next** button.
   1. On the **Set user speed limit** page, click the **Next** button.
   1. On the **Confirm settings** page, review and click the **Done** button.

{{< /details >}}

## Installation

For the next several steps we'll use a shell via the [`ssh` service](https://kb.synology.com/en-id/DSM/tutorial/How_to_login_to_DSM_with_root_permission_via_SSH_Telnet), logging in with the new user. If you use a custom SSH port, be sure to update the `-p` flag.

{{< shell >}}
{{< shellin >}}ssh -p 22 -l certupdater nas.internal.example.com{{< /shellin >}}
{{< /shell >}}

Check for the [latest `acme.sh` release](https://github.com/acmesh-official/acme.sh/releases), then download and extract it to the user's home directory.

{{< shell >}}
{{< shellin >}}mkdir -p src/acme.sh{{< /shellin >}}
{{< shellin >}}wget -O- https://github.com/acmesh-official/acme.sh/archive/refs/tags/3.0.2.tar.gz | tar -xzf- --strip-components=1 --C=src/acme.sh{{< /shellin >}}
{{< /shell >}}

Next, create a wrapper script to hold some of the common settings we'll use. The `SYNO_*` variables are used by the [`synology_dsm` deploy hook](https://github.com/acmesh-official/acme.sh/wiki/Synology-NAS-Guide) for connecting and updating certificates, so it needs the user's password to authenticate.

{{< shell >}}
{{< shellin >}}touch acme.sh{{< /shellin >}}
{{< shellin >}}chmod +x acme.sh{{< /shellin >}}
{{< shellin >}}vim acme.sh{{< /shellin >}}
{{< /shell >}}

```bash
#!/bin/bash

export SYNO_Certificate='acme'
export SYNO_Username="${USER}"
export SYNO_Password='replace-me' # TODO
export SYNO_Create='1'

exec ./src/acme.sh/acme.sh "$@"
```

If you're using a private ACME directory, you'll probably want to include a copy of your internally-trusted certificate(s). This file is referenced in later commands with the `--ca-bundle` flag.

{{< shell >}}
{{< shellin >}}vim ca.crt{{< /shellin >}}{{< /shell >}}

```
-----BEGIN CERTIFICATE-----
...snip...
-----END CERTIFICATE-----
```

Since we'll be relying on a web request to verify the certificate request, we also need to make sure the directory is configured. The web server configuration references it by default, but the directory may not already exist. Run the following with `sudo` (password will be required) to make sure it exists and the user can manage it.

{{< shell >}}
{{< shellin >}}sudo mkdir -p /var/lib/letsencrypt/.well-known/acme-challenge{{< /shellin >}}
{{< shellin >}}sudo chown -R certupdater:http /var/lib/letsencrypt{{< /shellin >}}{{< /shell >}}

## Certificate Setup

Once our utilities and configuration are in place, we can request our initial certificate with the `--issue` command. Be sure to review and update the arguments for your server and domain before running the following:

{{< shell >}}
{{< shellin >}}./acme.sh --issue \
--webroot /var/lib/letsencrypt \
--ca-bundle ca.crt \
--server https://pki.internal.example.com/acme/directory \
--domain nas.internal.example.com{{< /shellin >}}
{{< shellout summary="Sample Output" >}}[Fri Apr  8 10:29:41 MDT 2022] Using CA: https://pki.internal.example.com/acme/directory
[Fri Apr  8 10:29:41 MDT 2022] Single domain='nas.internal.example.com'
[Fri Apr  8 10:29:41 MDT 2022] Getting domain auth token for each domain
[Fri Apr  8 10:29:41 MDT 2022] Getting webroot for domain='nas.internal.example.com'
[Fri Apr  8 10:29:41 MDT 2022] Verifying: nas.internal.example.com
[Fri Apr  8 10:29:41 MDT 2022] Success
[Fri Apr  8 10:29:41 MDT 2022] Verify finished, start to sign.
[Fri Apr  8 10:29:41 MDT 2022] Lets finalize the order.
[Fri Apr  8 10:29:41 MDT 2022] Le_OrderFinalize='https://pki.internal.example.com/acme/order/TFtQ3YFptbEGXMx4wtubzxetGLdpxjmb/finalize'
[Fri Apr  8 10:29:41 MDT 2022] Downloading cert.
[Fri Apr  8 10:29:41 MDT 2022] Le_LinkCert='https://pki.internal.example.com/acme/certificate/D5XCEPWEUh321t4kIgU4CD75e1tLOdAW'
[Fri Apr  8 10:29:42 MDT 2022] Cert success.
-----BEGIN CERTIFICATE-----
...snip...
-----END CERTIFICATE-----
[Fri Apr  8 10:29:42 MDT 2022] Your cert is in: /volume1/homes/certupdater/.acme.sh/nas.internal.example.com/nas.internal.example.com.cer
[Fri Apr  8 10:29:42 MDT 2022] Your cert key is in: /volume1/homes/certupdater/.acme.sh/nas.internal.example.com/nas.internal.example.com.key
[Fri Apr  8 10:29:42 MDT 2022] The intermediate CA cert is in: /volume1/homes/certupdater/.acme.sh/nas.internal.example.com/ca.cer
[Fri Apr  8 10:29:42 MDT 2022] And the full chain certs is there: /volume1/homes/certupdater/.acme.sh/nas.internal.example.com/fullchain.cer{{< /shellout >}}{{< /shell >}}

After it completes, the certificate details and configuration are stored in the `~/.acme.sh` directory, but we still need to configure it within DSM. Use the `--deploy` command along with the built-in hook for `synology_dsm` (which uses the `SYNO_` environment variables from earlier).

{{< shell >}}
{{< shellin >}}./acme.sh --deploy \
--deploy-hook synology_dsm \
--domain nas.internal.example.com{{< /shellin >}}
{{< shellout summary="Sample Output" >}}[Fri Apr  8 10:31:04 MDT 2022] Logging into localhost:5000
[Fri Apr  8 10:31:04 MDT 2022] Getting certificates in Synology DSM
[Fri Apr  8 10:31:04 MDT 2022] Generate form POST request
[Fri Apr  8 10:31:04 MDT 2022] Upload certificate to the Synology DSM
[Fri Apr  8 10:31:04 MDT 2022] http services were NOT restarted
[Fri Apr  8 10:31:04 MDT 2022] Success{{< /shellout >}}{{< /shell >}}

If this is the first time the `acme` certificate is created on the server, you'll need to review the Certificate settings from the Security control panel to make sure it becomes the default and is used by services.

{{< details summary="Detailed Steps (DSM 7)" >}}

1. Log in to Synology DSM as an administrator.
1. Open the **Control Panel** application.
1. Go to the **Security** component.
1. Go to the **Certificate** tab.
1. Select the **acme** certificate that was just created.

   1. Activate **Set as default certificate**.
   1. Click the **OK** button.

1. Click the **Settings** button.

   1. For all existing services, change **Certificate** to be the new **acme** certificate.
   1. Click the **OK** button. It may take a few moments to apply and reconnect to DSM.

{{< /details >}}

Finally, if you reload your browser's connection to DSM you should see it using your new certificate. At this point we're done with the shell, so don't forget to close the connection.

{{< shell >}}
{{< shellin >}}exit{{< /shellin >}}
{{< /shell >}}

## Certificate Renewal

ACME certificates are typicaly shorter-lived, so we want to make sure the renewal and update process is automated. Although `acme.sh` has an `--install` command to configure `cron`, it feels more appropriate to use DSM's [Task Scheduler](https://kb.synology.com/en-uk/DSM/help/DSM/AdminCenter/system_taskscheduler?version=7) to configure a task which runs `./acme.sh --cron` daily as the `certupdater` user.

{{< details summary="Detailed Steps (DSM 7)" >}}

1. Log in to Synology DSM as an administrator.
1. Open the **Control Panel** application.
1. Go to the **Task Scheduler** component.
1. Click **Create**, choose **Scheduled Task**, and choose **User-defined script**.
1. From the **General** tab...

    1. For **Task**, enter **acme.sh --cron**.
    1. For **User**, choose **certupdater**.

1. From the **Schedule** tab...

    1. For **Run on the following days**, choose **Daily**.
    1. For **Time**, choose an off-peak time to run every day, such as **04:55**.

1. From the **Task Settings** tab...

    1. For **User-defined script**, enter **/bin/bash -c './acme.sh --cron 2>&1 >> task.log'**. *This uses `bash` and `task.log` to keep the log output in the user's home directory.*
    1. Optionally, configure **Send run details by email**.

1. Click the **OK** button.

{{< /details >}}

Once configured, it should automatically renew the certificates when it's approaching the expiration. You can test the behavior by temporarily editing the task to add the `--force` option to the `acme.sh` call, and then use the **Run** button from the task list. Any success and error messages should be reported in the log output.

## Troubleshooting

### libcurl error 60

> ```
> Please refer to https://curl.haxx.se/libcurl/c/libcurl-errors.html for error code: 60
> Can not init api for: https://pki.internal.example.com/acme/directory.
> ```

Review your `ca-bundle` flag and file to make sure it has the certificate being used by your ACME directory. Learn more about the error [here](https://curl.se/libcurl/c/libcurl-errors.html#CURLEPEERFAILEDVERIFICATION), and try debugging the connection with `curl` directly:

{{< shell >}}
{{< shellin >}}curl -v https://pki.internal.example.com/acme/directory --cacert ca.crt{{< /shellin >}}{{< /shell >}}

### Can not write token to file

> ```
> Can not write token to file : /var/lib/letsencrypt/.well-known/acme-challenge/7NckvfDW0gO98NgeOXcChpSBqBb7bLFE
> ```

Make sure the `/var/lib/letsencrypt` directory exists and has the correct permissions. Alternatively, recreate it with the steps earlier in the guide.
