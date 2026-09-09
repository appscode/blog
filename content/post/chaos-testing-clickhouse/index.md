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
dataset. They preserved ClickHouse data through the recoverable database
faults. TimeChaos did not demonstrate a live clock skew: it stopped the target
process and required manual `SIGCONT` after cleanup, so experiment 22 is
reported as a chaos-tool limitation rather than a pass. Both replicas
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
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-0-0   Bound    pvc-9773c0e0-8bd6-4a67-b445-444255b61074   4Gi        RWO            local-path     <unset>                 2m15s
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-0-1   Bound    pvc-af84e532-ec8b-4ddc-938f-34d6a21bdeeb   4Gi        RWO            local-path     <unset>                 2m9s
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-1-0   Bound    pvc-eecf7db8-b9b5-4888-b1aa-71e70ea65119   4Gi        RWO            local-path     <unset>                 2m13s
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-1-1   Bound    pvc-35ef4112-0c9d-42b1-973a-f26389b04864   4Gi        RWO            local-path     <unset>                 2m8s
persistentvolumeclaim/data-clickhouse-chaos-keeper-0                  Bound    pvc-531f098d-5196-490d-b6ba-b3584755b06c   1Gi        RWO            local-path     <unset>                 2m17s
persistentvolumeclaim/data-clickhouse-chaos-keeper-1                  Bound    pvc-41c5eecc-1f51-477d-aa7f-22e2e1c88ba9   1Gi        RWO            local-path     <unset>                 2m11s
persistentvolumeclaim/data-clickhouse-chaos-keeper-2                  Bound    pvc-e893bff9-8ba5-4da4-92f7-b0b125e4777f   1Gi        RWO            local-path     <unset>                 2m6s

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
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Kubernetes generates the suffix in this pod name. In the commands below, use
the workload pod name returned by your own cluster.

```bash
$ kubectl logs -n demo -f clickhouse-chaos-workload-64d7d5c85f-fqjdp
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
2. Require the ClickHouse resource and all four ClickHouse plus three Keeper pods to be `Ready`.
3. Confirm that no test Chaos Mesh object remains.
4. Perform an authenticated probe insert and require `count() == uniqExact(id)`
   on the Distributed table.
5. Require matching local counts and checksums for each shard in two
   consecutive checks. A check is a set of query results
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
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- touch /state/pause
```

The command prints nothing.

```bash
$ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- bash -c   'if pgrep -x clickhouse-client >/dev/null; then echo "client still active"; else echo "workload paused"; fi'
workload paused
```

Require ClickHouse and all seven database pods to be healthy:

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
pod/clickhouse-chaos-workload-64d7d5c85f-fqjdp   1/1     Running   0          6m39s

```

Confirm that no test fault remains:

```bash
➤ kubectl get podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos -n demo
No resources found in demo namespace.
```


Check the Distributed table. The first two values must match, and the count
must be at least `successful_batches × 100`:

```bash
kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- bash -c '
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
$ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- bash -c '
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


Shard-0's two lines match, and shard-1's two lines match. Run the same four
commands again. The second check must return the same
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

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Start the workload using the pod name returned above:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- rm -f /state/pause
```

The command prints nothing on success. Validate the manifest against the
API server:

```shell
➤ kubectl apply --dry-run=server -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 created (server dry run)
```

Inject the fault:

```shell
➤ kubectl apply -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 created
```

Prove that Chaos Mesh injected it:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  -f tests/01-pod-kill.yaml --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 condition met
```

Observe ClickHouse during the fault:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    7m
```

Read the workload counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- bash -c '
  printf "attempts="; cat /state/attempt_batches
  printf "success="; cat /state/success_batches
  printf "failed="; cat /state/failed_batches
'
attempts=46
success=46
failed=0
```

Remove the experiment:

```shell
➤ kubectl delete -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-01" deleted from demo namespace
```

Deletion only removes the injected fault; it does not prove that ClickHouse
recovered. After deletion, run every command in the **Mandatory Recovery Gate**
section immediately above. In this guide, “run the full gate” means:

1. Pause the workload and confirm its active client has stopped.
2. Wait for ClickHouse to become `Ready` and check all seven database pods.
3. Confirm that no Chaos Mesh resource from the test remains.
4. Perform the one-row probe insert and query the Distributed table.
5. Check every local replica in two stable snapshots.
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

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  -f tests/02-pod-failure.yaml --timeout=150s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-02 condition met
```

```shell
➤ kubectl delete -f tests/02-pod-failure.yaml
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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```


Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

Keep the workload running while observing the fault and recovery transition.

Record the target UID before applying the manifest, then confirm that it
changes after injection.

#### Demonstrate impact and recovery

Confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    6m
```

Record the original pod UID:

```shell
➤ kubectl get pod -n demo \
        clickhouse-chaos-chaos-cluster-shard-0-0 \
        -o jsonpath='{.metadata.uid}{"\n"}'
e8970f8d-477f-42ba-87ed-edc6d0f41a89
```


Apply this experiment:

```shell
➤ kubectl apply -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 created

```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-01 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-01 condition met
```

Observe the live impact:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    7m
```

Wait for PetSet to make the replacement pod ready:

```shell
➤ kubectl wait -n demo --for=condition=Ready \
        pod/clickhouse-chaos-chaos-cluster-shard-0-0 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-0-0 condition met
```

Confirm that the replacement has a new UID:

```shell
➤ kubectl get pod -n demo \
        clickhouse-chaos-chaos-cluster-shard-0-0 \
        -o jsonpath='{.metadata.uid}{"\n"}'
a1519479-69cf-45a5-b50a-550bcc5f941c
```


Delete the experiment:

```shell
➤ kubectl delete -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-01" deleted from demo namespace
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met

```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing. Record the stable workload counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
46 46 0
```

Verify the two shard-0 replicas separately:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
2319	2319	13499264341016960427
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- \
        bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
2319	2319	13499264341016960427
```

**Observed behavior:**

The target pod UID changed from `e8970f8d-477f-42ba-87ed-edc6d0f41a89` to `a1519479-69cf-45a5-b50a-550bcc5f941c`. ClickHouse stayed `Ready`; the workload advanced from `18 18 0` to `46 46 0`. After cleanup, the replica was writable with an empty queue and two active replicas.

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

**Expected behavior:** ClickHouse should report a degraded state while the sibling
replica continues serving the shard. When the fault ends, the same pod should
become reachable and converge without manual repair.

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

Keep the workload running while observing the fault and recovery transition.

Before injecting the fault, the fresh cluster was healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    8m
```


```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{" "}{.status.containerStatuses[0].restartCount}{"\n"}'
61f18b79-43f9-44a8-b9bc-d1db5baf6b29 0
```


Apply the file and confirm that Chaos Mesh really injected the failure. A
created object alone is not evidence that the fault reached the target:

```shell
➤ kubectl apply -f tests/02-pod-failure.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-02 created
```


```shell
➤ kubectl get podchaos -n demo clickhouse-chaos-exp-02 \
        -o jsonpath='{range .status.conditions[*]}{.type}={.status}{"\n"}{end}'
Selected=True
AllInjected=True
AllRecovered=False
Paused=False
```

While the 45-second fault was active, the target pod stayed present but its
container restarted. ClickHouse status may be `Critical` at this sample but the
healthy sibling could serve the shard:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS     AGE
clickhouse-chaos   26.2.6    Critical   9m
```

Prove that the selected replica rejects a direct SQL connection while the
fault is active:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'clickhouse-client --host clickhouse-chaos-chaos-cluster-shard-0-1.clickhouse-chaos-pods.demo.svc --user "$CH_USER" --password "$CH_PASSWORD" --query "SELECT 1"; echo exit_code=$?'
Code: 210. DB::NetException: Connection refused
exit_code=210
```

```shell
➤ kubectl get pod -n demo \
        clickhouse-chaos-chaos-cluster-shard-0-1
NAME                                       READY   STATUS    RESTARTS      AGE
clickhouse-chaos-chaos-cluster-shard-0-1   1/1     Running   1              9m
```

After the duration elapsed, Chaos Mesh reported recovery.

```shell
➤ kubectl get podchaos -n demo clickhouse-chaos-exp-02 \
        -o jsonpath='{range .status.conditions[*]}{.type}={.status}{"\n"}{end}'
Selected=True
AllInjected=False
AllRecovered=True
Paused=False

```


```shell
➤ kubectl delete -f tests/02-pod-failure.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-02" deleted from demo namespace
```

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    10m
```

The UID is unchanged and the final restart count is two:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{" "}{.status.containerStatuses[0].restartCount}{"\n"}'
61f18b79-43f9-44a8-b9bc-d1db5baf6b29 2
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT is_readonly, queue_size, total_replicas, active_replicas
             FROM system.replicas
             WHERE database='\''chaos_v2'\'' AND table='\''events_local'\''
  FORMAT TSV"'
