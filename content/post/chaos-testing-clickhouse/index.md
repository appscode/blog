---
title: 'Chaos Testing KubeDB ClickHouse: Building Resilience with Chaos Mesh'
date: "2026-09-03"
weight: 26
authors:
- Shuvo Kumar
tags:
- chaos-engineering
- chaos-mesh
- clickhouse
- cloud-native
- database
- disaster-recovery
- high-availability
- kubedb
- kubernetes
---

# Chaos Testing KubeDB ClickHouse: Building Resilience with Chaos Mesh

> New to KubeDB? Start with the [KubeDB documentation](https://kubedb.com/docs/).

## Setup Cluster

To follow this tutorial, you need:

1. A running Kubernetes cluster.
2. KubeDB installed in the cluster.
3. `kubectl` configured for that cluster.
4. Chaos Mesh installed with the correct container runtime socket.
5. A disposable `demo` namespace. Never run these tests against valuable
   data.

For K3s, Chaos Mesh must use `/run/k3s/containerd/containerd.sock`:

```bash
➤ helm upgrade --install chaos-mesh chaos-mesh/chaos-mesh \
        --namespace chaos-mesh \
        --create-namespace \
        --version 2.8.4 \
        --set dashboard.create=true \
        --set dashboard.securityMode=false \
        --set dnsServer.create=true \
        --set chaosDaemon.runtime=containerd \
        --set chaosDaemon.socketPath=/run/k3s/containerd/containerd.sock \
        --set chaosDaemon.privileged=true
Release "chaos-mesh" does not exist. Installing it now.
NAME: chaos-mesh
LAST DEPLOYED: Tue Sep  8 10:13:50 2026
NAMESPACE: chaos-mesh
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
```

If you use a different Kubernetes distribution, change the runtime and socket
path to match it.

Create the disposable namespace:

```bash
kubectl create ns demo
```

Create the local manifest directories:

```bash
mkdir -p clickhouse-chaos-testing/setup
mkdir -p clickhouse-chaos-testing/tests
cd clickhouse-chaos-testing
```

Confirm that the ClickHouse version and storage class used by this guide are
available:

```bash
➤ kubectl get clickhouseversion 26.2.6
NAME     VERSION   DB_IMAGE                                        DEPRECATED   AGE
26.2.6   26.2.6    docker.io/clickhouse/clickhouse-server:26.2.6                27d
```

```bash
➤ kubectl get storageclass local-path
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   true                   27d

```

## Verify KubeDB and Chaos Mesh Installation

```bash
➤ kubectl get pods -n kubedb
NAME                                            READY   STATUS             RESTARTS         AGE
kubedb-kubedb-autoscaler-0                      1/1     Running            3 (2d9h ago)     21d
kubedb-kubedb-ops-manager-0                     1/1     Running            1 (2d9h ago)     7d22h
kubedb-kubedb-provisioner-0                     1/1     Running            1 (2d9h ago)     5d13h
kubedb-kubedb-webhook-server-65949766c4-xlfpz   1/1     Running            1 (2d9h ago)     12d
kubedb-petset-85c9d79865-lfqww                  1/1     Running            3 (2d9h ago)     21d
kubedb-sidekick-86f897c579-djk65                1/1     Running            3 (2d9h ago)     21d
```

```bash
➤ kubectl get pods -n chaos-mesh

NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7d44d6dd54-7chxr   1/1     Running   0          3m50s
chaos-controller-manager-7d44d6dd54-fd5cv   1/1     Running   0          3m50s
chaos-controller-manager-7d44d6dd54-prstq   1/1     Running   0          3m50s
chaos-daemon-dkfxd                          1/1     Running   0          3m50s
chaos-dashboard-5db97f969f-gwm96            1/1     Running   0          3m50s
chaos-dns-server-6d8fd4b8b5-kjwg7           1/1     Running   0          3m50s

```

```bash
➤ kubectl get crd \
        podchaos.chaos-mesh.org \
        networkchaos.chaos-mesh.org \
        stresschaos.chaos-mesh.org \
        iochaos.chaos-mesh.org \
        dnschaos.chaos-mesh.org \
        timechaos.chaos-mesh.org
NAME                          CREATED AT
podchaos.chaos-mesh.org       2026-08-28T07:50:32Z
networkchaos.chaos-mesh.org   2026-08-28T07:50:32Z
stresschaos.chaos-mesh.org    2026-08-28T07:50:32Z
iochaos.chaos-mesh.org        2026-08-28T07:50:32Z
dnschaos.chaos-mesh.org       2026-08-28T07:50:32Z
timechaos.chaos-mesh.org      2026-08-28T07:50:33Z

```

All operator and Chaos Mesh pods must be Ready before continuing.

## Introduction to Chaos Engineering

Chaos engineering deliberately injects controlled failures before those
failures happen unexpectedly in production. A database test must establish
more than whether a pod restarted: it must prove client availability,
acknowledged-data safety, replica convergence, coordination health, and clean
fault removal.

### What This Blog Covers

In this guide, we will:

1. Deploy a highly available KubeDB-managed ClickHouse cluster.
2. Run a continuous workload that inserts rows with unique IDs.
3. Compare the expected behavior with the behavior actually observed.
4. Verify recovery by running database queries and checking replica
   consistency, Keeper health, pod state, and storage state.

The experiments were executed against a fresh KubeDB-managed ClickHouse
26.2.6 cluster with two shards, two replicas per shard, and three dedicated
ClickHouse Keeper members.

All 25 experiments used the same `clickhouse-chaos` resource and accumulated
dataset. They preserved ClickHouse data and passed their availability and
integrity gates. Experiments 19 and 22 required one manual `SIGCONT` command
because Chaos Mesh did not resume the target process while cleaning up
IOChaos or TimeChaos. Experiment 23 cleaned up automatically in this fresh
run. These were Chaos Mesh cleanup observations, not ClickHouse data failures. Both replicas
of each shard had identical row counts and payload checksums, replication
queues were empty, and Keeper had exactly one leader and two followers.

## Create a ClickHouse Cluster

The test topology deliberately separates two different kinds of redundancy:

- Two replicas protect each data shard.
- Three Keeper members provide a coordination quorum.

#### Create `setup/clickhouse-chaos.yaml`
Save the following manifest as `setup/clickhouse-chaos.yaml`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-chaos
  namespace: demo
  labels:
    chaos-test.kubedb.com/suite: clickhouse-chaos
spec:
  version: 26.2.6
  clusterTopology:
    cluster:
      name: chaos-cluster
      shards: 2
      replicas: 2
      storageType: Durable
      storage:
        storageClassName: local-path
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 4Gi
      podTemplate:
        spec:
          containers:
            - name: clickhouse
              resources:
                requests:
                  cpu: 250m
                  memory: 1Gi
                limits:
                  cpu: "1"
                  memory: 4Gi
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storageType: Durable
        storage:
          storageClassName: local-path
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        podTemplate:
          spec:
            containers:
              - name: clickhouse-keeper
                resources:
                  requests:
                    cpu: 100m
                    memory: 256Mi
                  limits:
                    cpu: 500m
                    memory: 512Mi
  deletionPolicy: WipeOut
```

Create the cluster and wait for the database:

```bash
➤ kubectl apply -f setup/clickhouse-chaos.yaml
clickhouse.kubedb.com/clickhouse-chaos created
```

```bash
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=15m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

```bash
➤ kubectl get clickhouse,petset,pods,pvc -n demo
NAME                                     VERSION   STATUS   AGE
clickhouse.kubedb.com/clickhouse-chaos   26.2.6    Ready    2m21s

NAME                                                                  AGE
petset.apps.k8s.appscode.com/clickhouse-chaos-chaos-cluster-shard-0   2m16s
petset.apps.k8s.appscode.com/clickhouse-chaos-chaos-cluster-shard-1   2m13s
petset.apps.k8s.appscode.com/clickhouse-chaos-keeper                  2m18s

NAME                                           READY   STATUS    RESTARTS   AGE
pod/clickhouse-chaos-chaos-cluster-shard-0-0   1/1     Running   0          2m15s
pod/clickhouse-chaos-chaos-cluster-shard-0-1   1/1     Running   0          2m9s
pod/clickhouse-chaos-chaos-cluster-shard-1-0   1/1     Running   0          2m13s
pod/clickhouse-chaos-chaos-cluster-shard-1-1   1/1     Running   0          2m8s
pod/clickhouse-chaos-keeper-0                  1/1     Running   0          2m17s
pod/clickhouse-chaos-keeper-1                  1/1     Running   0          2m11s
pod/clickhouse-chaos-keeper-2                  1/1     Running   0          2m6s

NAME                                                                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-0-0   Bound    pvc-ff62c4b6-3ab2-420f-afe1-0cbaa951cadc   4Gi        RWO            local-path     <unset>                 2m15s
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-0-1   Bound    pvc-4ce90a1f-0135-42aa-895f-8a84c2bb5506   4Gi        RWO            local-path     <unset>                 2m9s
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-1-0   Bound    pvc-e858eab7-2a22-4fd9-8543-957efaeb6b85   4Gi        RWO            local-path     <unset>                 2m13s
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-1-1   Bound    pvc-5c795565-f990-4af8-880f-c110bb1b96df   4Gi        RWO            local-path     <unset>                 2m8s
persistentvolumeclaim/data-clickhouse-chaos-keeper-0                  Bound    pvc-62b1a8b4-13e8-4485-9a18-5ffd6e4165f4   1Gi        RWO            local-path     <unset>                 2m17s
persistentvolumeclaim/data-clickhouse-chaos-keeper-1                  Bound    pvc-63f6a03b-b4fb-4a5a-903b-d54e278d8508   1Gi        RWO            local-path     <unset>                 2m11s
persistentvolumeclaim/data-clickhouse-chaos-keeper-2                  Bound    pvc-65ecb3a0-d259-4cdc-b8df-0c61b863afd1   1Gi        RWO            local-path     <unset>                 2m6s

```

KubeDB creates and references `clickhouse-chaos-auth` automatically:

```bash
➤ kubectl get secret -n demo clickhouse-chaos-auth
NAME                    TYPE                       DATA   AGE
clickhouse-chaos-auth   kubernetes.io/basic-auth   2      4m18s
```

### Test Environment

| Component | Value |
| --- | --- |
| Kubernetes | K3s v1.36.3+k3s1, single node |
| Namespace | `demo` |
| ClickHouse resource | `clickhouse-chaos` |
| ClickHouse version | `26.2.6` |
| Data topology | 2 shards × 2 replicas |
| Coordination | 3 ClickHouse Keeper members |
| Storage | 4Gi `local-path` PVC per ClickHouse pod |
| ClickHouse limit | 1 CPU, 4Gi memory |
| Keeper limit | 500m CPU, 512 MiB memory |
| Chaos Mesh | 2.8.4 |
| Container runtime | K3s containerd |
| Chaos daemon socket | `/run/k3s/containerd/containerd.sock` |

## Chaos Testing

We kept a write client active during the experiments so that availability
failures and ambiguous writes were visible instead of being hidden by an idle
database.

### ClickHouse High-Write Load Client

ClickHouse is a column-oriented analytical database with a SQL interface. We
use database queries to create tables, insert test data, count unique IDs, and
inspect replication state.

The table used `ReplicatedMergeTree`, while clients wrote through a
`Distributed` table across `chaos-cluster`. Each batch inserted 100 rows
with server-generated UUIDs and `insert_distributed_sync=1`.

A UUID is a randomly generated unique identifier assigned to each row. Unique
IDs let us detect accidental duplicates by checking whether the total row
count equals the number of unique IDs.

Create the schema from one data pod. The operator already injects
`CLICKHOUSE_USER` and `CLICKHOUSE_PASSWORD` into the ClickHouse container, so
no credential needs to be copied into the command:

```bash
kubectl exec -i -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c '
    clickhouse-client \
      --user "$CLICKHOUSE_USER" \
      --password "$CLICKHOUSE_PASSWORD" \
      --multiquery
  ' <<'SQL'
CREATE DATABASE IF NOT EXISTS chaos_v2 ON CLUSTER `chaos-cluster`;

CREATE TABLE IF NOT EXISTS chaos_v2.events_local
ON CLUSTER `chaos-cluster`
(
  id UUID,
  inserted_at DateTime64(3),
  payload UInt64
)
ENGINE = ReplicatedMergeTree(
  '/clickhouse/{installation}/{cluster}/tables/{shard}/{database}/{table}',
  '{replica}'
)
ORDER BY id;

CREATE TABLE IF NOT EXISTS chaos_v2.events
ON CLUSTER `chaos-cluster`
AS chaos_v2.events_local
ENGINE = Distributed(
  'chaos-cluster',
  'chaos_v2',
  'events_local',
  cityHash64(id)
);
SQL
```

The command exited successfully; ClickHouse created the
database and both tables on the cluster.

#### Create `setup/clickhouse-workload.yaml`

The workload client inserts 100 rows, waits for synchronous Distributed
delivery, records whether the batch was acknowledged, then repeats.

Save the following manifest as `setup/clickhouse-workload.yaml`:

The `while true` below belongs inside the workload container's script. It is
not a loop that the reader runs manually; Kubernetes starts this script once
and it keeps producing traffic until the workload is paused.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: clickhouse-chaos-workload
  namespace: demo
  labels:
    chaos-test.kubedb.com/suite: clickhouse-chaos
data:
  run.sh: |
    #!/usr/bin/env bash
    set -u
    mkdir -p /state
    test -f /state/attempt_batches || printf '0\n' > /state/attempt_batches
    test -f /state/success_batches || printf '0\n' > /state/success_batches
    test -f /state/failed_batches || printf '0\n' > /state/failed_batches

    while true; do
      if test -f /state/pause; then
        sleep 1
        continue
      fi

      attempt=$(( $(cat /state/attempt_batches) + 1 ))
      printf '%s\n' "$attempt" > /state/attempt_batches

      if timeout --signal=TERM --kill-after=5s 20s clickhouse-client \
        --host clickhouse-chaos.demo.svc \
        --user "$CH_USER" \
        --password "$CH_PASSWORD" \
        --connect_timeout 5 \
        --send_timeout 10 \
        --receive_timeout 10 \
        --query "INSERT INTO chaos_v2.events
          SELECT generateUUIDv4(), now64(3), rand64()
          FROM numbers(100)
          SETTINGS insert_distributed_sync=1"; then
        value=$(( $(cat /state/success_batches) + 1 ))
        printf '%s\n' "$value" > /state/success_batches
        printf '%s success attempt=%s rows=100\n' \
          "$(date -Iseconds)" "$attempt"
      else
        value=$(( $(cat /state/failed_batches) + 1 ))
        printf '%s\n' "$value" > /state/failed_batches
        printf '%s failure attempt=%s\n' \
          "$(date -Iseconds)" "$attempt" >&2
      fi
      sleep 1
    done
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: clickhouse-chaos-workload
  namespace: demo
  labels:
    chaos-test.kubedb.com/suite: clickhouse-chaos
spec:
  replicas: 1
  selector:
    matchLabels:
      app: clickhouse-chaos-workload
  template:
    metadata:
      labels:
        app: clickhouse-chaos-workload
        chaos-test.kubedb.com/suite: clickhouse-chaos
    spec:
      containers:
        - name: workload
          image: docker.io/clickhouse/clickhouse-server:26.2.6
          command:
            - /bin/bash
            - /scripts/run.sh
          env:
            - name: CH_USER
              valueFrom:
                secretKeyRef:
                  name: clickhouse-chaos-auth
                  key: username
            - name: CH_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: clickhouse-chaos-auth
                  key: password
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 250m
              memory: 256Mi
          volumeMounts:
            - name: script
              mountPath: /scripts
              readOnly: true
            - name: state
              mountPath: /state
      volumes:
        - name: script
          configMap:
            name: clickhouse-chaos-workload
            defaultMode: 365
        - name: state
          emptyDir: {}
```

Start the client and wait for at least ten successful batches:

```bash
➤ kubectl apply -f setup/clickhouse-workload.yaml
configmap/clickhouse-chaos-workload configured
deployment.apps/clickhouse-chaos-workload created

```

```bash
➤ kubectl rollout status -n demo \
        deployment/clickhouse-chaos-workload --timeout=3m
deployment "clickhouse-chaos-workload" successfully rolled out
```


```bash
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
    -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-sgzlc
```

```bash
$ kubectl logs -n demo -f clickhouse-chaos-workload-64d7d5c85f-sgzlc
2026-09-08T04:44:40+00:00 success attempt=1 rows=100
2026-09-08T04:44:41+00:00 success attempt=2 rows=100
2026-09-08T04:44:43+00:00 success attempt=3 rows=100
2026-09-08T04:44:44+00:00 success attempt=4 rows=100
2026-09-08T04:44:45+00:00 success attempt=5 rows=100
2026-09-08T04:44:46+00:00 success attempt=6 rows=100
2026-09-08T04:44:47+00:00 success attempt=7 rows=100
2026-09-08T04:44:48+00:00 success attempt=8 rows=100
2026-09-08T04:44:49+00:00 success attempt=9 rows=100
2026-09-08T04:44:51+00:00 success attempt=10 rows=100
```

When the log shows at least ten `success` lines, press `Ctrl-C`. This stops
only log streaming; the workload Deployment continues writing in the cluster.

The client marked a batch as acknowledged only when ClickHouse returned a
success response. If the client timed out, the outcome was ambiguous. For
example, ClickHouse could accept the rows but the response could arrive after
the client stopped waiting. Those rows remain in the database even though the
client recorded the attempt as failed. Therefore, `acknowledged batches ×
100` is the minimum number of rows known to have been accepted, not
necessarily the final row count.

Before the first test, the client had 35 successful batches, zero failures,
and 3,500 unique rows. At the end it had:

```text
attempted batches:     1,342
acknowledged batches:  1,208
failed/ambiguous:      134
acknowledged rows:     120,800
actual rows:           122,695
unique IDs:            122,695
```

The client attempted 1,342 batches. It received success for 1,208 batches, so
`1,208 × 100 = 120,800` rows were definitely acknowledged. Another 134 attempts
timed out or returned an error. Some rows from those attempts had already
reached ClickHouse.
Together they account for the additional 1,895 rows. Since all 122,695 IDs
were unique, these additional rows were neither duplicate rows nor evidence
of corruption.

### Mandatory Recovery Gate

After every experiment, we deleted the Chaos Mesh object, paused writes, and
required all of the
following before continuing:

1. Pause the workload and confirm its active client stopped.
2. Require KubeDB `Ready` and all four ClickHouse plus three Keeper pods Ready.
3. Confirm that no test Chaos Mesh object remains.
4. Perform an authenticated probe insert and require `count() == uniqExact(id)`
   on the Distributed table.
5. Require matching local counts and checksums for each shard in two
   consecutive checks five seconds apart. A check is a set of query results
   collected at one moment, not a volume snapshot or backup. Repeating it
   avoids judging a brief part-visibility delay immediately after Keeper
   returns.
6. Require `system.replicas` to show writable replicas, queue size zero, two
   total replicas, and two active replicas.
7. Require all three Keepers to answer `mntr`, with one leader and two
   followers.
8. Require running ClickHouse processes and normal data mounts, with no stale
   `toda` FUSE layer.

Pause the workload and let the active client finish:

```bash
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
    -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-sgzlc
```

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- touch /state/pause
```

```bash
$ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c   'if pgrep -x clickhouse-client >/dev/null; then echo "client still active"; else echo "workload paused"; fi'
workload paused
```

Require KubeDB and all seven database pods to be healthy:

```bash
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=10m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

```bash
➤ kubectl get clickhouse,petset,pods -n demo
NAME                                     VERSION   STATUS   AGE
clickhouse.kubedb.com/clickhouse-chaos   26.2.6    Ready    28m

NAME                                                                  AGE
petset.apps.k8s.appscode.com/clickhouse-chaos-chaos-cluster-shard-0   28m
petset.apps.k8s.appscode.com/clickhouse-chaos-chaos-cluster-shard-1   28m
petset.apps.k8s.appscode.com/clickhouse-chaos-keeper                  28m

NAME                                             READY   STATUS    RESTARTS   AGE
pod/clickhouse-chaos-chaos-cluster-shard-0-0     1/1     Running   0          28m
pod/clickhouse-chaos-chaos-cluster-shard-0-1     1/1     Running   0          28m
pod/clickhouse-chaos-chaos-cluster-shard-1-0     1/1     Running   0          28m
pod/clickhouse-chaos-chaos-cluster-shard-1-1     1/1     Running   0          28m
pod/clickhouse-chaos-keeper-0                    1/1     Running   0          28m
pod/clickhouse-chaos-keeper-1                    1/1     Running   0          28m
pod/clickhouse-chaos-keeper-2                    1/1     Running   0          28m
pod/clickhouse-chaos-workload-64d7d5c85f-sgzlc   1/1     Running   0          6m39s

```

Confirm that no test fault remains:

```bash
➤ kubectl get podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos -n demo
No resources found in demo namespace.
```


Check the Distributed table. The first two values must match, and the count
must be at least `successful_batches × 100`:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
  clickhouse-client \
    --host clickhouse-chaos.demo.svc \
    --user "$CH_USER" \
    --password "$CH_PASSWORD" \
    --query "INSERT INTO chaos_v2.events
      SELECT generateUUIDv4(), now64(3), rand64()
      SETTINGS insert_distributed_sync=1"
'
```


```bash
$ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
  clickhouse-client \
    --host clickhouse-chaos.demo.svc \
    --user "$CH_USER" \
    --password "$CH_PASSWORD" \
    --query "SELECT count(), uniqExact(id), sum(payload)
             FROM chaos_v2.events FORMAT TSV"
'
4201	4201	18251318426044052401
```

Check every local replica:

```bash
➤ kubectl exec -n demo \
        clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
    clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
      --query "SELECT count(), uniqExact(id), sum(payload)
               FROM chaos_v2.events_local FORMAT TSV"
  '
2100	2100	15819459328062010837

➤ kubectl exec -n demo \
        clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
    clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
      --query "SELECT count(), uniqExact(id), sum(payload)
               FROM chaos_v2.events_local FORMAT TSV"
  '
2100	2100	15819459328062010837

➤ kubectl exec -n demo \
        clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c '
    clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
      --query "SELECT count(), uniqExact(id), sum(payload)
               FROM chaos_v2.events_local FORMAT TSV"
  '
2101	2101	2431859097982041564

➤ kubectl exec -n demo \
        clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- bash -c '
    clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
      --query "SELECT count(), uniqExact(id), sum(payload)
               FROM chaos_v2.events_local FORMAT TSV"
  '
2101	2101	2431859097982041564
```


Shard-0's two lines match, and shard-1's two lines match. Wait five seconds
and run the same four commands again. The second check must return the same
four lines before you continue.

Now check `system.replicas` on each pod, one at a time:

```bash
➤ kubectl exec -n demo \
        clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
    clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
      --database chaos_v2 \
      --query "SELECT is_readonly, queue_size, total_replicas, active_replicas,
                      lost_part_count, absolute_delay
               FROM system.replicas WHERE database=currentDatabase() FORMAT TSV"
  '
0	0	2	2	0	0

➤ kubectl exec -n demo \
        clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
    clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
      --database chaos_v2 \
      --query "SELECT is_readonly, queue_size, total_replicas, active_replicas,
                      lost_part_count, absolute_delay
               FROM system.replicas WHERE database=currentDatabase() FORMAT TSV"
  '
0	0	2	2	0	0

➤ kubectl exec -n demo \
        clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c '
    clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
      --database chaos_v2 \
      --query "SELECT is_readonly, queue_size, total_replicas, active_replicas,
                      lost_part_count, absolute_delay
               FROM system.replicas WHERE database=currentDatabase() FORMAT TSV"
  '
0	0	2	2	0	0

➤ kubectl exec -n demo \
        clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- bash -c '
    clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
      --database chaos_v2 \
      --query "SELECT is_readonly, queue_size, total_replicas, active_replicas,
                      lost_part_count, absolute_delay
               FROM system.replicas WHERE database=currentDatabase() FORMAT TSV"
  '
0	0	2	2	0	0

```

The values mean writable, empty queue, two configured replicas, two active
replicas, no lost parts, and no replication delay.

Check Keeper directly rather than inferring quorum from ClickHouse status:

```bash
➤ kubectl exec -n demo clickhouse-chaos-keeper-0 \
        -c clickhouse-keeper -- bash -c '
    exec 3<>/dev/tcp/127.0.0.1/9181
    printf "mntr\n" >&3
    timeout 3 cat <&3
  ' | awk '$1=="zk_server_state" {print $2}'
follower

➤ kubectl exec -n demo clickhouse-chaos-keeper-1 \
        -c clickhouse-keeper -- bash -c '
    exec 3<>/dev/tcp/127.0.0.1/9181
    printf "mntr\n" >&3
    timeout 3 cat <&3
  ' | awk '$1=="zk_server_state" {print $2}'
follower

➤ kubectl exec -n demo clickhouse-chaos-keeper-2 \
        -c clickhouse-keeper -- bash -c '
    exec 3<>/dev/tcp/127.0.0.1/9181
    printf "mntr\n" >&3
    timeout 3 cat <&3
  ' | awk '$1=="zk_server_state" {print $2}'
leader
```

The leader can change, but the result must contain exactly one leader and two
followers. Finally, verify each ClickHouse PID 1 one at a time:

```bash
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
        -c clickhouse -- ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
      
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 \
        -c clickhouse -- ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
      
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
        -c clickhouse -- ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
      
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 \
        -c clickhouse -- ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

Check each data mount separately:

```bash
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
        -c clickhouse -- mount | grep /var/lib/clickhouse
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)

➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 \
        -c clickhouse -- mount | grep /var/lib/clickhouse
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)

➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
        -c clickhouse -- mount | grep /var/lib/clickhouse
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)

➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 \
        -c clickhouse -- mount | grep /var/lib/clickhouse
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

PID 1 must not contain state `T`, and data must be on the normal filesystem,
not a `toda` FUSE mount. These checks matter because `Ready` alone cannot
prove Keeper availability, replica equality, or complete Chaos Mesh cleanup.

### Running Each Chaos Experiment

For every test, we used the same safe sequence:

```bash
kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

```text
clickhouse-chaos-workload-64d7d5c85f-sgzlc
```

Start the workload using the pod name returned above:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- rm -f /state/pause
```

The command prints nothing on success. Validate the manifest against the
API server:

```bash
kubectl apply --dry-run=server -f tests/01-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 created (server dry run)
```

Inject the fault:

```bash
kubectl apply -f tests/01-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 created
```

Prove that Chaos Mesh injected it:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/01-pod-kill.yaml --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 condition met
```

Observe ClickHouse during the fault:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```

```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Critical
```

Read the workload counters:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
  printf "attempts="; cat /state/attempt_batches
  printf "success="; cat /state/success_batches
  printf "failed="; cat /state/failed_batches
'
```

Output from test 1 before its recovery gate:

```text
attempts=173
success=173
failed=0
```

Remove the experiment:

```bash
kubectl delete -f tests/01-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-01" deleted from demo namespace
```

Deletion only removes the injected fault; it does not prove that ClickHouse
recovered. After deletion, run every command in the **Mandatory Recovery Gate**
section immediately above. In this guide, “run the full gate” means:

1. Pause the workload and confirm its active client has stopped.
2. Wait for KubeDB `Ready` and check all seven database pods.
3. Confirm that no Chaos Mesh resource from the test remains.
4. Perform the one-row probe insert and query the Distributed table.
5. Check every local replica twice, five seconds apart.
6. Check `system.replicas` on all four data pods.
7. Check all three Keeper roles.
8. Check PID 1 and the data mount on all four ClickHouse pods.

All eight checks must pass before unpausing the workload for the next
experiment.

The commands above use test 1 as the example. For another experiment, replace
`tests/01-pod-kill.yaml` with the filename printed at the start of that
experiment. Run the commands in order; do not start the next fault until the
mandatory recovery gate passes.

For a timed experiment, wait for its duration to expire before deletion. For
example, test 2 uses:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  -f tests/02-pod-failure.yaml --timeout=150s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-02 condition met
```

