---
title: "Barcoding Inventory with QR Codes"
layout: post
tags: [ 'barcode', 'qr', 'retail', 'product', 'label', 'scan' ]
description: A web-centric, user-friendly approach for using barcodes in a retail shop.
---

Most decently-sized stores will have barcodes on their products. For the store, it makes the checkout process extremely
easy and accurate. For the consumer, barcodes might be useful with a phone app to scan them. I needed to make the
inventory scannable at the [shop][1], and I really wanted to do it in a more meaningful way than 1D barcodes could
support.


## Barcodes: 1D vs 2D

There are two different kinds of barcodes: 1 dimensional and 2 dimensional. The 1D allows for a purely linear scan of
simple, [UPC][2]-like barcodes. While 1D barcodes are extremely commonplace on many products, I dislike them because
they can't provide any context.

For example, if I were shopping in [Target][3] and scanned a UPC barcode with a regular phone app, it might take me to
the [Amazon][4] listing first - not necessarily great for Target's business, but it also becomes a completely separate
brand channel distracting my thoughts. Another example is when UPCs aren't registered on a product - different retail
stores will make up their own internal barcode which isn't helpful at all if I try to scan it.

On the otherhand, 2D barcodes require complex parsing but they can hold much more data. [QR codes][5] are one extremely
common form of 2D barcodes and they typically encode URLs. With my goal of providing more context, URLs provide just
that - not only with a domain name, but an arbitrary path. If somebody scanned an item at our shop, they'd at least get
redirected through the shop's website.

One disadvantage that QR codes have compared to 1D barcodes is their size and resolution requirements. All 1D barcodes
could theoretically be 1 pixel high, but QR codes must be square. To help ensure a reasonable QR codes, most people
will use a URL shortener service - shorter URLs mean simpler QR designs, simpler designs mean the QR code can be read
more easily and doesn't need to be large.

Another disadvantage to QR codes is that 2D handheld scanners are significantly more expensive than 1D. Fortunately,
many previously-used 2D scanners can be found on [eBay][6] for very reasonable prices. Unfortunately, I found that
some of the used ones would quickly turn unreliable after a period of time.


## Mapping URLs to retail "things"

While inventory was the primary target of barcoding, I really wanted to barcode most things involved with retail
workflows (like order receipts). With that in mind I figured I needed to store three properties:

 * `insignia` - the unique, short identifier (e.g. `EyV3chYax`)
 * `target_ref` - the type of "thing" (e.g. `inventory` or `order`)
 * `target_id` - the ID of the "thing" (e.g. `010035EA-9F6D-41A2-97C4-EEB5A3F3034A`)

I created a manager which supports three basic operations (internally it uses a map of the different types of
"things"):

 * `getInsignia($target)` - which returns the short identifier/insignia
 * `getTarget($insignia)` - which returns the application object
 * `getResponse($insignia)` - which returns an appropriate HTTP response

I created a couple of HTTP endpoints which utilize the manager:

 * `/io/{insignia}` - which returns the result of `getResponse` (typically a redirect)
 * `/io/{insignia}.png` - which returns the QR code image

Then, whenever I want to print a QR code on a document, I just have to do:

    <img src="{% raw %}{{ web_insignia_png(transaction) }}{% endraw %}" style="float:right;" />
    Your Receipt for Order #{% raw %}{{ transaction.id }}{% endraw %}