0	0	2	2
```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
107 99 8
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

The target kept the same UID, `61f18b79-43f9-44a8-b9bc-d1db5baf6b29`, and restarted twice. A direct query returned `Connection refused`, ClickHouse became `Critical`, and the workload moved from `46 46 0` to `107 99 8`. After `AllRecovered=True`, the test still waited for SQL and ClickHouse `Ready`.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

Compare the `clickhouse` container restart count before and after injection.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    11m
```

Record the pod UID and restart count before injection:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
        -o jsonpath='{.metadata.uid}{" "}{.status.containerStatuses[0].restartCount}{"\n"}'
326909c8-1fe4-4c3e-99f8-3c14f8680d4b 0
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/03-container-kill.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-03 created
```
Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-03 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-03 condition met
```

Observe the live impact:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
        -o jsonpath='{.metadata.uid}{"\n"}{.status.containerStatuses[0].restartCount}{"\n"}'
326909c8-1fe4-4c3e-99f8-3c14f8680d4b
1
```

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    12m
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/03-container-kill.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-03" deleted from demo namespace
```


Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
134 126 8
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

The pod UID remained `326909c8-1fe4-4c3e-99f8-3c14f8680d4b`, while its restart count changed from 0 to 1. ClickHouse stayed `Ready`; the workload moved from `107 99 8` to `134 126 8`.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

Keep the workload running while observing the fault and recovery transition.

Apply the files one at a time. Wait 20 seconds between kills, delete each
one-shot `PodChaos` after `AllInjected`, and run the complete recovery gate
after the third kill.

```shell
➤ kubectl apply -f tests/04-a-pod-kill.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-a created
```

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        -f tests/04-a-pod-kill.yaml --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-a condition met
```

```shell
➤ kubectl delete -f tests/04-a-pod-kill.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-04-a" deleted from demo namespace
```

```shell
➤ kubectl wait -n demo --for=condition=Ready \
        pod/clickhouse-chaos-chaos-cluster-shard-0-0 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-0-0 condition met

```

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
        -o jsonpath='{.metadata.uid}{"\n"}'
add2a88c-ed3c-46f3-bbb4-b2cb61fec420
```

Before starting the next fault, compare both shard-0 replicas:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
7810  7810  18036471919944389959
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
7810  7810  18036471919944389959
```

```shell
➤ kubectl apply -f tests/04-b-pod-kill.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-b created
```

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        -f tests/04-b-pod-kill.yaml --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-b condition met
```


```shell
➤ kubectl delete -f tests/04-b-pod-kill.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-04-b" deleted from demo namespace
```

```shell
➤ kubectl wait -n demo --for=condition=Ready \
        pod/clickhouse-chaos-chaos-cluster-shard-1-1 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-1-1 condition met
```

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-1 \
        -o jsonpath='{.metadata.uid}{"\n"}'
d8da0cb3-168b-40ad-840b-ecb835c27cea
```

Before starting the third fault, compare both shard-1 replicas:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
8846  8846  12894136727626022465
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
8846  8846  12894136727626022465
```

```shell
➤ kubectl apply -f tests/04-c-pod-kill.yaml

podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-c created
```

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        -f tests/04-c-pod-kill.yaml --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-04-c condition met
```

```shell
➤ kubectl delete -f tests/04-c-pod-kill.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-04-c" deleted from demo namespace

```

```shell
➤ kubectl wait -n demo --for=condition=Ready \
        pod/clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-1 \
        -o jsonpath='{.metadata.uid}{"\n"}'
0a753027-af14-428b-8b78-9192f190cfeb
```

#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
204 195 9
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

Shard-0 replica-0, shard-1 replica-1, and shard-0 replica-1 each received a new UID. The explicit SQL, replica-state, and checksum checks passed before each following kill. The workload moved from `134 126 8` to `204 195 9`; both replica pairs ended with matching counts and checksums.

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
ClickHouse should report `Critical`, then both replicas should return with equal
data.

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

Keep the workload running while observing the fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    19m
```


Apply this experiment:

```shell
➤ kubectl apply -f tests/05-full-shard-outage.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-05 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-05 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-05 condition met
```

Observe the live impact:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS     AGE
clickhouse-chaos   26.2.6    Critical   20m
```

Prove that each selected replica is unreachable:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'clickhouse-client --host clickhouse-chaos-chaos-cluster-shard-0-0.clickhouse-chaos-pods.demo.svc --user "$CH_USER" --password "$CH_PASSWORD" --query "SELECT 1"; echo exit_code=$?'
Code: 210. DB::NetException: Connection refused
exit_code=210
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'clickhouse-client --host clickhouse-chaos-chaos-cluster-shard-0-1.clickhouse-chaos-pods.demo.svc --user "$CH_USER" --password "$CH_PASSWORD" --query "SELECT 1"; echo exit_code=$?'
Code: 210. DB::NetException: Connection refused
exit_code=210
```

Run a Distributed query from the healthy shard. It fails explicitly because
the unavailable shard cannot be silently skipped:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count() FROM chaos_v2.events"'
Code: 279. DB::Exception: All connection tries failed. (ALL_CONNECTION_TRIES_FAILED)
command terminated with exit code 279
```


Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        podchaos/clickhouse-chaos-exp-05 --timeout=2m
podchaos.chaos-mesh.org/clickhouse-chaos-exp-05 condition met

```

Delete the experiment:

```shell
➤ kubectl delete -f tests/05-full-shard-outage.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-05" deleted from demo namespace
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1491 1441 50
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
72274	72274	17024341840210372511
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- \
        bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
72274	72274	17024341840210372511
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

Both shard-0 replicas were unavailable. Direct queries returned `Connection refused`, a Distributed query from shard 1 returned `ALL_CONNECTION_TRIES_FAILED`, and ClickHouse became `Critical`. At the recovery checkpoint the workload was `1491 1441 50`; both replicas then matched at `72274` rows with queue 0.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    36m

```

Apply this experiment:

```shell
➤ kubectl apply -f tests/06-data-plane-outage.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-06 created

```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-06 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-06 condition met

```

Observe the live impact:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS     AGE
clickhouse-chaos   26.2.6    Critical   36m
```

Prove that the SQL service is unavailable:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'clickhouse-client --host clickhouse-chaos.demo.svc --user "$CH_USER" --password "$CH_PASSWORD" --query "SELECT 1"; echo exit_code=$?'
Code: 210. DB::NetException: Connection refused (clickhouse-chaos.demo.svc:9000). (NETWORK_ERROR)
exit_code=210
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        podchaos/clickhouse-chaos-exp-06 --timeout=2m
podchaos.chaos-mesh.org/clickhouse-chaos-exp-06 condition met

```

Delete the experiment:

```shell
➤ kubectl delete -f tests/06-data-plane-outage.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-06" deleted from demo namespace
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met

```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1531 1441 90
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

All four data containers were failed while Keeper stayed online. The service query returned `Connection refused`, ClickHouse became `Critical`, and the workload moved from `1491 1441 50` to `1531 1441 90`: 40 attempts, no acknowledgements. All four pods reopened their existing PVCs and ClickHouse returned to `Ready`.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

The example uses `keeper-1`; replace it if `mntr` reports that member as the
leader.


#### Demonstrate impact and recovery

Confirm the role of each Keeper member before selecting the follower:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
leader
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-1 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
follower
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-2 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
follower
```

Record the selected follower UID:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-keeper-1 -o jsonpath='{.metadata.uid}{"\n"}'
2eb512bc-c64f-448e-a6bc-ad43cdaffa82
```

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    37m

```

Apply this experiment:

```shell
➤ kubectl apply -f tests/07-keeper-follower-kill.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-07 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-07 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-07 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- \
        bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | \
        awk '$1=="zk_server_state" {print $2}'
leader

```

Delete the experiment:

```shell
➤ kubectl delete -f tests/07-keeper-follower-kill.yaml

podchaos.chaos-mesh.org "clickhouse-chaos-exp-07" deleted from demo namespace
```


Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl get pod -n demo clickhouse-chaos-keeper-1 -o jsonpath='{.metadata.uid}{" "}{.status.phase}{"\n"}'
cd8ef14e-548d-4bd6-8784-8c964cf877eb Running
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1557 1467 90
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

Keeper-1 was verified as a follower before it was killed. Its UID changed from `2eb512bc-c64f-448e-a6bc-ad43cdaffa82` to `cd8ef14e-548d-4bd6-8784-8c964cf877eb`; Keeper-0 remained leader, ClickHouse stayed `Ready`, and all 26 new attempts were acknowledged.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

 Keep the workload running while observing the fault and recovery transition.

The example uses `keeper-0`; replace it with the actual leader. Time how long
another member takes to report `leader`.


