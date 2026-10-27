---
title: File Server Backups + Read-only rsync
description: Working around "rsync code 43" with a tiny SSH container.
publishDate: 2026-10-26
---

I use small VMs on my local network to run things like DNS and [Home Assistant](https://www.home-assistant.io). Some of them have persistent data, but I haven't had a reliable backup strategy for them yet. This weekend I consolidated the backups to run from Synology DSM via their [Active Backup for Business](https://www.synology.com/en-us/dsm/feature/active-backup-business/pc%20) (ABB) feature. Working backwards from what finally worked...

* [Read-only SSH+rsync Container](#read-only-ssh-rsync-container) with an ABB File Server connection
* [abb-cli uses a custom CA bundle](#failed-attempt-1) ignoring self-signed CAs already trusted by the host.
* [Full ISO image backups were too large](#failed-attempt-1b) when using an ABB Virtual Machine connection.
* [Incompatible `rrsync` usage](#failed-attempt-2) because ABB wants to use the `-s` option.

## Read-only SSH+rsync Container {#read-only-ssh-rsync-container}

Ultimately, I ended up using a dedicated, read-only containerized SSH server which mounts the persistent data that I want backed up. Most services are already provisioned as containers, so this was a quick addition. This approach also helped avoid extra changes to the host's users and SSH configuration. A few other principles I ended up applying here:

* **`rsyncbackup` user** - runs as `uid=0` for access to multi-user service directories, but seeing `root` in the backup manager just felt... security-wrong.
* **read-only container and mount** - avoids any unexpected modifications of persistent or container image files.
* **`ForceCommand` option** - forces the use of the `rsync` command. Not quite `rrsync` containment, but close enough in this setup and it keeps Synology happy.

It's worth noting that this approach does not try to pause the services or disks for the snapshot. For now, at least, that's less of a concern for these kinds of VMs. For more technical details, review the small [`rsync-sender.sh`]([rsync-sender.sh]({{< appendix-ref "2026-10-26-file-server-backups-read-only-rsync/rsync-sender.sh" >}})), [sshd_config]({{< appendix-ref "2026-10-26-file-server-backups-read-only-rsync/sshd_config" >}}), or [`Dockerfile`]({{< appendix-ref "2026-10-26-file-server-backups-read-only-rsync/Dockerfile" >}}) files.

### Basic Usage {#basic-usage}

As part of my deployment script, the following steps are effectively run on each VM.

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    # generate a host-specific ssh server host key
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    ssh-keygen -t ed25519 -N '' -f /srv/rsync-sender/ssh_host_ed25519_key
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    # add an externally-generated ssh public key (ABB uses the private key)
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    cat > /srv/rsyncbackup/authorized_keys
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    # run the container, one way or another
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    podman run -d --name rsyncbackup --read-only \
      -p 20261:22 \
      -v /srv/rsyncbackup/ssh_host_ed25519_key:/etc/ssh/ssh_host_ed25519_key:ro \
      -v /srv/rsyncbackup/authorized_keys:/etc/ssh/authorized_keys:ro \
      -v /mnt/persistent:/mnt/persistent:ro \
      rsyncbackup
    ```

  {{< /terminal-input >}}

{{< /terminal >}}

### Sanity Checks {#sanity-checks}

After the container is up, you can test it out with a few commands...

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    # read operations should succeed
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    rsync -a -s -e 'ssh -p 20261 -i ./id_ed25519' rsyncbackup@192.0.2.42:/mnt/persistent/ ./restore-test/
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    # write operations should be rejected
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    rsync -a -e 'ssh -p 20261 -i ./id_ed25519' /dev/urandom rsyncbackup@192.0.2.42:/mnt/persistent/unwritable
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    error: only read-only rsync is permitted
    rsync: connection unexpectedly closed (0 bytes received so far) [sender]
    rsync error: error in rsync protocol data stream (code 12) at io.c(232) [sender=3.2.7]
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    # read-then-sneaky-writes should fail due to the read-only filesystem
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    rsync -a -s --remove-source-files -e 'ssh -p 20261 -i ./id_ed25519' rsyncbackup@192.0.2.42:/mnt/persistent/ ./restore-test/
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    rsync: [sender] sender failed to remove example.txt: Read-only file system (30)
    rsync error: some files/attrs were not transferred (see previous errors) (code 23) at main.c(1887) [generator=3.2.7]
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

### Configure Active Backup for Business {#configure-active-backup-for-business}

Once provisioned, set up for ABB is pretty straight-forward.

1. Open *Active Backup for Business*.
1. Navigate to *File Server*.
1. Click *Add Server*.
1. Choose *rsync server* and click *Next*.
1. For *Connection Method*, be sure that *rsync shell mode via SSH* is selected.
1. Fill out the rest of the connection parameters (including the container's non-standard SSH port).
1. Click *Done*, and a quick verification test should confirm the configuration.
1. Configure the backup strategy (e.g. mirror, incremental) and target storage.
1. Optionally, initiate a backup immediately.

From there, Synology DSM will provide notifications on success/failure, and you might want to configure other offsite backups for these backups, depending on your strategy.

Of course, a final solution is rarely the first, so now I need to document what I should avoid trying again...

## Failed Attempt #1: Native Linux Agent {#failed-attempt-1}

I had first started by installing the native client software for ABB. And the first hiccup came before any backups - connecting to my NAS which uses a private CA's certificate.

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    sudo abb-cli -c
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    Enter server address: nas.example.com
    Enter username: abb-client
    Enter password: 
    Connecting...
    The SSL certificate of the Synology NAS is not trusted. To learn how to obtain a valid certificate, please have a registered domain by setting up DDNS (http://sy.to/ddns) and applying for Let's Encrypt (http://sy.to/letsencrypt) certificate. Proceed anyway? [y/n](default: y): 
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

The private CA was already installed in the host's trusted certificates, so that was weird. A quick Google search and I found a workaround for manually installing my private CA to Synology's trust store.

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    sudo systemctl stop synology-active-backup-business-linux-service
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    sudo rm -f /opt/Synology/ActiveBackupforBusiness/data/certs.pem
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    curl https://pki.example.com/roots.pem | sudo tee /opt/Synology/ActiveBackupforBusiness/data/certs.pem
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    sudo systemctl start synology-active-backup-business-linux-service
    ```

  {{< /terminal-input >}}

{{< /terminal >}}

After that, `sudo abb-cli -c` connected without complaint.

## Failed Attempt #1b: Full VM ISO Images {#failed-attempt-1b}

Once my VM was connected to ABB, I started an initial backup. I quickly learned it creates a full disk image as the backup - no way to point it at just `/mnt/persistent`. Considering that some of the persistent data is under 100 MB *and* the VM/OS/containers are disposable and recreatable... it seemed like a poor use of disk space. Especially considering further bloat for multiple backup versions and for some servers that end up in additional, offsite backups.

## Failed Attempt #2: incompatible `rrsync` options {#failed-attempt-2}

Avoiding full VM backups, I next went with the `rsync` approach. More specifically, [`rrsync`](https://download.samba.org/pub/rsync/rrsync.1), which ships with `rsync`, and restricts access to a single directory. I created a backup user on the VM with a `ForceCommand` of `/usr/bin/sudo -n /usr/bin/rrsync -ro /mnt/persistent`.

After verifying it manually, I switched over to ABB, created a new File Server with rsync, clicked next, and then... error.

{{< image alt="ABB error: Server can be reached but the given port is not opened. ( [rsync code 43])" src="./media/abb-error-rsync-code-43.png" caption="rsync code 43 (Active Backup for Business)" >}}

Confusingly, 43 is not an error code according to [`rsync` man pages](https://linux.die.net/man/1/rsync). Finding nothing in Synology's *Log Center*, I connected to the NAS via SSH and took a look at `journalctl`. In the logs, I found the following...

```
Oct 24 11:11:22 nas synoscgi_SYNO.ActiveBackup.Agentless_1_test_connection[5610]: synoabk_synoscgi::[ERROR] rsync-tool.cpp:225(5610,140305762057856) Failed to test connection [43] Warning: Permanently added '192.0.2.42' (ECDSA) to the list of known hosts.
Oct 24 11:11:22 nas synoscgi_SYNO.ActiveBackup.Agentless_1_test_connection[5610]: synoabk_synoscgi::[ERROR] rsync-tool.cpp:225(5610,140305762057856) Failed to test connection [43] /usr/bin/rrsync error: option -s has been disabled on this server.
Oct 24 11:11:22 nas synoscgi_SYNO.ActiveBackup.Agentless_1_test_connection[5610]: synoabk_synoscgi::[ERROR] rsync-tool.cpp:225(5610,140305762057856) Failed to test connection [43] rsync error: rsync service is no running (code 43) at io.c(254) [Receiver=3.1.2]
Oct 24 11:11:22 nas synoscgi_SYNO.ActiveBackup.Agentless_1_test_connection[5610]: synoabk_synoscgi::[ERROR] rsync-tool.cpp:227(5610,140305762057856) Failed to test connection [43] result
Oct 24 11:11:22 nas synoscgi_SYNO.ActiveBackup.Agentless_1_test_connection[5610]: synoabk_synoscgi::[ERROR] protocol-wrapper.cpp:248(5610,140305762057856) ProtocolWrapper: Failed to TestConnection '-17', '0', ' [rsync code 43]'
```

Notably, `/usr/bin/rrsync error: option -s has been disabled on this server` which indicated it had indeed connected to the port and `rrsync` process. Apparently, ABB uses the following flag in its connection parameters.

```
-s, --protect-args          no space-splitting; wildcard chars only
```

The flag is used to prevent interpreting of special characters (which is a convenience for some setups), but, sadly, I could not find a way to disable it in ABB. I also took a quick look at the `rrsync` wrapper source code, but it seemed a little silly to be disabling the singular option the command disabled by default -- as if it were there for good reason.
