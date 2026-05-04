---
title: Chaos Engineering KubeDB MySQL InnoDB Cluster with MySQL Router on Kubernetes
date: "2026-04-13"
weight: 14
authors:
- SK Ali Arman
tags:
- chaos-engineering
- chaos-mesh
- database
- high-availability
- innodb-cluster
- kubedb
- kubernetes
- mysql
- mysql-router
---

## Overview

We conducted **12 chaos experiments** on **MySQL 8.4.8 InnoDB Cluster with MySQL Router** running on a KubeDB-managed 3-node cluster. The goal: validate that KubeDB MySQL InnoDB Cluster delivers **zero data loss**, **automatic failover**, **transparent connection re-routing**, and **self-healing recovery** under realistic failure conditions with production-level write loads.

**The result: every experiment passed with zero data loss, zero split-brain, and zero errant GTIDs.**

This post summarizes the methodology, results, and key findings from comprehensive chaos testing of KubeDB MySQL InnoDB Cluster with MySQL Router.

## Why Chaos Testing?

Running databases on Kubernetes introduces failure modes that traditional infrastructure does not have — pods can be evicted, nodes can go down, network policies can partition traffic, and resource limits can trigger OOMKills at any time. Chaos engineering deliberately injects these failures to verify that the system recovers correctly **before** they happen in production.

For a MySQL InnoDB Cluster managed by KubeDB, we needed to answer:

- Does the cluster **lose data** when a primary is killed mid-transaction?
- Does **MySQL Router automatically re-route** traffic to the new primary after failover?
- Can the cluster **self-heal** after a full outage with no manual intervention?
- Are **GTIDs consistent** across all nodes after recovery?
- Does the **RW-Split port (6450)** work correctly under chaos conditions?

## What is InnoDB Cluster?

InnoDB Cluster is MySQL's integrated high-availability solution that combines Group Replication with MySQL Router for automatic connection routing. Unlike standalone Group Replication where applications connect directly to MySQL pods, InnoDB Cluster adds a Router layer that automatically routes:

- **Port 6446** (Read-Write) — routes to the PRIMARY
- **Port 6447** (Read-Only) — round-robin across SECONDARYs
- **Port 6450** (Read-Write Split) — single port with automatic routing: writes go to PRIMARY, reads distributed across all members (MySQL 8.2+)

The key advantage is **transparent failover** — applications don't need to know which pod is the primary. They connect to the Router, and the Router handles the routing automatically.

## Test Environment

| Component | Details |
|---|---|
| Kubernetes | kind (local cluster) |
| KubeDB Version | 2026.2.26 |
| Cluster Topology | 3-node InnoDB Cluster + 1 MySQL Router |
| MySQL Version | 8.4.8 |
| Storage | 2Gi PVC per node (Durable, ReadWriteOnce) |
| Memory Limit | 1.5Gi per MySQL pod |
| CPU Request | 500m per pod |
| Chaos Engine | Chaos Mesh |
| Load Generator | sysbench `oltp_write_only` via Router (port 6446), 4 tables, 8 threads |
| Baseline TPS | ~1,100 (via Router) |

All experiments were run under **sustained sysbench write load** through the MySQL Router to simulate production traffic during failures.

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

### Step 4: Deploy InnoDB Cluster

Create the namespace:

```bash
kubectl create namespace demo
```

Deploy the InnoDB Cluster:

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
    mode: InnoDBCluster
    innoDBCluster:
      router:
        replicas: 1
  version: 8.4.8
```

This deploys 3 MySQL pods + 1 MySQL Router pod. The Router automatically discovers the cluster topology and exposes 3 ports via the Kubernetes Service: 6446 (RW), 6447 (RO), and 6450 (RW-Split).

Deploy and wait for Ready:

```bash
kubectl apply -f mysql-innodb-cluster.yaml
kubectl wait --for=jsonpath='{.status.phase}'=Ready mysql/mysql-ha-cluster -n demo --timeout=5m
```

### Step 5: Verify Router Service

```shell
$ kubectl get svc -n demo | grep router
mysql-ha-cluster-router    ClusterIP   10.96.196.159   <none>   6446/TCP,6447/TCP,6450/TCP   63m
```

Test each port:

```shell
# Port 6446 (RW) → routes to PRIMARY
$ mysql -h mysql-ha-cluster-router -P 6446 -e "SELECT @@hostname;"
@@hostname
mysql-ha-cluster-0

