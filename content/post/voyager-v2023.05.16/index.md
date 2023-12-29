---
title: Announcing Voyager v2023.05.16
date: "2023-05-16"
weight: 25
authors:
- Tamal Saha
tags:
- cloud-native
- gateway
- haproxy
- ingress
- kubernetes
- voyager
---

We are pleased to announce the release of Voyager v2023.05.16. In this release, we have updated HAProxy image to 2.6.13 and fixed various CVEs. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2023.05.16/README.md).

## **CVE Fixes**

We have updated the docker images in this release to address the following CVEs:

- GHSA-vvpx-j8f3-3w6h	
- CVE-2023-0464	
- CVE-2022-41723	
- CVE-2020-15888	

## **HAProxy Version**

We have updated HAProxy images to the following version:

- ghcr.io/voyagermesh/haproxy:2.7.8-alpine
- ghcr.io/voyagermesh/haproxy:2.6.13-alpine
- ghcr.io/voyagermesh/haproxy:2.5.14-alpine

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
