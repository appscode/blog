---
title: Introducing KubeStash v2023.12.28
date: "2024-01-02"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- cli
- kubernetes
- restore
- kubestash
- disaster-recovery
---

We are pleased to announce the release of `KubeStash v2023.12.28`, packed with new features and important bug fixes. You can check out the full changelog [here](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2023.12.28/README.md). 
In this post, we'll highlight the key updates.

### New Features

1. You can now restore manifest of `MySQL` components ([#89](https://github.com/kubestash/apimachinery/pull/89)). Here is an example how to configure `RestoreSession` to restore manifest of `MySQL` components:

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
   
    Lets, the applied BackupConfiguration is configured with multiple sessions. Now you want to trigger backup for specific sessions. To do so, you have to provide comma seperated sessions name using `sessions` flag.
    ```bash
   kubectl kubestash trigger -n <namespace> <backupconfiguration-name> --sessions=<sessions-name>
    ```
   
4. You can now apply the Pod Security Policy profile to `Restricted`. This profile is a highly restrictive policy aligned with best practices for hardening Pods. Learn more about pod security [here](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted). You can restrict a namespace by the following command:

   `kubectl label namespace <namespace_name> pod-security.kubernetes.io/enforce=restricted`

    Use below command to install KubeStash enforcing Pod Security Policy profile to `Restricted`:
    ```bash 
    helm install kubestash oci://ghcr.io/appscode-charts/kubestash \
          --version v2023.12.28 \
          --namespace kubestash --create-namespace \
          --set-file global.license=<LICENCE_FILE> \
          --wait --burst-limit=10000 --debug \
          --set kubestash-operator.operator.securityContext.seccompProfile.type=RuntimeDefault \
          --set kubestash-operator.rbacproxy.securityContext.seccompProfile.type=RuntimeDefault \
          --set kubestash-operator.cleaner.securityContext.seccompProfile.type=RuntimeDefault
    ```
   
    To enforce Pod Security Policy profile to `Restricted` in storage initialize/cleanup jobs or in backup/restore jobs, add Pod Security and Container Security to BackupStorage, BackupConfiguration and RestoreSession as below:

    Pod Security:
    ```yaml
    securityContext:
      runAsUser: 65535
      runAsNonRoot: true
      seccompProfile:
        type: RuntimeDefault
    ```
    Container Security:
    ```yaml
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
    ```

### Improvements & bug fixes
- Now you can create multiple `RetentionPolicy` along with one default `RetentionPolicy` in a namespace. There can be maximum one default `RetentionPolicy` in each namespace.
- Fixed a bug in validation for selector type usage policy.
- Resolved addon version for MySQL `8.1.0` and `8.2.0`.


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [here](https://github.com/kubestash/installer/blob/master/charts/kubestash-operator/README.md).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [here](https://github.com/kubestash/installer/blob/master/charts/kubestash-operator/README.md).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with KubeStash or want to request new features, please [file an issue](https://github.com/kubestash/project/issues/new).
