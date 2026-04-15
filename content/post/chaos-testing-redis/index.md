---
title: Chaos Testing Redis on Kubernetes with KubeDB and Chaos Mesh
date: "2026-04-15"
weight: 14
authors:
- name: Hiranmoy Chowdhury
  image: /images/author/hiranmoy.jpg
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
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f redis-cluster.yaml
```

Wait for Redis to become ready:

```bash
kubectl get redis -n demo -w
```

```
NAME            VERSION   STATUS   AGE
redis-cluster   7.4.0     Ready    3m
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
```

```
"before-chaos"
```

---

## Step 3: Verify Chaos Mesh is Ready

```bash
kubectl get pods -n chaos-mesh
```

```
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-xxxxxxxxx-xxxxx    3/3     Running   0          5m
chaos-daemon-xxxxx                          1/1     Running   0          5m
chaos-daemon-xxxxx                          1/1     Running   0          5m
chaos-dashboard-xxxxxxxxx-xxxxx             1/1     Running   0          5m
```

---

## Chaos Experiments

For each experiment below, the workflow is:

1. Apply the manifest
2. Watch pod and Redis status
3. Verify data is still accessible
4. Delete the experiment and wait for full recovery before running the next one

---

### Experiment 1: Pod Failure

**What it does:** Marks a Redis pod as unavailable without killing the process. This simulates Kubernetes declaring a pod unhealthy.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: redis-pod-failure
  namespace: demo
spec:
  action: pod-failure
  mode: one
  duration: "30s"
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
```

Observe:

```bash
kubectl get pods -n demo -w
kubectl get redis -n demo redis-cluster
```

One pod will go into a not-ready state. After the 30-second duration, the experiment ends and the pod recovers automatically.

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c GET chaos-test
"before-chaos"
```

---

### Experiment 2: Pod Kill

**What it does:** Sends `SIGKILL` to a Redis pod, simulating an OOM kill or a sudden node failure. Unlike pod-failure, the pod process is actually terminated and Kubernetes must reschedule it.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: redis-pod-kill
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
```

Observe:

```bash
kubectl get pods -n demo -w
```

The killed pod transitions to `Terminating` then gets recreated by the StatefulSet controller. KubeDB reconciles the Redis resource back to the desired state.

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=redis-cluster
```

```
NAME                      READY   STATUS    RESTARTS   AGE
redis-cluster-shard0-0    1/1     Running   1          15m
```

Verify data:

```bash
kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c GET chaos-test
  
"before-chaos"
```

---

### Experiment 3: Container Kill

**What it does:** Kills only the `redis` container inside a pod, without deleting the pod itself. This tests in-place container restart behavior and is useful when sidecars or init containers are involved.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: redis-container-kill
  namespace: demo
spec:
  action: container-kill
  mode: one
  containerNames:
    - redis
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
```

Observe:

```bash
kubectl get pods -n demo -w
```

The pod stays alive. Only the `redis` container restarts. Once it comes back up, it rejoins the cluster automatically.

---

### Experiment 4: Network Delay

**What it does:** Injects artificial latency into the network traffic of Redis pods. This simulates cross-availability-zone communication, saturated network links, or noisy neighbour conditions.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: redis-network-delay
  namespace: demo
spec:
  action: delay
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
  delay:
    latency: "200ms"
    correlation: "25"
    jitter: "50ms"
  duration: "60s"
```

Observe:

```bash
time kubectl exec -it -n demo redis-cluster-shard0-0 -- \
  redis-cli -a $PASSWORD -c PING
```

Commands that normally complete in under 1ms will now take 200ms or more. After the 60-second window, latency returns to normal.

---

### Experiment 5: Network Loss

**What it does:** Randomly drops packets for Redis pods. This simulates flaky networks, overloaded switches, or unreliable cloud networking.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: redis-network-loss
  namespace: demo
spec:
  action: loss
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
  loss:
    loss: "50"
    correlation: "25"
  duration: "60s"
```

Observe:

Some Redis commands will time out or fail during this window. After the 60-second experiment, the packet loss stops and clients reconnect automatically. Use this experiment to validate your application's retry logic.

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
  name: redis-network-partition
  namespace: demo
spec:
  action: partition
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      app.kubernetes.io/instance: "redis-cluster"
      app.kubernetes.io/name: "redis"
  direction: both
  duration: "30s"
```

Observe:

```bash
kubectl get redis -n demo redis-cluster -o yaml
kubectl get pods -n demo -w
```

The isolated pod is cut off from its peers. Redis Cluster will mark the affected shard as unavailable. Once the 30-second window ends, the partition is lifted and the cluster re-converges.


---

### Experiment 8: CPU Stress

**What it does:** Saturates the CPU of a Redis pod to simulate CPU-intensive workloads on the same node or a resource-constrained environment.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: redis-cpu-stress
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
    cpu:
      workers: 2
      load: 100
  duration: "60s"
```

Observe:

```bash
kubectl top pods -n demo
```

Redis is single-threaded for command processing, so heavy CPU stress will noticeably increase command latency. After the experiment, CPU usage drops and performance returns to baseline.

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
kubectl describe pod -n demo redis-cluster-shard0-0 | grep -A5 "Limits\|Requests"
```

Watch whether Redis begins evicting keys (depending on your `maxmemory-policy` setting) and whether it recovers cleanly once the stressor is removed.

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
