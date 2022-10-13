---
title: Announcing Voyager v2021.10.18
date: "2021-10-18"
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

We are pleased to announce the release of Voyager v2021.10.18. This release is a patch release for `v2021.09.15`. The post highlights the import bug fixes in this release. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2021.10.18/README.md).

## **Voyager v1/v1beta1 Ingress api conversion**

In this release, we have fixed a panic that would occur when converting v1beta1 Ingress with multiple tls secrets in v1 api format.

In `v1` api, we have removed the deprecated `headerRules` and `rewriteRules` from `v1beta1` api. In this release, the api version converter has been updated to automatically convert `headerRules` and `rewriteRules` into `backendRules`. You can see how these rules are converted [here](https://github.com/voyagermesh/apimachinery/blob/v0.1.3/apis/voyager/v1beta1/conversion.go#L125-L131).

## **HAProxy Templates**

The GO templates used to generate HAProxy 2.x configuration are now open source. You can find them here: https://github.com/voyagermesh/haproxy-templates

## **Separate Operator and Webhook Server Deployment**

In this release, we have split the Voyager operator and webhook server into two separate deployments. This change will be applied when you `helm upgrade` to this release.

## Deprecating Previous Voyager Releases

This is just a quick reminder that Voyager v13.0.x, v12.0.x & v11.0.x releases have been deprecated since v2021.09.15 and will become unavailable by Dec 31, 2021. So, we encourage users to upgrade to the latest version of Voyager.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Voyager community, join us in the AppsCode Slack team channel `#general`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
