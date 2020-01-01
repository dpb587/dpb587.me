---
date: 2013-03-23
title: Bank Card Readers for Web Applications
description: Scanning credit cards into website forms.
code: https://gist.github.com/dpb587/5229239
tags:
- bank card
- forms
- javascript
- reader
aliases:
- /blog/2013/03/23/bank-card-readers-for-web-applications.html
---

I made a web-based [point of sale][1] for [The Loopy Ewe][2], but it needed an easier way to accept credit cards aside
from manually typing in the credit card details. To help with that, we got a keyboard-emulating USB magnetic card reader
and I wrote a [parser][3] for the [card data][4] and convert it to an object. It is fairly simple to hook up to a form
and enable a card to be scanned while the user is focused in the name or number fields...

```javascript
require(
    [ 'payment/form/cardparser', 'vendor/mootools' ],
    function (paymentFormCardparser) {
        function storeCard(card) {
            $('payment[card][name]').value = card.name;
            $('payment[card][number]').value = card.number;
            $('payment[card][expm]').value = card.expm;
            $('payment[card][expy]').value = card.expy;
            $('payment[card][code]').focus();
        }

        paymentFormCardparser
            .listen($('payment[card][name]'), storeCard)
            .listen($('payment[card][number]'), storeCard)
        ;
    }
);
```

It acts as a very passive listener without requiring the user to do anything special - if there is no card reader
connected then the form field is simply a regular field for keyboard input.


 [1]: http://en.wikipedia.org/wiki/Point_of_sale
 [2]: http://www.theloopyewe.com/
 [3]: https://gist.github.com/dpb587/5229239#file-cardparser-js
 [4]: http://en.wikipedia.org/wiki/Magnetic_stripe_card
