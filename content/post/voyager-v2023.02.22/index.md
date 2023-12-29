---
title: Announcing Voyager v2023.02.22
date: "2023-02-22"
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

We are pleased to announce the release of Voyager v2023.02.22. In this release, we have updated HAProxy image to 2.6.9 and fixed various bugs. The detailed commit by commit changelog can be found [here](https://github.com/voyagermesh/CHANGELOG/blob/master/releases/v2023.02.22/README.md).

## **CVE Fixes**

We have updated the docker images in this release to address the following CVEs:

- CVE-2023-0286
- CVE-2023-0215
- CVE-2022-4450

## **HAProxy Version**

We have updated HAProxy images to the following version:

- appscode/haproxy:2.7.3-alpine
- appscode/haproxy:2.6.9-alpine
- appscode/haproxy:2.5.12-alpine

## **HAProxy Server State**

We have updated the HAProxy configuration templates to stop storing server state in a file. As HAProxy pods restart this state is lost as data is stores in an `emptyDir`. So, this is not stores anymore and avoids related error logs from HAProxy. You can learn more from here: https://www.haproxy.com/documentation/hapee/latest/api/runtime-api/show-servers-state/

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Voyager, please follow the installation instruction from [here](https://voyagermesh.com/docs/latest/setup).

* If you want to upgrade Voyager from a previous version, please follow the upgrade instruction from [here](https://voyagermesh.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Voyagermesh).

If you have found a bug with Voyager or want to request for new features, please [file an issue](https://github.com/voyagermesh/project/issues/new).
