---
title: Deploy Azure Databases with KubeDB Crossplane Provider
date: "2023-12-28"
weight: 26
authors:
- SK Ali Arman
tags:
- azure
- crossplane
- kubedb
- kubernetes
---

KubeDB is now a [Crossplane](https://www.crossplane.io/) distribution for Hyper Clouds. Crossplane connects your Kubernetes cluster to external, non-Kubernetes resources, and allows platform teams to build custom Kubernetes APIs to consume those resources. We have introduced providers for Azure.

You need [crossplane](https://docs.crossplane.io/v1.14/) already installed in your cluster. This will allow KubeDB users to provision and manage Cloud provider managed databases in a Kubernetes native way.

## Install Crossplane

```bash
helm upgrade -i crossplane \
  oci://ghcr.io/appscode-charts/crossplane \
  -n crossplane-system --create-namespace
```

Check the installation with the following command. You will see two pod with prefix name crossplane in case of successful installation. 

```bash
kubectl get pod -n crossplane-system
```

## Provider-Azure

### Installation

Install the Azure provider into Kubernetes cluster with helm chart.

```bash
helm upgrade -i kubedb-provider-azure \
  oci://ghcr.io/appscode-charts/kubedb-provider-azure \
  --version=v2023.12.28 \
  -n crossplane-system --create-namespace
```

The command deploys a KubeDB Azure provider on the Kubernetes cluster in the default configuration. This will install CRDs representing Azure database services. These CRDs allow you to create Azure database resources inside Kubernetes.
 
### Setup Provider Config

Install azure cli and login to azure cli with the following command. See [Azure](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) cli documentaion

```bash
az login
```

Create an Azure service principal and save it as azure-credentials.json

```bash
az ad sp create-for-rbac --sdk-auth --role Owner --scopes /subscriptions/your-subscription-id >azure-credentials.json
```

Create a Kubernetes secret with the Azure credentials.

```bash
kubectl create secret generic azure-secret -n crossplane-system --from-file=creds=./azure-credentials.json
```

Create the ProviderConfig with the following yaml file

```bash
cat <<EOF | kubectl apply -f -
apiVersion: azure.kubedb.com/v1beta1
metadata:
  name: default
kind: ProviderConfig
spec:
  credentials:
    source: Secret
    secretRef:
      name: azure-secret
      key: creds
      namespace: crossplane-system
EOF
```

### Create MSSQL Database

We need to create resource group, MSSQL server to deploy database. Lets create these.

Create resource group

```bash
cat <<EOF | kubectl apply -f -
apiVersion: azure.kubedb.com/v1alpha1
kind: ResourceGroup
metadata:
  annotations:
    meta.kubedb.com/example-id: sql/v1alpha1/mssqldatabase
  labels:
    testing.kubedb.com/example-name: example
  name: db
spec:
  forProvider:
    location: West Europe
EOF
```

Create Secret

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: example-secret
  namespace: crossplane-system
type: Opaque
data:
  example-key: dGVzdFBhc3N3b3JkITEyMw==
EOF
```
Create MSSQL Server

```bash
cat <<EOF | kubectl apply -f -
apiVersion: sql.azure.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  annotations:
    meta.kubedb.com/example-id: sql/v1alpha1/mssqlserver
  labels:
    testing.kubedb.com/example-name: example
  name: mssqlservername
spec:
  forProvider:
    administratorLogin: missadministrator
    administratorLoginPasswordSecretRef:
      key: example-key
      name: example-secret
      namespace: crossplane-system
    location: West Europe
    resourceGroupNameRef:
      name: db
    version: "12.0"
EOF
```
Create Database

```bash
cat <<EOF | kubectl apply -f -
apiVersion: sql.azure.kubedb.com/v1alpha1
kind: MSSQLDatabase
metadata:
  annotations:
    meta.kubedb.com/example-id: sql/v1alpha1/mssqldatabase
  labels:
    testing.kubedb.com/example-name: mssqlservername
  name: database
spec:
  forProvider:
    serverIdRef:
      name: mssqlservername
EOF
```
Check the deployment with the following command

```bash
kubectl get managed -n crossplane-system
```
### Provider Azure also supports

- MySQL
- CosmosDB
  - GremlinGraph
  - Mongo
  - SQL
  - Table
  - Cassandra

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).


