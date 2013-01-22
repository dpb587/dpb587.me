---
title: OpenGrok CLI
layout: post
tags: opengrok php symfony xpath
description: Making it easier to search code from the command line.
code: https://github.com/dpb587/opengrok-cli
---

One tool that makes my life as a software developer easier is [OpenGrok][1] - it lets me quickly find application code
and it knows more context than a simple `grep`. It has a built-in web interface, but sometimes I want to work with
search results from the command line (particularly for automated tasks). Since I couldn't find an API, I created a
command to load and parse results using [symfony/console][3] and [xpath][4].


### Usage

It's straightforward to use, just provide the OpenGrok server, project to search, and the query. Mimicking grep, the
output format should look familiar:

{% highlight console %}
$ opengrok-cli --server=http://lxr.php.net --project=PHP_5_4 oci_internal_debug
/ext/oci8/oci8.c:777: PHP_FUNCTION(oci_internal_debug);
/ext/oci8/oci8.c:862: 	PHP_FE(oci_internal_debug,			arginfo_oci_internal_debug)
/ext/oci8/oci8.c:932: 	PHP_FALIAS(ociinternaldebug,	oci_internal_debug,		arginfo_oci_internal_debug)
/ext/oci8/oci8_interface.c:1307: /* {{ "{{{" }} proto void oci_internal_debug(int onoff)
/ext/oci8/oci8_interface.c:1309: PHP_FUNCTION(oci_internal_debug)
{% endhighlight %}

When run from an ANSI-friendly terminal, the output is nicely colorized. And just like the web interface, the `query`
argument can include operators, nested queries, field specifiers, and wildcard searches.

It also has a `--list` option to only output paths. Useful if I'm in the repository's top-level and I want to work
through all the results with `vim`:

{% highlight console %}
$ cd php-src/
$ export OPENGROK_SERVER=http://lxr.php.net OPENGROK_PROJECT=PHP_5_4
$ vim $(opengrok-cli --list refs:PHP_MODE_PROCESS_STDIN)
{% endhighlight %}


### Open Source

I published the code to [dpb587/opengrok-cli][5]. Check the `README`, but it's easy to get started:

{% highlight console %}
$ git clone git://github.com/dpb587/opengrok-cli.git opengrok-cli && cd !$
$ php composer.phar install
$ ./bin/opengrok-cli --help
{% endhighlight %}

Or take the easier route and use the pre-compiled version:

{% highlight console %}
$ wget static.dpb587.me/opengrok-cli.phar
$ php opengrok-cli.phar --help
{% endhighlight %}


 [1]: http://hub.opensolaris.org/bin/view/Project+opengrok/
 [2]: http://ctags.sourceforge.net/
 [3]: http://symfony.com/doc/master/components/console/introduction.html
 [4]: http://us.php.net/domxpath
 [5]: https://github.com/dpb587/opengrok-cli
