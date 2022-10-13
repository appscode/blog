---
title: Stash v2021.03.17 - A Better User Experience
date: "2021-03-18"
weight: 15
authors:
- Emruz Hossain
tags:
- backup
- kubernetes
- stash
aliases:
- /post/stash-v20201.03.17/
---

We are very excited to announce Stash `v2021.03.17`. In this release, we have focused on improving the user experience with Stash. We have simplified the installation process, improved KubeDB integration, added backup support for TLS secured databases, etc. We have also fixed various bugs and made other improvements.

In this post, we are going to highlight the major changes. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG).

### Simplified Installation Process

Previously, we had two separate charts for the community edition ( `stash-community`) and enterprise edition (`stash-enterprise`). We also had separate charts for each individual addon version. We used to provide a bash script to install those addon charts easily which wasn't platform-independent. Also, it was tedious for the users to keep track of which addon chart is usable with which operator version.

In this release, we have moved all the addon catalogs under a single chart named `stash-catalog`. Then, we have wrapped all the charts under a single parent chart named `stash` using [Helm dependency](https://helm.sh/docs/helm/helm_dependency/). Now, the user just has to care about only the parent chart. They can install, uninstall, upgrade all the Stash components using this combined chart. When they upgrade the parent chart, all the dependent charts will be automatically upgraded to the appropriate version. It also solves the platform dependency of the bash script-based addon installation method.

Check out our new installation method using the combined chart from [here](https://stash.run/docs/v2021.03.17/setup/).

### Seamless Migration between Community and Enterprise Edition

Since we have combined all the charts under a single parent chart, now it is very easy to migrate between the community edition and enterprise edition. Now, users can seamlessly migrate between the two editions using a simple `helm upgrade` command. Check out our new guide for migrating between the community edition and enterprise edition from [here](https://stash.run/docs/v2021.03.17/setup/upgrade/#migration-between-community-edition-and-enterprise-edition).

### Pre-install Official Addons

Now, we automatically install all the official addons along with the Stash Enterprise edition. So, the user no longer needs to worry about installing the addons. Whenever you upgrade the operator the addons will be upgraded automatically to the appropriate version. You can still use your custom addons. You just need to install them separately as you did before.

### A Better Integration with KubeDB

Stash is now aware of [KubeDB](https://kubedb.com/) managed databases. The KubeDB catalogs now come with embedded Stash addon information. Our KubeDB team includes the addon name, version, and respective parameters that are necessary to backup the database, in DBVersion objects. Here is an example of [ElasticsearchVersion](https://kubedb.com/docs/latest/guides/elasticsearch/concepts/catalog/) object with the pre-configured Stash addon information.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: ElasticsearchVersion
metadata:
  name: searchguard-7.9.3
spec:
  authPlugin: SearchGuard
  db:
    image: kubedb/elasticsearch:7.9.3-searchguard
  distribution: SearchGuard
  exporter:
    image: kubedb/elasticsearch_exporter:1.1.0
  initContainer:
    image: kubedb/toybox:0.8.4
    yqImage: kubedb/elasticsearch-init:7.9.3-searchguard
  podSecurityPolicies:
    databasePolicyName: elasticsearch-db
  stash:
    addon:
      backupTask:
        name: elasticsearch-backup-7.3.2
        params:
        - name: args
          value: --match=^(?![.])(?!searchguard).+
      restoreTask:
        name: elasticsearch-restore-7.3.2
  version: 7.9.3
```

So, as a KubeDB user, you no longer have to worry about the Stash addons. You no longer need to provide the addon information in `BackupConfiguration` or `RestoreSession` objects.

### A Stable Addon Name

We have also removed the `-vX` suffix form `Task` and `Function` name as it is vulnerable to any minor changes in the addon. Whenever we made any update to the addon, we had to update the version suffix too. This was to ensure that if the addon images get updated, it does not break the existing backup setup of the users for any version skew between the operator and the new addon image. However, it resulted in an unpleasant situation where a user had to update the `Task` name in all the existing BackupConfigurations whenever they update the addons.

Since we are now automatically upgrading the addons along with the operator, we no longer need to keep this suffix. Now, we can have a stable name for the addons so that the users no longer need to update the `Task` name in their existing BackupConfigurations after an upgrade.

### Restic 0.12.0

We have upgraded [restic](https://restic.net/) version to `v0.12.0`. This comes with a significant performance boost in prune operation. You can find the full changelog of restic `v0.12.0` from [here](https://github.com/restic/restic/releases/tag/v0.12.0).

### TLS Support for all databases

Now, all the database addons support TLS secured connection. Previously, only Elasticsearch and MongoDB addon supported TLS secured connection.

### What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2021.03.17/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2021.03.17/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
