---
date: 2013-05-13
title: Structured Data with schema.org
description: Ensuring content is useful to both humans and robots.
tags:
- product
- schema.org
- structured data
- xpath
aliases:
- /blog/2013/05/13/structured-data-with-schema-org.html
---

Good website content is important so people can learn and interact, but robots are the ones interpreting content to
figure out if the content is actually useful to people. With the new [website][1] I wanted to be sure I was using
standards and metadata so the content could be programmatically useful. I chose to use the markup from [schema.org][2]
due to its fairly comprehensive data types and broad adoption by search engines.


## Introduction

I think the importance of structured data is growing. Not only does it make things easier for search engines to
consistently interpret content, it can also help encourage properly designed website architecture. For example, if I
want search engines to know what the brand of a product is, it probably means I should ensure the product is linked to
the main brand page. A byproduct of this means a regular user can then click back to the main brand listings as well.

One of the most difficult things about embedding structured data is verifying that the markup looks how I expect. There
are tools on both [Google][5] and [Bing][8] for testing structured data, but they really work best for
publicly accessible pages (not development-local content). I found a [few][6] [other][7] [tools][9], but either they
were limited in their features or had some inconvenient bugs in how they represented data.

Ultimately, I wanted to see the website from a robots perspective and make sure I could traverse it as one. To help
myself out with that, I created a tool which would parse arbitrary local pages into JSON data based on my understanding
of how robots would interpret data. For example, I could view the [home page][1] in [raw JSON][10], or I could pretend I
was a robot and browse it in a [formatted HTML][11] page where links are rewritten for followup.


## Basic Pages

Even basic pages can provide some useful structured data. For example, the page describing the [Loopy Groupies][12]
doesn't have complicated content, but it still uses the basic [`WebPage`][14] type to identify breadcrumbs, titles, main
content, and a significant image on the page. By integrating the main site template, it also identifies the header and
footer as [`WPHeader`][16] and [`WPFooter`][16].

```json
{
    "_type": "http://schema.org/WebPage",
    "headline": "Loopy Groupies",
    "name": "Loopy Groupies",
    "breadcrumb": "Home \u00bb About \u00bb Loopy Groupies",
    "mainContentOfPage": {
        "_type": "http://schema.org/WebPageElement",
        "mainContentOfPage": "[Photo: Little Loopy with some fun] So - about those Loopy Groupies - what IS that exactly? In addition to our Loopy Rewards program, with your sixth package, you become an official member of the Loopy Groupie Club. With that, you'll receive: a fun care package to welcome you in (a cool tote with a couple of goodies inside) Loopy Kisses with each order (you'll have to wait to see those in person) advance notice of all new yarns and products right when they go up on the website (although if any of you have a particular yarn line you're watching for, of course you can email us and request to get notice of that yarn early. We're always happy to do that. The only exceptions to that are for Wollmeise, simply because it sells out too quickly for the notice to get to you in time.) an extra appreciation gift a couple of times a year, when we find something fun that we want to include in your order that month! We hope to have YOU as a Loopy Groupie, soon!",
        "primaryImageOfPage": {
            "_type": "http://schema.org/ImageObject",
            "contentUrl": "https://dy2k2bbze5kvv.cloudfront.net/static/9fdc4b5787/web/about/loopy-groupies.jpg"
        }
    }
}
```

Of course it's not limited to [`schema.org`][2] data types. The robot data also includes detailed breadcrumb data in the
[raw JSON][13] structure.


## Products

One of the most useful types in an e-commerce environment is [`SomeProducts`][3]. It lets robots see things like
pricing, inventory, availability, company, model, and various product attributes. For example, here's what our
[Slate Blue][4] product currently looks like to robots:

```json
{
    "_type": "http://schema.org/SomeProducts",
    "name": "57-61 Slate Blue",
    "model": {
        "_type": "http://schema.org/ProductModel",
        "name": "Solid Series",
        "url": "https://www.theloopyewe.com/shop/g/yarn/the-loopy-ewe/solid-series/"
    },
    "brand": {
        "_type": "http://schema.org/Brand",
        "name": "The Loopy Ewe",
        "url": "https://www.theloopyewe.com/shop/c/the-loopy-ewe/"
    },
    "offers": {
        "_type": "http://schema.org/Offer",
        "priceCurrency": "USD",
        "itemCondition": "http://schema.org/NewCondition",
        "availability": "http://schema.org/InStock",
        "price": "11.00"
    },
    "weight": {
        "_type": "http://schema.org/QuantitativeValue",
        "unitCode": "ONZ",
        "value": "2.00"
    },
    "image": "https://ehlo-a0.theloopyewe.net/asset/catalog-entry-photo/e9a1b966-2747-11d5-a74b-d0ba4caf395e~v2-702x702.jpg",
    "description": "Our Solid Series line brings you 90 solid colors in a smooshy fingering base, perfect for showing off your most intricate sock and shawl designs. You'll also love having this extensive palette of colors to choose from when working with colorwork, whether it's in socks, mitts, gloves, hats, cowls, shawls, or fine sweaters and vests. Dyed exclusively for The Loopy Ewe.",
    "inventoryLevel": {
        "_type": "http://schema.org/QuantitativeValue",
        "unitCode": "SW",
        "minValue": "15"
    },
    "_extra": [
        {
            "_type": "http://schema.org/QuantitativeValue",
            "name": "Fiber Material",
            "unitCode": "P1",
            "value": "100",
            "valueReference": {
                "_type": "http://schema.org/QualitativeValue",
                "url": "https://www.theloopyewe.com/shop/a/fiber-material/superwash-merino/",
                "name": "Superwash Merino"
            }
        },
        {
            "_type": "http://schema.org/QualitativeValue",
            "url": "https://www.theloopyewe.com/shop/a/fiber-weight/fingering-weight/",
            "name": "Fingering Weight"
        },
        {
            "_type": "http://schema.org/QuantitativeValue",
            "name": "Yardage",
            "unitCode": "YRD",
            "value": "220"
        },
        {
            "_type": "http://schema.org/ImageObject",
            "contentUrl": "https://ehlo-a0.theloopyewe.net/asset/catalog-entry-photo/e9a1b966-2747-11d5-a74b-d0ba4caf395e~v2-702x702.jpg",
            "thumbnailUrl": "https://ehlo-a1.theloopyewe.net/asset/catalog-entry-photo/e9a1b966-2747-11d5-a74b-d0ba4caf395e~v2-96x96.jpg"
        },
        {
            "_type": "http://schema.org/ImageObject",
            "contentUrl": "https://ehlo-a1.theloopyewe.net/asset/catalog-entry-photo/4a087c72-bdba-fc92-8867-3b7f1bd4fb24~v2-702x702.jpg",
            "thumbnailUrl": "https://ehlo-a1.theloopyewe.net/asset/catalog-entry-photo/4a087c72-bdba-fc92-8867-3b7f1bd4fb24~v2-96x96.jpg"
        }
    ]
}
```

