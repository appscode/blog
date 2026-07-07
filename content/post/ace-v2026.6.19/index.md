---
title: Announcing ACE v2026.6.19
date: "2026-06-30"
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
- platform-backend
- platform-ui
---

We are pleased to announce the release of `ACE v2026.6.19`.

ACE **v2026.6.19** brings a batch of new capabilities across the ACE interfaces, headlined by gateway configuration support in Cluster UI, a central monitoring system with Perses dashboards, expanded billing and contract management, and new database/OpsRequest support in KubeDB UI. In this post, we'll highlight the changes in this release.

### Key Changes
- **Cluster UI gains telemetry support & gateway configuration page**, including monitoring cluster imports and new feature-sets.
- **KubeDB UI adds Perses-based graphs and support for creating new databases and OpsRequests**.
- **Billing UI expands contract management** with cluster tagging, auto-assignment, and separate contact pages for viewers and admins. It also add **Local Billing** for offline mode.
- **Platform Backend introduces a central monitoring system** with a Perses UI.

Here are the components specific changes:

### Cluster UI
#### Enhancements
- Added a gateway configurations page.
- Added new options in the exposure block in kubedb-ui presets.
- Added a delete button in the hub-ui for clustersets.
- Added a telemetry stack, including importing monitoring clusters and new feature-sets for monitoring.
- Removed `helm_release_` from the cluster feature preview.

#### Fixes & Improvements
- Fixed a bug in the Cluster UI.

### Platform UI
#### Fixes & Improvements
- Fixed permission issues.
- Fixed the application settings show permission.

#### Enhancements
- Removed `k8sVersion` and `nodeCount` from client-org platform pages for client organizations.

### KubeDB UI
#### Fixes & Improvements
- Fixed the DB connection-card issue.
- Fixed the OpsRequest error-block.

#### Enhancements
- Added Perses graphs instead of Grafana.
- Added support for creating new databases and OpsRequests.

### Billing UI

#### Enhancements
- Added `AllowOffline` info display.
- Added a tagging system when assigning a cluster to a contract.
- Added auto-assign-cluster support when creating or modifying a contract.
- Added separate contact pages for viewers and admins.
- Added billing support when `onlineInstaller` is true.
- Added an event list table under the Admin licensed user product page.

### Platform Backend
#### Enhancements
- Added a central monitoring system.
- Added a Perses UI.
- Fixed the application settings show permission.

### External products
Here is the summary of external dependency updates for `ACE v2026.6.19`:

- `KubeDB`: v2026.6.19 [Release blog](https://appscode.com/blog/post/kubedb-v2026.6.19/).
- `KubeStash`: v2026.6.19 [Release blog](https://appscode.com/blog/post/kubestash-v2026.6.19/).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).
