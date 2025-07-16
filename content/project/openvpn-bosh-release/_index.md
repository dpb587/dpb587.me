---
description: A cloud-agnostic method of deploying OpenVPN with BOSH.
identifier:
- propertyID: github/repository
  value: https://github.com/dpb587/openvpn-bosh-release
keywords:
- bosh-release
- openvpn
title: OpenVPN (BOSH Release)
---

This was one of my first BOSH releases and I created it to deploy [OpenVPN](https://openvpn.net/) with [BOSH](https://bosh.io/) &ndash; something I'd been deploying manually, Puppet, and CloudFormation previously. I'm particularly proud of:

* adoption by several teams in my company to protect both development and production environments.
* the CI which automatically monitors, tests, and upgrades upstream dependencies.
* the [docs site](https://dpb587.github.io/openvpn-bosh-release/docs/latest/) which is generated for each release version.
