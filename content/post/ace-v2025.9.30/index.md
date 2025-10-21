---
title: Announcing ACE v2025.9.30
date: "2025-10-10"
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

We are pleased to announce the release of `ACE v2025.9.30`. ACE **v2025.9.30** focuses on improving scalability and automation for production-grade deployments. In this post, we’ll highlight the changes done in this release.

### Key Changes
- **Improved handling of billing usage values** by ensuring consistent rounding and normalizing the fractional part.
- Skip name from license and **reuse existing licenses only if valid**
- **Fix featureExclusionGroup issue**

Here are the components specific changes:

### Billing UI

#### Fixes & Improvements
Improved handling of billing usage values by ensuring consistent rounding and normalizing the fractional part. Previously, usage values were displayed as 0.0000 even when the fractional part was zero. Now, these values are normalized for a cleaner and more consistent display — for example:

* `0.0001` → remains `0.0001`
* `0.0100` → shown as `0.01`
* `0.1000` → shown as `0.1`
* `0.0000` → shown as `0`

### Platform Backend

#### Fixes & Improvements
- Skip name from license and **reuse existing licenses only if valid**. This fixes the "license status unknown, reason: x509: SAN rfc822Name is malformed" error.
> ⚠️ **Note:** If you still face this issue on ACE versions >= v2025.9.15, then you might need to restart the license-proxyserver pod. `kubectl delete pod -n kubeops -l app.kubernetes.io/instance=license-proxyserver` to fetch updated non-malformed license.

- **Fix featureExclusionGroup issue** on cluster import.

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).