#### Demonstrate impact and recovery

Before injection, verify that Keeper-0 is the leader:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
leader
```

Record its UID:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-keeper-0 -o jsonpath='{.metadata.uid}{"\n"}'
2a540a7e-dbad-4fd6-a7b1-f57d75d7a643
```

Confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos

NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    38m
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/08-keeper-leader-kill.yaml

podchaos.chaos-mesh.org/clickhouse-chaos-exp-08 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-08 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-08 condition met

```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-1 -c clickhouse-keeper -- \
        bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | \
        awk '$1=="zk_server_state" {print $2}'
leader

```

Delete the experiment:

```shell
➤ kubectl delete -f tests/08-keeper-leader-kill.yaml

podchaos.chaos-mesh.org "clickhouse-chaos-exp-08" deleted from demo namespace
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Confirm the recovered quorum one member at a time:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
follower
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-1 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
leader
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-2 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
follower
```



#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl get pod -n demo clickhouse-chaos-keeper-0 -o jsonpath='{.metadata.uid}{" "}{.status.phase}{"\n"}'
838451cd-cc02-42b5-8a5c-fbc023c33d9d Running
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1567 1476 91
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

Keeper-0 was the leader before injection. Its UID changed from `2a540a7e-dbad-4fd6-a7b1-f57d75d7a643` to `838451cd-cc02-42b5-8a5c-fbc023c33d9d`; Keeper-1 became leader and the other two members reported follower. ClickHouse stayed `Ready`.

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
  duration: 90s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-keeper-0
        - clickhouse-chaos-keeper-2
```

What this chaos does: Holds two of the three Keeper members failed for 90
seconds, removing the majority required for coordination.

**Expected behavior:** Existing reads may continue, but coordination-dependent
writes or replication can stall or fail. After quorum returns, queued work
should settle and both replicas of every shard should converge without data
repair.

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

Do not repair a brief replica mismatch while queues are still moving. Require
two consecutive equal checks within ten minutes.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    42m
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/09-keeper-quorum-loss.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-09 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-09 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-09 condition met

```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-1 -c clickhouse-keeper -- \
        bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3'
This instance is not currently serving requests
```

Prove a coordination-dependent write stalls instead of being acknowledged:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'timeout 8 clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "INSERT INTO chaos_v2.events_local VALUES (generateUUIDv4(), now(), 9)"; echo exit_code=$?'
exit_code=124
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        podchaos/clickhouse-chaos-exp-09 --timeout=2m
podchaos.chaos-mesh.org/clickhouse-chaos-exp-09 condition met
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/09-keeper-quorum-loss.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-09" deleted from demo namespace
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Verify that the recovered replica queue is empty:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT is_readonly, is_session_expired, queue_size, total_replicas, active_replicas FROM system.replicas WHERE database='\''chaos_v2'\'' AND table='\''events_local'\''"'
0  0  0  2  2
```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1611 1513 97
```

After any in-flight batch finishes, run the mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

Keeper-0 and Keeper-2 were selected together for 90 seconds. The lone Keeper-1 returned `This instance is not currently serving requests`, and a replicated insert timed out with exit code 124. ClickHouse still showed `Ready`; recovery restored an empty replica queue.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

During injection, check `mntr` directly even if ClickHouse still reports `Ready`.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    43m
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/10-full-keeper-outage.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-10 created
```


Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        podchaos/clickhouse-chaos-exp-10 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-10 condition met

```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- \
        bash -c 'printf "mntr\n"'
error: Internal error occurred: Internal error occurred: error executing command in container: failed to exec in container: failed to start exec "628453296b89642b1b9551dfba67fe273943dc7f77f1bd17c3c2ae9965933a96": OCI runtime exec failed: exec failed: unable to start container process: exec: "bash": executable file not found in $PATH

```

Prove a replicated write cannot complete without Keeper:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'timeout 8 clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "INSERT INTO chaos_v2.events_local VALUES (generateUUIDv4(), now(), 10)"; echo exit_code=$?'
exit_code=124
```


Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        podchaos/clickhouse-chaos-exp-10 --timeout=2m
podchaos.chaos-mesh.org/clickhouse-chaos-exp-10 condition met
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/10-full-keeper-outage.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-10" deleted from demo namespace
```


Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met

```

Confirm that Keeper formed one leader and two followers again:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
leader
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-1 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
follower
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-2 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
follower
```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1618 1517 100
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

`AllInjected=True` selected all three Keeper pods, the sampled Keeper-0 exec was rejected while the fault was active, and a replicated insert timed out with exit code 124. ClickHouse data processes remained present and ClickHouse still showed `Ready`. Keeper returned with one leader and two followers.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    45m
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/11-network-delay.yaml
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-11 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        networkchaos/clickhouse-chaos-exp-11 --timeout=90s
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-11 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        ping -c 5 -W 2 10.42.0.42
5 packets transmitted, 5 packets received, 0% packet loss
round-trip min/avg/max = 468.815/517.919/548.130 ms
```


Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        networkchaos/clickhouse-chaos-exp-11 --timeout=2m
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-11 condition met
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/11-network-delay.yaml
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-11" deleted from demo namespace
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- ping -c 5 -W 2 10.42.0.42
5 packets transmitted, 5 packets received, 0% packet loss
round-trip min/avg/max = 0.049/0.070/0.087 ms
```


Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met

```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1633 1532 100
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

Ping average rose from `0.113 ms` to `517.919 ms` under the configured delay and returned to `0.070 ms` after cleanup. ClickHouse remained `Ready`; the counters moved from `1618 1517 100` to `1633 1532 100`, so all 15 new attempts were acknowledged. The target finished writable with queue 0.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```


Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause

```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    46m
```
Apply this experiment:

```shell
➤ kubectl apply -f tests/12-network-loss.yaml
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-12 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        networkchaos/clickhouse-chaos-exp-12 --timeout=90s
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-12 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        ping -c 20 -W 1 10.42.0.42
20 packets transmitted, 13 packets received, 35% packet loss
round-trip min/avg/max = 0.063/0.082/0.120 ms
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        networkchaos/clickhouse-chaos-exp-12 --timeout=2m
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-12 condition met

```

Delete the experiment:

```shell
➤ kubectl delete -f tests/12-network-loss.yaml
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-12" deleted from demo namespace
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- ping -c 5 -W 2 10.42.0.42
5 packets transmitted, 5 packets received, 0% packet loss
round-trip min/avg/max = 0.045/0.106/0.234 ms
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met

```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1658 1558 100
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

The 20-packet probe received 13 replies and reported `35% packet loss`; the post-recovery probe received all five replies. ClickHouse remained `Ready`, the workload advanced to `1658 1558 100`, and the target finished writable with queue 0.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause

```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    48m
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/13-network-duplicate.yaml
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-13 created
```


Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        networkchaos/clickhouse-chaos-exp-13 --timeout=90s
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-13 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        ping -c 20 -W 1 10.42.0.42
20 packets transmitted, 20 packets received, 8 duplicates, 0% packet loss
round-trip min/avg/max = 0.049/0.082/0.140 ms

```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        networkchaos/clickhouse-chaos-exp-13 --timeout=2m
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-13 condition met
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/13-network-duplicate.yaml
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-13" deleted from demo namespace
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
80556	80556	11545698640526106827
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
80556	80556	11545698640526106827
```


Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met

```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause

```
The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1697 1597 100
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

The 20-packet probe reported eight `DUP!` replies. After recovery, both shard-0 replicas returned `80556` rows, `80556` unique IDs, and the same checksum.

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
        - clickhouse-chaos-workload-64d7d5c85f-fqjdp
  direction: both
  target:
    mode: all
    selector:
      namespaces:
        - demo
      pods:
        demo:
          - clickhouse-chaos-chaos-cluster-shard-0-0
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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause

```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    49m
```


Apply this experiment:

```shell
➤ kubectl apply -f tests/14-bandwidth.yaml
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-14 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        networkchaos/clickhouse-chaos-exp-14 --timeout=90s
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-14 condition met

```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c \
        '/usr/bin/time -f "elapsed=%e" clickhouse-client --host 10.42.0.42 --user "$CH_USER" --password "$CH_PASSWORD" --compression=0 --query "SELECT hex(randomString(450000)) FROM numbers(5) FORMAT TSV" >/dev/null'
elapsed=4.80

```


Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        networkchaos/clickhouse-chaos-exp-14 --timeout=2m
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-14 condition met

```

Delete the experiment:

```shell
➤ kubectl delete -f tests/14-bandwidth.yaml
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-14" deleted from demo namespace
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c \
        '/usr/bin/time -f "elapsed=%e" clickhouse-client --host 10.42.0.42 --user "$CH_USER" --password "$CH_PASSWORD" --compression=0 --query "SELECT hex(randomString(450000)) FROM numbers(5) FORMAT TSV" >/dev/null'
elapsed=0.21
```


Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met

