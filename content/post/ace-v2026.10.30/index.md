---
title: Announcing ACE v2026.10.30
date: "2026-11-24"
weight: 16
authors:
- Topoti Sarkar
tags:
- billing-ui
- cloud-native
- cluster-ui
- database
- kubedb
- kubedb-ui
- kubernetes
- kubestash
- kubevault
- platform-backend
- platform-ui
---

We are pleased to announce the release of `ACE v2026.10.30`.

ACE **v2026.10.30** is a major UI release: all ACE interfaces move to `3.0.0` with a complete migration from Bulma to Tailwind on a new design. On the backend, background jobs move from NATS to Cadence, credentials are kept out of HelmRelease values, Argo CD (fargocd) based setups are much better supported, and a new selfhost server and AKP installer chart are introduced. In this post, we'll highlight the changes in this release.

### Key Changes
- **New look across all UIs**: Cluster UI, KubeDB UI, Platform UI and Billing UI are migrated from Bulma to Tailwind with the new design system, with paged tables, restyled cards and navbar tooltips.
- **Telemetry retention per tenant**: client organizations can now pick their Thanos / VictoriaLogs retention from the TelemetryStack, and the TelemetryStack form supports per-retention storage size.
- **KubeDB UI**: new `RotateAuth` OpsRequest, BackupStorage selection for backup and archiver, and custom `MSSQL_PID` product key stored in a Secret.
- **Cadence-based workflows**: background jobs now run on Cadence instead of NATS.
- **Safer credential handling**: opscenter-features, service-backend and gateway credentials are now stored in Secrets instead of being written into HelmRelease values.
- **Better Argo CD (fargocd) support**, including OpenShift.
- **New AKP installer chart and OLM bundle**, supporting OpenShift 4.18 to 5.0.

Here are the components specific changes:

### Cluster UI
#### Enhancements
- Integrated TelemetryStack into Cluster UI, with support for editing a TelemetryStack.
- Added an overview page for Telemetry.
- Tenant retention can now be chosen from the TelemetryStack list.
- Shows a single metrics retention field for VictoriaMetrics clusters.
- Added a separate cluster profile page.
- Added a "configure DB versions" button for feature-sets.
- Added an edit button for hub KubeDB UI presets.
- Fixed the Central Argo CD picker.
- The current page is kept when switching clusters.

#### Fixes & Improvements
- Long labels and annotations no longer widen resource tables.

### KubeDB UI
#### Enhancements
- Datastore groups can now be collapsed.

#### Fixes & Improvements
- Fixed the instant backup deploy issue.
- Fixed link updates and scrolling in forms.

### Platform UI
#### Enhancements
- Added site admin controls for organization-based licensing.
- Client organizations moved from Site Administration to Settings.
- The "Create New Org" button on the overview page now respects permissions.

### Billing UI
#### Enhancements
- Shows the cluster name on the licensed clusters page.
- Added a contract-level license passthrough toggle.
- The issue license button is shown for all contract clusters.
- Latest licenses of a contract cluster are shown in a modal.
- Added copy buttons to tables.

### UI Wizards
#### Enhancements
- Added `RotateAuth` to OpsRequest editors.
- Allowed selecting a BackupStorage for backup and archiver.
- Allowed toggling BackupStorage selection in the datastore feature-set editor.
- Added per-retention storage size to the TelemetryStack form.
- Added `config-syncer-manager` to the ocm-hub feature-set editor.
- Custom `MSSQL_PID` product keys are now stored in a Secret.

#### Fixes & Improvements
- Switching the OpsRequest type now clears fields left over from the previous type.

### Platform & Installation
#### Enhancements
- Background jobs moved from NATS to Cadence.
- Added a selfhost server.
- Gateway DNS provider credentials and the JKS keystore are now also stored in Secrets.
- Honors the managed Argo CD cluster name and namespace.
- Added OpenShift support for Argo CD (fargocd) based setups.
- Added the AKP installer chart and OLM bundle, with Cadence workflow support; the bundle supports OpenShift 4.18 to 5.0.
- Added the `config-syncer-manager` OCM addon feature.
- Added an ACE removal script for Flux and fargocd (Argo CD) setups, and sped up ACE removal.
- Added pgpool settings to the ACE installer options.
- Gateway API versions are picked from what the cluster serves.

#### Fixes & Improvements
- Fixed spoke cluster removal.
- Stopped Perses sync from duplicating teams and members.
- Updated the sign-in/sign-up and activation UI.
- Fixed ACE sync under Argo CD and Argo CD mode cleanup.
- Fixed the Cadence schema version resetting to `0.0` on ACE chart upgrades.
- Gateway certificates now work with domains longer than 64 bytes.
- Uninstalling ACE no longer deletes the gateway certificate.

### External products
Here is the summary of external dependency updates for `ACE v2026.10.30`:

- `KubeDB`: v2026.11.15 [Release blog](https://appscode.com/blog/post/kubedb-v2026.11.15/).
- `KubeStash`: v2026.11.15 [Release blog](https://appscode.com/blog/post/kubestash-v2026.11.15/).
- `KubeVault`: v2026.11.15 [Release blog](https://appscode.com/blog/post/kubevault-v2026.11.15/).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).