# Port 6447 (RO) → routes to SECONDARY
$ mysql -h mysql-ha-cluster-router -P 6447 -e "SELECT @@hostname;"
@@hostname
mysql-ha-cluster-1

# Port 6450 (RW-Split) → round-robin across all members
$ mysql -h mysql-ha-cluster-router -P 6450 -e "SELECT @@hostname;"
@@hostname
mysql-ha-cluster-2
```

### Step 6: Deploy sysbench Load Generator

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
              value: "mysql-ha-cluster-router.demo.svc.cluster.local"
            - name: MYSQL_PORT
              value: "6446"
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

### Step 7: Prepare sysbench Tables

```bash
# Get the MySQL root password
PASS=$(kubectl get secret mysql-ha-cluster-auth -n demo -o jsonpath='{.data.password}' | base64 -d)

# Create the sbtest database (via Router RW port)
kubectl exec -n demo mysql-ha-cluster-0 -c mysql -- \
  mysql -uroot -p"$PASS" -e "CREATE DATABASE IF NOT EXISTS sbtest;"

# Get the sysbench pod name
SBPOD=$(kubectl get pods -n demo -l app=sysbench -o jsonpath='{.items[0].metadata.name}')

# Prepare tables via Router (4 tables x 50k rows)
kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
  --mysql-host=mysql-ha-cluster-router.demo.svc.cluster.local \
  --mysql-port=6446 --mysql-user=root --mysql-password="$PASS" \
  --mysql-db=sbtest --tables=4 --table-size=50000 \
  --threads=8 prepare
```

### Step 8: Run sysbench via Router During Chaos

```bash
# Write load via RW port (6446)
kubectl exec -n demo $SBPOD -- sysbench oltp_write_only \
  --mysql-host=mysql-ha-cluster-router.demo.svc.cluster.local \
  --mysql-port=6446 --mysql-user=root --mysql-password="$PASS" \
  --mysql-db=sbtest --tables=4 --table-size=50000 \
  --threads=8 --time=60 --report-interval=10 run

# Read-write split load via RW-Split port (6450) — requires SSL
kubectl exec -n demo $SBPOD -- sysbench oltp_read_write \
  --mysql-host=mysql-ha-cluster-router.demo.svc.cluster.local \
  --mysql-port=6450 --mysql-user=root --mysql-password="$PASS" \
  --mysql-db=sbtest --tables=4 --table-size=50000 \
  --threads=8 --time=60 --report-interval=10 --mysql-ssl=REQUIRED run
