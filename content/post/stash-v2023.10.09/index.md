---
title: Introducing Stash v2023.10.9
date: "2023-10-09"
weight: 10
authors:
- Hossain Mahmud
tags:
- backup
- cli
- kubernetes
- mongodb
- restore
- stash
---

We are pleased to announce the relase of [Stash v2023.10.9](https://stash.run/docs/v2023.10.9/setup/), packed with new features and important bug fixes. You can check out the full changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2023.10.9/README.md). 
In this post, we'll highlight the key updates.

### New Features

1. You can now manage your restic keys (passwords) using Stash CLI ([#189](https://github.com/stashed/cli/pull/189)). Here are some examples:

   - List all restic keys associated with a specific repository:
       ```bash
       kubectl stash key list -n <namespace> <repo-name>
       ```

   - Add a new restic key to a specifix repository:
        ```bash
        kubectl stash key add -n <namespace> <repo-name> --new-password-file <password-filepath>
        ```     

   - Update current restic key:
        ```bash
        kubectl stash key update -n <namespace> <repo-name> --new-password-file <password-filepath>
        ```
   - Remove a restic key:
        ```bash
        kubectl stash key remove -n <namespace> <repo-name> --id <key-id>
        ```

2. You can now use the Stash CLI to create `rules` for a RestoreSession  to recover your database and application backups. This command helps you find the nearest repository snapshots for a given timestamp and generates two rules for you: one for the snapshots just before the specified timestamp and another for those at or after the specified timestamp.
    ```bash
    kubectl stash gen rules -n <namespace> <repo-name> --timestamp <timestamp>
    ```
3. Previously, downloading snapshots with Stash CLI didn't work when using Google Workload Identity or AWS IRSA. However, now you can download snapshots even when using these identity mechanisms.
4. Previously, restoring specific snapshots other than the latest one was not supported for MongoDB shard clusters. Now, you have the flexibility to restore specific snapshots as needed ([#1927](https://github.com/stashed/mongodb/pull/1927)). Here is an example:
    ```yaml
    apiVersion: stash.appscode.com/v1beta1
    kind: RestoreSession
    metadata:
      name: sample-mongodb-restore
      namespace: demo
    spec:
      repository:
        name: gcs-repo
      target:
        ref:
          apiVersion: appcatalog.appscode.com/v1alpha1
          kind: AppBinding
          name: sample-mgo-sh
      rules:
      - snapshots:
        - 39f64846
        targetHosts:
        - confighost
      - snapshots:
        - bc13b4a8
        targetHosts:
        - host-0
      - snapshots:
        - e9cb7ab9
        targetHosts:
        - host-1
    ```
5. We've introduced a new hook execution policy called `OnFinalRetryFailure`. This policy triggers a backup hook only when the backup process has failed with no more retry attempts available ([#179](https://github.com/stashed/apimachinery/pull/210)).


### Improvements & bug fixes
#### MongoDB Addon
- We now resync the already lagged secondaries of a replicaSet before locking it up. That's how we ensure that the secondary has up-to-date oplogs while running `mongodump` command.
- Increasing stopBalancer timeout : We need to stop the sharded cluster's balancer, before taking backup. We now wait for 10 minutes to stop it gracefully. We also try the setBalancerState command multiple times while re-enabling balancer. This should avoid the timing-related issues.
- A huge amount of improvement was done in logging. 



## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/latest/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
