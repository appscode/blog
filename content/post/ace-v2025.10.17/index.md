---
title: Announcing ACE v2025.10.17
date: "2025-10-30"
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

We are pleased to announce the release of `ACE v2025.10.17`. ACE **v2025.10.17** focuses on improving scalability and automation for production-grade deployments. In this post, we’ll highlight the changes done in this release.

### Key Changes
- **Upgrade cluster-manager-spoke feature** on spoke upgrade
- **Add Credential-less Support** for AWS deployment.
- Fix backup session creation command

Here are the components specific changes:

### KubeDB UI
We are directly using the kubestash cli for generating the name of the kubestash-related objects to `fix backup session creation command`.


### Platform Backend

#### Fixes & Improvements
- The cluster-manager-spoke feature will also be upgraded on the spoke clusters from now on.
- We’ve introduced credential-less support for the Selfhost installer in production mode for AWS deployments. When the credential-less option is selected, users no longer need to provide AWS-specific credentials and can also import clusters in credential-less mode. All services within imported clusters will continue to operate seamlessly without requiring credentials.


### External products
Here is the summary of external dependency updates for `ACE v2025.10.17` :

- `KubeDB`: v2025.10.17 [Release blog](https://appscode.com/blog/post/kubedb-v2025.10.17/).
- `Stash`: v2025.10.17 [Release blog](https://appscode.com/blog/post/stash-v2025.10.17/).
- `KubeStash`: v2025.10.17 [Release blog](https://appscode.com/blog/post/kubestash-v2025.10.17/).
- `Voyager-gateway`: v2025.10.24 [Repo](https://github.com/voyagermesh/installer/tree/release-v2025.10.24).
- 
## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).
