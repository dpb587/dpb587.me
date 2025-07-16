---
description: Open sourcing a library to consume the knitting-oriented API.
params:
    nav:
        tag:
            api: true
            php: true
            ravelry: true
publishDate: "2016-01-21"
title: Writing a PHP Client for the Ravelry API
---

[Ravelry][1] is, in [their own words][2], "a place for knitters, crocheters, designers, spinners, weavers and dyers to keep track of their yarn, tools, project and pattern information, and look to others for ideas and inspiration." It's no wonder so many of [TLE][3]'s customers are also "Ravelers". Several years ago Ravelry created an API so developers could write apps and create integrations that users would love. Classifying myself as more a developer than a knitter, the API piqued my interest. Fast-forward a bit and I've created and now open-sourced a [couple][4] [projects][5] for the Ravelry API.


# With PHP Scripts {#with-php-scripts}

The library is registered on [packagist][6], so it's easy to install with a package manager like [composer][7]. Internally, the library is built on top of [Guzzle][8], an HTTP abstraction layer used by many other PHP projects. To help the object-oriented usage more closely mirror the Ravelry API hierarchy, I added a few helper classes on top. For example...

```php
// for https://api.ravelry.com/yarns/search.json?query=cascade%20220&sort=rating
$api->yarns->search([ 'query' => 'cascade 220', 'sort' => 'rating' ]);

// requests https://api.ravelry.com/messages/list.json?folder=inbox
$api->messages->list([ 'folder' => 'inbox' ]);
```

This was actually more time-consuming to implement than it may look. Some APIs provide schemas to download their capabilities, but right now the only way to grok the Ravelry API is to read the long [documentation page][9] with lots of headers and tables and links. When I first started, I manually typed out the specs for the methods I wanted to use. After doing that for three methods I wasn't thrilled about doing more (and then having to make corrections as the API changed). So, I spent a couple hours hacking together some [XQueries][11], [regexes][16], and a few manual corrections into a [mess of code][10] which dumps out a meaningful JSON schema describing the API. In addition to listing the parameter names, it includes types, documentation, and validations for both the request and response payloads...

```json
{ "yarns_search": {
    "description": "Search yarn database",
    "documentationUrl": "http://www.ravelry.com/api#yarns_search",
    "httpMethod": "GET",
    "parameters": {
      "page": {
          "description": "Result page to retrieve. Defaults to first page.",
          "location": "query",
          "required": false,
          "type": "number" },
      "query": {
          "description": "Search term for fulltext searching yarns",
          "location": "query",
          "required": false,
          "type": "string" },
      "sort": {
          "description": "Sort order.",
          "enum": [ "best", "rating", "projects" ],
          "location": "query",
          "required": false,
          "type": "string" } } }
```

I use Guzzle's [Services][12] library which takes the schema and then handles the smarts of serializing everything to and from the Ravelry API servers. Once the `$api` call is made, an array-like object is returned and can be used...

```php
echo $searchResults['yarns'][0]['yarn_company_name'];
```

To provide some assurances that the API schema was being generated properly and calls are being executed as expected, I wrote some [functional tests][13] which use [PHPUnit][14] to execute calls against the API with a test Ravelry account.


# With CLI Scripts {#with-cli-scripts}

Having the API available in my PHP scripts was a great step. Eventually, though, I found myself wanting to quickly test requests to the API, but didn't want to bother setting up a script with the class names, authentication details, and having to remember how parameters were named. This led me to adding a small CLI wrapper using Symfony's [Console][15] component. By reusing schema I extracted earlier, the CLI is able to show me details about the available API methods in addition to details about individual method parameters...

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    ./ravelry-api
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    ...snip...
      topics:reply                    Post a reply to a topic
      topics:show                     Get topic information
      topics:update                   Update a topic
    upload
      upload:image                    Upload an image file for later processing or attaching
      upload:request-token            Generate an upload token
      upload:status                   Get uploaded image IDs
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    ./ravelry-api yarns:search --help
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    Usage:
    yarns:search [--page="..."] [--page-size="..."] [--query="..."] [--sort="..."] [--facet="..."] [--debug] [--etag="..."] [--extras]

    Options:
    --query                Search term for fulltext searching yarns
    --page                 Result page to retrieve. Defaults to first page.
    ...snip...
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