```bash
kubectl delete -f tests/02-pod-failure.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-02" deleted from demo namespace
```

For one-shot `pod-kill` and `container-kill` tests, delete the Chaos object
after `AllInjected` and after proving the replacement pod or container is
healthy; those tests do not have a duration to wait for.

The `AllInjected` check matters. For example, our first DNS test used a
pattern that did not match the full name queried by the resolver. Treating
manifest creation as proof of injection would have produced a false pass.

## Pod and Keeper Chaos

### Chaos#1: Kill One ClickHouse Replica

#### Create `tests/01-pod-kill.yaml`
Save this YAML as `tests/01-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-01
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  gracePeriod: 0
```

What this chaos does: Abruptly deletes shard-0 replica-0 with no graceful
shutdown.

**Expected behavior:** The sibling replica should keep the shard available,
PetSet should create a replacement pod, and replication should catch it up.
The target pod UID must change, but acknowledged data must not.

Record the target UID before applying the manifest, then confirm that it
changes after injection.

#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    67m
```

Record the original pod UID:

```bash
➤ kubectl get pod -n demo \
        clickhouse-chaos-chaos-cluster-shard-0-0 \
        -o jsonpath='{.metadata.uid}{"\n"}'
9feb08ef-57ff-4256-8e29-ca14b9769000
```


Apply this experiment:

```bash
➤ kubectl apply -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 created

```

Confirm that Chaos Mesh reached the target:

```bash
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-01 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 condition met
```

Observe the live impact:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Critical
```

Wait for PetSet to make the replacement pod ready:

```bash
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-0-0 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-0-0 condition met
```

Confirm that the replacement has a new UID:

```bash
kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```

```text
0421be3e-436b-492d-ab6d-c2998debb5fc
```

Delete the experiment:

```bash
kubectl delete -f tests/01-pod-kill.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-01" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

The target pod UID changed from `0f32c2fb-1869-4521-ad03-8aead8f55a20` to `0421be3e-436b-492d-ab6d-c2998debb5fc`. KubeDB may briefly report `Critical`, but the workload advanced from 35 to 59 acknowledged batches with no failures. After cleanup, the replica was writable with an empty queue and two active replicas.

Result: **PASS** — the sibling kept the shard available and the replacement converged automatically.

### Chaos#2: Hold One Replica Failed

#### Create `tests/02-pod-failure.yaml`
Save this YAML as `tests/02-pod-failure.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-02
  namespace: demo
spec:
  action: pod-failure
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-1
```

What this chaos does: Makes shard-0 replica-1 continuously unavailable for
45 seconds instead of allowing Kubernetes to replace it immediately.

**Expected behavior:** KubeDB should report a degraded state while the sibling
replica continues serving the shard. When the fault ends, the same pod should
become reachable and converge without manual repair.

Before injecting the fault, the fresh cluster was healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```

```text
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    4m
```

```bash
kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1
```

```text
NAME                                             READY   STATUS    RESTARTS   AGE
clickhouse-chaos-chaos-cluster-shard-0-1         1/1     Running   0          4m
```

Apply the file and confirm that Chaos Mesh really injected the failure. A
created object alone is not evidence that the fault reached the target:

```bash
kubectl apply -f tests/02-pod-failure.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-02 created
```

```bash
kubectl get podchaos -n demo clickhouse-chaos-exp-02 \
  -o jsonpath='{range .status.conditions[*]}{.type}={.status}{"\n"}{end}'
```

```text
Selected=True
AllInjected=True
AllRecovered=False
Paused=False
```

While the 45-second fault was active, the target pod stayed present but its
container restarted. KubeDB was still `Ready` at this sample because the
healthy sibling could serve the shard:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```

```text
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    5m
```

```bash
kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1
```

```text
NAME                                             READY   STATUS    RESTARTS
clickhouse-chaos-chaos-cluster-shard-0-1         1/1     Running   1
```

After the duration elapsed, Chaos Mesh reported recovery. Deleting the
experiment can briefly leave KubeDB `Critical` while the replica reconnects;
wait for the database condition, rather than treating that short transition
as data loss:

```bash
kubectl get podchaos -n demo clickhouse-chaos-exp-02 \
  -o jsonpath='{range .status.conditions[*]}{.type}={.status}{"\n"}{end}'
```

```text
AllRecovered=True
```

```bash
kubectl delete -f tests/02-pod-failure.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-02" deleted from demo namespace
```

```bash
kubectl wait --for=condition=Ready clickhouse/clickhouse-chaos \
  -n demo --timeout=5m
```

```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```

```text
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    6m
```

```bash
kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT is_readonly, queue_size, total_replicas, active_replicas
           FROM system.replicas
           WHERE database='\''chaos_v2'\'' AND table='\''events_local'\''
           FORMAT TSV"'
```

```text
0  0  2  2
```

**Observed behavior:**

The 45-second failure restarted the target twice. KubeDB was initially `Ready`, then became `Critical` while the replica reconnected. The workload moved from 98 successful/0 failed to 138 successful/13 failed or ambiguous attempts. `AllRecovered=True` was not treated as complete database recovery; the test waited until KubeDB returned to `Ready`.

Result: **PASS** — the sustained replica failure was visible and the replica healed without manual repair.

### Chaos#3: Kill Only the ClickHouse Container

#### Create `tests/03-container-kill.yaml`
Save this YAML as `tests/03-container-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-03
  namespace: demo
spec:
  action: container-kill
  mode: one
  containerNames:
    - clickhouse
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-1-0
```

What this chaos does: Kills only the `clickhouse` process container while
leaving the pod object and its PVC in place.

**Expected behavior:** Kubernetes should restart the container, ClickHouse
should reconnect to Keeper and replication, and the restart count should
increase without a pod UID change.

Compare the `clickhouse` container restart count before and after injection.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/03-container-kill.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-03 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-03 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-03 condition met
```

Observe the live impact:

```bash
kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
  -o jsonpath='{.metadata.uid}{"\n"}{.status.containerStatuses[0].restartCount}{"\n"}'
```
```text
7b2c478c-f034-464e-bfce-6040f068a6ba
1
```

Delete the experiment:

