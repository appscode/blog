---
title: Announcing Voyager v2022.03.17
date: "2022-03-17"
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

We are pleased to announce the release of Voyager v2022.03.17. We have updated the HAProxy version to 2.5.5 in this release. The post highlights the import bug fixes in this release. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2022.03.17/README.md).

## **CVE Fixes**

We have updated the base image used for Voyager operator to address all known CVE reports by Trivy scanner.

## **HAProxy Support**

In this release we added support for alpine and debian based image for HAProxy 2.5.5 and 2.4.15. We have also added images with `major.minor-flavor` tags, so that users can stay up to date on the HAProxy image version.

``` sh
docker pull appscode/haproxy:2.5-alpine
docker pull appscode/haproxy:2.5-debian
docker pull appscode/haproxy:2.5.5-alpine
docker pull appscode/haproxy:2.5.5-debian

docker pull appscode/haproxy:2.4-alpine
docker pull appscode/haproxy:2.4-debian
docker pull appscode/haproxy:2.4.15-alpine
docker pull appscode/haproxy:2.4.15-debian
```

## **Fix HAProxy Template**

We have updated the HAProxy templates to replace use of `block` with `http-request deny`.

## **Mac ARM64 Support**

Voyager cli binary is now provided for Mac M1 arm64 architecture.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Voyager community, join us in the AppsCode Slack team channel `#general`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
