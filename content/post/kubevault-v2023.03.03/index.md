---
title: Introducing KubeVault v2023.03.03
date: "2023-03-06"
weight: 25
authors:
- Abdullah Al Shaad
tags:
- grafana
- hashicorp
- kubernetes
- kubevault
- monitoring
- prometheus
- secret-management
- security
- vault
---

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines' management, policy management in the Kubernetes native way.

We are very excited to announce the release of [KubeVault v2023.03.03](https://kubevault.com/docs/v2023.03.03/setup/) Edition. In this release, we have added grafana dashboard and alerting for Vault servers.

In this post, we are going to highlight the major changes. You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2023.03.03/README.md).


### Monitoring

This release includes a Grafana dashboard for easier monitoring of KubeVault managed Vault servers. The grafana dashboard shows several vault server specific data, status and diagram of memory and cpu consumption.
You can check the dashboard to see the overall health of your vault server easily.

To learn more about it go to the [link](https://github.com/appscode/grafana-dashboards/tree/master/kubevault)

### Alert

We have added configurable alerting support for KubeVault. You can configure Alert-manager to get notified when a metrics of vault server exceeds a given threshold.

To learn more, have a look [here](https://github.com/appscode/alerts/tree/master/charts/vaultserver)

## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2023.03.03/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).