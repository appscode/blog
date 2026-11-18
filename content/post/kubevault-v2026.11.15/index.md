---
title: Introducing KubeVault v2026.11.15
date: "2026-11-18"
weight: 25
authors:
- Abdullah Al Masood
tags:
- aerospike
- alicloud-oss
- cassandra
- cockroachdb
- couchdb
- google-cloud-spanner
- kubernetes
- kubevault
- mssql
- oci-object-storage
- openbao
- secret-management
- storage-backend
- zookeeper
---

[KubeVault](https://kubevault.com) is a Kubernetes operator for [OpenBao](https://openbao.org/). OpenBao is an open source tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage OpenBao servers in Kubernetes. It also manages secret engines, auth methods and policies in a Kubernetes-native way.

We are excited to announce the release of [KubeVault v2026.11.15](https://kubevault.com/docs/v2026.11.15/setup/). The main change in this release is **nine new storage backends for `VaultServer`**. It also adds support for **Sigilr 2.6.5.1**. You can find the full commit-by-commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2026.11.15/README.md).

---

### Quick highlights

- `VaultServer` now supports nine more storage backends: **Aerospike**, **Alibaba Cloud OSS**, **Cassandra**, **CockroachDB**, **CouchDB**, **Google Cloud Spanner**, **Microsoft SQL Server (MSSQL)**, **OCI Object Storage** and **ZooKeeper**.
- Most of the database backends can point at a **KubeDB-managed database** through `databaseRef`. KubeVault reads the address, and where available the TLS material, from the database's `AppBinding`.
- Credentials and TLS material always come from Kubernetes `Secret`s. They are never written into the `VaultServer` spec.
- The admission webhook checks **HA capability** for every new backend, so `spec.exposePrimary` is only accepted on a backend that can actually elect a leader.
- Added support for **Sigilr 2.6.5.1**.

---

## New Version Support

KubeVault now supports Sigilr 2.6.5.1, which is built on OpenBao 2.6.5. To deploy a `VaultServer` with it, set `spec.version` to `sigilr-2.6.5.1`:

````yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  replicas: 3
  version: sigilr-2.6.5.1
  backend:
    raft:
      storage:
        storageClassName: "standard"
        resources:
          requests:
            storage: 1Gi
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
````

---

## Nine New Storage Backends

Before this release, a `VaultServer` could store its data in Raft, Consul, etcd, PostgreSQL, MySQL, DynamoDB, GCS, S3, Azure, Swift, the local filesystem or in memory. This release adds nine more under `spec.backend`:

| Backend | `spec.backend` field | High availability |
|---|---|---|
| Aerospike | `aerospike` | No |
| Alibaba Cloud OSS | `alicloudoss` | No |
| Apache Cassandra | `cassandra` | No |
| Apache CouchDB | `couchdb` | No |
| Apache ZooKeeper | `zookeeper` | Always |
| CockroachDB | `cockroachdb` | With `haEnabled: true` |
| Google Cloud Spanner | `spanner` | With `haEnabled: true` |
| Microsoft SQL Server | `mssql` | No |
| OCI Object Storage | `oci` | With `haEnabled: true` and `lockBucketName` |

A few things apply to all of them:

- **Secrets stay in Secrets.** Usernames, passwords, service account keys and API keys are read from the `Secret` named in `credentialSecretRef` (or `authInfoSecretRef` for ZooKeeper). Depending on how the underlying storage backend reads its configuration, KubeVault either passes them as environment variables (Alibaba Cloud OSS, CouchDB, OCI, Spanner) or writes them into the generated storage config (Aerospike, Cassandra, CockroachDB, MSSQL, ZooKeeper).
- **KubeDB integration.** When `databaseRef` points at a KubeDB `AppBinding`, KubeVault generates the host address from it. For Cassandra, CockroachDB and MSSQL it can also take the CA bundle (and client certificate, where supported) from the `AppBinding` when you don't set `tlsSecretRef`.
- **Pre-created storage.** Some backends expect the database, keyspace, namespace or bucket to exist already. The notes under each example say which.

Below is a short example for each backend. Each one shows only `spec.backend`; the rest of the `VaultServer` stays the same. All examples use the `demo` namespace and Sigilr 2.6.5.1. Here is a complete manifest, using CockroachDB as the backend:

````yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  replicas: 3
  version: sigilr-2.6.5.1
  backend:
    cockroachdb:
      databaseRef:
        name: cockroachdb
        namespace: demo
      database: vault
      sslMode: verify-full
      haEnabled: true
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
````

### Aerospike

The Aerospike namespace must already exist on the server.

````yaml
  backend:
    aerospike:
      hostList: aerospike.demo.svc:3000
      namespace: vault
      set: vault-data
      credentialSecretRef:
        name: aerospike-cred   # username, password
      authMode: INTERNAL
````

Use `hostname`/`port` for a single seed node, or `databaseRef` for a KubeDB-managed Aerospike. `clusterName`, `timeout` and `idleTimeout` are also available.

### Alibaba Cloud OSS

The bucket must already exist, and `credentialSecretRef` is required.

````yaml
  backend:
    alicloudoss:
      endpoint: https://oss-us-east-1.aliyuncs.com
      bucket: vault-data
      credentialSecretRef:
        name: alicloudoss-cred   # access_key, secret_key
````

The `Secret` holds `access_key` and `secret_key`.

### Cassandra

The keyspace and table must already exist. `protocolVersion` must be `"2"` or higher for username/password authentication.

````yaml
  backend:
    cassandra:
      databaseRef:
        name: cassandra
        namespace: demo
      keyspace: vault
      table: entries
      consistency: LOCAL_QUORUM
      protocolVersion: "4"
      credentialSecretRef:
        name: cassandra-cred   # username, password
      tlsEnabled: true
````

Without a `databaseRef`, list the nodes in `hosts` as a comma-separated string. For TLS, `tlsSecretRef` points at a `Secret` with a `pem_bundle` key; `tlsServerName` lets you verify nodes against a shared DNS name instead of their IPs.

### CockroachDB

CockroachDB uses the PostgreSQL wire protocol, so the spec looks a lot like the existing `postgresql` backend. The database must already exist; OpenBao creates the tables. CockroachDB supports HA, so you can set `haEnabled: true` and use `spec.exposePrimary`.

````yaml
  backend:
    cockroachdb:
      databaseRef:
        name: cockroachdb
        namespace: demo
      database: vault
      sslMode: verify-full
      haEnabled: true
````

With `databaseRef`, KubeVault reads the host and the credentials from the database's `AppBinding`. Without it, set `address` to the CockroachDB host and point `credentialSecretRef` at a `Secret` holding `username` and `password`. That `Secret` can instead hold a full `connection_url`, which replaces `address`, `database` and `sslMode`.

### CouchDB

`endpoint` is the full URL of the database, including its name. The database must already exist.

````yaml
  backend:
    couchdb:
      endpoint: http://couchdb.demo.svc:5984/vault
      credentialSecretRef:
        name: couchdb-cred   # username, password
````

With `databaseRef`, KubeVault builds the endpoint as `<appbinding url>/<database>`, so set `database` instead of `endpoint`.

### Google Cloud Spanner

`database` is the full Spanner database name. Credentials come from a service account key stored under `sa.json`. If you leave out `credentialSecretRef`, Google's default credentials are used instead, for example Workload Identity on GKE. The service account needs the `roles/spanner.databaseUser` role on the database.

````yaml
  backend:
    spanner:
      database: projects/<project-id>/instances/<instance-id>/databases/vault
      haEnabled: true
      credentialSecretRef:
        name: spanner-cred   # sa.json
````

OpenBao does not create tables, so create them before deploying. The data table is named `Vault` by default (set `table` to change it), and with `haEnabled: true` OpenBao also needs a leader-election table, named after `table` with `HA` appended (set `haTable` to change it):

````sql
CREATE TABLE Vault (
  Key   STRING(MAX) NOT NULL,
  Value BYTES(MAX),
) PRIMARY KEY (Key);

CREATE TABLE VaultHA (
  Key       STRING(MAX) NOT NULL,
  Value     STRING(MAX),
  Identity  STRING(MAX),
  Timestamp TIMESTAMP,
) PRIMARY KEY (Key);
````

### Microsoft SQL Server

OpenBao creates the database, schema and table if they are missing, provided the user has permission to do so.

````yaml
  backend:
    mssql:
      databaseRef:
        name: mssqlserver
        namespace: demo
      database: vault
      schema: dbo
      table: vault
      credentialSecretRef:
        name: mssql-cred   # username, password
      encrypt: "true"
````

`encrypt` accepts `true`, `false`, `strict` (TDS 8.0, SQL Server 2022+) or `disable`. If you leave it unset, KubeVault turns encryption on whenever a CA is available, either from `tlsSecretRef` or from the `databaseRef` `AppBinding`.

### OCI Object Storage

Both buckets must already exist. By default the backend authenticates with **instance principals**, which works when OpenBao runs on OCI compute and needs no `Secret`. To use an API key instead, set `authTypeAPIKey: true`:

````yaml
  backend:
    oci:
      bucketName: vault-data
      namespaceName: <object-storage-namespace>
      region: us-ashburn-1
      authTypeAPIKey: true
      credentialSecretRef:
        name: oci-cred   # config, oci_api_key.pem
      haEnabled: true
      lockBucketName: vault-lock
````

Every key of the `Secret` is mounted at `/etc/vault/storage/oci/.oci`. The `config` file is in the usual `~/.oci/config` format, and its `key_file` must point at the full mounted path, for example `key_file=/etc/vault/storage/oci/.oci/oci_api_key.pem`. HA needs a second bucket, set in `lockBucketName`.

### ZooKeeper

ZooKeeper is always HA-capable, just like Raft and Consul.

````yaml
  backend:
    zookeeper:
      address: zookeeper.demo.svc:2181
      path: vault/
      znodeOwner: digest:vault:<hashed-password>
      authInfoSecretRef:
        name: zookeeper-auth   # authInfo
````

The `zookeeper-auth` `Secret` holds an `authInfo` key with a `scheme:auth` pair (for example `digest:vault:<password>`). For TLS, set `tlsEnabled: true` and point `tlsSecretRef` at a `Secret` containing `ca.crt`, `tls.crt` and `tls.key`.

---

## HA Validation for `exposePrimary`

`spec.exposePrimary` sends traffic only to the active OpenBao node, so it only works on a storage backend with HA support. The admission webhook now knows about all nine new backends:

- **Aerospike, Alibaba Cloud OSS, Cassandra, CouchDB and MSSQL** are never HA-capable, so the webhook rejects `exposePrimary` with any of them.
- **CockroachDB** and **Spanner** are HA-capable when `haEnabled: true`. **OCI Object Storage** also needs `lockBucketName`.
- **ZooKeeper** is always HA-capable, like Raft and Consul.

The webhook rejects an invalid combination when you apply the `VaultServer`, so you don't find out later from a cluster that never elects a primary.

---

## What's Next?

Please try the latest release and give us your valuable feedback.

- To install KubeVault, follow the installation instructions [here](https://kubevault.com/docs/v2026.11.15/setup).
- To upgrade an existing installation, follow the upgrade guide [here](https://kubevault.com/docs/v2026.11.15/setup/upgrade).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