```

> **Note:** Port 6450 (RW-Split) requires `--mysql-ssl=REQUIRED` because MySQL Router's `connection_sharing=1` with `caching_sha2_password` needs a secure connection. Ports 6446 and 6447 do not require this.

## Chaos Testing

We will run chaos experiments to see how our InnoDB Cluster behaves under failure scenarios like pod kill, OOM kill, network partition, network latency, IO latency, and more. We will use sysbench via the MySQL Router to simulate high write load on the cluster during each experiment.

> **Important Notes on Database Status:**
> - **`Ready`** — Database is fully operational. All pods are ONLINE.
> - **`Critical`** — Primary is accepting connections and operational, but one or more replicas may be down.
> - **`NotReady`** — Primary is not available. No writes can be accepted.
>
> You can read/write in your database in both **`Ready`** and **`Critical`** states. So even if your db is in `Critical` state, your uptime is not compromised.

### InnoDB Cluster Results Summary

| # | Experiment | Failover | Router Re-route | Data Loss | GTIDs | Checksums | Verdict |
|---|---|---|---|---|---|---|---|
| 1 | Pod Kill Primary | Yes | Auto | Zero | MATCH | MATCH | PASS |
| 2 | OOMKill Primary (1200MB stress) | Yes (OOMKill) | Auto | Zero | MATCH | MATCH | PASS |
| 3 | Network Partition (2 min) | Yes | Auto | Zero | MATCH | MATCH | PASS |
| 4 | IO Latency (100ms) | No | N/A | Zero | MATCH | MATCH | PASS |
| 5 | Network Latency (1s) | No | N/A | Zero | MATCH | MATCH | PASS |
| 6 | CPU Stress (98%) | No | N/A | Zero | MATCH | MATCH | PASS |
| 7 | Packet Loss (30%) | No | N/A | Zero | MATCH | MATCH | PASS |
| 8 | Full Cluster Kill | Yes | Auto | Zero | MATCH | MATCH | PASS |
| 9 | Double Primary Kill | Yes (x2) | Auto (x2) | Zero | MATCH | MATCH | PASS |
| 10 | Rolling Restart (0→1→2) | Yes | Auto | Zero | MATCH | MATCH | PASS |
| 11 | DNS Failure on Primary | No | N/A | Zero | MATCH | MATCH | PASS |
| 12 | Clock Skew (-5 min) | No | N/A | Zero | MATCH | MATCH | PASS |

**All 12 experiments passed with zero data loss.**

### InnoDB Chaos#1: Kill the Primary Pod

- **Expected behavior:**
  Primary killed → GR elects new primary → MySQL Router (metadata-cache TTL 0.5s) detects topology change and re-routes RW traffic on port 6446 → killed pod rejoins as secondary → cluster `Ready`. Zero data loss.

- **Actual result:**
  Pod-0 killed → pod-2 elected new primary → Router re-routed within seconds (confirmed via `SELECT @@hostname` on port 6446). Pod-0 rejoined as secondary. All 3 members `ONLINE`, GTIDs and checksums match. **PASS.**

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
    namespaces: [demo]
    labelSelectors:
      "app.kubernetes.io/instance": "mysql-ha-cluster"
      "kubedb.com/role": "primary"
  gracePeriod: 0
```

Before kill — pod-0 is PRIMARY:

```shell
$ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
MEMBER_HOST                                           MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         ONLINE        SECONDARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo         ONLINE        SECONDARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         ONLINE        PRIMARY
```

After kill — pod-2 elected as new PRIMARY, Router automatically re-routes:

```shell
$ SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
MEMBER_HOST                                           MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         ONLINE        PRIMARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo         ONLINE        SECONDARY

# Router automatically re-routed RW traffic to new primary
$ mysql -h mysql-ha-cluster-router -P 6446 -e "SELECT @@hostname;"
@@hostname
mysql-ha-cluster-2
```

After recovery — all 3 ONLINE, GTIDs MATCH, checksums MATCH:

```shell
# GTIDs — all match
pod-0: dcb462a5-...:1-4, e8ac711c-...:1-563:1000093-1000107
pod-1: dcb462a5-...:1-4, e8ac711c-...:1-563:1000093-1000107
pod-2: dcb462a5-...:1-4, e8ac711c-...:1-563:1000093-1000107

# Checksums — all match
pod-0: sbtest1=309802666
pod-1: sbtest1=309802666
pod-2: sbtest1=309802666
```

**Result: PASS** — Router detected the failover and re-routed within seconds. Zero data loss.

### InnoDB Chaos#2: OOMKill Primary (Memory Stress)

Applied 1200MB memory stress on the primary via StressChaos. Unlike Group Replication mode where 8.4.8 survived 1600MB stress, InnoDB Cluster with 1200MB triggered an OOMKill.

- **Expected behavior:**
  Memory stress pushes primary past 1.5Gi limit → OOMKill → GR expels member, new primary elected → Router re-routes RW traffic → killed pod restarts and rejoins → cluster `Ready`. Zero data loss.

