---
title: 'Chaos Engineering Results: KubeDB MySQL Achieves Zero Data Loss Across 69 Experiments — Group Replication & InnoDB Cluster'
date: "2026-04-08"
weight: 14
authors:
- SK Ali Arman
tags:
- chaos-engineering
- chaos-mesh
- database
- group-replication
- innodb-cluster
- high-availability
- kubedb
- kubernetes
- mysql
- mysql-router
---

## Overview

We conducted **69 chaos experiments** across **3 MySQL versions** (8.0.36, 8.4.8, 9.6.0), **2 Group Replication topologies** (Single-Primary and Multi-Primary), and **InnoDB Cluster with MySQL Router** on KubeDB-managed 3-node clusters. The goal: validate that KubeDB MySQL delivers **zero data loss**, **automatic failover**, and **self-healing recovery** under realistic failure conditions with production-level write loads.

**The result: every experiment passed with zero data loss, zero split-brain, and zero errant GTIDs.**

This post summarizes the methodology, results, and key findings from the most comprehensive chaos testing effort we have run on KubeDB MySQL to date — now including **InnoDB Cluster mode** with MySQL Router's read-write split capabilities.

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
| Cluster Topology | 3-node Group Replication (Single-Primary & Multi-Primary) + InnoDB Cluster with MySQL Router |
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

We will ignore the sysbench load for this experiment and focus on the failover behavior.

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

Note the `Restarts: 1` on pod-2 — it was OOMKilled and restarted by Kubernetes. The status `NotReady` means failover is in progress. After ~60 seconds, the cluster is fully recovered:

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

