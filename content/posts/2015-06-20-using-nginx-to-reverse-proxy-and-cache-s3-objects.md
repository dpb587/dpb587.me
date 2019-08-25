---
'@context': http://schema.org
'@type': BlogPosting
datePublished: "2015-06-20"
description: Using S3 as an upstream server for improving long-tail traffic.
keywords:
- aws
- aws-s3
- caching
- nginx
- reverse-proxy
- s3
- upstream
name: Using nginx to Reverse Proxy and Cache S3 Objects
url:
- /blog/2015/06/20/using-nginx-to-reverse-proxy-and-cache-s3-objects.html
---

My most recent project for [TLE][1] has been focused on making the infrastructure much more "cloud-friendly" and resilient to failures. One step in the project was going to require that more than one version might be running at a given time (typically just while a new version is still being rolled out to servers). The application itself doesn't have an issue with that sort of transition period, however, the way we were handling static assets (like stylesheets, scripts, and images) was going to cause problems. First, some background...

When the frontend application code gets built and packaged up, it only contains the static assets for its own version. The static assets get dumped into `/docroot/static/{hash}/`, where the hash is generated based on when they were last modified and build runtime details. Once the application gets deployed and symlinked live, the old versions are no longer accessible from the document root. This obviously has implications like:

 0. Late requests for those old assets result in 404s (infrequently users, usually bots).
 0. Application servers must be reloaded onto the new version at the same time (otherwise, an old server without the new assets might be used by the proxy).

Additionally, we use [CloudFront][2] as a CDN for those static assets with our website configured as the origin. If the CDN gets back a 404 for an asset (old or new) it is cached for a short period and potentially affects a lot of clients (particularly bad if it happens on the upcoming, new version). Since CloudFront supports [S3][4] buckets as origins, I figured we could use it to store all the versions of our static assets. I quickly added a step to the deployment process which uploads new assets to a bucket. However, that was only part of the solution.

Unfortunately, CloudFront doesn't support dynamic [gzip][5] compression - it will only send back, byte-for-byte, what the origin delivers and we were storing the plain, non-gzipped versions in S3. The options were to...

 0. no longer provide the files in gzip form (bad option... some files are genuinely large);
 0. store both plain and gzip versions in separate S3 objects, then change the web application to dynamically rewrite the `link`/`script`/URLs based on browser headers (a lot of work, fragile, and bad use of existing web standards); or
 0. continue using our website as the origin where responses could correctly be `Vary`'d and conditionally compressed.

The last one was definitely my preferred choice, but we would still have the problem of a single version being on the filesystem and unpredictable results when multiple application server versions were running behind the proxy. After some thought, I wanted to try using the S3 bucket as an upstream and avoiding the application servers altogether. And to improve latency and minimize the external, S3 requests I could cache them locally. After some experimentation, I ended up with something like the following in our [nginx][3] configs...

```nginx
location /static/ {
  # we can only ever GET/HEAD these resources
  limit_except GET {
      deny all;
  }

  # cookies are useless on these static, public resources
  proxy_ignore_headers set-cookie;
  proxy_hide_header set-cookie;
  proxy_set_header cookie "";
  
  # avoid passing along amazon headers
  # http://docs.aws.amazon.com/AmazonS3/latest/API/RESTCommonResponseHeaders.html
  proxy_hide_header x-amz-delete-marker;
  proxy_hide_header x-amz-id-2;
  proxy_hide_header x-amz-request-id;
  proxy_hide_header x-amz-version-id;
  
  # only rely on last-modified (which will never change)
  proxy_hide_header etag;

  # heavily cache results locally
  proxy_cache staticcache;
  proxy_cache_valid 200 28d;
  proxy_cache_valid 403 24h;
  proxy_cache_valid 404 24h;

  # s3 replies with 403 if an object is inaccessible; essentially not found
  proxy_intercept_errors on;
  error_page 403 =404 /_error/http-404.html;

  # go get it from s3
  proxy_pass https://s3-us-west-1.amazonaws.com/example-static-bucket$1;

  # annotate response about when it was originally retrieved
  add_header x-cache '$upstream_cache_status $upstream_http_date';
  
  # heavily cache results downstream
  expires max;
}
```

So, with the above configuration...

 * CloudFront still points to our website and we can serve gzip/plain at the same resource;
 * assets are kept around indefinitely (and we could utilize bucket lifecycle policies if it becomes an issue);
 * frontend web server no longer relies on a particular application server's filesystem;
 * access to the S3 bucket/prefix can be restricted via bucket policy; and
 * most importantly... deployment timing is no longer critical - versions can be deployed at whatever pace is appropriate and possible.

Since deploying these changes over a month ago, everything has been working very well and the number of static 404 nuissances in our error logs have dropped significantly. It also made it much easier to move onto the next problem on the path to cloud-friendliness and resiliency...


 [1]: https://www.theloopyewe.com/
 [2]: http://aws.amazon.com/cloudfront/
 [3]: http://nginx.org/
 [4]: http://aws.amazon.com/s3/
 [5]: https://en.wikipedia.org/wiki/Gzip
