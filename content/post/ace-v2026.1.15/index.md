---
title: Announcing ACE v2026.1.15
date: "2026-1-30"
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

We are pleased to announce the release of `ACE v2026.1.15`. ACE **v2026.1.15** focuses on improving scalability and automation for production-grade deployments. This release have introduced so many important features. In this post, we’ll highlight the changes done in this release.

### Key Changes
- **A new form builder is added in kubeDB ui** which makes DB related forms faster and user-friendly.
- **All security vulnerabilities are resolved in npm packages**.
- **Added foundational metrics (ROUTES, NATS, DB, OPENFGA and more) to improve observability**.
- Clusters are now **automatically assigned to active contracts** when a free contract expires
- Added CA certificate support during BlobFS bucket client initialization to **resolve secure connection issues** in on-prem environments.

Here are the components specific changes:

### Kubedb UI
#### Enhancements
- A new form builder is added in kubeDB ui which makes DB related forms more faster and user friendly
- All security vulnerabilities are resolved in npm packages
- New design for DB related actions and backup
- DB related Actions can be accessed through new intuitive sidebar in DB details page


### Cluster UI
##### Enhancements
- Going to platform from cluster will now open in current tab instead of new tab as users requested
- All security vulnerabilities are resolved in npm packages
- New version is now notified in cluster details page and user can go there directly from notification

#### Fixes
- Create credential from organization was taking user to their user profile which is now fixed
- In kubevirt when user go to import cluster there was a validation error which stopped user from importing that was fixed
- A glitch from an alert message was also removed from cluster details page
- All security vulnerabilities are resolved in npm packages

### Platform UI
#### Enhancements
- All security vulnerabilities are resolved in npm packages
- Organization creation level restriction added for client organization’s member
- Going to cluster from platform will now open in current tab instead of new tab as users requested
- Form styles were improved

#### Fixes
- User account creation had some validation error which are fixed in this version
- False toast error from organization page resolved
- Rancher validation issue is fixed

### Billing UI
#### Enhancements
- All security vulnerabilities are resolved in npm packages
- Admin can copy the license of user from the contract page


### Platform Backend

#### Enhancements
- **Monitoring for Core ACE Components** :
Added foundational metrics (ROUTES, NATS, DB, OPENFGA and more) to improve observability and operational visibility in production environments.
- **Automate Cluster Assignment in Contracts** :
Clusters are now automatically assigned to active contracts when a free contract expires.


#### Fixes
- **On-Prem S3 Proxy (Demo Mode)** :
Added CA certificate support during BlobFS bucket client initialization to resolve secure connection issues in on-prem environments.
- **Upgrader Job** :
i) Fixed platform-api container image selection during upgrader job creation.
ii) Fixed platform-api container volumeMount selection during upgrader job creation.
- **Installer & Cluster Import** :
i) Fixed unmarshal error during installer creation in v2026.1.15 version.
ii) Automatically set kubestash.storageSecret.create=false when importing clusters in credential-less mode.

### External products
Here is the summary of external dependency updates for `ACE v2026.1.15` :

- `KubeDB`: v2026.1.19 [Release blog](https://appscode.com/blog/post/kubedb-v2026.1.19/).
- `KubeStash`: v2026.1.19 [Release blog](https://appscode.com/blog/post/kubestash-v2026.1.19/).
- `Voyager-gateway`: v2026.1.15 [Repo](https://github.com/voyagermesh/installer/tree/release-v2026.1.15).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).
