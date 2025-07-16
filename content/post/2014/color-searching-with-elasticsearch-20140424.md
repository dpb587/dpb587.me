---
description: Some mappings, strategies, and queries for advanced color searching with elasticsearch.
params:
    nav:
        tag:
            color: true
            ecommerce: true
            elasticsearch: true
            hsv: true
            search: true
            weighted: true
publishDate: "2014-04-24"
title: Search by Color with Elasticsearch
---

A [year ago][1] when I updated the [TLE website][2] I dropped the "search by color" functionality. Originally, all the
colors were indexed into a database table and the frontend generated some complex queries to support specific and
multi-color searching. On occasion, it caused some database bottlenecks during peak loads and with some particularly
complex color combinations. The color search was also a completely separate interface from searching other product
attributes and availability. It was neat, but it was not a great user experience.

It took some time to get back to the search by color functionality, but I've finally been able to get back to it and,
with [elasticsearch][3], significantly improve it.

{{< image alt="Screenshot: colorized yarn" src="https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2014-04-24-color-searching-with-elasticsearch/search0.png" href="http://www.theloopyewe.com/shop/search/cd/0-100~75-90-50~18-12-12/g/59A9BAC5/" >}}


# Color Quantification {#color-quantification}

One of the most difficult processes of supporting color searches is to figure out the colors in products. In our case,
where we had thousands of items to "colorize", it would be easier to create an algorithm than have somebody manually
pick out significant colors. When it comes to algorithms and research, the process is called [color quantization][8].
A lot of the inventory at the shop is yarn and, unfortunately, the tools I tried didn't do a good job at picking out the
fiber colors (they would find significance in the numerous shadows or average colors).

Ultimately I ended up creating my own algorithm based on several strategies. In addition to finding the significant
colors it also keeps track of their ratios making it easy to realize multi-color items vs items with accent colors.
After batch processing inventory to bring colors up to date, I added hooks to ensure new images are processed for colors
as they're uploaded.

{{< image alt="Screenshot: colorized yarn" src="https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2014-04-24-color-searching-with-elasticsearch/colorizer-yarn.png" href="https://www.theloopyewe.com/shop/p/78C97118-Gobelin-A-moi-le-coco" >}}

{{< image alt="Screenshot: colorized fabric" src="https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2014-04-24-color-searching-with-elasticsearch/colorizer-fabric.png" href="https://www.theloopyewe.com/shop/p/86330BB1-DS23-Seafaring" >}}

You can see it noticed the significant colors of the yarn and fabric above, along with their approximate ratios. With
some types of items, it may be possible to infer additional meaning such as the "background color" of fabric.


# Color Theory {#color-theory}

When it comes to color, there are a few standard methods for measuring it. Probably the most familiar one from a web
perspective is [RGB][6]. Unfortunately, RGB doesn't efficiently quantify the "color" or hue. For example,
rgb(244, 40, 5) and
rgb(244, 214, 214)
are both obviously reddish, but the second color has high green and blue values yet the blue and green colors are not
present.

A much better model for this is [HSV][7] (or HSL). The "color" (hue, `H`) cycles from 0 thru 360 where 0 and 360 are
both red. The `S` for "saturation" ranges from 0 to 100 and describes how much "color" there is. Finally, the `V` for
"value" (or `B` for "brightness") ranges from 0 to 100 and describes how bright or dark it is. Compare the following
examples for a better idea:

 * rgb(0, 70, 70)
 * rgb(0, 30, 70)
 * rgb(0, 70, 30)
 * rgb(0, 30, 30)
 * rgb(180, 70, 70)
 * rgb(180, 30, 70)
 * rgb(180, 70, 30)
 * rgb(180, 30, 30)

Within elasticsearch we can easily map an object with the three color properties as integers:

```json
{ "color" : {
    "properties" : {
      "h" : { "type" : "integer" },
      "s" : { "type" : "integer" },
      "v" : { "type" : "integer" } } } }
```


# Mappings {#mappings}

Elasticsearch will natively handle [arrays][5] of multiple colors, but `color` needs to become a [`nested`][4] mapping
type in order to support realistic searches. For example, we could write a query looking for a dark blue, but unless
it's a nested object the query could match items which have any sort of blue (`color.h = 240`) and any sort of dark
(`color.v < 50`). To make `color` nested, we just have to add `type = nested`. Then we're able to write `nested` filters
which will look like:

```json
{ "nested" : {
    "path" : "color",
    "filter" : {
      "bool" : {
        "must" : [
          { "term" : { "color.h" : 240 } },
          { "range" : {
              "color.v" : { "lt" : 50 } } } ] } } } }
```

With the extra color proportion value mentioned earlier, we're also able to add a `ratio` range alongside `h`, `s`, and
`v`. This will allow us to find items where blue is more of a dominant color (e.g. more than 80%). Another searchable
fact which may be useful is `color_count` - then we would be able to find all solid-color products, or all dual-color
products, or just any products with more than four significant colors.

