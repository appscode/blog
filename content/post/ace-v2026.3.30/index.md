---
title: Announcing ACE v2026.3.30
date: "2026-03-30"
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

We are pleased to announce the release of `ACE v2026.3.30`.

ACE **v2026.3.30** brings improvements to installer configuration, site administration, and form-driven workflows across the platform. This release also includes several backend and upgrade-related improvements to make cluster onboarding, routing, and operational management more reliable. In this post, we'll highlight the changes done in this release.

### Key Changes
- **Alertmanager configuration is now exposed through ACE installer options** with email and webhook routing support.
- **Site Administration is improved** with a dedicated cluster inventory page and better orphan organization management flows.
- **Feature set and preset workflows are improved** in Cluster UI with dedicated pages and manifest preview support.
- **Installer and backend flows are more reliable** with safer `user-config` handling, better credential-less EKS onboarding, and safer upgrade handling for backup components.

Here are the components specific changes:

### Billing UI
#### Enhancements
- Added an **Auto Assign Cluster** toggle in contract cluster settings to manage the existing auto-assign workflow more easily for eligible users.
- Improved deep-link handling with `redirect_to` support so billing links can take users directly to the intended page.
- Simplified the **Add Cluster** flow with a lighter editor and inline error handling.

#### Fixes & Improvements
- Fixed routing issues when switching usage type or product in usage views.
- Updated customer-facing copy in contract link actions.

### Cluster UI
#### Enhancements
- Moved feature-set configuration from modal dialogs to dedicated pages across cluster, hub, and spoke flows.
- Added generated manifest preview and edit support before deployment in feature-set workflows.
- Extended the same guided flow to preset editing.

#### Fixes & Improvements
- Improved public cluster import routing to support cluster names that contain slashes.
- Creating KubeDB resources from Cluster UI now redirects users to the dedicated **KubeDB UI** flow.
- Reduced noisy upgrade prompts by waiting for version data before showing new-version notices.

### KubeDB UI
#### Enhancements
- Improved project quota visibility in the initial database create flow, including cases where only CPU or memory limits are partially set.
- Added an **Edit** action for `kubestash.com` resources to make related edit flows easier to reach.
- Updated several related navigation flows to stay in the same tab.

#### Fixes & Improvements
- Clarified quota charts when no CPU or memory limit is set instead of showing misleading empty states.
- Hid the snapshot create button where it should not be exposed.
- Added earlier validation for overly long resource names in create flows.

### Platform UI
#### Enhancements
- Added a dedicated **Clusters** page in Site Administration with search and key cluster metadata.
- Added visibility for likely orphaned organizations in the Site Administration organization list and direct delete action from the table.
- Improved URL handling by automatically adding `https` where needed.

#### Fixes & Improvements
- Cluster navigation now stays in the same tab instead of opening a new tab unexpectedly.

### Platform Backend
#### Enhancements
- Exposed **Alertmanager** email and webhook routing through ACE installer options.
- Added support for Gateway API based service presets in installer-driven deployments.
- Improved credential-less **EKS** onboarding with better IAM setup and smoother OCM registration behavior.
- Added APIs to list and delete orphan organizations.
- Implemented **get-update API pagination** for update workflows.
- Added an admin API to list cluster information for site administration workflows.
- Updated scripts to support running outbox workflows on macOS.

#### Fixes & Improvements
- Preserved existing platform `user-config` secrets instead of overwriting them during install or upgrade.
- Added safer handling for existing **KubeStash** components during ACE upgrades.
- Made `storageClass` optional in ACE options schema.
- Improved billing event handling and contract reminder timing in backend flows.
- Updated observability and networking dependencies including **Envoy**, **KEDA**, **ingress-nginx**, **grafana-tools**, and **Voyager**.
- Improved OpenShift monitoring deployment behavior and fixed RBAC for license proxyserver metrics.

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).
