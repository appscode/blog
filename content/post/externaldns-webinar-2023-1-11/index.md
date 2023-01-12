---
title: Manage ExternalDNS with CRD and Kubernetes Operator
date: "2023-01-12"
weight: 26
authors:
- Rasel Hossain
tags:
- azure
- cloudflare
- dns
- dns-provider
- external-dns
- external-dns-operator
- google
- operator
- route53
---

## Summary
AppsCode held a webinar on January 11, 2023, on **Manage ExternalDNS with CRD and Kubernetes Operator**. The contents of the webinar are:

1) Introduction to DNS
2) DNS in Kubernetes
3) External DNS Project
4) External DNS Operator
5) Demo
6) Q & A Session

## Description of the Webinar
At the beginning of the webinar, we talked about the `DNS` service and how it works. Then there was an overview of the `ExternalDNS` project by `Kubernetes`. The `External DNS` project dynamically syncs the exposed resources with DNS providers. In the webinar, we discussed how to use `External DNS` and why we need an operator.

Then we had an introduction to `External DNS Operator` and also there was an overview of the CRD fields and provider-specific secrets. The `External DNS Operator` can create and manage DNS records of cluster resources (`Node`, `Service`, `Ingress`). It requires a single CRD for a single set of configurations. Then we discussed the phases and workflow of the operator and how it watches the resources and updates the DNS record changes in the `provider`.

Later in this webinar, we demonstrate how to use this `operator` in different providers and how the records change when a resource record gets updated.

Take a deep dive into the full webinar below:

<iframe width="560" height="315" src="https://youtu.be/l96AJWBsnhc" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install the **ExternalDNS Operator**, please follow the installation instruction [here](https://github.com/kubeops/installer/tree/master/charts/external-dns-operator)
- You can check the GitHub public repo of **ExternalDNS Operator** from [here](https://github.com/kubeops/external-dns-operator)
- You can follow the official **ExternalDNS** project by `Kubernetes` from [here](https://github.com/kubernetes-sigs/external-dns)

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/kubedb).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
