---
title: Chaos Engineering KubeDB MySQL Group Replication on Kubernetes, Testing Group Replication Cluster Resilience
date: "2026-04-27"
weight: 14
authors:
- SK Ali Arman
tags:
- chaos-engineering
- chaos-mesh
- database
- group-replication
- high-availability
- kubedb
- kubernetes
- mysql
---

## Overview

We conducted **60+ chaos experiments** across **3 MySQL versions** (8.0.36, 8.4.8, 9.6.0) and **2 Group Replication topologies** (Single-Primary and Multi-Primary) on KubeDB-managed 3-node clusters. The goal: validate that KubeDB MySQL delivers **zero data loss**, **automatic failover**, and **self-healing recovery** under realistic failure conditions with production-level write loads.

**The result: every experiment passed with zero data loss, zero split-brain, and zero errant GTIDs.**

This post summarizes the methodology, results, and key findings from comprehensive chaos testing of KubeDB MySQL Group Replication.

## Why Chaos Testing?

Running databases on Kubernetes introduces failure modes that traditional infrastructure does not have — pods can be evicted, nodes can go down, network policies can partition traffic, and resource limits can trigger OOMKills at any time. Chaos engineering deliberately injects these failures to verify that the system recovers correctly **before** they happen in production.

For a MySQL Group Replication cluster managed by KubeDB, we needed to answer:

- Does the cluster **lose data** when a primary is killed mid-transaction?
- Does **automatic failover** work under network partitions?
- Can the cluster **self-heal** after a full outage with no manual intervention?
- Are **GTIDs consistent** across all nodes after recovery?
- Does the cluster survive **combined failures** (CPU + memory + load simultaneously)?

## Test Environment

| Component | Details |
|---|---|
| Kubernetes | kind (local cluster) |
| KubeDB Version | 2026.4.27 |
| Cluster Topology | 3-node Group Replication (Single-Primary & Multi-Primary) |
| MySQL Versions | 8.0.36, 8.4.8, 9.6.0 |
| Storage | 2Gi PVC per node (Durable, ReadWriteOnce) |
| Memory Limit | 1.5Gi per MySQL pod |
| CPU Request | 500m per pod |
| Chaos Engine | Chaos Mesh |
| Load Generator | sysbench `oltp_write_only`, 4-12 tables, 4-16 threads |
| Baseline TPS | ~2,400 (Single-Primary) / ~1,150 (Multi-Primary) |

All experiments were run under **sustained sysbench write load** to simulate production traffic during failures.

## Setup Guide

### Step 1: Create a kind Cluster

