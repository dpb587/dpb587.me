---
title: Comparing PHP Application Definitions
layout: post
tags: code diff language php xslt
description: Identifying how classes/interfaces changed between versions.
code: https://github.com/dpb587/diff-defn.php
---

While working to update a PHP project, I thought it'd be helpful if I could systematically qualify significant code
changes between versions. I could weed through a massive line diff, but that's costly if many changes aren't ultimately
affecting my API dependencies. Typically I only care about how interfaces and classes change in their usage of methods,
method arguments, variables, and scope.

I did a bit of research on the idea and found [several][7] [different][8] [questions][9], a few [referenced][10]
[products][11], and a short [article][12] on the idea. However I wasn't able to find a good PHP (or even generic) option
which was open-source and something I could easily try out.

To that end, I made a prototype for a language-intelligent/OOP-diff/structured diff engine that can summarize many of
the programmatic changes in an easily readable report which links definitions back to their file and line number for
more detailed review...

<img height="343" src="/blog-data/2013-03-07-comparing-php-application-definitions/console-diff.png" width="536" />


### Usage

If I were upgrading my application with a [`symfony/Console`][1] dependency from `v2.0.22` to `v2.2.0`, I could generate
the diff of definitions with:

{% highlight console %}
$ git clone git://github.com/dpb587/diff-defn.php.git diff-defn.php && cd !$
$ php composer.phar install
$ ./bin/diff-defn.php diff:show --exclude='/Tests/' git://github.com/symfony/Console.git v2.0.22 v2.2.0 > output.html
$ open output.html
{% endhighlight %}

Take a look at several other reports using the default stylesheet:

 * [`doctrine/dbal`][2] (`2.1.7` &rarr; `2.3.2`)
 * [`fabpot/Twig`][3] (`v1.10.0` &rarr; `v1.12.2`)
 * [`symfony/symfony`][4] (`v2.0.22` &rarr; `v2.2.0`)
 * [`zendframework/zf2`][5] (`release-2.0.0` &rarr; `release-2.1.3`)


### Behind the Scenes

The logic behind the command looks like:

 1. Use version control to diff the two versions and see what files were changed.
 2. Use [nikic/php-parser][6] to parse the PHP files in both their initial and final commit...
     * Build separate structures for both the initial and final code states.
     * Use visitors to analyze definitions, language or application-specific definitions.
 3. Use some logic to compare the initial and final structures and create a new structured diff with only the relevant
    definitions that changed (including both old and new).
 4. Apply a stylesheet to the diff structure to generate human-readable output.

The structures are simple classes which can be dumped to XML. And technically, aside from the step of parsing of PHP
files, this is very language-agnostic. For example, the XML representation of the initial or final commit looks like:

{% highlight xml %}
<root id="root">
    <defn id="source" repository="git://github.com/symfony/Security.git" repository-link="https://github.com/symfony/Security/" file-link="https://github.com/symfony/Security/blob/%commit%/%file%#L%line%" commit-link="https://github.com/symfony/Security/tree/%commit%">
        <defn id="commit" value="8cd00e30f4a13b0c57c5d98613c3dd533bc1c35a" friendly="v2.0.22"/>
    </defn>
    <class id="Symfony\Component\Security\Http\Firewall\UsernamePasswordFormAuthenticationListener">
        <defn-source id="source" file="Http/Firewall/UsernamePasswordFormAuthenticationListener.php" line="33"/>
        <class-extends id="Symfony\Component\Security\Http\Firewall\AbstractAuthenticationListener"/>
        <class-property id="csrfProvider">
            <defn-source id="source" file="Http/Firewall/UsernamePasswordFormAuthenticationListener.php" line="35"/>
            <defn-attr id="visibility" value="private"/>
        </class-property>
        <function id="__construct">
            <defn-source id="source" file="Http/Firewall/UsernamePasswordFormAuthenticationListener.php" line="40"/>
            <defn-attr id="visibility" value="public"/>
            <function-param id="securityContext">
                <defn-attr id="typehint" value="Symfony\Component\Security\Core\SecurityContextInterface"/>
            </function-param>
            <!-- ... -->
            <function-param id="providerKey"/>
            <function-param id="options">
                <defn-attr id="default" type="array" value="[]"/>
                <defn-attr id="typehint" value="array"/>
            </function-param>
            <!-- ... -->
            <function-param id="logger">
                <defn-attr id="default" type="const" value="null"/>
                <defn-attr id="typehint" value="Symfony\Component\HttpKernel\Log\LoggerInterface"/>
            </function-param>
            <!-- ... -->
        </function>
        <function id="attemptAuthentication">
            <defn-source id="source" file="Http/Firewall/UsernamePasswordFormAuthenticationListener.php" line="56"/>
            <defn-attr id="visibility" value="protected"/>
            <function-param id="request">
                <defn-attr id="typehint" value="Symfony\Component\HttpFoundation\Request"/>
            </function-param>
        </function>
    </class>
