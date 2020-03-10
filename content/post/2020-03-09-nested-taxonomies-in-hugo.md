---
title: Nested Taxonomies with Hugo
description: Finding ways to support content classified within hierarchies.
date: 2020-03-09
tags:
- hierarchy
- hugo
- nested
- taxonomy
---

I was looking into a [WordPress](https://wordpress.org/) to [Hugo](https://gohugo.io/) site migration recently. One problem I noticed was around Hugo's incomplete support for hierarchical taxonomies - something that WordPress offers with its notion of [categories](https://make.wordpress.org/support/user-manual/content/categories-and-tags/categories/). I spent a while trying to adapt the built-in [taxonomy conventions](https://gohugo.io/content-management/taxonomies/), but eventually I decided it wasn't practical. Instead, I switched to using regular content [sections](https://gohugo.io/content-management/sections/) and some extra templating magic to support these more complex taxonomies. The following are general strategies that I found to work for me.


## Starting with Content

The first type in this example is the taxonomy-equivalent type which I call a *collection*. To maintain the hierarchy, each collection must be its own section. For example, the following represents the *Confections* category (which itself is a subcategory of *Sweets*):

{{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="content/collection/sweets/confections/_index.md" lang="yaml" >}}

The next type is a *product*, and these are regular pages which can include a list of the collections they belong to. For example:

{{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="content/product/schoggi-schokolade.md" lang="yaml" >}}


## Listing Pages

When it comes to rendering collections, I primarily use the `section.html` template. One of the traditional views is to show all the products in the collection. For this, a straightforward [`where` function](https://gohugo.io/functions/where/) with an `intersect` can be used:

{{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="layouts/collection/section.html" lines="11-13" lang="go-html-template" >}}

The result for a collection with two products then looks something like:

> {{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="public/collection/sweets/confections/index.html" lines="56-58" lang="html" >}}


## Listing Nested Pages

A more complicated view (which I was not able to reproduce with built-in Hugo taxonomies) was to list all items in this taxonomy collection *and* sub-collections. To support this, I created a template helper function to recursively capture all the sub-collections into a [`.Scratch` variable](https://gohugo.io/functions/scratch/):

{{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="layouts/_default/baseof.html" lines="59-64" lang="go-html-template" >}}

Then, after executing it, I use it with another `where`/`intersect` lookup (and I could add [`Paginate`](https://gohugo.io/templates/pagination/) for long lists):

{{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="layouts/collection/section.html" lines="18-24" lang="go-html-template" >}}

The result for a collection containing two sub-collections with three total products then looks like:

> {{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="public/collection/sweets/index.html" lines="55-58" lang="html" >}}


## Page Navigation

For the product pages, a common view is to show breadcrumbs for the taxonomy collections it is in. To help with that, I created a template function that I can call with a collection to show the hierarchy:

{{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="layouts/_default/baseof.html" lines="53-57" lang="go-html-template" >}}

Then, on the individual product pages, I can show a set of breadcrumbs for each collection. I use the [`.GetPage`](https://gohugo.io/functions/getpage/) function to load the collection before passing it to the template function (with [`errorf`](https://gohugo.io/functions/errorf/) helping to avoid frontmatter typos):

{{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="layouts/product/single.html" lines="4-11" lang="go-html-template" >}}

The result for a product in the *Confections* and *On Sale* collections then looks like:

> {{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="public/product/schoggi-schokolade/index.html" lines="48-50" lang="html" >}}


## Still, a taxonomy

One last thing: in order to avoid confusion of the `collection/` root directory being a collection itself, I explicitly give it a type of `taxonomy`. There's nothing magic about the name `taxonomy` here – it is just another content type – but it does help simplify some of the earlier template functions and make it easier to have a different theme template at that path.

{{< snippet dir="appendix/2020-03-09-nested-taxonomies-with-hugo" file="content/collection/_index.md" lang="yaml" >}}

---

*Note: the sample site for examples from this post is [available](https://github.com/dpb587/dpb587.me/tree/master/appendix/2020-03-09-nested-taxonomies-with-hugo) if you want more details or to run `hugo serve` yourself.*