- **Actual result:**
  Primary pod-2 OOMKilled (restart count 1). Pod-1 elected new primary. Router re-routed automatically. Pod-2 rejoined as secondary. All 3 members `ONLINE`, GTIDs and checksums match. **PASS.**

```shell
$ kubectl get pods -n demo -l app.kubernetes.io/instance=mysql-ha-cluster
NAME                        READY   STATUS    RESTARTS      AGE
mysql-ha-cluster-0          2/2     Running   0             89s
mysql-ha-cluster-1          2/2     Running   0             17m
mysql-ha-cluster-2          2/2     Running   1 (3s ago)    17m    # OOMKill restart
mysql-ha-cluster-router-0   1/1     Running   0             17m
```

pod-1 elected as new PRIMARY. After recovery — all 3 ONLINE, GTIDs MATCH, checksums MATCH.

**Result: PASS** — OOMKill triggered failover. Router re-routed automatically. Zero data loss.

### InnoDB Chaos#3: Network Partition

- **Expected behavior:**
  Primary isolated from standbys for 2 min → GR marks it `UNREACHABLE` → after timeout, expelled and new primary elected from remaining 2 nodes → Router re-routes RW to new primary → after partition lifts, expelled node rejoins via coordinator → cluster `Ready`. Zero data loss.

- **Actual result:**
  Pod-1 (old primary) shown `UNREACHABLE`, expelled. Pod-2 elected as new primary. Router re-routed. Pod-1 auto-rejoined ~90s after partition lifted. GTIDs and checksums match. **PASS.**

Partitioned the primary from standbys for 2 minutes:

```shell
# During partition — primary shown as UNREACHABLE
MEMBER_HOST                                           MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         ONLINE        SECONDARY
mysql-ha-cluster-1.mysql-ha-cluster-pods.demo         UNREACHABLE   PRIMARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         ONLINE        SECONDARY

# After GR timeout — pod-2 elected as new PRIMARY, pod-1 expelled
MEMBER_HOST                                           MEMBER_STATE  MEMBER_ROLE
mysql-ha-cluster-2.mysql-ha-cluster-pods.demo         ONLINE        PRIMARY
mysql-ha-cluster-0.mysql-ha-cluster-pods.demo         ONLINE        SECONDARY
```

After partition ended, the coordinator rejoined pod-1 automatically (~90s). All GTIDs and checksums matched.

**Result: PASS** — Partition caused failover. Expelled member auto-rejoined. Zero data loss.

### InnoDB Chaos#4: IO Latency (100ms) + Write Load via Router