Once I know my method and parameters, I can execute it to get the JSON result...

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    ./ravelry-api yarns:search --query 'cascade 220' --sort rating
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    { "yarns": [
      { "rating_count": 924,
        "machine_washable": false,
        "texture": "Plied",
        "yarn_company_name": "Cascade Yarns",
        ...snip...
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

To further help in my development and debugging, I also made sure that I could increase verbosity with `-vv` to dump out raw HTTP requests and responses. This turned out to be immensely useful when tracking down a [strange bug][17] around the ordering of parameters in the `Content-Disposition` request header used for file uploads.

If you happen to be on OS X or have PHP, you can visit the [releases][18] page to download the [Phar file][19] and give it a try.

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    wget https://github.com/dpb587/ravelry-api-cli.php/releases/download/v0.2.0/ravelry-api-cli.phar
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    chmod +x ravelry-api-cli.phar
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    ./ravelry-api-cli.phar list
    ```

  {{< /terminal-input >}}

{{< /terminal >}}

# With TLE Orders {#with-tle-orders}

One of Ravelry's main features is the ability for users to "stash" yarn that they buy so they can remember and allocate it to projects as they use it. A couple weeks ago, I enabled a button on TLE that allows customers to automagically stash their yarn purchases from an order. For customers who don't enjoy spending the time finding and entering the stash details, we're now able to do it for them in several seconds. With two clicks, a popup will show them that we have taken care of stashing their purchases...

{{< image alt="Orders Screenshot" src="https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2016-01-21-writing-a-php-client-for-the-ravelry-api/with-tle-orders.png" >}}

When customers return to their order page in the future, they see a Stash link which takes them to their stash page on Ravelry to further plan or see what project they used the purchase in. Customers are really enjoying the added convenience of this new feature and, for the time being, it's a privilege to be the only shop offering such a feature.


# Open Source {#open-source}

When I first started experimenting with the API, there were not any useful PHP solutions. Still today there are only a [couple][20] [snippets][21] and they show how to sign Ravelry API requests. By open-sourcing this project, I'm hoping it can enable others to more easily innovate and create projects that the knitting+Ravelry community can really enjoy. If you end up using it, feel free to make an issue or pull request about troubles or areas that can be improved.


 [1]: http://www.ravelry.com/
 [2]: http://www.ravelry.com/about
 [3]: https://www.theloopyewe.com/
 [4]: https://github.com/dpb587/ravelry-api.php
 [5]: https://github.com/dpb587/ravelry-api-cli.php
 [6]: https://packagist.org/packages/dpb587/ravelry-api
 [7]: https://getcomposer.org/
 [8]: http://docs.guzzlephp.org/en/latest/
 [9]: http://www.ravelry.com/api
 [10]: https://github.com/dpb587/ravelry-api.php/blob/f3a36eaf3cf853264bef88faa3d71c5f3301b279/bin/schemagen.php
 [11]: https://www.w3.org/TR/xquery/
 [12]: https://github.com/guzzle/guzzle-services
 [13]: https://github.com/dpb587/ravelry-api.php/tree/master/test
 [14]: https://phpunit.de/
 [15]: https://github.com/symfony/console
 [16]: https://en.wikipedia.org/wiki/Regular_expression
 [17]: http://www.ravelry.com/discuss/ravelry-api/2936052/1-25
 [18]: https://github.com/dpb587/ravelry-api-cli.php/releases
 [19]: http://php.net/manual/en/intro.phar.php
 [20]: https://github.com/willsteinmetz/ravelry-php-api
 [21]: https://gist.github.com/gavinr/886876
