---
title: Announcing Voyager v2022.08.17
date: "2022-08-17"
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

We are pleased to announce the release of Voyager v2022.08.17. In this release, we have released operator and HAProxy images to fix a number of CVEs. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2022.08.17/README.md).

## **CVE Fixes**

We have updated the docker images in this release to address the following CVEs:

- CVE-2022-1996
- CVE-2022-37434
- CVE-2021-33194
- CVE-2021-44716
- CVE-2021-38561
- CVE-2022-28948
- CVE-2022-2097
- CVE-2021-20329
- CVE-2021-31525
- CVE-2022-29526
- GHSA-r489-9g5r-8q2h
- GHSA-f6mq-5m25-4r72
- GHSA-hp87-p4gw-j4gq

## **HAProxy Version**

We have updated HAProxy images to the following version:

- appscode/haproxy:2.6.2-alpine
- appscode/haproxy:2.5.8-alpine

## **Support for Kubernetes 1.25**

In this release we have updated Kubernetes client library dependencies to 1.25.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Voyager community, join us in the AppsCode Slack team channel `#general`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
