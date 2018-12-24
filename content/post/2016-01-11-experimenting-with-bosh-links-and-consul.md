---
date: 2016-01-11
title: "Experimenting with BOSH Links and Consul"
description: "Integrating consul and links metadata for inter-deployment service dependencies."
tags:
- bosh
- consul
- service-discovery
aliases:
- /blog/2016/01/11/experimenting-with-bosh-links-and-consul.html
---

With BOSH, I use deployments to segment various services. For example, [TLE][1] has several services like web and database servers, WordPress blogs, the main e-commerce application, statistics, and internal services. Many of them are interconnected in some way. Historically I've used a combination of hard-coded IP addresses in the deployment properties and dynamic service discovery with [consul][2]. With a small bit of tweaking and an extra pre-parser, I'm now able to emulate much of the proposed links features, but from a more dynamic, distributed perspective.

<!--more-->


## About Links

The main goal of links is to interlink services across jobs and deployments without having to hardcode IP addresses everywhere. Links are a feature with a [bosh-notes plan][3] and will hopefully be implemented soon.

Since I have already been using consul for service discovery, I figured it would be an easy experiment to add the proposed links metadata to my internal releases' job `spec`s and see how well everything could play together. Since this is a userland implementation, things are a bit different than what the actual BOSH features will eventually support. For example, link properties will have their own template function, but here they're injected as a regular property.


## Consul

Since I already have a consul deployment using the [community release][9], I'll continue to use it. However, since I'll be using more of the DNS features, I needed to add those IPs to my networks' DNS configuration...

```yaml
networks:
 - name: "default"
    subnets:
      dns:
        - "192.0.2.21"
        - "192.0.2.22"
        - "192.0.2.23"
...snip...
```

Now, instead of using IPs I can switch to something like `{deployment}--{link_name}.service.consul`.


## DNS

Sometimes it's easiest to just use DNS which will point to all healthy hosts. For that, I can use `.dns` of my link which will be something like `wordpress--db.service.consul`. Since DNS is managed by consul, it will always return IPs of healthy instances. This was the easiest method for migrating my existing releases which were previously using hard-coded IPs.

    $ cat jobs/app/parameters.yml
    parameters:
      database_host: "<%= p('_links.webapp.consumes.db.dns') %>"
      database_port: <%= p('_links.webapp.consumes.db.properties.port') %>


## Health Checks

One of the advantages of consul is that it has built-in support for [health checks][5] to detect if a service has a fault and should be removed. To utilize this, I let my release jobs include a file at `$JOBDIR/links/{link-name}.consul.json` with the JSON of their health checks. For example...

```json
[ { "http": "http://127.0.0.1:8080/status/fpm",
    "interval": "15s",
    "timeout": "3s" } ]
```

The file gets picked up by my consul-links release and the consul agent will locally check whether the app server is healthy. When it's unhealthy, consul will immediately remove itself from the list of healthy endpoints that DNS will return.


## Templates

A more advanced solution to DNS is [consul-template][6]. This tool will automatically update a configuration file whenever your healthy endpoints change for a service. For example, in nginx you could write a config template which looks like...

```go-text-template
upstream webapp {
  {{ range service "<%= p('_links.frontend.consumes.app.service') %>" }}
    server {{.Address }}:{{ .Port }};
  {{ else }}
    server 127.0.0.1:8911;
  {{ end }}
}
```

Within the job, you'd then have two monit processes - one which actually runs the web server and another which runs consul-template to monitor for changes. When consul-template notices changes, it updates the config file and sends a `SIGHUP` to nginx to make it live-reload the config. The control script looks something like...

```bash
NGINX_TMPL="/var/vcap/jobs/$JOB_NAME/etc/nginx.conf.tmpl"
NGINX_CONF="/var/vcap/jobs/$JOB_NAME/etc/nginx.conf"
NGINX_RELOAD="/var/vcap/jobs/$JOB_NAME/bin/nginx-control reload"

CMD="/var/vcap/packages/consul-template/bin/consul-template -consul localhost:8500 \
    -template \"$NGINX_TMPL:$NGINX_CONF:$NGINX_RELOAD\""

case $1 in
  once)  eval $CMD -once ;;
  start) exec $CMD --pid-file="$PIDFILE" ;;
  stop)  kill -s INT $( cat "$PIDFILE" ) ;;
esac
```


## Maintenance

One other feature of consul that I use is the ability to mark a service as going into maintenance mode. Instead of waiting for consul health checks to notice an issue, I can forcefully remove it from the list of healthy endpoints. I often use this approach in my [`drain script`][7]...

```bash
/var/vcap/packages/consul/bin/consul maint -enable \
  -service="<%= p('_links.frontend.provides.app.service') %>" \
  -reason=bosh.deployment.job.drain \
  > /dev/null
```

This will immediately remove the app server (causing consul-template to remove it from the downstream web server) and then I can poll until the app is no longer finishing requests before exiting.


## consul-ui

One of the nice perks of consul is being able to get a live view of all the services which are shared within a deployment or with other deployments.

![Screenshot: consul-ui](https://dpb587-website-us-east-1.s3.amazonaws.com/asset/blog/2016-01-11-experimenting-with-bosh-links-and-consul/consul-ui.png)


## Conclusion

I published my code for this userland implementation as [`consul-bosh-links`][8]. Since annotating my release jobs with the proposed provides/consumes metadata and using my userland implementation, my deployment manifests have become a bit cleaner, my releases are a bit cleaner, and I'm better prepared for possibly using official links features whenever they get released.


 [1]: https://www.theloopyewe.com/
 [2]: https://consul.io/
 [3]: https://github.com/cloudfoundry/bosh-notes/blob/master/links.md
 [4]: https://consul.io/docs/agent/options.html#dns_config
 [5]: https://consul.io/docs/agent/checks.html
 [6]: https://github.com/hashicorp/consul-template/
 [7]: https://bosh.io/docs/drain.html
 [8]: https://github.com/dpb587/consul-bosh-links-release
 [9]: https://github.com/cloudfoundry-community/consul-boshrelease/
