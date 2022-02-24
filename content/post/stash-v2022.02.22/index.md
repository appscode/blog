---
title: Introducing Stash v2022.02.22
date: 2022-02-22
weight: 15
authors:
  - Piyush Kanti Das
tags:
  - kubernetes
  - stash
  - backup
  - restore
  - cross-namespace
  - grafana
  - cli
  - docs
---
We are very excited to announce Stash v2022.02.22. In this release, we have made many enhancements for Stash. You can find the complete changelog here. We are going to highlight the major changes in this post.

## Cross-namespace Repository
We are introducing cross-namespace Repository in this release. We brought a few changes to our Repository spec. We added `spec.usagePolicy` to the Repository CRD.  From now, you can take backup of your target or restore the target in any namespace you wish by creating only one Repository. We implemented double opt-in checking in the Repository usagePolicy. You can control from which namespaces your Repository can be referred. If you refer to a Repository from a restricted namespace, Stash will not complete that backup/restore process.

For example,
The following spec of a Repository will allow Stash to take backup or restore in any namespace of a cluster by using this Repository. 
```yaml
spec:
  usagePolicy:
    allowedNamespaces:
      from: All
```
But, this Repository spec from below won’t allow referring this Repository from any namespace other than the same namespace of the Repository. 
```yaml
spec:
  usagePolicy:
    allowedNamespaces:
      from: All
```

## BackupConfiguration Phase
We have added a new field to BackupConfiguration. This Phase can be of any three values at a moment. They are:
Invalid: Invalid indicates that the BackupConfiguration is referring to any restricted object or resource for the current namespace
Not Ready: This indicates that either the BackupConfiguration is doing it’s intermediate processing before getting Ready or creating the Cronjob or this BackupConfiguration could not discover any of the referred Reposiroty/Secret/other CRD.
Ready: This Phase indicates that the BackupConfiguration is now ready to take backup and created Cronjob as well.


## Add `Invalid` Phase to RestoreSession Phase
If the Restoresession object refers to any object that is restricted for that respective namespace, the RestoreSession Phase will go into the `Invalid` phase and will not proceed with the restoration process until the RestoreSession Phase turns into `Running` 

## AppBinding serviceReference cross-namespace support 
Stash now supports serviceReference from any namespace in a cluster for Appbinding. Along with that, Stash now accepts URL to resolve service connections. With these changes, a user can take backup or restore in any namespace where the service can be referred. By URL support, Stash can take backup or restore databases outside of Kubernetes as well.

## Show Snapshot ID in kubectl list command
From this release, if you run the Spanshot listing command like `kubectl list snapshots -n demo`,  we will display the first 8 characters of our Snapshot ID in a separate column. This will make restoring/deleting a specific Snapshot convenient. 

## Add commands to Stash CLI
We have added two useful commands to our Stash CLI. The newly added commands are:
kubectl-stash pause: Pause Stash backup temporarily
kubectl-stash resume: Resume stash resources
Both of the above commands have several flags to enhance the user capabilities. 

## Fixes and upgrades into Stash CLI
We have upgraded and fixed known issues of Stash CLI in this release. Some of them are:
Released cli for darwin/arm64
Used stashed/restic image for darwin/arm64 support
Fixed stash broken cli 


## Add Grafana dashboards
We have added Grafana dashboards to Stash. Now you can monitor Stash CRDs through Grafana dashboards. We have added support for custom Pushgateway as well. 


## Updated documentation
Cleaned up deprecated documentation
Add docs for newly introduced changes
Added troubleshooting guides in the documentation
Updated documentation for Stash CLI and Restoresession customization

## Updated Stash Addons
Updated all Stash addons to cope with the newly introduced changes in this release. 
Refactored Addons implementation for improved performance

## Made the operator ready for arm64 nodes
Stash operator now supports arm64 nodes 

## Always keep the last completed BackupSession
Now, Stash will always keep the last completed BackupSession  when `backuphistorylimit>0`. It will keep the last completed BackupSession even if it exceeds the history limit. 

## Bug fixes and improvements
We have also squashed a few bugs caught in the testing process that will make Stash more resilient. We refactored the API and the controller logic for boosting the performance and ensuring the extendability of Stash. We cleaned up the legacy implementations that were no longer used.  We have updated dependencies in the Stash UI server.

### What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.02.22/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.02.22/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
