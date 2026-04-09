---
title: "Chaos Engineering Results: KubeDB MySQL Group Replication Achieves Zero Data Loss Across 57 Experiments"
date: "2026-04-08"
weight: 14
authors:
- SK Ali Arman
tags:
- chaos-engineering
- chaos-mesh
- database
- kubedb
- kubernetes
- mysql
- group-replication
- high-availability
---

## Overview

We conducted **57 chaos experiments** across **3 MySQL versions** (8.0.36, 8.4.8, 9.6.0) and **2 Group Replication topologies** (Single-Primary and Multi-Primary) on KubeDB-managed 3-node clusters. The goal: validate that KubeDB MySQL delivers **zero data loss**, **automatic failover**, and **self-healing recovery** under realistic failure conditions with production-level write loads.

**The result: every experiment passed with zero data loss, zero split-brain, and zero errant GTIDs.**

This post summarizes the methodology, results, and key findings from the most comprehensive chaos testing effort we have run on KubeDB MySQL to date.

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
| KubeDB Version | 2026.2.26 |
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
  --version v2026.2.26 \
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
# Get the MySQL root password
PASS=$(kubectl get secret mysql-ha-cluster-auth -n demo -o jsonpath='{.data.password}' | base64 -d)

# Create the sbtest database
kubectl exec -n demo mysql-ha-cluster-0 -c mysql -- \
  mysql -uroot -p"$PASS" -e "CREATE DATABASE IF NOT EXISTS sbtest;"

# Get the sysbench pod name
SBPOD=$(kubectl get pods -n demo -l app=sysbench -o jsonpath='{.items[0].metadata.name}')

# Prepare tables (12 tables x 100k rows)
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
NAME                                VERSION   STATUS   AGE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    18h

NAME                             READY   STATUS    RESTARTS   AGE     ROLE
mysql-ha-cluster-0               2/2     Running   0          23m     primary
mysql-ha-cluster-1               2/2     Running   0          102m    standby
mysql-ha-cluster-2               2/2     Running   0          51m     standby
sysbench-load-849bdc4cdc-h2zpx   1/1     Running   0          7d22h
```

Verify GR members:

```shell
➤ kubectl exec -n demo mysql-ha-cluster-0 -c mysql -- \
    mysql -uroot -p"$PASS" -e "SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE \
    FROM performance_schema.replication_group_members;"
MEMBER_HOST                                         MEMBER_PORT   MEMBER_STATE   MEMBER_ROLE
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo       3306          ONLINE         PRIMARY
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY
```

The pod with `kubedb.com/role=primary` is the primary and `kubedb.com/role=standby` are the secondaries.

### Chaos#1: Kill the Primary Pod

We are about to kill the primary pod and see how fast the failover happens.

Save this yaml as `tests/01-pod-kill.yaml`:

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

Before running, let's see who is the primary:

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          55m   primary
mysql-ha-cluster-1               2/2     Running   0          79m   standby
mysql-ha-cluster-2               2/2     Running   0          28m   standby
```

Now run `watch kubectl get mysql,pods -n demo -L kubedb.com/role` in one terminal, and apply the chaos in another:

```shell
➤ kubectl apply -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org/mysql-primary-pod-kill created
```

Within seconds, the primary pod is killed and a new primary is elected. The database goes `NotReady` briefly:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     NotReady   18h

NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          10s   standby
mysql-ha-cluster-1               2/2     Running   0          79m   standby
mysql-ha-cluster-2               2/2     Running   0          28m   primary
```

`NotReady` means the old primary was killed and the new primary (pod-2) is now being promoted. After ~30 seconds, the old primary comes back as a standby and the cluster returns to `Ready`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    18h

NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          52s   standby
mysql-ha-cluster-1               2/2     Running   0          80m   standby
mysql-ha-cluster-2               2/2     Running   0          29m   primary
```

Verify data integrity — all 3 nodes must have matching GTIDs and checksums:

```shell
➤ # GR Members — all ONLINE
MEMBER_HOST                                         MEMBER_PORT   MEMBER_STATE   MEMBER_ROLE
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo       3306          ONLINE         PRIMARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY

➤ # GTIDs — all match ✅
pod-0: b4385007-...:1-271885:1085881-1128433
pod-1: b4385007-...:1-271885:1085881-1128433
pod-2: b4385007-...:1-271885:1085881-1128433

➤ # Checksums — all match ✅
pod-0: sbtest1=762643126, sbtest2=3738248816, sbtest3=3334660391, sbtest4=2415767356
pod-1: sbtest1=762643126, sbtest2=3738248816, sbtest3=3334660391, sbtest4=2415767356
pod-2: sbtest1=762643126, sbtest2=3738248816, sbtest3=3334660391, sbtest4=2415767356
```

