---
title: Distributed Docker Containers
layout: post
tags: aws-ec2 docker nodejs scs-utils
description: A strategy for integrating Docker services across multiple hosts and data centers.
---

One thing I've been working with lately is [Docker][1]. You've probably seen it referenced in various tech articles
lately as the next greatest thing for cloud computing. Docker runs "containers" from base "images" which essentially
allow running many lightweight virtual machines on any recent, Linux-based system. Internally, the magic behind it is
[lxc][2], although Docker adds a lot more magic to improve and make it more usable.

For a long time now I've used virtual machines for development - it allows me to better simulate how software runs out
on production servers. Historically, [Vagrant][3] + [VirtualBox][4]/[VMWare Fusion][5]/[EC2][6] have been great tools
for that, but they have limitations and they tend to drift a bit from production architecture.


## The Problem

In trying to duplicate the production environments, it's not typically feasible for me to run more than one virtual
machine on my laptop. I could split my single local virtual machine to multiple EC2 instances; but then it becomes more
difficult to manage IP addresses for the various service dependencies as the instances get stopped/started between
working sessions (in addition to the extra costs). VPCs with private IP addresses do help with that a lot, as long as
there's a sane way to manage those resources.

Another issue that comes up when combining services on a single host is dependency overlap. One example of this is
shared modules. Some newer features of nginx require a newer version of the openssl libraries. However, PHP doesn't
necessarily support the newer version of openssl without upgrading quite a few other components. While there may be
workarounds, the inconvenience of it all typically just prompts me to avoid working on that particular feature,
unfortunately.

Ultimately, I want to have the same software and network stack that I use in a production environment, but in a
development environment and, if possible, locally on my laptop.


## The Alternatives

This problem is certainly not unique, but a practical solution has been difficult for me to find. I've been
experimenting with a few different technologies over the years trying to solve this sort of thing.

Vagrant is obviously the first practical solution. For me, it has been a functional solution for quite a while, but not
an optimal one. Like I mentioned before, it's a bit bulky when attempting to mimic non-trivial architectures on a
standard laptop. For a while now, I've been finding the motivation and time to migrate to a better setup.

With the advent of Docker, many of my software requirements become much simpler. Each piece of software can run in its
own container and I don't have to worry about dependency overlap. Multiple containers are *significantly* cheaper than
trying to run multiple virtual machines. I could even reuse containers built on my development machine out on
production. One thing Docker doesn't effectively solve is service dependency. It can support them on a single host with
links, but not across multiple hosts.

I've been keeping an eye out for other tools which may help solve these problems. Some of them are:

 * [decking][7] - seems to primarily build on top of Docker's built-in link functionality for service dependency within
   a single host
 * [etcd][11] - an excellent distributed, hierarchical key-value store; very useful for monitoring configuration values
   and being notified when they change (related: [confd][22])
 * [fig][8] - seems like [Foreman][21], but geared for Docker containers
 * [flynn][11] - originally I was very excited about this, however it still seems underdeveloped for the purposes of
   service discovery of arbitrary services; I'm still very hopeful
 * [serf][9] - a very new client for distributing data across a cluster and taking action on it. To me it seems like
   more of a management tool (like half of the [mcollective][10] utility)

Recently, I've been becoming more aquainted with [bosh][12], an interesting tool for managing large deployments along
with all their dependencies. To me, bosh always seems overly complicated for whatever I'd want to accomplish and has
quite a few bosh-specific practices to learn. Its resource and service management is very thorough, although it takes a
while to get comfortable with it. It seems more like an infrastructure management tool rather than a service management
tool, and I was hoping to keep those responsibilities separate and simpler. Ultimately, I think bosh could be made to
work... but I was still hoping for something different, lighter, and utilizing more common open source tools that I was
already familiar with.


## The Ideas

