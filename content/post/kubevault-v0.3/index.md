![The KubeVault Overview](/content/post/kubevault-v0.3/KubeVault_overview.svg)

KubeVault operator is a Kubernetes controller for [HashiCorp Vault](https://www.vaultproject.io/). Vault is a tool for secrets management, encryption as a service, and privileged access management. Deploying, maintaining, and managing Vault in Kubernetes could be challenging. KubeVault operator eases these operational tasks so that developers can focus on solving business problems.

# About this release

The KubeVault operator version [0.3.0]() is now available with more stabilities and features. New features, changes, and known issues that are solved are included in this topic.

## New features and enhancements

This release adds improvements related to the following components and concepts.

### SecretEngine CRD

The SecretEngine CustomResourceDefinition (CRD) which is designed to automate the process of enabling and configuring secret engines in Vault in a Kubernetes native way, is introduced in this new release.

Supported Vault secret engine list:

- AWS Secrets Engine
- Azure Secrets Engine **(New)**
- Database Secret Engine
  - MongoDB
  - MySQL
  - PostgreSQL
- GCP Secrets Engine **(New)**

### Azure secrets engine

The Azure secrets engine dynamically generates Azure service principals along with role and group assignments. Vault roles can be mapped to one or more Azure roles, and optionally group assignments, providing a simple, flexible way to manage the permissions granted to generated service principals.

The `AzureRole` and `AzureAccessKeyRequest` CRDs are introduced in this release to allow a user to create an Azure secret engine role and to request for credentials respectively in a Kubernetes native way.

### GCP secret engine

The GCP secrets engine dynamically generates Google Cloud service account keys and OAuth tokens based on IAM policies. This enables users to gain access to Google Cloud resources without needing to create or manage a dedicated service account.

The `GCPRole` and `GCPAccessKeyRequest` CRDs are introduced in this release to allow a user to create a GCP secret engine role and to request for credentials respectively in a Kubernetes native way.

### Vault authentication methods

Auth methods are the components in Vault that perform authentication and are responsible for assigning identity and a set of policies to a user. The support for `Azure authN` and `GCP authN` is added in this release.

The list of supported authentication methods:

- AWS IAM Auth Method
- Kubernetes Auth Method
- TLS Certificates Auth Method
- Token Auth Method
- Userpass Auth Method
- GCP IAM Auth Method **(New)**
- Azure Auth Method **(New)**

### Vault storage backend

The Vault storage backend represents the location for the durable storage of Vault's information. Each backend has pros, cons, advantages, and trade-offs. For example, some backends support high availability while others provide some robust features.

The support for `consul` and `filesystem` storage backend is added in this release. The `filesystem` storage backend is coupled with a [PersitentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) so that the Vault data can be stored in PersistentVolume.

The list of supported storage backend:

- Azure
- Consul **(New)**
- DynamoDB
- Etcd
- Filesystem **(New)**
- Google Cloud Storage
- In-memory
- MySQL
- PostgreSQL
- S3
- Swift

## Issues fixed

- Vault operator helm template for the cleaner job is missing namespace support [#80](https://github.com/kubevault/issues/issues/80)
- Support CSI driver for k8s 1.14+ [#65](https://github.com/kubevault/issues/issues/65)
- Feature Request: Allow mounting configmaps/secrets into Vault container [#60](https://github.com/kubevault/issues/issues/60)
- Feature request: Consul backend [#51](https://github.com/kubevault/issues/issues/51)