```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1849 1749 100
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

With compression disabled, the same 4.5 MB response took `4.80s` under the 1 Mbps cap and `0.21s` after recovery. The workload moved from `1729 1629 100` to `1849 1749 100`, and the replication queue was empty afterward.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
        -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp

```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    8h
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/15-data-partition.yaml
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-15 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
        networkchaos/clickhouse-chaos-exp-15 --timeout=90s
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-15 condition met

```


Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'timeout 3 bash -c "</dev/tcp/10.42.0.44/9000"; echo exit_code=$?'
exit_code=124
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id) FROM chaos_v2.events_local; SELECT queue_size FROM system.replicas WHERE database='\''chaos_v2'\''"'
88120	88120
12
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- \
        bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id) FROM chaos_v2.events_local; SELECT queue_size FROM system.replicas WHERE database='\''chaos_v2'\''"'
88532	88532
0
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
        networkchaos/clickhouse-chaos-exp-15 --timeout=2m
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-15 condition met

```

Delete the experiment:

```shell
➤ kubectl delete -f tests/15-data-partition.yaml
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-15" deleted from demo namespace
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c 'timeout 3 bash -c "</dev/tcp/10.42.0.44/9000"; echo exit_code=$?'
exit_code=0
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local; SELECT queue_size FROM system.replicas WHERE database='\''chaos_v2'\''"'
89357	89357	2709723990317646529
0
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local; SELECT queue_size FROM system.replicas WHERE database='\''chaos_v2'\''"'
89357	89357	2709723990317646529
0
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
        clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met

```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
        touch /state/pause
```

The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1875 1773 102
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

Peer TCP changed from exit code 0 to timeout exit code 124. During isolation, replica-0 had `88120` rows and queue 12 while its sibling had `88532` rows and queue 0. Without `SYSTEM SYNC REPLICA`, both automatically converged to `89357` rows with the same checksum and queue 0. The workload ended at `1875 1773 102`.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/16-keeper-partition.yaml
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-16 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  networkchaos/clickhouse-chaos-exp-16 --timeout=90s
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-16 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT is_readonly, is_session_expired, queue_size FROM system.replicas \
  WHERE database='\''chaos_v2'\'' AND table='\''events_local'\'' FORMAT TSV"'
1  1  0
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'timeout 8 clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "INSERT INTO chaos_v2.events_local VALUES (generateUUIDv4(), now(), 16)"; echo exit_code=$?'
exit_code=124
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  networkchaos/clickhouse-chaos-exp-16 --timeout=2m
networkchaos.chaos-mesh.org/clickhouse-chaos-exp-16 condition met
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/16-keeper-partition.yaml
networkchaos.chaos-mesh.org "clickhouse-chaos-exp-16" deleted from demo namespace
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c 'timeout 3 bash -c "</dev/tcp/10.42.0.46/9181"; echo exit_code=$?'
exit_code=0
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT is_readonly, is_session_expired, queue_size, total_replicas, active_replicas FROM system.replicas WHERE database='\''chaos_v2'\''"'
0	0	0	2	2
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1889 1784 105
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

The target lost all Keeper connectivity. After 20 seconds it reported `is_readonly=1` and `is_session_expired=1`, correctly refusing uncoordinated replicated writes. After cleanup it returned `is_readonly=0`, `queue_size=0`, and `active_replicas=2`. The workload ended at `1889 1784 105`.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.


#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Record the CPU limit and baseline throttling counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'cat /sys/fs/cgroup/cpu.max; grep -E "usage_usec|nr_throttled|throttled_usec" /sys/fs/cgroup/cpu.stat'
100000 100000
usage_usec 117916959
nr_throttled 203
throttled_usec 23817473
```

Record the restart count before stress:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.status.containerStatuses[0].restartCount}{"\n"}'
4
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/17-cpu-stress.yaml
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-17 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  stresschaos/clickhouse-chaos-exp-17 --timeout=90s
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-17 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        bash -c 'ps -eo pid,pcpu,comm | grep -E "stress-ng|clickhouse" | head; grep -E "usage_usec|nr_throttled|throttled_usec" /sys/fs/cgroup/cpu.stat'
      1 10.4 clickhouse-serv
   1249 44.2 stress-ng-cpu
   1250 44.2 stress-ng-cpu
usage_usec 143313148
nr_throttled 455
throttled_usec 58200837
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  stresschaos/clickhouse-chaos-exp-17 --timeout=2m
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-17 condition met
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/17-cpu-stress.yaml
stresschaos.chaos-mesh.org "clickhouse-chaos-exp-17" deleted from demo namespace
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Verify that CPU stress did not restart ClickHouse:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.status.containerStatuses[0].restartCount}{"\n"}'
4
```

Verify that the replication queue is empty:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT queue_size FROM system.replicas WHERE database='\''chaos_v2'\'' AND table='\''events_local'\''"'
0
```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1940 1835 105
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

Two CPU workers each consumed about 44% CPU. Cgroup throttling rose from 203 to 455 periods and from 23,817,473 to 58,200,837 microseconds. The target restart count remained 4, ClickHouse stayed `Ready`, the replication queue returned to 0, and the workload ended at `1940 1835 105`.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

#### Demonstrate impact and recovery

Before injection, confirm the database is healthy and record the cgroup
baseline:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c '
  printf "memory_current="; cat /sys/fs/cgroup/memory.current
  printf "memory_max="; cat /sys/fs/cgroup/memory.max
'
memory_current=1301417984
memory_max=4294967296
```

Record the restart count before memory stress:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
  -o jsonpath='{.status.containerStatuses[0].restartCount}{"\n"}'
3
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/18-memory-stress.yaml
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-18 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  stresschaos/clickhouse-chaos-exp-18 --timeout=90s
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-18 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c '
printf "memory_current="; cat /sys/fs/cgroup/memory.current
printf "memory_max="; cat /sys/fs/cgroup/memory.max
ps -o pid,stat,comm -p 1'
memory_current=2407194624
memory_max=4294967296
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

Confirm that the restart count did not change:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
  -o jsonpath='{.status.containerStatuses[0].restartCount}{"\n"}'
3
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  stresschaos/clickhouse-chaos-exp-18 --timeout=2m
stresschaos.chaos-mesh.org/clickhouse-chaos-exp-18 condition met
```

Delete the experiment and measure recovery:

```shell
➤ kubectl delete -f tests/18-memory-stress.yaml
stresschaos.chaos-mesh.org "clickhouse-chaos-exp-18" deleted from demo namespace
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c '
printf "memory_current="; cat /sys/fs/cgroup/memory.current
printf "memory_max="; cat /sys/fs/cgroup/memory.max
ps -o pid,stat,comm -p 1'
memory_current=1358225408
memory_max=4294967296
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
1978 1873 105
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

Before injection, memory usage was 1,301,417,984 bytes against a 4,294,967,296-byte limit. A 1GiB stress worker raised usage to 2,407,194,624 bytes without an OOM or restart. After cleanup it fell to 1,358,225,408 bytes, PID 1 remained `Ssl`, and the workload ended at `1978 1873 105`.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Measure the same read-only operation before injection:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c \
        '/usr/bin/time -f "elapsed=%e" sh -c "find /var/lib/clickhouse/store -type f | head -n 20 | xargs stat >/dev/null"'
elapsed=0.00
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/19-io-latency.yaml
iochaos.chaos-mesh.org/clickhouse-chaos-exp-19 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  iochaos/clickhouse-chaos-exp-19 --timeout=90s
iochaos.chaos-mesh.org/clickhouse-chaos-exp-19 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  findmnt -T /var/lib/clickhouse
TARGET              SOURCE FSTYPE OPTIONS
/var/lib/clickhouse toda   fuse   rw,nosuid,nodev,relatime,user_id=0,group_id=0
```

Run the read-only latency probe while the FUSE layer is active:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c \
        '/usr/bin/time -f "elapsed=%e" sh -c "find /var/lib/clickhouse/store -type f | head -n 20 | xargs stat >/dev/null"'
