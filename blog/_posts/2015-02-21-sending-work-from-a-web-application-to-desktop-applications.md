---
title: "Sending Work from a Web Application to Desktop Applications"
layout: "post"
tags: [ "applescript", "automation", "aws-sqs", "box", "dymo", "endicia", "hazel", "launchd", "osx", "phar", "php", "usps" ]
description: "Using queues and PHP to automate third-party applications running on staff workstations."
code: https://github.com/theloopyewe/elfbot
---

I prefer working on the web application side of things, but there are frequently tasks that need to be automated outside the context of a browser and server. For [TLE][10], there's a physical shop where inventory, order, and shipping tasks need to happen, and those tasks revolve around web-based systems of one form or another. To help unify and simplify things for the staff (aka [elves][11]), I've been connecting scripts on the workstations with internal web applications via queues in the cloud.


## Evolution of a bot

Over the past 8+ years, the need for running commands on the desktop has changed. The easiest example to follow is how we have printed shipping labels over the years:

 0. For the first few months, we would copy/paste the address into the [USPS Print & Ship][1] website, click through the shipping options ourselves, print out a label on sticky paper with inkjet, and copy/paste back the delivery confirmation into an order note. Averaging a few orders a day, it was quite manageable.
 0. With more orders we needed something faster, so I created a form posting to USPS which prefilled all the fields. This way, all we needed to do was confirm/print and copy/paste the delivery confirmation back. That helped for a bit longer.
 0. With a growing number of orders, we still needed something more, so we switched to [Endicia][2], a desktop application which had several integration options and the ability to print directly to a label printer. I switched from USPS links/forms to pre-composed links using Endicia's [custom URI handler][3]. This helped speed things up, save money on label paper, and also automatically copied confirmation codes for us to paste.
 0. Occasionally we would have a couple problems with the URI approach, so I changed to using file downloads:

     1. Instead of Endicia's custom URI handler, the server would send a file download with the XML-based postage details.
     1. Using the [watched folder approach][4], OS X would notice the new file and send it to Endicia for printing.

 0. This worked fairly well, but we quickly ran into a few quirks related to AppleScript's watched folder features and browser downloads - some files not being noticed at all or being noticed multiple times. We switched to [Hazel][5] which not only sidestepped the bugs we were seeing, but also provided me with better insight if something failed.
 0. A bit later I discovered the `OutputFile` attribute of the [DAZzle spec][6] which would allow me to capture the results of the printed postage. By using and monitoring a different file extension for the output, I updated the script to parse the results and post the confirmation code to the website. This became an immense timesaver since it would allow postage to be queued instead of having to wait to paste each confirmation code manually. We used this approach for a long time.

Eventually we needed to do more than just printing postage. The Hazel setup was straightforward, but the AppleScript implementation had become a bit too complex and inconvenient to test and change. We also needed this setup to be easily deployed on multiple systems. At this point I decided to spend some time coming up with a different solution which would better meet our needs.


## The Bot

Today's bot operates a bit differently. Rather than depending on monitored folders for file downloads, each workstation has its own queue (via [Amazon SQS][8]). Rather than complex logic in AppleScript, it is primarily based in PHP (as a [Phar][13]). Rather than Hazel managing processes, [launchd][9] typically runs it as an agent daemon. Rather than only printing shipping labels, it helps with several different tasks. Here are some of them...


**Printing Postage** - the long-lived task of printing postage. The server pushes a resource URL which has the DAZzle XML data with address/contents/weights, the task gets the resource and sends it to Endicia, and then, once finished, it pushes the results back to the server where shipment costs and confirmation codes get extracted to update the order.

**Purchasing Postage** - Endicia uses an account balance when printing postage, so whenever it gets low we need to reload it. Typically this requires user intervention since they don't support automatic reloading, but this task runs through the menus and dialogs with AppleScript ([discussed here][21]) to avoid any real interruptions. Whenever the system notices the balance getting low, it automatically sends this task to a capable workstation.

**Archiving the Mailing Log** - Endicia keeps track of the postage it prints/buys/refunds in a mailing log. Over time this grows and slows things down, so Endicia provides an option to archive the log. Normally this is a manual process, but this task automates it. In addition to archiving, it also takes care of uploading the log to an encrypted S3 bucket where a server process can later go through to reconcile the transactions. A scheduler regularly sends this task to workstations running Endicia.