Additionally, **12 InnoDB Cluster experiments** (with MySQL Router) all PASSED — see the [InnoDB Cluster section](#innodb-cluster-mode-mysql-848--mysql-router) below for details.

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

---

## InnoDB Cluster Mode (MySQL 8.4.8 + MySQL Router)

InnoDB Cluster is MySQL's integrated high-availability solution that combines Group Replication with MySQL Router for automatic connection routing. Unlike standalone Group Replication where applications connect directly to MySQL pods, InnoDB Cluster adds a Router layer that automatically routes:

- **Port 6446** (Read-Write) — routes to the PRIMARY
- **Port 6447** (Read-Only) — round-robin across SECONDARYs
- **Port 6450** (Read-Write Split) — single port with automatic routing: writes go to PRIMARY, reads distributed across all members (MySQL 8.2+)

### InnoDB Cluster Setup

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

### Verify Router Service

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

### Sysbench via Router

All InnoDB Cluster chaos experiments use the Router service for load generation:

```shell
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

Applied 1-second network delay between primary and replicas:

```shell
[ 10s ] thds: 8 tps: 0.80 qps: 8.80 lat (ms,95%): 7615.89 err/s: 0.00
[ 30s ] thds: 8 tps: 1.10 qps: 6.60 lat (ms,95%): 6476.48 err/s: 0.00
[ 60s ] thds: 8 tps: 1.20 qps: 7.20 lat (ms,95%): 6476.48 err/s: 0.00
```

Average TPS: 1.26 (99.9% reduction). 95th percentile latency: 7,616ms. Zero errors. No failover.

**Result: PASS** — GR tolerated extreme latency. Zero data loss.

### InnoDB Chaos#6: CPU Stress (98%) + Write Load

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

Applied 30% packet loss on all cluster pods. Sysbench could not connect through the Router (error 2003), but the cluster itself remained stable — all 3 members stayed ONLINE with no failover.

This is a notable difference from Group Replication mode, where 30% packet loss triggered a failover.

**Result: PASS** — Cluster stable despite packet loss. Router connections failed but no data loss.

### InnoDB Chaos#8: Full Cluster Kill

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

Killed the primary, waited for a new primary to be elected, then immediately killed the new primary:

```
First kill:  pod-0 (PRIMARY) → pod-2 elected as new PRIMARY
Second kill: pod-2 (PRIMARY) → pod-1 elected as third PRIMARY
```

Both killed pods rejoined as SECONDARY after restart. All GTIDs and checksums matched.

**Result: PASS** — Survived two consecutive primary failures. Zero data loss.

### InnoDB Chaos#10: Rolling Restart (0→1→2)

Sequentially force-deleted pod-0, pod-1, and pod-2. pod-1 maintained PRIMARY throughout the rolling restart. pod-2 required the coordinator's 10-attempt restart cycle before rejoining.

**Result: PASS** — Rolling restart completed with zero data loss.

### InnoDB Chaos#11: DNS Failure on Primary

Applied DNS error mode on the primary for 3 minutes. No impact — GR uses IP addresses for group communication. Cluster remained Ready, no failover.

**Result: PASS** — DNS failure has no impact on InnoDB Cluster. Zero data loss.

### InnoDB Chaos#12: Clock Skew (-5 min)

Shifted the primary's clock back by 5 minutes:

```shell
# During chaos — primary shows time 5 min behind secondaries
Primary (pod-1): 2026-04-13 08:24:44
Secondary (pod-0): 2026-04-13 08:29:44
```

No impact — GR's Paxos protocol uses logical clocks for consensus, not wall-clock time. Cluster remained Ready, no failover.

**Result: PASS** — Clock skew tolerated. Zero data loss.

### InnoDB Cluster vs Group Replication — Key Differences

| Aspect | Group Replication (8.4.8) | InnoDB Cluster (8.4.8) |
|---|---|---|
| OOMKill (1200MB stress) | Survived (no OOMKill) | OOMKill triggered, failover |
| Packet Loss (30%) | Failover triggered | No failover (stable) |
| Connection routing | Direct to MySQL pods | Via MySQL Router (auto-failover) |
| RW-Split (port 6450) | N/A | Available (auto-enabled in 8.4+) |
| Recovery mechanism | Coordinator signals | Same + `rebootClusterFromCompleteOutage` |
| Rejoin after expulsion | 10-attempt restart cycle | Same 10-attempt restart cycle |

### Router-Specific Observations

1. **Automatic failover re-route**: Router detects primary changes via metadata cache (TTL=0.5s) and re-routes RW traffic within seconds
2. **During pod kill**: Existing connections get "Lost connection" (error 2013) — new connections route to new primary after re-route
3. **Port 6450 (RW-Split)** is auto-enabled by MySQL Router 8.4+ bootstrap — no custom configuration needed
4. **Packet loss**: Router could not establish new connections during 30% packet loss, but the cluster itself remained stable

## Failover Performance (Single-Primary)

| Scenario | Failover Time | Full Recovery Time |
|---|---|---|
| Pod Kill Primary | ~2-3 seconds | ~30-33 seconds |
| OOMKill Primary | ~2-3 seconds | ~30 seconds |
| Network Partition | ~3 seconds | ~3 minutes |
| Packet Loss (30%) | ~30 seconds | ~2 minutes |
| Full Cluster Kill | ~10 seconds | ~1-2 minutes |
| Combined Stress (OOMKill) | ~3 seconds | ~4 minutes |

### InnoDB Cluster Failover (via Router)

| Scenario | Failover Time | Router Re-route | Full Recovery Time |
|---|---|---|---|
| Pod Kill Primary | ~3 seconds | ~5 seconds | ~30 seconds |
| OOMKill Primary | ~3 seconds | ~5 seconds | ~60 seconds |
| Network Partition | ~5 seconds | ~5 seconds | ~90 seconds |
| Full Cluster Kill | ~10 seconds | After reboot | ~60 seconds |
| Double Primary Kill | ~3 seconds (x2) | Auto (x2) | ~120 seconds |

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

### InnoDB Cluster (via Router port 6446)

| Chaos Type | TPS During Chaos | Impact |
|---|---|---|
| IO Latency (100ms) | 0.1-0.2 (recovered to 1242) | ~99.9% during, full recovery after |
| Network Latency (1s) | 1.26 | 99.9% drop |
| CPU Stress (98%) | 763 avg (dipped to 188) | ~25% avg drop |
| Packet Loss (30%) | Router connection failed | Cluster stable, no failover |
| RW-Split (port 6450) | 555 TPS (`oltp_read_write`) | Baseline for mixed workload |

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
| InnoDB Cluster (12 tests) | Not tested | **12/12** | Not tested |

## Key Takeaways

1. **KubeDB MySQL achieves zero data loss** across all 69 chaos experiments in Single-Primary, Multi-Primary, and InnoDB Cluster topologies.

2. **Automatic failover works reliably** — primary election completes in 2-3 seconds, full recovery in under 4 minutes for all scenarios, including double primary kill and disk failure.

3. **Multi-Primary mode is production-ready** — all 12 experiments passed on MySQL 8.4.8. Be aware that multi-primary has higher sensitivity to CPU stress and network issues due to Paxos consensus requirements on all writable nodes.

4. **Full data rebuild works automatically** — even after complete PVC deletion, the CLONE plugin rebuilds a node from scratch in ~90 seconds with zero manual intervention.

5. **Coordinator crash has zero impact** — MySQL GR operates independently of the coordinator sidecar. Killing the coordinator does not trigger failover or interrupt writes.

6. **Disk failures trigger safe failover** — 50% I/O error rate eventually crashes MySQL, but InnoDB crash recovery + GR distributed recovery handles it with zero data loss after pod restart.

7. **Clock skew and bandwidth limits are tolerated** — GR's Paxos protocol is resilient to 5-minute clock drift (~45% TPS drop, no errors) and 1mbps bandwidth limits (~80% TPS drop, no errors).

8. **Transient GTID mismatches are normal** — brief mismatches (15-30 seconds) during recovery are expected and resolve automatically via GR distributed recovery.

9. **InnoDB Cluster with MySQL Router adds transparent failover** — applications connecting through the Router (port 6446) experience automatic re-routing when the primary fails. The RW-Split port (6450) enables single-port read/write optimization with zero configuration.

10. **InnoDB Cluster survived all 12 chaos experiments** — including full cluster kill (auto-reboot from complete outage), double primary kill, and rolling restart. Router re-routed traffic automatically in every failover scenario.

## What's Next

- **Multi-Primary testing on additional MySQL versions** — extend chaos testing to MySQL 9.6.0 in Multi-Primary mode
- **Long-duration soak testing** — extended chaos runs (hours/days) to validate stability under sustained failure injection

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