stat: cannot statx '/var/lib/clickhouse/store/...': Transport endpoint is not connected
Command exited with non-zero status 123
elapsed=5.78
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  iochaos/clickhouse-chaos-exp-19 --timeout=2m
iochaos.chaos-mesh.org/clickhouse-chaos-exp-19 condition met
```

Delete the recovered experiment:

```shell
➤ kubectl delete -f tests/19-io-latency.yaml
iochaos.chaos-mesh.org "clickhouse-chaos-exp-19" deleted from demo namespace
```

Check that Chaos Mesh removed its FUSE layer:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  mount | grep /var/lib/clickhouse
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

Confirm that PID 1 is still running:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

Wait for ClickHouse recovery after the fault has been removed:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing. Record the stable counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
2014 1890 124
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

IOChaos changed the mount from ext4 to `toda`; the read-only metadata scan changed from `0.00s` to `5.78s` and returned explicit transport errors. After `AllRecovered`, ext4 returned, the scan took `0.00s`, PID 1 was already `Ssl`, and the workload ended at `2014 1890 124`.

Result: **PASS** — the latency was directly visible and Chaos Mesh restored the mount and running process automatically.

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

`EIO` is the operating-system error for an input/output failure. It is
recoverable in this experiment because Chaos Mesh returns the error only while
the experiment is active; it does not deliberately corrupt stored bytes.

**Expected behavior:** ClickHouse should expose explicit disk errors rather
than silently accepting bad data. Some writes may fail. After injection ends,
the normal mount, SQL service, and equal replica checksums must return.

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

Record the workload counters before applying the experiment. The values are
attempted, acknowledged, and failed batches, in that order:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
2014 1890 124
```

#### Demonstrate impact and recovery

Before injection, confirm that ClickHouse is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Confirm that the selected replica is using its normal ext4-backed PVC:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  findmnt -T /var/lib/clickhouse
TARGET              SOURCE                                                                                                                              FSTYPE OPTIONS
/var/lib/clickhouse /dev/vda1[/var/lib/rancher/k3s/storage/pvc-eecf7db8-b9b5-4888-b1aa-71e70ea65119_demo_data-clickhouse-chaos-chaos-cluster-shard-1-0] ext4   rw,relatime,discard,errors=remount-ro,commit=30
```

Run the read-only probe against existing ClickHouse part files. Before fault
injection, it must finish successfully:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  sh -c 'find /var/lib/clickhouse/store -type f -exec stat {} + >/dev/null; echo exit_code=$?'
exit_code=0
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/20-io-fault.yaml
iochaos.chaos-mesh.org/clickhouse-chaos-exp-20 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  iochaos/clickhouse-chaos-exp-20 --timeout=90s
iochaos.chaos-mesh.org/clickhouse-chaos-exp-20 condition met
```

While `AllInjected=True`, inspect the data mount again:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  findmnt -T /var/lib/clickhouse
TARGET              SOURCE FSTYPE OPTIONS
/var/lib/clickhouse toda   fuse   rw,nosuid,nodev,relatime,user_id=0,group_id=0,default_permissions,allow_other
```

The `toda` FUSE source proves that Chaos Mesh has interposed its fault layer on
the selected mount. Run the same read-only probe again. It performs a metadata
lookup for every existing part file, giving the 10 percent fault enough
operations to encounter `EIO` without creating or changing data:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  sh -c 'find /var/lib/clickhouse/store -type f -exec stat {} + >/dev/null; echo exit_code=$?'
find: ‘/var/lib/clickhouse/store’: Input/output error
exit_code=1
```

This is the direct proof that the experiment worked: the same path was ext4
before injection, became the `toda` fault mount, and returned
`Input/output error` with a non-zero exit code. `AllInjected=True` alone would
only prove that Chaos Mesh attached the experiment, not that an operation
actually encountered the configured error.

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  iochaos/clickhouse-chaos-exp-20 --timeout=2m
iochaos.chaos-mesh.org/clickhouse-chaos-exp-20 condition met
```

#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing. Read the same three workload counters:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
2040 1903 137
```

During this test window, the client attempted 26 batches: 13 were
acknowledged and 13 failed or were ambiguous. We do not claim that every one of
those failures was an `EIO`; the read-only probe above proves the
configured storage error, while the workload counters prove that ClickHouse
experienced an availability impact during the same injection window.

Verify that the original filesystem is mounted again:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  findmnt -T /var/lib/clickhouse
TARGET              SOURCE                                                                                                                              FSTYPE OPTIONS
/var/lib/clickhouse /dev/vda1[/var/lib/rancher/k3s/storage/pvc-eecf7db8-b9b5-4888-b1aa-71e70ea65119_demo_data-clickhouse-chaos-chaos-cluster-shard-1-0] ext4   rw,relatime,discard,errors=remount-ro,commit=30
```

Verify that ClickHouse PID 1 is runnable. `Ssl` is a normal sleeping server
process; importantly, it does not contain `T`, which would mean stopped:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/20-io-fault.yaml
iochaos.chaos-mesh.org "clickhouse-chaos-exp-20" deleted from demo namespace
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Prove the same read-only operation succeeds again after recovery:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  sh -c 'find /var/lib/clickhouse/store -type f -exec stat {} + >/dev/null; echo exit_code=$?'
exit_code=0
```

The zero exit code proves that the selected filesystem became readable again
after the fault.

**Observed behavior:**

The `toda` mount returned a real `Input/output error` to the read-only probe,
and the workload recorded failed inserts during the fault window. After
`AllRecovered`, the PVC was ext4 again, PID 1 was `Ssl`, SQL responded, and the
same read-only traversal returned exit code 0. No signal or pod replacement
was required.

Result: **PASS** — the transient EIO was visible, ClickHouse recovered without
manual intervention, and the read-only recovery probe passed.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Prove the target can resolve Keeper before injection:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        getent hosts clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
10.42.0.39      clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/21-keeper-dns-error.yaml
dnschaos.chaos-mesh.org/clickhouse-chaos-exp-21 created
```

Confirm that Chaos Mesh reached the target:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  dnschaos/clickhouse-chaos-exp-21 --timeout=90s
dnschaos.chaos-mesh.org/clickhouse-chaos-exp-21 condition met
```

Observe the live impact:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  getent hosts clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
command terminated with exit code 2
```

The sibling is the control and must still resolve the same name:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- \
        sh -c 'getent hosts clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local; echo exit_code=$?'
10.42.0.39      clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
exit_code=0
```

Wait for Chaos Mesh to remove the fault:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  dnschaos/clickhouse-chaos-exp-21 --timeout=2m
dnschaos.chaos-mesh.org/clickhouse-chaos-exp-21 condition met
```

Delete the experiment:

```shell
➤ kubectl delete -f tests/21-keeper-dns-error.yaml
dnschaos.chaos-mesh.org "clickhouse-chaos-exp-21" deleted from demo namespace
```

Prove DNS recovery on the original target:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
        sh -c 'getent hosts clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local; echo exit_code=$?'
10.42.0.39      clickhouse-chaos-keeper-2.clickhouse-chaos-keeper-pods.demo.svc.cluster.local
exit_code=0
```

Wait for ClickHouse to report full recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```


#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing.

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
2079 1942 137
```

**Observed behavior:**

Before injection, the Keeper FQDN resolved to `10.42.0.39`. During DNSChaos, the target returned exit code 2 while the sibling resolved the same name with exit code 0. After recovery, the target resolved it again; the workload moved from `2040 1903 137` to `2079 1942 137`.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

An incorrect timezone and an incorrect clock are different conditions. A
wrong timezone normally changes only how local time is displayed; the
underlying UTC clock remains correct, and ClickHouse can continue functioning.
An actual two-hour clock error changes the time returned by `now()` and can
affect inserted timestamps, TTL processing, scheduled work, logs, certificate
validation, and other time-based behavior. This experiment changes one
ClickHouse process's real-time clock. It does not change the node timezone.

Record the target and control clocks before applying the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- bash -c \
        'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT now(), toUnixTimestamp(now())"'
2026-09-09 09:20:47	1788945647
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c \
        'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT now(), toUnixTimestamp(now())"'
2026-09-09 09:20:47	1788945647
```

Validate and apply the fault:

```shell
➤ kubectl apply --dry-run=server -f tests/22-clock-skew.yaml
timechaos.chaos-mesh.org/clickhouse-chaos-exp-22 created (server dry run)
```

```shell
➤ kubectl apply -f tests/22-clock-skew.yaml
timechaos.chaos-mesh.org/clickhouse-chaos-exp-22 created
```

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  -f tests/22-clock-skew.yaml --timeout=90s
timechaos.chaos-mesh.org/clickhouse-chaos-exp-22 condition met
```

The target SQL query did not return before the timeout:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  bash -c 'timeout 8 clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT now(), toUnixTimestamp(now())"; echo exit_code=$?'
exit_code=124
```

Inspect PID 1 to explain the timeout:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
        ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Tsl  clickhouse-serv
```

The control replica continued returning current time:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c \
        'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT now(), toUnixTimestamp(now())"'
2026-09-09 09:21:32	1788945692
```

Wait for the 45-second fault to finish:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  -f tests/22-clock-skew.yaml --timeout=150s
timechaos.chaos-mesh.org/clickhouse-chaos-exp-22 condition met
```

```shell
➤ kubectl delete -f tests/22-clock-skew.yaml
timechaos.chaos-mesh.org "clickhouse-chaos-exp-22" deleted from demo namespace
```

