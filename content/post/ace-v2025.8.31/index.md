---
title: Announcing ACE v2025.8.31
date: "2025-09-11"
weight: 16
authors:
- Nihal Azmain
tags:
- billing-ui
- cloud-native
- cluster-ui
- database
- kubedb
- kubedb-ui
- kubernetes
- kubestashi
- platform-ui
- voyager-gateway
---

We are pleased to announce the release of `ACE v2025.8.31`, packed with major improvements. ACE **v2025.8.31** brings a bunch of new features & fixes. This release focuses on improving scalability and automation for production-grade deployments. In this post, we’ll highlight the changes done in this release.

### Key Changes
- **Site Admin** can now disable registration for new account. 

Here are the components specific changes:

### Cluster UI

### Fixes & Improvements
- `Fix namespace issue for rancher clusters` when creating project quota.
- `Fix loader` when upgrading ace version.

### KubeDB UI

#### Features
- Add cluster level restrictions for `create, delete and update` operations.

#### Fixes & Improvements
- `Add Configuration options` for Clickhouse keeper for both externally and internally managed.
- `Fix securityContext issue` in ferretdb pg-backend .

### Billing UI

#### Features
- UI changes to show granular information about `trail usage, PROD & NON-PROD usage`. Please visit the [Billing & usage doc](https://appscode.com/docs/en/guides/billing-and-usage-guide/overview.html) for more details.

### Platform UI

#### Features
- You can now download `ca.cert` from our platform in the `Rancher Proxy` page which you will find in `organization settings` 
- `Site Admin` can now disbale registration for new account. 

#### Fixes & Improvements
- Fix `create credential` for `AzureStorage, Scaleway, Swift and Vultr`
- Switching `App` from `Appdrawer` will now be switched in the same tab.

### Platform Backend

#### Fixes & Improvements
- Fix `aceproxy` installation command
- `Each database will get it's first month as free trial usage` if the database namespace is explicitly annotated to allow for trail usage. Please visit the [Billing & usage doc](https://appscode.com/docs/en/guides/billing-and-usage-guide/overview.html) for more details.

### External products
Here is the summary of external dependency updates for `ACE v2025.8.31` :

- `KubeDB`: v2025.8.31 [Release blog](https://appscode.com/blog/post/kubedb-v2025.8.31/).
- `Stash`: v2025.7.31 [Release blog](https://appscode.com/blog/post/stash-v2025.7.31/).
- `KubeStash`: v2025.7.31 [Release blog](https://appscode.com/blog/post/kubestash-v2025.7.31/).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).