---
title: "KubeDB AutoOps: Automate Day 2 Life-cycle Management for Databases on Kubernetes"
date: 2022-04-18
weight: 26
authors:
  - Pulak Kanti Bhowmick
tags:
  - kubedb
  - autoops
  - supervisor
  - recommendation
  - maintenance-window
  - approval-policy
  - rotote-tls
  - version-upgrade
---

## Summary
On April 13, 2022, AppsCode held a webinar on **KubeDB AutoOps: Automate Day 2 Life-cycle Management for Databases on Kubernetes**. The key contents of the webinars are:
1) Introduction to KubeDB AutoOps
2) KubeDB AutoOps Components:
    * Recommendation Generator
    * Supervisor
3) Demo
4) Q & A Session

## Description of the Webinar
At first, we gave an overview of `KubeDB AutoOps`. AutoOps is an addition of technology or concept that performs tasks automatically or with minimal human assistance. Similarly, `KubeDB AutoOps` is a concept that is focusing to automate the day 2 life-cycle management for databases on Kubernetes. Currently, KubeDB manages the day 2 life-cycle for a Database using `OpsRequest`. But these processes are manual and users have to create those OpsRequest with immediate effect. By using `KubeDB AutoOps`, these processes can be fully automated & configurable.

To achieve this, we introduced two KubeDB AutoOps components. One is `Recommendation Generator` which will generate `Recommendation` by inspecting KubeDB Database resources. And another one is `Supervisor` which will execute those recommendations in the specified `Maintenance Window`.

Later in this webinar, we demonstrated how `Recommendation Generator` automatically generated `Rotate TLS` & `Version Upgrade` recommendations for the KubeDB Databases. We also showed how to execute those recommendations in a specific `Maintenance Window` by `Supervisor` with minimal human assistance.


Take a deep dive into the full webinar below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/-TiyOS1QbhI" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.03.28/welcome/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/kubedb).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