**Result: PASS** — Zero data loss. Failover completed in ~5 seconds. All GTIDs and checksums match across all 3 nodes.

Clean up:

```shell
➤ kubectl delete -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org "mysql-primary-pod-kill" deleted
```

---

### Chaos#2: OOMKill the Primary Pod

Now we stress the primary pod's memory to exceed its 1.5Gi limit and see if it triggers an OOMKill.

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

**What this chaos does:** Allocates 1200MB of extra memory on the primary pod. Combined with MySQL's memory usage, this can exceed the 1.5Gi limit and trigger an OOMKill.

First, start the sysbench load test, then apply the chaos while writes are in-flight:

```shell
➤ kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
    --mysql-host=mysql-ha-cluster --mysql-port=3306 \
    --mysql-user=root --mysql-password="$PASS" \
    --mysql-db=sbtest --tables=12 --table-size=100000 \
    --threads=4 --time=120 --report-interval=10 run

[ 10s ] thds: 4 tps: 770.97 qps: 4627.51 lat (ms,95%): 9.22 err/s: 0.00
```

Now apply the memory stress while sysbench is running:

```shell
➤ kubectl apply -f tests/02-oomkill.yaml
stresschaos.chaos-mesh.org/mysql-primary-memory-stress created
```

Watch the database status during the stress:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    18h

NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          3m    standby
mysql-ha-cluster-1               2/2     Running   0          83m   standby
mysql-ha-cluster-2               2/2     Running   0          32m   primary
```

On MySQL 8.4.8, the primary **survived** the memory stress — it did not OOMKill. The database stayed `Ready` throughout. TPS degraded from ~770 to ~258 at the lowest point, but writes continued with zero errors:

```shell
[ 10s ] thds: 4 tps: 770.97 lat (ms,95%): 9.22 err/s: 0.00   # before stress
[ 20s ] thds: 4 tps: 426.20 lat (ms,95%): 19.65 err/s: 0.00  # stress applied
[ 80s ] thds: 4 tps: 257.90 lat (ms,95%): 44.17 err/s: 0.00  # lowest point
[120s ] thds: 4 tps: 367.10 lat (ms,95%): 39.65 err/s: 0.00  # stabilized

SQL statistics:
    transactions:                        48558  (404.56 per sec.)
    ignored errors:                      0      (0.00 per sec.)
```

Verify data integrity:

```shell
➤ # GR Members — all ONLINE, primary unchanged (no failover needed)
MEMBER_HOST                                         MEMBER_PORT   MEMBER_STATE   MEMBER_ROLE
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo       3306          ONLINE         PRIMARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY

➤ # GTIDs — all match ✅
pod-0: b4385007-...:1-320527:1085881-1128433
pod-1: b4385007-...:1-320527:1085881-1128433
pod-2: b4385007-...:1-320527:1085881-1128433

➤ # Checksums — all match ✅
pod-0: sbtest1=4287163059, sbtest2=2043983926, sbtest3=3614909082, sbtest4=164911003
pod-1: sbtest1=4287163059, sbtest2=2043983926, sbtest3=3614909082, sbtest4=164911003
pod-2: sbtest1=4287163059, sbtest2=2043983926, sbtest3=3614909082, sbtest4=164911003
```

**Result: PASS** — MySQL 8.4.8 survived the memory stress without OOMKill. Zero data loss, zero errors. TPS degraded ~47% under memory pressure but recovered immediately after stress removed.

> **Note:** MySQL 8.4.8 handles memory allocation more conservatively than 9.6.0. The same test triggers OOMKill on MySQL 9.6.0, which also passes with zero data loss after automatic failover.

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

Before running:

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          28m   primary
mysql-ha-cluster-1               2/2     Running   0          108m  standby
mysql-ha-cluster-2               2/2     Running   0          56m   standby
```

Apply the chaos:

```shell
➤ kubectl apply -f tests/03-network-partition.yaml
networkchaos.chaos-mesh.org/mysql-primary-network-partition created
```

Within ~20 seconds (the `group_replication_unreachable_majority_timeout`), the isolated primary loses quorum. The standbys elect a new primary, and the database goes `Critical`:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS     AGE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Critical   18h

NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          28m   standby
mysql-ha-cluster-1               2/2     Running   0          108m  standby
mysql-ha-cluster-2               2/2     Running   0          56m   primary
```

`Critical` means the new primary (pod-2) is accepting connections and operational, but the isolated pod-0 is not yet back in the group. Let's check GR from pod-2 (the new primary):

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
  FROM performance_schema.replication_group_members;
MEMBER_HOST                                         MEMBER_PORT   MEMBER_STATE   MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo       3306          ONLINE         PRIMARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY
```

Only 2 members visible — pod-0 is isolated. After the 2-minute partition expires, the coordinator automatically restarts the isolated node and it rejoins the group:

