---
title: Announcing Voyager v2021.09.15
date: "2021-09-16"
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

We are pleased to announce the release of Voyager v2021.09.15. This post lists all the major changes done in this release since `v12.0.0`. This release offers support for the latest `Kubernetes version 1.22` and upgrades HAProxy to `2.4.4`. Voyager v2021.09.15 introduces Community & Enterprise Edition and deprecates prior releases of Voyager operators. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2021.09.15/README.md).

## **Kubernetes 1.22**

As you may know, [Kubernetes 1.22 removed several deprecated beta APIs](https://kubernetes.io/blog/2021/07/14/upcoming-changes-in-kubernetes-1-22/) that were used by Voyager. So, we have updated Voyager project to use the corresponding GA version of those APIs. With this change, Voyager v2021.09.15 supports Kubernetes 1.19 - 1.22 . With this change, we have removed support for `networking.k8s.io/v1beta1` Ingress api and only support `networking.k8s.io/v1` api.

## **Voyager v1 Ingress api**

In this release, we have introduced `voyager.appscode.com/v1` api group with changes similar to `networking.k8s.io/v1` api group. The documentation has been updated to only use the `v1` api group. The original `v1beta1` api group is still supported via a CRD conversion webhook built into the operator. So, you should be able to use your existing `v1beta1` YAMLs. But we recommend upgrading to `v1` YAML format.

## **HAProxy v2.4.4**

In this release, we have upgraded to HAProxy version to latest 2.4.4 release. We have also removed the prometheus exporter sidecar and use the HAPRoxy's built-in prometheus exporter for exposing metrics.

## **Certificate Management**

In this release, we have removed Voyager's built-in `Certificate` CRD. We now recommend using Jetstack's [cert-manager](https://cert-manager.io/) project for certificate management. Currently we only support DNS based domain validator for Let's Encrypt. HTTP01 based domain validation is not supported currently due to [cert-manager#4288](https://github.com/jetstack/cert-manager/issues/4288) limitation in cert-manager project.

## Voyager Operator uses `voyager` namespace

In previous releases, the Voyager documentation will show Voyager operator deployed in `kube-system` namespace. From this release, we have updated the documentation to use the `voyager` namespace for installing the operator. This is generally recommended to deploy operators in their own namespace instead of `kube-system` namespace. But you can still deploy Voyager is the `kube-system` namespace or any other namespace you like using Helm.

## Community Edition vs Enterprise Edition

Kubernetes v2021.09.15 Community Edition will manage ingress resources in Kubernetes `demo` namespace. Community Edition is feature limited and with this change we are making it clear to end users that Community Edition is primarily targeted for demo use-cases and the Enterprise Edition is targeted for production usage. To manage databases in any namespace, please try the Enterprise Edition.

## Deprecating Previous Voyager Releases

Our plans for Voyager has evolved quite a bit since [our decision to adopt an open/core model](https://appscode.com/blog/post/relicensing/) last year and provide a sustainable future for the project. With this release, we are announcing the deprecation of all prior Voyager releases. Currently Voyager v12.0.x and v11.0.x are available to users. The previous versions of Voyager operator has been retired and v12.0.x and v11.0.x will become unavailable by Dec 31, 2021. So, we encourage users to upgrade to the latest version of Voyager.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
