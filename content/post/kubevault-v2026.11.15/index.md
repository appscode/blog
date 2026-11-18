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
- hashicorp
- kubernetes
- kubevault
- mssql
- oci-object-storage
- openbao
- secret-management
- storage-backend
- vault
- zookeeper
---

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/) and [OpenBao](https://openbao.org/). Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also manages secret engines, auth methods and policies in a Kubernetes-native way.

We are excited to announce the release of [KubeVault v2026.11.15](https://kubevault.com/docs/v2026.11.15/setup/). The main change in this release is **nine new storage backends for `VaultServer`**. You can find the full commit-by-commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2026.11.15/README.md).

---

### Quick highlights

- `VaultServer` now supports nine more storage backends: **CockroachDB**, **Cassandra**, **ZooKeeper**, **CouchDB**, **Microsoft SQL Server (MSSQL)**, **Google Cloud Spanner**, **Aerospike**, **OCI Object Storage** and **Alibaba Cloud OSS**.
- Most of the database backends can point at a **KubeDB-managed database** through `databaseRef`. KubeVault reads the address, and where available the TLS material, from the database's `AppBinding`.
- Credentials and TLS material always come from Kubernetes `Secret`s. They are never written into the `VaultServer` spec.
- The admission webhook checks **HA capability** for every new backend, so `spec.exposePrimary` is only accepted on a backend that can actually elect a leader.
- Fixed a bug where a CockroachDB `connection_url` lost its host when only `credentialSecretRef` was set.

---

## Nine New Storage Backends

Before this release, a `VaultServer` could store its data in Raft, Consul, etcd, PostgreSQL, MySQL, DynamoDB, GCS, S3, Azure, Swift, the local filesystem or in memory. This release adds nine more under `spec.backend`:

| Backend | `spec.backend` field | `databaseRef` support | High availability |
|---|---|---|---|
| CockroachDB | `cockroachdb` | Yes | With `haEnabled: true` |
| Apache Cassandra | `cassandra` | Yes | No |
| Apache ZooKeeper | `zookeeper` | Yes | Always |
| Apache CouchDB | `couchdb` | Yes | No |
| Microsoft SQL Server | `mssql` | Yes | No |
| Google Cloud Spanner | `spanner` | No | With `haEnabled: true` |
| Aerospike | `aerospike` | Yes | No |
| OCI Object Storage | `oci` | No | With `haEnabled: true` and `lockBucketName` |
| Alibaba Cloud OSS | `alicloudoss` | No | No |

A few things apply to all of them:

- **Secrets stay in Secrets.** Usernames, passwords, service account keys and API keys are read from the `Secret` named in `credentialSecretRef` (or `authInfoSecretRef` for ZooKeeper). Depending on how the underlying storage backend reads its configuration, KubeVault either passes them as environment variables (CouchDB, Spanner, Alibaba Cloud OSS, OCI) or writes them into the generated storage config (CockroachDB, Cassandra, ZooKeeper, MSSQL, Aerospike).
- **KubeDB integration.** When `databaseRef` points at a KubeDB `AppBinding`, KubeVault generates the host address from it. For CockroachDB, Cassandra and MSSQL it can also take the CA bundle (and client certificate, where supported) from the `AppBinding` when you don't set `tlsSecretRef`.
- **Pre-created storage.** Some backends expect the database, keyspace, namespace or bucket to exist already. The notes under each example say which.

Below is a short example for each backend. All examples use the `demo` namespace and OpenBao 2.7.1.

### CockroachDB

CockroachDB uses the PostgreSQL wire protocol, so the spec looks a lot like the existing `postgresql` backend. The database must already exist; Vault creates the tables. CockroachDB supports HA, so you can set `haEnabled: true` and use `spec.exposePrimary`.

````yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  replicas: 3
  version: openbao-2.7.1
  backend:
    cockroachdb:
      databaseRef:
        name: cockroach-db
        namespace: demo
      database: vault
      sslMode: verify-full
      credentialSecretRef:
        name: cockroach-cred
      haEnabled: true
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
````

The credential `Secret` holds `username` and `password`. You can also give it a full `connection_url`, which then replaces `address`, `database` and `sslMode`. Without a `databaseRef`, set `address` to the CockroachDB host instead.

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
        name: cassandra-cred
      tlsEnabled: true
````

Without a `databaseRef`, list the nodes in `hosts` as a comma-separated string. For TLS, `tlsSecretRef` points at a `Secret` with a `pem_bundle` key; `tlsServerName` lets you verify nodes against a shared DNS name instead of their IPs.

### ZooKeeper

ZooKeeper is always HA-capable, just like Raft and Consul.

````yaml
  backend:
    zookeeper:
      address: zk-0.zk-pods.demo.svc:2181,zk-1.zk-pods.demo.svc:2181,zk-2.zk-pods.demo.svc:2181
      path: vault/
      znodeOwner: digest:vault:<hashed-password>
      authInfoSecretRef:
        name: zk-auth
````

The `zk-auth` `Secret` holds an `authInfo` key with a `scheme:auth` pair (for example `digest:vault:password`). For TLS, set `tlsEnabled: true` and point `tlsSecretRef` at a `Secret` containing `ca.crt`, `tls.crt` and `tls.key`.

### CouchDB

`endpoint` is the full URL of the database, including its name. The database must already exist.

````yaml
  backend:
    couchdb:
      endpoint: http://couchdb.demo.svc:5984/vault
      credentialSecretRef:
        name: couchdb-cred
````

With `databaseRef`, KubeVault builds the endpoint as `<appbinding url>/<database>`, so set `database` instead of `endpoint`.

### Microsoft SQL Server

Vault creates the database, schema and table if they are missing, provided the user has permission to do so.

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
        name: mssql-cred
      encrypt: "true"
````

`encrypt` accepts `true`, `false`, `strict` (TDS 8.0, SQL Server 2022+) or `disable`. If you leave it unset, KubeVault turns encryption on whenever a CA is available, either from `tlsSecretRef` or from the `databaseRef` `AppBinding`.

### Google Cloud Spanner

`database` is the full Spanner database name. Credentials come from a service account key stored under `sa.json`.

````yaml
  backend:
    spanner:
      database: projects/my-project/instances/vault/databases/vault
      table: vault-data
      haEnabled: true
      credentialSecretRef:
        name: spanner-sa
````

With `haEnabled: true`, the leader-election table defaults to the value of `table` with `HA` appended. Use `haTable` to choose a different name.

### Aerospike

The Aerospike namespace must already exist on the server.

````yaml
  backend:
    aerospike:
      hostList: aerospike-0.demo.svc:3000,aerospike-1.demo.svc:3000
      namespace: vault
      set: vault-data
      credentialSecretRef:
        name: aerospike-cred
      authMode: INTERNAL
````

Use `hostname`/`port` for a single seed node, or `databaseRef` for a KubeDB-managed Aerospike. `clusterName`, `timeout` and `idleTimeout` are also available.

### OCI Object Storage

Both buckets must already exist. By default the backend authenticates with **instance principals**, which works when Vault runs on OCI compute and needs no `Secret`. To use an API key instead, set `authTypeAPIKey: true`:

````yaml
  backend:
    oci:
      bucketName: vault-data
      namespaceName: my-tenancy-namespace
      region: us-ashburn-1
      authTypeAPIKey: true
      credentialSecretRef:
        name: oci-api-key
      haEnabled: true
      lockBucketName: vault-lock
````

Every key of the `Secret` is mounted at `/etc/vault/storage/oci/.oci`. The `config` file is in the usual `~/.oci/config` format, and its `key_file` must point at the full mounted path, for example `key_file=/etc/vault/storage/oci/.oci/oci_api_key.pem`. HA needs a second bucket, set in `lockBucketName`.

### Alibaba Cloud OSS

The bucket must already exist, and `credentialSecretRef` is required.

````yaml
  backend:
    alicloudoss:
      endpoint: https://oss-us-east-1.aliyuncs.com
      bucket: vault-data
      credentialSecretRef:
        name: alicloud-cred
````

The `Secret` holds `access_key` and `secret_key`.

---

## HA Validation for `exposePrimary`

`spec.exposePrimary` sends traffic only to the active Vault node, so it only works on a storage backend with HA support. The admission webhook now knows about all nine new backends:

- **ZooKeeper** is always HA-capable, like Raft and Consul.
- **CockroachDB** and **Spanner** are HA-capable when `haEnabled: true`. **OCI Object Storage** also needs `lockBucketName`.
- **Cassandra, CouchDB, MSSQL, Aerospike and Alibaba Cloud OSS** are never HA-capable, so the webhook rejects `exposePrimary` with any of them.

The webhook rejects an invalid combination when you apply the `VaultServer`, so you don't find out later from a cluster that never elects a primary.

---

## Bug Fixes

- **CockroachDB connection URL:** when only `credentialSecretRef` was set (no `databaseRef`), the generated `connection_url` silently dropped the host, because `address` was only used when *neither* field was set. `address` is now always used as the host unless `databaseRef` overrides it.

---

## What's Next?

Please try the latest release and give us your valuable feedback.

- To install KubeVault, follow the installation instructions [here](https://kubevault.com/docs/v2026.11.15/setup).
- To upgrade an existing installation, follow the upgrade guide [here](https://kubevault.com/docs/v2026.11.15/setup/upgrade).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
