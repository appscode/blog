---
title: Introducing Stash v2024.12.18
date: "2024-12-18"
weight: 10
authors:
- Md Anisur Rahman
tags:
- backup
- cli
- disaster-recovery
- kubedump
- kubernetes
- mongodb
- postgresql
- restore
- stash
---

We are pleased to announce the release of [Stash v2024.8.27](https://stash.run/docs/v2024.8.27/setup/), packed with new features. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2024.4.8/README.md). In this post, we'll highlight the changes done in this release.

### New Features

#### Introducing the `multiDumpArgs` Parameter for MySQL Backups

We are excited to introduce the `multiDumpArgs` parameter, which is available for all MySQL versions during the backup process.

**What is `multiDumpArgs`?**

The `multiDumpArgs` parameter can be specified under the `spec.task.params` section in the `BackupConfiguration`. It allows you to define multiple dump commands, which will be executed sequentially and separated by the Bash `&&` operator.

**Why Use `multiDumpArgs`?**

This feature is designed for use cases where you need to take database backups into different dump files, such as:

- Backing up `table-schema` and `table-data` into separate files.

- Customizing multiple dump operations to suit your specific needs.

Using `multiDumpArgs`, you have to provide dump arguments by separating each command with the `$args` placeholder.

**Format Configuration:**

```YAML
task:
  params:
    - name: multiDumpArgs
      value: >-
        $args= <arguments for the first dump command>  
        $args= <arguments for the second dump command>  
        $args= <arguments for the third dump command>  

```

**Example Configuration:**

```YAML
task:
  params:
    - name: args
      value: --set-gtid-purged=OFF
    - name: multiDumpArgs
      value: >-
        $args=--no-tablespaces --no-data --skip-triggers --skip-opt --single-transaction --create-options --disable-keys --extended-insert --set-charset --quick --databases playground  
        $args=--no-tablespaces --no-create-info --skip-triggers --skip-opt --single-transaction --create-options --disable-keys --extended-insert --set-charset --quick --ignore-table=playground.equipment --databases playground  
        $args=--no-tablespaces --no-data --no-create-info --skip-opt --single-transaction --create-options --disable-keys --extended-insert --set-charset --quick --databases playground
```

#### Templating for Exec Hook Command

We've added support for [Go template](https://pkg.go.dev/text/template) in exec type as well. Previously it was only limited to httpPost hook. Here is an example of configuring template in exec hook:

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: mysql-backup
  namespace: demo
spec:
  hooks:
    postBackup:
      executionPolicy: Always
      exec:
        command:
          - curl
          - -X
          - POST
          - -H
          - 'Content-type: application/json'
          - -d
          - '{"text":"ENV Backup status for {{ .Namespace }}/{{ .Target.Name }} Phase: {{ if eq .Status.Phase `Succeeded`}}Succeeded{{ else }}Failed. Reason: {{ .Status.Error }}{{ end}}."}'
          - https://hooks.slack.com
      containerName: mysql 
  schedule: "*/5 * * * *"
  task:
    name: mysql-backup-8.0.21
    params:
    - name: args
      value: --set-gtid-purged=OFF --single-transaction --all-databases
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: mysql
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [HERE](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [HERE](https://stash.run/docs/latest/setup/upgrade/).


### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).