```shell
➤ kubectl get mysql,pods -n demo -L kubedb.com/role
NAME                                VERSION   STATUS   AGE
mysql.kubedb.com/mysql-ha-cluster   8.4.8     Ready    18h

NAME                             READY   STATUS    RESTARTS   AGE   ROLE
mysql-ha-cluster-0               2/2     Running   0          29m   standby
mysql-ha-cluster-1               2/2     Running   0          109m  standby
mysql-ha-cluster-2               2/2     Running   0          57m   primary
```

Verify data integrity:

```shell
➤ SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE
  FROM performance_schema.replication_group_members;
MEMBER_HOST                                         MEMBER_PORT   MEMBER_STATE   MEMBER_ROLE
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo       3306          ONLINE         PRIMARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo       3306          ONLINE         SECONDARY

➤ # GTIDs — all match ✅
pod-0: b4385007-...:1-320695:1085881-1128519
pod-1: b4385007-...:1-320695:1085881-1128519
pod-2: b4385007-...:1-320695:1085881-1128519

➤ # Checksums — all match ✅
pod-0: sbtest1=4287163059, sbtest2=2043983926, sbtest3=3614909082, sbtest4=164911003
pod-1: sbtest1=4287163059, sbtest2=2043983926, sbtest3=3614909082, sbtest4=164911003
pod-2: sbtest1=4287163059, sbtest2=2043983926, sbtest3=3614909082, sbtest4=164911003
```

**Result: PASS** — Network partition triggered failover in ~20 seconds. Isolated node rejoined automatically after partition removed. Zero data loss. All GTIDs and checksums match.

Clean up:

```shell
➤ kubectl delete -f tests/03-network-partition.yaml
networkchaos.chaos-mesh.org "mysql-primary-network-partition" deleted
```

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

**All 12 experiments PASSED with zero data loss.**

## Extended Experiments — Details (MySQL 8.4.8 Single-Primary)

### Exp 13: Double Primary Kill

Kill the primary, wait for new election, then immediately kill the new primary. Tests survival of two consecutive leader failures.

```bash
# Kill first primary
kubectl delete pod mysql-ha-cluster-2 -n demo --force --grace-period=0
# Wait 15s for new primary election, then kill the new primary
sleep 15
NEW_PRIMARY=$(kubectl get pods -n demo \
  -l "app.kubernetes.io/instance=mysql-ha-cluster,kubedb.com/role=primary" \
  -o jsonpath='{.items[0].metadata.name}')
kubectl delete pod $NEW_PRIMARY -n demo --force --grace-period=0
```

**Result:** Pod-0 was elected as the third primary. Cluster recovered in ~90 seconds. Zero data loss.

### Exp 14: Rolling Restart (0→1→2)

Simulate a rolling upgrade — delete each pod sequentially with 40-second gaps under write load.

```bash
kubectl delete pod mysql-ha-cluster-0 -n demo --force --grace-period=0
sleep 40
kubectl delete pod mysql-ha-cluster-1 -n demo --force --grace-period=0
sleep 40
kubectl delete pod mysql-ha-cluster-2 -n demo --force --grace-period=0
```

**Result:** Each pod recovered and rejoined within ~30 seconds. Two failovers triggered (when primary was deleted). Zero data loss.

### Exp 15: Coordinator Crash

Kill only the mysql-coordinator sidecar container, leaving MySQL running. Tests whether the cluster stays stable without coordinator.

```bash
kubectl exec -n demo mysql-ha-cluster-1 -c mysql-coordinator -- kill 1
```

**Result:** Kubernetes auto-restarted the coordinator container. MySQL was completely unaffected — no failover, no write interruption, 728 TPS (normal). The coordinator is a management layer; MySQL GR operates independently.

### Exp 16: Long Network Partition (10 min)

Isolate the primary from replicas for 10 minutes — 5x longer than the standard 2-minute test.

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

**Result:** Failover triggered within seconds. After 10-minute partition removed, the isolated node rejoined cleanly via GR distributed recovery. All 3 nodes ONLINE within ~2 minutes. Zero data loss.

### Exp 17: DNS Failure on Primary

Block all DNS resolution on the primary for 3 minutes. GR uses hostnames for inter-node communication, so this tests a critical infrastructure dependency.

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

**Result:** Primary survived without failover. TPS dropped ~32% (497 vs ~730 baseline) due to DNS-dependent operations timing out. No errors, no data loss. Existing TCP connections between GR members stayed open.

### Exp 18: PVC Delete + Pod Kill

Completely destroy a node's data — delete both the pod and its PVC. The node must rebuild from scratch using the CLONE plugin.

```bash
kubectl delete pod mysql-ha-cluster-0 -n demo --force --grace-period=0
kubectl delete pvc data-mysql-ha-cluster-0 -n demo
```

