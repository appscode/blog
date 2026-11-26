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

ACE **v2026.10.30** is a major UI release: all ACE interfaces move to `3.0.0` with a complete migration from Bulma to Tailwind on a new design system. On the backend, background jobs move from NATS to Cadence, credentials are kept out of HelmRelease values, Argo CD (fargocd) based setups are much better supported, and a new selfhost server and AKP installer chart are introduced. In this post, we'll highlight the changes in this release.

### Key Changes
- **New look across all UIs**: Cluster UI, KubeDB UI, Platform UI and Billing UI are migrated from Bulma to Tailwind with the new design system, with paged tables, restyled cards and navbar tooltips.
- **Telemetry retention per tenant**: client organizations can now pick their Thanos / VictoriaLogs retention from the TelemetryStack, and the TelemetryStack form supports per-retention storage size.
- **KubeDB UI wizards**: new `RotateAuth` OpsRequest, BackupStorage selection for backup and archiver, and custom `MSSQL_PID` product key stored in a Secret.
- **Cadence-based workflows**: NATS jobs are migrated to Cadence.
- **Safer credential handling**: opscenter-features, service-backend and gateway credentials are moved into Secrets referenced via `valuesFrom`, instead of being written into HelmRelease `spec.values`.
- **Better Argo CD (fargocd) support**, including OpenShift.
- **New AKP installer chart and OLM bundle**, supporting OpenShift 4.18 to 5.0.

Here are the components specific changes:

### Cluster UI
#### Enhancements
- Migrated from Bulma to Tailwind and updated to design system `2.13.0`.
- Integrated TelemetryStack into Cluster UI, with support for editing a TelemetryStack.
- Added an overview page for Telemetry.
- Tenant retention can now be chosen from the TelemetryStack list.
- Shows a single metrics retention field for VictoriaMetrics clusters.
- Added a separate cluster profile page.
- Added a "configure DB versions" button for feature-sets.
- Added an edit button for hub KubeDB UI presets.
- Sends the restructured Argo CD options, and fixed the Central Argo CD picker.
- Tables are now paginated.
- Added tooltips to navbar icons.
- The current page is kept when switching clusters, and the navbar user type is updated.

#### Fixes & Improvements
- Long labels and annotations no longer widen resource tables.

### KubeDB UI
#### Enhancements
- Migrated from Bulma to Tailwind and updated to design system `2.13.0`.
- Datastore groups can now be collapsed.
- Tables are now paginated, with a full-height section nav.
- Added tooltips to navbar icons.
- Added a loader for the preview button, and the preview button can be customized.

#### Fixes & Improvements
- Fixed the instant backup deploy issue.
- Fixed link updates and scrolling in forms.

### Platform UI
#### Enhancements
- Migrated from Bulma to Tailwind with the new design system.
- Client organization metrics retention can be picked from the TelemetryStack.
- Shows a single metrics retention field for VictoriaMetrics clusters.
- Added site admin controls for organization-based licensing.
- Client organizations moved from Site Administration to Settings.
- The "Create New Org" button on the overview page now respects permissions.

### Billing UI
#### Enhancements
- Migrated from Bulma to Tailwind and updated to design system `2.13.0`.
- The dashboard table is now paginated.
- Shows the cluster name on the licensed clusters page.
- Added a contract-level license passthrough toggle.
- The issue license button is shown for all contract clusters.
- Latest licenses of a contract cluster are shown in a modal.
- Added copy buttons to tables (shown only when the table has data).

#### Fixes & Improvements
- Fixed the usage details route link and tooltip footer.
- Fixed the loading state, month picker and checkbox issues.

### UI Wizards (v0.38.0 → v0.39.1)
#### Enhancements
- Added `RotateAuth` to OpsRequest editors.
- Allowed selecting a BackupStorage for backup and archiver.
- Allowed toggling BackupStorage selection in the datastore feature-set editor.
- Added per-retention storage size to the TelemetryStack form.
- Added `config-syncer-manager` to the ocm-hub feature-set editor.
- Custom `MSSQL_PID` product keys are now stored in a Secret.

#### Fixes & Improvements
- Stale spec fields are cleared when the OpsRequest type changes.

### Platform Backend
#### Enhancements
- Background jobs moved from NATS to Cadence.
- Added a selfhost server.
- Switched to the AKP installer chart.
- Client organizations can pick a Thanos and VictoriaLogs retention from the TelemetryStack.
- Issued licenses are verified against the issuing CA.
- Opscenter-features and service-backend credentials are moved to `valuesFrom` Secrets; gateway DNS provider credentials and the JKS keystore are also treated as credentials.
- Opscenter-features credentials are shipped to spokes via the config-syncer addon, and spoke upgrades wait for the synced credential Secret.
- Uses the restructured Argo CD options and honors the managed Argo CD cluster name and namespace.
- Set Argo CD / argocd-agent chart values for OpenShift and pinned the fargocd version.
- Uses cert-manager for KubeStash webhook serving certificates, including on OCM spokes.

#### Fixes & Improvements
- Fixed spoke cluster removal.
- Stopped Perses sync from duplicating teams and members.
- Updated the sign-in/sign-up and activation UI.

### Installer
#### Enhancements
- Bumped KubeDB UI, Platform UI, Cluster UI and Billing UI to `3.0.0`.
- Added the AKP installer chart and OLM bundle, with Cadence workflow support; the bundle supports OpenShift 4.18 to 5.0.
- Added the `selfhost-api` chart for the b3 selfhost server.
- Added the `config-syncer-manager` OCM addon feature and moved to config-syncer `v1.0.0`.
- Added `remove-ace.sh` for Flux and fargocd (Argo CD) setups, and sped up ACE removal.
- Added pgpool settings to the ACE installer options.
- Added `backupStorage` toggle and enabled the Rotate Auth action in `kubedb-ui-presets`.
- Gateway API versions are picked from what the cluster serves.
- Raw credentials are kept out of HelmRelease `spec.values` and the bootstrap-presets ClusterChartPreset.
- Always uses cert-manager for KubeStash webhook certificates.
- Updated dependencies: envoy `v1.38.8-ac`, panopticon `v2026.9.18`, cluster-profile `v0.14.0`, cluster-proxy `v2026.11.11`, scanner `v2026.11.11`, aceshifter `v0.0.4`, openfga-alerts `v2026.7.14`, hub-cluster-robot `v2026.11.11`.

#### Fixes & Improvements
- Fixed ACE sync under Argo CD and Argo CD mode cleanup.
- Fixed the Cadence schema version resetting to `0.0` on ACE chart upgrades.
- Gateway certificates now work with domains longer than 64 bytes.
- The ACE cleaner hook no longer deletes `ace-gw-cert`.
- Disabled `ace-user-roles` in the KubeStash feature.

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
