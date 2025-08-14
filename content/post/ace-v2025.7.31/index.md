---
title: Announcing ACE v2025.7.31
date: "2025-07-31"
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

We are pleased to announce the release of `ACE v2025.7.31`, packed with major improvements. ACE **v2025.7.31** brings a bunch of new features & fixes. This release focuses on improving scalability and automation for production-grade deployments. In this post, we’ll highlight the changes done in this release.

## Key Changes
- **Add kubevirt cluster create and delete flow** with managed and self-managed option 
- **Site Admin** related settings(Ace upgrade, Allowed Domains, Branding, Client Organizations) is moved from `User Settings tab` to `Site Administration tab`
- **SelfSubjectRulesReview** permissions added for client-organizations in [KubeDB](https://appscode.com/blog/post/kubedb-v2025.7.31/)  

Here are the components specific changes:

## Cluster UI

### Fixes & Improvements
- `Fix editor changes` when importing a public cluster
- `Fix appName` when updated in branding
- `Allow to delete a cluster` if the cluster contains infraNamespace
- `fix featureSets cards css` in hub cluster
- `Add instructions` when adding namespace in hub permissions 


## KubeDB UI

### Fixes & Improvements
- `Fix SelfSubjectRulesReview permissions` when creating, deleting or updating a db in a client-organization

### Platform UI


#### Fixes & Improvements
- `Full name field` added when creating user from site-administration 
- `Admission permission` for deleting auto generated users are `restricted`
- Remove `permissions api` when viewing a public profile
- Update e2e tests form `claim and permissions settings`
- `Fix error messages disappearing` when creating a client-organization
- `Fix cookies issue` when switching to PLatform-ui from appDrawer
- `Update permissions` for adding a member in a organization using openfga 

### Platform Backend

#### Features
- `Add client-org context` in cluster & org permissions
- `Allow viewer to connect to cluster` and view kubeconfig

#### Fixes & Improvements
- `Update cluster deletion flow` to clean up cluster components and skip feature uninstallation 
- `Add CAPIProvider field` and `include Cluster_IsCAPICluster relation`
- `Add support for rancher-extension`
- `Enforce unique constraint on Email field`

### External products
Here is the summary of external dependency updates for `ACE v2025.7.31` :

- `KubeDB`: v2025.7.31 [Release blog](https://appscode.com/blog/post/kubedb-v2025.7.31/).
- `Stash`: v2025.7.31 [Release blog](https://appscode.com/blog/post/stash-v2025.7.31/).
- `KubeStash`: v2025.7.31 [Release blog](https://appscode.com/blog/post/kubestash-v2025.7.31/).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).