---
title: Chaos Testing KubeDB MariaDB on Kubernetes, Testing Galera Cluster Resilience
date: "2026-04-17"
weight: 25
authors:
- SK Ali Arman
tags:
- chaos-engineering
- chaos-mesh
- database
- galera-cluster
- high-availability
- kubedb
- kubernetes
- mariadb
---

> New to KubeDB? Please start [here](https://kubedb.com/docs/v2026.2.26/welcome/).

# Chaos Testing KubeDB Managed MariaDB Galera Cluster with Chaos Mesh

## Setup Cluster

To follow along with this tutorial, you will need:

1. A running Kubernetes cluster.
2. KubeDB [installed](https://kubedb.com/docs/v2026.2.26/setup/install/kubedb/) in your cluster.
3. kubectl command-line tool configured to communicate with your cluster.
4. Chaos-Mesh [installed](https://chaos-mesh.org/docs/production-installation-using-helm/) in your cluster.
    ```shell
    helm upgrade -i chaos-mesh chaos-mesh/chaos-mesh \
     -n chaos-mesh \
    --create-namespace \
    --set dashboard.create=true \
    --set dashboard.securityMode=false \
    --set chaosDaemon.runtime=containerd \
    --set chaosDaemon.socketPath=/run/containerd/containerd.sock \
    --set chaosDaemon.privileged=true
    ```
> Note: Make sure to set correct path to your container runtime socket and runtime in the above command. For ex: `socketPath=/run/containerd/containerd.sock`, or if in **k3s**, set `chaosDaemon.socketPath=/run/k3s/containerd/containerd.sock`.

## Introduction to Chaos Engineering

**Chaos Engineering** is a disciplined approach to testing distributed systems by deliberately introducing controlled failure scenarios to discover vulnerabilities and weaknesses before they impact your users. Rather than waiting for production incidents, chaos engineering proactively identifies how your system behaves under adverse conditions — such as pod failures, network outages, resource exhaustion, and data corruption.

This methodology is particularly crucial for database systems, where failures can lead to data loss, service downtime, and compromised data consistency. By testing these scenarios in controlled environments, you gain confidence that your system can recover gracefully and maintain availability.

### What This Blog Covers

In this comprehensive guide, we will:

1. **Deploy a Highly Available MariaDB Galera Cluster** on Kubernetes using KubeDB, configured with synchronous multi-master replication
2. **Run 18 Chaos Engineering Experiments** using Chaos Mesh to simulate real-world failure scenarios
3. **Observe Cluster Behavior** during failures including pod crashes, network issues, resource exhaustion, and disk I/O errors
4. **Measure Resilience** by tracking data consistency, failover speed, and recovery capabilities
5. **Learn Best Practices** for configuring MariaDB Galera Cluster for maximum resilience

Each experiment progressively tests different aspects of the system — from simple pod failures to complex scenarios involving packet corruption and IO data manipulation. By the end, you'll have a thorough understanding of how your MariaDB Galera cluster behaves under various failure modes.

You can see the [`Chaos Testing Results Summary`](#chaos-testing-results-summary) for a quick view of what we have done in this blog.

## Test Environment

| Component | Details |
|---|---|
| Kubernetes | kind (local cluster) |
| KubeDB Version | 2026.2.26 |
| Cluster Topology | 3-node Galera Cluster (all nodes read-write) |
| MariaDB Version | 11.8.5 |
| Storage | 2Gi PVC per node (Durable, ReadWriteOnce) |
| Memory Limit | 1.5Gi per MariaDB pod |
| CPU Request | 500m per pod |
| Chaos Engine | Chaos Mesh |
| Load Generator | sysbench `oltp_read_write`, 4 tables x 50k rows, 4 threads |
| Baseline TPS | ~1,039 |

All experiments were run under **sustained sysbench read-write load** to simulate production traffic during failures.

## Create a High-Availability MariaDB Galera Cluster

First, we need to deploy a MariaDB cluster configured for High Availability. Unlike a Standalone instance, a Galera Cluster consists of 3 multi-master nodes that all accept reads and writes simultaneously. If any node fails, the remaining nodes continue serving traffic with zero downtime.

Save the following YAML as `setup/kubedb-mariadb.yaml`:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: md
  namespace: demo
spec:
  deletionPolicy: Delete
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  storageType: Durable
  podTemplate:
    spec:
      containers:
        - name: mariadb
          resources:
            limits:
              memory: 1.5Gi
            requests:
              cpu: 500m
              memory: 1.5Gi
  topology:
    mode: GaleraCluster
  version: 11.8.5
```

> **Important Notes:**
> - You can read/write in your database in both **`Ready`** and **`Critical`** states. So it means even if your db is in `Critical` state, your uptime is not compromised. `Critical` means one or more nodes are offline, but the remaining nodes have quorum and accept connections.
> - All the results/metrics shown in this blog are related to the chaos scenarios. In general, a **Galera node recovery takes ~5-30 seconds** with **zero data loss**, ensuring high availability and data safety.

Now, create the namespace and apply the manifest:

```shell
kubectl create ns demo
kubectl apply -f setup/kubedb-mariadb.yaml
```

You can monitor the status until all pods are ready:

```shell
watch kubectl get mariadb,pods -n demo -L kubedb.com/role
```

See the database status is ready:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                       VERSION   STATUS   AGE
mariadb.kubedb.com/md      11.8.5    Ready    68m

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          68m   Primary
pod/md-1   2/2     Running   0          68m   Primary
pod/md-2   2/2     Running   0          68m   Primary
```

Note: In Galera Cluster, **all nodes have role `Primary`** because every node accepts reads and writes. This is different from MySQL Group Replication or PostgreSQL where you have a single primary and standbys.

### Deploy sysbench Load Generator

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
              value: "md.demo.svc.cluster.local"
            - name: MYSQL_PORT
              value: "3306"
            - name: MYSQL_USER
              value: "root"
            - name: MYSQL_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: md-auth
                  key: password
            - name: MYSQL_DB
              value: "sbtest"
```

```bash
kubectl apply -f sysbench.yaml
```

### Prepare sysbench Tables

```bash
# Get the MariaDB root password
PASS=$(kubectl get secret md-auth -n demo -o jsonpath='{.data.password}' | base64 -d)

# Create the sbtest database
kubectl exec -n demo md-0 -c mariadb -- \
  mariadb -uroot -p"$PASS" -e "CREATE DATABASE IF NOT EXISTS sbtest;"

# Get the sysbench pod name
SBPOD=$(kubectl get pods -n demo -l app=sysbench -o jsonpath='{.items[0].metadata.name}')

# Prepare tables (4 tables x 50k rows)
kubectl exec -n demo $SBPOD -- sysbench oltp_read_write \
  --mysql-host=md --mysql-port=3306 \
  --mysql-user=root --mysql-password="$PASS" \
  --mysql-db=sbtest --tables=4 --table-size=50000 \
  --threads=4 prepare
```

### Run sysbench During Chaos

```bash
kubectl exec -n demo $SBPOD -- sysbench oltp_read_write \
  --mysql-host=md --mysql-port=3306 \
  --mysql-user=root --mysql-password="$PASS" \
  --mysql-db=sbtest --tables=4 --table-size=50000 \
  --threads=4 --time=60 --report-interval=10 run
```

## Galera Cluster Key Concepts

Unlike MySQL Group Replication which has a single primary and secondaries, MariaDB Galera Cluster is **multi-master** — all nodes accept reads and writes simultaneously. Key status variables:

| Variable | Meaning |
|---|---|
| `wsrep_cluster_size` | Number of nodes in the cluster |
| `wsrep_cluster_status` | `Primary` = cluster has quorum and is operational |
| `wsrep_local_state_comment` | `Synced` / `Joined` / `Donor` / `Desynced` |
| `wsrep_ready` | `ON` = node accepts queries |
| `wsrep_connected` | `ON` = node connected to cluster |
| `wsrep_flow_control_paused` | Fraction of time paused for flow control (0.0 = healthy) |

> **Important Notes on Database Status:**
> - **`Ready`** — Database is fully operational. All pods are Synced.
> - **`Critical`** — Cluster has quorum but one or more nodes may be down or desynced.
> - **`NotReady`** — Cluster has lost quorum. No writes can be accepted.
>
> You can read/write in your database in both **`Ready`** and **`Critical`** states. Even if your db is in `Critical` state, your uptime is not compromised.

## Chaos Testing

We will run chaos experiments to see how our Galera cluster behaves under failure scenarios. We use sysbench to simulate high read-write load during each experiment.

### Verify Cluster is Ready

Before starting chaos experiments, verify the cluster is healthy:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                       VERSION   STATUS   AGE
mariadb.kubedb.com/md      11.8.5    Ready    68m

NAME                                 READY   STATUS    RESTARTS      AGE    ROLE
pod/md-0                             2/2     Running   1 (34m ago)   68m    Primary
pod/md-1                             2/2     Running   0             12m    Primary
pod/md-2                             2/2     Running   0             68m    Primary
```

Note: In Galera Cluster, **all nodes have role `Primary`** because every node accepts reads and writes.

Inspect the Galera cluster status:

```shell
➤ kubectl exec -n demo md-0 -c mariadb -- \
    mariadb -uroot -p"$PASS" -e "SHOW GLOBAL STATUS WHERE Variable_name IN (
      'wsrep_cluster_size','wsrep_cluster_status',
      'wsrep_local_state_comment','wsrep_ready',
      'wsrep_connected','wsrep_flow_control_paused');"
Variable_name              Value
wsrep_flow_control_paused  0
wsrep_local_state_comment  Synced
wsrep_cluster_size         3
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON
```

All 3 nodes Synced, cluster_size=3, wsrep_ready=ON. With the cluster ready and sysbench tables prepared, we are ready to run chaos experiments.

---

### Chaos#1: Kill a Pod

We kill one MariaDB pod and see how fast the Galera cluster recovers. In Galera, since all nodes are equal, killing any node should be handled gracefully.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: mariadb-primary-pod-kill
  namespace: chaos-mesh
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  gracePeriod: 0
```

**What this chaos does:** Terminates one MariaDB pod abruptly with `grace-period=0`, forcing the remaining 2 nodes to handle all traffic while the killed pod recovers.

Before running:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                       VERSION   STATUS   AGE
mariadb.kubedb.com/md      11.8.5    Ready    68m

NAME       READY   STATUS    RESTARTS      AGE   ROLE
pod/md-0   2/2     Running   1 (34m ago)   68m   Primary
pod/md-1   2/2     Running   0             12m   Primary
pod/md-2   2/2     Running   0             68m   Primary
```

Apply the chaos:

```shell
➤ kubectl apply -f 1-single-experiments/pod-kill-primary.yaml
podchaos.chaos-mesh.org/mariadb-primary-pod-kill created
```

Within seconds, one pod is killed and recreated. The database goes `Critical` briefly:

```shell
➤ kubectl get mariadb,pods -n demo
NAME                       VERSION   STATUS   AGE
mariadb.kubedb.com/md      11.8.5    Ready    68m

NAME       READY   STATUS    RESTARTS   AGE
pod/md-0   2/2     Running   0          8s
pod/md-1   2/2     Running   0          12m
pod/md-2   2/2     Running   0          68m
```

md-0 was killed and recreated (age=8s). After about 5 seconds, the pod rejoins the Galera cluster and syncs via IST (Incremental State Transfer). Check Galera status:

```shell
➤ SHOW GLOBAL STATUS WHERE Variable_name IN ('wsrep_cluster_size','wsrep_cluster_status',
    'wsrep_local_state_comment','wsrep_ready','wsrep_connected');

md-0:
Variable_name              Value
wsrep_local_state_comment  Synced
wsrep_cluster_size         3
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON

md-1:
Variable_name              Value
wsrep_local_state_comment  Synced
wsrep_cluster_size         3
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON

md-2:
Variable_name              Value
wsrep_local_state_comment  Synced
wsrep_cluster_size         3
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON
```

All 3 nodes Synced. Run sysbench to verify the cluster is fully operational:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1063.05 qps: 21272.82 (r/w/o: 14891.32/2327.87/4053.63) lat (ms,95%): 5.00 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1059.67 qps: 21198.98 (r/w/o: 14840.36/2343.15/4015.46) lat (ms,95%): 4.91 err/s: 0.20 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1060.19 qps: 21202.99 (r/w/o: 14841.86/2407.58/3953.56) lat (ms,95%): 4.91 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        15919  (1060.99 per sec.)
    queries:                             318397 (21220.87 per sec.)
    ignored errors:                      1      (0.07 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

Verify data integrity:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=2941988609, sbtest2=1454430013, sbtest3=496174579, sbtest4=1322761405
md-1: sbtest1=2941988609, sbtest2=1454430013, sbtest3=496174579, sbtest4=1322761405
md-2: sbtest1=2941988609, sbtest2=1454430013, sbtest3=496174579, sbtest4=1322761405
```

**Result: PASS** — Zero data loss. Pod recreated in ~5 seconds, auto-rejoined via IST. All 25 tracking rows preserved, checksums match across all 3 nodes. Post-recovery throughput: 1061 TPS (back to baseline).

Clean up:

```shell
➤ kubectl delete -f 1-single-experiments/pod-kill-primary.yaml
podchaos.chaos-mesh.org "mariadb-primary-pod-kill" deleted
```

---

### Chaos#2: OOMKill (Memory Stress)

We stress-test memory on one node to see if the cluster survives under extreme memory pressure.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: mariadb-primary-memory-stress
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  stressors:
    memory:
      workers: 2
      size: "1200MB"
  duration: "10m"
```

**What this chaos does:** Allocates 1200MB of extra memory on one pod. With MariaDB's memory usage, this approaches the 1.5Gi limit.

Apply the chaos:

```shell
➤ kubectl apply -f 1-single-experiments/stress-memory-primary.yaml
stresschaos.chaos-mesh.org/mariadb-primary-memory-stress created
```

After 20 seconds, check pods — no OOMKill triggered:

```shell
➤ kubectl get mariadb,pods -n demo
NAME                       VERSION   STATUS   AGE
mariadb.kubedb.com/md      11.8.5    Ready    76m

NAME       READY   STATUS    RESTARTS   AGE
pod/md-0   2/2     Running   0          7m54s
pod/md-1   2/2     Running   0          20m
pod/md-2   2/2     Running   0          76m
```

MariaDB survived at 1200MB stress — no OOMKill. Run sysbench during stress:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1051.05 qps: 21032.31 (r/w/o: 14723.04/2446.84/3862.43) lat (ms,95%): 4.91 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1046.48 qps: 20930.60 (r/w/o: 14652.32/2489.19/3789.09) lat (ms,95%): 5.00 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1052.99 qps: 21064.27 (r/w/o: 14745.71/2533.78/3784.78) lat (ms,95%): 4.91 err/s: 0.20 reconn/s: 0.00

SQL statistics:
    transactions:                        15757  (1050.21 per sec.)
    queries:                             315156 (21005.21 per sec.)
    ignored errors:                      1      (0.07 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

Verify data integrity:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=3400554968, sbtest2=1909458598
md-1: sbtest1=3400554968, sbtest2=1909458598
md-2: sbtest1=3400554968, sbtest2=1909458598
```

**Result: PASS** — MariaDB survived 1200MB memory stress without OOMKill. Cluster fully operational at 1050 TPS (no degradation). All 25 tracking rows preserved, checksums match.

Clean up:

```shell
➤ kubectl delete -f 1-single-experiments/stress-memory-primary.yaml
stresschaos.chaos-mesh.org "mariadb-primary-memory-stress" deleted
```

---

### Chaos#3: Network Partition

We isolate one Galera node from the other two for 2 minutes. This tests whether the remaining nodes maintain quorum and continue serving traffic, and whether the isolated node rejoins cleanly.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mariadb-primary-network-partition
  namespace: chaos-mesh
spec:
  action: partition
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "kubedb.com/role": "Primary"
  direction: both
  duration: "2m"
```

**What this chaos does:** Creates a complete network partition between one node and the rest of the cluster for 2 minutes. The isolated node loses quorum and becomes `non-Primary`. The remaining 2 nodes maintain quorum and continue accepting writes.

Before running:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS   AGE
mariadb.kubedb.com/md   11.8.5    Ready    82m

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          13m   Primary
pod/md-1   2/2     Running   0          25m   Primary
pod/md-2   2/2     Running   0          82m   Primary

➤ SHOW GLOBAL STATUS ...
wsrep_flow_control_paused  0.0277989
wsrep_local_state_comment  Synced
wsrep_cluster_size         3
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON
```

Apply the chaos:

```shell
➤ kubectl apply -f 1-single-experiments/network-partition-primary.yaml
networkchaos.chaos-mesh.org/mariadb-primary-network-partition created
```

Within ~15 seconds, the isolated node loses quorum. The database goes `Critical`:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS     AGE
mariadb.kubedb.com/md   11.8.5    Critical   82m

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          14m   Primary
pod/md-1   2/2     Running   0          26m   Primary
pod/md-2   2/2     Running   0          82m   non-Primary
```

Note: md-2 is `non-Primary` — it's the isolated node. Let's check Galera status from each node:

```shell
➤ # md-0 (in quorum):
wsrep_flow_control_paused  0.0257333
wsrep_local_state_comment  Synced
wsrep_cluster_size         2
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON

➤ # md-1 (in quorum):
wsrep_flow_control_paused  0.0209364
wsrep_local_state_comment  Synced
wsrep_cluster_size         2
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON

➤ # md-2 (ISOLATED):
wsrep_flow_control_paused  0.00670716
wsrep_local_state_comment  Initialized
wsrep_cluster_size         1
wsrep_cluster_status       non-Primary
wsrep_connected            ON
wsrep_ready                OFF
```

The isolated node (md-2) shows `wsrep_cluster_size=1`, `wsrep_cluster_status=non-Primary`, `wsrep_ready=OFF` — it cannot accept queries. The remaining 2 nodes still have quorum (`wsrep_cluster_status=Primary`) and accept both reads and writes.

Run sysbench during the partition:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1435.24 qps: 28715.89 (r/w/o: 20102.42/3503.22/5110.25) lat (ms,95%): 3.49 err/s: 0.20 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1437.64 qps: 28756.01 (r/w/o: 20129.37/3626.70/4999.94) lat (ms,95%): 3.43 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1416.81 qps: 28336.92 (r/w/o: 19836.08/3607.82/4893.02) lat (ms,95%): 3.55 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        21453  (1429.89 per sec.)
    queries:                             429076 (28598.90 per sec.)
    ignored errors:                      1      (0.07 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

TPS **increased from 1039 to 1430** during partition — a 37% improvement! This is because with only 2 nodes, Galera certification has less overhead (fewer nodes to coordinate with).

After the 2-minute partition expires, the isolated node automatically rejoins:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS   AGE
mariadb.kubedb.com/md   11.8.5    Ready    85m

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          17m   Primary
pod/md-1   2/2     Running   0          29m   Primary
pod/md-2   2/2     Running   0          85m   Primary
```

All 3 nodes back to `Primary`, cluster `Ready`. Verify:

```shell
➤ # Galera status — all Synced
md-0: wsrep_cluster_size=3, Synced, wsrep_ready=ON, wsrep_flow_control_paused=0.0208
md-1: wsrep_cluster_size=3, Synced, wsrep_ready=ON, wsrep_flow_control_paused=0.0186
md-2: wsrep_cluster_size=3, Synced, wsrep_ready=ON, wsrep_flow_control_paused=0.0064

➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=3192592587, sbtest2=1620218475, sbtest3=827673677, sbtest4=2199205073
md-1: sbtest1=3192592587, sbtest2=1620218475, sbtest3=827673677, sbtest4=2199205073
md-2: sbtest1=3192592587, sbtest2=1620218475, sbtest3=827673677, sbtest4=2199205073
```

**Result: PASS** — Network partition handled correctly. Isolated node became `non-Primary` and stopped accepting queries (no split-brain). Remaining 2 nodes maintained quorum at 1430 TPS. After partition expired, isolated node auto-rejoined and synced. Zero data loss — all 25 tracking rows preserved, checksums match.

Clean up:

```shell
➤ kubectl delete -f 1-single-experiments/network-partition-primary.yaml
networkchaos.chaos-mesh.org "mariadb-primary-network-partition" deleted
```

---

### Chaos#4: IO Latency (100ms)

We inject 100ms latency on all disk operations on one node. This simulates degraded storage — a common issue with cloud block storage or overloaded disk controllers.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: mariadb-primary-io-latency
  namespace: chaos-mesh
spec:
  action: latency
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  volumePath: "/var/lib/mysql"
  path: "/**"
  delay: "100ms"
  percent: 100
  duration: "3m"
```

**What this chaos does:** Adds 100ms delay to every disk read/write on `/var/lib/mysql` for one node. This makes every InnoDB flush, WAL write, and data page read significantly slower.

Apply the chaos:

```shell
➤ kubectl apply -f 1-single-experiments/io-latency-primary.yaml
iochaos.chaos-mesh.org/mariadb-primary-io-latency created
```

After ~10 seconds, the affected node (md-1) becomes completely unresponsive — MariaDB cannot handle 100ms latency on every disk operation:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS   AGE
mariadb.kubedb.com/md   11.8.5    Ready    88m

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          19m   Primary
pod/md-1   2/2     Running   0          32m   Primary
pod/md-2   2/2     Running   0          88m   Primary
```

The pod shows Running/Ready (Kubernetes doesn't know MariaDB inside is frozen), but the Galera cluster has already expelled the node:

```shell
➤ # md-0 (healthy):
wsrep_flow_control_paused  0.0189172
wsrep_local_state_comment  Synced
wsrep_cluster_size         2
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON

➤ # md-1 (IO latency target):
ERROR 2002 (HY000): Can't connect to local server through socket '/run/mysqld/mysqld.sock' (111)

➤ # md-2 (healthy):
wsrep_flow_control_paused  0.00631477
wsrep_local_state_comment  Synced
wsrep_cluster_size         2
wsrep_cluster_status       Primary
wsrep_connected            ON
wsrep_ready                ON
```

The affected node is completely unreachable — `wsrep_cluster_size=2` on the healthy nodes confirms Galera expelled it. Run sysbench against the 2 remaining nodes:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1452.19 qps: 29057.21 (r/w/o: 20341.06/3789.65/4926.49) lat (ms,95%): 3.36 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1456.70 qps: 29129.26 (r/w/o: 20390.44/3873.07/4865.74) lat (ms,95%): 3.30 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1440.25 qps: 28803.66 (r/w/o: 20162.34/3885.60/4755.71) lat (ms,95%): 3.43 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        21750  (1449.63 per sec.)
    queries:                             435000 (28992.63 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

1450 TPS on 2 nodes, zero errors. After the 3-minute IO latency expires, the affected node auto-recovers and rejoins:

```shell
➤ kubectl get mariadb -n demo md
NAME   VERSION   STATUS   AGE
md     11.8.5    Ready    94m

➤ # All nodes recovered:
md-0: wsrep_cluster_size=3, Synced, wsrep_ready=ON, wsrep_flow_control_paused=0.0145
md-1: wsrep_cluster_size=3, Synced, wsrep_ready=ON, wsrep_flow_control_paused=0
md-2: wsrep_cluster_size=3, Synced, wsrep_ready=ON, wsrep_flow_control_paused=0.0059
```

Verify data integrity:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=3774570434, sbtest2=107955818, sbtest3=4222587866, sbtest4=2113239503
md-1: sbtest1=3774570434, sbtest2=107955818, sbtest3=4222587866, sbtest4=2113239503
md-2: sbtest1=3774570434, sbtest2=107955818, sbtest3=4222587866, sbtest4=2113239503
```

**Result: PASS** — IO latency caused the affected node to become unresponsive and expelled from Galera. The remaining 2 nodes continued serving 1450 TPS with zero errors. After chaos expired, the affected node auto-rejoined via IST. Zero data loss — all checksums match.

Clean up:

```shell
➤ kubectl delete -f 1-single-experiments/io-latency-primary.yaml
iochaos.chaos-mesh.org "mariadb-primary-io-latency" deleted
```

---

### Chaos#5: Network Latency (1s)

We inject 1 second network latency between one node and all others. This tests how Galera certification handles slow cross-node communication — a common scenario with cross-region deployments or degraded network links.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mariadb-replication-latency
  namespace: chaos-mesh
spec:
  action: delay
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "md"
        "kubedb.com/role": "Primary"
  delay:
    latency: "1s"
    jitter: "50ms"
  duration: "10m"
  direction: both
```

**What this chaos does:** Adds 1 second latency (+ 50ms jitter) between one node and all other cluster nodes. Every Galera certification message must wait 1 second each way.

Apply the chaos:

```shell
➤ kubectl apply -f 1-single-experiments/network-latency-primary-to-replicas.yaml
networkchaos.chaos-mesh.org/mariadb-replication-latency created
```

Check Galera — all 3 nodes stay Synced (1s latency isn't enough to trigger expulsion):

```shell
➤ # All nodes:
md-0: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0136
md-1: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0
md-2: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0058
```

But sysbench throughput is **severely impacted**:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 2.40 qps: 63.19 (r/w/o: 44.79/8.40/10.00) lat (ms,95%): 1938.16 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 2.80 qps: 56.00 (r/w/o: 39.20/7.80/9.00) lat (ms,95%): 1869.60 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 3.00 qps: 60.00 (r/w/o: 42.00/7.00/11.00) lat (ms,95%): 1973.38 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        45     (2.77 per sec.)
    queries:                             900    (55.49 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)

Latency (ms):
         min:                                  985.37
         avg:                                 1418.98
         max:                                 2090.38
         95th percentile:                     2045.74
```

**TPS dropped from 1039 to 2.77** — a 99.7% reduction! This is because Galera uses synchronous replication: every write must be certified by all nodes. With 1s network latency, each certification round-trip takes ~2 seconds. The 95th percentile latency is 2045ms, confirming this.

However, note the key metrics: **0 errors, 0 reconnects**. The cluster never broke — it just became extremely slow.

After removing the chaos, throughput returns to normal immediately. Verify data:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=620696091, sbtest2=702654982, sbtest3=2895801545, sbtest4=86505438
md-1: sbtest1=620696091, sbtest2=702654982, sbtest3=2895801545, sbtest4=86505438
md-2: sbtest1=620696091, sbtest2=702654982, sbtest3=2895801545, sbtest4=86505438
```

**Result: PASS** — 1s network latency caused severe throughput degradation (99.7% TPS drop) due to Galera's synchronous certification, but the cluster remained operational with zero errors. No split-brain, no data loss. This is a key trade-off of synchronous replication: latency directly impacts throughput, but data safety is guaranteed.

Clean up:

```shell
➤ kubectl delete -f 1-single-experiments/network-latency-primary-to-replicas.yaml
networkchaos.chaos-mesh.org "mariadb-replication-latency" deleted
```

---

### Chaos#6: CPU Stress (98%)

We apply 98% CPU stress on one node to see if the Galera cluster can handle a CPU-starved member.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: mariadb-primary-cpu-stress
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  stressors:
    cpu:
      workers: 2
      load: 98
  duration: "5m"
```

**What this chaos does:** Consumes 98% CPU on one node using 2 stress workers. Tests whether the cluster maintains throughput and data consistency under extreme CPU pressure.

Apply the chaos:

```shell
➤ kubectl apply -f 1-single-experiments/stress-cpu-primary.yaml
stresschaos.chaos-mesh.org/mariadb-primary-cpu-stress created
```

All 3 nodes remain Synced:

```shell
➤ # Galera status during CPU stress:
md-0: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0121
md-1: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0
md-2: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0056
```

Run sysbench:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1033.26 qps: 20679.04 (r/w/o: 14476.27/2825.42/3377.35) lat (ms,95%): 5.18 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1026.65 qps: 20530.52 (r/w/o: 14371.44/2828.13/3330.95) lat (ms,95%): 5.09 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1041.79 qps: 20836.99 (r/w/o: 14586.25/2930.17/3320.57) lat (ms,95%): 5.00 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        15513  (1033.96 per sec.)
    queries:                             310260 (20679.19 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

**1034 TPS — virtually no impact!** MariaDB on Galera is largely IO-bound, not CPU-bound for write workloads.

Verify data:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=3417352532, sbtest2=403445326
md-1: sbtest1=3417352532, sbtest2=403445326
md-2: sbtest1=3417352532, sbtest2=403445326
```

**Result: PASS** — 98% CPU stress had negligible impact on Galera cluster. 1034 TPS (99.5% of baseline). Zero errors, all nodes Synced throughout. MariaDB's write path is IO-bound, not CPU-bound.

Clean up:

```shell
➤ kubectl delete -f 1-single-experiments/stress-cpu-primary.yaml
stresschaos.chaos-mesh.org "mariadb-primary-cpu-stress" deleted
```

---

### Chaos#7: Packet Loss (30%)

We inject 30% packet loss across all cluster nodes. This simulates unreliable network infrastructure — congested switches, flaky NICs, or cloud provider network issues.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mariadb-cluster-packet-loss
  namespace: chaos-mesh
spec:
  action: loss
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
  loss:
    loss: "30"
    correlation: "25"
  duration: "5m"
```

**What this chaos does:** Drops 30% of all network packets on every cluster node with 25% correlation (burst losses). This affects both Galera replication traffic and client connections.

Apply the chaos:

```shell
➤ kubectl apply -f 1-single-experiments/packet-loss.yaml
networkchaos.chaos-mesh.org/mariadb-cluster-packet-loss created
```

All 3 nodes remain Synced — Galera handles retransmissions:

```shell
➤ # Galera status during 30% packet loss:
md-0: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0124
md-1: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0077
md-2: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0064
```

But throughput is severely impacted due to TCP retransmissions and Galera certification delays:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1.00 qps: 28.99 (r/w/o: 21.80/3.20/4.00) lat (ms,95%): 4203.93 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1.60 qps: 33.40 (r/w/o: 23.40/4.40/5.60) lat (ms,95%): 5507.54 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 0.80 qps: 20.80 (r/w/o: 13.60/4.00/3.20) lat (ms,95%): 3326.55 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        21     (1.32 per sec.)
    queries:                             420    (26.40 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)

Latency (ms):
         min:                                 1373.02
         avg:                                 2987.19
         max:                                 5524.30
         95th percentile:                     4517.90
```

**TPS dropped from 1039 to 1.32** — even worse than 1s network latency! 30% packet loss causes massive TCP retransmissions, and each Galera certification round can take several seconds. But critically: **zero errors, zero reconnects, no node expulsion**.

Verify data integrity after cleanup:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=2426416412, sbtest2=1176558970, sbtest3=4292186888, sbtest4=1305861152
md-1: sbtest1=2426416412, sbtest2=1176558970, sbtest3=4292186888, sbtest4=1305861152
md-2: sbtest1=2426416412, sbtest2=1176558970, sbtest3=4292186888, sbtest4=1305861152
```

**Result: PASS** — 30% packet loss caused extreme throughput degradation (TPS: 1039 → 1.32) but no node was expelled and no data was lost. Galera's TCP-based replication handles retransmissions correctly. All checksums match.

Clean up:

```shell
➤ kubectl delete -f 1-single-experiments/packet-loss.yaml
networkchaos.chaos-mesh.org "mariadb-cluster-packet-loss" deleted
```

---

### Chaos#8: Full Cluster Kill

We kill ALL 3 pods simultaneously — the worst-case scenario. This tests Galera's ability to bootstrap from a complete outage with no surviving member.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: mariadb-full-cluster-kill
  namespace: chaos-mesh
spec:
  action: pod-kill
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
  gracePeriod: 0
```

**What this chaos does:** Kills all 3 MariaDB pods simultaneously. No surviving node means the cluster must bootstrap from scratch using the data on PVCs.

Apply the chaos:

```shell
➤ kubectl apply -f full-cluster-kill.yaml
podchaos.chaos-mesh.org/mariadb-full-cluster-kill created
```

All pods are killed and recreated. The database immediately goes `NotReady`:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS     AGE
mariadb.kubedb.com/md   11.8.5    NotReady   112m

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          10s   Unknown
pod/md-1   2/2     Running   0          10s   Unknown
pod/md-2   2/2     Running   0          10s   Unknown
```

All pods show role `Unknown` — the Galera cluster has no quorum, no primary component. The KubeDB coordinator detects this and initiates a Galera bootstrap sequence:

1. The coordinator identifies the node with the most recent GTID (highest `seqno`)
2. That node is bootstrapped as the new primary component (`--wsrep-new-cluster`)
3. The other 2 nodes join the bootstrapped cluster via IST/SST

After approximately **3 minutes**, all nodes rejoin and the cluster becomes `Ready`:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS   AGE
mariadb.kubedb.com/md   11.8.5    Ready    115m

NAME       READY   STATUS    RESTARTS   AGE     ROLE
pod/md-0   2/2     Running   0          2m59s   Primary
pod/md-1   2/2     Running   0          2m59s   Primary
pod/md-2   2/2     Running   0          2m59s   Primary
```

All 3 nodes back to `Primary`, Galera fully operational:

```shell
➤ # All nodes:
md-0: wsrep_cluster_size=3, Synced, wsrep_ready=ON
md-1: wsrep_cluster_size=3, Synced, wsrep_ready=ON
md-2: wsrep_cluster_size=3, Synced, wsrep_ready=ON
```

Run sysbench to confirm throughput is restored:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1014.09 qps: 20295.00 lat (ms,95%): 5.47 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1027.82 qps: 20559.23 lat (ms,95%): 5.28 err/s: 0.20 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1030.61 qps: 20616.72 lat (ms,95%): 5.18 err/s: 0.20 reconn/s: 0.00

SQL statistics:
    transactions:                        15367  (1024.26 per sec.)
    queries:                             307372 (20487.38 per sec.)
    ignored errors:                      2      (0.13 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

Verify data integrity:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=3947591857, sbtest2=3914957582, sbtest3=901119555, sbtest4=4023705335
md-1: sbtest1=3947591857, sbtest2=3914957582, sbtest3=901119555, sbtest4=4023705335
md-2: sbtest1=3947591857, sbtest2=3914957582, sbtest3=901119555, sbtest4=4023705335
```

**Result: PASS** — Full cluster kill (all 3 pods) recovered automatically via Galera bootstrap in ~3 minutes. KubeDB coordinator handled the bootstrap sequence — no manual intervention needed. 1024 TPS post-recovery. Zero data loss — all 25 tracking rows preserved, all checksums match.

Clean up:

```shell
➤ kubectl delete podchaos mariadb-full-cluster-kill -n chaos-mesh
podchaos.chaos-mesh.org "mariadb-full-cluster-kill" deleted
```

---

### Chaos#9: DNS Error

We inject DNS errors on one cluster node to see if it affects Galera replication.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: DNSChaos
metadata:
  name: mariadb-dns-error-primary
  namespace: chaos-mesh
spec:
  action: error
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  duration: "3m"
```

**What this chaos does:** Makes all DNS lookups fail on one MariaDB pod. Tests whether Galera replication (which uses IP addresses internally) is affected by DNS failures.

Apply and check:

```shell
➤ kubectl apply -f 1-single-experiments/dns-error-primary.yaml
dnschaos.chaos-mesh.org/mariadb-dns-error-primary created
```

All 3 nodes remain Synced — Galera uses IP addresses for cluster communication, not DNS:

```shell
➤ # Galera status during DNS error:
md-0: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0193
md-1: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0193
md-2: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0182

➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1011.69 qps: 20251.63 lat (ms,95%): 5.47 err/s: 0.20 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1022.82 qps: 20449.23 lat (ms,95%): 5.09 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1013.21 qps: 20267.98 lat (ms,95%): 5.47 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        15242  (1015.94 per sec.)
    queries:                             304857 (20319.98 per sec.)
    ignored errors:                      1      (0.07 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

1016 TPS — virtually no impact. Verify:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=286824257, sbtest2=4261772161
md-1: sbtest1=286824257, sbtest2=4261772161
md-2: sbtest1=286824257, sbtest2=4261772161
```

**Result: PASS** — DNS errors had no impact on Galera replication. Galera communicates via IP addresses, so DNS failures don't affect cluster operations. 1016 TPS (97.8% of baseline).

---

### Chaos#10: IO Fault (EIO 50%)

We inject IO errors (errno 5 = EIO) on 50% of disk operations on one node. This simulates disk corruption or failing storage hardware.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: mariadb-primary-io-fault
  namespace: chaos-mesh
spec:
  action: fault
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  volumePath: "/var/lib/mysql"
  path: "/**"
  errno: 5
  percent: 50
  duration: "3m"
```

**What this chaos does:** Returns `EIO` (Input/output error) for 50% of all disk operations on `/var/lib/mysql`. This is more severe than IO latency — it causes actual data access failures.

Apply the chaos:

```shell
➤ kubectl apply -f 1-single-experiments/io-fault-primary.yaml
iochaos.chaos-mesh.org/mariadb-primary-io-fault created
```

The affected node (md-0) crashes with a segfault — MariaDB cannot handle random IO errors on its data files:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS     AGE
mariadb.kubedb.com/md   11.8.5    Critical   129m

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          17m   Unknown
pod/md-1   2/2     Running   0          17m   Primary
pod/md-2   2/2     Running   0          17m   Primary

➤ # md-0 (crashed):
ERROR 2002 (HY000): Can't connect to local server through socket '/run/mysqld/mysqld.sock' (111)

➤ # md-1 (healthy):
wsrep_flow_control_paused  0.0222942
wsrep_local_state_comment  Synced
wsrep_cluster_size         2
wsrep_cluster_status       Primary
wsrep_ready                ON

➤ # md-2 (healthy):
wsrep_flow_control_paused  0.0215375
wsrep_local_state_comment  Synced
wsrep_cluster_size         2
wsrep_cluster_status       Primary
wsrep_ready                ON
```

MariaDB logs show: `Segmentation fault (core dumped)`. The remaining 2 nodes continue serving traffic:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1391.02 qps: 27830.18 lat (ms,95%): 3.55 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1409.46 qps: 28191.89 lat (ms,95%): 3.55 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1411.61 qps: 28231.51 lat (ms,95%): 3.55 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        21065  (1404.01 per sec.)
    queries:                             421300 (28080.23 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

1404 TPS on 2 nodes. After chaos expires, the KubeDB coordinator detects the crashed node, restarts it, and it rejoins via IST:

```shell
➤ # After recovery:
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=3560624458, sbtest2=1590320566
md-1: sbtest1=3560624458, sbtest2=1590320566
md-2: sbtest1=3560624458, sbtest2=1590320566
```

**Result: PASS** — IO faults caused MariaDB to segfault on the affected node, but the remaining 2 nodes continued at 1404 TPS. After chaos expired, the crashed node was recovered by the coordinator. Zero data loss.

---

### Chaos#11: Clock Skew (-5 min)

We shift the clock backward by 5 minutes on one node. This tests whether time-dependent operations (certificates, timeouts, GTID ordering) are affected.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: TimeChaos
metadata:
  name: mariadb-primary-clock-skew
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  timeOffset: "-5m"
  duration: "3m"
```

**What this chaos does:** Shifts the system clock backward by 5 minutes on one pod. This can confuse time-dependent operations like TLS certificate validation, timeout calculations, and log ordering.

Apply and check:

```shell
➤ kubectl apply -f 1-single-experiments/clock-skew-primary.yaml
timechaos.chaos-mesh.org/mariadb-primary-clock-skew created
```

All 3 nodes remain Synced:

```shell
➤ # Galera status during clock skew:
md-0: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0
md-1: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0165
md-2: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0161

➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 973.31 qps: 19488.47 lat (ms,95%): 6.09 err/s: 0.40 reconn/s: 0.00
[ 10s ] thds: 4 tps: 995.79 qps: 19915.27 lat (ms,95%): 5.88 err/s: 0.20 reconn/s: 0.00
[ 15s ] thds: 4 tps: 996.42 qps: 19930.23 lat (ms,95%): 5.47 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        14832  (988.48 per sec.)
    queries:                             296691 (19772.96 per sec.)
    ignored errors:                      3      (0.20 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

988 TPS — minor 5% drop. 3 ignored errors (likely deadlocks from skewed timestamps). Verify:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=1081956110, sbtest2=4193306675
md-1: sbtest1=1081956110, sbtest2=4193306675
md-2: sbtest1=1081956110, sbtest2=4193306675
```

**Result: PASS** — 5-minute clock skew caused only a minor throughput dip (5%) and 3 ignored errors. Galera certification is based on write-set ordering, not wall-clock time, so clock skew has minimal impact.

---

### Chaos#12: Bandwidth Throttle (1 mbps)

We throttle network bandwidth to 1 mbps on one node. This simulates a severely constrained network link.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mariadb-bandwidth-throttle
  namespace: chaos-mesh
spec:
  action: bandwidth
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  bandwidth:
    rate: "1mbps"
    limit: 20971520
    buffer: 10000
  duration: "3m"
```

**What this chaos does:** Limits the outbound bandwidth of one node to 1 mbps. This severely restricts the amount of data the node can send for Galera replication and client responses.

Apply and check:

```shell
➤ kubectl apply -f 1-single-experiments/bandwidth-throttle.yaml
networkchaos.chaos-mesh.org/mariadb-bandwidth-throttle created
```

All 3 nodes remain Synced, but flow control kicks in:

```shell
➤ # Galera status during bandwidth throttle:
md-0: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0247
md-1: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0202
md-2: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0197

➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 282.18 qps: 5651.74 (r/w/o: 3957.88/903.93/789.94) lat (ms,95%): 41.85 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 278.00 qps: 5565.80 (r/w/o: 3895.40/885.00/785.40) lat (ms,95%): 41.85 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 281.40 qps: 5627.60 (r/w/o: 3939.20/898.00/790.40) lat (ms,95%): 41.85 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        4212   (280.27 per sec.)
    queries:                             84240  (5605.35 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)

Latency (ms):
         min:                                    2.80
         avg:                                   14.25
         max:                                   50.38
         95th percentile:                       41.85
```

**TPS dropped from 1039 to 280** (73% reduction). The 1 mbps bandwidth limit constrains how fast write-sets can be replicated. P95 latency increased from 5ms to 42ms. But zero errors — the cluster self-throttles via Galera flow control.

Note the `wsrep_flow_control_paused` values are slightly elevated (0.02) — Galera is pausing writers to prevent the slow node from falling too far behind.

Verify data integrity after cleanup:

```shell
➤ # Tracking rows — all match
md-0 markers: 25
md-1 markers: 25
md-2 markers: 25

➤ # Checksums — all match
md-0: sbtest1=410521570, sbtest2=126349543, sbtest3=2842752298, sbtest4=2843785900
md-1: sbtest1=410521570, sbtest2=126349543, sbtest3=2842752298, sbtest4=2843785900
md-2: sbtest1=410521570, sbtest2=126349543, sbtest3=2842752298, sbtest4=2843785900
```

**Result: PASS** — Bandwidth throttle caused significant throughput degradation (73%) but Galera flow control kept all nodes Synced with zero errors. No data loss.

---

### Chaos#13: Pod Failure (5 min pause)

Unlike pod-kill which deletes and recreates the pod, pod-failure **pauses** the pod — making it completely unresponsive while the container appears to be Running.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: mariadb-primary-pod-failure
  namespace: chaos-mesh
spec:
  action: pod-failure
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  duration: "5m"
```

**What this chaos does:** Freezes one pod completely — the container appears Running but is unresponsive. This simulates a hung process or kernel-level freeze.

During chaos, the frozen pod (md-1) shows `container not found` when exec'd:

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS     AGE
mariadb.kubedb.com/md   11.8.5    Critical   149m

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          37m   Primary
pod/md-1   2/2     Running   0          37m   Unknown
pod/md-2   2/2     Running   0          37m   Primary

➤ # md-0: wsrep_cluster_size=2, Synced, wsrep_ready=ON
➤ # md-1: error: container not found ("mariadb")
➤ # md-2: wsrep_cluster_size=2, Synced, wsrep_ready=ON
```

The remaining 2 nodes continue serving traffic:

```shell
➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1415.21 qps: 28314.94 lat (ms,95%): 3.49 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1420.69 qps: 28413.41 lat (ms,95%): 3.43 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1390.60 qps: 27812.97 lat (ms,95%): 3.68 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        21137  (1408.83 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

**Result: PASS** — Pod freeze handled gracefully. 1409 TPS on 2 nodes, 0 errors. After chaos expired, frozen pod restarted and rejoined. 25/25 markers, checksums match.

---

### Chaos#14: Container Kill (mariadb process only)

We kill only the mariadb container, not the entire pod. This tests whether the pod-level restart (by kubelet) and the coordinator can recover the database process.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: mariadb-kill-mariadb-process
  namespace: chaos-mesh
spec:
  action: container-kill
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  containerNames:
    - mariadb
  duration: "30s"
```

**What this chaos does:** Kills only the `mariadb` container inside the pod. The coordinator container keeps running. Kubelet restarts the killed container automatically.

```shell
➤ kubectl get pods -n demo -L kubedb.com/role
NAME   READY   STATUS    RESTARTS      AGE   ROLE
md-0   2/2     Running   0             16h   Primary
md-1   2/2     Running   5 (15s ago)   16h   Unknown
md-2   2/2     Running   0             16h   Primary

➤ # md-0: wsrep_cluster_size=2, Synced, wsrep_flow_control_paused=0.0001
➤ # md-1: ERROR 2002 - Can't connect to local server through socket
➤ # md-2: wsrep_cluster_size=2, Synced, wsrep_flow_control_paused=0.0004

➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1403.41 qps: 28080.43 lat (ms,95%): 3.55 err/s: 0.20 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1368.89 qps: 27381.54 lat (ms,95%): 3.82 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1370.59 qps: 27412.22 lat (ms,95%): 3.68 err/s: 0.20 reconn/s: 0.00

SQL statistics:
    transactions:                        20719  (1380.96 per sec.)
    ignored errors:                      2      (0.13 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

**Result: PASS** — Container kill handled by kubelet restart. 1381 TPS on 2 nodes. Coordinator detected the restart and rejoined the node. 25/25 markers, checksums match.

---

### Chaos#15: Packet Duplicate (50%)

We inject 50% packet duplication. Unlike packet loss, duplicated packets arrive multiple times, which can confuse protocols that aren't idempotent.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mariadb-primary-packet-duplicate
  namespace: chaos-mesh
spec:
  action: duplicate
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "md"
  duplicate:
    duplicate: "50"
    correlation: "25"
  duration: "10m"
  direction: both
```

All 3 nodes remain Synced:

```shell
➤ # Galera status during 50% packet duplication:
md-0: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0001
md-1: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0
md-2: wsrep_cluster_size=3, Synced, wsrep_flow_control_paused=0.0004

➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 988.07 qps: 19773.62 lat (ms,95%): 5.67 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1009.84 qps: 20196.64 lat (ms,95%): 5.57 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 986.39 qps: 19727.30 lat (ms,95%): 5.57 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        14926  (994.79 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

**Result: PASS** — 50% packet duplication caused only a minor 4% TPS drop (995 vs 1039). TCP handles duplicate packets natively (sequence numbers), so Galera is unaffected. Zero errors, all nodes Synced. 25/25 markers, checksums match.

---

### Chaos#16: Packet Corrupt (50%)

We corrupt 50% of all packets. This is the most severe network chaos — corrupt packets fail TCP checksums, causing retransmissions that compound with Galera's certification protocol.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: mariadb-primary-packet-corrupt
  namespace: chaos-mesh
spec:
  action: corrupt
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "md"
  corrupt:
    corrupt: "50"
    correlation: "25"
  duration: "10m"
  direction: both
```

**This is the most severe chaos experiment.** 50% packet corruption completely broke the Galera cluster:

```shell
➤ # ALL 3 nodes lost quorum:
md-0: wsrep_cluster_size=1, non-Primary, wsrep_ready=OFF, Initialized
md-1: wsrep_cluster_size=1, non-Primary, wsrep_ready=OFF, Initialized
md-2: wsrep_cluster_size=1, non-Primary, wsrep_ready=OFF, Initialized

➤ sysbench:
FATAL: mysql_stmt_execute() returned error 1047 (WSREP has not yet prepared node for application use)
```

The cluster was completely non-functional — all nodes became standalone (`cluster_size=1`) with `non-Primary` status. No node could accept writes (`error 1047`). This happened because corrupt packets broke the group communication protocol — nodes couldn't exchange valid heartbeats or certification messages.

After removing the chaos, the KubeDB coordinator detected the complete outage and bootstrapped the cluster:

```shell
➤ kubectl get mariadb -n demo md
NAME   VERSION   STATUS   AGE
md     11.8.5    Ready    18h
```

**Result: PASS** — 50% packet corruption caused a complete cluster outage (all nodes non-Primary). This is the expected worst case — Galera cannot function when half of all packets are corrupted. However, after chaos removal, the coordinator automatically bootstrapped the cluster with zero data loss. 25/25 markers, checksums match.

---

### Chaos#17: IO Attr Override (read-only filesystem)

We override file attributes to make the data directory read-only. This simulates a read-only filesystem mount or permissions issue.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: mariadb-primary-io-attr-override
  namespace: chaos-mesh
spec:
  action: attrOverride
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  volumePath: /var/lib/mysql
  path: /var/lib/mysql/**/*
  attr:
    perm: 444
  percent: 100
  duration: "10m"
  containerNames:
    - mariadb
```

**What this chaos does:** Makes all files in `/var/lib/mysql` read-only (perm 444). MariaDB cannot write WAL, flush pages, or commit transactions.

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS     AGE
mariadb.kubedb.com/md   11.8.5    Critical   18h

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          16h   Primary
pod/md-1   2/2     Running   5          16h   Primary
pod/md-2   2/2     Running   0          16h   Unknown

➤ # md-0: wsrep_cluster_size=2, Synced, wsrep_ready=ON
➤ # md-2 (affected): ERROR 2002 - Can't connect to local server
➤ # md-1: wsrep_cluster_size=2, Synced, wsrep_ready=ON

➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1384.82 qps: 27707.45 lat (ms,95%): 3.62 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1381.28 qps: 27625.71 lat (ms,95%): 3.55 err/s: 0.20 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1397.80 qps: 27962.05 lat (ms,95%): 3.49 err/s: 0.20 reconn/s: 0.00

SQL statistics:
    transactions:                        20824  (1387.98 per sec.)
    ignored errors:                      2      (0.13 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

**Result: PASS** — Read-only filesystem crashed MariaDB on the affected node. Remaining 2 nodes served 1388 TPS. After chaos expired, coordinator recovered the node. 25/25 markers, checksums match.

---

### Chaos#18: IO Mistake (random data corruption in IO)

We inject random bytes into 50% of disk read/write operations. This is the most dangerous IO chaos — it corrupts actual data being read from or written to disk.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: mariadb-primary-io-mistake
  namespace: chaos-mesh
spec:
  action: mistake
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "md"
      "kubedb.com/role": "Primary"
  volumePath: /var/lib/mysql
  path: /var/lib/mysql/**/*
  mistake:
    filling: random
    maxOccurrences: 10
    maxLength: 100
  percent: 50
  duration: "10m"
  containerNames:
    - mariadb
```

**What this chaos does:** Replaces up to 10 occurrences of up to 100 bytes with random data in 50% of IO operations. This corrupts InnoDB pages, WAL entries, and metadata.

```shell
➤ kubectl get mariadb,pods -n demo -L kubedb.com/role
NAME                    VERSION   STATUS     AGE
mariadb.kubedb.com/md   11.8.5    Critical   18h

NAME       READY   STATUS    RESTARTS   AGE   ROLE
pod/md-0   2/2     Running   0          16h   Primary
pod/md-1   2/2     Running   5          16h   Unknown
pod/md-2   2/2     Running   0          16h   Primary

➤ # md-0: wsrep_cluster_size=2, Synced, wsrep_ready=ON
➤ # md-1 (affected): ERROR 2002 - Can't connect
➤ # md-2: wsrep_cluster_size=2, Synced, wsrep_ready=ON

➤ sysbench oltp_read_write ... --time=15 --report-interval=5 run
[ 5s ] thds: 4 tps: 1352.04 qps: 27051.31 lat (ms,95%): 3.96 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 4 tps: 1396.45 qps: 27929.02 lat (ms,95%): 3.55 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 4 tps: 1390.02 qps: 27800.64 lat (ms,95%): 3.55 err/s: 0.00 reconn/s: 0.00

SQL statistics:
    transactions:                        20697  (1379.51 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)
```

**Result: PASS** — Random IO corruption crashed MariaDB on the affected node (data page checksums detected the corruption). Remaining 2 nodes served 1380 TPS with zero errors. The corrupted node was recovered by the coordinator (likely via SST — full state transfer from a healthy node). 25/25 markers, checksums match across all 3 nodes.

---

## Chaos Testing Results Summary

### Test Results Overview

**All 18 experiments passed with zero data loss.**

| # | Experiment | TPS During | Impact | Recovery | Data |
|---|---|---|---|---|---|
| 1 | Pod Kill | 1061 | None | ~5s auto-rejoin | 25/25, checksums MATCH |
| 2 | OOMKill (1200MB) | 1050 | None (survived) | N/A | 25/25, checksums MATCH |
| 3 | Network Partition | 1430 (+37%) | TPS increased (2 nodes) | Auto-rejoin after expiry | 25/25, checksums MATCH |
| 4 | IO Latency (100ms) | 1450 (2 nodes) | Node unresponsive | Auto-rejoin after expiry | 25/25, checksums MATCH |
| 5 | Network Latency (1s) | 3 (-99.7%) | Severe degradation | Instant after removal | 25/25, checksums MATCH |
| 6 | CPU Stress (98%) | 1034 | Negligible | N/A | 25/25, checksums MATCH |
| 7 | Packet Loss (30%) | 1.3 (-99.9%) | Severe degradation | Instant after removal | 25/25, checksums MATCH |
| 8 | Full Cluster Kill | 1024 | Full outage | ~3 min bootstrap | 25/25, checksums MATCH |
| 9 | DNS Error | 1016 | None | N/A | 25/25, checksums MATCH |
| 10 | IO Fault (EIO 50%) | 1404 (2 nodes) | Node crash (segfault) | Coordinator recovery | 25/25, checksums MATCH |
| 11 | Clock Skew (-5 min) | 988 (-5%) | Minor | Instant after removal | 25/25, checksums MATCH |
| 12 | Bandwidth Throttle | 280 (-73%) | Significant | Instant after removal | 25/25, checksums MATCH |
| 13 | Pod Failure (freeze) | 1409 (2 nodes) | Node frozen | Auto-rejoin after expiry | 25/25, checksums MATCH |
| 14 | Container Kill | 1381 (2 nodes) | Process killed | Kubelet restart + rejoin | 25/25, checksums MATCH |
| 15 | Packet Duplicate (50%) | 995 (-4%) | Minor | N/A | 25/25, checksums MATCH |
| 16 | Packet Corrupt (50%) | 0 (all down) | **Complete outage** | Auto-bootstrap after removal | 25/25, checksums MATCH |
| 17 | IO Attr Override (r/o) | 1388 (2 nodes) | Node crash | Coordinator recovery | 25/25, checksums MATCH |
| 18 | IO Mistake (corruption) | 1380 (2 nodes) | Node crash | SST recovery | 25/25, checksums MATCH |

## Key Findings

### Galera Cluster Strengths
1. **Zero data loss** across all 18 experiments — Galera's synchronous replication guarantees consistency
2. **Automatic recovery** from all failure modes — no manual intervention needed
3. **No split-brain** — isolated nodes become `non-Primary` and stop accepting writes
4. **CPU stress resilience** — 98% CPU had virtually no impact (write path is IO-bound)
5. **Packet duplication resilience** — 50% packet duplication caused only 4% TPS drop (TCP handles duplicates natively)
6. **Full cluster kill recovery** — KubeDB coordinator handles Galera bootstrap automatically
7. **IO corruption recovery** — random data corruption and read-only filesystem both recovered via SST/IST

### Galera Cluster Sensitivities
1. **Network latency** — 1s latency caused 99.7% TPS drop (synchronous replication amplifies latency)
2. **Packet loss** — 30% loss caused 99.9% TPS drop (TCP retransmissions compound with certification)
3. **Packet corruption** — 50% corruption caused **complete cluster outage** (all nodes non-Primary) — the only chaos that broke the entire cluster
4. **IO faults** — EIO errors, read-only filesystem, and random corruption all crash MariaDB, but cluster continues on remaining nodes
5. **Bandwidth** — 1 mbps limit caused 73% TPS drop, but flow control kept nodes Synced

### Galera vs MySQL Group Replication

| Aspect | Galera Cluster | MySQL GR (Single-Primary) |
|---|---|---|
| Topology | Multi-master (all nodes write) | Single primary + secondaries |
| Network latency impact | Severe (every write certified across all nodes) | Moderate (only primary writes) |
| CPU stress impact | Negligible | Negligible |
| Packet loss 30% | Severe TPS drop, no expulsion | Node expulsion, failover |
| Recovery from full kill | ~3 min (coordinator bootstrap) | ~1 min (coordinator + GR protocol) |
| Flow control | `wsrep_flow_control_paused` | N/A (group_replication handles internally) |

### Performance Metrics Summary

| Metric | Value |
|---|---|
| **Baseline TPS** | ~1,039 |
| **Best TPS during chaos** | 1,450 (2-node, during partition/IO chaos) |
| **Worst TPS during chaos** | 0 (packet corruption — full outage) |
| **Worst TPS (cluster functional)** | 1.3 (30% packet loss) |
| **Recovery Time (single node)** | ~5-30 seconds |
| **Recovery Time (full cluster kill)** | ~3 minutes |
| **Data Loss** | Zero across all 18 experiments |
| **Tracking Rows Preserved** | 25/25 on every experiment |
| **Checksum Mismatches** | Zero |

### Conclusion

The KubeDB-managed MariaDB Galera Cluster demonstrates excellent resilience across all 18 tested failure scenarios. Key takeaways:

- **Zero data loss** in every experiment — Galera's synchronous replication guarantees consistency
- **Automatic recovery** from all failure modes including full cluster kill — no manual intervention needed
- **No split-brain** — isolated nodes become `non-Primary` and stop accepting writes
- **Network quality is critical** — Galera's synchronous certification amplifies network issues (latency/loss → severe TPS drop)
- **IO failures are handled gracefully** — affected node crashes but remaining nodes continue serving traffic, and the coordinator recovers the crashed node automatically

The cluster achieves a strong balance of high availability and data consistency, making it suitable for production workloads that require zero data loss guarantees.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2026.2.26/setup).
- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2026.2.26/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).

