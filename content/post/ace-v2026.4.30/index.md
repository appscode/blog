---
title: Announcing ACE v2026.4.30
date: "2026-04-30"
weight: 16
authors:
- Samiul Islam
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
- selfhost-ui
---

We are pleased to announce the release of `ACE v2026.4.30`.

ACE **v2026.4.30** is a smaller stabilization release focused on installer usability, member-management improvements, feature-set workflow fixes, and dependency updates across the ACE interfaces. In this post, we'll highlight the changes in this release.

### Key Changes
- **Selfhost installer workflows are improved** with better Kube API Server handling and DNS-related fixes.
- **Platform UI improves organization management** with member actions and member counts.
- **Cluster feature-set flows are more reliable** with fixes for spoke configuration and enable/configure behavior.
- **ACE interfaces received dependency and security updates** across the `v2.2.0` UI releases and the latest Selfhost UI updates.

Here are the components specific changes:

### Billing UI
#### Fixes & Improvements
- Updated Billing UI dependencies and resolved security issues as part of the `v2.2.0` release line.

### Cluster UI
#### Fixes & Improvements
- Fixed feature-set issues in spoke configuration flows.
- Resolved enable/configure confusion in feature-set workflows to make cluster settings behavior more predictable.
- Updated Cluster UI dependencies and resolved security issues as part of the `v2.2.0` release line.

### KubeDB UI
#### Fixes & Improvements
- Updated KubeDB UI dependencies and resolved security issues as part of the `v2.2.0` release line.

### Platform UI
#### Enhancements
- Added member count visibility in the Site Administration organization list.
- Added a remove-member action in organization member management flows.

#### Fixes & Improvements
- Updated Platform UI dependencies and resolved security issues as part of the `v2.2.0` release line.

### Selfhost UI
#### Enhancements
- Improved installer forms by making the **Kube API Server** field more generic and removing overly strict validation.

#### Fixes & Improvements
- Fixed DNS-related issues in Selfhost installer workflows.

### Platform Backend
#### Fixes & Improvements
- Synced the latest `lib-selfhost` changes into the backend to keep installer-driven workflows aligned with the current self-hosted release behavior.
- Updated bundled wizard and resource metadata for newer database and editor definitions, including **DocumentDB**, **DB2**, **HanaDB**, **Milvus**, **Neo4j**, **Qdrant**, and **Weaviate**.
- Added backend test coverage around ACE installer and upgrader flows while refreshing dependencies.

### External products
Here is the summary of external dependency updates for `ACE v2026.4.30`:

- `KubeDB`: v2026.4.27 [Release blog](https://appscode.com/blog/post/kubedb-v2026.4.27/).
- `KubeStash`: v2026.4.27 [Release blog](https://appscode.com/blog/post/kubestash-v2026.4.27/).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).
