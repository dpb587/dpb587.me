---
date: 2013-04-27
title: New Website for The Loopy Ewe
description: A summary of the customer-facing changes I worked on for the site.
tags:
- elasticsearch
- migration
- redesign
- theloopyewe
aliases:
- /blog/2013/04/27/new-website-for-the-loopy-ewe.html
---

I've spent the past several months working on some website changes for [The Loopy Ewe][1]. On Thursday I was able to
push many of those frontend changes out. I thought I'd briefly discuss some of those changes here.


## Before and After

First off, it's fun to show before and after screenshots of many key areas...


### Home Page

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-homepage.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-homepage.jpg)](http://theloopyewe.com/) |
| --- | --- |

So the home page is one of the first welcome pages to new visitors. I wanted to make sure it was warm and welcoming,
primarily through the central photos we show; the default one being the entry view of our shop (with a dynamic thumbnail
of our webcam in the bottom right). Over time we'll be able to rotate through different photos for different events,
product updates, and more clever things.

I wanted to get rid of the multi-color sidebar from every page so it could be better filled with more useful,
page-specific content. Visually, I increased the page width from 784px to 960px, so combined with dropping the sidebar
it allows for about 75% more content area.

Previously the sidebar was the main method of navigation, so I regrouped the old blue navigation link box into about 6
different topics to use as the main header links.

Instead of a simple, almost-non-existant footer on the old site, I took advantage of that area to include store
information, social links, payment options, and numerous other credentials that customers can appreciate.


### Contact Us

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-contactus.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-contactus.jpg)](http://www.theloopyewe.com/contact/) |
| --- | --- |

Contact information is important for customers. In addition to the information now being in the footer, there is a
cleaner page with a new interactive map to help people visually realize where exactly the shop is located.


### Wonderful Customers

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-testimonials.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-testimonials.jpg)](http://www.theloopyewe.com/about/wonderful-customers/) |
| --- | --- |

It's always nice to be able to show feedback customers send in. The new site reorganizes everything in a nicer, more
readable way, and on separate pages. It's also much simpler to submit a testimonial through the on-screen form.


### Shop

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-shop.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-shop.jpg)](http://www.theloopyewe.com/shop/) |
| --- | --- |

Generally speaking, I wanted the photos to be the main defining experience that a visitor has. To that end, product
photos became significantly larger in an effort to fill in the missing colors of the simple color palette I used.
Since it's the main shop page, I also included useful links like new products, gift certificates, search, and links for
browsing by some attributes.

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-shop-category.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-shop-category.jpg)](http://www.theloopyewe.com/shop/g/yarn/cascade/) |
| --- | --- |

Within specific shop categories, I only slightly increased the thumbnails and instead favored focusing more on the
different brands and their distinctions.

One other significant addition to the new website is the social sharing functionality. On most shop pages, there are new
social sharing links to Twitter, Pinterest, and Facebook. Using a custom short domain and campaign URL arguments, we can
get better insight into customer interests.

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-shop-brand.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-shop-brand.jpg)](http://www.theloopyewe.com/shop/g/yarn/cascade/220/) |
| --- | --- |

In my opinion, one of the best changes has been to viewing products on pages like this. Using a sidebar to show the
description and attributes allows customers to more quickly see the enticing and larger product photos together.

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-shop-product.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-shop-product.jpg)](http://www.theloopyewe.com/shop/p/F2FDB8A1-220-8905-Robin-Egg-Blue) |
| --- | --- |

I think the second best improvement is the individual product page where the photo takes precedence and shows off the
quality of the product. A larger call-to-action makes it easier to add the item to carts and wishlists. I reorganized
the product information as well to better prioritize it, visually.

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-shop-search-grid.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-shop-search-grid.jpg)](http://www.theloopyewe.com/shop/search/a/fiber-weight/fingering-weight/availability/in-stock/?q=red) |
| --- | --- |

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-shop-search-list.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-shop-search-list.jpg)](http://www.theloopyewe.com/shop/search/a/fiber-weight/fingering-weight/availability/in-stock/?q=red&amp;r%5Bview%5D=list-tn) |
| --- | --- |

One major feature addition has been a real search engine. The old site used some complex and inefficient database
queries (which actually caused noticeable performance issues at rare times). With the new site, all the products are
properly indexed and searched via [elasticsearch][2]. I'm looking forward to adding more elasticsearch integrations on
the site in the future.


### Help

| ![Screenshot: before](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/old-help.jpg) | [![Screenshot: after](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-help.jpg)](http://www.theloopyewe.com/help/) |
| --- | --- |

Previously we had a single, text-heavy and difficult to read help page, also known as "frequently asked questions." The
new site breaks things down into different topics and adds creative pictures to make things more readable. There's also
a new inline form where customers can ask for help instead of bothering to open an email client and compose an email.


## New Stuff

Although I disabled a number of things for later release and chatter, it's always fun to include some completely new
functionality...


### Local

[![Screenshot: web page](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-local.jpg)](http://www.theloopyewe.com/local/classes.html)

I created a new topic dedicated to our local customers. Since it's not only an online store anymore, we wanted a way to
publicize some of the local activities that Fort Collins people would be interested in. It also lets online-only
customers see how we exist and work in real life to create more of a connection.


### About

[![Screenshot: web page](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-about.jpg)](http://www.theloopyewe.com/about/loopy-central/fort-collins.html)

Along with a local page, I also wanted a better page for showing our real world existence so customers could feel more
connected and understand both who and where they're purchasing from.


### Shop Attributes

[![Screenshot: web page](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-shop-attribute.jpg)](http://www.theloopyewe.com/shop/a/fiber-material/merino-wool/)

In an effort to make navigating the shop easier, I created new pages to view products by attributes in a more organized
way. If somebody is interested in "Fingering Weight" they can easily see all the companies and brands that offer it. If
they need more complicated searches, there's an Advanced Search link at the bottom of each page.


### Site Feedback

[![Screenshot: web page](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-sitefeedback.jpg)](http://www.theloopyewe.com/contact/site-feedback.html?uri=%2Fshop%2Fg%2Fyarn%2Fthe-loopy-ewe%2Floopy-cakes%2F)

For both the cases of bugs and hearing ideas for improvement, I wanted to be sure visitors could easily send technical
feedback. Links at the footer of every page include information like what page they were looking at, what browser,
authenticated username information, and whatever notes they want to add.


### humans.txt

[![Screenshot: web page](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2013-04-27-new-website-for-the-loopy-ewe/new-humans.jpg)](http://www.theloopyewe.com/humans.txt)

Whenever possible, I like discussing and linking to technical resources that I have found useful. For the nerdy types, I
created the `humans.txt` file to document many of the resources that have helped make the website possible.


## Conclusion

So there's the basic overview about some of the less-technical changes. I'm looking forward to several additional
features to rollout over time and help keep things fresh over the next few months. Later blog posts can discuss some of
the more technical processes and decisions that have helped in making the new site.


 [1]: http://www.theloopyewe.com/
 [2]: http://www.elasticsearch.org/
