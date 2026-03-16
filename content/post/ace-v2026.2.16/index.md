---
title: Announcing ACE v2026.2.16
date: "2026-03-16"
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
---

We are pleased to announce the release of `ACE v2026.2.16`.

ACE **v2026.2.16** focuses on improving observability, billing workflows, and everyday usability across the platform. In this post, we'll highlight the changes done in this release.

### Key Changes
- **Billing usage projection and auto-assign workflows** make contract management easier for operators.
- **Observability is improved** with billing pod metrics endpoints and service-mode-aware dashboard panels.
- **Cluster and database workflows are more streamlined** with preset form-builder updates, backup editing, and cleaner in-app navigation.
- **Credential-less and EKS flows are improved** in the platform backend for smoother self-hosted operations.

Here are the components specific changes:

### Billing UI
#### Enhancements
- Added an **Auto Assign Cluster** action to simplify assigning clusters to contracts.
- Added a **lighter and faster editor setup** for a smoother editing experience.
- Updated copy actions and labels for a cleaner contract and customer workflow.
- Resolved dependency and security issues.

### Cluster UI
#### Enhancements
- Updated **presets design** for a faster UI experience.
- Updated create links to take users directly to **KubeDB UI** where appropriate.
- Opening **Platform UI** from Cluster UI now happens in the **current tab**.
- Improved alert messaging behavior after required conditions are met.
- Resolved dependency and security issues.

#### Fixes
- Fixed store-related issues.
- Fixed reactivity issues in the **KubeVirt** cluster creation flow.

### KubeDB UI
#### Enhancements
- Added an **Edit** action on the backup page for easier backup management.
- Improved **project quota** visibility when only the group or partial CPU/memory limits are set.
- Removed unnecessary new-tab behavior in related flows for a more consistent navigation experience.
- Resolved dependency and security issues.

### Platform UI
#### Enhancements
- Cluster cards now open related destinations in the **current tab** instead of a new one.
- Website URLs are normalized by automatically adding `https` when needed.
- Resolved dependency and security issues.

### Platform Backend
#### Enhancements
- Exposed **metrics endpoints for billing pods** and improved dashboards so different **service modes** of B3 can be visualized separately.
- Added **automated email alerts for expiring contracts** for operational follow-up.
- Added **end-of-month billable usage projection** support.
- Added **credential-less support** to the virtual secret server.
- Improved **EKS support** with OCM-related updates for cluster flows.
- Clusters can now still be **auto-assigned to active contracts** after free contracts expire, with matching UI support.

#### Fixes & Improvements
- Skipped installing **KubeStash** from the ACE chart where appropriate.
- Enabled **DS metrics** by default.
- Fixed the **License Bucket** setting.
- Set `infra.tenantSpreadPolicy` to `single` by default when not provided.
- Added missing registries and included dependency and CVE fixes.
- Converted ingress handling to **Gateway**.

### External products
Here is the summary of external dependency updates for `ACE v2026.2.16` :

- `KubeDB`: v2026.2.26 [Release blog](https://appscode.com/blog/post/kubedb-v2026.2.26/).
- `KubeStash`: v2026.2.26 [Release blog](https://appscode.com/blog/post/kubestash-v2026.2.26/).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).
