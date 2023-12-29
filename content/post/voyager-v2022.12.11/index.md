---
title: Announcing Voyager v2022.12.11
date: "2022-12-11"
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

We are pleased to announce the release of Voyager v2022.12.11. In this release, we have released operator and HAProxy images to fix a number of CVEs. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2022.12.11/README.md).

## **CVE Fixes**

We have updated the docker images in this release to address the following CVEs:

- CVE-2022-41717 

## **HAProxy Version**

We have updated HAProxy images to the following version:

- appscode/haproxy:2.7.0-alpine
- appscode/haproxy:2.6.7-alpine
- appscode/haproxy:2.5.10-alpine

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
