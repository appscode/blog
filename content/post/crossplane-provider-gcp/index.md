---
title: Deploy GCP Databases with KubeDB Crossplane Provider
date: "2023-12-28"
weight: 26
authors:
- SK Ali Arman
tags:
- crossplane
- gcp
- kubedb
- kubernetes
---

KubeDB is now a [Crossplane](https://www.crossplane.io/) distribution for Hyper Clouds. Crossplane connects your Kubernetes cluster to external, non-Kubernetes resources, and allows platform teams to build custom Kubernetes APIs to consume those resources. We have introduced providers for GCP.

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

## Provider-GCP

### Installation

Install the GCP provider into Kubernetes cluster with helm chart.

```bash
helm upgrade -i kubedb-provider-gcp \
  oci://ghcr.io/appscode-charts/kubedb-provider-gcp \
  --version=v2023.12.28 \
  -n crossplane-system --create-namespace
```

The command deploys a KubeDB GCP provider on the Kubernetes cluster in the default configuration. This will install CRDs representing GCP database services. These CRDs allow you to create GCP database resources inside Kubernetes.

### Setup Provider Config

Generate a GCP service account JSON file and save it as appscode-testing.json. See the [GCP](https://cloud.google.com/iam/docs/keys-create-delete) documentaion.

Create a Kubernetes secret with the GCP credentials.

```bash
kubectl create secret generic gcp-secret -n crossplane-system --from-file=creds=./appscode-testing.json
```

Create the ProviderConfig with the following yaml file

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gcp.kubedb.com/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  projectID: <PROJECT_ID>
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: gcp-secret
      key: creds
EOF
```

### Create SQL Database

Ensure that your service account has the control of Cloud SQL resource.

Create a database instance for the sql database

```bash
cat <<EOF | kubectl apply -f -
apiVersion: sql.gcp.kubedb.com/v1alpha1
kind: DatabaseInstance
metadata:
  annotations:
    meta.kubedb.com/example-id: sql/v1alpha1/databaseinstance
  labels:
    testing.kubedb.com/example-name: example_instance
  name: example-instance
spec:
  forProvider:
    region: "us-central1"
    databaseVersion: "MYSQL_5_7"
    settings:
      - tier: "db-f1-micro"
        diskSize: 20
    deletionProtection: false
EOF
```

Create database

```bash
cat <<EOF | kubectl apply -f -
apiVersion: sql.gcp.kubedb.com/v1alpha1
kind: Database
metadata:
  annotations:
    meta.kubedb.com/example-id: sql/v1alpha1/database
  labels:
    testing.kubedb.com/example-name: example_database
  name: example-database
spec:
  forProvider:
    instanceRef:
      name: example-instance
EOF
```

### Provider GCP also supports

- Redis
- Spanner

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).