I had a simple application in mind to roughly define my "[minimum viable product][13]":

 0. run WordPress web application, a MySQL server, and a backup MySQL server as separate services
 0. runtime parity (between development and production)
     1. configure services the exact same way
     1. run services the exact same way
     1. depend on other services the exact same way
 0. architecture flexibility
     1. in production, run the services on three separate hosts across two separate data centers
     1. in development, run all services on a single virtual machine on my laptop
 0. service flexibility - be able to dynamically relocate services without manual reconfiguration and minimal downtime
     * combine services into one or two hosts during quiet hours
     * move a service to a more powerful instance during high load
 0. self-provisioning - when a container requires a particular volume or network, make sure it can be automatically
    provisioned and de-provisioned

First off, I knew I wanted to run the services inside of Docker containers. I can only imagine Docker's ubiquity will
continue to grow, and the ability to run completely arbitrary software anywhere with minimal host dependencies seemed
like a perfect, lightweight solution.

I've used [Puppet][14] to configure servers and applications for a long time. While I dislike the overhead it requires
for smaller use cases, I really like the consistency and declarative nature that it provides. Since I'll continue to use
it for host server configuration, it's a small stretch to also use it for configuring the service runtimes.

When it comes down to it, I think there are two main questions that a service must answer:

 * How should I work? and
 * How do I connect with the rest of the world?

The first question can be managed and configured via Puppet. Once a service is configured and compiled to run as
requested, it never needs to go through that process again. This approach lets compiled Docker images be consistently
reused across time and servers.

The second question deals with pointing WordPress to the MySQL server, or pointing MySQL server to the data directory,
or running the MySQL backup server on a specific network segment. These decisions and connections have nothing to do
with how the service should work, so they can be changed as needed. So far, I have four main dependencies about how
these containers get connected:

 0. volumes - giving containers a place to write persistent data (e.g. WordPress `wp-content/uploads` directory)
 0. provided services - a service that the container is running (e.g. `http` on `80/tcp`)
 0. required services - a service that the container needs (e.g. `mysql`)
 0. network - how the container is attached to the network

I think these basic aspects effectively describe everything needed to manage a self-contained service.


## The Implementation

The next step of an idea is to prototype it, and that's where I am today. There are several pieces that I've been
working on, but three general topics...


### Service Discovery

One of the most interesting concepts is service discovery. I wanted containers to be able to connect with each other
across multiple hosts and data centers. I've been using DNS for host discovery and, while it works great it doesn't seem
entirely appropriate for "containerized" discovery. Through [`A`][23] records, DNS easily picks up on hosts changing,
but is not so good for dynamic ports. DNS [`SRV`][24] records seem *much* more appropriate with attributes for both
hostname and port, but `SRV` records are rarely used in internal APIs.

Originally I was using etcd to register and discover services, but I found it to be inefficient for filtering services
and propagating changes. Instead, I created a specialized client/server protocol to handle the registration and
discovery process. In technical terms, the protocol works like the following...

WordPress needs a database, so before it starts the container, it connects with the disco server:

 > **container**: Hi, I need a `mysql` service to talk to - who's available?  
 > **disco**: You should talk with `192.0.2.11:39313` - I'll keep you posted if it changes, but let me know if you no
 > longer need it

The results are injected as environment variables when the container is started and can use them however it likes.
WordPress obviously runs a web server, so, once the container is started, the container manager connects with disco:

 > **container**: Hi, I'm `wordpress` and I have an `http` service available at `192.0.2.12` on port `39212`  
 > **disco**: Nice to meet you; let me know if you no longer provide it

Then things are running happily and you could ask the disco server where to find `wordpress/http` to pull it up in your
web browser. If the database server crashes and recovers elsewhere, a few things will happen. First, when disco realizes
MySQL is no longer available (either by a clean disconnect, heartbeat timeout, or socket disconnect), it notifies
everyone who is subscribed that the endpoint has been dropped:

 > **disco**: Looks like you were using `mysql`, but I'm sorry to tell you it's no longer available  
 > **container**: Thanks for letting me know

The container manager then attaches to the container to run an update command letting it know about the change. The
command can take care of updating the runtime configuration and restarting the application server.

