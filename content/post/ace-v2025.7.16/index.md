---
title: Announcing ACE v2025.7.16
date: "2025-07-21"
weight: 16
authors:
- Arnob Kumar Saha
tags:
- billing-ui
- cloud-native
- cluster-ui
- database
- kubedb
- kubedb-ui
- kubernetes
- kubestash
- platform-ui
- voyager-gateway
---

We are pleased to announce the release of `ACE v2025.7.16`, packed with major improvements. ACE **v2025.7.16** brings a bunch of new features & fixes. This release focuses on improving scalability and automation for production-grade deployments. In this post, we’ll highlight the changes done in this release.

## Key Changes
- **EKS with private endpoint**: Support added for deploying ACE on an EKS cluster with private endpoint.
- **Domain redirection**: Fix domain redirection issue in all the ui components.
- **BackupConfig creation**: Backupconfig creation issue has been resolved both on db-specific pages, & common page.

Here are the components specific changes:

## Cluster UI

### Fixes
- Sort database names in the "Datastore" feature, to avoid unnecessary HelmRelease patching.


## KubeDB UI

### Features

#### MSSQLServer
- Add `Evaluation` mode in mssql.

### Fixes

#### Common
- `Fix backupConfig creation` from both db-specific pages, & common page.

#### Elasticsearch
- `Fix showing es configSecert` in preview page

#### MSSQLServer
- `Fix exec` into mssql containers.

#### Redis
- Add validation for the number of shards for the `announce` field in Redis.
- `Fix horizontal scaling` for sentinel enabled redis.



### Platform UI


#### Fixes
- `Fix Team creation` with type Basic.
- Correctly show the `spinning in version-upgrade`.


### Billing UI

#### Features
- `Redesign total usage` data section


### Selfhost UI
#### Features
- Support added for deploying ACE on an `EKS cluster with private endpoint`. See this [video](https://www.youtube.com/watch?v=C1cC1qghl-g&list=PLoiT1Gv2KR1iqWFGkCozbLqYe31QMsQcX&index=3) for how to setup that.


### Platform Backend

#### Fixes & Improvements
- Ignore hub featureset when ace deployed on EKS cluster.
- `Fix installer generation for "Cloud Demo"` mode with dns enabled.


### External products
Here is the summary of external dependency updates for `ACE v2025.7.16` :

- `OCM`: v1.0.0 [Release note](https://github.com/open-cluster-management-io/ocm/releases/tag/v1.0.0).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).