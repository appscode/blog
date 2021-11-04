---
title: Neo Jezek v2021.10.11 - Introducing NATS & ETCD Add-ons
date: 2021-10-12
weight: 15
authors:
  - Neo Jezek
tags:
  - kubernetes
  - stash
  - backup
  - nats
  - etcd
---


# User Feedback `by Neo Jezek`

![Neo Jezek](./jan.png)

# Setup KubeDB (on Minikube) - The fast way


## Install Minikube

Create a Minikube (or Kind) cluster with 

- at least 2 CPUs
- nailed down to Kubernetes version 1.21.x
- and enough memory

Use that Kubernetes version (higher one was not able to get KubeDB running at this moment)

`minikube start --memory=4096 --driver=virtualbox --kubernetes-version=v1.21.6 --cpus=4`

`kubectl get storageclass` (must return a "default" storageclass)

> **HINT:** If not present activate storageclass with:
> `minikube addons enable default-storageclass`.
---
> **another Hint:** If the storageclass of Minikube wont work, apply a local-path storageclass provided by Rancher:
> <https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml>
---

## Get ID of cluster

`kubectl get ns kube-system -o=jsonpath='{.metadata.uid}'`

Obtain a license here: <https://license-issuer.appscode.com/>

> **HINT:** That License MUST be recreated EACH TIME you recreate your Cluster/Minikube!

## Add KubeDB repos

```console
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm search repo appscode/kubedb
```

## Install KubeDB controller

Use most recent "latest" KubeDB version

```bash
helm install kubedb appscode/kubedb \
  --version v2021.09.30 \
  --namespace kubedb --create-namespace \
  --set-file global.license=./kubedb-community-minikube-license.txt
```

## Verify KubeDB installation

`watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"`

### Confirm CRDs are created

`kubectl get crd -l app.kubernetes.io/name=kubedb`

## Install a DB

KubeDB can install many different databases:

- Postgres
- Postgis (Postgres with Geocoding)
- MongoDB
- MariaDB
- PGBouncer
- Redis
- Elasticsearch
- TimescaleDB (Postgres Time series data with Postgres)
- and more...

### DB driver Manifest

Apply manifest for the PostgisDB deployment process (only needed with Postgis, others are preinstalled) (You can look up here: `k get postgresversion`)

### Select your desired DB

```bash
❯ kubectl get postgresversions
NAME                     VERSION   DISTRIBUTION   DB_IMAGE                               DEPRECATED   AGE
13.2                     13.2      PostgreSQL     postgres:13.2-alpine                                3m45s
13.2-debian              13.2      PostgreSQL     postgres:13.2                                       3m45s
9.6.21-debian            9.6.21    PostgreSQL     postgres:9.6.21                                     3m45s
timescaledb-2.1.0-pg12   12.6      TimescaleDB    timescale/timescaledb:2.1.0-pg12-oss                3m45s
timescaledb-2.1.0-pg13   13.2      TimescaleDB    timescale/timescaledb:2.1.0-pg13-oss                3m45s
```

> If your desired db is listed above you can skip the next step.

### Custom database manifest

For postgis, we need that definition, it's used by the KubeDB controller:

`kubectl apply -f postgis-version.yml`

That's the content:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: PostgresVersion
metadata:
  name: 13.2-postgis
spec:
  coordinator:
    image: kubedb/pg-coordinator:v0.6.0
  db:
    image: kubedb/postgres:13.2-postgis
  distribution: PostgreSQL
  exporter:
    image: prometheuscommunity/postgres-exporter:v0.9.0
  initContainer:
    image: kubedb/postgres-init:0.3.0 # 0.3 was important, 0.2 was not working
  podSecurityPolicies:
    databasePolicyName: postgres-db
  securityContext:
    runAsAnyNonRoot: true
    runAsUser: 999
  stash:
    addon:
      backupTask:
        name: postgres-backup-13.1
      restoreTask:
        name: postgres-restore-13.1
  version: "13.2"
```

## Install your (custom) Postgis database

Change that file to your needs (size,..)

`kubectl apply -f ./postgis.yml`


That is the content:

```yaml
# postgis.yml
# https://raw.githubusercontent.com/kubedb/postgres-docker/release-13.2-postgis/example/demo.yaml
---
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: demo-postgis
  namespace: demo
spec:
  version: "13.2-postgis"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut # in PRODUCTiON USE: "DoNotTerminate" !!!!
```

### Troubleshoot

> **HINT:** If you have errors on write access to your storageclass then you have to explicitly use Kubernetes `1.20.x`

 `kubectl describe pod -n demo demo-postgis-1`
 and

`kubectl logs -n demo demo-postgis-0 -c postgres-init-container`

## Install PGAdmin to test

`kubectl create -f ../pgadmin/demo.yaml`

To get the URL of PGAdmin:

`minikube service pgadmin -n demo --url`

Example

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: quick-postgres
  namespace: demo
spec:
  version: "13.2"
  storageType: Durable # or "Ephemeral"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate # 
```

## Deleting a "DoNotTerminate" database

```bash
kubectl patch -n demo demo-postgis -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo demo-postgis
```

https://github.com/kubedb/postgres-docker/blob/release-13.2-postgis/example/README.md

#### Confirm CRDs

`kubectl get crd -l app.kubernetes.io/name=stash`

## Stash - KubeDB Backup solution

```bash
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm search repo appscode/stash --version v2021.10.11
```

```bash
helm install stash appscode/stash  \
  --version v2021.10.11            \
  --namespace kube-system          \
  --set features.community=true    \
  --set-file global.license=/path/to/the/license.txt
```

                            
                                    