```bash
kubectl delete -f tests/03-container-kill.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-03" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

The pod UID remained `7b2c478c-f034-464e-bfce-6040f068a6ba`, while its restart count changed from 0 to 1. KubeDB briefly reported `Critical`; 16 further batches were acknowledged and no new failure was recorded.

Result: **PASS** — Kubernetes restarted only the ClickHouse container and it rejoined replication.

### Chaos#4: Repeat Alternating Pod Kills

#### Create `tests/04-a-pod-kill.yaml`
Save the first fault as `tests/04-a-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-04-a
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  gracePeriod: 0
```

#### Create `tests/04-b-pod-kill.yaml`
Save the second fault as `tests/04-b-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-04-b
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-1-1
  gracePeriod: 0
```

#### Create `tests/04-c-pod-kill.yaml`
Save the third fault as `tests/04-c-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-04-c
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-1
  gracePeriod: 0
```

What this chaos does: Kills three different data replicas one at a time,
with 20 seconds between injections.

**Expected behavior:** Every killed pod should be replaced before the next
fault is injected. Replication queues, restarts, and checksums should return to
baseline after the sequence.

Apply the files one at a time. Wait 20 seconds between kills, delete each
one-shot `PodChaos` after `AllInjected`, and run the complete recovery gate
after the third kill.

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- rm -f /state/pause
```

The command prints nothing on success.

```bash
kubectl apply -f tests/04-a-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-a created
```

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/04-a-pod-kill.yaml --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-a condition met
```

```bash
kubectl delete -f tests/04-a-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-04-a" deleted from demo namespace
```

```bash
kubectl wait -n demo --for=create \
  pod/clickhouse-chaos-chaos-cluster-shard-0-0 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-0-0 condition met
```

```bash
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-0-0 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-0-0 condition met
```

```bash
sleep 20
```

Output: none.

```bash
kubectl apply -f tests/04-b-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-b created
```

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/04-b-pod-kill.yaml --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-b condition met
```

```bash
kubectl delete -f tests/04-b-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-04-b" deleted from demo namespace
```

```bash
kubectl wait -n demo --for=create \
  pod/clickhouse-chaos-chaos-cluster-shard-1-1 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-1-1 condition met
```

```bash
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-1-1 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-1-1 condition met
```

```bash
sleep 20
```

Output: none.

```bash
kubectl apply -f tests/04-c-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-c created
```

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/04-c-pod-kill.yaml --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-c condition met
```

```bash
kubectl delete -f tests/04-c-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-04-c" deleted from demo namespace
```

```bash
kubectl wait -n demo --for=create \
  pod/clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

```bash
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

**Observed behavior:**

Shard-0 replica-0, shard-1 replica-1, and shard-0 replica-1 each received a new UID. The full gate passed before each following kill. Across the sequence, 136 batches were acknowledged and one attempt became failed or ambiguous; both replica pairs ended with matching counts and checksums.

Result: **PASS** — three sequential recoveries did not accumulate replica drift.

### Chaos#5: Lose an Entire Shard

#### Create `tests/05-full-shard-outage.yaml`
Save this YAML as `tests/05-full-shard-outage.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-05
  namespace: demo
spec:
  action: pod-failure
  mode: all
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
        - clickhouse-chaos-chaos-cluster-shard-0-1
```

What this chaos does: Holds both replicas of shard 0 unavailable for 45
seconds.

**Expected behavior:** Distributed inserts that require shard 0 should fail
clearly; the remaining shard cannot substitute for missing shard data.
KubeDB should report `Critical`, then both replicas should return with equal
data.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/05-full-shard-outage.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-05 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-05 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-05 condition met
```

Observe the live impact:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    NotReady
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  podchaos/clickhouse-chaos-exp-05 --timeout=2m
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-05 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/05-full-shard-outage.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-05" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Both shard-0 replicas were unavailable. A Distributed query from shard 1 returned `ALL_CONNECTION_TRIES_FAILED`, KubeDB progressed to `NotReady`, and 42 attempts failed or became ambiguous. After `AllRecovered`, the test still waited for KubeDB `Ready`; the shard replicas then matched.

Result: **PASS** — loss of a complete shard caused an explicit outage and recovered without silent inconsistency.

### Chaos#6: Lose the Entire ClickHouse Data Plane

#### Create `tests/06-data-plane-outage.yaml`
Save this YAML as `tests/06-data-plane-outage.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-06
  namespace: demo
spec:
  action: pod-failure
  mode: all
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
        - clickhouse-chaos-chaos-cluster-shard-0-1
        - clickhouse-chaos-chaos-cluster-shard-1-0
        - clickhouse-chaos-chaos-cluster-shard-1-1
```

What this chaos does: Makes all four ClickHouse data pods unavailable
while leaving the three Keeper members running.

**Expected behavior:** SQL clients should see a complete outage and no batch
should be acknowledged during it. When the fault ends, all four replicas
should reopen their existing PVC data and converge automatically.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/06-data-plane-outage.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-06 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-06 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-06 condition met
```

Observe the live impact:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Critical
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  podchaos/clickhouse-chaos-exp-06 --timeout=2m
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-06 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/06-data-plane-outage.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-06" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

All four data containers were failed while Keeper stayed online. KubeDB reported `Critical` and then `NotReady`; the measured window added 42 failed or ambiguous attempts. All four pods reopened their existing PVCs and the cluster returned to `Ready`.

Result: **PASS** — the full data-plane outage was recoverable and did not lose acknowledged rows.

### Chaos#7: Kill a Keeper Follower

Discover the Keeper roles immediately before the test.

#### Create `tests/07-keeper-follower-kill.yaml`
Save this YAML as `tests/07-keeper-follower-kill.yaml`, using a current
follower as the target:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-07
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-keeper-1
  gracePeriod: 0
```

What this chaos does: Discovers the Keeper roles with `mntr`, then abruptly
kills one current follower.

**Expected behavior:** The leader and remaining follower still form a
two-member majority, so coordination and writes should continue. The recreated
member should rejoin as a follower.

The example uses `keeper-1`; replace it if `mntr` reports that member as the
leader.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/07-keeper-follower-kill.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-07 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-07 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-07 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- \
  bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | \
  awk '$1=="zk_server_state" {print $2}'
```
```text
leader
```

Delete the experiment:

```bash
kubectl delete -f tests/07-keeper-follower-kill.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-07" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Keeper-1, a follower, was killed and received a new pod UID. Keeper-0 remained leader, KubeDB stayed `Ready`, and the workload added 39 acknowledged batches without a new error.

Result: **PASS** — the remaining two Keeper members retained quorum.

### Chaos#8: Kill the Keeper Leader

Rediscover the Keeper roles immediately before the test.

#### Create `tests/08-keeper-leader-kill.yaml`
Save this YAML as `tests/08-keeper-leader-kill.yaml`, using the current leader
as the target:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-08
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-keeper-0
  gracePeriod: 0
```

What this chaos does: Discovers and kills the current Keeper leader.

**Expected behavior:** The two surviving members should elect a new leader
quickly. ClickHouse should tolerate the short election, and the old leader
should return as a follower rather than forming a second leader.

The example uses `keeper-0`; replace it with the actual leader. Time how long
another member takes to report `leader`.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/08-keeper-leader-kill.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-08 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-08 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-08 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-keeper-2 -c clickhouse-keeper -- \
  bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | \
  awk '$1=="zk_server_state" {print $2}'
```
```text
leader
```

Delete the experiment:

```bash
kubectl delete -f tests/08-keeper-leader-kill.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-08" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Keeper-0 was the leader before injection. After it was killed, Keeper-2 reported `leader`, Keeper-1 remained a follower, and the replacement Keeper-0 rejoined as a follower. KubeDB stayed `Ready`.

Result: **PASS** — Keeper elected a new leader automatically.

### Chaos#9: Lose Keeper Quorum

#### Create `tests/09-keeper-quorum-loss.yaml`
Save this YAML as `tests/09-keeper-quorum-loss.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-09
  namespace: demo
spec:
  action: pod-failure
  mode: all
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-keeper-0
        - clickhouse-chaos-keeper-1
```

What this chaos does: Holds two of the three Keeper members failed for 45
seconds, removing the majority required for coordination.

**Expected behavior:** Existing reads may continue, but coordination-dependent
writes or replication can stall or fail. After quorum returns, queued work
should settle and both replicas of every shard should converge without data
repair.

Do not repair a brief replica mismatch while queues are still moving. Require
two consecutive equal checks within ten minutes.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/09-keeper-quorum-loss.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-09 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-09 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-09 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-keeper-2 -c clickhouse-keeper -- \
  bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3'
```
```text
This instance is not currently serving requests
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  podchaos/clickhouse-chaos-exp-09 --timeout=2m
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-09 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/09-keeper-quorum-loss.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-09" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Keeper-0 and Keeper-1 were failed together. The survivor returned `This instance is not currently serving requests`, which proved that a `leader` label alone would not establish quorum. KubeDB still showed `Ready`; four batches succeeded and two attempts failed before quorum returned.

Result: **PASS** — Keeper quorum reformed and all replica queues drained.

### Chaos#10: Fail All Keeper Members

#### Create `tests/10-full-keeper-outage.yaml`
Save this YAML as `tests/10-full-keeper-outage.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-10
  namespace: demo
spec:
  action: pod-failure
  mode: all
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-keeper-0
        - clickhouse-chaos-keeper-1
        - clickhouse-chaos-keeper-2
```

What this chaos does: Holds all three Keeper members unavailable for 45
seconds.

**Expected behavior:** ClickHouse processes can remain reachable, but Keeper
operations should be unavailable and replicated writes may fail. Recovery
requires a new one-leader/two-follower quorum and drained replica queues.

During injection, check `mntr` directly even if KubeDB still reports `Ready`.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/10-full-keeper-outage.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-10 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-10 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-10 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- \
  bash -c 'printf "mntr\n"'
```
```text
OCI runtime exec failed: exec: "bash": executable file not found in $PATH
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  podchaos/clickhouse-chaos-exp-10 --timeout=2m
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-10 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/10-full-keeper-outage.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-10" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

All three Keeper containers were failed. An exec attempt returned an OCI error because Chaos Mesh had replaced the container entrypoint. ClickHouse data processes remained present, 13 batches succeeded around existing sessions, and two attempts failed. Keeper returned with one leader and two followers.

Result: **PASS** — a complete coordination outage recovered automatically.

## Network Chaos

### Chaos#11: Add Network Delay

#### Create `tests/11-network-delay.yaml`
Save this YAML as `tests/11-network-delay.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-exp-11
  namespace: demo
spec:
  action: delay
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  direction: to
  delay:
    latency: 500ms
    jitter: 50ms
    correlation: "50"
```

What this chaos does: Adds 500 ms latency with 50 ms jitter to traffic
reaching one data replica for 45 seconds.

**Expected behavior:** Queries may slow down, but TCP and the healthy replica
should keep the tested workload available. The target must finish with no
replication delay or queued work.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/11-network-delay.yaml
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-11 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  networkchaos/clickhouse-chaos-exp-11 --timeout=90s
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-11 condition met
```

Observe the live impact:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  networkchaos/clickhouse-chaos-exp-11 --timeout=2m
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-11 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/11-network-delay.yaml
```
```text
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-11" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

The target received 500ms inbound delay with 50ms jitter. KubeDB remained `Ready`; 27 batches were acknowledged and no new failure appeared during the measured window. The target finished writable with an empty queue.

Result: **PASS** — the healthy sibling and TCP retries absorbed the delay.

### Chaos#12: Drop 30 Percent of Packets

#### Create `tests/12-network-loss.yaml`
Save this YAML as `tests/12-network-loss.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-exp-12
  namespace: demo
spec:
  action: loss
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  direction: to
  loss:
    loss: "30"
    correlation: "25"
```

What this chaos does: Drops 30 percent of packets sent to one ClickHouse
replica.

