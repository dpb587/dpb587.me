---
title: Embeddable and Context-Aware Web Pages
layout: post
tags: [ 'architecture', 'http', 'javascript', 'symfony', 'symfony2' ]
description: Embedding content in an absolutely relative manner.
---

In my [symfony][5] website applications I frequently make multiple subrequests to reuse content from other controllers.
For simple, non-dynamic content this is trivial, but when arguments can change data or when the browser may want to
update those subrequests things start to get complicated. Usually it requires tying the logic of the subrequest
controller in the main request controller (e.g. knowing that the `q` argument needs to be passed to the template, and
then making sure the template passes it in the subrequest). I wanted to simplify it and get rid of those inner
dependencies.

As an example, take a look at this [product search][1]. The [facets][2] and [results][3] are actually subrequests, but
the main results content is taking advantage of the request design I implemented. My goals were:

 * remove logic from controller code to keep them independent from each other,
 * pages work without JavaScript and without requiring newer browsers,
 * pages work the same whether it's a subrequest or a master request, and
 * any page should be capable of being a self-contained subrequest.


## Steps

When a subrequest is self-contained, I call it a *subcontext*. These subcontext requests have an additional requirement
of being publicly accessible. In the product search, the [results][3] page is publicly routed and all the pagination and
view links will work properly within the `./results.html` page. This makes it easy for using XHR to load updated
content.

Another minor piece of this design is that views don't need to be fully rendered. This means an Ajax request can ask for
just the page content and exclude the typical header/footer. In [Twig][4] parlance it is a `frag_content` block which
has all the useful content.

When it comes to passing query parameters down through subcontexts, I decided that each subcontext gets its own scoped
variable. So whenever I render a subcontext in a template, I always specify a name for it. The name should be unique
within the template context. In the product search example, the facets subcontext is named `f` and the results
subcontext is named `r`. When a request arrives for `/?r[offset]=54`, the subrequest will arrive at the results
controller looking like `/results.html?offset=54` (which is equivalent to navigating that page directly).

To keep track of the subcontext names, template content, query data, and relative locations I started using a custom
request header named `tle-subcontext`. In practice it looks like:

    tle-subcontext: r:content@/shop/search/availability/in-stock/?q=red

When that request header exists it means:

 * we're within a subcontext named `r`,
 * we want to get the view fragment named `content`, and
 * the root URL we started at was `/shop/search/availability/in-stock/?q=red`.

Within the controller code that header information should not be relevant. In templating though it becomes useful for
rewriting URLs. Whenever a template is going to give a link to itself, I wrap it in a custom `subcontext_rewrite`
function. For example, given the `tle-subcontext` configuration above, it would rewrite:

    dataset_generic(...snip...)
    => /shop/.../in-stock/results.html?q=red&view=list-tn&offset=54

    subcontext_rewrite(dataset_generic(...snip...))
    => /shop/.../in-stock/?q=red&r[view]=list-tn&r[offset]=54#r

The rewritten URL is completely valid and can be accessed without fancy JavaScript calls. Now, to make that possible I
don't use the standard inline renderer in Twig. I created a custom renderer with a little additional logic which takes
care of rewriting the subcontext data and injecting the header:

{% highlight php %}
$rootUri = $request->getRequestUri();

if (preg_match('/^([a-z0-9\-]+):([a-z0-9]+)@(.*)$/', $request->server->get('HTTP_TLE_SUBCONTEXT'), $match)) {
    # this means a subcontext already exists and a sub-subcontext is being created

    # append our context name to the parent context name
    $options['name'] = $match[1] . '-' . $options['name'];

    # use the root uri from the header since $request is only a subrequest
    $rootUri = $match[3];

    # pull out our context-specific query data from the root uri and update our request
    parse_str(parse_url($match[3], PHP_URL_QUERY), $rootQuery);
    $subRequest->query->replace(isset($rootQuery[$options['name']]) ? $rootQuery[$options['name']] : array());
} elseif ((null !== $subdata = $request->query->get($options['name'])) && (is_array($subdata))) {
    # pull out our context-specific query data
    $subRequest->query->replace($subdata);
}

# now add the header with all our combined data to the request
$subRequest->server->set(
    'HTTP_TLE_SUBCONTEXT',
    $options['name'] . ':' . (empty($options['frag']) ? 'content' : $options['frag']) . '@' . $rootUri
);

unset($options['name'], $options['frag']);
{% endhighlight %}

So now whenever I want a subcontext within a view, I can use the custom renderer:

{% highlight jinja %}{% raw %}
{{ render_subcontext(path('search_results', passthru), { 'name' : 'r' }) }}
{% endraw %}{% endhighlight %}

With those simple customizations I no longer have to worry about knowing what parameters need to be passed on to
template subrequests. It also paves the way for some more fancy behavior...


## Adding Some Magic

Since the subcontext pages are publicly accessible, it should be easy to let Ajax reload individual subcontexts without
having to reload the whole page. To enable that, I went ahead and configured subcontext requests to always end up in a
specific layout which will wrap it with the subcontext metadata. The template looks like:

{% highlight jinja %}{% raw %}
<article id="{{ subcontext_name() }}" data-href="{{ app.request.uri }}#{{ subcontext_frag() }}">
    <header><h3>{{ block('def_title') }}</h3></header>
    <section>{{ block('frag_' ~ subcontext_frag()) }}</section>
</article>
{% endraw %}{% endhighlight %}

The `subcontext_*` custom functions simply peek at the request to find the `tle-subcontext` header and appropriate
values.

Now that the extra data is available, we can have JavaScript build links to send partial requests. If a
`subcontext_rewrite` link is clicked, it's a matter of starting with the `article[@data-href]` value and peeking at the
clicked `a[@href]` to find query parameters that were within the `article[@id]` name. For example:

    // window.location
    /shop/search/availability/in-stock/?q=red

    // subcontext
    <article id="r" data-href="/shop/search/availability/in-stock/results.html?q=red#content">

    // clicked anchor
    <a href="/shop/search/availability/in-stock/?q=red&r[offset]=54#r">

    // becomes the request
    GET /shop/search/availability/in-stock/results.html?q=red&offset=54
    tle-subcontext: tle-subcontext: r:content@/shop/search/availability/in-stock/?q=red

Something easily processable with an Ajax request. And since the clicked anchor was a canonical URL, that easily becomes
the new window URL location by using the [HTML5 History API][6].


## Conclusion

Once I implemented the code snippets for tying all the ideas together, it became much quicker and simpler for me to
embed other dynamic controllers within my requests. So far it has been working out quite well and I no longer have to
worry about page-specific hacks for passing data to subrequests.


 [1]: http://www.theloopyewe.com/shop/search/availability/in-stock/?q=red
 [2]: http://www.theloopyewe.com/shop/search/availability/in-stock/facets.html?q=red
 [3]: http://www.theloopyewe.com/shop/search/availability/in-stock/results.html?q=red
 [4]: http://twig.sensiolabs.org/
 [5]: http://symfony.com/
 [6]: http://diveintohtml5.info/history.html
