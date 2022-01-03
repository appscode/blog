---
title: Announcing Voyager v2022.01.01
date: 2022-01-01
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

We are pleased to announce the release of Voyager v2022.01.01. We have updated the HAProxy version to 2.5.0 in this release. The post highlights the import bug fixes in this release. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2022.01.01/README.md).

## **HAProxy Support**

In this release we added support for alpine and debian based image for HAProxy 2.5.0 and 2.4.10. We have also added images with `major.minor-flavor` tags, so that users can stay up to date on the HAProxy image version.

``` sh
docker pull appscode/haproxy:2.5-alpine
docker pull appscode/haproxy:2.5-debian
docker pull appscode/haproxy:2.5.0-alpine
docker pull appscode/haproxy:2.5.0-debian

docker pull appscode/haproxy:2.4-alpine
docker pull appscode/haproxy:2.4-debian
docker pull appscode/haproxy:2.4.10-alpine
docker pull appscode/haproxy:2.4.10-debian
```

## **Security Context Support HAProxy containers**

In this release we have added a `spec.proxySecurityContext` field to the Ingress CRD of Voyager. This will allow users to set security context for the HAProxy container.

## **Improved logging for HAProxy reloads**

In this release we have added detailed logging when a config reload happens to HAProxy. This information is logged to the coordinator sidecar container. This can help indentify excessive reload of HAProxy configuration.

## Deprecating Previous Voyager Releases

This is just a quick reminder that Voyager v12.0.x & v11.0.x releases have been deprecated since v2021.09.15 and will become unavailable by Dec 31, 2021. So, we encourage users to upgrade to the latest version of Voyager.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Voyager community, join us in the AppsCode Slack team channel `#general`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