While working on a frontend interface, I was having trouble faceting popular colors. A lot of dull colors were coming
back. As a first step, I started using some [`terms`][9] aggregations with a `value_script` which created large buckets
of colors from the `h`, `s`, and `v` tuple. That helped significantly, but then it seemed like there was a
disproportionate number of very dark and very light colors. Instead of adding additional calculations to the aggregation
during runtime, I decided to pre-compute the buckets that the colors should belong to. Now it's doing more advanced
calculations and no runtime calculations. For example, all low-`v` colors will end up in a single bucket
`{ h : 360 , s : 10 , v : 10 , ... }`. Similar rules trim low-saturation colors and create the appropriate buckets for
colors.


# Searches {#searches}

Given four key properties (hue, saturation, value, and color ratio), I needed a way to represent the searches from
users. For searching individual colors, I settled on the following syntax:

```
{ratio-min}-{ratio-max}~{hue}-{sat}-{val}~{hue-range}-{sat-range}-{val-range}
```

This way, if a user is very specific about the dark blue they want, and they want at least 80% of the item to be blue,
the color slug might look like: [`80-100~190-100-50~10-5-5`][10]. Within the application, this gets translated into a
[`filtered`][11] query. The filter part looks like:

```json
{ "filter": {
    "and": [
      { "nested": {
          "path": "color",
          "filter": {
            "and": [
              { "range": {
                  "ratio": {
                    "gte": 80,
                    "lte": 100 } } },
              { "range": {
                  "h": {
                    "gte": 180,
                    "lte": 200 } } },
              { "range": {
                  "s": {
                    "gte": 95,
                    "lte": 100 } } },
              { "range": {
                  "v": {
                    "gte": 45,
                    "lte": 55 } } } ] } } } ] } }
```

The query part then becomes responsible for ranking using a basic calculation which roughly computes the distance
between the requested color and the matched color. The [`function_score`][13] query currently looks like:

```json
{ "function_score": {
    "boost_mode": "replace",
    "query": {
      "nested": {
        "path": "color",
        "query": {
          "function_score": {
            "score_mode": "multiply",
            "functions": [
              { "exp": {
                  "h": {
                    "origin": 190,
                    "offset": 2,
                    "scale": 4 } } },
              { "exp": {
                  "s": {
                    "origin": 100,
                    "offset": 4,
                    "scale": 8 } } },
              { "exp": {
                  "v": {
                    "origin": 50,
                    "offset": 4,
                    "scale": 8 } } },
              { "linear": {
                  "ratio": {
                    "origin": 100,
                    "offset": 5,
                    "scale": 10 } } } ] } },
        "score_mode": "sum" } },
    "functions": [
        { "script_score": { "script": "_score" } } ] } }
```

The `_score` can then be used in sorting to show the closest color matches first.

{{< image alt="Screenshot: search screen shot" src="https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2014-04-24-color-searching-with-elasticsearch/search1.png" href="http://www.theloopyewe.com/shop/search/cd/80-100~190-100-50~10-5-5/g/59A9BAC5/" >}}

Of course, these color searches can be added alongside the other facet searches like product availability, attributes,
and regular keyword searches.


# User Interface {#user-interface}

One of the more difficult tasks of the color search was to create a reasonable user interface to front the powerful
capabilities. This initial version uses the same interface as a year ago, letting users pick from the available "color
dots". Ultimately I hope to improve it with a more advanced, yet simple, [Raphaël][12] interface which would let them
pick a specific color and say how picky they want to be. That goal requires a fair bit of time and learning though...


# Summary {#summary}

I'm excited to have the search by color functionality back. I'm even more excited about the possibilities of better,
advanced user searches further down the road. After it gets used a bit more, I hope we can more prominently promote the
color search functionality around the site. Elasticsearch has been an excellent tool for our product searching and it's
exciting to continue expanding the role it takes in powering the website.


 [1]: /blog/2013/04/27/new-website-for-the-loopy-ewe.html
 [2]: http://www.theloopyewe.com/
 [3]: http://www.elasticsearch.org/
 [4]: http://www.elasticsearch.org/guide/en/elasticsearch/reference/1.x/mapping-nested-type.html
 [5]: http://www.elasticsearch.org/guide/en/elasticsearch/reference/1.x/mapping-array-type.html
 [6]: http://en.wikipedia.org/wiki/RGB_color_model
 [7]: http://en.wikipedia.org/wiki/HSL_and_HSV
 [8]: http://en.wikipedia.org/wiki/Color_quantization
 [9]: http://www.elasticsearch.org/guide/en/elasticsearch/reference/1.x/search-aggregations-bucket-terms-aggregation.html
 [10]: http://www.theloopyewe.com/shop/search/cd/80-100~190-100-50~10-5-5/g/59A9BAC5/
 [11]: http://www.elasticsearch.org/guide/en/elasticsearch/reference/1.x/query-dsl-filtered-query.html#query-dsl-filtered-query
 [12]: http://raphaeljs.com/
 [13]: http://www.elasticsearch.org/guide/en/elasticsearch/reference/1.x/query-dsl-function-score-query.html