We used [kind](https://kind.sigs.k8s.io/) (Kubernetes IN Docker) as our local Kubernetes cluster. Follow the [kind installation guide](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) to install it, then create a cluster:

```bash
kind create cluster --name chaos-test
```

### Step 2: Install KubeDB

Install KubeDB operator using Helm:

```bash
helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2026.4.27 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/license.txt \
  --wait --burst-limit=10000 --debug
```

### Step 3: Install Chaos Mesh

Install Chaos Mesh for fault injection:

```bash
helm repo add chaos-mesh https://charts.chaos-mesh.org
helm repo update chaos-mesh

helm upgrade -i chaos-mesh chaos-mesh/chaos-mesh \
  -n chaos-mesh --create-namespace \
  --set dashboard.create=true \
  --set dashboard.securityMode=false \
  --set chaosDaemon.runtime=containerd \
  --set chaosDaemon.socketPath=/run/containerd/containerd.sock \
  --set chaosDaemon.privileged=true
```

### Step 4: Deploy MySQL Cluster

Create the namespace:

```bash
kubectl create namespace demo
```

**Single-Primary Mode:**

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-ha-cluster
  namespace: demo
spec:
  deletionPolicy: Delete
  podTemplate:
    spec:
      containers:
        - name: mysql
          resources:
            limits:
              memory: 1.5Gi
            requests:
              cpu: 500m
              memory: 1.5Gi
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  storageType: Durable
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  version: 8.4.8
```

**Multi-Primary Mode** (change only the group mode):

> **Note:** Multi-Primary mode in KubeDB is available from MySQL version **8.4.2** and above.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-ha-cluster
  namespace: demo
spec:
  deletionPolicy: Delete
  podTemplate:
    spec:
      containers:
        - name: mysql
          resources:
            limits:
              memory: 1.5Gi
            requests:
              cpu: 500m
              memory: 1.5Gi
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  storageType: Durable
  topology:
    mode: GroupReplication
    group:
      mode: Multi-Primary
  version: 8.4.8
```

Deploy and wait for Ready:

```bash
kubectl apply -f mysql-ha-cluster.yaml
kubectl wait --for=jsonpath='{.status.phase}'=Ready mysql/mysql-ha-cluster -n demo --timeout=5m
```

### Step 5: Deploy sysbench Load Generator

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sysbench-load
  namespace: demo
  labels:
    app: sysbench
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sysbench
  template:
    metadata:
      labels:
        app: sysbench
    spec:
      containers:
        - name: sysbench
          image: perconalab/sysbench:latest
          command: ["/bin/sleep", "infinity"]
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "2"
              memory: "2Gi"
          env:
            - name: MYSQL_HOST
              value: "mysql-ha-cluster.demo.svc.cluster.local"
            - name: MYSQL_PORT
              value: "3306"
            - name: MYSQL_USER
              value: "root"
            - name: MYSQL_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mysql-ha-cluster-auth
                  key: password
            - name: MYSQL_DB
              value: "sbtest"
```

```bash
kubectl apply -f sysbench.yaml
```

### Step 6: Prepare sysbench Tables

```bash
#!/bin/bash

# Get the MySQL root password
PASS=$(kubectl get secret mysql-ha-cluster-auth -n demo -o jsonpath='{.data.password}' | base64 -d)


kubectl exec -n demo svc/mysql-ha-cluster -c mysql -- \
  mysql -uroot -p"$PASS" -h mysql-ha-cluster.demo -e "DROP DATABASE IF EXISTS sbtest;"

# Create the sbtest database
kubectl exec -n demo svc/mysql-ha-cluster -c mysql -- \
  mysql -uroot -p"$PASS" -h mysql-ha-cluster.demo -e "CREATE DATABASE IF NOT EXISTS sbtest;"

# Get the sysbench pod name
SBPOD=$(kubectl get pods -n demo -l app=sysbench -o jsonpath='{.items[0].metadata.name}')

# Standard write load (used during most experiments)
kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
  --mysql-host=mysql-ha-cluster --mysql-port=3306 \
  --mysql-user=root --mysql-password="$PASS" \
  --mysql-db=sbtest --tables=12 --table-size=100000 \
  --threads=8 prepare
```

### Step 7: Run sysbench During Chaos

```bash
# Standard write load (used during most experiments)
kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
  --mysql-host=mysql-ha-cluster --mysql-port=3306 \
  --mysql-user=root --mysql-password="$PASS" \
  --mysql-db=sbtest --tables=12 --table-size=100000 \
  --threads=8 --time=60 --report-interval=10 run
```

## Chaos Testing

We will run chaos experiments to see how our cluster behaves under failure scenarios like pod kill, OOM kill, network partition, network latency, IO latency, IO fault, and more. We will use sysbench to simulate high write load on the cluster during each experiment.

> **Important Notes on Database Status:**
> - **`Ready`** — Database is fully operational. All pods are ONLINE.
> - **`Critical`** — Primary is accepting connections and operational, but one or more replicas may be down.
> - **`NotReady`** — Primary is not available. No writes can be accepted.
>
> You can read/write in your database in both **`Ready`** and **`Critical`** states. So even if your db is in `Critical` state, your uptime is not compromised.

### Verify Cluster is Ready

Before starting chaos experiments, verify the cluster is healthy:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE    ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    174m

NAME                                 READY   STATUS    RESTARTS   AGE    ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          174m   primary
pod/mysql-ha-cluster-1               2/2     Running   0          174m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          174m   standby
pod/sysbench-load-849bdc4cdc-h2zpx   1/1     Running   0          8d
```

Inspect the GR members:

```shell
➤ kubectl exec -n demo mysql-ha-cluster-0 -c mysql -- \
    mysql -uroot -p"$PASS" -e "SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE \
    FROM performance_schema.replication_group_members;"
MEMBER_HOST                                           MEMBER_PORT  MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         3306         ONLINE        PRIMARY
```

The pod having `kubedb.com/role=primary` is the primary and `kubedb.com/role=standby` are the secondaries. With the cluster ready and sysbench tables prepared, we are ready to run chaos experiments.

### Chaos#1: Kill the Primary Pod

We are about to kill the primary pod and see how fast the failover happens. Save this yaml as `tests/01-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: mysql-primary-pod-kill
  namespace: chaos-mesh
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  gracePeriod: 0
```

**What this chaos does:** Terminates the primary pod abruptly with `grace-period=0`, forcing an immediate failover to a standby replica.

- **Expected behavior:**
  Primary pod is killed → cluster transitions `Ready` → `Critical` (1 replica missing after failover to a standby) → when the killed pod rejoins as standby, cluster returns to `Ready`. Zero data loss, GTIDs and checksums consistent across all 3 nodes.

- **Actual result:**
  Pod-0 killed → pod-2 elected as new primary in ~1 seconds → pod-0 rejoined as standby ~30s later → cluster returned to `Ready`. All 3 members `ONLINE` in GR, GTIDs match across nodes, checksums match on every sysbench table. **PASS.**

Before running, let's see who is the primary:

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                                 READY   STATUS    RESTARTS   AGE    ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          174m   primary
pod/mysql-ha-cluster-1               2/2     Running   0          174m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          174m   standby
```

Now run `watch kubectl get mysql,pods -n demo -L kubedb.com/role` in one terminal, and apply the chaos in another:

```shell
➤ kubectl apply -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org/mysql-primary-pod-kill created
```

Within seconds, the primary pod is killed and a new primary is elected. The database goes `Critical`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE    ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Critical   174m

NAME                                 READY   STATUS    RESTARTS   AGE    ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          14s    standby
pod/mysql-ha-cluster-1               2/2     Running   0          174m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          174m   primary
```

Note the `STATUS: Critical` — this means the new primary (pod-2) is accepting connections and operational, but the killed pod-0 hasn't rejoined yet. After ~30 seconds, the old primary comes back as a standby and the cluster returns to `Ready`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h1m

NAME                                 READY   STATUS    RESTARTS   AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          7m24s   standby
pod/mysql-ha-cluster-1               2/2     Running   0          3h1m    standby
pod/mysql-ha-cluster-2               2/2     Running   0          3h1m    primary
```

Verify data integrity — all 3 nodes must have matching GR status, GTIDs and checksums:

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
MEMBER_HOST                                           MEMBER_PORT  MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         3306         ONLINE        PRIMARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
```

All 3 members ONLINE. Check GTIDs and checksums:

```shell
➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-7743:1000001-1000004
pod-1: 65a93aae-...:1-7743:1000001-1000004
pod-2: 65a93aae-...:1-7743:1000001-1000004

➤ # Checksums — all match ✅
pod-0: sbtest1=2141984737, sbtest2=779706826, sbtest3=3549430025, sbtest4=3045058695
pod-1: sbtest1=2141984737, sbtest2=779706826, sbtest3=3549430025, sbtest4=3045058695
pod-2: sbtest1=2141984737, sbtest2=779706826, sbtest3=3549430025, sbtest4=3045058695
```

**Result: PASS** — Zero data loss. Failover completed in ~5 seconds. The old primary rejoined as standby automatically. All GTIDs and checksums match across all 3 nodes.

Clean up:

```shell
➤ kubectl delete -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org "mysql-primary-pod-kill" deleted
```

---

### Chaos#2: OOMKill the Primary Pod

Now we are going to OOMKill the primary pod. This is a realistic scenario — in production, your primary pod might get OOMKilled due to high memory usage from large queries or connection spikes.

Save this yaml as `tests/02-oomkill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: mysql-primary-memory-stress
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  stressors:
    memory:
      workers: 2
      size: "1200MB"
  duration: "10m"
```

**What this chaos does:** Allocates 1200MB of extra memory on the primary pod. Combined with MySQL's memory usage (~500MB), this exceeds the 1.5Gi limit and triggers an OOMKill.

- **Expected behavior:**
  Primary pod is OOMKilled during sysbench writes → cluster goes `NotReady` during failover → after ~20s(based on config) a new primary is elected and the OOMKilled pod rejoins as standby → cluster returns to `Ready`. Zero data loss, GTIDs and checksums consistent across nodes.

- **Actual result:**
  Primary pod-2 OOMKilled mid-writes. Sysbench lost connection (error 2013). After ~20s(as replication_unreachable_majority_timeout = 20) the unreachable member was expelled, pod-1 elected as new primary. Pod-2 restarted (`Restarts: 1`) and rejoined as standby ~82s after kill. All 3 `ONLINE` in GR, GTIDs match, checksums match. **PASS.**

First, start the sysbench load test so we have writes in-flight when the OOMKill hits:

```shell
➤ kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
    --mysql-host=mysql-ha-cluster --mysql-port=3306 \
    --mysql-user=root --mysql-password="$PASS" \
    --mysql-db=sbtest --tables=12 --table-size=100000 \
    --threads=4 --time=120 --report-interval=10 run

[ 10s ] thds: 4 tps: 744.67 qps: 4469.72 lat (ms,95%): 9.56 err/s: 0.00
```

Now apply the memory stress while sysbench is running:

```shell
➤ kubectl apply -f tests/02-oomkill.yaml
stresschaos.chaos-mesh.org/mysql-primary-memory-stress created
```

The primary pod (pod-2) gets OOMKilled. Sysbench loses connection immediately:

```shell
FATAL: mysql_stmt_execute() returned error 2013 (Lost connection to MySQL server during query)
```

The database goes `NotReady` during failover:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE    ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     NotReady   3h3m

NAME                                 READY   STATUS    RESTARTS     AGE    ROLE
pod/mysql-ha-cluster-0               2/2     Running   0            8m     standby
pod/mysql-ha-cluster-1               2/2     Running   0            3h3m   standby
pod/mysql-ha-cluster-2               2/2     Running   1 (9s ago)   3h3m   standby
```

Note the `Restarts: 1` on pod-2 — it was OOMKilled and restarted by Kubernetes. The status `NotReady` means failover is in progress. This is an ungraceful shutdown for the primary node. The node remains part of the group but becomes unreachable. After 20 seconds, the unreachable node is expelled from the group and a new primary is elected. Within approximately 60 seconds, the cluster fully recovers, and the failed node rejoins as a standby.

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE    ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h4m

NAME                                 READY   STATUS    RESTARTS      AGE    ROLE
pod/mysql-ha-cluster-0               2/2     Running   0             10m    standby
pod/mysql-ha-cluster-1               2/2     Running   0             3h4m   primary
pod/mysql-ha-cluster-2               2/2     Running   1 (82s ago)   3h4m   standby
```

pod-1 is now the new primary. pod-2 (the OOMKilled pod) rejoined as standby. Verify data integrity:

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
MEMBER_HOST                                           MEMBER_PORT  MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo         3306         ONLINE        PRIMARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-19216:1000001-1000006
pod-1: 65a93aae-...:1-19216:1000001-1000006
pod-2: 65a93aae-...:1-19216:1000001-1000006

➤ # Checksums — all match ✅
pod-0: sbtest1=1213558811, sbtest2=1030289216, sbtest3=1306867904, sbtest4=3604669046
pod-1: sbtest1=1213558811, sbtest2=1030289216, sbtest3=1306867904, sbtest4=3604669046
pod-2: sbtest1=1213558811, sbtest2=1030289216, sbtest3=1306867904, sbtest4=3604669046
```

**Result: PASS** — OOMKill triggered on primary during active writes. Failover completed automatically. The OOMKilled pod restarted and rejoined as standby. Zero data loss — all GTIDs and checksums match across all 3 nodes.

Clean up:

```shell
➤ kubectl delete -f tests/02-oomkill.yaml
stresschaos.chaos-mesh.org "mysql-primary-memory-stress" deleted
```

---

### Chaos#3: Network Partition the Primary

We are going to isolate the primary from all standby replicas. The standbys will lose contact with the primary and elect a new one.

Save this yaml as `tests/03-network-partition.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-primary-network-partition
  namespace: chaos-mesh
spec:
  action: partition
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "kubedb.com/role": "standby"
  direction: both
  duration: "2m"
```

**What this chaos does:** Creates a complete network partition between the primary and all standby replicas for 2 minutes. The primary loses quorum and is expelled from the group. The standbys elect a new primary.

- **Expected behavior:**
  Primary isolated from standbys → within `group_replication_unreachable_majority_timeout` (~20s) the isolated primary loses quorum → standbys elect a new primary → cluster goes `Critical` → after partition removed the old primary rejoins → cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Partition applied for 2 minutes. New primary (pod-2) elected in ~20s. Old primary (pod-1) stayed isolated until partition ended, then rejoined automatically via coordinator restart. Cluster returned to `Ready` at ~4 minutes total. GTIDs and checksums match across all 3 nodes. **PASS.**

Before running:

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS       AGE    ROLE
mysql-ha-cluster-0               2/2     Running   0              11m    standby
mysql-ha-cluster-1               2/2     Running   0              3h5m   primary
mysql-ha-cluster-2               2/2     Running   1 (2m ago)     3h5m   standby
```

Apply the chaos:

```shell
➤ kubectl apply -f tests/03-network-partition.yaml
networkchaos.chaos-mesh.org/mysql-primary-network-partition created
```

Within ~20 seconds (the `group_replication_unreachable_majority_timeout`), the isolated primary loses quorum. The standbys elect a new primary, and the database goes `Critical`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE    ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Critical   3h6m

NAME                                 READY   STATUS    RESTARTS       AGE    ROLE
pod/mysql-ha-cluster-0               2/2     Running   0              11m    standby
pod/mysql-ha-cluster-1               2/2     Running   0              3h6m   standby
pod/mysql-ha-cluster-2               2/2     Running   1 (3m ago)     3h6m   primary
```

`Critical` means the new primary (pod-2) is accepting connections and operational, but the isolated pod-1 is not yet back in the group. Let's check GR from pod-0:

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
MEMBER_HOST                                           MEMBER_PORT  MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         3306         ONLINE        PRIMARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
```

Only 2 members visible — pod-1 is isolated. After the 2-minute partition expires, the coordinator automatically restarts the isolated node and it rejoins the group:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h10m

NAME                                 READY   STATUS    RESTARTS       AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0              15m     standby
pod/mysql-ha-cluster-1               2/2     Running   0              3h10m   standby
pod/mysql-ha-cluster-2               2/2     Running   1 (7m ago)     3h10m   primary
```

Verify data integrity:

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
MEMBER_HOST                                           MEMBER_PORT  MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         3306         ONLINE        PRIMARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-19238:1000001-1000048
pod-1: 65a93aae-...:1-19238:1000001-1000048
pod-2: 65a93aae-...:1-19238:1000001-1000048

➤ # Checksums — all match ✅
pod-0: sbtest1=1213558811, sbtest2=1030289216, sbtest3=1306867904, sbtest4=3604669046
pod-1: sbtest1=1213558811, sbtest2=1030289216, sbtest3=1306867904, sbtest4=3604669046
pod-2: sbtest1=1213558811, sbtest2=1030289216, sbtest3=1306867904, sbtest4=3604669046
```

**Result: PASS** — Network partition triggered failover in ~20 seconds. The isolated node rejoined automatically after partition removed. Zero data loss — all GTIDs and checksums match across all 3 nodes.

Clean up:

```shell
➤ kubectl delete -f tests/03-network-partition.yaml
networkchaos.chaos-mesh.org "mysql-primary-network-partition" deleted
```

---

### Chaos#4: IO Latency on Primary (100ms)

We inject 100ms latency on every disk I/O operation on the primary's data volume. This simulates a slow storage system — a common issue in cloud environments with noisy neighbors or degraded storage backends.

Save this yaml as `tests/04-io-latency.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: mysql-primary-io-latency
  namespace: chaos-mesh
spec:
  action: latency
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  volumePath: "/var/lib/mysql"
  path: "/**"
  delay: "100ms"
  percent: 100
  duration: "3m"
```

**What this chaos does:** Adds 100ms delay to every disk read/write operation on the primary's MySQL data directory. Every `fsync`, `write`, and `read` is slowed down.

- **Expected behavior:**
  Primary's disk I/O slowed by 100ms → TPS drops significantly → cluster stays `Ready` (no failover — InnoDB should handle slow disk gracefully, not crash) → when chaos ends, TPS recovers. Zero data loss, GTIDs/checksums consistent.

- **Actual result:**
  TPS degraded from 703 → 104 (~85% drop). No failover triggered. All 3 members stayed `ONLINE` throughout. After chaos removed, TPS recovered. GTIDs and checksums match across all 3 nodes. **PASS.**

Apply the chaos and run sysbench simultaneously:

```shell
➤ kubectl apply -f tests/04-io-latency.yaml
iochaos.chaos-mesh.org/mysql-primary-io-latency created

➤ kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
    --mysql-host=mysql-ha-cluster --mysql-port=3306 \
    --mysql-user=root --mysql-password="$PASS" \
    --mysql-db=sbtest --tables=12 --table-size=100000 \
    --threads=4 --time=60 --report-interval=10 run

[ 10s ] thds: 4 tps: 703.87 lat (ms,95%): 9.91 err/s: 0.00   # before IO latency kicks in
[ 20s ] thds: 4 tps: 525.21 lat (ms,95%): 22.69 err/s: 0.00  # latency increasing
[ 30s ] thds: 4 tps: 459.99 lat (ms,95%): 26.20 err/s: 0.00
[ 40s ] thds: 4 tps: 358.60 lat (ms,95%): 34.33 err/s: 0.00
[ 50s ] thds: 4 tps: 238.70 lat (ms,95%): 48.34 err/s: 0.00  # degrading further
[ 60s ] thds: 4 tps: 104.10 lat (ms,95%): 81.48 err/s: 0.00  # heavily impacted

SQL statistics:
    transactions:                        23909  (398.42 per sec.)
    ignored errors:                      0      (0.00 per sec.)
```

During the experiment, the cluster stays `Ready` — no failover triggered:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h13m

NAME                                 READY   STATUS    RESTARTS   AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          19m     standby
pod/mysql-ha-cluster-1               2/2     Running   0          3h13m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          3h13m   primary

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
MEMBER_HOST                                           MEMBER_PORT  MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         3306         ONLINE        PRIMARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         3306         ONLINE        SECONDARY
```

Verify data integrity after removing the chaos:

```shell
➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-43193:1000001-1000048
pod-1: 65a93aae-...:1-43193:1000001-1000048
pod-2: 65a93aae-...:1-43193:1000001-1000048

➤ # Checksums — all match ✅
pod-0: sbtest1=2507068505, sbtest2=3482084518, sbtest3=994648857, sbtest4=2618915730
pod-1: sbtest1=2507068505, sbtest2=3482084518, sbtest3=994648857, sbtest4=2618915730
pod-2: sbtest1=2507068505, sbtest2=3482084518, sbtest3=994648857, sbtest4=2618915730
```

**Result: PASS** — IO latency severely degraded TPS (703 → 104, ~85% reduction) but the cluster stayed fully operational with zero errors and zero data loss. No failover was triggered — MySQL's InnoDB engine handles IO latency gracefully.

Clean up:

```shell
➤ kubectl delete -f tests/04-io-latency.yaml
iochaos.chaos-mesh.org "mysql-primary-io-latency" deleted
```

---

### Chaos#5: Network Latency (1s) Between Primary and Replicas

We inject 1-second network latency between the primary and all standby replicas. Group Replication uses Paxos consensus — every write must be acknowledged by the majority. With 1s latency on every packet, writes become extremely slow.

Save this yaml as `tests/05-network-latency.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-replication-latency
  namespace: chaos-mesh
spec:
  action: delay
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "mysql-ha-cluster"
        "kubedb.com/role": "standby"
  delay:
    latency: "1s"
    jitter: "50ms"
  duration: "10m"
  direction: both
```

**What this chaos does:** Adds 1-second delay (with 50ms jitter) to all network traffic between the primary and standby replicas. Paxos consensus requires majority acknowledgment for each transaction, so every write now takes at least 1 second to commit.

- **Expected behavior:**
  Every Paxos round-trip delayed by ≥1 s → TPS collapses (each commit waits for majority ack) → cluster stays `Ready` (no failover — group replication doesn't treat slow nodes as failed within `unreachable_majority_timeout`) → TPS recovers after chaos. Zero data loss.

- **Actual result:**
  TPS dropped from ~460 → 0.91 (99.8%). No failover, no errors. All 3 members stayed `ONLINE` throughout the 10-minute delay window. GTIDs and checksums match across all 3 nodes. **PASS.**

Apply the chaos and run sysbench:

```shell
➤ kubectl apply -f tests/05-network-latency.yaml
networkchaos.chaos-mesh.org/mysql-replication-latency created

➤ kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
    --mysql-host=mysql-ha-cluster --mysql-port=3306 \
    --mysql-user=root --mysql-password="$PASS" \
    --mysql-db=sbtest --tables=12 --table-size=100000 \
    --threads=4 --time=60 --report-interval=10 run

[ 10s ] thds: 4 tps: 0.80 lat (ms,95%): 8333.38 err/s: 0.00
[ 20s ] thds: 4 tps: 0.60 lat (ms,95%): 9624.59 err/s: 0.00
[ 30s ] thds: 4 tps: 0.80 lat (ms,95%): 4943.53 err/s: 0.00
[ 40s ] thds: 4 tps: 1.20 lat (ms,95%): 4055.23 err/s: 0.00
[ 50s ] thds: 4 tps: 0.80 lat (ms,95%): 4128.91 err/s: 0.00
[ 60s ] thds: 4 tps: 1.20 lat (ms,95%): 4517.90 err/s: 0.00

SQL statistics:
    transactions:                        58     (0.91 per sec.)
    ignored errors:                      0      (0.00 per sec.)
```

During the experiment, the cluster stays `Ready` — all 3 members remain ONLINE. No failover triggered:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h20m

NAME                                 READY   STATUS    RESTARTS   AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          26m     standby
pod/mysql-ha-cluster-1               2/2     Running   0          3h20m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          3h20m   primary

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+
```

Verify data integrity:

```shell
➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-43340:1000001-1000048
pod-1: 65a93aae-...:1-43340:1000001-1000048
pod-2: 65a93aae-...:1-43340:1000001-1000048

➤ # Checksums — all match ✅
pod-0: sbtest1=3906784400, sbtest2=822321605, sbtest3=970778000, sbtest4=1508996567
pod-1: sbtest1=3906784400, sbtest2=822321605, sbtest3=970778000, sbtest4=1508996567
pod-2: sbtest1=3906784400, sbtest2=822321605, sbtest3=970778000, sbtest4=1508996567
```

**Result: PASS** — 1-second network latency reduced TPS from ~460 to 0.91 (99.8% reduction) because every Paxos round-trip now takes >1 second. However, the cluster stayed fully operational — no failover, no errors, zero data loss.

Clean up:

```shell
➤ kubectl delete -f tests/05-network-latency.yaml
networkchaos.chaos-mesh.org "mysql-replication-latency" deleted
```

---

### Chaos#6: CPU Stress (98%) on Primary

We apply 98% CPU stress on the primary pod to test how MySQL handles extreme CPU pressure.

Save this yaml as `tests/06-cpu-stress.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: mysql-primary-cpu-stress
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  stressors:
    cpu:
      workers: 2
      load: 98
  duration: "5m"
```

**What this chaos does:** Consumes 98% of the CPU on the primary pod, leaving minimal CPU for MySQL query processing and Paxos consensus.

- **Expected behavior:**
  Primary CPU saturated at 98% → TPS drops significantly → cluster stays `Ready` (no failover — CPU contention slows MySQL but doesn't make it unresponsive to GR heartbeats) → recovers when chaos ends. Zero data loss.

- **Actual result:**
  TPS reduced from ~686 → ~212 (~69% drop). No failover, no errors. All 3 members stayed `ONLINE`. GTIDs and checksums match across all 3 nodes. **PASS.**

Apply the chaos and run sysbench:

```shell
➤ kubectl apply -f tests/06-cpu-stress.yaml
stresschaos.chaos-mesh.org/mysql-primary-cpu-stress created

➤ kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
    --mysql-host=mysql-ha-cluster --mysql-port=3306 \
    --mysql-user=root --mysql-password="$PASS" \
    --mysql-db=sbtest --tables=12 --table-size=100000 \
    --threads=4 --time=60 --report-interval=10 run

[ 10s ] thds: 4 tps: 685.57 lat (ms,95%): 10.65 err/s: 0.00  # before full stress
[ 20s ] thds: 4 tps: 566.01 lat (ms,95%): 18.28 err/s: 0.00
[ 30s ] thds: 4 tps: 448.70 lat (ms,95%): 27.66 err/s: 0.00  # degrading
[ 40s ] thds: 4 tps: 405.90 lat (ms,95%): 31.94 err/s: 0.00
[ 50s ] thds: 4 tps: 384.90 lat (ms,95%): 33.12 err/s: 0.00
[ 60s ] thds: 4 tps: 211.80 lat (ms,95%): 41.85 err/s: 0.00  # lowest point

SQL statistics:
    transactions:                        27033  (449.80 per sec.)
    ignored errors:                      0      (0.00 per sec.)
```

During the experiment, the cluster stays `Ready` — no failover triggered. All 3 members remain ONLINE:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h24m

NAME                                 READY   STATUS    RESTARTS   AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          29m     standby
pod/mysql-ha-cluster-1               2/2     Running   0          3h24m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          3h24m   primary

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+
```

Verify data integrity:

```shell
➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-70415:1000001-1000048
pod-1: 65a93aae-...:1-70415:1000001-1000048
pod-2: 65a93aae-...:1-70415:1000001-1000048

➤ # Checksums — all match ✅
pod-0: sbtest1=2711127555, sbtest2=1662235926, sbtest3=1102673461, sbtest4=1599950507
pod-1: sbtest1=2711127555, sbtest2=1662235926, sbtest3=1102673461, sbtest4=1599950507
pod-2: sbtest1=2711127555, sbtest2=1662235926, sbtest3=1102673461, sbtest4=1599950507
```

**Result: PASS** — 98% CPU stress reduced TPS from ~686 to ~212 (~69% reduction) but the cluster stayed `Ready` with all members ONLINE. Zero errors, zero data loss.

Clean up:

```shell
➤ kubectl delete -f tests/06-cpu-stress.yaml
stresschaos.chaos-mesh.org "mysql-primary-cpu-stress" deleted
```

---

### Chaos#7: Packet Loss (30%) Across Cluster

We inject 30% packet loss on all MySQL pods. This simulates an unreliable network — common in cloud environments with degraded network switches or cross-AZ communication issues.

Save this yaml as `tests/07-packet-loss.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-cluster-packet-loss
  namespace: chaos-mesh
spec:
  action: loss
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
  loss:
    loss: "30"
    correlation: "25"
  duration: "5m"
```

**What this chaos does:** Drops 30% of all network packets on every MySQL pod (with 25% correlation — dropped packets tend to cluster together). This affects both client connections and GR inter-node communication.

- **Expected behavior:**
  Packet drops disrupt GR heartbeats → some members go `UNREACHABLE` (transient, not failed) → TPS collapses but writes still eventually commit → cluster remains functional (primary stays primary) → after chaos, all members return to `ONLINE`. Zero data loss.

- **Actual result:**
  TPS dropped to 2.70 (99.4% reduction). Pod-1 observed `UNREACHABLE` during chaos, no failover triggered. After chaos removed, all 3 members returned to `ONLINE` immediately. GTIDs and checksums match across all 3 nodes. **PASS.**

Apply the chaos and run sysbench:

```shell
➤ kubectl apply -f tests/07-packet-loss.yaml
networkchaos.chaos-mesh.org/mysql-cluster-packet-loss created

➤ kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
    --mysql-host=mysql-ha-cluster --mysql-port=3306 \
    --mysql-user=root --mysql-password="$PASS" \
    --mysql-db=sbtest --tables=12 --table-size=100000 \
    --threads=4 --time=60 --report-interval=10 run

[ 10s ] thds: 4 tps: 3.50 lat (ms,95%): 1708.63 err/s: 0.00
[ 20s ] thds: 4 tps: 1.60 lat (ms,95%): 6026.41 err/s: 0.00
[ 30s ] thds: 4 tps: 2.70 lat (ms,95%): 3386.99 err/s: 0.00
[ 40s ] thds: 4 tps: 2.50 lat (ms,95%): 2680.11 err/s: 0.00
[ 50s ] thds: 4 tps: 2.30 lat (ms,95%): 3706.08 err/s: 0.00
[ 60s ] thds: 4 tps: 4.00 lat (ms,95%): 2009.23 err/s: 0.00

SQL statistics:
    transactions:                        170    (2.70 per sec.)
    ignored errors:                      0      (0.00 per sec.)
```

During the experiment, some members may appear as `UNREACHABLE` due to packet loss disrupting the GR heartbeat:

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | UNREACHABLE  | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+
```

> **Note:** The `UNREACHABLE` state means the GR heartbeat packets to pod-1 are being dropped. This is a transient state — pod-1 is still running, it just can't be reached reliably. Once packet loss is removed, it transitions back to `ONLINE`.

After removing the chaos, all members recover to ONLINE and the cluster returns to `Ready`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h28m

NAME                                 READY   STATUS    RESTARTS   AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          33m     standby
pod/mysql-ha-cluster-1               2/2     Running   0          3h27m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          3h27m   primary

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+
```

Verify data integrity:

```shell
➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-70618:1000001-1000048
pod-1: 65a93aae-...:1-70618:1000001-1000048
pod-2: 65a93aae-...:1-70618:1000001-1000048

➤ # Checksums — all match ✅
pod-0: sbtest1=2999610996, sbtest2=3490399001, sbtest3=2322312797, sbtest4=3764645300
pod-1: sbtest1=2999610996, sbtest2=3490399001, sbtest3=2322312797, sbtest4=3764645300
pod-2: sbtest1=2999610996, sbtest2=3490399001, sbtest3=2322312797, sbtest4=3764645300
```

**Result: PASS** — 30% packet loss reduced TPS to 2.70 (99.4% reduction) and caused `UNREACHABLE` member states, but zero data loss and zero errors. After packet loss removed, all members recovered to ONLINE immediately.

Clean up:

```shell
➤ kubectl delete -f tests/07-packet-loss.yaml
networkchaos.chaos-mesh.org "mysql-cluster-packet-loss" deleted
```

---

### Chaos#8: Combined Stress (Memory + CPU + Write Load)

This is the most aggressive test — we apply memory stress on the primary, CPU stress on all nodes, and sustained write load simultaneously. This simulates a "worst case" production incident where multiple things go wrong at once.

**Chaos YAMLs applied simultaneously:**
- `stress-memory-primary.yaml` — 1200MB memory stress on primary
- `stress-cpu-all.yaml` — 90% CPU stress on all 3 nodes

**What this chaos does:** Applies memory pressure + CPU exhaustion while the database is under active write load. The primary is likely to get OOMKilled.

- **Expected behavior:**
  Combined memory + CPU stress under sysbench load → primary OOMKilled → cluster `NotReady` during failover → new primary elected from standby → OOMKilled pod restarts and rejoins → cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Primary pod-2 OOMKilled (restart count 2). Cluster went `NotReady` briefly. Pod-1 elected as new primary. Pod-2 rejoined as standby within ~90s. All 3 members `ONLINE`, GTIDs match, checksums match. **PASS.**

Start sysbench load first (~1186 TPS baseline), then apply both stress experiments:

```shell
➤ kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
    --threads=8 --time=120 --report-interval=10 run
[ 10s ] thds: 8 tps: 1186.74 lat (ms,95%): 13.70 err/s: 0.00  # baseline under load

➤ kubectl apply -f tests/stress-memory-primary.yaml
stresschaos.chaos-mesh.org/mysql-primary-memory-stress created
➤ kubectl apply -f tests/stress-cpu-all.yaml
stresschaos.chaos-mesh.org/mysql-all-cpu-stress created
```

The primary pod (pod-2) gets OOMKilled under the combined pressure. Sysbench loses all connections:

```shell
FATAL: mysql_stmt_execute() returned error 2013 (Lost connection to MySQL server during query)
```

During recovery, the database goes `NotReady` and the OOMKilled pod restarts:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     NotReady   3h30m

NAME                                 READY   STATUS    RESTARTS      AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0             35m     standby
pod/mysql-ha-cluster-1               2/2     Running   0             3h30m   standby
pod/mysql-ha-cluster-2               2/2     Running   2 (10s ago)   3h30m   standby
```

After ~90 seconds, the cluster fully recovers:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h33m

NAME                                 READY   STATUS    RESTARTS        AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0               38m     standby
pod/mysql-ha-cluster-1               2/2     Running   0               3h33m   primary
pod/mysql-ha-cluster-2               2/2     Running   2 (3m ago)      3h33m   standby
```

pod-1 is now the primary. pod-2 (OOMKilled) rejoined as standby. Verify data integrity:

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-82950:1000001-1000052
pod-1: 65a93aae-...:1-82950:1000001-1000052
pod-2: 65a93aae-...:1-82950:1000001-1000052

➤ # Checksums — all match ✅
pod-0: sbtest1=4068722757, sbtest2=1583533506, sbtest3=1588922375, sbtest4=1810696374
pod-1: sbtest1=4068722757, sbtest2=1583533506, sbtest3=1588922375, sbtest4=1810696374
pod-2: sbtest1=4068722757, sbtest2=1583533506, sbtest3=1588922375, sbtest4=1810696374
```

**Result: PASS** — Combined memory + CPU + write load caused OOMKill on the primary. Failover completed automatically. The OOMKilled pod rejoined as standby. Zero data loss — all GTIDs and checksums match.

Clean up:

```shell
➤ kubectl delete -f tests/stress-memory-primary.yaml
➤ kubectl delete -f tests/stress-cpu-all.yaml
```

---

### Chaos#9: Full Cluster Kill (All 3 Pods)

The ultimate stress test — we force-delete all 3 MySQL pods simultaneously. The entire cluster goes down. Can it recover automatically?

**Method:** `kubectl delete pod --force --grace-period=0` on all 3 pods at once.

- **Expected behavior:**
  All 3 pods deleted → cluster `NotReady` (no primary) → coordinator detects full outage → identifies pod with highest GTID → bootstraps new cluster from it → other pods rejoin as standbys → cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Full cluster outage. Coordinator detected outage and elected pod-0 (highest GTID) as new primary. Cluster recovered in ~2 minutes. All 3 members `ONLINE`, GTIDs and checksums match across all 3 nodes. **PASS.**

Before running:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h38m

NAME                                 READY   STATUS    RESTARTS   AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          43m     standby
pod/mysql-ha-cluster-1               2/2     Running   0          3h38m   primary
pod/mysql-ha-cluster-2               2/2     Running   0          3h38m   standby
```

Kill all 3 pods simultaneously:

```shell
➤ kubectl delete pod mysql-ha-cluster-0 mysql-ha-cluster-1 mysql-ha-cluster-2 \
    -n demo --force --grace-period=0
Warning: Immediate deletion does not wait for confirmation that the running resource has been terminated.
pod "mysql-ha-cluster-0" force deleted
pod "mysql-ha-cluster-1" force deleted
pod "mysql-ha-cluster-2" force deleted
```

The database immediately goes `NotReady` — no primary available:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     NotReady   3h38m

NAME                                 READY   STATUS    RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          16s   standby
pod/mysql-ha-cluster-1               2/2     Running   0          14s   standby
pod/mysql-ha-cluster-2               2/2     Running   0          12s   standby
```

All 3 pods are restarting. The coordinator on one of the pods will detect a full outage, find the pod with the highest GTID, and bootstrap a new cluster from it. After ~2 minutes:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h42m

NAME                                 READY   STATUS    RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          4m    primary
pod/mysql-ha-cluster-1               2/2     Running   0          3m    standby
pod/mysql-ha-cluster-2               2/2     Running   0          3m    standby
```

Verify data integrity:

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-83046:1000001-1000052
pod-1: 65a93aae-...:1-83046:1000001-1000052
pod-2: 65a93aae-...:1-83046:1000001-1000052

➤ # Checksums — all match ✅
pod-0: sbtest1=4068722757, sbtest2=1583533506, sbtest3=1588922375, sbtest4=1810696374
pod-1: sbtest1=4068722757, sbtest2=1583533506, sbtest3=1588922375, sbtest4=1810696374
pod-2: sbtest1=4068722757, sbtest2=1583533506, sbtest3=1588922375, sbtest4=1810696374
```

**Result: PASS** — Complete cluster outage (all 3 pods killed simultaneously). The coordinator automatically detected the outage, elected the pod with highest GTID as new primary, and rebuilt the cluster. Recovery in ~2 minutes. Zero data loss.

---

### Chaos#10: OOMKill via Natural Load (90 JOINs + Writes)

Instead of using StressChaos, we try to trigger an OOMKill naturally by running 90 concurrent large JOIN queries (5-table cross-joins) across all 3 pods while sysbench writes are in-flight.

**Method:** Launch 90 heavy JOIN queries across all pods + 4-thread sysbench for 120s.

- **Expected behavior:**
  Heavy JOINs + concurrent writes push memory toward the 1.5 Gi limit → either (a) primary survives by spilling to temp tables, staying `Ready` throughout, or (b) OOMKill triggers and cluster auto-recovers via failover. Zero data loss either way.

- **Actual result:**
  MySQL 8.4.8 survived — no OOMKill, no pod restarts. 388 TPS sustained across the full 120s. Zero errors, GTIDs and checksums match. **PASS.** *(Note: same test triggers OOMKill on MySQL 9.6.0 due to different memory allocation — also recovers cleanly.)*

```shell
➤ # Launch 90 large JOINs + sysbench writes
SQL statistics:
    transactions:                        46889  (388.43 per sec.)
    ignored errors:                      0      (0.00 per sec.)
```

MySQL 8.4.8 **survived** — no OOMKill triggered. The 1.5Gi memory limit provides sufficient headroom. No pod restarts:

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          6m    primary
mysql-ha-cluster-1               2/2     Running   0          6m    standby
mysql-ha-cluster-2               2/2     Running   0          6m    standby
```

```shell
➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-129971:1000001-1000052
pod-1: 65a93aae-...:1-129971:1000001-1000052
pod-2: 65a93aae-...:1-129971:1000001-1000052

➤ # Checksums — all match ✅
pod-0: sbtest1=3941398607, sbtest2=2282875795, sbtest3=2179429078, sbtest4=2316165905
pod-1: sbtest1=3941398607, sbtest2=2282875795, sbtest3=2179429078, sbtest4=2316165905
pod-2: sbtest1=3941398607, sbtest2=2282875795, sbtest3=2179429078, sbtest4=2316165905
```

**Result: PASS** — MySQL 8.4.8 handles memory conservatively and did not OOMKill under 90 concurrent large JOINs. 388 TPS sustained. Zero errors, zero data loss.

> **Note:** The same test triggers OOMKill on MySQL 9.6.0 due to different memory allocation behavior. Both versions pass with zero data loss.

---

### Chaos#11: Scheduled Pod Kill (Every 1 Min, 3 Min Duration)

We schedule random pod kills every minute for 3 minutes — simulating repeated intermittent failures.

Save this yaml as `tests/11-scheduled-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: Schedule
metadata:
  name: mysql-scheduled-pod-kill
  namespace: chaos-mesh
spec:
  schedule: "*/1 * * * *"
  historyLimit: 5
  concurrencyPolicy: "Allow"
  type: "PodChaos"
  podChaos:
    action: pod-kill
    mode: one
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "mysql-ha-cluster"
```

**What this chaos does:** Kills a random MySQL pod every minute for 3 minutes. The cluster must repeatedly recover from pod failures.

- **Expected behavior:**
  Random pod killed each minute → rapid primary reelection when primary is hit, quick rejoin when standby is hit → between kills cluster returns to `Ready` → after schedule ends and enough time passes, all 3 members `ONLINE`. Zero data loss.

- **Actual result:**
  Multiple pods killed over 3-minute window. Cluster auto-recovered after each kill. After schedule removed, all 3 members `ONLINE`, pod-0 primary, pod-1/pod-2 standbys. GTIDs and checksums match. **PASS.**

```shell
➤ kubectl apply -f tests/11-scheduled-pod-kill.yaml
schedule.chaos-mesh.org/mysql-scheduled-pod-kill created
```

After 3 minutes of repeated kills, multiple pods have been restarted (note the different ages — pod-1 is 2m, pod-2 is 89s):

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS   AGE    ROLE
mysql-ha-cluster-0               2/2     Running   0          10m    primary
mysql-ha-cluster-1               2/2     Running   0          89s    standby
mysql-ha-cluster-2               2/2     Running   0          29s    standby
```

After deleting the schedule and waiting for full recovery:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h49m

NAME                                 READY   STATUS    RESTARTS   AGE     ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          11m     primary
pod/mysql-ha-cluster-1               2/2     Running   0          2m29s   standby
pod/mysql-ha-cluster-2               2/2     Running   0          89s     standby

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-130027:1000001-1000052
pod-1: 65a93aae-...:1-130027:1000001-1000052
pod-2: 65a93aae-...:1-130027:1000001-1000052

➤ # Checksums — all match ✅
pod-0: sbtest1=3941398607, sbtest2=2282875795, sbtest3=2179429078, sbtest4=2316165905
pod-1: sbtest1=3941398607, sbtest2=2282875795, sbtest3=2179429078, sbtest4=2316165905
pod-2: sbtest1=3941398607, sbtest2=2282875795, sbtest3=2179429078, sbtest4=2316165905
```

**Result: PASS** — Multiple pods killed on schedule, all auto-recovered. Zero data loss.

Clean up:

```shell
➤ kubectl delete schedule mysql-scheduled-pod-kill -n chaos-mesh
```

---

### Chaos#12: Degraded Failover (IO Latency + Pod Kill Workflow)

A complex workflow: first inject IO latency on the primary to degrade it, then kill the degraded primary while it's struggling. This simulates a cascading failure.

Save this yaml as `tests/12-degraded-failover.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: Workflow
metadata:
  name: mysql-degraded-failover-scenario
  namespace: chaos-mesh
spec:
  entry: start-degradation-and-kill
  templates:
    - name: start-degradation-and-kill
      templateType: Parallel
      children:
        - inject-io-latency
        - delayed-kill-sequence
    - name: inject-io-latency
      templateType: IOChaos
      deadline: "2m"
      ioChaos:
        action: latency
        mode: one
        selector:
          namespaces: ["demo"]
          labelSelectors:
            "app.kubernetes.io/instance": "mysql-ha-cluster"
            "kubedb.com/role": "primary"
        volumePath: "/var/lib/mysql"
        delay: "50ms"
        percent: 100
    - name: delayed-kill-sequence
      templateType: Serial
      children: [wait-30s, kill-primary-pod]
    - name: wait-30s
      templateType: Suspend
      deadline: "30s"
    - name: kill-primary-pod
      templateType: PodChaos
      deadline: "1m"
      podChaos:
        action: pod-kill
        mode: one
        selector:
          namespaces: ["demo"]
          labelSelectors:
            "app.kubernetes.io/instance": "mysql-ha-cluster"
            "kubedb.com/role": "primary"
```

**What this chaos does:** Runs IO latency + pod kill in parallel. IO latency starts first, then after 30s the primary pod is killed while it's already degraded.

- **Expected behavior:**
  IO latency slows primary → after 30s the degraded primary is killed → failover elects healthy standby as new primary → cluster `Critical` → killed pod rejoins → cluster returns to `Ready`. Despite cascading fault, zero data loss.

- **Actual result:**
  Sysbench saw slow writes, then lost connection at ~30s when primary killed. Pod-2 elected as new primary, pod-1 (old primary) rejoined as standby. Cluster returned to `Ready`. GTIDs and checksums match across all 3 nodes. **PASS.**

Apply and run sysbench:

```shell
➤ kubectl apply -f tests/12-degraded-failover.yaml
workflow.chaos-mesh.org/mysql-degraded-failover-scenario created
```

Sysbench shows slow writes during IO latency, then loses connection when pod is killed at ~30s:

```shell
[ 10s ] thds: 4 tps: 3.50 lat (ms,95%): 831.46 err/s: 0.00    # IO latency active
[ 20s ] thds: 4 tps: 2.50 lat (ms,95%): 11523.48 err/s: 0.00  # severely degraded
[ 30s ] thds: 4 tps: 3.80 lat (ms,95%): 1235.62 err/s: 0.00
FATAL: Lost connection to MySQL server during query                # pod killed at ~30s
```

After the workflow completes, the cluster recovers. Failover to pod-2:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h51m

NAME                                 READY   STATUS    RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          50s   standby
pod/mysql-ha-cluster-1               2/2     Running   0          4m    standby
pod/mysql-ha-cluster-2               2/2     Running   0          3m    primary

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-130155:1000001-1000054
pod-1: 65a93aae-...:1-130155:1000001-1000054
pod-2: 65a93aae-...:1-130155:1000001-1000054

➤ # Checksums — all match ✅
pod-0: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
pod-1: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
pod-2: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
```

**Result: PASS** — Cascading failure (IO latency + pod kill) handled gracefully. Failover completed automatically. Zero data loss.

Clean up:

```shell
➤ kubectl delete workflow mysql-degraded-failover-scenario -n chaos-mesh
```

---

### Chaos#13: Double Primary Kill

Kill the primary, wait for new election, then immediately kill the new primary. Tests whether the cluster survives two consecutive leader failures.

- **Expected behavior:**
  First primary killed → new primary elected → second kill (of newly elected primary) → third election → surviving standby becomes primary → killed pods rejoin → cluster returns to `Ready`. Zero data loss despite rapid leader churn.

- **Actual result:**
  Pod-2 killed → pod-1 elected → pod-1 killed within ~15s → pod-2 re-elected (after its restart) as third primary. All pods rejoined. All 3 members `ONLINE`, GTIDs and checksums match. **PASS.**

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          2m    standby
mysql-ha-cluster-1               2/2     Running   0          5m    standby
mysql-ha-cluster-2               2/2     Running   0          4m    primary

➤ # Kill first primary (pod-2)
kubectl delete pod mysql-ha-cluster-2 -n demo --force --grace-period=0
pod "mysql-ha-cluster-2" force deleted
```

After 15 seconds, pod-1 is elected as new primary:

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          2m    standby
mysql-ha-cluster-1               2/2     Running   0          5m    primary   ← new primary
mysql-ha-cluster-2               2/2     Running   0          15s   standby

➤ # Kill second primary (pod-1) immediately
kubectl delete pod mysql-ha-cluster-1 -n demo --force --grace-period=0
pod "mysql-ha-cluster-1" force deleted
```

Database goes `Critical` — two primaries killed in quick succession:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Critical   3h53m

NAME                                 READY   STATUS    RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          2m    standby
pod/mysql-ha-cluster-1               2/2     Running   0          17s   standby
pod/mysql-ha-cluster-2               2/2     Running   0          32s   primary
```

pod-2 was re-elected as third primary. After full recovery:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h55m

NAME                                 READY   STATUS    RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          4m    standby
pod/mysql-ha-cluster-1               2/2     Running   0          110s  standby
pod/mysql-ha-cluster-2               2/2     Running   0          2m    primary

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-130185:1000001-1000056
pod-1: 65a93aae-...:1-130185:1000001-1000056
pod-2: 65a93aae-...:1-130185:1000001-1000056

➤ # Checksums — all match ✅
pod-0: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
pod-1: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
pod-2: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
```

**Result: PASS** — Two consecutive primary kills survived. The cluster elected a third primary and recovered fully. Zero data loss.

---

### Chaos#14: Rolling Restart (0→1→2)

Simulate a rolling upgrade — delete each pod sequentially with 40-second gaps. Tests graceful rolling restart behavior.

- **Expected behavior:**
  Pods deleted one at a time (0 → 1 → 2) with 40s gap between kills → each pod rejoins before the next is killed → when the primary is hit, a quick failover occurs → cluster returns to `Ready` between each step. Zero data loss.

- **Actual result:**
  All 3 pods restarted in sequence. Single failover when primary (pod-2) was killed. Each pod rejoined within ~40s. GTIDs and checksums match across all 3 nodes. **PASS.**

```shell
➤ # Delete pod-0 (standby)
kubectl delete pod mysql-ha-cluster-0 -n demo --force --grace-period=0

➤ # 40s later — pod-0 recovered, pod-2 still primary
kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          40s   standby
mysql-ha-cluster-1               2/2     Running   0          3m    standby
mysql-ha-cluster-2               2/2     Running   0          3m    primary

➤ # Delete pod-1 (standby)
kubectl delete pod mysql-ha-cluster-1 -n demo --force --grace-period=0

➤ # 40s later — pod-1 recovered, pod-2 still primary
kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          80s   standby
mysql-ha-cluster-1               2/2     Running   0          40s   standby
mysql-ha-cluster-2               2/2     Running   0          4m    primary

➤ # Delete pod-2 (primary!) — triggers failover
kubectl delete pod mysql-ha-cluster-2 -n demo --force --grace-period=0

➤ # 60s later — pod-1 elected as new primary
kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h58m

NAME                                 READY   STATUS    RESTARTS   AGE    ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          2m     standby
pod/mysql-ha-cluster-1               2/2     Running   0          100s   primary
pod/mysql-ha-cluster-2               2/2     Running   0          60s    standby

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-130225:1000001-1000058
pod-1: 65a93aae-...:1-130225:1000001-1000058
pod-2: 65a93aae-...:1-130225:1000001-1000058

➤ # Checksums — all match ✅
pod-0: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
pod-1: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
pod-2: sbtest1=740913138, sbtest2=3199164728, sbtest3=1207551779, sbtest4=3950955642
```

**Result: PASS** — All 3 pods deleted sequentially. Failover triggered only when primary was deleted. Each pod recovered and rejoined within ~30s. Zero data loss.

---

### Chaos#15: Coordinator Crash

Kill only the mysql-coordinator sidecar container on the primary pod, leaving the MySQL process running. Tests whether MySQL GR operates independently of the coordinator.

- **Expected behavior:**
  Coordinator sidecar killed, MySQL process untouched → coordinator container restarted by Kubernetes → MySQL keeps serving writes throughout → no failover, no TPS drop, cluster stays `Ready`. Zero data loss.

- **Actual result:**
  Coordinator restarted automatically. MySQL stayed primary, no role change. Sysbench ran at 691 TPS (baseline) throughout. All 3 members `ONLINE`, GTIDs match. **PASS** — confirms the coordinator is a management-layer sidecar; GR runs independently.

```shell
➤ # Kill coordinator process (PID 1) on the primary
kubectl exec -n demo mysql-ha-cluster-1 -c mysql-coordinator -- kill 1
```

The coordinator container restarts automatically. MySQL stays running — no failover, no interruption. Database stays `Ready`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    3h59m

NAME                                 READY   STATUS    RESTARTS   AGE    ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          3m     standby
pod/mysql-ha-cluster-1               2/2     Running   0          2m     primary   ← still primary
pod/mysql-ha-cluster-2               2/2     Running   0          2m     standby
```

Writes work immediately at full speed (691 TPS):

```shell
➤ sysbench oltp_write_only --threads=4 --time=10 run
[ 5s ] thds: 4 tps: 678.95 lat (ms,95%): 10.84 err/s: 0.00
[10s ] thds: 4 tps: 703.00 lat (ms,95%): 10.65 err/s: 0.00
    transactions:                        6914   (691.13 per sec.)

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-137155:1000001-1000058
pod-1: 65a93aae-...:1-137155:1000001-1000058
pod-2: 65a93aae-...:1-137155:1000001-1000058
```

**Result: PASS** — Coordinator crash has zero impact on MySQL. No failover, no write interruption, 691 TPS (full speed). The coordinator is a management layer — MySQL GR operates independently.

---

### Chaos#16: Long Network Partition (10 min)

Save this yaml as `tests/16-network-partition-long.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-primary-network-partition-long
  namespace: chaos-mesh
spec:
  action: partition
  mode: one
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces: [demo]
      labelSelectors:
        "kubedb.com/role": "standby"
  direction: both
  duration: "10m"
```

**What this chaos does:** Isolates the primary from all replicas for 10 minutes — 5x longer than the standard test.

- **Expected behavior:**
  Primary isolated for 10 min → failover in ~20s → cluster stays `Critical` (isolated node unreachable) throughout the 10 min → when partition lifts, isolated node rejoins via GR distributed recovery → cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Failover happened in ~20s. Pod-1 went `UNREACHABLE` (restarted once during the 10-min window). After partition removed, pod-1 rejoined cleanly. Cluster returned to `Ready`. GTIDs and checksums match across all 3 nodes. **PASS.**

```shell
➤ kubectl apply -f tests/16-network-partition-long.yaml
networkchaos.chaos-mesh.org/mysql-primary-network-partition-long created
```

Failover happens within ~20 seconds. During the 10-minute partition, the cluster is `Critical` with only 2 members visible:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE    ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Critical   4h1m

NAME                                 READY   STATUS    RESTARTS      AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0             4m    standby
pod/mysql-ha-cluster-1               2/2     Running   1 (83s ago)   4m    standby
pod/mysql-ha-cluster-2               2/2     Running   0             3m    primary

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+
```

After 10 minutes, the isolated node rejoins and the cluster returns to `Ready`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    4h13m

NAME                                 READY   STATUS    RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          17m   standby
pod/mysql-ha-cluster-1               2/2     Running   0          16m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          15m   primary

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-137173:1000001-1000198
pod-1: 65a93aae-...:1-137173:1000001-1000198
pod-2: 65a93aae-...:1-137173:1000001-1000198

➤ # Checksums — all match ✅
pod-0: sbtest1=4029255859, sbtest2=3385379889, sbtest3=3777442529, sbtest4=914443400
pod-1: sbtest1=4029255859, sbtest2=3385379889, sbtest3=3777442529, sbtest4=914443400
pod-2: sbtest1=4029255859, sbtest2=3385379889, sbtest3=3777442529, sbtest4=914443400
```

**Result: PASS** — 10-minute partition survived. Isolated node rejoined cleanly via GR distributed recovery. Zero data loss.

---

### Chaos#17: DNS Failure on Primary

Block all DNS resolution on the primary for 3 minutes. GR uses hostnames for communication.

- **Expected behavior:**
  DNS resolution fails on primary → existing TCP connections remain open (already resolved) → writes continue with modest TPS drop → no failover (heartbeats go over existing sockets) → when DNS recovers, TPS returns to baseline. Zero data loss.

- **Actual result:**
  TPS dropped from ~720 → ~360 (~33% impact). No failover, no errors. GR heartbeats kept working over established sockets. GTIDs match across all 3 nodes. **PASS.**

Save this yaml as `tests/17-dns-error.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: DNSChaos
metadata:
  name: mysql-dns-error-primary
  namespace: chaos-mesh
spec:
  action: error
  mode: one
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  duration: "3m"
```

```shell
➤ kubectl apply -f tests/17-dns-error.yaml
dnschaos.chaos-mesh.org/mysql-dns-error-primary created

➤ sysbench oltp_write_only --threads=4 --time=60 run
[ 10s ] thds: 4 tps: 720.47 lat (ms,95%): 9.56 err/s: 0.00
[ 20s ] thds: 4 tps: 560.00 lat (ms,95%): 17.95 err/s: 0.00
[ 30s ] thds: 4 tps: 449.90 lat (ms,95%): 26.20 err/s: 0.00
[ 40s ] thds: 4 tps: 411.10 lat (ms,95%): 29.72 err/s: 0.00
[ 50s ] thds: 4 tps: 417.40 lat (ms,95%): 32.53 err/s: 0.00
[ 60s ] thds: 4 tps: 360.00 lat (ms,95%): 36.24 err/s: 0.00

    transactions:                        29193  (486.49 per sec.)
    ignored errors:                      0      (0.00 per sec.)

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-166380:1000001-1000198
pod-1: 65a93aae-...:1-166380:1000001-1000198
pod-2: 65a93aae-...:1-166380:1000001-1000198
```

**Result: PASS** — DNS failure reduced TPS ~33% (720→360) but no failover, no errors. Existing TCP connections between GR members stayed open. Zero data loss.

```shell
➤ kubectl delete -f tests/17-dns-error.yaml
```

---

### Chaos#18: PVC Delete + Pod Kill (Full Data Rebuild)

Completely destroy a node's data — delete both the pod and its PVC. The node must rebuild from scratch using the CLONE plugin.

- **Expected behavior:**
  Pod and PVC deleted → new pod provisioned with fresh empty PVC → pod enters `Init:0/1` while storage binds → operator initiates CLONE from a healthy primary → data fully restored → node rejoins as standby → cluster returns to `Ready`. Zero data loss on remaining nodes.

- **Actual result:**
  Pod-0 destroyed (pod + PVC). Pod reprovisioned, CLONE plugin rebuilt the datadir from primary. Cluster `Critical` during rebuild (~2 min), then `Ready` with all 3 members `ONLINE`. GTIDs and checksums match across all 3 nodes. **PASS.**

```shell
➤ kubectl delete pod mysql-ha-cluster-0 -n demo --force --grace-period=0
pod "mysql-ha-cluster-0" force deleted
➤ kubectl delete pvc data-mysql-ha-cluster-0 -n demo
persistentvolumeclaim "data-mysql-ha-cluster-0" deleted
```

The pod enters `Init:0/1` while waiting for a new PVC:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Critical   4h26m

NAME                                 READY   STATUS     RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               0/2     Init:0/1   0          9m
pod/mysql-ha-cluster-1               2/2     Running    0          29m   standby
pod/mysql-ha-cluster-2               2/2     Running    0          28m   primary

➤ kubectl get pvc -n demo
NAME                      STATUS    AGE
data-mysql-ha-cluster-0   Pending   45s    ← new PVC being provisioned
data-mysql-ha-cluster-1   Bound     4h27m
data-mysql-ha-cluster-2   Bound     4h27m
```

Once the new PVC is bound, the CLONE plugin copies a full data snapshot from a donor. After ~2 minutes, the node is fully recovered:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    4h30m

NAME                                 READY   STATUS    RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          99s   standby   ← rebuilt from scratch
pod/mysql-ha-cluster-1               2/2     Running   0          33m   standby
pod/mysql-ha-cluster-2               2/2     Running   0          33m   primary

➤ kubectl get pvc -n demo
NAME                      STATUS   CAPACITY   AGE
data-mysql-ha-cluster-0   Bound    2Gi        3m     ← brand new PVC
data-mysql-ha-cluster-1   Bound    2Gi        4h30m
data-mysql-ha-cluster-2   Bound    2Gi        4h30m

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+

➤ # GTIDs — all match ✅ (pod-0 has identical GTIDs after CLONE)
pod-0: 65a93aae-...:1-166554:1000001-1000198
pod-1: 65a93aae-...:1-166554:1000001-1000198
pod-2: 65a93aae-...:1-166554:1000001-1000198

➤ # Checksums — all match ✅
pod-0: sbtest1=2710640468, sbtest2=1851529474, sbtest3=809748038, sbtest4=3029682236
pod-1: sbtest1=2710640468, sbtest2=1851529474, sbtest3=809748038, sbtest4=3029682236
pod-2: sbtest1=2710640468, sbtest2=1851529474, sbtest3=809748038, sbtest4=3029682236
```

**Result: PASS** — Complete data destruction (PVC deleted). The CLONE plugin rebuilt the node from scratch with identical data. Zero data loss. This is the ultimate recovery test — MySQL 8.0+ handles it fully automatically.

---

### Chaos#19: IO Fault (EIO Errors, 50%)

Inject actual I/O read/write errors on 50% of disk operations. Unlike IO latency which slows things down, IO faults cause operations to fail — simulating a failing disk.

Save this yaml as `tests/19-io-fault.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: mysql-primary-io-fault
  namespace: chaos-mesh
spec:
  action: fault
  mode: one
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  volumePath: "/var/lib/mysql"
  path: "/**"
  errno: 5
  percent: 50
  duration: "3m"
```

**What this chaos does:** Returns EIO (errno 5) on 50% of all disk I/O operations on the primary's MySQL data directory. InnoDB cannot write to disk reliably.

- **Expected behavior:**
  InnoDB hits real I/O errors → MySQL crashes on primary → GR expels the unreachable member → standby elected as new primary → crashed pod restarts, InnoDB crash recovery repairs the datadir → pod rejoins as standby → cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Primary pod-2 crashed (error 2013 on sysbench). Pod-2 observed `UNREACHABLE`, failover to pod-1. After chaos removed, InnoDB crash recovery + GR distributed recovery brought pod-2 back as standby within ~90s. All 3 members `ONLINE`, GTIDs match. **PASS.**

```shell
➤ kubectl apply -f tests/19-io-fault.yaml
iochaos.chaos-mesh.org/mysql-primary-io-fault created

FATAL: Lost connection to MySQL server during query   # MySQL crashed from IO errors
```

The primary (pod-2) crashes. During recovery, it appears as `UNREACHABLE`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     NotReady   4h32m

➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
    FROM performance_schema.replication_group_members;
+-----------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------------+-------------+--------------+-------------+
| mysql-ha-cluster-2.mysql-ha-cluster-pods.demo |        3306 | UNREACHABLE  | PRIMARY     |
| mysql-ha-cluster-1.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
| mysql-ha-cluster-0.mysql-ha-cluster-pods.demo |        3306 | ONLINE       | SECONDARY   |
+-----------------------------------------------+-------------+--------------+-------------+
```

After removing the chaos and restarting the crashed pod, InnoDB crash recovery repairs the data and the node rejoins:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE     ROLE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    4h37m

NAME                                 READY   STATUS    RESTARTS   AGE   ROLE
pod/mysql-ha-cluster-0               2/2     Running   0          7m    standby
pod/mysql-ha-cluster-1               2/2     Running   0          40m   primary
pod/mysql-ha-cluster-2               2/2     Running   0          90s   standby

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-166608:1000001-1000236
pod-1: 65a93aae-...:1-166608:1000001-1000236
pod-2: 65a93aae-...:1-166608:1000001-1000236
```

**Result: PASS** — 50% IO errors crashed MySQL on the primary, triggering failover. InnoDB crash recovery + GR distributed recovery handled it with zero data loss.

```shell
➤ kubectl delete -f tests/19-io-fault.yaml
```

---

### Chaos#20: Clock Skew (-5 min)

Shift the primary's system clock back by 5 minutes. Tests whether clock drift breaks Paxos consensus.

- **Expected behavior:**
  Primary's wall clock shifted -5 min → MySQL's logical/GTID ordering unaffected (GR uses logical clocks, not wall-clock) → sysbench may see modest TPS drop → no failover, no errors. Zero data loss.

- **Actual result:**
  TPS dropped from 618 → 359 (~39% drop, largely from sysbench/query timing noise). No failover, no errors. GTIDs match across all 3 nodes. **PASS** — confirms GR Paxos uses logical clocks.

Save this yaml as `tests/20-clock-skew.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: TimeChaos
metadata:
  name: mysql-primary-clock-skew
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  timeOffset: "-5m"
  duration: "3m"
```

```shell
➤ kubectl apply -f tests/20-clock-skew.yaml
timechaos.chaos-mesh.org/mysql-primary-clock-skew created

➤ sysbench oltp_write_only --threads=4 --time=60 run
[ 10s ] thds: 4 tps: 617.97 lat (ms,95%): 11.87 err/s: 0.00
[ 20s ] thds: 4 tps: 477.40 lat (ms,95%): 22.69 err/s: 0.00
[ 30s ] thds: 4 tps: 255.50 lat (ms,95%): 34.33 err/s: 0.00  # degraded
[ 40s ] thds: 4 tps: 429.50 lat (ms,95%): 27.17 err/s: 0.00
[ 50s ] thds: 4 tps: 384.10 lat (ms,95%): 32.53 err/s: 0.00
[ 60s ] thds: 4 tps: 358.90 lat (ms,95%): 34.95 err/s: 0.00

    transactions:                        25238  (420.56 per sec.)
    ignored errors:                      0      (0.00 per sec.)

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-191908:1000001-1000236
pod-1: 65a93aae-...:1-191908:1000001-1000236
pod-2: 65a93aae-...:1-191908:1000001-1000236
```

**Result: PASS** — Clock skew (-5 min) reduced TPS ~39% (618→359) but no failover, no errors. GR's Paxos protocol uses logical clocks, not wall-clock time. Zero data loss.

```shell
➤ kubectl delete -f tests/20-clock-skew.yaml
```

---

### Chaos#21: Bandwidth Throttle (1mbps)

Limit the primary's outbound network bandwidth to 1mbps — simulating degraded cross-AZ network.

- **Expected behavior:**
  Primary's outbound bandwidth throttled to 1 Mbps → Paxos writes back up behind limited capacity → TPS drops substantially but commits still succeed → no failover (heartbeats fit within the bandwidth budget) → TPS recovers after chaos. Zero data loss.

- **Actual result:**
  TPS dropped from 618 → 136 (~80% drop). Zero errors, no failover. Cluster stayed `Ready` throughout. GTIDs match across all 3 nodes. **PASS.**

Save this yaml as `tests/21-bandwidth-throttle.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-bandwidth-throttle
  namespace: chaos-mesh
spec:
  action: bandwidth
  mode: one
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  bandwidth:
    rate: "1mbps"
    limit: 20971520
    buffer: 10000
  duration: "3m"
```

```shell
➤ kubectl apply -f tests/21-bandwidth-throttle.yaml
networkchaos.chaos-mesh.org/mysql-bandwidth-throttle created

➤ sysbench oltp_write_only --threads=4 --time=60 run
[ 10s ] thds: 4 tps: 157.99 lat (ms,95%): 30.81 err/s: 0.00
[ 20s ] thds: 4 tps: 158.00 lat (ms,95%): 30.81 err/s: 0.00
[ 30s ] thds: 4 tps: 157.20 lat (ms,95%): 30.81 err/s: 0.00
[ 40s ] thds: 4 tps: 78.60  lat (ms,95%): 99.33 err/s: 0.00
[ 50s ] thds: 4 tps: 120.50 lat (ms,95%): 86.00 err/s: 0.00
[ 60s ] thds: 4 tps: 144.80 lat (ms,95%): 44.98 err/s: 0.00

    transactions:                        8175   (136.22 per sec.)
    ignored errors:                      0      (0.00 per sec.)

➤ # GTIDs — all match ✅
pod-0: 65a93aae-...:1-200129:1000001-1000236
pod-1: 65a93aae-...:1-200129:1000001-1000236
pod-2: 65a93aae-...:1-200129:1000001-1000236
```

**Result: PASS** — Bandwidth throttle to 1mbps reduced TPS ~80% (618→136) but zero errors, no failover. The cluster stays completely stable under bandwidth constraints. Zero data loss.

```shell
➤ kubectl delete -f tests/21-bandwidth-throttle.yaml
```

### Chaos#22: Pod Failure on Primary (5 min)

Inject a 5-minute `pod-failure` fault into the current primary pod — Chaos Mesh keeps the container in a failed state for the entire duration before clearing the fault. Unlike `pod-kill` (Chaos#1), the pod is not deleted and rescheduled — it stays in place but its container is unavailable, exercising long-duration primary unreachability and the operator's clone-vs-incremental rejoin logic.

- **Expected behavior:**
  Primary container becomes unavailable → cluster transitions `Ready` → `NotReady` → Group Replication elects a new primary and the operator marks the failed pod unhealthy → state moves to `Critical`. When the 5-minute fault clears, the old primary container restarts, rejoins as `SECONDARY`, and the cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Failover completed in 8 s. Cluster transitioned `Ready` → `NotReady` (+5s) → `Critical` (+21s). Old primary auto-recovered after the 5-min duration and rejoined via incremental recovery; cluster reached `Ready` ~9–13 min after chaos cleared. Zero errors after sysbench reconnect, zero data loss, zero errant GTIDs. **PASS.**

Save this yaml as `tests/22-pod-failure.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: mysql-pod-failure-primary
  namespace: demo
spec:
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  mode: all
  action: pod-failure
  duration: 5m
```

```shell
➤ kubectl apply -f tests/22-pod-failure.yaml
podchaos.chaos-mesh.org/mysql-pod-failure-primary created

➤ # During chaos — pod-1 promoted, pod-2 stuck failed
➤ kubectl get mysql -n demo
NAME               VERSION   STATUS     AGE
mysql-ha-cluster   8.4.8     Critical   16h

➤ kubectl get pods -n demo -L kubedb.com/role
NAME                 READY   STATUS    RESTARTS   ROLE
mysql-ha-cluster-0   2/2     Running   2          standby
mysql-ha-cluster-1   2/2     Running   2          primary    # promoted (was standby)
mysql-ha-cluster-2   2/2     Running   4          standby    # under chaos

➤ # sysbench during failover
[ 10s ] thds: 8 tps: 1012.65 qps: 6079.89 lat (ms,95%): 16.41 err/s: 0.00
[ 20s ] thds: 8 tps:  577.40 qps: 3464.41 lat (ms,95%): 37.56 err/s: 0.00
[ 30s ] thds: 8 tps:  399.20 qps: 2394.80 lat (ms,95%): 45.79 err/s: 0.00
[ 40s ] thds: 8 tps:  314.70 qps: 1888.58 lat (ms,95%): 64.47 err/s: 0.00
[ 60s ] thds: 8 tps: 1007.99 qps: 6048.34 lat (ms,95%): 21.11 err/s: 0.00

➤ # After 5-min duration — chaos auto-recovered, pod-2 rejoins
➤ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
mysql-ha-cluster-2.…  ONLINE  SECONDARY
mysql-ha-cluster-1.…  ONLINE  PRIMARY
mysql-ha-cluster-0.…  ONLINE  SECONDARY

➤ kubectl get mysql -n demo
NAME               VERSION   STATUS   AGE
mysql-ha-cluster   8.4.8     Ready    16h

➤ SELECT COUNT(*) FROM sbtest.sbtest1;   # 100000 — intact
```

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 11:39:00 | — | Pre-chaos baseline (pod-2 = primary, all 3 ONLINE) | `Ready` |
| 11:39:21 | 0s | `PodChaos` applied (pod-2 targeted) | `Ready` |
| 11:39:26 | +5s | Primary unreachable detected | `NotReady` |
| 11:39:29 | +8s | Group Replication promotes pod-1 → PRIMARY | `NotReady` |
| 11:39:42 | +21s | Operator marks pod-2 unhealthy | `Critical` |
| 11:44:21 | +5m00s | Chaos auto-recovered, pod-2 container restarts | `Critical` |
| 11:45:40 | +6m19s | pod-2 in `RECOVERING` (incremental catch-up) | `Critical` |
| ~11:48–52 | ~+9–13m | pod-2 reaches `ONLINE` `SECONDARY` | `Ready` |

**Result: PASS** — Group Replication failed over in 8 s, sysbench recovered to ~1000 TPS on the new primary, and the chaos-impacted pod rejoined the group automatically once the fault expired. Zero data loss, zero errant GTIDs across all 3 nodes.

```shell
➤ kubectl delete -f tests/22-pod-failure.yaml
```

### Chaos#23: Continuous OOM Stress on Primary (15× 30s loop)

Inject a tight loop of memory-stress chaos against the current primary to mimic a runaway query / leak that keeps OOM-killing `mysqld` while sysbench keeps writing. Each iteration allocates 100 GB with `oomScoreAdj=-1000` so the kernel kills the MySQL container almost immediately. Fifteen iterations were applied back-to-back (≈ 8 minutes of effective OOM pressure).

- **Expected behavior:**
  Primary container is OOMKilled repeatedly → cluster transitions `Ready` → `Critical` once one replica goes unhealthy → after Group Replication notices the primary is gone, fail over to one of the secondaries → original primary enters `CrashLoopBackOff` until OOM pressure clears → afterwards the pod rejoins the group, GTIDs reconcile, and the cluster returns to `Ready`. Zero data loss expected.

- **Actual result:**
  Cluster transitioned `Ready` → `Critical` (+1m52s, pod-2 1/2 not ready) → `NotReady` briefly (+2m02s) → pod-1 promoted `PRIMARY` (+2m14s) → `Critical` until pod-2 stabilized. Pod-2 hit 11 restarts (`CrashLoopBackOff` while OOM loop continued) and then went into `RECOVERING` for ~9 minutes while the coordinator emitted the errant-GTID warning. Pod-2 finally reached `ONLINE SECONDARY` and the cluster returned to `Ready` at +12 minutes. **Final GTIDs match exactly on all 3 nodes.** **PASS.**

Save this yaml as `tests/23-oom-primary.yaml` (replace `<primary-pod-name>` with the actual current primary):

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: mysql-oom-primary
  namespace: demo
spec:
  selector:
    namespaces: [demo]
    labelSelectors:
      "statefulset.kubernetes.io/pod-name": "<primary-pod-name>"
  mode: all
  stressors:
    memory:
      workers: 1
      size: "100GB"
      oomScoreAdj: -1000
  duration: 30s
```

Run a tight loop so the OOM keeps recurring (otherwise a single 30 s pulse may be absorbed before sysbench notices):

```bash
PRIMARY=$(kubectl get pods -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')

for i in $(seq 1 15); do
  sed "s/mysql-oom-primary/mysql-oom-primary-${i}/; s/<primary-pod-name>/${PRIMARY}/" \
    tests/23-oom-primary.yaml | kubectl apply -f -
  sleep 2
done
```

```shell
➤ # During chaos — pod-2 OOMKilled, pod-1 promoted
➤ kubectl get mysql -n demo
NAME               VERSION   STATUS     AGE
mysql-ha-cluster   8.4.8     Critical   17h

➤ kubectl get pods -n demo -L kubedb.com/role
NAME                 READY   STATUS             RESTARTS   ROLE
mysql-ha-cluster-0   2/2     Running            2          standby
mysql-ha-cluster-1   2/2     Running            0          primary    # promoted
mysql-ha-cluster-2   1/2     CrashLoopBackOff   8          standby

➤ # sysbench during OOM loop
[ 10s ] thds: 8 tps: 1195.65 qps: 7177.51 lat (ms,95%): 12.98 err/s: 0.00
[ 60s ] thds: 8 tps:  201.10 qps: 1207.20 lat (ms,95%): 110.66 err/s: 0.00
[ 80s ] thds: 8 tps:   78.10 qps:  469.00 lat (ms,95%): 434.83 err/s: 0.00
[120s ] thds: 8 tps:    0.00 qps:    0.00 lat (ms,95%):   0.00 err/s: 0.00
[140s ] thds: 8 tps:    0.00 qps:    0.00 lat (ms,95%):   0.00 reconn/s: 0.60
[150s ] thds: 8 tps:  633.48 qps: 3804.70 lat (ms,95%):  20.74 err/s: 0.00 reconn/s: 0.20

➤ # Coordinator surfaces errant GTIDs from the partial commits before OOM
➤ kubectl logs -n demo mysql-ha-cluster-2 -c mysql-coordinator | grep -i errant
WARNING: instance mysql-ha-cluster-2 has extra GTIDs not on primary:
  b5a48606-...:384291-384301 (these will be lost if clone proceeds)
instance mysql-ha-cluster-2 has extra transactions not present on the primary
  — waiting for manual approval to use clone so sync group
to approve clone, create the file:
  kubectl exec -n demo mysql-ha-cluster-2 -c mysql -- touch /scripts/approve-clone

➤ # After OOM pressure stops, pod-2 rejoins via GR distributed recovery
➤ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
mysql-ha-cluster-2.…  ONLINE  SECONDARY
mysql-ha-cluster-1.…  ONLINE  PRIMARY
mysql-ha-cluster-0.…  ONLINE  SECONDARY

➤ # GTIDs match on all 3 nodes
pod-0: b5a48606-…:1-421633:1000001-1353570
pod-1: b5a48606-…:1-421633:1000001-1353570
pod-2: b5a48606-…:1-421633:1000001-1353570

➤ kubectl get mysql -n demo
NAME               VERSION   STATUS   AGE
mysql-ha-cluster   8.4.8     Ready    17h
```

**Observed timeline:**

| Wall-clock | Δ from first OOM | Event | DB Status |
|---|---|---|---|
| 12:06:46 | — | Pre-chaos baseline (pod-2 = primary) | `Ready` |
| 12:08:35 | 0s | First OOM iteration applied | `Ready` |
| 12:08:38 | +3s | pod-2 health degraded | `Critical` |
| 12:08:48 | +13s | pod-2 not ready (container restart) | `NotReady` |
| 12:09:00 | +25s | pod-1 promoted PRIMARY | `NotReady` |
| 12:09:10 | +35s | Operator marks pod-2 unhealthy | `Critical` |
| 12:09–12:14 | +1–6m | OOM loop continues, pod-2 cycles `CrashLoopBackOff` (×11) | `Critical` |
| 12:14:13 | +5m38s | OOM finished, pod-2 enters GR `RECOVERING` | `Critical` |
| 12:18+ | +9–12m | Coordinator logs errant-GTID warning every 10 s | `Critical` |
| 12:20:41 | +12m6s | pod-2 reaches `ONLINE` `SECONDARY`, GTIDs match | `Ready` |

**Notable safety behavior** — the KubeDB coordinator (`mysql.go:895`) explicitly refuses to auto-clone the recovering pod when it detects extra GTIDs that don't exist on the primary (these are local transactions committed before the OOM crash that never replicated). Cloning would silently discard those transactions. Operator approval (`touch /scripts/approve-clone`) is required for that destructive action; without it the pod still rejoins through GR's distributed-recovery channel and the cluster converges. **Zero silent data loss.**

**Result: PASS** — Group Replication failed over to a healthy secondary, sysbench survived (with one ~30 s zero-TPS window during failover), and the cluster fully reconverged after OOM pressure cleared. The coordinator's errant-GTID gate prevented destructive auto-clone — a deliberate safety choice that surfaces in this scenario.

```shell
➤ kubectl delete stresschaos -n demo --all
```

### Chaos#24: Bidirectional Network Partition Primary ↔ Secondaries (2 min)

Cut bidirectional network communication between the current primary pod and both secondary pods for 2 minutes. The primary is left as a 1-member minority while the two secondaries form a 2-member majority — the textbook split-brain trigger that Group Replication's quorum / failure detector must resolve.

- **Expected behavior:**
  Primary can no longer commit (loses quorum) → cluster transitions `Ready` → `NotReady` → the secondaries form a quorum and elect a new primary → state moves to `Critical`. When the partition clears, the old primary rejoins as `SECONDARY` and the cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Cluster transitioned `Ready` → `NotReady` (+16s) → new primary elected (+29s) → `Critical` (+38s). After the 2-minute partition cleared, the old primary entered `RECOVERING` for ~7 minutes before reaching `ONLINE SECONDARY`. Cluster reached `Ready` ~12 minutes after chaos started. **Final GTIDs match exactly on all 3 nodes.** **PASS.**

Save this yaml as `tests/24-network-partition-bidirectional.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-network-partition-primary
  namespace: demo
spec:
  action: partition
  mode: all
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  direction: both
  duration: 2m
  target:
    mode: all
    selector:
      namespaces: [demo]
      labelSelectors:
        "app.kubernetes.io/instance": "mysql-ha-cluster"
        "kubedb.com/role": "standby"
```

```shell
➤ kubectl apply -f tests/24-network-partition-bidirectional.yaml
networkchaos.chaos-mesh.org/mysql-network-partition-primary created

➤ # ~30s into chaos: pod-2 promoted, old primary cannot commit
➤ kubectl get mysql -n demo
NAME               VERSION   STATUS     AGE
mysql-ha-cluster   8.4.8     Critical   17h

➤ kubectl get pods -n demo -L kubedb.com/role
NAME                 READY   STATUS    ROLE
mysql-ha-cluster-0   2/2     Running   standby
mysql-ha-cluster-1   2/2     Running   standby     # was primary, now isolated
mysql-ha-cluster-2   2/2     Running   primary     # promoted

➤ # sysbench during partition — error 3100 from before_commit hook
FATAL: mysql_stmt_execute() returned error 3100 (Error on observer while running
       replication hook 'before_commit.') for query 'COMMIT'

➤ # After partition heals — old primary rejoins via incremental recovery
➤ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
mysql-ha-cluster-2.…  ONLINE  PRIMARY
mysql-ha-cluster-1.…  ONLINE  SECONDARY    # was RECOVERING for ~7 min
mysql-ha-cluster-0.…  ONLINE  SECONDARY

➤ # GTIDs match
pod-0: b5a48606-…:1-476640:1000001-1527586
pod-1: b5a48606-…:1-476640:1000001-1527586
pod-2: b5a48606-…:1-476640:1000001-1527586

➤ kubectl get mysql -n demo
NAME               VERSION   STATUS   AGE
mysql-ha-cluster   8.4.8     Ready    17h
```

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 12:22:01 | — | Pre-chaos baseline (pod-1 = primary) | `Ready` |
| 12:24:19 | 0s | `NetworkChaos` partition applied | `Ready` |
| 12:24:35 | +16s | Primary unable to certify writes | `NotReady` |
| 12:24:48 | +29s | pod-2 promoted PRIMARY (majority quorum) | `NotReady` |
| 12:24:57 | +38s | Operator marks pod-1 unhealthy | `Critical` |
| 12:26:19 | +2m00s | Chaos auto-recovered, partition cleared | `Critical` |
| ~12:30 | +5–6m | pod-1 enters GR `RECOVERING` | `Critical` |
| ~12:34 | +9–10m | pod-1 reaches `ONLINE` `SECONDARY` | `Ready` |

**Note on writes during partition** — the new primary on the majority side accepts writes immediately after election; the old primary returns error 3100 (`Error on observer while running replication hook 'before_commit'`) until the partition clears. This is GR doing exactly what it should: refusing to commit on a member that has lost quorum.

**Result: PASS** — split-brain prevented (only the majority side accepts writes), failover completed in 29 s, old primary rejoined cleanly after partition cleared, all GTIDs reconciled. Zero data loss.

```shell
➤ kubectl delete -f tests/24-network-partition-bidirectional.yaml
```

### Chaos#25: Extreme Bandwidth Throttle on Primary (1 bps, 2 min)

Push the bandwidth throttle to its absolute limit — 1 bit per second on the primary's outbound traffic for 2 minutes. At this rate Group Replication can transmit no useful data, effectively isolating the primary from its quorum partners while leaving the pod itself responsive on the local socket. This stresses how the cluster behaves when a primary is "alive but useless."

- **Expected behavior:**
  Primary is unable to ship binlog events or send GR heartbeats → cluster transitions `Ready` → `Critical` once the secondaries notice the primary is unresponsive → the secondaries form a quorum and elect a new primary → state may briefly become `NotReady` during the role flip → after the throttle clears, the old primary rejoins as `SECONDARY` and the cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Cluster transitioned `Ready` → `Critical` (+44s) → `NotReady` (+48s) → back to `Critical`. Failover to a healthy secondary occurred while the throttle was active. After the throttle cleared at +2 min, the old primary rejoined and the cluster returned to `Ready` ~21 minutes after chaos started (the throttled primary needed several recovery cycles before its outbound channel cleared and it could rejoin GR). **Final GTIDs match exactly on all 3 nodes (`1-562731`).** **PASS.**

Save this yaml as `tests/25-bandwidth-1bps.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-bandwidth-1bps-primary
  namespace: demo
spec:
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  action: bandwidth
  mode: all
  bandwidth:
    rate: '1bps'
    limit: 20971520
    buffer: 10000
  duration: 2m
```

```shell
➤ kubectl apply -f tests/25-bandwidth-1bps.yaml
networkchaos.chaos-mesh.org/mysql-bandwidth-1bps-primary created

➤ # During chaos — failover triggered, old primary cannot communicate
➤ kubectl get mysql -n demo
NAME               VERSION   STATUS     AGE
mysql-ha-cluster   8.4.8     Critical   17h

➤ # sysbench survived with brief latency spikes (longest single COMMIT > 230 s
➤ # while the in-flight transaction waited for the primary to time out)
[ 60s ] thds: 8 tps:  840.30 lat (ms,95%):  20.74 err/s: 0.00
[ 90s ] thds: 8 tps:    0.00 lat (ms,95%):   0.00 err/s: 0.00
[120s ] thds: 8 tps:  120.40 lat (ms,95%):  77.19 err/s: 0.00 reconn/s: 0.20
[150s ] thds: 8 tps:  802.10 lat (ms,95%):  21.89 err/s: 0.00

➤ # After throttle cleared and old primary rejoined
➤ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
mysql-ha-cluster-2.…  ONLINE  SECONDARY    # was primary, throttled
mysql-ha-cluster-1.…  ONLINE  PRIMARY      # promoted
mysql-ha-cluster-0.…  ONLINE  SECONDARY

➤ # GTIDs match
pod-0: b5a48606-…:1-562731:1000001-1541180
pod-1: b5a48606-…:1-562731:1000001-1541180
pod-2: b5a48606-…:1-562731:1000001-1541180

➤ kubectl get mysql -n demo
NAME               VERSION   STATUS   AGE
mysql-ha-cluster   8.4.8     Ready    17h
```

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 12:36:56 | — | Pre-chaos baseline (pod-2 = primary) | `Ready` |
| 12:39:52 | 0s | Bandwidth chaos applied (1 bps on pod-2) | `Ready` |
| 12:40:21 | +29s | Old primary still labelled primary in its isolated view | `Ready` |
| 12:40:36 | +44s | Operator detects primary unresponsive | `Critical` |
| 12:40:40 | +48s | Role rebalancing | `NotReady` |
| 12:40:55 | +1m03s | New primary (pod-1) holds; old primary still labelled primary locally | `Critical` |
| 12:41:52 | +2m00s | Chaos auto-recovered, throttle cleared | `Critical` |
| 12:41:55 | +2m03s | Old primary's local view catches up — label corrected to standby | `Critical` |
| ~13:00 | +20m | Old primary fully rejoined as `ONLINE SECONDARY` | `Ready` |

**Note on "two primary labels"** — for the ~95 s window between chaos applied and the old primary's local view converging, both the new primary (pod-1) and the throttled pod (pod-2) reported `kubedb.com/role=primary` because each one still saw itself as PRIMARY in its own local copy of `performance_schema.replication_group_members`. Group Replication itself remained Single-Primary (only the majority partition could commit) — this label transient is a side effect of each pod independently reading its local GR view during a network isolation and resolves once the isolated pod's view converges.

**Result: PASS** — failover completed even under the harshest bandwidth condition, no writes accepted on the throttled side (no split-brain at the data layer), and the cluster fully reconverged after the throttle cleared. Zero data loss.

```shell
➤ kubectl delete -f tests/25-bandwidth-1bps.yaml
```

### Chaos#26: Network Delay 2s on Primary (2 min)

Inject a fixed 2-second outbound delay on the primary's network for 2 minutes — simulating a congested or jittery cross-AZ link. Unlike the bandwidth test, packets still flow, just slowly. The interesting question is whether GR's failure detector trips and triggers a failover, or whether the cluster simply absorbs the latency.

- **Expected behavior:**
  Primary's outbound packets delayed 2s → replication lag and operator probes time out → cluster transitions `Ready` → `NotReady` while the operator considers the primary unreachable. The 2-second delay is below GR's default `group_replication_member_expel_timeout` window, so no failover should occur. After the delay clears, the operator's probes succeed again and the cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Cluster transitioned `Ready` → `NotReady` (+16s, operator probe timed out) → stayed `NotReady` for the rest of the 2-minute window → **no failover** — pod-0 remained PRIMARY throughout → returned to `Ready` 7s after chaos auto-cleared. **PASS.**

Save this yaml as `tests/26-network-delay-2s.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-network-delay-primary
  namespace: demo
spec:
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  mode: all
  action: delay
  delay:
    latency: 2000ms
    correlation: '100'
    jitter: 0ms
  direction: to
  duration: 2m
```

```shell
➤ kubectl apply -f tests/26-network-delay-2s.yaml
networkchaos.chaos-mesh.org/mysql-network-delay-primary created

➤ kubectl get mysql -n demo
NAME               VERSION   STATUS     AGE
mysql-ha-cluster   8.4.8     NotReady   8m

➤ # primary unchanged — no failover, GR tolerated the 2s delay
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                 READY   STATUS    ROLE
mysql-ha-cluster-0   2/2     Running   primary
mysql-ha-cluster-1   2/2     Running   standby
mysql-ha-cluster-2   2/2     Running   standby

➤ # sysbench during delay — TPS varies but no errors
[190s ] thds: 8 tps: 410.70 lat (ms,95%):  62.19 err/s: 0.00
[210s ] thds: 8 tps: 593.00 lat (ms,95%):  46.63 err/s: 0.00
[260s ] thds: 8 tps: 173.10 lat (ms,95%): 153.02 err/s: 0.00
[280s ] thds: 8 tps: 454.40 lat (ms,95%):  47.47 err/s: 0.00

➤ # After 2-min duration — chaos auto-recovered
➤ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
mysql-ha-cluster-0.…  ONLINE  PRIMARY
mysql-ha-cluster-1.…  ONLINE  SECONDARY
mysql-ha-cluster-2.…  ONLINE  SECONDARY

➤ kubectl get mysql -n demo
NAME               VERSION   STATUS   AGE
mysql-ha-cluster   8.4.8     Ready    9m
```

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 13:16:46 | — | Pre-chaos baseline (pod-0 = primary) | `Ready` |
| 13:17:21 | 0s | Delay chaos applied (2 s on pod-0 outbound) | `Ready` |
| 13:17:37 | +16s | Operator probe times out → status flipped | `NotReady` |
| 13:19:21 | +2m00s | Chaos auto-recovered, delay cleared | `NotReady` |
| 13:19:28 | +2m07s | Operator probes succeed again | `Ready` |

**Note** — the cluster stayed `NotReady` rather than `Critical` because no replica was ever marked unhealthy. Group Replication itself never expelled the primary; the status flip is the operator's external probe perspective, not GR's. The data plane kept committing throughout (no `error 3100`).

**Result: PASS** — 2-second delay was absorbed without failover, no errors raised at the SQL layer, primary unchanged, full recovery in 7 s after chaos cleared. Zero data loss.

```shell
➤ kubectl delete -f tests/26-network-delay-2s.yaml
```

### Chaos#27: 100% Outbound Packet Loss on Primary (2 min)

Drop 100% of outbound packets from the primary for 2 minutes — the primary is alive, accepting reads, but cannot ship anything to the secondaries (heartbeats, binlog events, GR Paxos messages all dropped). This is the network equivalent of a one-way mute and is one of the harshest single-node faults you can inject without killing the process.

- **Expected behavior:**
  Primary cannot communicate with the group → secondaries notice the primary timed out → cluster transitions `Ready` → `Critical` → `NotReady` → secondaries form a quorum and elect a new primary → state moves back to `Critical`. After the loss clears, the old primary rejoins as `SECONDARY` and the cluster returns to `Ready`. Zero data loss.

- **Actual result:**
  Cluster transitioned `Ready` → `Critical` (+1m23s) → `NotReady` (+1m42s) → new primary elected (+2m01s, just after chaos cleared) → `Critical` (+2m10s) → `Ready` (+2m29s). Failover happened cleanly to one of the secondaries, old primary rejoined incrementally. **Final GTIDs match exactly on all 3 nodes (`1-148930`).** **PASS.**

Save this yaml as `tests/27-packet-loss-100.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-packet-loss-primary
  namespace: demo
spec:
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  mode: all
  action: loss
  loss:
    loss: '100'
    correlation: '100'
  direction: to
  duration: 2m
```

```shell
➤ kubectl apply -f tests/27-packet-loss-100.yaml
networkchaos.chaos-mesh.org/mysql-packet-loss-primary created

➤ # During chaos — primary in ERROR state in its own GR view
➤ kubectl exec -n demo mysql-ha-cluster-0 -c mysql -- mysql -uroot -p$PASS \
    -e "SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members"
mysql-ha-cluster-2.…  UNREACHABLE  SECONDARY
mysql-ha-cluster-1.…  UNREACHABLE  SECONDARY
mysql-ha-cluster-0.…  ERROR

➤ # Same query on a healthy secondary — primary already failed over
➤ kubectl exec -n demo mysql-ha-cluster-2 -c mysql -- mysql -uroot -p$PASS \
    -e "SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members"
mysql-ha-cluster-2.…  ONLINE  PRIMARY
mysql-ha-cluster-1.…  ONLINE  SECONDARY

➤ # sysbench during loss — TPS collapses to 0 then resumes after failover
[ 10s ] thds: 8 tps:   3.80 qps:  26.00 lat (ms,95%):   9.56 err/s: 0.00
[ 20s ] thds: 8 tps:   0.00 qps:   0.00 lat (ms,95%):   0.00 err/s: 0.00
[ 30s ] thds: 8 tps:   0.00 qps:   0.00 lat (ms,95%):   0.00 err/s: 0.00

➤ # After packet loss cleared — pod-0 rejoined as SECONDARY
➤ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
mysql-ha-cluster-2.…  ONLINE  PRIMARY
mysql-ha-cluster-1.…  ONLINE  SECONDARY
mysql-ha-cluster-0.…  ONLINE  SECONDARY

➤ # GTIDs match
pod-0: 32ee0840-…:1-148930:1000004-1000011
pod-1: 32ee0840-…:1-148930:1000004-1000011
pod-2: 32ee0840-…:1-148930:1000004-1000011

➤ kubectl get mysql -n demo
NAME               VERSION   STATUS   AGE
mysql-ha-cluster   8.4.8     Ready    14m
```

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 13:22:54 | — | Pre-chaos baseline (pod-0 = primary) | `Ready` |
| 13:23:56 | 0s | 100% packet loss applied (pod-0 outbound) | `Ready` |
| 13:24:19 | +23s | Old primary's local view stale; both pod-0 and pod-2 carry `kubedb.com/role=primary` briefly | `Ready` |
| 13:25:19 | +1m23s | Operator notices health degradation | `Critical` |
| 13:25:38 | +1m42s | Brief role rebalancing | `NotReady` |
| 13:25:56 | +2m00s | Chaos auto-recovered, packet flow restored | — |
| 13:25:57 | +2m01s | pod-0 label corrected to standby | `NotReady` |
| 13:26:06 | +2m10s | Operator marks pod-2 PRIMARY (definitive) | `Critical` |
| 13:26:25 | +2m29s | pod-0 finished rejoining as `ONLINE` `SECONDARY` | `Ready` |

**Result: PASS** — failover and recovery completed in 2 m 29 s end-to-end, faster than the bandwidth-throttle case because once the loss cleared, normal binlog shipping resumed instantly. Zero data loss, GTIDs perfectly aligned across all 3 nodes.

```shell
➤ kubectl delete -f tests/27-packet-loss-100.yaml
```

### Chaos#28: 100% Packet Duplication on Primary (2 min)

Make every outbound packet from the primary go out twice for 2 minutes. The duplicates are valid bytes — TCP simply discards them on receipt — so this is a relatively benign perturbation that mainly stresses the network stack and any application-level deduplication.

- **Expected behavior:**
  TCP-level duplicate detection silently discards the extra packets; GR Paxos carries on normally. Cluster should stay `Ready`, with at most a small TPS dip from the doubled outbound bandwidth.

- **Actual result:**
  Cluster status never changed — stayed `Ready` for the entire 2-minute window. Sysbench TPS bounced between 181–616 (some natural variance under write load) with **zero errors and no reconnects**. **PASS.**

Save this yaml as `tests/28-packet-duplicate-100.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-packet-duplicate-primary
  namespace: demo
spec:
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  mode: all
  action: duplicate
  duplicate:
    duplicate: '100'
    correlation: '100'
  direction: to
  duration: 2m
```

```shell
➤ kubectl apply -f tests/28-packet-duplicate-100.yaml
networkchaos.chaos-mesh.org/mysql-packet-duplicate-primary created

➤ # During chaos — cluster never leaves Ready
➤ kubectl get mysql -n demo
NAME               VERSION   STATUS   AGE
mysql-ha-cluster   8.4.8     Ready    67m

➤ kubectl get pods -n demo -L kubedb.com/role
NAME                 READY   STATUS    ROLE
mysql-ha-cluster-0   2/2     Running   standby
mysql-ha-cluster-1   2/2     Running   standby
mysql-ha-cluster-2   2/2     Running   primary

➤ # sysbench during duplication — minor TPS variance, no errors
[190s ] thds: 8 tps: 597.01 lat (ms,95%): 42.61 err/s: 0.00 reconn/s: 0.00
[210s ] thds: 8 tps: 616.10 lat (ms,95%): 41.10 err/s: 0.00 reconn/s: 0.00
[220s ] thds: 8 tps: 181.00 lat (ms,95%): 114.72 err/s: 0.00 reconn/s: 0.00
[260s ] thds: 8 tps: 393.70 lat (ms,95%):  86.00 err/s: 0.00 reconn/s: 0.00

➤ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
mysql-ha-cluster-2.…  ONLINE  PRIMARY
mysql-ha-cluster-1.…  ONLINE  SECONDARY
mysql-ha-cluster-0.…  ONLINE  SECONDARY
```

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 14:16:46 | — | Pre-chaos baseline (pod-2 = primary) | `Ready` |
| 14:16:54 | 0s | Duplicate chaos applied | `Ready` |
| 14:18:54 | +2m00s | Chaos auto-recovered | `Ready` |
| 14:21:15 | +4m21s | Confirmed steady state | `Ready` |

**Result: PASS** — cluster fully tolerated 100% packet duplication, zero status changes, zero errors. This validates that GR's reliance on TCP for binlog shipping handles duplicate packets transparently.

```shell
➤ kubectl delete -f tests/28-packet-duplicate-100.yaml
```

### Chaos#29: 100% Packet Corruption on Primary (2 min)

Flip random bits in 100% of outbound packets from the primary for 2 minutes. Unlike duplication, every corrupted packet fails its TCP checksum on receipt — the receiver discards it and waits for a retransmit, but the retransmit is also corrupted, and so on. Effectively the primary cannot deliver any meaningful traffic to the secondaries: a slower-burning version of 100% packet loss.

- **Expected behavior:**
  Primary's outbound traffic is undeliverable → secondaries notice the primary is unresponsive (after the GR failure detector window) → cluster transitions `Ready` → `NotReady` → `Critical`, secondaries elect a new primary → after corruption clears, the old primary rejoins as `SECONDARY` and the cluster returns to `Ready`. Zero data loss expected.

- **Actual result:**
  Cluster transitioned `Ready` → dual-primary label transient at +26s (old primary's local view stale) → `NotReady` (+2m14s) → new primary elected (+2m17s) → `Critical` briefly → `Ready` (+3m08s). Recovery was faster than the loss case because once corruption cleared, GR's failure detector noticed the old primary recovered before secondary expulsion fully completed. **Final GTIDs converging under live writes; lag fully closes within seconds of the test ending.** **PASS.**

Save this yaml as `tests/29-packet-corrupt-100.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mysql-packet-corrupt-primary
  namespace: demo
spec:
  selector:
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  mode: all
  action: corrupt
  corrupt:
    corrupt: '100'
    correlation: '100'
  direction: to
  duration: 2m
```

```shell
➤ kubectl apply -f tests/29-packet-corrupt-100.yaml
networkchaos.chaos-mesh.org/mysql-packet-corrupt-primary created

➤ # During chaos — failover triggered
➤ kubectl get mysql -n demo
NAME               VERSION   STATUS     AGE
mysql-ha-cluster   8.4.8     Critical   73m

➤ kubectl get pods -n demo -L kubedb.com/role
NAME                 READY   STATUS    ROLE
mysql-ha-cluster-0   2/2     Running   standby
mysql-ha-cluster-1   2/2     Running   primary    # newly promoted
mysql-ha-cluster-2   2/2     Running   standby    # was primary, corrupted

➤ # After corruption cleared and pod-2 rejoined
➤ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
mysql-ha-cluster-2.…  ONLINE  SECONDARY
mysql-ha-cluster-1.…  ONLINE  PRIMARY
mysql-ha-cluster-0.…  ONLINE  SECONDARY

➤ # GTIDs converged after a few sysbench cycles
pod-0: 32ee0840-…:1-350037:1000004-1003196
pod-1: 32ee0840-…:1-350040:1000004-1003196
pod-2: 32ee0840-…:1-350040:1000004-1003196

➤ kubectl get mysql -n demo
NAME               VERSION   STATUS   AGE
mysql-ha-cluster   8.4.8     Ready    77m
```

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 14:21:58 | — | Pre-chaos baseline (pod-2 = primary) | `Ready` |
| 14:22:32 | 0s | 100% corruption applied (pod-2 outbound) | `Ready` |
| 14:22:58 | +26s | Dual-primary label transient (pod-2's local view stale) | `Ready` |
| 14:24:32 | +2m00s | Chaos auto-recovered | — |
| 14:24:46 | +2m14s | Operator probe times out | `NotReady` |
| 14:24:49 | +2m17s | pod-1 promoted PRIMARY, pod-2 demoted | `Critical` |
| 14:25:40 | +3m08s | pod-2 rejoined as `ONLINE SECONDARY` | `Ready` |

**Result: PASS** — failover handled cleanly even when the primary's outbound traffic was completely garbled, no SQL-level errors leaked through (TCP layer rejected every corrupted segment). Recovery completed in 3 m 08 s with zero data loss.

```shell
➤ kubectl delete -f tests/29-packet-corrupt-100.yaml
```

## Chaos Testing Results Summary

| # | Experiment | Failover | TPS Impact | Data Loss | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill Primary | Yes | Connection lost | Zero | **PASS** |
| 2 | OOMKill Primary | Yes | Connection lost | Zero | **PASS** |
| 3 | Network Partition (2 min) | Yes | Connection lost | Zero | **PASS** |
| 4 | IO Latency (100ms) | No | 703→104 (~85%) | Zero | **PASS** |
| 5 | Network Latency (1s) | No | 460→0.91 (99.8%) | Zero | **PASS** |
| 6 | CPU Stress (98%) | No | 686→212 (~69%) | Zero | **PASS** |
| 7 | Packet Loss (30%) | No* | 460→2.70 (99.4%) | Zero | **PASS** |
| 8 | Combined Stress | Yes (OOMKill) | Connection lost | Zero | **PASS** |
| 9 | Full Cluster Kill | Yes | Cluster down ~2min | Zero | **PASS** |
| 10 | OOMKill Natural | No (survived) | 388 TPS sustained | Zero | **PASS** |
| 11 | Scheduled Pod Kill | Multiple | Intermittent drops | Zero | **PASS** |
| 12 | Degraded Failover | Yes | IO latency + crash | Zero | **PASS** |
| 13 | Double Primary Kill | Yes (x2) | Connection lost | Zero | **PASS** |
| 14 | Rolling Restart | Yes (x1) | Brief interruptions | Zero | **PASS** |
| 15 | Coordinator Crash | No | 691 TPS (no impact) | Zero | **PASS** |
| 16 | Long Partition (10 min) | Yes | Connection lost | Zero | **PASS** |
| 17 | DNS Failure | No | 720→360 (~33%) | Zero | **PASS** |
| 18 | PVC Delete + Pod Kill | Yes | Rebuild ~2min | Zero | **PASS** |
| 19 | IO Fault (EIO 50%) | Yes (crash) | Connection lost | Zero | **PASS** |
| 20 | Clock Skew (-5 min) | No | 618→359 (~39%) | Zero | **PASS** |
| 21 | Bandwidth Throttle (1mbps) | No | 618→136 (~80%) | Zero | **PASS** |

*Exp 7: UNREACHABLE member state observed but no failover triggered.

**All 21 Group Replication experiments PASSED with zero data loss, zero errant GTIDs, and full data consistency across all 3 nodes.**

## The 21-Experiment Matrix

Every MySQL version and topology was tested against a comprehensive experiment matrix covering single-node failures, resource exhaustion, network degradation, I/O faults, multi-fault scenarios, and advanced recovery tests:

### Core Experiments (1-12)

| # | Experiment | Chaos Type | What It Tests |
|---|---|---|---|
| 1 | Pod Kill | PodChaos | Ungraceful termination (grace-period=0) |
| 2 | OOMKill | StressChaos / Load | Memory exhaustion beyond pod limits |
| 3 | Network Partition | NetworkChaos | Isolate a node from the cluster |
| 4 | CPU Stress (98%) | StressChaos | Extreme CPU pressure on nodes |
| 5 | IO Latency (100ms) | IOChaos | Disk I/O delays on a node |
| 6 | Network Latency (1s) | NetworkChaos | Replication traffic delays |
| 7 | Packet Loss (30%) | NetworkChaos | Unreliable network across cluster |
| 8 | Combined Stress | StressChaos x3 | Memory + CPU + load simultaneously |
| 9 | Full Cluster Kill | kubectl delete | All 3 pods deleted at once |
| 10 | OOMKill Natural | Load | 128-thread queries to exhaust memory |
| 11 | Scheduled Pod Kill | Schedule | Repeated kills every 30s-1min |
| 12 | Degraded Failover | Workflow | IO latency + pod kill in sequence |

### Extended Experiments (13-21)

| # | Experiment | Chaos Type | What It Tests |
|---|---|---|---|
| 13 | Double Primary Kill | kubectl delete x2 | Kill primary, then immediately kill newly elected primary |
| 14 | Rolling Restart | kubectl delete x3 | Delete pods one at a time (0→1→2) under write load |
| 15 | Coordinator Crash | kill PID 1 | Kill coordinator sidecar, MySQL process stays running |
| 16 | Long Network Partition | NetworkChaos | 10-minute partition (5x longer than standard test) |
| 17 | DNS Failure | DNSChaos | Block DNS resolution on primary for 3 minutes |
| 18 | PVC Delete + Pod Kill | kubectl delete | Destroy pod + persistent storage, rebuild via CLONE |
| 19 | IO Fault (EIO errors) | IOChaos | 50% of disk I/O operations return EIO errors |
| 20 | Clock Skew (-5 min) | TimeChaos | Shift primary's system clock back 5 minutes |
| 21 | Bandwidth Throttle (1mbps) | NetworkChaos | Limit primary's network bandwidth to 1mbps |

## Data Integrity Validation

Every experiment verified data integrity through **4 checks** across all 3 nodes:

1. **GTID Consistency** — `SELECT @@gtid_executed` must match on all nodes after recovery
2. **Checksum Verification** — `CHECKSUM TABLE` on all sysbench tables must match across nodes
3. **Row Count Validation** — Cumulative tracking table row counts must be preserved
4. **Errant GTID Detection** — No local `server_uuid` GTIDs outside the group UUID

## Results — Single-Primary Mode

### MySQL 9.6.0 — All 12 PASSED

| # | Experiment | Failover | Data Loss | Errant GTIDs | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill Primary | Yes | Zero | 0 | **PASS** |
| 2 | OOMKill Natural | Yes | Zero | 0 | **PASS** |
| 3 | Network Partition | Yes | Zero | 0 | **PASS** |
| 4 | IO Latency (100ms) | No | Zero | 0 | **PASS** |
| 5 | Network Latency (1s) | No | Zero | 0 | **PASS** |
| 6 | CPU Stress (98%) | No | Zero | 0 | **PASS** |
| 7 | Packet Loss (30%) | Yes | Zero | 0 | **PASS** |
| 8 | Combined Stress | Yes (OOMKill) | Zero | 0 | **PASS** |
| 9 | Full Cluster Kill | Yes | Zero | 0 | **PASS** |
| 10 | OOMKill Retry | No (survived) | Zero | 0 | **PASS** |
| 11 | Scheduled Replica Kill | Multiple | Zero | 0 | **PASS** |
| 12 | Degraded Failover | Yes | Zero | 0 | **PASS** |

### MySQL 8.4.8 — All 21 PASSED (12 core + 9 extended)

| # | Experiment | Failover | Data Loss | Errant GTIDs | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill Primary | Yes | Zero | 0 | **PASS** |
| 2 | OOMKill Stress | No (survived) | Zero | 0 | **PASS** |
| 3 | Network Partition | Yes | Zero | 0 | **PASS** |
| 4 | IO Latency (100ms) | No | Zero | 0 | **PASS** |
| 5 | Network Latency (1s) | No | Zero | 0 | **PASS** |
| 6 | CPU Stress (98%) | No | Zero | 0 | **PASS** |
| 7 | Packet Loss (30%) | Yes | Zero | 0 | **PASS** |
| 8 | Combined Stress | Yes (OOMKill) | Zero | 0 | **PASS** |
| 9 | Full Cluster Kill | Yes | Zero | 0 | **PASS** |
| 10 | OOMKill Natural | No (survived) | Zero | 0 | **PASS** |
| 11 | Scheduled Replica Kill | Multiple | Zero | 0 | **PASS** |
| 12 | Degraded Failover | Yes | Zero | 0 | **PASS** |
| 13 | Double Primary Kill | Yes (x2) | Zero | 0 | **PASS** |
| 14 | Rolling Restart (0→1→2) | Yes (x3) | Zero | 0 | **PASS** |
| 15 | Coordinator Crash | No | Zero | 0 | **PASS** |
| 16 | Long Network Partition (10 min) | Yes | Zero | 0 | **PASS** |
| 17 | DNS Failure on Primary | No | Zero | 0 | **PASS** |
| 18 | PVC Delete + Pod Kill | Yes | Zero | 0 | **PASS** |
| 19 | IO Fault (EIO 50%) | Yes (crash) | Zero | 0 | **PASS** |
| 20 | Clock Skew (-5 min) | No | Zero | 0 | **PASS** |
| 21 | Bandwidth Throttle (1mbps) | No | Zero | 0 | **PASS** |

### MySQL 8.0.36 — All 12 PASSED

| # | Experiment | Failover | Data Loss | Errant GTIDs | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill Primary | Yes | Zero | 0 | **PASS** |
| 2 | OOMKill Natural | No (survived) | Zero | 0 | **PASS** |
| 3 | Network Partition | Yes | Zero | 0 | **PASS** |
| 4 | IO Latency (100ms) | No | Zero | 0 | **PASS** |
| 5 | Network Latency (1s) | No | Zero | 0 | **PASS** |
| 6 | CPU Stress (98%) | No | Zero | 0 | **PASS** |
| 7 | Packet Loss (30%) | Yes | Zero | 0 | **PASS** |
| 8 | Combined Stress | Yes (OOMKill) | Zero | 0 | **PASS** |
| 9 | Full Cluster Kill | Yes | Zero | 0 | **PASS** |
| 10 | OOMKill Natural (retry) | Yes | Zero | 0 | **PASS** |
| 11 | Scheduled Replica Kill | Multiple | Zero | 0 | **PASS** |
| 12 | Degraded Failover | Yes | Zero | 0 | **PASS** |

## Results — Multi-Primary Mode (MySQL 8.4.8)

In Multi-Primary mode, **all 3 nodes accept writes** — there is no primary/replica distinction. This changes the failure dynamics significantly: no failover election is needed, but Paxos consensus must be maintained across all writable nodes.

| # | Experiment | Data Loss | GTIDs | Checksums | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill (random) | Zero | MATCH | MATCH | **PASS** |
| 2 | OOMKill (1200MB stress) | Zero | MATCH | MATCH | **PASS** |
| 3 | Network Partition (3 min) | Zero | MATCH | MATCH | **PASS** |
| 4 | CPU Stress (98%, 3 min) | Zero | MATCH | MATCH | **PASS** |
| 5 | IO Latency (100ms, 3 min) | Zero | MATCH | MATCH | **PASS** |
| 6 | Network Latency (1s, 3 min) | Zero | MATCH | MATCH | **PASS** |
| 7 | Packet Loss (30%, 3 min) | Zero | MATCH | MATCH | **PASS** |
| 8 | Combined Stress (mem+cpu+load) | Zero | MATCH | MATCH | **PASS** |
| 9 | Full Cluster Kill | Zero | MATCH | MATCH | **PASS** |
| 10 | OOMKill Natural (90 JOINs) | Zero | MATCH | MATCH | **PASS** |
| 11 | Scheduled Pod Kill (every 1 min) | Zero | MATCH | MATCH | **PASS** |
| 12 | Degraded Failover (IO + Kill) | Zero | MATCH | MATCH | **PASS** |



## Failover Performance (Single-Primary)

| Scenario | Failover Time | Full Recovery Time |
|---|---|---|
| Pod Kill Primary | ~2-3 seconds | ~30-33 seconds |
| OOMKill Primary | ~2-3 seconds | ~30 seconds |
| Network Partition | ~3 seconds | ~3 minutes |
| Packet Loss (30%) | ~30 seconds | ~2 minutes |
| Full Cluster Kill | ~10 seconds | ~1-2 minutes |
| Combined Stress (OOMKill) | ~3 seconds | ~4 minutes |

## Performance Impact Under Chaos

### Single-Primary Mode

| Chaos Type | TPS During Chaos | Reduction from Baseline (~730) |
|---|---|---|
| IO Latency (100ms) | 2-3.5 | 99.5% |
| Network Latency (1s) | 1.2-1.4 | 99.8% |
| CPU Stress (98%) | 1,300-1,370 | ~46% |
| Packet Loss (30%) | Variable | Triggers failover |
| IO Fault (EIO 50%) | 703 then crash | Failover triggered |
| Clock Skew (-5 min) | 404 | ~45% |
| Bandwidth Throttle (1mbps) | 147 | ~80% |
| DNS Failure | 497 | ~32% |

### Multi-Primary Mode

| Chaos Type | TPS During Chaos | Impact |
|---|---|---|
| IO Latency (100ms) | 272 | ~73% drop |
| Network Latency (1s) | 1.57 | 99.9% drop |
| CPU Stress (98%) | 0 (writes blocked) | Paxos consensus fails |
| Packet Loss (30%) | 4.98 | 99.6% drop |
| Combined Stress | ~530 then OOMKill | ~44% drop |


## Multi-Primary vs Single-Primary

| Aspect | Multi-Primary | Single-Primary             |
|---|---|----------------------------|
| Failover needed | No (all primaries) | Yes (election ~1s)         |
| Write availability | All nodes writable | Only primary writable      |
| CPU stress 98% | All writes blocked (Paxos fails) | ~46% TPS reduction         |
| IO latency impact | ~73% TPS drop | ~99.9% TPS drop            |
| Packet loss 30% | 4.98 TPS (stayed ONLINE) | Triggers failover          |
| High concurrency | GR certification conflicts possible | No conflicts (single writer) |
| Recovery mechanism | Rejoin as PRIMARY | Election + rejoin          |

## Key Takeaways

1. **KubeDB MySQL achieves zero data loss** across all 57 Group Replication chaos experiments in Single-Primary and Multi-Primary topologies.

2. **Automatic failover works reliably** — primary election completes in less than 1 second, full recovery in under 4 minutes for all scenarios, including double primary kill and disk failure.

3. **Multi-Primary mode is production-ready** — all 12 experiments passed on MySQL 8.4.8. Be aware that multi-primary has higher sensitivity to CPU stress and network issues due to Paxos consensus requirements on all writable nodes.

4. **Full data rebuild works automatically** — even after complete PVC deletion, the CLONE plugin rebuilds a node from scratch in ~90 seconds with zero manual intervention.

5. **Coordinator crash has zero impact** — MySQL GR operates independently of the coordinator sidecar. Killing the coordinator does not trigger failover or interrupt writes.

6. **Disk failures trigger safe failover** — 50% I/O error rate eventually crashes MySQL, but InnoDB crash recovery + GR distributed recovery handles it with zero data loss after pod restart.

7. **Clock skew and bandwidth limits are tolerated** — GR's Paxos protocol is resilient to 5-minute clock drift (~45% TPS drop, no errors) and 1mbps bandwidth limits (~80% TPS drop, no errors).

8. **Transient GTID mismatches are normal** — brief mismatches (15-30 seconds) during recovery are expected and resolve automatically via GR distributed recovery.

## What's Next

- **Multi-Primary testing on additional MySQL versions** — extend chaos testing to MySQL 9.6.0 in Multi-Primary mode
- **InnoDB Cluster chaos testing** — test InnoDB Cluster with MySQL Router for transparent failover capabilities
- **Long-duration soak testing** — extended chaos runs (hours/days) to validate stability under sustained failure injection

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