**Expected behavior:** TCP retransmission should absorb some loss. Timeouts
are acceptable, but after the fault the replica must be writable, caught up,
and byte-for-byte equivalent at the logical checksum level.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/12-network-loss.yaml
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-12 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  networkchaos/clickhouse-chaos-exp-12 --timeout=90s
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-12 condition met
```

Observe the live impact:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  networkchaos/clickhouse-chaos-exp-12 --timeout=2m
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-12 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/12-network-loss.yaml
```
```text
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-12" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Thirty percent packet loss was injected into one replica. KubeDB remained `Ready`; 39 batches were acknowledged without a new client failure, and the replica converged after recovery.

Result: **PASS** — packet loss caused no lasting replica damage.

### Chaos#13: Duplicate 50 Percent of Packets

#### Create `tests/13-network-duplicate.yaml`
Save this YAML as `tests/13-network-duplicate.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-exp-13
  namespace: demo
spec:
  action: duplicate
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  direction: to
  duplicate:
    duplicate: "50"
    correlation: "25"
```

What this chaos does: Duplicates 50 percent of packets reaching one data
replica.

**Expected behavior:** TCP and ClickHouse should not turn duplicated network
packets into duplicated table rows. `count()` must still equal
`uniqExact(id)` after recovery.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/13-network-duplicate.yaml
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-13 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  networkchaos/clickhouse-chaos-exp-13 --timeout=90s
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-13 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id) FROM chaos_v2.events FORMAT TSV"'
```
```text
59500  59500
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  networkchaos/clickhouse-chaos-exp-13 --timeout=2m
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-13 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/13-network-duplicate.yaml
```
```text
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-13" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Fifty percent of inbound packets were duplicated. The workload continued, and the post-fault query returned 59,500 total rows and 59,500 unique IDs.

Result: **PASS** — network duplication did not create duplicate table rows.

### Chaos#14: Limit Bandwidth to 1 Mbps

#### Create `tests/14-bandwidth.yaml`
Save this YAML as `tests/14-bandwidth.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-exp-14
  namespace: demo
spec:
  action: bandwidth
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  direction: to
  bandwidth:
    rate: 1mbps
    limit: 20971520
    buffer: 10000
```

What this chaos does: Restricts inbound traffic to one replica to 1 Mbps
for 45 seconds.

**Expected behavior:** Throughput and latency may degrade. The client may
timeout if demand exceeds the cap, but replication should drain completely
after normal bandwidth returns.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/14-bandwidth.yaml
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-14 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  networkchaos/clickhouse-chaos-exp-14 --timeout=90s
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-14 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
printf "successful="; cat /state/success_batches
printf "failed="; cat /state/failed_batches'
```
```text
successful=645
failed=103
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  networkchaos/clickhouse-chaos-exp-14 --timeout=2m
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-14 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/14-bandwidth.yaml
```
```text
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-14" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Inbound bandwidth to one replica was limited to 1Mbps. The small 100-row workload continued from 590 to 645 acknowledged batches without increasing the failure counter, and the replication queue was empty afterward.

Result: **PASS** — the workload fit within the constrained link and recovered cleanly.

### Chaos#15: Partition One Replica from Data Peers

#### Create `tests/15-data-partition.yaml`
Save this YAML as `tests/15-data-partition.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-exp-15
  namespace: demo
spec:
  action: partition
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  direction: both
  target:
    mode: all
    selector:
      namespaces:
        - demo
      pods:
        demo:
          - clickhouse-chaos-chaos-cluster-shard-0-1
          - clickhouse-chaos-chaos-cluster-shard-1-0
          - clickhouse-chaos-chaos-cluster-shard-1-1
```

What this chaos does: Isolates shard-0 replica-0 in both directions from
the other three ClickHouse data pods.

**Expected behavior:** The isolated replica can fall behind while its sibling
serves the shard. After reconnection it should fetch missing parts and match
the sibling without deleting its pod or PVC.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/15-data-partition.yaml
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-15 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  networkchaos/clickhouse-chaos-exp-15 --timeout=90s
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-15 condition met
```

Observe the live impact:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  networkchaos/clickhouse-chaos-exp-15 --timeout=2m
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-15 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/15-data-partition.yaml
```
```text
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-15" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Shard-0 replica-0 was isolated from the other data pods. KubeDB stayed `Ready`, one attempt failed during the observed window, and after reconnection both shard-0 replicas returned 33,777 rows with the same checksum.

Result: **PASS** — the isolated replica fetched missing work and converged.

### Chaos#16: Partition One Replica from Keeper

#### Create `tests/16-keeper-partition.yaml`
Save this YAML as `tests/16-keeper-partition.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-exp-16
  namespace: demo
spec:
  action: partition
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  direction: both
  target:
    mode: all
    selector:
      namespaces:
        - demo
      pods:
        demo:
          - clickhouse-chaos-keeper-0
          - clickhouse-chaos-keeper-1
          - clickhouse-chaos-keeper-2
```

What this chaos does: Blocks one data replica from communicating with all
three Keeper members for 45 seconds.

**Expected behavior:** An existing Keeper session may mask a short partition.
If the session expires, replicated-table operations on that replica should
stop safely rather than accept uncoordinated state. It should become writable
again after reconnection.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/16-keeper-partition.yaml
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-16 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  networkchaos/clickhouse-chaos-exp-16 --timeout=90s
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-16 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT is_readonly, is_session_expired, queue_size FROM system.replicas \
  WHERE database='\''chaos_v2'\'' AND table='\''events_local'\'' FORMAT TSV"'
```
```text
1  1  0
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  networkchaos/clickhouse-chaos-exp-16 --timeout=2m
```
```text
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-16 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/16-keeper-partition.yaml
```
```text
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-16" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

The target lost all Keeper connectivity. After 20 seconds it reported `is_readonly=1` and `is_session_expired=1`, correctly refusing uncoordinated replicated writes. After cleanup it returned `is_readonly=0`, `queue_size=0`, and `active_replicas=2`.

Result: **PASS** — Keeper session loss failed safely and recovered automatically.

## Resource Stress

### Chaos#17: Stress CPU

#### Create `tests/17-cpu-stress.yaml`
Save this YAML as `tests/17-cpu-stress.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: clickhouse-chaos-exp-17
  namespace: demo
spec:
  mode: one
  duration: 60s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  containerNames:
    - clickhouse
  stressors:
    cpu:
      workers: 2
      load: 80
```

What this chaos does: Runs two stress workers at 80 percent CPU load in one
ClickHouse container for 60 seconds.

**Expected behavior:** Latency can increase, but Kubernetes should not restart
the pod merely because CPU is throttled. Writes and replication should recover
with no lasting queue.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/17-cpu-stress.yaml
```
```text
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-17 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  stresschaos/clickhouse-chaos-exp-17 --timeout=90s
```
```text
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-17 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  cat /sys/fs/cgroup/cpu.stat
```
```text
nr_periods 7006
nr_throttled 339
throttled_usec 51868624
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  stresschaos/clickhouse-chaos-exp-17 --timeout=2m
```
```text
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-17 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/17-cpu-stress.yaml
```
```text
stresschaos.chaos-mesh.org "clickhouse-chaos-exp-17" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Two CPU workers at 80 percent load increased cgroup throttling to 339 periods and 51,868,624 microseconds. The target restart count remained 4 before and after the fault, KubeDB stayed `Ready`, and no new workload failure appeared.

Result: **PASS** — CPU throttling increased without restarting ClickHouse or damaging replication.

### Chaos#18: Stress Memory

#### Create `tests/18-memory-stress.yaml`
Save this YAML as `tests/18-memory-stress.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: clickhouse-chaos-exp-18
  namespace: demo
spec:
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-1-0
  containerNames:
    - clickhouse
  stressors:
    memory:
      workers: 1
      size: 1GiB
```

What this chaos does: Allocates an additional 1GiB in a ClickHouse pod whose
memory limit is 4GiB.

**Expected behavior:** This test should create strong pressure without
intentionally forcing an OOM kill. The process should remain alive and cgroup
usage should fall after cleanup. A restart or sustained near-limit usage would
fail the test.

#### Demonstrate impact and recovery

Before injection, confirm the database is healthy and record the cgroup
baseline:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c '
  printf "memory_current="; cat /sys/fs/cgroup/memory.current
  printf "memory_max="; cat /sys/fs/cgroup/memory.max
'
```

```text
memory_current=1305341952
memory_max=4294967296
```

Apply this experiment:

```bash
kubectl apply -f tests/18-memory-stress.yaml
```
```text
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-18 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  stresschaos/clickhouse-chaos-exp-18 --timeout=90s
```
```text
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-18 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/18-memory-stress.yaml
```

```text
stresschaos.chaos-mesh.org "clickhouse-chaos-exp-18" deleted from demo namespace
```

Allow memory usage to settle:

```bash
sleep 15
```

Output: none.

Measure the recovered cgroup:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c '
  printf "memory_current="; cat /sys/fs/cgroup/memory.current
  printf "memory_max="; cat /sys/fs/cgroup/memory.max
'
```

```text
memory_current=1400053760
memory_max=4294967296
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c '
printf "memory_current="; cat /sys/fs/cgroup/memory.current
printf "memory_max="; cat /sys/fs/cgroup/memory.max'
```
```text
memory_current=2417717248
memory_max=4294967296
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  stresschaos/clickhouse-chaos-exp-18 --timeout=2m
```
```text
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-18 condition met
```


**Observed behavior:**

Before injection, memory usage was 1,305,341,952 bytes against a 4,294,967,296-byte limit. A 1GiB stress worker raised usage to 2,417,717,248 bytes without an OOM or restart. Fifteen seconds after cleanup it fell to 1,400,053,760 bytes.

Result: **PASS** — the 4GiB limit provided safe headroom and memory returned toward baseline.

## IO Chaos

### Chaos#19: Add Filesystem Latency

#### Create `tests/19-io-latency.yaml`
Save this YAML as `tests/19-io-latency.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: clickhouse-chaos-exp-19
  namespace: demo
spec:
  action: latency
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  containerNames:
    - clickhouse
  volumePath: /var/lib/clickhouse
  path: /var/lib/clickhouse/**/*
  delay: 100ms
  percent: 50
```

What this chaos does: Uses IOChaos to delay 50 percent of operations below
`/var/lib/clickhouse` by 100 ms for 45 seconds.

**Expected behavior:** Queries can slow or timeout, but ClickHouse must not
lose acknowledged parts. Chaos Mesh must remove its `toda` FUSE layer and
restore the normal data mount when the experiment ends.

#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/19-io-latency.yaml
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-19 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  iochaos/clickhouse-chaos-exp-19 --timeout=90s
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-19 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  findmnt -T /var/lib/clickhouse
```
```text
TARGET              SOURCE FSTYPE OPTIONS
/var/lib/clickhouse toda   fuse   rw,nosuid,nodev,relatime,user_id=0,group_id=0
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  iochaos/clickhouse-chaos-exp-19 --timeout=2m
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-19 condition met
```

Delete the recovered experiment:

```bash
kubectl delete -f tests/19-io-latency.yaml
```

```text
iochaos.chaos-mesh.org "clickhouse-chaos-exp-19" deleted from demo namespace
```

Check that Chaos Mesh removed its FUSE layer:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  mount | grep /var/lib/clickhouse
```

```text
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

The normal mount returned, but PID 1 was still stopped:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

```text
    PID STAT COMMAND
      1 Tsl  clickhouse-serv
```

Because `T` means stopped, resume the existing process:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  kill -CONT 1
```

The signal command prints nothing. Confirm that PID 1 is running again:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