Further, the QR code can be used with a redirecting short domain for even simpler codes:

 > ![QR Code](http://www.theloopyewe.com/io/EyV3chYax.png?s=2 "http://tle.io/EyV3chYax")  
 > [`http://tle.io/EyV3chYax`](http://www.theloopyewe.com/io/EyV3chYax)


## Adding More Context

One of the reasons I wanted to use QR codes was context. Aside from scans now landing on the shop's website, they can
be even more context-aware through security roles. For example, if a customer scans the QR code above, they'll end up
on the product page for the shop as you would expect; but if an admin scans it they'll end up on the main inventory
page to see current quantities and recent transactions.

Or, a better example is with order receipts. If somebody scans their order receipt, they'll be required to login and
then will be taken to their order details page, assuming the order was on their account. If an admin scans the receipt
(perhaps while packing it) they'll be taken to the administrative, detailed view of the order.

The idea of context doesn't only apply to where a user might end up, it also applies to how they get there. For
example, sometimes there will be more than one bolt of a single fabric pattern, and, since each bolt is a different
"thing" in the system, they each have a different QR code. If a customer scans either of the bolts, they would get
taken to the exact same public product page. However, when an admin scans a bolt they'll get taken to the detailed
view showing which orders were cut on that specific bolt and how much yardage the bolt still has.



## Integrated Context

At this point, the barcodes were extremely accessible for one-off scans, but I also wanted to integrate the barcodes
into specific points of the system. For the computers we're using USB 2D barcode scanners which are capable of acting
like a keyboard device (the computer sees it "typing" whatever it scans, followed by an Enter). The most useful
integration point was the POS for handling in-store shoppers.

For the POS, I created a new UI component which auto-focused itself. Once something gets scanned, it sends the scanned
data to the server so it can figure out what should happen. For QR code scans, it performs the insignia lookup to find
the actual inventory item. Then, for simple inventory items it can just add the scanned item to the order. For fabric
on the bolt, it comes back with a dialog about how much to cut. For complex items, it shows a dialog for further
specifications. Or there might just be a discrepancy and it needs to come back and show a message. Once the item is
added it provides visual feedback, the scan field is re-focused and the cycle continues. It works something like the
following in a browser...

<blockquote>
    <div style="color:#666666;padding-left:5px;">
        <span id="demoscan-dotty" style="background-color:#CC0000;border:#999999 solid 1px;border-radius:3px;display:inline-block;margin-bottom:-1px;width:12px;height:12px;"></span>
        <a id="demoscan-talkr" class="subtle" href="#" style="display:inline-block;padding:3px 2px;">Click to Scan</a>
        <input id="demoscan-input" type="text" style="border:transparent;background-color:transparent;height:1px;margin:0;padding:0;width:1px;" />
    </div>
    <script src="//ajax.googleapis.com/ajax/libs/mootools/1.4.5/mootools-yui-compressed.js"></script>
    <script type="text/javascript">
        var talkr = $('demoscan-talkr');
        var input = $('demoscan-input');
        var dotty = $('demoscan-dotty');

        dotty.set('tween', { link : 'cancel', duration : 200 });

        input
            .addEvent(
                'focus',
                function () {
                    talkr.set('text', 'Ready to Scan');
                    dotty.tween('background-color', '#66FF66');
                }
            )
            .addEvent(
                'blur',
                function () {
                    talkr.set('text', 'Click to Scan');
                    dotty.tween('background-color', '#CC0000');
                }
            ).addEvent(
                'keydown',
                function (e) {
                    if ('enter' != e.key) {
                        return;
                    } else if (!this.value) {
                        return;
                    }

                    prompt('Seems like you scanned...', this.value);

                    this.value = '';
                    this.focus();
                }
            )
        ;

        talkr
            .addEvent(
                'click',
                function (e) {
                    input.focus();

                    e.preventDefault();
                }
            )
        ;
    </script>
</blockquote>


## Conclusion

I feel like the shop is able to better grow both technically and logistically by having used QR codes as opposed to a
classic barcode system. A few techy customers have tried the QR codes, but it's not really something we've been
promoting. Once the website has a proper mobile-friendly version we'll have a better opportunity and reason to try and
impress customers with the QR codes. In the meantime, the QR codes have been an immense time-saver for both staff and
shoppers checking out at the shop.


 [1]: http://www.theloopyewe.com/
 [2]: http://en.wikipedia.org/wiki/Universal_Product_Code
 [3]: http://www.target.com/
 [4]: http://www.amazon.com/
 [5]: http://en.wikipedia.org/wiki/QR_code
 [6]: http://www.ebay.com/