After `AllRecovered`, inspect PID 1:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Tsl  clickhouse-serv
```

The expected state does not contain `T`. If it does, Chaos Mesh did not fully
clean up the experiment. Resume the existing process and rerun the complete
gate:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  kill -CONT 1
```

`kill -CONT` printed nothing. Checking PID 1 again produced:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing.

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
2088 1947 140
```

**Observed behavior:**

Both replicas returned the same current timestamp before injection. After `AllInjected`, the target SQL call hung and PID 1 was `Tsl`; the control replica still returned current time. `AllRecovered=True` left PID 1 stopped, so `kill -CONT 1` was required to restore `Ssl`. This run did not produce a usable two-hour-skew observation.

Result: **CHAOS TOOL LIMITATION** — ClickHouse data recovered, but Chaos Mesh stopped the process instead of demonstrating a live skewed clock and required manual `SIGCONT` cleanup.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

#### Demonstrate impact and recovery

Before injection, confirm the database is healthy:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Ready
```

Apply this experiment:

```shell
➤ kubectl apply -f tests/23-io-plus-sibling-failure.yaml
iochaos.chaos-mesh.org/clickhouse-chaos-exp-23-io created
podchaos.chaos-mesh.org/clickhouse-chaos-exp-23-pod created
```

Confirm that the I/O fault reached replica-0:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  iochaos/clickhouse-chaos-exp-23-io --timeout=90s
iochaos.chaos-mesh.org/clickhouse-chaos-exp-23-io condition met
```

Confirm that the sibling failure reached replica-1:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-23-pod --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-23-pod condition met
```

Observe the live impact:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS
clickhouse-chaos   26.2.6    Critical
```

Confirm that IOChaos replaced the normal data mount with its `toda` FUSE
layer:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  findmnt -T /var/lib/clickhouse
TARGET              SOURCE FSTYPE OPTIONS
/var/lib/clickhouse toda   fuse   rw,nosuid,nodev,relatime,user_id=0,group_id=0
```

Prove that the sibling failure also reached its target:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c 'printf ok'
error: Internal error occurred: OCI runtime exec failed: exec: "bash": executable file not found in $PATH
```

The service had no healthy path for shard 0:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c \
        'timeout 8 clickhouse-client --host clickhouse-chaos.demo.svc --user "$CH_USER" --password "$CH_PASSWORD" --query "SELECT count() FROM chaos_v2.events"; echo exit_code=$?'
Code: 210. DB::NetException: Connection refused (clickhouse-chaos.demo.svc:9000). (NETWORK_ERROR)
exit_code=210
```

The workload counters during the combined fault were:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- bash -c '
printf "attempted="; cat /state/attempt_batches
printf "successful="; cat /state/success_batches
printf "failed="; cat /state/failed_batches'
attempted=2121
successful=1949
failed=172
```

Wait for the I/O fault duration to finish:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  iochaos/clickhouse-chaos-exp-23-io --timeout=2m
iochaos.chaos-mesh.org/clickhouse-chaos-exp-23-io condition met
```

Wait for the sibling failure duration to finish:

```shell
➤ kubectl wait -n demo --for=condition=AllRecovered \
  podchaos/clickhouse-chaos-exp-23-pod --timeout=2m
podchaos.chaos-mesh.org/clickhouse-chaos-exp-23-pod condition met
```

Delete the `PodChaos` first:

```shell
➤ kubectl delete podchaos -n demo clickhouse-chaos-exp-23-pod
podchaos.chaos-mesh.org "clickhouse-chaos-exp-23-pod" deleted from demo namespace
```

Delete the `IOChaos` second:

```shell
➤ kubectl delete iochaos -n demo clickhouse-chaos-exp-23-io
iochaos.chaos-mesh.org "clickhouse-chaos-exp-23-io" deleted from demo namespace
```

Check that the normal filesystem is mounted again:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  mount | grep /var/lib/clickhouse
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

Check the ClickHouse process state:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

Unlike experiments 19 and 22, this process was not stopped after cleanup, so
we did not run `kill -CONT 1`.

Compare the replicas after automatic recovery:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c \
        'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local; SELECT queue_size FROM system.replicas WHERE database='\''chaos_v2'\''"'
99019	99019	2073066991253321069
0
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c \
        'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local; SELECT queue_size FROM system.replicas WHERE database='\''chaos_v2'\''"'
99019	99019	2073066991253321069
0
```

Finally, wait for ClickHouse to return to `Ready`:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing.

Record the stable counters after recovery:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
2121 1949 172
```

After any in-flight batch finishes, run the mandatory recovery gate.

**Observed behavior:**

Shard-0 replica-0 had a `toda` latency mount and returned delayed transport errors while replica-1 rejected exec. The service query returned `Connection refused`, ClickHouse became `Critical`, and the workload ended at `2121 1949 172`. Cleanup restored ext4 and PID `Ssl`; both replicas matched at `99019` rows with queue 0.

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

#### Resume the workload

Discover the workload pod by its label:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Resume the workload before injecting the fault:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

The command prints nothing. Keep the workload running while observing the
fault and recovery transition.

Apply the files in numeric order. Each cycle must recover completely before
the next pod is killed.

#### Cycle 1: clickhouse-chaos-chaos-cluster-shard-1-0

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
326909c8-1fe4-4c3e-99f8-3c14f8680d4b
```

Inject the pod kill:

```shell
➤ kubectl apply -f tests/24-1-recovery-soak.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-1 created
```

Confirm injection:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-24-1 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-1 condition met
```

Delete this one-shot experiment:

```shell
➤ kubectl delete -f tests/24-1-recovery-soak.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-24-1" deleted from demo namespace
```

Wait for this replica:

```shell
➤ kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-1-0 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-1-0 condition met
```

Run the ClickHouse recovery gate before continuing:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Confirm the replacement UID:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
24daa5d8-08ab-48b8-aeec-d417b7f10d27
```

Verify cycle 1 with SQL on both replicas:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
101256	101256	6139974127982147424
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
101256	101256	6139974127982147424
```


#### Cycle 2: clickhouse-chaos-chaos-cluster-shard-0-0

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
add2a88c-ed3c-46f3-bbb4-b2cb61fec420
```

Inject the pod kill:

```shell
➤ kubectl apply -f tests/24-2-recovery-soak.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-2 created
```

Confirm injection:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-24-2 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-2 condition met
```

Delete this one-shot experiment:

```shell
➤ kubectl delete -f tests/24-2-recovery-soak.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-24-2" deleted from demo namespace
```

Wait for this replica:

```shell
➤ kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-0-0 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-0-0 condition met
```

Run the ClickHouse recovery gate before continuing:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Confirm the replacement UID:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
ce5234f0-f938-4927-9b3b-f9244ab22607
```

Verify cycle 2 with SQL on both replicas:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
101943	101943	14304192949753297155
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
101943	101943	14304192949753297155
```


#### Cycle 3: clickhouse-chaos-chaos-cluster-shard-1-1

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-1 \
  -o jsonpath='{.metadata.uid}{"\n"}'
d8da0cb3-168b-40ad-840b-ecb835c27cea
```

Inject the pod kill:

```shell
➤ kubectl apply -f tests/24-3-recovery-soak.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-3 created
```

Confirm injection:

```shell
➤ kubectl wait -n demo --for=condition=AllInjected \
  podchaos/clickhouse-chaos-exp-24-3 --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-exp-24-3 condition met
```

Delete this one-shot experiment:

```shell
➤ kubectl delete -f tests/24-3-recovery-soak.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-exp-24-3" deleted from demo namespace
```

Wait for this replica:

```shell
➤ kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-1-1 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-1-1 condition met
```

Run the ClickHouse recovery gate before continuing:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Confirm the replacement UID:

```shell
➤ kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-1-1 \
  -o jsonpath='{.metadata.uid}{"\n"}'
b4d71c1f-6045-4fe5-93e1-1856b9e511cf
```

Verify cycle 3 with SQL on both replicas:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
105543	105543	10953483407728974470
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
105543	105543	10953483407728974470
```



#### Pause the workload

After capturing the recovery transition, stop the workload from starting new
batches:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- sh -c 'paste -d " " /state/attempt_batches /state/success_batches /state/failed_batches'
2225 2053 172
```

After any in-flight batch finishes, run the
mandatory recovery gate and record the stable integrity result.

**Observed behavior:**

Three one-shot kills replaced shard-1 replica-0, shard-0 replica-0, and shard-1 replica-1. Their new UIDs were `24daa5d8-08ab-48b8-aeec-d417b7f10d27`, `ce5234f0-f938-4927-9b3b-f9244ab22607`, and `b4d71c1f-6045-4fe5-93e1-1856b9e511cf`. SQL, replica-state, and checksum checks passed between cycles; the workload ended at `2225 2053 172`.

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

#### Pause the workload

Pause the continuous workload so the baseline remains stable:

```shell
➤ kubectl get pods -n demo -l app=clickhouse-chaos-workload \
  -o jsonpath='{.items[0].metadata.name}{"\n"}'