**Result:** StatefulSet auto-created a new PVC. The CLONE plugin copied a full data snapshot from a donor node. Pod recovered and joined GR in ~90 seconds with identical GTIDs and checksums. This is the ultimate recovery test — MySQL 8.0+ handles it fully automatically.

### Exp 19: IO Fault (EIO Errors)

Inject I/O read/write errors (errno 5 = EIO) on 50% of disk operations on the primary's data volume. Unlike IO latency which slows things down, IO faults cause actual operation failures — simulating a failing disk or storage system.

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

**Result:** Initially 703 TPS (InnoDB retries handle some errors), but the 50% EIO rate eventually crashed the MySQL process on the primary. Failover triggered to a secondary. After chaos removed and pod force-restarted, InnoDB crash recovery repaired the data directory and the node rejoined GR cleanly. Zero data loss.

### Exp 20: Clock Skew (-5 min)

Shift the primary's system clock back by 5 minutes. GR uses timestamps for conflict detection, Paxos message ordering, and timeout calculations. Tests whether clock drift breaks consensus or causes split-brain.

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

**Result:** 404 TPS (~45% reduction from baseline). No failover triggered, no errors. GR's Paxos protocol is resilient to clock drift — it uses logical clocks for consensus, not wall-clock time. All 3 nodes stayed ONLINE throughout. Zero data loss.

### Exp 21: Bandwidth Throttle (1mbps)

Limit the primary's outbound network bandwidth to 1mbps. Simulates degraded network in cross-AZ or cross-region deployments where bandwidth is limited but not broken.

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

**Result:** 147 TPS (~80% reduction from baseline). The bandwidth limit heavily throttles GR's Paxos consensus traffic, but the cluster stays completely stable — no failover, no errors, no member state changes. All 3 nodes remained ONLINE. Zero data loss.

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

| Aspect | Multi-Primary | Single-Primary |
|---|---|---|
| Failover needed | No (all primaries) | Yes (election ~2-3s) |
| Write availability | All nodes writable | Only primary writable |
| CPU stress 98% | All writes blocked (Paxos fails) | ~46% TPS reduction |
| IO latency impact | ~73% TPS drop | ~99.9% TPS drop |
| Packet loss 30% | 4.98 TPS (stayed ONLINE) | Triggers failover |
| High concurrency | GR certification conflicts possible | No conflicts (single writer) |
| Recovery mechanism | Rejoin as PRIMARY | Election + rejoin |

## Version Compatibility

| Capability | 8.0.36 | 8.4.8 | 9.6.0 |
|---|---|---|---|
| Pod Kill Recovery | Yes | Yes | Yes |
| OOMKill Recovery | Yes | Yes | Yes |
| Network Partition Recovery | Yes | Yes | Yes |
| CLONE Plugin | Yes | Yes | Yes |
| Single-Primary (core 12) | **12/12** | **12/12** | **12/12** |
| Single-Primary (extended 13-21) | Not tested | **9/9** | Not tested |
| Multi-Primary (12 tests) | Not tested | **12/12** | Not tested |

## Key Takeaways

1. **KubeDB MySQL achieves zero data loss** across all 57 chaos experiments in both Single-Primary and Multi-Primary topologies.

2. **Automatic failover works reliably** — primary election completes in 2-3 seconds, full recovery in under 4 minutes for all scenarios, including double primary kill and disk failure.

3. **Multi-Primary mode is production-ready** — all 12 experiments passed on MySQL 8.4.8. Be aware that multi-primary has higher sensitivity to CPU stress and network issues due to Paxos consensus requirements on all writable nodes.

4. **Full data rebuild works automatically** — even after complete PVC deletion, the CLONE plugin rebuilds a node from scratch in ~90 seconds with zero manual intervention.

5. **Coordinator crash has zero impact** — MySQL GR operates independently of the coordinator sidecar. Killing the coordinator does not trigger failover or interrupt writes.

6. **Disk failures trigger safe failover** — 50% I/O error rate eventually crashes MySQL, but InnoDB crash recovery + GR distributed recovery handles it with zero data loss after pod restart.

7. **Clock skew and bandwidth limits are tolerated** — GR's Paxos protocol is resilient to 5-minute clock drift (~45% TPS drop, no errors) and 1mbps bandwidth limits (~80% TPS drop, no errors).

8. **Transient GTID mismatches are normal** — brief mismatches (15-30 seconds) during recovery are expected and resolve automatically via GR distributed recovery.

## What's Next

- **Multi-Primary testing on additional MySQL versions** — extend chaos testing to MySQL 9.6.0 in Multi-Primary mode
- **Long-duration soak testing** — extended chaos runs (hours/days) to validate stability under sustained failure injection

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
