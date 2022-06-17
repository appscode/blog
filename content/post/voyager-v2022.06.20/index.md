---
title: Announcing Voyager v2022.06.20
date: 2022-06-20
weight: 25
authors:
  - Tamal Saha
tags:
  - cloud-native
  - kubernetes
  - ingress
  - gateway
  - haproxy
  - voyager
---

We are pleased to announce the release of Voyager v2022.06.20. In this release, we have released operator and HAProxy images to fix CVE-2022-1586 & CVE-2022-1587. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2022.06.20/README.md).

## **CVE Fixes**

We have updated the base image used for Voyager operator to address CVE-2022-1586 & CVE-2022-1587.


## **ExternalName Service Fixes**

In this release we have fixed a regression bug when Ingresses use ExternalName Services as backends.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Voyager community, join us in the AppsCode Slack team channel `#general`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