- **Expected behavior:**
  IO latency on primary → TPS collapses (disk-bound writes) but cluster stays `Ready` → Router keeps connection to the degraded primary (it's still alive, just slow) → when chaos ends, TPS recovers. Zero data loss.

- **Actual result:**
  TPS dropped to ~0 during the 30s latency window, then recovered to 1242 TPS after chaos expired. No failover, no errors. GTIDs and checksums match. **PASS.**

Applied 100ms IO latency on the primary while running 8-thread sysbench write load through the Router (port 6446):

```shell
[ 10s ] thds: 8 tps: 0.10 qps: 1.90 lat (ms,95%): 1678.14 err/s: 0.00
[ 20s ] thds: 8 tps: 0.00 qps: 0.00 lat (ms,95%): 0.00 err/s: 0.00
[ 30s ] thds: 8 tps: 0.20 qps: 1.50 lat (ms,95%): 26861.48 err/s: 0.00
[ 40s ] thds: 8 tps: 0.00 qps: 0.00 lat (ms,95%): 0.00 err/s: 0.00
[ 50s ] thds: 8 tps: 254.30 qps: 1528.22 lat (ms,95%): 13.22 err/s: 0.00   # chaos expired
[ 60s ] thds: 8 tps: 1242.00 qps: 7452.01 lat (ms,95%): 11.24 err/s: 0.00   # full recovery
```

TPS dropped to near-zero during IO latency, then recovered to ~1242 TPS after chaos expired. No failover triggered, no errors. The Router maintained the connection to the degraded primary throughout.

**Result: PASS** — Severe performance degradation but no data loss. Zero errors.

### InnoDB Chaos#5: Network Latency (1s) + Write Load

- **Expected behavior:**
  1s delay on GR traffic → each Paxos commit waits ≥1s → TPS near zero but writes still succeed → no failover (delay within unreachable timeout) → recovers after chaos. Zero data loss.

- **Actual result:**
  Average TPS 1.26 (99.9% reduction), 95p latency 7.6s. Zero errors, no failover. All 3 members stayed `ONLINE`. GTIDs and checksums match. **PASS.**

Applied 1-second network delay between primary and replicas:

```shell
[ 10s ] thds: 8 tps: 0.80 qps: 8.80 lat (ms,95%): 7615.89 err/s: 0.00
[ 30s ] thds: 8 tps: 1.10 qps: 6.60 lat (ms,95%): 6476.48 err/s: 0.00
[ 60s ] thds: 8 tps: 1.20 qps: 7.20 lat (ms,95%): 6476.48 err/s: 0.00
```

Average TPS: 1.26 (99.9% reduction). 95th percentile latency: 7,616ms. Zero errors. No failover.

**Result: PASS** — GR tolerated extreme latency. Zero data loss.

### InnoDB Chaos#6: CPU Stress (98%) + Write Load

- **Expected behavior:**
  98% CPU stress on primary → TPS drops significantly → no failover (primary still responds to heartbeats) → TPS recovers when chaos ends. Zero data loss.

- **Actual result:**
  Baseline 1113 TPS dropped to 188 at peak stress, then stabilized around 810 as MySQL adapted. Average 763 TPS. No failover, zero errors. GTIDs and checksums match. **PASS.**

Applied 98% CPU stress on the primary:

```shell
[ 10s ] thds: 8 tps: 1113.75 qps: 6686.48 lat (ms,95%): 13.22 err/s: 0.00
[ 30s ] thds: 8 tps: 187.50 qps: 1125.00 lat (ms,95%): 137.35 err/s: 0.00   # stress kicks in
[ 50s ] thds: 8 tps: 877.90 qps: 5267.80 lat (ms,95%): 29.19 err/s: 0.00   # adapting
[ 60s ] thds: 8 tps: 810.70 qps: 4863.39 lat (ms,95%): 31.37 err/s: 0.00
```

Average TPS: 763. Dipped to 188 at peak stress, then recovered. No failover, zero errors.

**Result: PASS** — CPU stress caused temporary TPS drop. Zero data loss.

### InnoDB Chaos#7: Packet Loss (30%)

- **Expected behavior:**
  30% packet loss on all pods → Router and GR member-to-member connections degraded → cluster either stays `Ready` (if heartbeats get through) or sees `UNREACHABLE`/failover. Zero data loss either way.

- **Actual result:**
  Sysbench could not connect through Router (error 2003), but cluster internals stayed stable — all 3 members `ONLINE`, **no failover** triggered. Notable contrast with GR-only mode where 30% loss did cause an `UNREACHABLE` state. GTIDs and checksums match. **PASS.**

Applied 30% packet loss on all cluster pods. Sysbench could not connect through the Router (error 2003), but the cluster itself remained stable — all 3 members stayed ONLINE with no failover.

This is a notable difference from Group Replication mode, where 30% packet loss triggered a failover.

**Result: PASS** — Cluster stable despite packet loss. Router connections failed but no data loss.

### InnoDB Chaos#8: Full Cluster Kill

- **Expected behavior:**
  All 3 pods killed → no primary → coordinator detects complete outage, identifies pod with highest GTID → invokes `dba.rebootClusterFromCompleteOutage()` from that pod → other pods rejoin → cluster `Ready`. Zero data loss.

- **Actual result:**
  Pod-0 elected as bootstrap candidate (max transacted pod). `rebootClusterFromCompleteOutage` succeeded, pod-1 and pod-2 rejoined. Recovery in ~60s. All 3 members `ONLINE`, GTIDs and checksums match. **PASS.**

Force-deleted all 3 MySQL pods simultaneously:

```shell
$ kubectl delete pod -n demo mysql-ha-cluster-0 mysql-ha-cluster-1 mysql-ha-cluster-2 \
    --force --grace-period=0
```

The coordinator detected the complete outage and initiated `rebootClusterFromCompleteOutage`:

```
I0413 07:47:51 mysql.go:563] pod mysql-ha-cluster-0 is a superset of all other pods
I0413 07:47:51 mysql.go:622] max transacted pod: mysql-ha-cluster-0
I0413 07:47:51 mysql.go:435] all peers acknowledged bootstrap
I0413 07:47:51 mysql.go:364] cluster rebooting from complete outage
```

pod-0 was elected as bootstrap candidate (highest GTID), bootstrapped the cluster, and pod-1 and pod-2 rejoined. Full recovery in ~60 seconds.

**Result: PASS** — Cluster rebooted from complete outage automatically. Zero data loss.

### InnoDB Chaos#9: Double Primary Kill

- **Expected behavior:**
  First primary killed → new primary elected (Router re-routes) → second primary killed → third primary elected → both killed pods rejoin → cluster `Ready`. Zero data loss despite rapid leader churn.

- **Actual result:**
  Pod-0 killed → pod-2 elected → pod-2 killed → pod-1 elected as third primary. Both killed pods rejoined as secondary. Router re-routed correctly after each failover. GTIDs and checksums match across all 3 nodes. **PASS.**

Killed the primary, waited for a new primary to be elected, then immediately killed the new primary:

```
First kill:  pod-0 (PRIMARY) → pod-2 elected as new PRIMARY
Second kill: pod-2 (PRIMARY) → pod-1 elected as third PRIMARY
```

Both killed pods rejoined as SECONDARY after restart. All GTIDs and checksums matched.

**Result: PASS** — Survived two consecutive primary failures. Zero data loss.

### InnoDB Chaos#10: Rolling Restart (0→1→2)

- **Expected behavior:**
  Pods deleted sequentially (0 → 1 → 2) → each rejoins before the next is killed → primary maintained throughout (unless hit) → cluster returns to `Ready` between steps. Zero data loss.

- **Actual result:**
  Pod-1 maintained PRIMARY throughout. Pod-2 needed the coordinator's 10-attempt restart cycle before rejoining (known behavior for rapid rejoin), then joined cleanly. All 3 members `ONLINE`, GTIDs and checksums match. **PASS.**

Sequentially force-deleted pod-0, pod-1, and pod-2. pod-1 maintained PRIMARY throughout the rolling restart. pod-2 required the coordinator's 10-attempt restart cycle before rejoining.

**Result: PASS** — Rolling restart completed with zero data loss.

### InnoDB Chaos#11: DNS Failure on Primary

- **Expected behavior:**
  DNS resolution blocked on primary → no impact on existing TCP connections (GR uses IP addresses internally) → cluster stays `Ready`, no failover. Zero data loss.

- **Actual result:**
  Cluster stayed `Ready` throughout the 3-minute DNS chaos. No failover, no errors. GTIDs and checksums match. **PASS.**

Applied DNS error mode on the primary for 3 minutes. No impact — GR uses IP addresses for group communication. Cluster remained Ready, no failover.

**Result: PASS** — DNS failure has no impact on InnoDB Cluster. Zero data loss.

### InnoDB Chaos#12: Clock Skew (-5 min)

- **Expected behavior:**
  Primary wall clock shifted -5 min → GR's Paxos uses logical clocks, so consensus unaffected → no failover → cluster stays `Ready`. Zero data loss.

- **Actual result:**
  Primary showed time 5 min behind secondaries; cluster stayed `Ready` throughout, no failover, no errors. GTIDs and checksums match. **PASS** — confirms logical-clock-based Paxos.

Shifted the primary's clock back by 5 minutes:

```shell
# During chaos — primary shows time 5 min behind secondaries
Primary (pod-1): 2026-04-13 08:24:44
Secondary (pod-0): 2026-04-13 08:29:44
```

No impact — GR's Paxos protocol uses logical clocks for consensus, not wall-clock time. Cluster remained Ready, no failover.

**Result: PASS** — Clock skew tolerated. Zero data loss.

### Router-Specific Observations

1. **Automatic failover re-route**: Router detects primary changes via metadata cache (TTL=0.5s) and re-routes RW traffic within seconds
2. **During pod kill**: Existing connections get "Lost connection" (error 2013) — new connections route to new primary after re-route
3. **Port 6450 (RW-Split)** is auto-enabled by MySQL Router 8.4+ bootstrap — no custom configuration needed
4. **Packet loss**: Router could not establish new connections during 30% packet loss, but the cluster itself remained stable


## Failover Performance (InnoDB Cluster via Router)

| Scenario | Failover Time | Router Re-route | Full Recovery Time |
|---|---|---|---|
| Pod Kill Primary | ~3 seconds | ~5 seconds | ~30 seconds |
| OOMKill Primary | ~3 seconds | ~5 seconds | ~60 seconds |
| Network Partition | ~5 seconds | ~5 seconds | ~90 seconds |
| Full Cluster Kill | ~10 seconds | After reboot | ~60 seconds |
| Double Primary Kill | ~3 seconds (x2) | Auto (x2) | ~120 seconds |

## Performance Impact Under Chaos

| Chaos Type | TPS During Chaos | Impact |
|---|---|---|
| IO Latency (100ms) | 0-254 | ~100% drop during, then recovery |
| Network Latency (1s) | 1.26 | 99.9% drop |
| CPU Stress (98%) | 187-877 | ~30-80% drop |
| Packet Loss (30%) | 0 (connection failed) | Router connections blocked |
| Clock Skew (-5 min) | Normal | No impact |
| DNS Failure | Normal | No impact |

## Key Takeaways

1. **KubeDB MySQL InnoDB Cluster achieves zero data loss** across all 12 chaos experiments with transparent failover via MySQL Router.

2. **MySQL Router automatically re-routes traffic** — applications connecting through the Router (port 6446) experience automatic re-routing when the primary fails. The Router detects topology changes via its metadata cache (TTL=0.5s) and re-routes within seconds.

3. **Full cluster recovery works automatically** — even after all 3 pods are killed simultaneously, the coordinator detects the complete outage, identifies the pod with the highest GTID, and invokes `dba.rebootClusterFromCompleteOutage()` to restore the cluster.

4. **RW-Split port (6450) provides single-port optimization** — applications can use port 6450 for automatic read/write routing without any application changes. Writes go to PRIMARY, reads are distributed across all members.

5. **Packet loss affects Router connectivity but not cluster stability** — during 30% packet loss, the Router couldn't establish new connections, but the cluster itself remained stable with all members ONLINE.

6. **Clock skew and DNS failures have no impact** — GR's Paxos protocol uses logical clocks for consensus, and GR uses IP addresses internally for group communication.

7. **InnoDB Cluster vs Group Replication** — the key difference is transparent failover. With standalone GR, applications need to handle primary discovery themselves. With InnoDB Cluster, the Router handles it automatically.

## InnoDB Cluster vs Group Replication

| Aspect | InnoDB Cluster | Group Replication |
|---|---|---|
| Connection routing | Via MySQL Router (automatic) | Direct to pods (manual) |
| Failover handling | Transparent (Router re-routes) | Application must reconnect |
| Read-Write Split | Built-in (port 6450) | Application logic required |
| Additional component | MySQL Router pod | None |
| Best for | Applications needing transparent failover | Applications with custom routing |

## What's Next

- **InnoDB Cluster Multi-Primary testing** — test Multi-Primary topology with MySQL Router
- **MySQL 9.6.0 InnoDB Cluster** — validate all experiments on MySQL 9.6.0
- **Long-duration soak testing** — extended chaos runs (hours/days) to validate stability under sustained failure injection

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
