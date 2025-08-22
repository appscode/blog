---
title: Announcing ACE v2025.8.15
date: "2025-08-22"
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

We are pleased to announce the release of `ACE v2025.8.15`, packed with major improvements. ACE **v2025.8.15** brings a bunch of new features & fixes. This release focuses on improving scalability and automation for production-grade deployments. In this post, we’ll highlight the changes done in this release.

## Key Changes
- **Rancher Extension**: KubeDB is now available as an extension in rancher UI.
- **Billing improvements**: Improve the billing ui.
- **GitOps support**: Add support to deploy gitops-enabled databases from ui.


Here are the components specific changes:


## KubeDB UI

### Features
- Add support to deploy `gitops-enabled databases` from ui. You can't do any operations as it is synced from outer source.


### Platform UI

### Features
- Added `safety check on clientOrg deletion`, if it is in use.
- Shared Gateway Configs will only be created on ace-gw namespace


### Billing UI

#### Features
- Improve the billing ui.


### Selfhost UI
#### Features
- Users can now specify some `annotations for their ingress LB` to work.


### Platform Backend

#### Fixes & Improvements
- KubeDB is now available as an `extension in rancher UI`.

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).