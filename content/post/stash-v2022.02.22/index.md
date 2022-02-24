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

We are very excited to announce Stash `v2022.02.22`. In this release, we have introduced some exciting features and fixed some bugs. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.02.22/README.md). We are going to highlight the major changes in this post.

## New Features

Here, we are going to highlight the new features that have been introduced in this release.

### Support cross-namespace Repository reference

We are introducing support for cross-namespace Repository reference. Now, you can refer to a Repository in BackupConfiguration and RestoreSession from different namespaces. So, you can now easily restore into different namespaces without copying the Repository or you can keep your Repository and backend Secret isolated from the application namespaces where the backup happens.

To support this feature, we are introducing the `spec.usagePolicy` field in the Repository CRD. This lets you control which namespaces are allowed to use the Repository and which are not. If you refer to a Repository from a restricted namespace, Stash will reject creating the respective BackupConfiguration/RestoreSession from validating webhook.

You can use the `usagePolicy` to allow only the same namespace, a subset of namespaces, or all the namespaces to refer to the Repository.

For example, here is the sample YAML of a Repository that allows referencing it from all namespaces.

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: s3-repo
  namespace: demo
spec:
  backend:
    s3:
      endpoint: s3.amazonaws.com
      bucket: stash-demo
      region: us-west-1
      prefix: /backup/demo/deployment/stash-demo
    storageSecretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
```

Here, is the example of a Repository that allows referencing it only from the same namespace.

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: s3-repo
  namespace: demo
spec:
  backend:
    s3:
      endpoint: s3.amazonaws.com
      bucket: stash-demo
      region: us-west-1
      prefix: /backup/demo/deployment/stash-demo
    storageSecretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: Same
```

Finally, here is the example of a Repository that allows referencing it only from `prod` and `staging` namespaces.

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: s3-repo
  namespace: demo
spec:
  backend:
    s3:
      endpoint: s3.amazonaws.com
      bucket: stash-demo
      region: us-west-1
      prefix: /backup/demo/deployment/stash-demo
    storageSecretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: Selector
      selector:
        matchExpressions:
        - key: "kubernetes.io/metadata.name"
          operator: In
          values: ["prod","staging"]
```

### Introducing `phase` in BackupConfiguration status

We have added a `phase` field to the BackupConfiguration status. This can help you understand the backup setup status.

Currently, the `phase` field accepts the following values:

- **Invalid:** It indicates that the BackupConfiguration has failed the validation check. It can happen when the BackupConfiguration refers to a Repository of a different namespace and the current namespace is not allowed by the `usagePolicy` of the Repository.

- **NotReady:** It indicates that the backup setup is not completed yet. It can happen when some dependency (i.e. Repository, backend Secret, etc.) are missing, or Stash hasn't completed processing the BackupConfiguration yet. You can describe the BackupConfiguration to check why it is in the `NotReady` state.

- **Ready:** It indicates that the backup setup was completed successfully and the backup will be triggered from the next schedules.

### Introducing `Invalid` phase in RestoreSession status

We have added an `Invalid` phase in the RestoreSession status. It indicates that the RestoreSession has failed the validation check. This can happen when the RestoreSession references a Repository of a restricted namespace.

### Support cross-namespace Service reference in AppBinding

AppBinding now supports referencing Services of a different namespace. This can help you to backup a database from a different namespace.

### Support URL in the AppBinding

Stash now supports referencing your application via a URL in AppBinding. This allows you to backup/restore databases located outside of your Kubernetes cluster.

### Custom Pushgateway support

You can now use your custom pushgateway to push Stash metrics. If you use a custom pushgateway, Stash will not inject the pushgateway sidecar to the operator. You can now pass the following helm flag during installation to use the custom pushgateway.

 ```bash
 # For community edition
 --set stash-community.pushgateway.customURL=<your pushgateway url>

 # For enterprise edition
 --set stash-enterprise.pushgateway.customURL=<your pushgateway url>
 ```

### Show Snapshot ID in `kubectl get snapshots` command

We have added Snapshot ID in the response to the `kubectl get snapshots` command. This helps you easily identify the Snapshot ID which is necessary to restore a particular Snapshot.

### Add `pause` and `resume` commands to Stash `kubectl` plugin

We have added two useful commands to our Stash `kubectl` plugin. You can now easily pause and resume your backup using the commands.

For example, to pause a backup, you can run:

```bash
kubectl stash pause --backupconfig=sample-mongo-backup -n demo
```

To resume the backup, you can run:

```bash
kubectl stash resume --backupconfig=sample-mongo-backup -n demo
```

### Support for M1 Mac in Stash kubectl plugin

We have added support M1 Mac in our Stash kubectl plugin.

## Bug Fixes

Now, here are a few bugs we have squashed in this release.

### Fix broken `download` command in Stash kubectl plugin

You can now again run the `kubectl stash download` command to download a snapshot to your local machine.

### Always keep the last completed BackupSession

Now, Stash will always keep the last completed BackupSession when `backuphistorylimit>0`. It will keep the last completed BackupSession even if it exceeds the history limit. This will help to keep the backup history when a backup gets skipped due to another running backup.

## Deprecation and Cleanup

In this release, we have removed the deprecated v1alpha1 APIs. This means the `Restic` and `Recovery` CRDs are no longer available in Stash.

## Documentation Improvements

Now, here are a few improvements we made on the documentation side.

- Removed deprecated documentation
- Add docs for cross namespace Repository reference.
- Added troubleshooting guides that show how to identify and fix common Stash issues.
- Improved addon documentations

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.02.22/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.02.22/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
