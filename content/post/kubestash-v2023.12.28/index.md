---
title: Introducing KubeStash v2023.12.28
date: "2024-01-02"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- cli
- disaster-recovery
- kubernetes
- kubestash
- restore
---

We are pleased to announce the release of `KubeStash v2023.12.28`, packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2023.12.28/README.md). 
In this post, we'll highlight the key updates.

### New Features

1. You can now backup and restore the manifests of KubeDB managed `MySQL` database ([#89](https://github.com/kubestash/apimachinery/pull/89)). 
   
   Here is an example how to configure `BackupConfiguration` to backup manifests of KubeDB managed `MySQL` database:
   ```yaml
   apiVersion: core.kubestash.com/v1alpha1
   kind: BackupConfiguration
   metadata:
     name: mysql-manifest-backup
     namespace: demo
   spec:
     target:
       apiGroup: kubedb.com
       kind: MySQL
       namespace: demo
       name: sample-mysql
     sessions:
     - name: manifest-backup
       sessionHistoryLimit: 3
       scheduler:
         schedule: "*/10 * * * *"
         successfulJobsHistoryLimit: 1
         failedJobsHistoryLimit: 1
         jobTemplate:
           backoffLimit: 1
       repositories:
       - name: mysql-repo
         directory: /manifest/sample-mysql
         encryptionSecret:
           name: mysql-encry-secret
           namespace: demo
       addon:
         name: mysql-addon
         tasks:
         - name: ManifestBackup
       failurePolicy: Retry
       retryConfig:
         maxRetry: 2
         delay: 1m
   ```

   Here is an example how to configure `RestoreSession` to restore manifests of KubeDB managed `MySQL` database:

    ```yaml
    apiVersion: core.kubestash.com/v1alpha1
    kind: RestoreSession
    metadata:
      name: mysql-manifest-restore
      namespace: demo
    spec:
      manifestOptions:
        restoreNamespace: mysql-demo
        mySQL:
          db: true
          dbName: demo-mysql
          authSecret: true
          configSecret: true
      dataSource:
        snapshot: latest
        repository: mysql-repo
        encryptionSecret:
          name: mysql-encry-secret
          namespace: demo
      addon:
        name: mysql-addon
        tasks:
        - name: ManifestRestore
    ```

2. We've added support for pod hook execution strategy. Now you can provide a value for field `.spec.executor.pod.strategy` in a `HookTemplate`. Pod hook execution strategy specifies what should be the behavior when multiple pods are selected depending on `selector` for executing the hook. The valid values for this field are:
   
   - `ExecuteOnOne`: Execute hook on only one of the selected pods. This is default behavior.
   - `ExecuteOnAll`: Execute hook on all the selected pods.

      Here is an example of `HookTemplate` using pod executor with `ExecuteOnOne` strategy:
      ```yaml
      apiVersion: core.kubestash.com/v1alpha1
      kind: HookTemplate
      metadata:
       name: demo-pod-hook
       namespace: demo
      spec:
        usagePolicy:
          allowedNamespaces:
            from: All
        params:
        - name: TEST
          usage: This is a test param
          required: false
        action:
          exec:
            command:
            - /bin/sh
            - -c
            - echo data_test > /source/data/data.txt
        executor:
          type: Pod
          pod:
            selector: name=test-app, test=hook
            strategy: ExecuteOnOne
     ```

3. You can now trigger backup for specific sessions using KubeStash CLI ([#10](https://github.com/kubestash/cli/pull/10)). Here is an example:
   
    Assume that the applied `BackupConfiguration` is configured with multiple sessions. Now you want to trigger backup for specific sessions. To do so, you have to provide comma (,) separated sessions name using `sessions` flag.
    ```bash
   $ kubectl kubestash trigger -n <namespace> <backupconfiguration-name> --sessions=<sessions-name>
    ```

### Improvements & Bug Fixes
- Fixed a bug that was preventing the creation of multiple Retention Policies when a default one exists in a namespace.
- Fixed a bug in validation for selector type usage policy.
- Fixed a bug that was causing an "addon not found" error for MySQL versions `8.1.0` and `8.2.0`.


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://github.com/kubestash/installer/blob/master/charts/kubestash-operator/README.md).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://github.com/kubestash/installer/blob/master/charts/kubestash-operator/README.md).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).