clickhouse-chaos-workload-64d7d5c85f-fqjdp
```

Use the returned pod name to pause the workload:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing on success.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- bash -c '
if pgrep -x clickhouse-client >/dev/null; then
  echo "client still active"
else
  echo "workload paused"
fi'
workload paused
```

Record the workload counters accumulated across experiments 1–24:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- bash -c '
printf "attempted="; cat /state/attempt_batches
printf "successful="; cat /state/success_batches
printf "failed="; cat /state/failed_batches'
attempted=2225
successful=2053
failed=172
```

Check the Distributed table before deleting anything:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events FORMAT TSV"'
209804  209804  13047809195631774027
```

The row count is higher than `2053 × 100` because some timed-out Distributed
inserts reached ClickHouse even though the client did not receive a success
response. Equality between `count()` and `uniqExact(id)` proves those rows are
not duplicate IDs.

Before deleting anything, compare the automatically synchronized replicas
separately:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events_local FORMAT TSV"'
104261  104261  2094325787902799557
```

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events_local FORMAT TSV"'
104261  104261  2094325787902799557
```

Record the donor pod UID:

```shell
➤ kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 \
  -o jsonpath='{.metadata.uid}{"\n"}'
ce5234f0-f938-4927-9b3b-f9244ab22607
```

Record the target pod UID:

```shell
➤ kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{"\n"}'
0a753027-af14-428b-8b78-9192f190cfeb
```

Record the target PVC UID and PV:

```shell
➤ kubectl get pvc -n demo \
  data-clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{"\n"}{.spec.volumeName}{"\n"}'
af84e532-ec8b-4ddc-938f-34d6a21bdeeb
pvc-af84e532-ec8b-4ddc-938f-34d6a21bdeeb
```

Delete the target pod first:

```shell
➤ kubectl delete pod -n demo clickhouse-chaos-chaos-cluster-shard-0-1 --wait=false
pod "clickhouse-chaos-chaos-cluster-shard-0-1" deleted from demo namespace
```

Delete its PVC immediately afterward:

```shell
➤ kubectl delete pvc -n demo data-clickhouse-chaos-chaos-cluster-shard-0-1 --wait=false
persistentvolumeclaim "data-clickhouse-chaos-chaos-cluster-shard-0-1" deleted from demo namespace
```

Wait for the replacement pod:

```shell
➤ kubectl wait -n demo --for=create \
  pod/clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

Wait for the replacement PVC:

```shell
➤ kubectl wait -n demo --for=create \
  pvc/data-clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
persistentvolumeclaim/data-clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

Wait for the replacement pod to become Ready:

```shell
➤ kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-chaos-cluster-shard-0-1 --timeout=5m
pod/clickhouse-chaos-chaos-cluster-shard-0-1 condition met
```

The new pod UID is different:

```shell
➤ kubectl get pod -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{"\n"}'
ad93b11a-2e7a-492e-8aa3-e1f1048fe21d
```

The new PVC UID and PV are also different:

```shell
➤ kubectl get pvc -n demo \
  data-clickhouse-chaos-chaos-cluster-shard-0-1 \
  -o jsonpath='{.metadata.uid}{"\n"}{.spec.volumeName}{"\n"}'
11b51f42-bd68-483e-b537-e3bf349c40b3
pvc-11b51f42-bd68-483e-b537-e3bf349c40b3
```

The old PV no longer exists:

```shell
➤ kubectl get pv pvc-af84e532-ec8b-4ddc-938f-34d6a21bdeeb
Error from server (NotFound): persistentvolumes "pvc-af84e532-ec8b-4ddc-938f-34d6a21bdeeb" not found
```

Immediately after the replacement starts, prove that its new disk has no copy
of the workload table:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "EXISTS TABLE chaos_v2.events_local"'
0
```

This is the live impact: the pod is running, but its local table and data were
really lost. ClickHouse reported why it was not ready while automatic recovery
was in progress:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS     AGE
clickhouse-chaos   26.2.6    NotReady   68m
```

Wait for ClickHouse to report recovery:

```shell
➤ kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready clickhouse/clickhouse-chaos --timeout=5m
clickhouse.kubedb.com/clickhouse-chaos condition met
```

Compare the donor:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events_local FORMAT TSV"'
104261  104261  2094325787902799557
```

Compare the rebuilt replica:

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events_local FORMAT TSV"'
104261  104261  2094325787902799557
```

#### Resume and pause the workload

Resume the existing workload to prove new writes still work:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  rm -f /state/pause
```

Watch until the workload reports a successful post-rebuild batch:

```shell
➤ timeout 15 kubectl logs -n demo -f clickhouse-chaos-workload-64d7d5c85f-fqjdp --tail=0 | head -n 1
2026-09-09T09:32:37+00:00 success attempt=2226 rows=100
```

Pause it again for the final stable check:

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- \
  touch /state/pause
```

The command prints nothing on success.

```shell
➤ kubectl exec -n demo clickhouse-chaos-workload-64d7d5c85f-fqjdp -- bash -c '
printf "attempted="; cat /state/attempt_batches
printf "successful="; cat /state/success_batches
printf "failed="; cat /state/failed_batches'
attempted=2227
successful=2055
failed=172
```

```shell
➤ kubectl exec -n demo \
  clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- bash -c '
clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
  --query "SELECT count(), uniqExact(id), sum(payload)
           FROM chaos_v2.events FORMAT TSV"'
210004  210004  309272090909017131
```

**Observed behavior:** The same cluster retained all data accumulated during
experiments 1–24. KubeDB created a new 4Gi PVC and repaired the empty replica
from its sibling. The rebuilt replica matched the donor at 104,261 rows before
the workload resumed, and two additional 100-row batches succeeded
afterward.

Result: **PASS** — complete loss of one replica's pod and disk recovered
automatically from its sibling without manual schema or data repair.

## Chaos Testing Results Summary

| # | Fault | Fresh observed impact | Recovery |
| ---: | --- | --- | --- |
| 1 | Single replica pod kill | Target UID changed; ClickHouse stayed `Ready` | Both replicas matched at 2,319 rows |
| 2 | Replica pod failure, 45s | Direct query refused; target restarted twice; ClickHouse became `Critical` | SQL and `Ready` returned |
| 3 | ClickHouse container kill | Same pod UID; restart count 0 → 1 | Queue returned to 0 |
| 4 | Three alternating pod kills | Three target UIDs changed; one ambiguous attempt | SQL and checksum checks passed between kills |
| 5 | Both replicas of shard 0 failed | Direct queries refused; Distributed query returned `ALL_CONNECTION_TRIES_FAILED` | Both replicas matched at 72,274 rows |
| 6 | All four data pods failed | Complete SQL outage; 40 attempts and no acknowledgements | Four pods reopened their PVC data |
| 7 | Keeper follower kill | Follower UID changed; Keeper-0 remained leader | 26/26 attempts acknowledged |
| 8 | Keeper leader kill | Keeper-1 became leader | One leader and two followers restored |
| 9 | Keeper quorum loss, 90s | Survivor stopped serving; replicated insert timed out | Quorum and queue 0 returned |
| 10 | All Keeper members failed | All rejected exec; replicated insert timed out | Quorum reformed |
| 11 | 500ms network delay | Ping average rose to 517.919ms | Returned to 0.070ms; queue 0 |
| 12 | 30% packet loss | Probe measured 35% loss | Recovery probe measured 0% loss |
| 13 | 50% packet duplication | Probe reported 8 duplicate packets | Both replicas had 80,556 unique rows |
| 14 | 1Mbps bandwidth limit | 4.5MB transfer took 4.80s | Same transfer took 0.21s |
| 15 | Data-replica partition | Peer TCP timed out; replicas diverged; queue reached 12 | Automatically matched at 89,357 rows |
| 16 | Replica-to-Keeper partition | Target reported `is_readonly=1` and `is_session_expired=1` | Returned writable with two active replicas |
| 17 | CPU stress | Two workers used ~44% each; throttling 203 → 455 | No restart; queue 0 |
| 18 | 1GiB memory stress | Usage 1.30GB → 2.41GB under a 4GiB limit | Fell to 1.36GB; no OOM |
| 19 | 100ms filesystem latency | `toda` active; scan 0.00s → 5.78s with transport errors | ext4 and PID `Ssl` returned automatically |
| 20 | 10% EIO | Read-only probe returned `Input/output error` | Same probe returned exit 0; PID `Ssl` |
| 21 | Keeper DNS errors | Target lookup failed; control lookup succeeded | Target DNS resolved again |
| 22 | Clock skew −2h | Chaos Mesh stopped PID 1; live skew was not demonstrated | Manual `SIGCONT`; chaos-tool limitation |
| 23 | I/O latency plus sibling failure | `toda`, sibling exec failure, SQL refusal, `Critical` | Both replicas matched at 99,019 rows |
| 24 | Three-cycle recovery soak | Three target UIDs changed | SQL and checksum checks passed every cycle |
| 25 | Existing replica and PVC deletion | New pod/PVC/PV; `EXISTS TABLE` initially returned 0 | 104,261 rows restored; new writes succeeded |