**Label Printing** - another task we need to manage is printing labels for inventory through [DYMO Label][15]. The labels use a QR code ([discussed here][14]) and may include price and other product information. The server pushes a resource URL which has the XML-based label template appropriate for the product, embedding the product/inventory details. The task then downloads the label file to a temporary location and uses AppleScript to open it, printing however many copies are requested.

**Webcam** - in addition to the [virtual tour][16] of the shop, we also have a [public webcam][17]. The webcam software supports sending snapshots to a URL endpoint on a timed interval, but it doesn't support SSL/TLS connections. As a workaround, this task takes care of downloading the snapshot as JPG and then uploading it securely to the correct endpoint. A scheduler is responsible for pushing this task to a server at the shop during business hours.

**Printing** - a more recent experiment is for remotely printing regular documents. Sometimes the system sends emails to the staff when they need to reprint documents (such as pricing signs, pull details, or inventory locations). Rather than waiting for someone to see those emails and manually print them, I'm hoping the documents can just be waiting in the printer in the mornings for an elf to quickly pick up and handle.

**User Dialog** - sometimes there are one-off tasks which need interaction. For example, letting the user know if Endicia is having confirmed service issues where we need to wait on printing more shipping labels.

**Automatic Updates** - another more recent development is automatic updates. Historically I used read-only deployment keys and manually deployed the full repository to workstations. This was problematic on older machines since it needed `git`. Instead, I've started deploying Phars, creating them with [box][18] and publishing a versions manifest ([example][20]) for the [php-phar-update][19] component. Whenever it's convenient for the workstation, I can push the update task and let it self-update and restart.


## From the Web

From the server side of things, it maintains a hard-coded mapping of workstations and their available tasks. Whenever multiple workstations can handle a particular task, an extra field is presented to the user so they can pick where it should happen (defaulting to their own).

<img alt="Screenshot: Print To selection" src="{{ site.asset_prefix }}/blog/2015-02-21-sending-work-from-a-web-application-to-desktop-applications/print-to-interface.jpg" width="480" />

Whenever the app needs to send a task to a bot, it queues a JSON object where the key is the task name and the its value is the task options. For example, the payload for purchasing new postage looks like:

{% highlight json %}
{ "endicia.purchase_postage": {
    "amount": 500 } }
{% endhighlight %}


## Conclusion

PHP probably isn't most people's first thought for this sort of solution - there isn't any hypertext involved, after all. But since I didn't have to abuse PHP to fit here, and since it's a language I'm very productive with, it was the most efficient route to solving my problems. It has taken a few experiments to get to this point, but over the past ~2 years this queueing/PHP-based approach has been working out very well for us on the ~6 systems it runs on.

Although it probably doesn't make much sense for others, I recently cleaned up and open sourced the bot portion of the code that I've been using for this. The [elfbot][12] repository has most of the tasks, an example configuration, and a compiled Phar in the releases. Maybe you'll find something interesting.


 [1]: https://www.usps.com/
 [2]: http://www.endicia.com/
 [3]: http://mac.endicia.com/extras/urls/
 [4]: http://mac.endicia.com/extras/applescript/
 [5]: http://www.noodlesoft.com/hazel.php
 [6]: http://mac.endicia.com/extras/xml/
 [8]: http://aws.amazon.com/sqs/
 [9]: https://developer.apple.com/library/mac/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html
 [10]: https://www.theloopyewe.com/
 [11]: https://www.theloopyewe.com/sheri/2008/08/the-loopy-elves-in-the-loopy-limelight
 [12]: https://github.com/theloopyewe/elfbot
 [13]: http://php.net/manual/en/book.phar.php
 [14]: /blog/2014/01/13/barcoding-inventory-with-qr-codes.html
 [15]: http://www.dymo.com/en-US
 [16]: https://www.theloopyewe.com/about/loopy-central/fort-collins
 [17]: https://www.theloopyewe.com/about/loopy-central/webcam/
 [18]: http://box-project.org/
 [19]: https://github.com/herrera-io/php-phar-update
 [20]: https://theloopyewe.github.io/elfbot/versions.json
 [21]: /blog/2013/01/28/scripting-endicia-to-purchase-postage.html