```text
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

Wait for KubeDB recovery after resuming the process:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```

```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```

**Observed behavior:**

IOChaos installed a `toda` FUSE mount and delayed half of the selected filesystem operations by 100ms. It restored the ext4 mount after `AllRecovered`, but PID 1 was `Tsl`. `kill -CONT 1` changed it to `Ssl`, after which KubeDB and replica checks passed.

Result: **PASS WITH MANUAL CLEANUP** — data and the filesystem were intact, but Chaos Mesh did not resume the process.

### Chaos#20: Return Recoverable EIO

#### Create `tests/20-io-fault.yaml`
Save this YAML as `tests/20-io-fault.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: clickhouse-chaos-exp-20
  namespace: demo
spec:
  action: fault
  mode: one
  duration: 30s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-1-0
  containerNames:
    - clickhouse
  volumePath: /var/lib/clickhouse
  path: /var/lib/clickhouse/**/*
  errno: 5
  percent: 10
```

What this chaos does: Makes 10 percent of selected data-volume operations
return errno 5 (`EIO`) for 30 seconds.

**Expected behavior:** ClickHouse should expose explicit disk errors rather
than silently accepting bad data. Some writes may fail. After injection ends,
the normal mount, writable replicas, equal checksums, and empty queues must
return.

This returns explicit errors; it does not intentionally return incorrect file
contents.

Count the relevant ClickHouse log messages after the fault:

```bash
kubectl logs -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse | \
  grep -E 'Input/output error|CANNOT_STATVFS' | wc -l
```

Output from our run:

```text
261
```


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/20-io-fault.yaml
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-20 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  iochaos/clickhouse-chaos-exp-20 --timeout=90s
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-20 condition met
```

Observe the live impact:

```bash
kubectl logs -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse | \
  grep -E 'Input/output error|CANNOT_STATVFS' | wc -l
```
```text
261
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  iochaos/clickhouse-chaos-exp-20 --timeout=2m
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-20 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/20-io-fault.yaml
```
```text
iochaos.chaos-mesh.org "clickhouse-chaos-exp-20" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Ten percent of selected filesystem operations returned errno 5. During injection, 261 `Input/output error` or `CANNOT_STATVFS` messages were counted and the workload recorded failures. After recovery, the mount was ext4 and PID 1 was `Ssl`; no signal or pod replacement was required.

Result: **PASS** — explicit storage errors stopped when the fault ended.

## DNS and Time Chaos

### Chaos#21: Return Keeper DNS Errors

#### Create `tests/21-keeper-dns-error.yaml`
Save this YAML as `tests/21-keeper-dns-error.yaml`. Use the complete
`.svc.cluster.local` names; a pattern ending at `.svc` does not match the DNS
query made by the resolver.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: DNSChaos
metadata:
  name: clickhouse-chaos-exp-21
  namespace: demo
spec:
  action: error
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  containerNames:
    - clickhouse
  patterns:
    - clickhouse-chaos-keeper-0.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
    - clickhouse-chaos-keeper-1.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
    - clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
```

What this chaos does: Returns DNS errors for the three full Keeper service
FQDNs when queried from one ClickHouse container.

**Expected behavior:** A direct lookup must fail during injection and succeed
afterward. Established Keeper TCP sessions may keep working, so uninterrupted
writes do not by themselves prove that DNSChaos failed to inject.

Prove the injection with a direct lookup. This command must fail during the
fault:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  getent hosts \
  clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
```

Output during injection:

```text
command terminated with exit code 2
```

No address was returned. After `AllRecovered`, run the same command again:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  getent hosts \
  clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
```

Output after recovery:

```text
10.42.0.232     clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
```


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/21-keeper-dns-error.yaml
```
```text
dnschaos.chaos-mesh.org/clickhouse-chaos-exp-21 created
```

Confirm that Chaos Mesh reached the target:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  dnschaos/clickhouse-chaos-exp-21 --timeout=90s
```
```text
dnschaos.chaos-mesh.org/clickhouse-chaos-exp-21 condition met
```

Observe the live impact:

```bash
kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  getent hosts clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
```
```text
command terminated with exit code 2
```

Wait for Chaos Mesh to remove the fault:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  dnschaos/clickhouse-chaos-exp-21 --timeout=2m
```
```text
dnschaos.chaos-mesh.org/clickhouse-chaos-exp-21 condition met
```

Delete the experiment:

```bash
kubectl delete -f tests/21-keeper-dns-error.yaml
```
```text
dnschaos.chaos-mesh.org "clickhouse-chaos-exp-21" deleted from demo namespace
```

Wait for KubeDB to report full recovery:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```


**Observed behavior:**

Before injection, the Keeper FQDN resolved to `10.42.0.119`. During DNSChaos, the same `getent` command returned exit code 2. KubeDB remained `Ready` because established Keeper sessions continued; after recovery, the name resolved again and the failure counter had not increased.

Result: **PASS** — the DNS fault was proved independently from cached coordination connections.

### Chaos#22: Skew the Clock Back Two Hours

#### Create `tests/22-clock-skew.yaml`
Save this YAML as `tests/22-clock-skew.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: TimeChaos
metadata:
  name: clickhouse-chaos-exp-22
  namespace: demo
spec:
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-1-1
  containerNames:
    - clickhouse
  timeOffset: -2h
  clockIds:
    - CLOCK_REALTIME
```

What this chaos does: Changes only `CLOCK_REALTIME` for one ClickHouse
process by minus two hours for 45 seconds. It simulates a node with badly
incorrect wall-clock synchronization without changing the control replica's
clock.

**Expected behavior:** The target's `now()` values and log timestamps should
move two hours backward while the control replica remains correct. ClickHouse
should continue serving data, replication should converge, and no unique ID should
be lost or duplicated. When the fault expires, Chaos Mesh should restore the
clock automatically and leave PID 1 running; no manual signal or pod restart
should be necessary.

An incorrect timezone and an incorrect clock are different conditions. A
wrong timezone normally changes only how local time is displayed; the
underlying UTC clock remains correct, and ClickHouse can continue functioning.
An actual two-hour clock error changes the time returned by `now()` and can
affect inserted timestamps, TTL processing, scheduled work, logs, certificate
validation, and other time-based behavior. This experiment changes one
ClickHouse process's real-time clock. It does not change the node timezone.

Prove the skew with a ClickHouse query that is already running when the fault
starts. A new process created by `kubectl exec date` may not share the
injected process time namespace, and a new ClickHouse query may time out while
Chaos Mesh is attaching to the server process.

In terminal 1, start this 60-second timestamp stream before applying the
TimeChaos manifest:

```bash
kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

```text
clickhouse-chaos-workload-64d7d5c85f-sgzlc
```

Start the timestamp stream using the pod name returned above:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
  clickhouse-client \
    --host clickhouse-chaos-chaos-cluster-shard-1-1.clickhouse-chaos-pods.demo.svc \
    --user "$CH_USER" \
    --password "$CH_PASSWORD" \
    --query "SELECT nowInBlock64(3), sleepEachRow(0.5)
             FROM numbers(120)
             SETTINGS max_block_size=1 FORMAT TSV"
'
```

Captured output excerpt from terminal 1:

```text
2026-09-07 08:18:30.405    0
2026-09-07 06:18:30.939    0
2026-09-07 08:20:48.061    0
```

The timestamps prove that the same running ClickHouse query moved backward by
two hours during injection and returned to the current time after recovery.
In terminal 2, validate and apply the fault:

```bash
kubectl apply --dry-run=server -f tests/22-clock-skew.yaml
```

```text
timechaos.chaos-mesh.org/clickhouse-chaos-exp-22 created (server dry run)
```

```bash
kubectl apply -f tests/22-clock-skew.yaml
```

```text
timechaos.chaos-mesh.org/clickhouse-chaos-exp-22 created
```

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/22-clock-skew.yaml --timeout=90s
```

```text
timechaos.chaos-mesh.org/clickhouse-chaos-exp-22 condition met
```

Wait for the 45-second fault to finish:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  -f tests/22-clock-skew.yaml --timeout=150s
```

```text
timechaos.chaos-mesh.org/clickhouse-chaos-exp-22 condition met
```

```bash
kubectl delete -f tests/22-clock-skew.yaml
```

```text
timechaos.chaos-mesh.org "clickhouse-chaos-exp-22" deleted from demo namespace
```

After `AllRecovered`, inspect PID 1:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

Output from our run:

```text
    PID STAT COMMAND
      1 Tsl  clickhouse-serv
```

The expected state does not contain `T`. If it does, Chaos Mesh did not fully
clean up the experiment. Resume the existing process and rerun the complete
gate:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  kill -CONT 1
```

`kill -CONT` printed nothing. Checking PID 1 again produced:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

```text
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

**Observed behavior:**

The already-running timestamp stream first returned `08:18`, then `06:18` while TimeChaos was active, and returned to current time after cleanup. Chaos Mesh reported `AllRecovered` but PID 1 was `Tsl`; `SIGCONT` restored `Ssl`. The workload ended this window with 1,083 acknowledged batches and 130 failed or ambiguous attempts.

Result: **PASS WITH MANUAL CLEANUP** — the two-hour clock fault preserved data, but cleanup did not resume ClickHouse.

## Combined Chaos and Recovery Soak

### Chaos#23: Combine I/O Latency with Sibling Failure

#### Create `tests/23-io-plus-sibling-failure.yaml`

Save both resources in `tests/23-io-plus-sibling-failure.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: clickhouse-chaos-exp-23-io
  namespace: demo
spec:
  action: latency
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  containerNames:
    - clickhouse
  volumePath: /var/lib/clickhouse
  path: /var/lib/clickhouse/**/*
  delay: 100ms
  percent: 50
---
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-23-pod
  namespace: demo
spec:
  action: pod-failure
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-1
```

What this chaos does: Adds 100 ms storage latency to shard-0 replica-0
while holding its sibling replica-1 failed.

**Expected behavior:** This removes the healthy fast path for the shard, so
errors and `Critical` are acceptable. Once both faults clear, the siblings
must converge and the IOChaos FUSE mount must disappear without PVC changes.

#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```bash
kubectl apply -f tests/23-io-plus-sibling-failure.yaml
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-23-io created
podchaos.chaos-mesh.org/clickhouse-chaos-exp-23-pod created
```

Confirm that the I/O fault reached replica-0:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  iochaos/clickhouse-chaos-exp-23-io --timeout=90s
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-23-io condition met
```

Confirm that the sibling failure reached replica-1:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-23-pod --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-23-pod condition met
```

Observe the live impact:

```bash
kubectl get clickhouse -n demo clickhouse-chaos
```
```text
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Critical
```

Confirm that IOChaos replaced the normal data mount with its `toda` FUSE
layer:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  findmnt -T /var/lib/clickhouse
```

```text
TARGET              SOURCE FSTYPE OPTIONS
/var/lib/clickhouse toda   fuse   rw,nosuid,nodev,relatime,user_id=0,group_id=0
```

The workload counters during the combined fault were:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
printf "attempted="; cat /state/attempt_batches
printf "successful="; cat /state/success_batches
printf "failed="; cat /state/failed_batches'
```

```text
attempted=1237
successful=1104
failed=132
```

Wait for the I/O fault duration to finish:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  iochaos/clickhouse-chaos-exp-23-io --timeout=2m
```
```text
iochaos.chaos-mesh.org/clickhouse-chaos-exp-23-io condition met
```