## Final Integrity Evidence

The same cluster used by all 25 experiments finished healthy. Check the
ClickHouse resource first:

```shell
➤ kubectl get clickhouse -n demo clickhouse-chaos
NAME               VERSION   STATUS   AGE
clickhouse-chaos   26.2.6    Ready    105m
```

Check every database and workload pod:

```shell
➤ kubectl get pods -n demo | grep clickhouse-chaos
clickhouse-chaos-chaos-cluster-shard-0-0      1/1   Running   0             39m
clickhouse-chaos-chaos-cluster-shard-0-1      1/1   Running   0             37m
clickhouse-chaos-chaos-cluster-shard-1-0      1/1   Running   0             40m
clickhouse-chaos-chaos-cluster-shard-1-1      1/1   Running   1 (39m ago)   39m
clickhouse-chaos-keeper-0                     1/1   Running   6 (61m ago)   66m
clickhouse-chaos-keeper-1                     1/1   Running   2 (61m ago)   67m
clickhouse-chaos-keeper-2                     1/1   Running   6 (61m ago)   105m
clickhouse-chaos-workload-64d7d5c85f-fqjdp   1/1   Running   0             103m
```

The Distributed table contained 210,004 rows and every UUID was unique:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id) FROM chaos_v2.events"'
210004  210004
```

Compare the count, unique-ID count, and checksum on each replica. Matching
results within each shard prove that its replicas converged automatically:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
104352  104352  403484691276693588
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
104352  104352  403484691276693588
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
105652  105652  18352531473341875159
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT count(), uniqExact(id), sum(payload) FROM chaos_v2.events_local"'
105652  105652  18352531473341875159
```

Check replica health separately on all four data pods. The six columns are
`is_readonly`, `is_session_expired`, `queue_size`, `total_replicas`,
`active_replicas`, and `absolute_delay`:

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT is_readonly, is_session_expired, queue_size, total_replicas, active_replicas, absolute_delay FROM system.replicas WHERE database = '\''chaos_v2'\'' AND table = '\''events_local'\''"'
0  0  0  2  2  0
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-0-1 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT is_readonly, is_session_expired, queue_size, total_replicas, active_replicas, absolute_delay FROM system.replicas WHERE database = '\''chaos_v2'\'' AND table = '\''events_local'\''"'
0  0  0  2  2  0
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-0 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT is_readonly, is_session_expired, queue_size, total_replicas, active_replicas, absolute_delay FROM system.replicas WHERE database = '\''chaos_v2'\'' AND table = '\''events_local'\''"'
0  0  0  2  2  0
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-chaos-cluster-shard-1-1 -c clickhouse -- \
  bash -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "SELECT is_readonly, is_session_expired, queue_size, total_replicas, active_replicas, absolute_delay FROM system.replicas WHERE database = '\''chaos_v2'\'' AND table = '\''events_local'\''"'
0  0  0  2  2  0
```

Check each Keeper member separately:

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-0 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
leader
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-1 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
follower
```

```shell
➤ kubectl exec -n demo clickhouse-chaos-keeper-2 -c clickhouse-keeper -- bash -c 'exec 3<>/dev/tcp/127.0.0.1/9181; printf "mntr\n" >&3; timeout 3 cat <&3' | awk '$1=="zk_server_state" {print $2}'
follower
```

Verify that all data PVCs are 4Gi and all Keeper PVCs are 1Gi:

```shell
➤ kubectl get pvc -n demo | grep clickhouse-chaos
data-clickhouse-chaos-chaos-cluster-shard-0-0   Bound   pvc-9773c0e0-8bd6-4a67-b445-444255b61074   4Gi   RWO   local-path   <unset>   105m
data-clickhouse-chaos-chaos-cluster-shard-0-1   Bound   pvc-11b51f42-bd68-483e-b537-e3bf349c40b3   4Gi   RWO   local-path   <unset>   37m
data-clickhouse-chaos-chaos-cluster-shard-1-0   Bound   pvc-eecf7db8-b9b5-4888-b1aa-71e70ea65119   4Gi   RWO   local-path   <unset>   105m
data-clickhouse-chaos-chaos-cluster-shard-1-1   Bound   pvc-35ef4112-0c9d-42b1-973a-f26389b04864   4Gi   RWO   local-path   <unset>   105m
data-clickhouse-chaos-keeper-0                  Bound   pvc-531f098d-5196-490d-b6ba-b3584755b06c   1Gi   RWO   local-path   <unset>   105m
data-clickhouse-chaos-keeper-1                  Bound   pvc-41c5eecc-1f51-477d-aa7f-22e2e1c88ba9   1Gi   RWO   local-path   <unset>   105m
data-clickhouse-chaos-keeper-2                  Bound   pvc-e893bff9-8ba5-4da4-92f7-b0b125e4777f   1Gi   RWO   local-path   <unset>   105m
```

Finally, prove that no experiment remains active:

```shell
➤ kubectl get podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos -n demo
No resources found in demo namespace.
```

The final workload counters were 2,227 attempted batches, 2,055 acknowledged
batches, and 172 failed or ambiguous attempts. The database contained 210,004
unique rows. The 4,504 rows above `2,055 × 100` came from inserts for which
ClickHouse accepted data but the client did not receive a success response
before its timeout. They are not duplicate IDs.

## What the Errors Mean

- **Client timeout or connection failure:** the selected shard, all data pods,
  Keeper, network, or storage was temporarily unavailable. Failed attempts are
  expected during those tests.
- **`Critical` ClickHouse phase:** at least part of the desired cluster was not
  healthy. It does not always mean every query is unavailable.
- **ClickHouse stayed `Ready` during Keeper loss:** the health check could still
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
- **PID state `Tsl`:** the TimeChaos run left ClickHouse stopped even after
  Chaos Mesh reported recovery. `SIGCONT` resumed the
  process without deleting data or recreating the pod. This indicates an
  incomplete chaos-tool cleanup, not data corruption.

## Key Findings

1. Replica and Keeper redundancy protected acknowledged data across every
   tested recoverable failure.
2. ClickHouse `Ready` is useful but insufficient for Keeper-specific faults. A real
   database query, replica state, and Keeper `mntr` must be checked.
3. Unique row IDs are essential because a timed-out Distributed insert can
   have an ambiguous partial result.
4. Recovery verification needs consecutive matching replica checks, not one
   instantaneous queue sample.
5. Chaos Mesh 2.8.4 handled PodChaos, NetworkChaos, StressChaos, DNSChaos, and
   both I/O cases cleanly. TimeChaos stopped the target and required manual
   `SIGCONT`, so this run did not validate a functioning live clock skew.
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

Experiment 22 separates database resilience from chaos-tool behavior.
ClickHouse data recovered, but Chaos Mesh 2.8.4 stopped the affected process
and did not resume it automatically. Because a live two-hour skew was not
observed, that experiment is reported as a chaos-tool limitation, not a pass.

## What Next?

Repeat the same suite on a multi-node cluster with production-equivalent
storage, workload volume, resource limits, and monitoring. NodeChaos can then
be added safely because losing one worker will not also remove the Kubernetes
control plane and the chaos controller.

## Optional Cleanup

The tested cluster was kept running after the final evidence was captured. If
you created the cluster only for this campaign, first confirm that no chaos
experiment remains active:

```shell
➤ kubectl get podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos -n demo
No resources found in demo namespace.
```

Then remove the workload resources and the ClickHouse resource:

```shell
➤ kubectl delete deployment -n demo clickhouse-chaos-workload --ignore-not-found
```

```shell
➤ kubectl delete configmap -n demo clickhouse-chaos-workload --ignore-not-found
```

```shell
➤ kubectl delete clickhouse -n demo clickhouse-chaos --ignore-not-found
```

The test manifest uses `deletionPolicy: WipeOut`, so KubeDB also removes the
managed pods, generated authentication Secret, and PVCs. Do not delete a PVC
manually while its pod is terminating. Verify cleanup with:

```shell
➤ kubectl get clickhouse,petset,pods,pvc,secrets -n demo -o name | grep clickhouse-chaos
```

The final command should print nothing. These cleanup commands are intentionally
shown without captured deletion output because they were not run against the
cluster used to produce this article.
