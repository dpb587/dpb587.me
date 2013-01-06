---
title: Secure Git Repositories
layout: post
tags: git security
---

I use private repositories on [GitHub][1], but I still don't feel quite comfortable pushing sensitive data like
passwords, keys, and account information. Typically that information ends up just sitting on my local machine or in my
head ready for me to pull up as needed. It would be much better if that information was a bit more fault tolerant and,
even better, if I could follow similar workflows as the rest of my application code.

After some research I discovered [gist 873637][2] which discusses using `git`'s clean and smudge [filters][4] to pass
files through `openssl` for decryption and encryption. Soon I found the much more useful [`shadowhand/git-encrypt`][3]
tool.


### Initial Setup

First, I installed `gitcrypt` on my machine:

{% highlight console %}
$ git clone git://github.com/shadowhand/git-encrypt.git /usr/local/git-encrypt
$ chmod +x /usr/local/git-encrypt/gitcrypt
$ ln -s /usr/local/git-encrypt/gitcrypt /usr/local/bin/gitcrypt
{% endhighlight %}

Next, I created a new repo and used `gitcrypt init` with the auto-generated defaults:

{% highlight console %}
$ mkdir fort-knox ; cd !$
$ git init
$ gitcrypt init
Generate a random salt? [Y/n] Y
Generate a random password? [Y/n]Y
What encryption cipher do you want to use? [aes-256-ecb] 

This configuration will be stored:

salt:   7d9f6cc1512aa2b5
pass:   EAC8405A-DD64-43A3-A17F-EB28195B4B1E
cipher: aes-256-ecb

Does this look right? [Y/n] 
Do you want to use .git/info/attributes? [Y/n] 
What files do you want encrypted? [*] 
{% endhighlight %}

Now I just have to be sure to securely keep the salt and pass elsewhere for the next time I setup this repo. Other than
that, it's ready for me to use like any other `git` repository.


### A Practical Bit

Since I won't frequently be setting up this repository, it'd probably be best if I could keep a reminder about what I'll
need to do. So I add a `.gitattributes` file which excludes itself and README from encryption:

{% highlight vim %}
* filter=encrypt diff=encrypt
README -filter -diff
.gitattributes -filter -diff

[merge]
    renormalize=true
{% endhighlight %}

And include the necessary commands and reference in the README:

{% highlight console %}
$ git clone git@github.com:dpb587/fort-knox.git fort-knox ; cd !$
$ gitcrypt init # https://github.com/shadowhand/git-encrypt
$ git reset --hard HEAD
{% endhighlight %}


### Summary

With `gitcrypt` I can work with a repository and enjoy extra security on top of the redundancy and version control that
`git` provides. The only difference from my regular repos is I can't really view my files from [github.com][1] (with the
convenient exception of the README).


 [1]: https://github.com/
 [2]: https://gist.github.com/873637
 [3]: https://github.com/shadowhand/git-encrypt
 [4]: http://git-scm.com/book/ch7-2.html#Keyword-Expansion
