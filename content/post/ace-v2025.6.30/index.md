---
title: Announcing ACE v2025.6.30
date: "2025-07-07"
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

We are pleased to announce the release of `ACE v2025.6.30`, packed with major improvements. You can check out the full changelog [HERE](https://github.com/appscode-cloud/CHANGELOG/blob/master/releases/v2025.6.30/README.md). ACE **v2025.6.30** brings a bunch of new features & fixes. This release focuses on improving scalability and automation for production-grade deployments. In this post, we’ll highlight the changes done in this release.

## Key Changes
- **OpenFGA**: Added OpenFGA support to achieve more granular-level RBAC.
- **New Databases**: Introduced support for **Hazelcast** & **Oracle** to the UI.
- **Redis Hostname Support**: Added "announce" configuration for Redis clusters.

Here are the components specific changes:

## Cluster UI

### Features
- A user will require `can_remove` permission to delete an imported cluster
- Viewing kubeconfig also require `can_view/edit` kubeconfig permission
- Hub operation also require `can_view/edit` hub permission
- All the similar actions now require permissions. By default it’s permitted unless denied
- All these permissions are coming/configured from openfga

### Fixes
- Handle floating points of memory in the kubedb-ui-preset’s machine-profiles


## KubeDB UI

### Features

#### Hazelcast
Provisioning, monitoring, alerting & tls support added

#### Oracle
Provisioning support added

#### Postgres
Add support for remote-replica

#### Redis
Add `announce` field in Redis. This is a feature in Redis that enables external connections to Redis cluster deployed within Kubernetes

### Fixes

#### Common
- `Fix date-conversion` while point-in-time-recovery, in all archiver-enabled databases.
- `Sort the database versions` in version dropdown while db-creation
- ConnectionCard Update: Handle `Multiple Ports` in Gateway 
- Delete gateway class, & correct cleanup route resources on `client-org deletion`.

#### FerretDB
Fix resource field


### Platform UI

#### Features
- Update permission for teams, where only admin can edit the `editor` type team and members of editor teach can edit the `viewer` type teams. All the teams related permissions are now managed by openfga.
- Claim billings from appscode.com is made workable in the new ui.
- Site admin can now delete org owned by himself from site administration tab

#### Fixes
- Fix rendering issues & update `email validation`


### Billing UI

#### Fixes
- Fix stack issue (Linegraph in namespace view was showing wrong data)


### Selfhost UI
#### Features
- User can be editor, viewer, admin and some other `granular permission`
- For importing a cluster you need to be an editor (Admins are by default editor)
- Installer list delete and other actions require editor permission
- Viewer can still see the list and download
- Permissions are permitted unless denied by default as before


### Platform Backend

#### OpenFGA Feature
We have integrated OpenFGA support in this release. It is an open-source authorization solution that allows developers to build granular access control using an easy-to-read modeling language and friendly APIs.

**Key Points** :
- `OpenFGA-based Authorization` :
Introduced authorization using OpenFGA for our platform system.
- `Default Teams on Organization Creation` :
When a new organization is created, two additional teams are now provisioned automatically:
  1) Editors: Granted write access to all organization resources.
  2) Viewers: Granted read-only access to all organization resources, including clusters, tokens, installers etc.
- `Manual Team Assignment` :
Users can manually assign the Editors and Viewers teams to grant write or read access, respectively, across resources.
- `Organization Details Management` :
Members of the Editors team can now edit organization-related information and other resources.
- `Custom Team Creation` :
Users can also create custom teams with read or write permissions, assigning view or edit access to team members as needed.

**Role Capabilities** :
- `Owners` :
Full administrative privileges, including all Editors permissions.
Can delete organizations, manage all teams (including Editors and Viewers), and perform any action on organization resources.
- `Editors` :
Manage Viewers team membership.
Granted write/edit permissions on all organization resources.
- `Viewers` :
Read-only access to all organization resources.

Note that, The existing Owners team remains the primary organization owner with full administrative privileges.


#### Fixes & Improvements
- Add support for `converting imported cluster to spoke` cluster
- Initialize db on ace upgrade process.
- Add support for `inbox-server`.
- Add support for `credential-less EKS` cluster creation & import.
- Fix owner team can't be update due to reserve team name
- `Fix in upgrade process` for fluxcd.


### External products
Here is the summary of external dependency updates for `ACE v2025.6.30` :

- `KubeDB`: v2025.6.30 [Release blog](https://appscode.com/blog/post/kubedb-v2025.6.30/).
- `Stash`: v2025.6.30 [Release blog](https://appscode.com/blog/post/stash-v2025.6.30/).
- `KubeStash`: v2025.6.30 [Release blog](https://appscode.com/blog/post/kubestash-v2025.6.30/).
- `Voyager Gateway`: v2025.6.30 [Chart Ref](https://github.com/voyagermesh/installer/tree/release-v2025.6.30/charts/voyager-gateway). This uses envoy `v1.34.1-ac` & envoy-gateway `1.4.1`.
- `monitoring-operator`: v2025.6.30 [Chart Ref](https://appscode.com/blog/post/kubestash-v2025.6.30/).
- `panopticon`: v2025.6.30 [Chart Ref](https://github.com/open-viz/installer/tree/release-v2025.6.30/charts/monitoring-operator).


## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/appscode) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [ACE](https://appscode.com/docs/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/appscode-cloud/launchpad/issues).