Eventually the new MySQL server will come back online and register itself. Once registered, disco realizes that
WordPress is subscribed, so it lets it know:

 > **disco**: Great news, I have a new `mysql` endpoint for you at `192.0.2.14.39414`  
 > **container**: Excellent, thanks

And it again runs the live update command, updating the environment and restarting the application server.

The disco protocol has a few more features (like using a single server for more than one WordPress/MySQL setup, or
filtering services by arbitrary tags like availability zones to improve load balancing), but that's the general idea.


### Configuration Files

I'm using YAML files to describe images and containers. They get compiled to a static version, and then cached based on
the image configuration. For example, take a look at this example [scs-wordpress][16] image manifest. It describes the
various connection points, docker details, and how it's configured. Now, take a look at the [Puppet manifests][17] which
enumerates all the configuration options which affect how the service will run. Finally, take a look at the
[sample config][18] which ties together what kind of image it needs to be able to run (configuration) and how that image
will be connected to the world.


### Self-Provisioning

For each of the four dependency/connection types (volumes, service provider, service dependent, network), I'm trying to
make them suitable for local development and AWS EC2 deployment. For example:

 * AWS EC2 volumes can be auto-created, mounted, and attached to hosts for use by docker containers. This allows
   services to drift across instances
 * Likewise, I can also just use a local path for a volume and avoid an official network mount
 * Various other strategies can be added for each dependency:
    * nfs-volume: to attach a docker mount point to an external NFS mount
    * aws-ec2-eni: to attach an ENI as the network interface for a docker container

My goal is to provide a manifest configuration file to a machine and know that it will load up whatever it needs to run,
including recompiling the image from scratch if it's not available in any caches.


## The Prototype

So, all those ideas are currently under development in my [`scs-utils`][20] repository. I've created a repository called
[`scs-example-blog`][19] which is a functional implementation of my original MVP. It provides a `Vagrantfile` for you to
easily try it out yourself and it goes through the process of getting the containers running on a single virtual machine,
accessing the services from the host, and then splitting them up across multiple virtual machines. It's more a tutorial
describing the steps - typically the service deployment would be managed by Puppet.


## The Conclusion

All these ideas are absolutely a work in progress and I'm still actively tweaking the implementation, but it was in a
functional state to briefly discuss the idea. So far it has been an excellent learning opportunity for Docker, custom
network protocols, and splitting some of the services I've previously been running into more reusable components. Even
if `scs-utils` isn't still what I'm using in 2 years, the refactoring it has motivated makes it significantly easier to
port into whatever more valuable tool surfaces further down the road.


 [1]: https://www.docker.io/
 [2]: http://linuxcontainers.org/
 [3]: http://www.vagrantup.com/
 [4]: https://www.virtualbox.org/
 [5]: http://www.vmware.com/products/fusion
 [6]: http://aws.amazon.com/ec2/
 [7]: http://decking.io/
 [8]: http://orchardup.github.io/fig/
 [9]: http://www.serfdom.io/
 [10]: http://puppetlabs.com/mcollective
 [11]: https://flynn.io/
 [12]: http://docs.cloudfoundry.org/bosh/
 [13]: http://en.wikipedia.org/wiki/Minimum_viable_product
 [14]: http://puppetlabs.com/puppet/puppet-open-source
 [15]: https://github.com/coreos/etcd
 [16]: https://github.com/dpb587/scs-wordpress/blob/3ba391d4f82da5c9642d88962e0bce32eb692add/scs/image.yaml
 [17]: https://github.com/dpb587/scs-wordpress/tree/3ba391d4f82da5c9642d88962e0bce32eb692add/scs/puppet/scs/manifests
 [18]: https://github.com/dpb587/scs-example-blog/blob/master/wordpress/manifest.yaml
 [19]: https://github.com/dpb587/scs-example-blog
 [20]: https://github.com/dpb587/scs-utils
 [21]: http://ddollar.github.io/foreman/
 [22]: https://github.com/kelseyhightower/confd
 [23]: http://en.wikipedia.org/wiki/A_record#A
 [24]: http://en.wikipedia.org/wiki/SRV_record
