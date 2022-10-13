---
title: Stash v2021.08.02 - Introducing Redis Addon and Kubernetes 1.22 Support
date: 2021-08-02
weight: 15
authors:
- Emruz Hossain
tags:
- kubernetes
- stash
- backup
- redis
aliases:
- /post/stash-v2021.08.02/
---

We are very excited to announce Stash `v2021.08.02`. In this release, we are introducing Redis addon for Stash. We have also added support for Kubernetes version 1.22.

In this post, we are going to highlight the major changes. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG).

### Introducing Redis Addon

In this release, we have added Redis to our Stash addons family. Now, you can backup your Redis database running inside Kubernetes using Stash. We have added two addons versioned `6.2.5` and `5.0.13` for Redis. You should be able to backup your Redis `6.x` and `5.x` series Redis databases using those addons.

Related resources:

- [How the Redis backup works in Stash](https://stash.run/docs/v2021.08.02/addons/redis/overview/).
- [Step by step guide to backup Redis databases managed by Helm](https://stash.run/docs/v2021.08.02/addons/redis/helm/).
- [Using Stash Auto-backup with Redis](https://stash.run/docs/v2021.08.02/addons/redis/auto-backup/).
- [Customizing backup and restore process according to your environment](https://stash.run/docs/v2021.08.02/addons/redis/customization/).

### Bug Fix and Enhancements

We have also squashed a few bugs and added few enhancements in this release. Here are some of the highlighting fixes & enhancements.

- **Fix High Memory Usage Issue:** Some of our users were seeing spike in memory usage by Stash backup process after upgrading to Stash `v2021.06.23`. We have fixed the issue in this release. Your memory usage should be stable again after upgrading to `v2021.08.02`.
- **Support Kubernetes 1.22.x:** We have upgraded the underlying libraries to support Kubernetes 1.22.x. Now, Stash should work flawlessly with Kubernetes 1.22.x.
- **Fix Region Issue with S3 Backend:** Some of our users were facing issues with custom region in S3 backend. We have fixed this issue. Now, Stash should work with custom region for S3 and S3 compatible backends.
- **Added Duration Column:** We have added `DURATION` column in `kubectl get` output for BackupSession and RestoreSession. You can now check how much time it took to complete the backup and restore process directly in `kubectl get` output. Here, is an example:

```bash
# BackupSession
❯ kubectl get backupsession -n demo
NAME                             INVOKER-TYPE          INVOKER-NAME          PHASE       DURATION    AGE
sample-redis-backup-1627490702   BackupConfiguration   sample-redis-backup   Succeeded   1m18s       89s

# RestoreSession
❯ kubectl get restoresession -n demo -w
NAME                   REPOSITORY   PHASE       DURATION     AGE
sample-redis-restore   gcs-repo     Succeeded   16s          56s
```

### What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2021.08.02/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2021.08.02/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