Wait for the sibling failure duration to finish:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  podchaos/clickhouse-chaos-exp-23-pod --timeout=2m
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-23-pod condition met
```

Delete the `PodChaos` first:

```bash
kubectl delete podchaos -n demo clickhouse-chaos-exp-23-pod
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-23-pod" deleted from demo namespace
```

Delete the `IOChaos` second:

```bash
kubectl delete iochaos -n demo clickhouse-chaos-exp-23-io
```

```text
iochaos.chaos-mesh.org "clickhouse-chaos-exp-23-io" deleted from demo namespace
```

Check that the normal filesystem is mounted again:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  mount | grep /var/lib/clickhouse
```

```text
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

Check the ClickHouse process state:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

```text
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

Unlike experiments 19 and 22, this process was not stopped after cleanup, so
we did not run `kill -CONT 1`.

Finally, wait for KubeDB to return to `Ready`:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```

```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```

**Observed behavior:**

Shard-0 replica-0 had a `toda` latency mount while replica-1 was failed. KubeDB became `Critical` and three attempts failed or became ambiguous. In this fresh run, deleting the PodChaos and then IOChaos restored ext4 and PID `Ssl` automatically; no `SIGCONT` was required.

Result: **PASS** — the combined shard fault recovered fully and Chaos Mesh cleanup succeeded this time.

### Chaos#24: Run Three Recovery-Soak Cycles

#### Create `tests/24-1-recovery-soak.yaml`
Save the first cycle as `tests/24-1-recovery-soak.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-24-1
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-1-0
  gracePeriod: 0
```

#### Create `tests/24-2-recovery-soak.yaml`
Save the second cycle as `tests/24-2-recovery-soak.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-24-2
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-0-0
  gracePeriod: 0
```

#### Create `tests/24-3-recovery-soak.yaml`
Save the third cycle as `tests/24-3-recovery-soak.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-exp-24-3
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-chaos-cluster-shard-1-1
  gracePeriod: 0
```

What this chaos does: Repeats one-shot pod kills across three replicas,
running the complete recovery gate between cycles.

**Expected behavior:** Every cycle should return to the same healthy baseline.
No replication backlog, checksum difference, stopped process, stale FUSE
mount, or restart instability may accumulate across cycles.

Apply the files in numeric order. Each cycle must recover completely before
the next pod is killed.

#### Cycle 1: clickhouse-chaos-chaos-cluster-shard-1-0

```bash
kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```
```text
7b2c478c-f034-464e-bfce-6040f068a6ba
```

Inject the pod kill:

```bash
kubectl apply -f tests/24-1-recovery-soak.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-1 created
```

Confirm injection:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-24-1 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-1 condition met
```

Delete this one-shot experiment:

```bash
kubectl delete -f tests/24-1-recovery-soak.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-24-1" deleted from demo namespace
```

Wait for this replica:

```bash
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-1-0 --timeout=5m
```
```text
pod/clickhouse-chaos-chaos-cluster-shard-1-0 condition met
```

Run the KubeDB recovery gate before continuing:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Confirm the replacement UID:

```bash
kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```
```text
38d0b075-4ed0-4209-8178-e8c1f8cdd37c
```


#### Cycle 2: clickhouse-chaos-chaos-cluster-shard-0-0

```bash
kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```
```text
21d15784-8304-47b8-b06a-488aadff71af
```

Inject the pod kill:

```bash
kubectl apply -f tests/24-2-recovery-soak.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-2 created
```

Confirm injection:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-24-2 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-2 condition met
```

Delete this one-shot experiment:

```bash
kubectl delete -f tests/24-2-recovery-soak.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-24-2" deleted from demo namespace
```

Wait for this replica:

```bash
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-0-0 --timeout=5m
```
```text
pod/clickhouse-chaos-chaos-cluster-shard-0-0 condition met
```

Run the KubeDB recovery gate before continuing:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Confirm the replacement UID:

```bash
kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```
```text
813edc0a-8d71-42fb-b02d-d543e583abbb
```


#### Cycle 3: clickhouse-chaos-chaos-cluster-shard-1-1

```bash
kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-1 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```
```text
d4a7118b-fb22-4a6a-878e-f0278fa47834
```

Inject the pod kill:

```bash
kubectl apply -f tests/24-3-recovery-soak.yaml
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-3 created
```

Confirm injection:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-24-3 --timeout=90s
```
```text
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-3 condition met
```

Delete this one-shot experiment:

```bash
kubectl delete -f tests/24-3-recovery-soak.yaml
```
```text
podchaos.chaos-mesh.org "clickhouse-chaos-exp-24-3" deleted from demo namespace
```

Wait for this replica:

```bash
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-1-1 --timeout=5m
```
```text
pod/clickhouse-chaos-chaos-cluster-shard-1-1 condition met
```

Run the KubeDB recovery gate before continuing:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
```
```text
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Confirm the replacement UID:

```bash
kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-1 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```
```text
90b19b88-8f26-45da-b187-7d68792b7bcb
```



**Observed behavior:**

Three one-shot kills replaced shard-1 replica-0, shard-0 replica-0, and shard-1 replica-1. Their UIDs changed to `38d0b075-4ed0-4209-8178-e8c1f8cdd37c`, `813edc0a-8d71-42fb-b02d-d543e583abbb`, and `90b19b88-8f26-45da-b187-7d68792b7bcb`. The full KubeDB gate passed between cycles and only one attempt became ambiguous across the soak.

Result: **PASS** — repeated recovery remained stable with no accumulating backlog.

### Chaos#25: Delete One Shard Replica and Its PVC

This final experiment reuses the same `clickhouse-chaos` cluster and the data
written during experiments 1–24. It does not create a second ClickHouse
resource. Killing a pod normally preserves its PVC, so here we delete both one
replica's PVC and its pod to simulate permanent loss of that replica's disk.

**What this chaos does:** Removes the local metadata and data files of shard-0
replica-1. Shard-0 replica-0 remains online as the donor. The two replicas of
shard 1 remain untouched as a control.

**Expected behavior:** KubeDB should provision a new 4Gi PVC and pod, detect
the missing local schema, remove the stale Keeper registration, recreate the
schema from shard-0 replica-0, and let `ReplicatedMergeTree` fetch every part.
The replacement pod, PVC, and PV must have new identities. No manual table
creation, part attachment, or data copy is allowed.

Pause the continuous workload so the baseline remains stable:

```bash
kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

```text
clickhouse-chaos-workload-64d7d5c85f-sgzlc
```

Use the returned pod name to pause the workload:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- \
  touch /state/pause
```

Output: none.

```bash
sleep 5
```

Output: none.

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
if pgrep -x clickhouse-client >/dev/null; then
  echo "client still active"
else
  echo "workload paused"
fi'
```

```text
workload paused
```

Record the workload counters accumulated across experiments 1–24:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
printf "attempted="; cat /state/attempt_batches
printf "successful="; cat /state/success_batches
printf "failed="; cat /state/failed_batches'
```

```text
attempted=1338
successful=1204
failed=134
```

Check the Distributed table before deleting anything:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events FORMAT TSV"'
```

```text
122295  122295  7726510402707751844
```

The row count is higher than `1204 × 100` because some timed-out Distributed
inserts reached ClickHouse even though the client did not receive a success
response. Equality between `count()` and `uniqExact(id)` proves those rows are
not duplicate IDs.

Synchronize the target shard's two replicas, then compare them separately:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SYSTEM SYNC REPLICA chaos_v2.events_local"'
```

Output: none.

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SYSTEM SYNC REPLICA chaos_v2.events_local"'
```

Output: none.

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events_local FORMAT TSV"'
```

```text
61712  61712  629146192837794976
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events_local FORMAT TSV"'
```

```text
61712  61712  629146192837794976
```

Record the donor pod UID:

```bash
kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```

```text
813edc0a-8d71-42fb-b02d-d543e583abbb
```

Record the target pod UID:

```bash
kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```

```text
c956b56f-6582-46c4-b4a3-3d62cc122544
```

Record the target PVC UID and PV:

```bash
kubectl get pvc -n demo \
  data-clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{"\n"}{.spec.volumeName}{"\n"}'
```

```text
1c49ee01-87bf-46d9-b815-05a867174b81
pvc-1c49ee01-87bf-46d9-b815-05a867174b81
```

Delete the target PVC:

```bash
kubectl delete pvc -n demo \
  data-clickhouse-chaos-chaos-cluster-shard-0-1 --wait=false
```

```text
persistentvolumeclaim "data-clickhouse-chaos-chaos-cluster-shard-0-1" deleted from demo namespace
```

Delete the pod so PetSet can create a replacement attached to a new volume:

```bash
kubectl delete pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 \
  --grace-period=0 --force --wait=false
```

```text
Warning: Immediate deletion does not wait for confirmation that the running resource has been terminated.
pod "clickhouse-chaos-chaos-cluster-shard-0-1" force deleted from demo namespace
```

Wait for the replacement pod:

```bash
kubectl wait -n demo --for=create \
  pod/clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

Wait for the replacement PVC:

```bash
kubectl wait -n demo --for=create \
  pvc/data-clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
```

```text
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

Wait for the replacement pod to become Ready:

```bash
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
```

```text
pod/clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

The new pod UID is different:

```bash
kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{"\n"}'
```

```text
638165bc-59cc-4dc6-8039-5335ef9182be
```

The new PVC UID and PV are also different:

```bash
kubectl get pvc -n demo \
  data-clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{"\n"}{.spec.volumeName}{"\n"}'
```

```text
a1b86928-0305-4bdb-8481-1bc58940d933
pvc-a1b86928-0305-4bdb-8481-1bc58940d933
```

The old PV no longer exists:

```bash
kubectl get pv pvc-1c49ee01-87bf-46d9-b815-05a867174b81
```

```text
Error from server (NotFound): persistentvolumes "pvc-1c49ee01-87bf-46d9-b815-05a867174b81" not found
```

Immediately after the replacement starts, prove that its new disk has no copy
of the workload table:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "EXISTS TABLE chaos_v2.events_local"'
```

```text
0
```

This is the live impact: the pod is running, but its local table and data were
really lost. The operator logs first show the expected stale Keeper
registration, then the repair:

```bash
kubectl logs -n kubedb kubedb-kubedb-provisioner-0 --since=3h | \
  grep -E 'replication table kubedb_events_local is missing|REPLICA_ALREADY_EXISTS|replica recovery: repaired' | \
  tail -n 3
```

```text
I0907 08:26:52.721369       1 health.go:240] health check: could not ensure replication probe table on ClickHouse demo/clickhouse-chaos: code: 253, message: There was an error on [clickhouse-chaos-chaos-cluster-shard-0-1.clickhouse-chaos-pods:9000]: Code: 253. DB::Exception: Replica /clickhouse/clickhouse-chaos/chaos-cluster/tables/1/default/kubedb_events_local/replicas/2 already exists. (REPLICA_ALREADY_EXISTS) (version 26.2.6.27 (official build))
E0907 08:26:52.726549       1 health.go:174] failed to check health for db: demo/clickhouse-chaos pod: clickhouse-chaos-chaos-cluster-shard-0-1, error: replication table kubedb_events_local is missing on this replica; it has no local metadata and is awaiting recovery from a sibling
I0907 08:27:00.619727       1 replica_recovery.go:283] replica recovery: repaired clickhouse-chaos-chaos-cluster-shard-0-1 from clickhouse-chaos-chaos-cluster-shard-0-0, created 3 object(s)
```

Wait until the table has been recreated, then synchronize the replica:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SYSTEM SYNC REPLICA chaos_v2.events_local"'
```

Output: none.

Compare the donor:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events_local FORMAT TSV"'
```

```text
61712  61712  629146192837794976
```