</root>
{% endhighlight %}

And after the initial and final commit are compared, the resulting structured diff looks like:

{% highlight xml %}
<root id="root" diff="touched">
    <defn id="source" repository="git://github.com/symfony/Security.git" repository-link="https://github.com/symfony/Security/" file-link="https://github.com/symfony/Security/blob/%commit%/%file%#L%line%" commit-link="https://github.com/symfony/Security/tree/%commit%" diff="touched">
        <defn id="commit" value="9e53793548e403c155d28a01153026905ee53d5d" friendly="v2.2.0" diff="changed">
            <diff-old id="old">
                <defn id="commit" value="8cd00e30f4a13b0c57c5d98613c3dd533bc1c35a" friendly="v2.0.22"/>
            </diff-old>
        </defn>
    </defn>
    <class id="Symfony\Component\Security\Http\Firewall\UsernamePasswordFormAuthenticationListener" diff="touched">
        <defn-source id="source" file="Http/Firewall/UsernamePasswordFormAuthenticationListener.php" line="33"/>
        <function id="__construct" diff="touched">
            <defn-source id="source" file="Http/Firewall/UsernamePasswordFormAuthenticationListener.php" line="40"/>
            <function-param id="logger" diff="touched">
                <defn-attr id="typehint" value="Psr\Log\LoggerInterface" diff="changed">
                    <diff-old id="old">
                        <defn-attr id="typehint" value="Symfony\Component\HttpKernel\Log\LoggerInterface"/>
                    </diff-old>
                </defn-attr>
            </function-param>
        </function>
        <function id="requiresAuthentication" diff="added">
            <defn-source id="source" file="Http/Firewall/UsernamePasswordFormAuthenticationListener.php" line="56" diff="added"/>
            <defn-attr id="visibility" value="protected" diff="added"/>
            <function-param id="request" diff="added">
                <defn-attr id="typehint" value="Symfony\Component\HttpFoundation\Request"/>
            </function-param>
        </function>
    </class>
</root>
{% endhighlight %}


### Going Further

Being able to parse files and have their differences stored in static, semi-agnostic format allows for some interesting
usages:

 * search for specific changes, like which class methods have had their typehint changed (e.g. xpath
   `//class/function/function-param/defn-attr[@id="typehint" and @diff="changed" and @value="Psr\Log\LoggerInterface"]`)
 * combine search results with other automated tools for updating impacted application code or explicitly requiring
   reviews for changes breaking compatibility standards
 * generating lists about new interfaces/classes, dropped definitions, newly limited scopes
 * when using test naming conventions, specifically test and verify the code's tests are run
 * instead of simple "lines of code" stats, also track classes/methods/functions
 * writing post-commit rules based on definition searches (e.g. email a maintainer whenever a critical class is touched)

Since the analysis and the serialized, static representation are distinct steps, this also allows for custom,
application-specific analysis information like:

 * in aspect-oriented code, analyzing `@Aspects(...)` and including them in reports
 * tying code-linting tool results to flag specific methods/properties that have issues
 * additional flags to monitor if function logic changed vs formatting/comments (even if the API is unchanged)

And unlike some of the other tools I ran into, the static representation is not itself inherently readable; it needs a
stylesheet to make it human-friendly. This makes the results potentially reusable for multiple different reports.


### Summary

I've published this work-in-progress code to [dpb587/diff-defn.php][13] in case you want to try it out with your own PHP
repositories. It's certainly not a replacement of reading changelogs and understanding what upstream changes are being
made, but I have found it interesting and helpful to identifying breaking changes.


  [1]: https://github.com/symfony/Console
  [2]: http://static.dpb587.me/2013-03-07-comparing-php-application-definitions/doctrine-dbal-2.1.7..2.3.2.html
  [3]: http://static.dpb587.me/2013-03-07-comparing-php-application-definitions/fabpot-Twig-v1.10.0..v1.12.2.html
  [4]: http://static.dpb587.me/2013-03-07-comparing-php-application-definitions/symfony-symfony-v2.0.22..v2.2.0.html
  [5]: http://static.dpb587.me/2013-03-07-comparing-php-application-definitions/zendframework-zf2-release-2.0.0..release-2.1.3.html
  [6]: https://github.com/nikic/php-parser
  [7]: http://stackoverflow.com/questions/77931/do-you-know-of-any-language-aware-diffing-tools
  [8]: http://stackoverflow.com/questions/2828795/is-there-a-language-aware-diff
  [9]: http://discuss.fogcreek.com/joelonsoftware5/default.asp?cmd=show&ixPost=155585&ixReplies=18
 [10]: http://www.semdesigns.com/Products/SmartDifferencer/index.html
 [11]: http://www.schneidersoft.com/Products/OOP-DIFF/OOP-DIFF.aspx
 [12]: http://www.itworld.com/software/231515/usenix-dartmouth-expanding-diff-grep-unix-tools
 [13]: https://github.com/dpb587/diff-defn.php
