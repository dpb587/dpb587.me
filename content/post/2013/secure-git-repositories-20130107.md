---
description: Seamless data encryption of repository files.
params:
    nav:
        tag:
            git: true
            security: true
publishDate: "2013-01-07"
title: Secure Git Repositories
---

I use private repositories on [GitHub][1], but I still don't feel quite comfortable pushing sensitive data like
passwords, keys, and account information. Typically that information ends up just sitting on my local machine or in my
head ready for me to pull up as needed. It would be much better if that information was a bit more fault tolerant and,
even better, if I could follow similar workflows as the rest of my application code.

After some research I discovered [gist 873637][2] which discusses using `git`'s clean and smudge [filters][4] to pass
files through `openssl` for decryption and encryption. The result is `git`'s indexes only containing encrypted file
contents in base64. Soon I found [`shadowhand/git-encrypt`][3].


# Initial Setup {#initial-setup}

First, I did a one-time install of `shadowhand/git-encrypt` on my machine:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    git clone git://github.com/shadowhand/git-encrypt.git /usr/local/git-encrypt
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    chmod +x /usr/local/git-encrypt/gitcrypt
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    ln -s /usr/local/git-encrypt/gitcrypt /usr/local/bin/gitcrypt
    ```

  {{< /terminal-input >}}

{{< /terminal >}}

Next, I created a new repo and use `gitcrypt init` to set things up:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    mkdir fort-knox && cd !$
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    git init
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    Initialized empty Git repository in /private/tmp/fort-knox/.git/
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    gitcrypt init
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    Generate a random salt? [Y/n] Y
    Generate a random password? [Y/n]Y
    What encryption cipher do you want to use? [aes-256-ecb] 

    This configuration will be stored:

    salt:   7d9f6cc1512aa2b5
    pass:   EAC8405A-DD64-43A3-A17F-EB28195B4B1E
    cipher: aes-256-ecb

    Does this look right? [Y/n] Y
    Do you want to use .git/info/attributes? [Y/n] n
    What files do you want encrypted? [*] 
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

Now I just have to be sure to securely keep the salt and pass elsewhere for the next time I setup this repo. Other than
that, it's ready for me to use like any other `git` repository.


# A Practical Bit {#a-practical-bit}

Since I won't frequently be setting up this repository, it'd probably be best if I could keep a reminder about what I'll
need to do. So I update `.gitattributes` to exclude itself and `README` from encryption:

```ini
* filter=encrypt diff=encrypt
README -filter -diff
.gitattributes -filter -diff

[merge]
    renormalize=true
```

And include the necessary commands and reference in `README`:

```
Remember...

    git clone git@github.com:dpb587/fort-knox.git fort-knox && cd !$
    gitcrypt init # https://github.com/shadowhand/git-encrypt
    git reset --hard HEAD
```

So, my first commit looks like:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    git add .
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    git commit -m 'initial commit'
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [master (root-commit) 1077d71] initial commit
    2 files changed, 7 insertions(+)
    create mode 100644 .gitattributes
    create mode 100644 README
    ```

  {{< /terminal-output >}}

{{< /terminal >}}


# Under the Hood {#under-the-hood}

Originally I was a bit curious and wanted to verify that it's doing what I thought. So I created a simple test file:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    date > top-secret.txt
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    cat top-secret.txt 
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    Mon Jan  7 15:11:22 MST 2013
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    git add top-secret.txt
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    git commit -m 'top secret information'
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [master dd2272a] top secret information
    1 file changed, 1 insertion(+)
    create mode 100644 top-secret.txt
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

After committing I can look at the raw index data to see what's actually being stored:

{{< terminal >}}

    {{< terminal-input >}}

        ```bash
        git ls-tree HEAD
        ```

    {{< /terminal-input >}}

    {{< terminal-output >}}

        ```
        100644 blob 6a9e000e136a20858f65188f849d0bffed48a685	.gitattributes
        100644 blob 2221766ff8694dffa1e11ea5d0e7acd213e22d90	README
        100644 blob e847f7c05236ac1111a0f5495da87fec188d5420	top-secret.txt
        ```

    {{< /terminal-output >}}

    {{< terminal-input >}}

        ```bash
        git cat-file -p 2221766ff8694dffa1e11ea5d0e7acd213e22d90
        ```

    {{< /terminal-input >}}

    {{< terminal-output >}}

        ```
        Remember...

            git clone git@github.com:dpb587/fort-knox.git fort-knox && cd !$
            gitcrypt init # https://github.com/shadowhand/git-encrypt
            git reset --hard HEAD
        ```

    {{< /terminal-output >}}

    {{< terminal-input >}}

        ```bash
        git cat-file -p e847f7c05236ac1111a0f5495da87fec188d5420
        ```

    {{< /terminal-input >}}

    {{< terminal-output >}}

        ```bash
        U2FsdGVkX199n2zBUSqitTy46rTQ8tytPxnYmmdBahPCL5u1SwnPcYcDN+KFNgom
        ```

    {{< /terminal-output >}}

{{< /terminal >}}

As expected, `README` is readable, but `top-secret.txt` is not. I can manually verify my secret data is still there by
decoding it with my key:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    git cat-file -p e847f7c05236ac1111a0f5495da87fec188d5420 \
    | openssl base64 -d -aes-256-ecb -k "EAC8405A-DD64-43A3-A17F-EB28195B4B1E"
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    Mon Jan  7 15:11:22 MST 2013
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

# Summary {#summary}

With `gitcrypt` I can work with a repository and enjoy extra security on top of the redundancy and version control that
`git` provides. The only difference from my regular repos is I can't really view my files from [github.com][1] (with the
convenient exception of `README`).


 [1]: https://github.com/
 [2]: https://gist.github.com/873637
 [3]: https://github.com/shadowhand/git-encrypt
 [4]: http://git-scm.com/book/ch7-2.html#Keyword-Expansion