Compare the rebuilt replica:

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events_local FORMAT TSV"'
```

```text
61712  61712  629146192837794976
```

Resume the existing workload to prove new writes still work:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- \
  rm -f /state/pause
```

Output: none.

```bash
sleep 5
```

Output: none.

Pause it again for the final stable check:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- \
  touch /state/pause
```

Output: none.

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
printf "attempted="; cat /state/attempt_batches
printf "successful="; cat /state/success_batches
printf "failed="; cat /state/failed_batches'
```

```text
attempted=1342
successful=1208
failed=134
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events FORMAT TSV"'
```

```text
122695  122695  2580264676554960954
```

**Observed behavior:** The same cluster retained all data accumulated during
experiments 1–24. KubeDB created a new 4Gi PVC and repaired the empty replica
from its sibling. The rebuilt replica matched the donor at 61,712 rows before
the workload resumed, and four additional 100-row batches succeeded
afterward.

Result: **PASS** — complete loss of one replica's pod and disk recovered
automatically from its sibling without manual schema or data repair.

## Chaos Testing Results Summary

| # | Fault | Fresh observed impact | Recovery |
| ---: | --- | --- | --- |
| 1 | Single replica pod kill | KubeDB became `Critical`; target UID changed; no workload error | `Ready`; replica queue empty |
| 2 | Replica pod failure, 45s | Target restarted twice; 13 failed/ambiguous attempts | `Critical` → `Ready` |
| 3 | ClickHouse container kill | Same pod UID; restart count 0 → 1 | `Critical` → `Ready` |
| 4 | Three alternating pod kills | Three new pod UIDs; one failed/ambiguous attempt | Full gate passed after every kill |
| 5 | Both replicas of shard 0 failed | Distributed query returned `ALL_CONNECTION_TRIES_FAILED`; KubeDB became `NotReady` | Both replicas returned equal |
| 6 | All four data pods failed | Complete SQL outage; KubeDB became `Critical` then `NotReady` | Four pods reopened their PVC data |
| 7 | Keeper follower kill | Existing leader remained leader; writes continued | Quorum stayed available |
| 8 | Keeper leader kill | Keeper-2 became leader | One leader and two followers restored |
| 9 | Keeper quorum loss | Survivor said it was not serving requests; KubeDB still showed `Ready` | Two failed/ambiguous attempts; quorum reformed |
| 10 | All Keeper members failed | Keeper container exec unavailable; replicated writes stalled | Two failed/ambiguous attempts; quorum reformed |
| 11 | 500ms network delay | KubeDB remained `Ready`; no new workload error | Queue drained |
| 12 | 30% packet loss | KubeDB remained `Ready`; no new workload error | Replica converged |
| 13 | 50% packet duplication | 59,500 rows and 59,500 unique IDs | No duplicate database rows |
| 14 | 1Mbps bandwidth limit | Workload continued without a new error | Replica queue empty |
| 15 | Data-replica partition | One transient workload failure | Shard-0 replicas matched at 33,777 rows |
| 16 | Replica-to-Keeper partition | Target reported `is_readonly=1` and `is_session_expired=1` | Returned writable with two active replicas |
| 17 | CPU stress | Cgroup throttling increased; restart count unchanged | `Ready` throughout |
| 18 | 1GiB memory stress | Usage rose from 1.31GiB to 2.42GiB under a 4GiB limit | Fell to 1.40GiB; no OOM |
| 19 | 100ms filesystem latency | `toda` FUSE mount active | ext4 returned; `SIGCONT` required |
| 20 | 10% EIO | 261 matching storage errors observed during injection | ext4 and PID `Ssl` returned |
| 21 | Keeper DNS errors | Direct lookup failed with exit code 2; existing sessions kept writes alive | DNS resolved after recovery |
| 22 | Clock skew −2h | One running query moved from 08:18 to 06:18 | Clock restored; `SIGCONT` required |
| 23 | I/O latency plus sibling failure | KubeDB became `Critical`; `toda` active | ext4 and PID `Ssl` returned automatically |
| 24 | Three-cycle recovery soak | Three targets received new UIDs; one ambiguous attempt | Full gate passed after every cycle |
| 25 | Existing replica and PVC deletion | New pod/PVC/PV; `EXISTS TABLE` initially returned 0 | 61,712 rows restored from sibling; new writes succeeded |

## Final Integrity Evidence

The same cluster used by all 25 experiments finished with:

```text
ClickHouse phase: Ready
Ready database pods: 7/7
Distributed rows: 122695
Unique IDs: 122695

Shard 0 replica 0: 61911 rows, checksum 461971579980756496
Shard 0 replica 1: 61911 rows, checksum 461971579980756496
Shard 1 replica 0: 60784 rows, checksum 2118293096574204458
Shard 1 replica 1: 60784 rows, checksum 2118293096574204458

Every replica: is_readonly=0, is_session_expired=0, queue_size=0,
               total_replicas=2, active_replicas=2
Keeper-0: leader
Keeper-1: follower
Keeper-2: follower
Data PVCs: 4Gi and Bound
Old shard-0 replica-1 PV: deleted
Replacement shard-0 replica-1 PVC: Bound
Remaining test chaos objects: 0
```

The final workload counters were 1,342 attempted batches, 1,208 acknowledged
batches, and 134 failed or ambiguous attempts. The database contained 122,695
unique rows. The 1,895 rows above `1,208 × 100` came from inserts for which
ClickHouse accepted data but the client did not receive a success response
before its timeout. They are not duplicate IDs.

## What the Errors Mean

- **Client timeout or connection failure:** the selected shard, all data pods,
  Keeper, network, or storage was temporarily unavailable. Failed attempts are
  expected during those tests.
- **`Critical` KubeDB phase:** at least part of the desired cluster was not
  healthy. It does not always mean every query is unavailable.
- **KubeDB stayed `Ready` during Keeper loss:** the health check could still
  connect to ClickHouse, but Keeper-dependent operations were degraded. Check
  Keeper directly rather than relying on one status field.
- **Temporary replica count mismatch:** a part accepted around Keeper recovery
  was still propagating. It became a problem only if consecutive checks failed
  to converge within the recovery timeout.
- **`REPLICA_ALREADY_EXISTS` after PVC loss:** the new disk had no local table,
  but Keeper still remembered the old replica registration. This was the
  expected intermediate condition. The recovery controller removed the stale
  registration before replaying the sibling's table definition.
- **`EXISTS TABLE` returned `0`:** the replacement ClickHouse process was
  reachable, but its new PVC contained no local schema. This proved the test
  caused real disk loss rather than an ordinary pod restart.
- **`Input/output error` and `CANNOT_STATVFS`:** IOChaos deliberately made
  filesystem calls return EIO. These errors disappeared after the fault.
- **PID state `Tsl`:** the standalone I/O-latency run and the TimeChaos run left ClickHouse
  stopped even after Chaos Mesh reported recovery. `SIGCONT` resumed the
  process without deleting data or recreating the pod. This indicates an
  incomplete chaos-tool cleanup, not data corruption.

## Key Findings

1. Replica and Keeper redundancy protected acknowledged data across every
   tested recoverable failure.
2. KubeDB `Ready` is useful but insufficient for Keeper-specific faults. A real
   database query, replica state, and Keeper `mntr` must be checked.
3. Unique row IDs are essential because a timed-out Distributed insert can
   have an ambiguous partial result.
4. Recovery verification needs consecutive matching replica checks, not one
   instantaneous queue sample.
5. Chaos Mesh 2.8.4 handled PodChaos, NetworkChaos, StressChaos, DNSChaos, and
   the EIO case cleanly. The standalone I/O-latency and TimeChaos runs required
   manual `SIGCONT`; the combined I/O experiment cleaned up automatically.
6. Memory pressure rose to about 2.42GiB under a 4GiB limit without an OOM,
   then fell toward baseline after cleanup.
7. Losing both a replica pod and its PVC requires schema recovery before normal
   ClickHouse replication can resume. The recovery-enabled KubeDB operator
   removed stale Keeper registrations, restored the schema from the same-shard
   sibling, and then let `ReplicatedMergeTree` fetch the data automatically.

## Conclusion

This campaign showed that the tested ClickHouse topology protected
acknowledged data across pod, process, Keeper, network, CPU, memory, storage,
DNS, clock, combined failures, and permanent loss of one replica's PVC.
Expected downtime occurred when an entire shard or the whole data plane was
unavailable; that is correct behavior for this topology, not a failed recovery
mechanism.

The most important operational lesson is that one green status field is not a
complete health check. A trustworthy ClickHouse recovery decision combines a
real read and write, shard-local count and checksum equality, `system.replicas`
state, Keeper quorum, process state, and clean storage mounts.

Experiments 19 and 22 also separate database resilience from chaos-tool
cleanup. ClickHouse data remained correct under storage latency and a two-hour
clock offset, but Chaos Mesh 2.8.4 did not resume the affected process
automatically in those two runs. Experiment 23 cleaned up automatically.
Therefore the database-integrity checks passed, while automatic cleanup failed
in two experiments.

## What Next?

Repeat the same suite on a multi-node cluster with production-equivalent
storage, workload volume, resource limits, and monitoring. NodeChaos can then
be added safely because losing one worker will not also remove the Kubernetes
control plane and the chaos controller.

## Cleanup

Save the final evidence before removing anything. Pause the workload, record
its counters, then scale it down:

```bash
kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

```text
clickhouse-chaos-workload-64d7d5c85f-sgzlc
```

Pause the workload using the pod name returned above:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- touch /state/pause
```

The command prints nothing. Read the final counters:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-sgzlc -- bash -c '
  printf "attempts="; cat /state/attempt_batches
  printf "success="; cat /state/success_batches
  printf "failed="; cat /state/failed_batches
'
```

```text
attempts=1342
success=1208
failed=134
```

```bash
kubectl scale deployment -n demo \
  clickhouse-chaos-workload --replicas=0
```

```text
deployment.apps/clickhouse-chaos-workload scaled
```

Repeat the complete recovery gate directly from a ClickHouse pod. Record the
Distributed and local counts, unique-ID counts, checksums, `system.replicas`
state, Keeper roles, pod state, PID states, mounts, and
remaining Chaos Mesh resources.

Delete only resources belonging to this disposable campaign:

```bash
kubectl get podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos -n demo
```

```text
No resources found in demo namespace.
```

```bash
kubectl delete deployment -n demo clickhouse-chaos-workload \
  --ignore-not-found
```

```text
deployment.apps "clickhouse-chaos-workload" deleted
```

```bash
kubectl delete configmap -n demo clickhouse-chaos-workload \
  --ignore-not-found
```

```text
configmap "clickhouse-chaos-workload" deleted
```

```bash
kubectl delete clickhouse -n demo clickhouse-chaos \
  --ignore-not-found
```

```text
clickhouse.kubedb.com "clickhouse-chaos" deleted
```

Because the test ClickHouse uses `deletionPolicy: WipeOut`, its managed pods
and PVCs should disappear. Give the operator time to finish, then inspect the
namespace directly:

```bash
sleep 60
```

`sleep` prints nothing. If a campaign pod is still `Terminating`, wait another
30 seconds. Do not delete its PVC manually.

The generated auth Secret is owned by the ClickHouse resource and is removed
by `WipeOut`; do not delete credentials manually. Verify that the campaign
left nothing behind:

```bash
kubectl get clickhouse,petset,pods,pvc,secrets -n demo -o name | \
  grep clickhouse-chaos
```

Output from our final name check:

```text
(no output)
```

The ClickHouse, its PetSets, pods, PVCs, generated
auth Secret, and Chaos Mesh resources no longer existed. Unrelated resources
in `demo` remained unchanged.
