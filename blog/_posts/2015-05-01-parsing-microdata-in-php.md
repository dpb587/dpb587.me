---
title: "Parsing Microdata in PHP"
layout: "post"
tags: [ "microdata", "opensource", "php", "schema", "xpath" ]
description: "Open sourcing a library to easily traverse HTML for microdata."
code: https://github.com/dpb587/microdata-dom.php
---

A couple years ago I wrote about how I was [adding microdata][3] to [The Loopy Ewe][1] website to annotate things like products, brands, and contact details. I later wrote about how the internal search engine [depended on that microdata][4] for search results. During development and the initial release I was using some basic [XPath][2] queries, but as time passed the implementation became more fragile and incomplete. Since then, the parser has gone through several refactorings and this week I was able to extract it into a separate library that I can [open source][9].


## Implementation

My original implementation was a single helper class with a confusing mix of recursion, loops, and values by reference. The helper would receive the HTML string to parse and it would return a complex array with self-referencing values for multi-level scopes. Looking for a more reliable data structure to pass around, I decided to switch and extend the [`DOMDocument`][5]. I spent some time reading the [HTML Microdata][6] spec and wanted to try and find a balance between the spec's [DOM API][7] and existing PHP conventions.

Now I use the library's [`MicrodataDOM\DOMDocument`][8] class when I want to parse a microdata document. It works just like the built-in `DOMDocument` so I'm able to manage libxml errors, control how I import the HTML document, and pass it through methods which are expecting a regular `DOMDocument`. The key difference is the addition of a `getItems` method which lets me quickly retrieve the microdata items. Internally, `getItems` and subsequent calls are still using XPath queries.

In addition to extending `DOMDocument`, the library also extends `DOMElement`. This way, `getItems` is just returning a regular (but still specialized) list of DOM elements. The extended element class provides access to the microdata attributes like type, property name, and value.


## Usage

It's works like a low-level library, expecting other, more specialized classes to add their own friendlier methods on top. Here's the example I used in the readme...

{% highlight javascript %}
<?php

$dom = new MicrodataDOM\DOMDocument();
$dom->loadHTMLFile('http://dpb587.me/about.html');

// find Person types and get the first item
$dpb587 = $dom->getItems('http://schema.org/Person')->item(0);
echo $dpb587->itemId;

// items are still regular DOMElement objects
printf(" (from %s on line %s)\n", $dpb587->getNodePath(), $dpb587->getLineNo());

// there are a couple ways to access the first value of a named property
printf("givenName: %s\n", $dpb587->properties['givenName'][0]->itemValue);
printf("familyName: %s\n", $dpb587->properties['familyName']->getValues()[0]);

// or directly get the third, property-defining DOM element
$property = $dpb587->properties[3];
printf("%s: %s\n", $property->itemProp[0], $property->itemValue);

// use the toArray method to get a Microdata JSON structure
echo json_encode($dpb587->toArray(), JSON_UNESCAPED_SLASHES) . "\n";
{% endhighlight %}

Which will output something like...

    http://dpb587.me/ (from /html/body/article/section on line 97)
    givenName: Danny
    familyName: Berger
    jobTitle: Software Engineer
    {"id":"http://dpb587.me/","type":["http://schema.org/Person"],"properties":{"givenName":["Danny"],...snip...}

In addition to using it for the internal search, I've been using this library for other internal tools responsible for sanitizing, normalizing, and taking care of some validation during development and testing. Hopefully I'll be able to extract and open-source those features sometime as well.


## Summary

Back when I first started this, I couldn't find any good libraries for this sort of microdata parsing. Nowadays it looks like there's at least [one other project][10] which I would consider if I didn't already have an implementation. With bias, I do still favor mine because of the unit tests, `itemprop` properties implementation, and a bit closer mirroring of how the spec describes interacting with a microdata API.


 [1]: https://www.theloopyewe.com/
 [2]: http://php.net/manual/en/class.domxpath.php
 [3]: /blog/2013/05/13/structured-data-with-schema-org.html
 [4]: /blog/2013/06/01/search-engine-based-on-structured-data.html
 [5]: http://php.net/manual/en/class.domdocument.php
 [6]: http://www.w3.org/TR/microdata/
 [7]: http://www.w3.org/TR/microdata/#microdata-dom-api
 [8]: https://github.com/dpb587/microdata-dom.php/blob/master/src/MicrodataDOM/DOMDocument.php
 [9]: https://github.com/dpb587/microdata-dom.php
 [10]: https://github.com/linclark/MicrodataPHP
