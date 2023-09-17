---
title: Announcing Voyager v2023.9.18
date: "2023-09-18"
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

We are pleased to announce the release of Voyager v2023.9.18. In this release, we have updated HAProxy image to 2.6.15 and fixed various CVEs. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2023.9.18/README.md).

## **Custom SSL Ciphers**

In this release, we have added an option to customoize cipher suites used by the HAProxy pods. To customize this value, set the annotation `ingress.appscode.com/ssl-ciphers` and it will be passed as `ssl-default-bind-ciphers` to the HAProxy configuration. The relevant pr can be found here: https://github.com/voyagermesh/haproxy-templates/commit/d1fd1e37eb575a0df982d300f13657c7b7e47d8c 

## **CVE Fixes**

We have updated the docker images in this release to address the following CVEs:

- CVE-2023-2650
- CVE-2023-2975
- CVE-2023-3446
- CVE-2023-3817
- CVE-2023-2650
- CVE-2023-2975
- CVE-2023-3446
- CVE-2023-3817
- CVE-2023-2650
- CVE-2023-2975
- CVE-2023-3446
- CVE-2023-3817

## **HAProxy Version**

We have updated HAProxy images to the following version:

- ghcr.io/voyagermesh/haproxy:2.8.3-alpine
- ghcr.io/voyagermesh/haproxy:2.7.10-alpine
- ghcr.io/voyagermesh/haproxy:2.6.15-alpine

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Voyager community, join us in the AppsCode Slack team channel `#general`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