With the markup on the page, it's now possible for search engines to quickly show information such as pricing and
availability alongside results for the product. Not only that, but given sufficient parsing it can also infer the
relationships that specific page (marked as a product) has with other product concepts to create a more intelligent data
graph.


## Product Listings

For the main product types, pages also support listings that reference the individual products. The main
[Solid Series][17] listing has the following data:

```json
{
    "_type": "http://schema.org/CollectionPage",
    "mainContentOfPage": {
        "_type": "http://schema.org/WebPageElement",
        "_extra": [
            {
                "_type": "http://schema.org/ItemList",
                "_extra": [
                    {
                        "_type": "http://schema.org/SomeProducts",
                        "url": "https://www.theloopyewe.com/shop/p/BEA04EDF-Solid-Series-00-Color-Cards",
                        "image": "https://ehlo-a0.theloopyewe.net/asset/catalog-entry-photo/1341fe35-3260-4ece-df42-387e9ddcafe5~v2-210x130.jpg",
                        "name": "00 Color Cards",
                        "itemCondition": "http://schema.org/NewCondition",
                        "offers": {
                            "_type": "http://schema.org/Offer",
                            "priceCurrency": "USD",
                            "availability": "http://schema.org/InStock",
                            "price": "15.00"
                        },
                        "inventoryLevel": {
                            "_type": "http://schema.org/QuantitativeValue",
                            "unitCode": "SW",
                            "minValue": "15"
                        }
                    },
                    {
                        "_type": "http://schema.org/SomeProducts",
                        "url": "https://www.theloopyewe.com/shop/p/0300A54D-Solid-Series-01-39-White",
                        "image": "https://ehlo-a0.theloopyewe.net/asset/catalog-entry-photo/cb0e1bb9-1431-ef41-7232-c79a3c510f2a~v2-210x130.jpg",
                        "name": "01-39 White",
                        "itemCondition": "http://schema.org/NewCondition",
                        "offers": {
                            "_type": "http://schema.org/Offer",
                            "priceCurrency": "USD",
                            "availability": "http://schema.org/InStock",
                            "price": "11.00"
                        },
                        "inventoryLevel": {
                            "_type": "http://schema.org/QuantitativeValue",
                            "unitCode": "SW",
                            "minValue": "15"
                        }
                    }
                ]
            }
        ]
    }
}
```


## Rationale

Nearly all pages on the new [website][1] have at least some structured data present, if only the breadcrumb data. All
this markup isn't simply an academic exercise though. For example, [Ravelry][18] supports checking the pricing and
inventory of our product ads and displaying them to users. Instead of complex, fragile regular expressions or DOM
traversal, we can just say to use an XPath query like `*[@itemscope and @itemtype = "http://schema.org/SomeProducts"]`.

One of the original motivations behind focusing on structured data was the goal of having an internal search for the
site. Instead of writing web page scrapers that know what the DOM looks like and how to find significant content, it has
been much easier to rely on simple `schema.org` types which are consistent across all pages. The structured data on
pages is still a work in progress as I learn more about what robots are interested in and figure out the best way to
represent content.


 [1]: https://theloopyewe.com/
 [2]: http://schema.org/
 [3]: http://schema.org/SomeProducts
 [4]: https://theloopyewe.com/shop/p/C7CBF721-Solid-Series-57-61-Slate-Blue
 [5]: http://www.google.com/webmasters/tools/richsnippets
 [6]: http://linter.structured-data.org/
 [7]: http://foolip.org/microdatajs/live/
 [8]: http://www.bing.com/toolbox/markup-validator
 [9]: https://github.com/linclark/MicrodataPHP
 [10]: https://www.theloopyewe.com/api/search/resource/uri.json?uri=%2F&pretty
 [11]: https://www.theloopyewe.com/api/search/resource/uri.html?uri=%2F
 [12]: https://www.theloopyewe.com/about/loopy-groupies.html
 [13]: https://www.theloopyewe.com/api/search/resource/uri.json?uri=%2Fabout%2Floopy-groupies.html&pretty
 [14]: http://schema.org/WebPage
 [15]: http://schema.org/WPHeader
 [16]: http://schema.org/WPFooter
 [17]: https://www.theloopyewe.com/shop/g/yarn/the-loopy-ewe/solid-series/
 [18]: http://www.ravelry.com/
