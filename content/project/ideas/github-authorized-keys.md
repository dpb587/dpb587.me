---
draft: true
---

# github-authorized-keys

Generate a list of SSH public keys (i.e. `authorized_keys`) from GitHub users and teams.


## Command Line Usage

Provide a team to list all member SSH keys.

    github-authorized-keys dundermifflin/scranton

The output will be a standard [`authorized_keys` file](https://man.openbsd.org/sshd.8#AUTHORIZED_KEYS_FILE_FORMAT).

    ssh-rsa AAAAB3NzaA... dwightschrute@users.noreply.github.com
    ssh-rsa AAAAB3NzaB... michaelscott@users.noreply.github.com
    ssh-rsa AAAAB3NzaC... jimhalpert@users.noreply.github.com

Provide a username to list a specific user.

    github-authorized-keys sadiq

Provide multiple teams and users for larger lists.

    github-authorized-keys dundermifflin/scranton dundermifflin/corporate sadiq

Use `-o` to include option specifications...

    github-authorized-keys dundermifflin/scranton \
      -o no-port-forwarding -o no-x11-forwarding -o no-agent-forwarding \
      -o command="git-shell -c \"$SSH_ORIGINAL_COMMAND\""

To avoid [rate limiting](https://developer.github.com/v3/#rate-limiting) and reference unlisted organization and team members, configure a [personal access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line).

    export GITHUB_TOKEN=a1b2c3d4...
