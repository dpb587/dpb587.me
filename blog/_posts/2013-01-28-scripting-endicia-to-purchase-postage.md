---
title: Scripting Endicia to Purchase Postage
layout: post
tags: [ 'applescript', 'endicia', 'loopy' ]
description: Automating user interactions with AppleScript.
code: https://gist.github.com/dpb587/4660132
---

We currently use [Endicia for Mac][1] for postage processing at Loopy. We rarely use the UI since I've scripted most of
it, but one annoyance had been to regularly open it up and add postage since it doesn't reload automatically. If we
happen to forget, it ends up blocking things until we notice. I finally got around to scripting that, too.


## Scripted

In real life, whenever the balance gets too low it throws up an alert and you need to click through a few menus, select
a purchase amount, and confirm the selection before the application will continue. Using [System Events][2], it can all
be conveniently automated. Using [the script][4] I wrote, $500 can be purchased by running:

{% highlight console %}
$ osascript endicia-postage-purchase.applescript 500
ok
{% endhighlight %}

With that step automated, it can be tied in with the `endiciatool` output -- whenever `<Balance />` drops below $30,
automatically kick off the script to buy more postage.


## Summary

So now that's one less manual step everybody has to worry about, saving some time and hassle. If you happen to be new to
[Endicia][3], you should check them out (and use the promotional code <code>599888</code>). Their software has been a
valuable timesaver for us.


 [1]: http://www.dymoendicia.com/segments/all-products/endicia-for-mac
 [2]: https://developer.apple.com/library/mac/#documentation/applescript/conceptual/applescriptx/Concepts/as_related_apps.html#//apple_ref/doc/uid/TP40001570-1149074-BAJEIHJA
 [3]: http://www.dymoendicia.com/
 [4]: https://gist.github.com/dpb587/4660132#file-endicia-purchase-postage-applescript
