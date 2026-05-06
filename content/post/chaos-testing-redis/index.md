---
title: Chaos Testing Redis on Kubernetes with KubeDB and Chaos Mesh
date: "2026-04-15"
weight: 14
authors:
- Hiranmoy Das Chowdhury
tags:
- chaos-engineering
- chaos-mesh
- database
- kubedb
- kubernetes
- redis
coverImage: /images/blog/chaos-testing-redis/redis-chaos.png
---

[Redis](https://redis.io/) is a popular in-memory data store used for caching, session management, pub/sub messaging, leaderboards, and real-time analytics. In many architectures, Redis sits in the critical path — a failure can cause cascading degradation across the entire application stack.

**Chaos engineering** is the discipline of proactively injecting failures into a system to discover weaknesses before they manifest as production incidents.

In this post, we will:
1. Deploy a production-style Redis Cluster using **[KubeDB](https://kubedb.com)**
2. Inject ten different failure scenarios using **[Chaos Mesh](https://chaos-mesh.org)**
3. Observe and validate how Redis and KubeDB respond to each fault

---

## What is Chaos Engineering?

Chaos engineering is the practice of intentionally introducing controlled failures into a system to observe how it behaves. The goal is not to break things — it is to learn how the system responds so you can make it more resilient.

For a Redis deployment, the questions chaos engineering helps answer are:

- Does Redis recover automatically after a pod is killed?
- Does the cluster re-converge after a network partition?
- Can clients reconnect seamlessly after a disruption?
- Does KubeDB restore the desired state after a fault?
- How does Redis behave under CPU or memory pressure?

---

## Tools Used

| Tool | Purpose |
|---|---|
| [KubeDB](https://kubedb.com) | Manages Redis lifecycle on Kubernetes |
| [Chaos Mesh](https://chaos-mesh.org) | Injects chaos experiments into the cluster |
| Redis 7.4.0 | The database under test |

---

## Prerequisites

Before you begin, make sure you have the following:

- A running Kubernetes cluster (GKE, EKS, AKS, or a local cluster using Kind or Minikube)
- `kubectl` configured to access the cluster
- KubeDB operator installed — follow the [KubeDB setup guide](https://kubedb.com/docs/v2026.2.26/setup/)
- Chaos Mesh installed — follow the [Chaos Mesh installation guide](https://chaos-mesh.org/docs/production-installation-using-helm/)
- A default or usable `StorageClass` in the cluster

---

## Step 1: Deploy Redis Cluster with KubeDB

We will deploy a Redis Cluster with 3 shards and 2 replicas per shard. This gives us a proper distributed setup where we can observe failover and re-convergence behavior.

Create the namespace:

```bash
kubectl create ns demo
```

Apply the Redis manifest:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version: "7.4.0"
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f redis-cluster.yaml
```

Wait for Redis to become ready:

```bash
NAME                             VERSION   STATUS   AGE
redis.kubedb.com/redis-cluster   7.4.0     Ready    73s

NAME                         READY   STATUS    RESTARTS   AGE
pod/redis-cluster-shard0-0   1/1     Running   0          70s
pod/redis-cluster-shard0-1   1/1     Running   0          54s
pod/redis-cluster-shard1-0   1/1     Running   0          67s
pod/redis-cluster-shard1-1   1/1     Running   0          53s
pod/redis-cluster-shard2-0   1/1     Running   0          66s
pod/redis-cluster-shard2-1   1/1     Running   0          53s
```


## Step 2: Verify Redis is Healthy

Retrieve the Redis password from the secret KubeDB created:

```bash
export PASSWORD=$(kubectl get secret -n demo redis-cluster-auth \
  -o jsonpath='{.data.password}' | base64 -d)
```

Check the cluster state:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c CLUSTER INFO
```

Write a test key that we will check after each experiment to verify data integrity:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c SET chaos-test "before-chaos"
```

Read it back:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
                              redis-cli -a $PASSWORD -c GET chaos-test
Defaulted container "redis" out of: redis, redis-init (init)
"before-chaos"

```

---

## Step 3: Verify Chaos Mesh is Ready

```bash
kubectl get pods -n chaos-mesh
NAME                                      READY   STATUS    RESTARTS   AGE
chaos-controller-manager-b8d65b98-75s8w   1/1     Running   0          2m15s
chaos-controller-manager-b8d65b98-jcmnt   1/1     Running   0          2m13s
chaos-controller-manager-b8d65b98-tfwfd   1/1     Running   0          2m14s
chaos-daemon-jhth2                        1/1     Running   0          2m15s
chaos-dashboard-566b9f5c4b-zmplh          1/1     Running   0          2m15s
chaos-dns-server-85b8846dc9-ksljn         1/1     Running   0          116m
```

---

## Chaos Experiments

For each experiment below, the workflow is:

1. Apply the manifest
2. Watch pod and Redis status
3. Verify data is still accessible
4. Delete the experiment and wait for full recovery before running the next one

---

### Experiment 1: Master Pod Failure

**What it does:** Marks a Redis pod as unavailable without killing the process. This simulates Kubernetes declaring a pod unhealthy.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: rd-master-pod-failure-short
  namespace: demo
spec:
  action: pod-failure
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/name: redises.kubedb.com
      kubedb.com/role: master
  duration: "1m"
```

Observe:

```bash
kubectl get rd,rds,rdops,rdsops,pods -n demo
NAME                             VERSION   STATUS   AGE
redis.kubedb.com/redis-cluster   7.4.0     Ready    39m

NAME                         READY   STATUS    RESTARTS         AGE
pod/redis-cluster-shard0-0   1/1     Running   10 (3m16s ago)   39m
pod/redis-cluster-shard0-1   1/1     Running   0                39m
pod/redis-cluster-shard1-0   1/1     Running   10 (3m16s ago)   39m
pod/redis-cluster-shard1-1   1/1     Running   0                39m
pod/redis-cluster-shard2-0   1/1     Running   10 (3m16s ago)   39m
pod/redis-cluster-shard2-1   1/1     Running   0                39m
```

One pod will go into a not-ready state. After the 30-second duration, the experiment ends and the pod recovers automatically.

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
                                    redis-cli -a $PASSWORD -c GET chaos-test
                                    
Defaulted container "redis" out of: redis, redis-init (init)
"before-chaos"
```

markdown
**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 11:39:00 | — | Pre-chaos baseline (all 3 master pods ONLINE) | `Ready` |
| 11:44:07 | 0s | `PodChaos` applied (all master pods targeted) | `Ready` |
| 11:44:07 | +0s | All master pods marked unavailable | `NotReady` |
| 11:44:21 | +14s | Operator marks masters unhealthy | `Critical` |
| 11:45:07 | +1m00s | Chaos auto-recovered, master pods restart | `Critical` |
| 11:45:40 | ~+1m33s | Pods in `RECOVERING` (rejoining cluster) | `Critical` |
| ~11:48–52 | ~+4–8m | All masters reach `Running`, cluster re-converges | `Ready` |

**Result: PASS** — All master pods recovered automatically after the 1-minute fault window. KubeDB reconciled the Redis Cluster back to the desired state with zero data loss. The chaos-impacted pods rejoined the cluster and the test key `chaos-test` remained intact.

---

### Experiment 2: Pod Kill Master

**What it does:** Sends `SIGKILL` to a Redis pod, simulating an OOM kill or a sudden node failure. Unlike pod-failure, the pod process is actually terminated and Kubernetes must reschedule it.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: rd-master-pod-kill-short
  namespace: demo
spec:
  action: pod-kill
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/name: redises.kubedb.com
      kubedb.com/role: master
  duration: "5m"
  gracePeriod: 0
```

Observe:

```bash
kubectl get rd,rds,rdops,rdsops,pods -n demo
NAME                             VERSION   STATUS   AGE
redis.kubedb.com/redis-cluster   7.4.0     Ready    72m

NAME                         READY   STATUS    RESTARTS   AGE
pod/redis-cluster-shard0-0   1/1     Running   0          6m55s
pod/redis-cluster-shard0-1   1/1     Running   0          72m
pod/redis-cluster-shard1-0   1/1     Running   0          6m55s
pod/redis-cluster-shard1-1   1/1     Running   0          72m
pod/redis-cluster-shard2-0   1/1     Running   0          6m55s
pod/redis-cluster-shard2-1   1/1     Running   0          72m
```

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c GET chaos-test
  
"before-chaos"
```

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 12:14:19 | — | Pre-chaos baseline (all 3 master pods ONLINE) | `Ready` |
| 12:14:19 | 0s | `PodChaos` applied — all master pods SIGKILLed | `Ready` |
| 12:14:19 | +0s | All master pods terminated (`pod-kill` injected for shard0-0, shard1-0, shard2-0) | `Critical` |
| 12:14:30 | ~+11s | Kubernetes reschedules master pods; replicas promoted | `Critical` |
| 12:14:55 | ~+36s | New master pods reach `Running`; cluster re-elects primaries | `Critical` |
| 12:19:19 | +5m00s | Chaos experiment window ends | `Critical` |
| ~12:21:00 | ~+7m | All shards re-converge; cluster topology stabilized | `Ready` |

**Result: PASS** — All master pods were hard-killed simultaneously. Kubernetes rescheduled them via the StatefulSet controller and KubeDB reconciled the cluster back to the desired state. Replica pods were promoted to masters during the kill window and the test key `chaos-test` remained intact after full recovery.

---

### Experiment 3: Container Kill

**What it does:** Kills only the `redis` container inside a pod, without deleting the pod itself. This tests in-place container restart behavior and is useful when sidecars or init containers are involved.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: rd-master-container-kill-short
  namespace: demo
spec:
  action: container-kill
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/name: redises.kubedb.com
      kubedb.com/role: master
  duration: "5m"
  containerNames: ['redis']
```

Observe:

```bash
kubectl get rd,rds,rdops,rdsops,pods -n demo
NAME                             VERSION   STATUS   AGE
redis.kubedb.com/redis-cluster   7.4.0     Ready    80m

NAME                         READY   STATUS    RESTARTS      AGE
pod/redis-cluster-shard0-0   1/1     Running   1 (61s ago)   14m
pod/redis-cluster-shard0-1   1/1     Running   0             79m
pod/redis-cluster-shard1-0   1/1     Running   1 (61s ago)   14m
pod/redis-cluster-shard1-1   1/1     Running   0             79m
pod/redis-cluster-shard2-0   1/1     Running   1 (61s ago)   14m
pod/redis-cluster-shard2-1   1/1     Running   0             79m
```

The pod stays alive. Only the `redis` container restarts. Once it comes back up, it rejoins the cluster automatically.

markdown
**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 12:27:43 | — | Pre-chaos baseline (all 3 master pods ONLINE) | `Ready` |
| 12:27:43 | 0s | `PodChaos` applied — `redis` container killed in shard0-0, shard1-0, shard2-0 | `Ready` |
| 12:27:43 | +0s | All 3 master containers killed simultaneously (`container-kill` injected) | `Critical` |
| 12:27:44 | ~+1s | Kubernetes detects container exit; kubelet restarts `redis` container in-place | `Critical` |
| 12:28:04 | ~+21s | Containers restart with 1 restart count; pods remain scheduled on same nodes | `Critical` |
| 12:28:44 | ~+1m | Restarted containers rejoin the Redis Cluster; cluster re-converges | `Ready` |
| 12:32:43 | +5m00s | Chaos experiment window ends; no further kills injected | `Ready` |

**Result: PASS** — All three master containers were killed simultaneously. Kubernetes restarted them in-place (pod identity preserved, no rescheduling needed). KubeDB reconciled the cluster back to the desired state with zero data loss. The test key `chaos-test` remained intact after full recovery.

---

### Experiment 4: Network Delay

**What it does:** Injects artificial latency into the network traffic of Redis pods. This simulates cross-availability-zone communication, saturated network links, or noisy neighbour conditions.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: delay
  namespace: demo
spec:
  action: delay
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/name: redises.kubedb.com
      kubedb.com/role: master
  delay:
    latency: '1000ms'
    correlation: '100'
    jitter: '50ms'
  duration: "60s"
```

Observe:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c PING
```

Commands that normally complete in under 1ms will now take 1000ms or more. After the 60-second window, latency returns to normal.


markdown
**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| T+0:00 | — | Pre-chaos baseline (all 6 pods ONLINE) | `Ready` |
| T+0:00 | 0s | `NetworkChaos` applied — 1000ms latency injected on all master pods | `Ready` |
| T+0:05 | ~+5s | Redis commands begin timing out; write attempts fail with `context deadline exceeded` | `NotReady` |
| T+0:17 | ~+17s | Cluster marks affected shards unavailable; KubeDB detects degraded state | `NotReady` |
| T+1:00 | +60s | Chaos experiment window ends; latency injection removed | `NotReady` |
| T+3:00 | ~+3m | All shards re-converge; KubeDB reconciles cluster to desired state | `Ready` |

**Result: PASS** — During the 1000ms latency window, Redis write operations failed with `context deadline exceeded` errors due to cluster heartbeat timeouts exceeding the configured threshold. After the chaos window ended, all pods recovered automatically and KubeDB reconciled the cluster back to `Ready`. Data integrity was confirmed — the test key `chaos-test` remained intact after full recovery.

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c GET chaos-test

"before-chaos"
```

---

### Experiment 5: Network Bandwidth

markdown
**What it does:** Throttles the network bandwidth available to all Redis pods to 1 Mbps. This simulates a congested or limited network link, such as a cross-region replication scenario or a degraded network interface.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: bandwidth
  namespace: demo
spec:
  action: bandwidth
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/name: redises.kubedb.com
  bandwidth:
    rate: 1mbps
    limit: 20971520
    buffer: 10000
  duration: "60s"
```

Apply and observe:

```bash
kubectl get rd,pods -n demo
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c PING
```

With only 1 Mbps available, Redis inter-node gossip and replication traffic will compete with client traffic. Large key writes or bulk operations will slow significantly. After the 60-second window, bandwidth returns to normal.

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| T+0:00 | — | Pre-chaos baseline (all 6 pods ONLINE) | `Ready` |
| T+0:00 | 0s | `NetworkChaos` applied — 1 Mbps cap injected on all pods | `Ready` |
| T+0:05 | ~+5s | Replication and gossip traffic degraded; write latency increases | `Ready` |
| T+0:30 | ~+30s | Cluster remains functional but throughput is visibly throttled | `Ready` |
| T+1:00 | +60s | Chaos experiment window ends; bandwidth restriction lifted | `Ready` |
| T+1:10 | ~+1m10s | All pods return to full throughput; cluster fully converged | `Ready` |

**Result: PASS** — The 1 Mbps bandwidth cap throttled replication and gossip traffic across all shards. Redis remained available throughout the experiment with elevated latency on large operations, but the cluster did not lose quorum. After the chaos window ended, all pods returned to normal throughput and the test key `chaos-test` remained intact.

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c GET chaos-test

"before-chaos"
```

---

### Experiment 6: Network Corruption

**What it does:** Corrupts a percentage of network packets for Redis pods. This simulates bit-flip errors from faulty hardware or bad cables.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: redis-network-corruption
  namespace: demo
spec:
  action: corrupt
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
  corrupt:
    corrupt: "40"
    correlation: "25"
  duration: "60s"
```

Observe:

TCP will detect and retransmit corrupted packets, so most Redis commands will still succeed but with higher latency and occasional errors. After the experiment, the network returns to normal.

---

### Experiment 7: Network Partition

**What it does:** Completely isolates one Redis pod from the rest of the cluster. This is the most aggressive network experiment and tests how the Redis Cluster handles a split-brain scenario.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: partition
  namespace: demo
spec:
  action: partition
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/name: redises.kubedb.com
      kubedb.com/role: master
  direction: both
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        app.kubernetes.io/name: redises.kubedb.com
        kubedb.com/role: slave
  duration: "10m"
```

Observe:

```bash
kubectl get rd,rds,rdops,rdsops,pods -n demo
NAME                             VERSION   STATUS     AGE
redis.kubedb.com/redis-cluster   7.4.0     Critical   119m

NAME                         READY   STATUS    RESTARTS      AGE
pod/redis-cluster-shard0-0   1/1     Running   1 (40m ago)   53m
pod/redis-cluster-shard0-1   1/1     Running   0             118m
pod/redis-cluster-shard1-0   1/1     Running   1 (40m ago)   53m
pod/redis-cluster-shard1-1   1/1     Running   0             118m
pod/redis-cluster-shard2-0   1/1     Running   1 (40m ago)   53m
pod/redis-cluster-shard2-1   1/1     Running   0             118m
```

The isolated pod is cut off from its peers. Redis Cluster will mark the affected shard as unavailable. Once the 10m window ends, the partition is lifted and the cluster re-converges.

observe after 10m:

```bash
kubectl get rd,rds,rdops,rdsops,pods -n demo
NAME                             VERSION   STATUS   AGE
redis.kubedb.com/redis-cluster   7.4.0     Ready    126m

NAME                         READY   STATUS    RESTARTS      AGE
pod/redis-cluster-shard0-0   1/1     Running   1 (47m ago)   60m
pod/redis-cluster-shard0-1   1/1     Running   0             126m
pod/redis-cluster-shard1-0   1/1     Running   1 (47m ago)   60m
pod/redis-cluster-shard1-1   1/1     Running   0             126m
pod/redis-cluster-shard2-0   1/1     Running   1 (47m ago)   60m
pod/redis-cluster-shard2-1   1/1     Running   0             126m
```
markdown

**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 13:04:51 | — | Pre-chaos baseline (all 3 master pods ONLINE) | `Ready` |
| 13:04:51 | 0s | `NetworkChaos` applied — bidirectional partition between all masters and all replicas | `Ready` |
| 13:04:52 | ~+1s | All 6 pods partitioned (masters isolated from replicas across all 3 shards) | `Critical` |
| 13:05:10 | ~+20s | Redis Cluster marks affected shards unavailable; KubeDB detects degraded state | `Critical` |
| 13:14:51 | +10m00s | Chaos experiment window ends; partition lifted on all pods | `Critical` |
| ~13:16:00 | ~+11m | All shards re-converge; cluster topology stabilized | `Ready` |

**Result: PASS** — All master pods were bidirectionally partitioned from their replicas simultaneously. During the 10-minute window, the Redis Cluster entered a `Critical` state as shards lost quorum visibility. Once the partition was lifted, all pods automatically re-converged and KubeDB reconciled the cluster back to `Ready`. The test key `chaos-test` remained intact after full recovery.

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c GET chaos-test

"before-chaos"
```
---

### Experiment 8: CPU Stress

**What it does:** Saturates the CPU of a Redis pod to simulate CPU-intensive workloads on the same node or a resource-constrained environment.

current CPU usage:

```bash
kubectl top pods -n demo
NAME                     CPU(cores)   MEMORY(bytes)   
redis-cluster-shard0-0   3m           5Mi             
redis-cluster-shard0-1   4m           6Mi             
redis-cluster-shard1-0   4m           5Mi             
redis-cluster-shard1-1   4m           5Mi             
redis-cluster-shard2-0   3m           4Mi             
redis-cluster-shard2-1   3m           6Mi
```

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: memory-stress-example
  namespace: demo
spec:
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/name: redises.kubedb.com
      kubedb.com/role: master
  duration: "2m"
  stressors:
    cpu:
      workers: 4
      load: 50
```
Apply and Observe:

```bash
kubectl top pods -n demo
NAME                     CPU(cores)   MEMORY(bytes)   
redis-cluster-shard0-0   1901m        9Mi             
redis-cluster-shard0-1   3m           5Mi             
redis-cluster-shard1-0   2007m        9Mi             
redis-cluster-shard1-1   3m           6Mi             
redis-cluster-shard2-0   1104m        9Mi             
redis-cluster-shard2-1   3m           5Mi
```

Redis is single-threaded for command processing, so heavy CPU stress will noticeably increase command latency. After the experiment, CPU usage drops and performance returns to baseline.

markdown
**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 13:28:05 | — | Pre-chaos baseline (all 3 master pods ONLINE, ~3–4m CPU) | `Ready` |
| 13:28:05 | 0s | `StressChaos` applied — 4 workers at 50% CPU load injected on all master pods | `Ready` |
| 13:28:05 | +0s | CPU stress injected into shard0-0, shard1-0, shard2-0 | `Ready` |
| 13:28:10 | ~+5s | CPU usage spikes to ~1100–2000m per master pod; command latency increases | `Ready` |
| 13:28:05 | ~+30s | Redis remains available; single-threaded command processing visibly slower | `Ready` |
| 13:30:05 | +2m00s | Chaos experiment window ends; CPU stress removed from all master pods | `Ready` |
| 13:30:10 | ~+2m05s | CPU usage returns to baseline (~3–4m); cluster fully converged | `Ready` |

**Result: PASS** — CPU stress raised master pod usage from ~3–4m to ~1100–2000m cores. Redis remained available throughout the experiment, but command latency increased due to CPU contention. After the 2-minute chaos window ended, all pods recovered to baseline CPU usage automatically. The test key `chaos-test` remained intact.

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c GET chaos-test

"before-chaos"
```

---

### Experiment 9: Memory Stress

**What it does:** Allocates a large chunk of memory inside a Redis pod to simulate memory pressure, approaching OOM conditions.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: redis-memory-stress
  namespace: demo
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
  stressors:
    memory:
      workers: 2
      size: "256MB"
  duration: "60s"
```

Observe:

```bash
kubectl top pods -n demo
NAME                     CPU(cores)   MEMORY(bytes)   
redis-cluster-shard0-0   3m           5Mi             
redis-cluster-shard0-1   4m           6Mi             
redis-cluster-shard1-0   4m           252Mi           
redis-cluster-shard1-1   4m           5Mi             
redis-cluster-shard2-0   3m           4Mi             
redis-cluster-shard2-1   3m           5Mi
```

Watch whether Redis begins evicting keys (depending on your `maxmemory-policy` setting) and whether it recovers cleanly once the stressor is removed.

markdown
**Observed timeline:**

| Wall-clock | Δ from chaos | Event | DB Status |
|---|---|---|---|
| 14:04:17 | — | Pre-chaos baseline (all master pods ONLINE, ~5–6Mi memory) | `Ready` |
| 14:04:17 | 0s | `StressChaos` applied — 2 workers allocating 256MB injected on one master pod | `Ready` |
| 14:04:17 | +0s | Memory stress injected into `redis-cluster-shard1-0` | `Ready` |
| 14:04:22 | ~+5s | Memory usage on `shard1-0` spikes to ~252Mi; Redis under pressure | `Ready` |
| 14:04:47 | ~+30s | Redis remains available; no OOM kill triggered within limits | `Ready` |
| 14:05:17 | +1m00s | Chaos experiment window ends; memory stress removed from pod | `Ready` |
| 14:05:20 | ~+1m03s | Memory usage returns to baseline (~5Mi); cluster fully converged | `Ready` |

**Result: PASS** — Memory stress raised `shard1-0` usage from ~5Mi to ~252Mi. Redis remained available throughout the experiment and no OOM kill was triggered. After the 60-second chaos window ended, the stressor was removed and memory returned to baseline automatically. The test key `chaos-test` remained intact.

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c GET chaos-test

"before-chaos"
` `` 
```

---

### Experiment 10: I/O Chaos

**What it does:** Injects latency into disk I/O operations for the Redis data directory. This simulates a slow or degraded storage backend affecting AOF writes and RDB snapshots.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: redis-io-delay
  namespace: demo
spec:
  action: latency
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
  volumePath: /data
  path: "/data/**"
  delay: "100ms"
  percent: 50
  duration: "60s"
```

Observe:

```bash
kubectl logs -n demo redis-cluster-shard0-0 --tail=50 -f
```

Redis will log warnings about slow AOF flushes or delayed RDB saves. After the experiment ends, I/O returns to normal and persistence operations resume.

---

## Summary of Experiments

| # | Experiment | Chaos Kind | What is Validated                         |
|---|---|---|-------------------------------------------|
| 1 | Pod Failure | PodChaos | Pod unavailability, automatic recovery    |
| 2 | Pod Kill | PodChaos | Hard kill, PetSet reschedule behavior     |
| 3 | Container Kill | PodChaos | In-place container restart                |
| 4 | Network Delay | NetworkChaos | High latency, client timeout behavior     |
| 5 | Network Loss | NetworkChaos | Packet drops, client retry behavior       |
| 6 | Network Corruption | NetworkChaos | Corrupted packets, TCP retransmission     |
| 7 | Network Partition | NetworkChaos | Cluster split, re-convergence             |
| 8 | CPU Stress | StressChaos | Performance under CPU saturation          |
| 9 | Memory Stress | StressChaos | OOM behavior, key eviction                |
| 10 | I/O Chaos | IOChaos | Storage degradation, persistence behavior |

---

## Cleanup

Delete the Redis cluster:

```bash
kubectl delete redis -n demo redis-cluster
```

Delete the namespace:

```bash
kubectl delete ns demo
```
