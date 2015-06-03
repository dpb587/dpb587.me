---
title: "New BOSH Release for OpenVPN"
layout: "post"
tags: [ "bosh", "openvpn" ]
description: "Open sourcing a new BOSH release for managing an OpenVPN network."
code: https://github.com/dpb587/openvpn-boshrelease
---

I'm a big fan of [OpenVPN][1] - both for personal and professional VPNs. Seeing as how I've been deploying more things with [BOSH][2] lately, an OpenVPN release seemed like a good little project. I started one about nine months ago and have been using development releases ever since, but last week I went ahead and created a ["final" release][6] of it.

There is only a single job (`openvpn`) and the properties are [well documented][3]. Its primary purpose is to act as a server for other clients to connect to, however you can also configure it to connect as a client and connect to another OpenVPN network as well. This makes it very easy to join multiple networks from a single OpenVPN connection.

One of the more complicated steps of configuring an OpenVPN server is figuring out and remembering the correct commands for creating and signing security keys and certificates. The [README][4] includes all those steps to get a server running in a deployment and a client connected to it. There are also a few other examples about some fancier configuration options such as: setting up `iptables` for shared networks, allowing VPN clients to communicate with each other, and making sure specific clients are assigned static IPs.

After going through the process of setting up quite a few OpenVPN servers and trying to automate and maintain them, this BOSH release has become my preferred method given its flexibility, consistency, and handy readme so I'm no longer Googling at every step. Check out the [project page][5] if you'd like to learn more, or see the [releases][6] page there for a tarball that you can use in your own BOSH environment.


 [1]: https://openvpn.net/
 [2]: http://bosh.io/
 [3]: https://github.com/dpb587/openvpn-boshrelease/blob/89fd58982db3327e26cb8e2b9ed06835ffb08dd1/jobs/openvpn/spec#L17
 [4]: https://github.com/dpb587/openvpn-boshrelease/blob/master/README.md
 [5]: https://github.com/dpb587/openvpn-boshrelease
 [6]: https://github.com/dpb587/openvpn-boshrelease/releases
