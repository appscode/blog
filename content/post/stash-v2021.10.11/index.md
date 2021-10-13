---
title: Stash v2021.10.11 - Introducing NATS & ETCD Addons
date: 2021-10-11
weight: 15
authors:
  - Hossain Mahmud
tags:
  - kubernetes
  - stash
  - backup
  - nats
  - etcd
---

We are very excited to announce Stash v2021.10.11. In this release, we are introducing NATS and ETCD addons for Stash. We have also added TLS support for Redis Addon.

In this post, we are going to highlight the major changes. You can find the complete changelog [here.](https://github.com/stashed/CHANGELOG/blob/master/releases/v2021.10.11/README.md)

## Introducing NATS Addon

In this release, we have added NATS to our Stash addons family. Now, you can backup your NATS Jetstream server running inside Kubernetes using Stash. We have added the addon version 2.6.1 for NATS. You should be able to backup any NATS server of 2.x.x series.

Related resources:

* [How the NATS backup works in Stash.](https://stash.run/docs/v2021.10.11/addons/nats/overview/)
* [Step by step guide to backup NATS managed by Helm.](https://stash.run/docs/v2021.10.11/addons/nats/helm/)
* Step by step guide to backup NATS using different authentication methods.
  * [Basic Authentication](https://stash.run/docs/v2021.10.11/addons/nats/authentications/basic-auth/)
  * [Token Authentication](https://stash.run/docs/v2021.10.11/addons/nats/authentications/token-auth/)
  * [Nkey Authentication](https://stash.run/docs/v2021.10.11/addons/nats/authentications/nkey-auth/)
  * [JWT Authentication](https://stash.run/docs/v2021.10.11/addons/nats/authentications/jwt-auth/)
* [Step by step guide to backup TLS secured NATS.](https://stash.run/docs/v2021.10.11/addons/nats/tls/)
* [Customizing the backup and restore process according to your environment.](https://stash.run/docs/v2021.10.11/addons/nats/customization/)

## Introducing ETCD Addon

In this release, we have also added ETCD addon. Now, you can backup your ETCD database running inside Kubernetes using Stash. We have added the addon versioned 3.5.0 for ETCD. You should be able to backup your ETCD 3.x.x series ETCD database using this addon.

## Bug Fix and Enhancements

We have also squashed a few bugs and added a few enhancements in this release. Here are some of the highlighting fixes & enhancements.

* **Support passing args to restic backup/restore command:** We have added the support for passing arguments to restic backup and restore command. Now User can pass optional arguments to the restic backup/restore process via `spec.target.args` field of BackupConfiguration/RestoreSession. Here are some examples:

  **Pass args to restic backup command:**
  ```yaml
  apiVersion: stash.appscode.com/v1beta1
  kind: BackupConfiguration
  metadata:
    name: sample-deployment-backup
    namespace: demo
  spec:
    repository:
      name: deployment-backup-repo
    schedule: "*/5 * * * *"
    target:
      ref:
        apiVersion: apps/v1
        kind: Deployment
        name: sample-deployment
      volumeMounts:
        - name: source-data
          mountPath: /source/data
      paths:
        - /source/data
      args: ["--ignore-inode", "--tag=t1,t2"]
    retentionPolicy:
      name: "keep-last-5"
      keepLast: 5
      prune: true
  ```

  **Pass args to restic restore command:**

  ```yaml
  apiVersion: stash.appscode.com/v1beta1
  kind: RestoreSession
  metadata:
    name: sample-deployment-restore
    namespace: demo
  spec:
    repository:
      name: deployment-backup-repo
    target:
      ref:
        apiVersion: apps/v1
        kind: Deployment
        name: sample-deployment
      volumeMounts:
        - name: source-data
          mountPath: /source/data
      rules:
        - paths:
            - /source/data/
      args: ["--tag=t1,t2"]
  ```

* **Fix license-reader ClusterRoleBinding cleaning:** Some of our users were facing issues with license-reader ClusterRoleBinding not getting deleted after deleting the respective RestoreSession. We have fixed the issue in this release. Now, the ClusterRoleBinding should get deleted when a user deletes the respective RestoreSession.

* **Use  restic v0.12.1:** In this release, we have upgraded the restic version to v0.12.1.

* **Show actual repository size instead of logical size:** Previously, Stash showed the logical size of the repository which did not reflect the actual repository size. We have fixed this issue. Now, your Repository size should be the same as how much actual data is in the backend.

* **Update restic docker image in Stash kubectl plugin:** We have upgraded the underlying restic docker image used by Stash kubectl plugin.

* **TLS support in Redis addon:** We have added the TLS support for redis addon in this release.

* **MongoDB 5.0.3 and PostgreSQL 14.0 support:** We have added support for MongoDB 5.0.3. Now, you can backup your MongoDB 5.x.x using this addon. We have also added support for PostgreSQL 14.0.

### What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2021.10.11/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2021.10.11/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
