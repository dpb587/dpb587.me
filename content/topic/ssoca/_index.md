---
description: Service-oriented certificate authority for single sign-on.
identifier:
- propertyID: github/repository
  value: https://github.com/dpb587/ssoca
params:
  projectType:
    group: Network Security
    languages:
    - Go
    status: inactive
  topics:
    golang: {}
    openvpn: {}
title: SSOCA Service Manager
---

I started this project after getting tired of manually managing certificates for the OpenVPN servers I managed. It was also really helpful for learning more about PKIs and better understanding how services rely on certificate-based authentication. I'm particularly proud of:

* working through the broad implementations of user-authentication, certificate management, and service integrations.
* adoption by several teams in my company to protect both development and production environments.
