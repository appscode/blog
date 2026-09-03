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
helm upgrade --install chaos-mesh chaos-mesh/chaos-mesh \
  --namespace chaos-mesh \
  --create-namespace \
  --version 2.8.4 \
  --set dashboard.create=true \
  --set dashboard.securityMode=false \
  --set dnsServer.create=true \
  --set chaosDaemon.runtime=containerd \
  --set chaosDaemon.socketPath=/run/k3s/containerd/containerd.sock \
  --set chaosDaemon.privileged=true
```

Output from our installation:

```text
NAME: chaos-mesh
NAMESPACE: chaos-mesh
STATUS: deployed
REVISION: 2
DESCRIPTION: Upgrade complete
```

If you use a different Kubernetes distribution, change the runtime and socket
path to match it.

Create the disposable namespace:

```bash
kubectl create ns demo
```

Output:

```text
namespace/demo created
```

Create the local manifest directories:

```bash
mkdir -p clickhouse-chaos-testing/setup
mkdir -p clickhouse-chaos-testing/tests
cd clickhouse-chaos-testing
```

Output: none. These shell commands completed successfully.

Confirm that the ClickHouse version and storage class used by this guide are
available:

```bash
kubectl get clickhouseversion 26.2.6
```

Output from our cluster:

```text
NAME     VERSION   DB_IMAGE
26.2.6   26.2.6    docker.io/clickhouse/clickhouse-server:26.2.6
```

```bash
kubectl get storageclass local-path
```

```text
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer
```

If your cluster does not provide `local-path`, replace
`storageClassName: local-path` in the setup manifest with a storage class that
supports `ReadWriteOnce` volumes in your cluster.

## Verify KubeDB and Chaos Mesh Installation

```bash
kubectl get pods -n kubedb
```

Output from our cluster:

```text
NAME                                            READY   STATUS    RESTARTS   AGE
kubedb-kubedb-autoscaler-0                      1/1     Running   2          15d
kubedb-kubedb-ops-manager-0                     1/1     Running   0          2d2h
kubedb-kubedb-provisioner-0                     1/1     Running   0          2d2h
kubedb-kubedb-webhook-server-65949766c4-xlfpz   1/1     Running   0          6d4h
kubedb-petset-85c9d79865-lfqww                  1/1     Running   2          15d
kubedb-sidekick-86f897c579-djk65                1/1     Running   2          15d
```

```bash
kubectl get pods -n chaos-mesh
```

```text
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7ddc79b6dc-65qr6   1/1     Running   1          4d23h
chaos-controller-manager-7ddc79b6dc-ksfxm   1/1     Running   1          4d23h
chaos-controller-manager-7ddc79b6dc-lb8qr   1/1     Running   1          4d23h
chaos-daemon-vmc4j                          1/1     Running   0          4d23h
chaos-dashboard-5db97f969f-prb94            1/1     Running   0          5d
chaos-dns-server-6d8fd4b8b5-6lws4           1/1     Running   0          5d
```

```bash
kubectl get crd \
  podchaos.chaos-mesh.org \
  networkchaos.chaos-mesh.org \
  stresschaos.chaos-mesh.org \
  iochaos.chaos-mesh.org \
  dnschaos.chaos-mesh.org \
  timechaos.chaos-mesh.org
```

```text
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

All 25 experiments preserved ClickHouse data and eventually passed their
availability and integrity gates. Experiments 19, 22, and 23 each required one
manual `SIGCONT` command because Chaos Mesh did not resume the target process
while cleaning up IOChaos or TimeChaos. These were Chaos Mesh cleanup issues,
not ClickHouse data failures. The first 24 experiments finished with 50,987
rows and 50,987 unique IDs. Experiment 25 used a separate fresh 100,000-row
cluster to test complete loss of one replica's PVC. In both campaigns, the
replicas of each shard finished with identical row counts and payload
checksums, replication queues were empty, and Keeper had exactly one leader
and two followers.

## Create a ClickHouse Cluster

The test topology deliberately separates two different kinds of redundancy:

- Two replicas protect each data shard.
- Three Keeper members provide a coordination quorum.

Save the following manifest as `setup/clickhouse-chaos-v2.yaml`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-chaos-v2
  namespace: demo
  labels:
    chaos-test.kubedb.com/suite: clickhouse-chaos-v2
spec:
  version: 26.2.6
  clusterTopology:
    cluster:
      name: chaos-v2-cluster
      shards: 2
      replicas: 2
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
            - name: clickhouse
              resources:
                requests:
                  cpu: 250m
                  memory: 512Mi
                limits:
                  cpu: "1"
                  memory: 1Gi
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
kubectl apply -f setup/clickhouse-chaos-v2.yaml
```

```text
clickhouse.kubedb.com/clickhouse-chaos-v2 created
```

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos-v2 --timeout=15m
```

```text
clickhouse.kubedb.com/clickhouse-chaos-v2 condition met
```

```bash
kubectl get clickhouse,petset,pods,pvc -n demo
```

Output from our fresh deployment, with unrelated `demo` resources omitted:

```text
NAME                  VERSION   STATUS   AGE
clickhouse-chaos-v2   26.2.6    Ready    84s

NAME                                                                        AGE
petset.apps.k8s.appscode.com/clickhouse-chaos-v2-chaos-v2-cluster-shard-0   78s
petset.apps.k8s.appscode.com/clickhouse-chaos-v2-chaos-v2-cluster-shard-1   76s
petset.apps.k8s.appscode.com/clickhouse-chaos-v2-keeper                     81s

NAME                                               READY   STATUS    RESTARTS   AGE
clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0   1/1   Running   0   78s
clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1   1/1   Running   0   74s
clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0   1/1   Running   0   76s
clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1   1/1   Running   0   70s
clickhouse-chaos-v2-keeper-0                     1/1   Running   0   80s
clickhouse-chaos-v2-keeper-1                     1/1   Running   0   76s
clickhouse-chaos-v2-keeper-2                     1/1   Running   0   71s

NAME                                                                        STATUS   CAPACITY   STORAGECLASS
persistentvolumeclaim/data-clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0   Bound    1Gi        local-path
persistentvolumeclaim/data-clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1   Bound    1Gi        local-path
persistentvolumeclaim/data-clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0   Bound    1Gi        local-path
persistentvolumeclaim/data-clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1   Bound    1Gi        local-path
persistentvolumeclaim/data-clickhouse-chaos-v2-keeper-0                     Bound    1Gi        local-path
persistentvolumeclaim/data-clickhouse-chaos-v2-keeper-1                     Bound    1Gi        local-path
persistentvolumeclaim/data-clickhouse-chaos-v2-keeper-2                     Bound    1Gi        local-path
```

KubeDB creates and references `clickhouse-chaos-v2-auth` automatically:

```bash
kubectl get secret -n demo clickhouse-chaos-v2-auth
```

```text
NAME                       TYPE                       DATA   AGE
clickhouse-chaos-v2-auth   kubernetes.io/basic-auth   2      118s
```

### Test Environment

| Component | Value |
| --- | --- |
| Kubernetes | K3s v1.36.3+k3s1, single node |
| Namespace | `demo` |
| ClickHouse resource | `clickhouse-chaos-v2` |
| ClickHouse version | `26.2.6` |
| Data topology | 2 shards × 2 replicas |
| Coordination | 3 ClickHouse Keeper members |
| Storage | 1 GiB `local-path` PVC per pod |
| ClickHouse limit | 1 CPU, 1 GiB memory |
| Keeper limit | 500m CPU, 512 MiB memory |
| Chaos Mesh | 2.8.4 |
| Container runtime | K3s containerd |
| Chaos daemon socket | `/run/k3s/containerd/containerd.sock` |

This campaign used a disposable single-node K3s cluster. NodeChaos was
excluded because failing the only
K3s node would also remove the control plane and Chaos Mesh. IOChaos `mistake`
was excluded because it deliberately returns incorrect bytes and can cause
silent persistent corruption.

## Chaos Testing

We kept a write client active during the experiments so that availability
failures and ambiguous writes were visible instead of being hidden by an idle
database.

### ClickHouse High-Write Load Client

ClickHouse is a column-oriented analytical database with a SQL interface. We
use database queries to create tables, insert test data, count unique IDs, and
inspect replication state.

The local table used `ReplicatedMergeTree`, while clients wrote through a
`Distributed` table across `chaos-v2-cluster`. Each batch inserted 100 rows
with server-generated UUIDs and `insert_distributed_sync=1`.

A UUID is a randomly generated unique identifier assigned to each row. Unique
IDs let us detect accidental duplicates by checking whether the total row
count equals the number of unique IDs.

Create the schema from one data pod. The operator already injects
`CLICKHOUSE_USER` and `CLICKHOUSE_PASSWORD` into the ClickHouse container, so
no credential needs to be copied into the command:

```bash
kubectl exec -i -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  bash -c '
    clickhouse-client \
      --user "$CLICKHOUSE_USER" \
      --password "$CLICKHOUSE_PASSWORD" \
      --multiquery
  ' <<'SQL'
CREATE DATABASE IF NOT EXISTS chaos_v2 ON CLUSTER `chaos-v2-cluster`;

CREATE TABLE IF NOT EXISTS chaos_v2.events_local
ON CLUSTER `chaos-v2-cluster`
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
ON CLUSTER `chaos-v2-cluster`
AS chaos_v2.events_local
ENGINE = Distributed(
  'chaos-v2-cluster',
  'chaos_v2',
  'events_local',
  cityHash64(id)
);
SQL
```

The command printed nothing and exited successfully; ClickHouse created the
database and both tables on the cluster.

Save this YAML as `setup/clickhouse-workload.yaml`. The client inserts 100 rows,
waits for synchronous Distributed delivery, records whether the batch was
acknowledged, then repeats:

The `while true` below belongs inside the workload container's script. It is
not a loop that the reader runs manually; Kubernetes starts this script once
and it keeps producing traffic until the workload is paused.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: clickhouse-chaos-v2-workload
  namespace: demo
  labels:
    chaos-test.kubedb.com/suite: clickhouse-chaos-v2
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
        --host clickhouse-chaos-v2.demo.svc \
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
  name: clickhouse-chaos-v2-workload
  namespace: demo
  labels:
    chaos-test.kubedb.com/suite: clickhouse-chaos-v2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: clickhouse-chaos-v2-workload
  template:
    metadata:
      labels:
        app: clickhouse-chaos-v2-workload
        chaos-test.kubedb.com/suite: clickhouse-chaos-v2
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
                  name: clickhouse-chaos-v2-auth
                  key: username
            - name: CH_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: clickhouse-chaos-v2-auth
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
            name: clickhouse-chaos-v2-workload
            defaultMode: 365
        - name: state
          emptyDir: {}
```

Start the client and wait for at least ten successful batches:

```bash
kubectl apply -f setup/clickhouse-workload.yaml
```

```text
configmap/clickhouse-chaos-v2-workload created
deployment.apps/clickhouse-chaos-v2-workload created
```

```bash
kubectl rollout status -n demo \
  deployment/clickhouse-chaos-v2-workload --timeout=3m
```

```text
deployment "clickhouse-chaos-v2-workload" successfully rolled out
```

```bash
workload_pod=$(kubectl get pod -n demo \
  -l app=clickhouse-chaos-v2-workload \
  -o jsonpath='{.items[0].metadata.name}')
```

Output: none. The variable contains the workload pod name.

```bash
kubectl logs -n demo -f "$workload_pod"
```

Output from our fresh deployment:

```text
2026-09-02T06:16:43+00:00 success attempt=2 rows=100
2026-09-02T06:16:44+00:00 success attempt=3 rows=100
2026-09-02T06:16:53+00:00 success attempt=10 rows=100
2026-09-02T06:16:56+00:00 success attempt=13 rows=100
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

Before the first test, the client had 42 successful batches, zero failures,
and 4,201 unique rows including one gate-probe row. At the end it had:

```text
attempted batches:     611
acknowledged batches:  491
failed/ambiguous:      120
acknowledged rows:     49,100
actual rows:           50,987
unique IDs:            50,987
```

The client attempted 611 batches. It received success for 491 batches, so
`491 × 100 = 49,100` rows were definitely acknowledged. Another 120 attempts
timed out or returned an error. Some rows from those attempts had already
reached ClickHouse, and the recovery checks also inserted small probe rows.
Together they account for the additional 1,887 rows. Since all 50,987 IDs
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
workload_pod=$(kubectl get pod -n demo \
  -l app=clickhouse-chaos-v2-workload \
  -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n demo "$workload_pod" -- touch /state/pause
sleep 5
kubectl exec -n demo "$workload_pod" -- bash -c \
  'if pgrep -x clickhouse-client >/dev/null; then echo "client still active"; else echo "workload paused"; fi'
```

Output from our cluster:

```text
workload paused
```

Require KubeDB and all seven database pods to be healthy:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready \
  clickhouse/clickhouse-chaos-v2 --timeout=10m
```

```text
clickhouse.kubedb.com/clickhouse-chaos-v2 condition met
```

```bash
kubectl get clickhouse,petset,pods -n demo
```

Output from our cluster, with unrelated `demo` resources omitted:

```text
NAME                                           VERSION   STATUS   AGE
clickhouse.kubedb.com/clickhouse-chaos-v2      26.2.6    Ready    84s

NAME                                                                        AGE
petset.apps.k8s.appscode.com/clickhouse-chaos-v2-chaos-v2-cluster-shard-0   78s
petset.apps.k8s.appscode.com/clickhouse-chaos-v2-chaos-v2-cluster-shard-1   76s
petset.apps.k8s.appscode.com/clickhouse-chaos-v2-keeper                     81s

NAME                                                     READY   STATUS    RESTARTS   AGE
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0       1/1     Running   0          78s
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1       1/1     Running   0          74s
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0       1/1     Running   0          76s
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1       1/1     Running   0          70s
pod/clickhouse-chaos-v2-keeper-0                         1/1     Running   0          80s
pod/clickhouse-chaos-v2-keeper-1                         1/1     Running   0          76s
pod/clickhouse-chaos-v2-keeper-2                         1/1     Running   0          71s
```

Confirm that no test fault remains:

```bash
kubectl get podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos \
  -n demo
```

Output from our cluster:

```text
No resources found in demo namespace.
```

Check the Distributed table. The first two values must match, and the count
must be at least `successful_batches × 100`:

```bash
kubectl exec -n demo "$workload_pod" -- bash -c '
  clickhouse-client \
    --host clickhouse-chaos-v2.demo.svc \
    --user "$CH_USER" \
    --password "$CH_PASSWORD" \
    --query "INSERT INTO chaos_v2.events
      SELECT generateUUIDv4(), now64(3), rand64()
      SETTINGS insert_distributed_sync=1"
'
```

The insert returns no text when ClickHouse accepts it.

```bash
kubectl exec -n demo "$workload_pod" -- bash -c '
  clickhouse-client \
    --host clickhouse-chaos-v2.demo.svc \
    --user "$CH_USER" \
    --password "$CH_PASSWORD" \
    --query "SELECT count(), uniqExact(id), sum(payload)
             FROM chaos_v2.events FORMAT TSV"
'
```

Output from the fresh reader-validation cluster:

```text
4201	4201	6381530584011999872
```

Check every local replica:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- bash -c '
  clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id), sum(payload)
             FROM chaos_v2.events_local FORMAT TSV"
'
```

Output:

```text
2100	2100	14543385160371786853
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1 -c clickhouse -- bash -c '
  clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id), sum(payload)
             FROM chaos_v2.events_local FORMAT TSV"
'
```

Output:

```text
2100	2100	14543385160371786853
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 -c clickhouse -- bash -c '
  clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id), sum(payload)
             FROM chaos_v2.events_local FORMAT TSV"
'
```

Output:

```text
2101	2101	10284889497349764635
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 -c clickhouse -- bash -c '
  clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id), sum(payload)
             FROM chaos_v2.events_local FORMAT TSV"
'
```

Output:

```text
2101	2101	10284889497349764635
```

Shard-0's two lines match, and shard-1's two lines match. Wait five seconds
and run the same four commands again. The second check must return the same
four lines before you continue.

Now check `system.replicas` on each pod, one at a time:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- bash -c '
  clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --database chaos_v2 \
    --query "SELECT is_readonly, queue_size, total_replicas, active_replicas,
                    lost_part_count, absolute_delay
             FROM system.replicas WHERE database=currentDatabase() FORMAT TSV"
'
```

Output:

```text
0	0	2	2	0	0
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1 -c clickhouse -- bash -c '
  clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --database chaos_v2 \
    --query "SELECT is_readonly, queue_size, total_replicas, active_replicas,
                    lost_part_count, absolute_delay
             FROM system.replicas WHERE database=currentDatabase() FORMAT TSV"
'
```

Output:

```text
0	0	2	2	0	0
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 -c clickhouse -- bash -c '
  clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --database chaos_v2 \
    --query "SELECT is_readonly, queue_size, total_replicas, active_replicas,
                    lost_part_count, absolute_delay
             FROM system.replicas WHERE database=currentDatabase() FORMAT TSV"
'
```

Output:

```text
0	0	2	2	0	0
```

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 -c clickhouse -- bash -c '
  clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
    --database chaos_v2 \
    --query "SELECT is_readonly, queue_size, total_replicas, active_replicas,
                    lost_part_count, absolute_delay
             FROM system.replicas WHERE database=currentDatabase() FORMAT TSV"
'
```

Output:

```text
0	0	2	2	0	0
```

The values mean writable, empty queue, two configured replicas, two active
replicas, no lost parts, and no replication delay.

Check Keeper directly rather than inferring quorum from ClickHouse status:

```bash
kubectl exec -n demo clickhouse-chaos-v2-keeper-0 \
  -c clickhouse-keeper -- bash -c '
  exec 3<>/dev/tcp/127.0.0.1/9181
  printf "mntr\n" >&3
  timeout 3 cat <&3
' | awk '$1=="zk_server_state" {print $2}'
```

Output:

```text
leader
```

```bash
kubectl exec -n demo clickhouse-chaos-v2-keeper-1 \
  -c clickhouse-keeper -- bash -c '
  exec 3<>/dev/tcp/127.0.0.1/9181
  printf "mntr\n" >&3
  timeout 3 cat <&3
' | awk '$1=="zk_server_state" {print $2}'
```

Output:

```text
follower
```

```bash
kubectl exec -n demo clickhouse-chaos-v2-keeper-2 \
  -c clickhouse-keeper -- bash -c '
  exec 3<>/dev/tcp/127.0.0.1/9181
  printf "mntr\n" >&3
  timeout 3 cat <&3
' | awk '$1=="zk_server_state" {print $2}'
```

Output:

```text
follower
```

The leader can change, but the result must contain exactly one leader and two
followers. Finally, verify each ClickHouse PID 1 one at a time:

```bash
kubectl exec -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 \
  -c clickhouse -- ps -o pid,stat,comm -p 1
```

```text
PID STAT COMMAND
  1 Ssl  clickhouse-serv
```

```bash
kubectl exec -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1 \
  -c clickhouse -- ps -o pid,stat,comm -p 1
```

```text
PID STAT COMMAND
  1 Ssl  clickhouse-serv
```

```bash
kubectl exec -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 \
  -c clickhouse -- ps -o pid,stat,comm -p 1
```

```text
PID STAT COMMAND
  1 Ssl  clickhouse-serv
```

```bash
kubectl exec -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 \
  -c clickhouse -- ps -o pid,stat,comm -p 1
```

```text
PID STAT COMMAND
  1 Ssl  clickhouse-serv
```

Check each data mount separately:

```bash
kubectl exec -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 \
  -c clickhouse -- mount | grep /var/lib/clickhouse
```

```text
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

```bash
kubectl exec -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1 \
  -c clickhouse -- mount | grep /var/lib/clickhouse
```

```text
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

```bash
kubectl exec -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 \
  -c clickhouse -- mount | grep /var/lib/clickhouse
```

```text
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

```bash
kubectl exec -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 \
  -c clickhouse -- mount | grep /var/lib/clickhouse
```

```text
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

PID 1 must not contain state `T`, and data must be on the normal filesystem,
not a `toda` FUSE mount. These checks matter because `Ready` alone cannot
prove Keeper availability, replica equality, or complete Chaos Mesh cleanup.

### Running Each Chaos Experiment

For every test, we used the same safe sequence:

```bash
workload_pod=$(kubectl get pod -n demo \
  -l app=clickhouse-chaos-v2-workload \
  -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n demo "$workload_pod" -- rm -f /state/pause
```

These commands print nothing on success. Validate the manifest against the
API server:

```bash
kubectl apply --dry-run=server -f tests/01-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-01 created (server dry run)
```

Inject the fault:

```bash
kubectl apply -f tests/01-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-01 created
```

Prove that Chaos Mesh injected it:

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/01-pod-kill.yaml --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-01 condition met
```

Observe ClickHouse and the pods:

```bash
kubectl get clickhouse,petset,pods -n demo
```

Output from our recovered cluster, with unrelated `demo` resources omitted:

```text
NAME                                        VERSION   STATUS   AGE
clickhouse.kubedb.com/clickhouse-chaos-v2   26.2.6    Ready    69m

NAME                                                          READY   STATUS    RESTARTS   AGE
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0            1/1     Running   0          10m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1            1/1     Running   6          49m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0            1/1     Running   0          11m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1            1/1     Running   0          10m
pod/clickhouse-chaos-v2-keeper-0                              1/1     Running   4          43m
pod/clickhouse-chaos-v2-keeper-1                              1/1     Running   4          45m
pod/clickhouse-chaos-v2-keeper-2                              1/1     Running   2          69m
```

Read the workload counters:

```bash
kubectl exec -n demo "$workload_pod" -- bash -c '
  printf "attempts="; cat /state/attempt_batches
  printf "success="; cat /state/success_batches
  printf "failed="; cat /state/failed_batches
'
```

Output from test 1 before its recovery gate:

```text
attempts=47
success=47
failed=0
```

Remove the experiment:

```bash
kubectl delete -f tests/01-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-01" deleted from demo namespace
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
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-02 condition met
```

```bash
kubectl delete -f tests/02-pod-failure.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-02" deleted from demo namespace
```

For one-shot `pod-kill` and `container-kill` tests, delete the Chaos object
after `AllInjected` and after proving the replacement pod or container is
healthy; those tests do not have a duration to wait for.

The `AllInjected` check matters. For example, our first DNS test used a
pattern that did not match the full name queried by the resolver. Treating
manifest creation as proof of injection would have produced a false pass.

## Pod and Keeper Chaos

### Chaos#1: Kill One ClickHouse Replica

Save this YAML as `tests/01-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-01
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
  gracePeriod: 0
```

What this chaos does: Abruptly deletes shard-0 replica-0 with no graceful
shutdown.

**Expected behavior:** The sibling replica should keep the shard available,
PetSet should create a replacement pod, and replication should catch it up.
The target pod UID must change, but acknowledged data must not.

Record the target UID before applying the manifest, then confirm that it
changes after injection.

**Observed behavior:**

Chaos Mesh killed shard-0 replica-0 with zero grace period. Its pod UID
changed, proving that the old pod was actually removed and replaced. KubeDB
remained `Ready`; the writer had five successful batches and no failures.
The replacement caught up and the complete gate passed.

Result: **PASS** — a single replica can disappear without losing acknowledged
data.

### Chaos#2: Hold One Replica Failed

Save this YAML as `tests/02-pod-failure.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-02
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1
```

What this chaos does: Makes shard-0 replica-1 continuously unavailable for
45 seconds instead of allowing Kubernetes to replace it immediately.

**Expected behavior:** KubeDB should report a degraded state while the sibling
replica continues serving the shard. When the fault ends, the same pod should
become reachable and converge without manual repair.

**Observed behavior:**

`pod-failure` kept shard-0 replica-1 unavailable for 45 seconds. KubeDB still
reported `Ready` at the 15-second sample, showing that status can lag a
container-level failure. The workload recorded 27 successes and ten failed or
ambiguous attempts while the healthy sibling served the shard. The full gate
passed 26 seconds after cleanup began.

Result: **PASS** — the cluster degraded and healed automatically.

### Chaos#3: Kill Only the ClickHouse Container

Save this YAML as `tests/03-container-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-03
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0
```

What this chaos does: Kills only the `clickhouse` process container while
leaving the pod object and its PVC in place.

**Expected behavior:** Kubernetes should restart the container, ClickHouse
should reconnect to Keeper and replication, and the restart count should
increase without a pod UID change.

Compare the `clickhouse` container restart count before and after injection.

**Observed behavior:**

This test killed the `clickhouse` container without deleting its pod. The
container restart count changed from 0 to 1. One batch succeeded and none
failed. Both replicas remained consistent after restart.

Result: **PASS** — Kubernetes restarted the database process and ClickHouse
rejoined replication without manual work.

### Chaos#4: Repeat Alternating Pod Kills

Save the first fault as `tests/04-a-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-04-a
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
  gracePeriod: 0
```

Save the second fault as `tests/04-b-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-04-b
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1
  gracePeriod: 0
```

Save the third fault as `tests/04-c-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-04-c
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1
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
kubectl exec -n demo "$workload_pod" -- rm -f /state/pause
```

The command prints nothing on success.

```bash
kubectl apply -f tests/04-a-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-04-a created
```

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/04-a-pod-kill.yaml --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-04-a condition met
```

```bash
kubectl delete -f tests/04-a-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-04-a" deleted from demo namespace
```

```bash
kubectl wait -n demo --for=create \
  pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 --timeout=5m
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 --timeout=5m
sleep 20
```

```text
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 condition met
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 condition met
```

```bash
kubectl apply -f tests/04-b-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-04-b created
```

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/04-b-pod-kill.yaml --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-04-b condition met
```

```bash
kubectl delete -f tests/04-b-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-04-b" deleted from demo namespace
```

```bash
kubectl wait -n demo --for=create \
  pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 --timeout=5m
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 --timeout=5m
sleep 20
```

```text
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 condition met
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 condition met
```

```bash
kubectl apply -f tests/04-c-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-04-c created
```

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/04-c-pod-kill.yaml --timeout=90s
```

```text
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-04-c condition met
```

```bash
kubectl delete -f tests/04-c-pod-kill.yaml
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-04-c" deleted from demo namespace
```

```bash
kubectl wait -n demo --for=create \
  pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1 --timeout=5m
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1 --timeout=5m
```

```text
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1 condition met
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1 condition met
```

**Observed behavior:**

Three one-shot pod kills targeted shard-0 replica-0, shard-1 replica-1, then
shard-0 replica-1, with 20 seconds between injections. Each target received a
new pod UID. The workload completed 36 batches with one failed or ambiguous
attempts, and no queue or restart instability accumulated.

Result: **PASS** — repeated isolated failures did not cause recovery drift.

### Chaos#5: Lose an Entire Shard

Save this YAML as `tests/05-full-shard-outage.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-05
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1
```

What this chaos does: Holds both replicas of shard 0 unavailable for 45
seconds.

**Expected behavior:** Distributed inserts that require shard 0 should fail
clearly; the remaining shard cannot substitute for missing shard data.
KubeDB should report `Critical`, then both replicas should return with equal
data.

**Observed behavior:**

Both replicas of shard 0 were failed for 45 seconds. A two-second status probe
showed KubeDB progress from `Ready` to `Critical`, then `NotReady` as the
health checks observed the sustained outage. Every workload attempt in the
measured fault window failed because a Distributed insert needs all selected
shards: 39 failed or ambiguous attempts and no acknowledged writes. Once both
replicas returned, their counts and checksums matched; the gate passed in 25
seconds.

Result: **PASS** — shard loss caused an honest availability failure, not silent
data inconsistency.

### Chaos#6: Lose the Entire ClickHouse Data Plane

Save this YAML as `tests/06-data-plane-outage.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-06
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1
```

What this chaos does: Makes all four ClickHouse data pods unavailable
while leaving the three Keeper members running.

**Expected behavior:** SQL clients should see a complete outage and no batch
should be acknowledged during it. When the fault ends, all four replicas
should reopen their existing PVC data and converge automatically.

**Observed behavior:**

All four ClickHouse pods were failed for 45 seconds while Keeper remained up.
The client had 39 failures and no successes. Kubernetes still displayed the
containers as Ready during `pod-failure`; real SQL and KubeDB's `Critical`
phase were the reliable signals. All acknowledged data returned afterward,
and the gate passed in 16 seconds.

Result: **PASS** — complete client outage recovered automatically with no
acknowledged data loss.

### Chaos#7: Kill a Keeper Follower

Discover the Keeper roles immediately before the test.

Save this YAML as `tests/07-keeper-follower-kill.yaml`, using a current
follower as the target:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-07
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-keeper-1
  gracePeriod: 0
```

What this chaos does: Discovers the Keeper roles with `mntr`, then abruptly
kills one current follower.

**Expected behavior:** The leader and remaining follower still form a
two-member majority, so coordination and writes should continue. The recreated
member should rejoin as a follower.

The example uses `keeper-1`; replace it if `mntr` reports that member as the
leader.

**Observed behavior:**

The suite discovered Keeper roles through `mntr` and killed one follower.
The leader and remaining follower retained quorum. Three batches succeeded and
none failed; the replacement member rejoined as a follower.

Result: **PASS** — one Keeper failure did not interrupt writes.

### Chaos#8: Kill the Keeper Leader

Rediscover the Keeper roles immediately before the test.

Save this YAML as `tests/08-keeper-leader-kill.yaml`, using the current leader
as the target:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-08
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-keeper-0
  gracePeriod: 0
```

What this chaos does: Discovers and kills the current Keeper leader.

**Expected behavior:** The two surviving members should elect a new leader
quickly. ClickHouse should tolerate the short election, and the old leader
should return as a follower rather than forming a second leader.

The example uses `keeper-0`; replace it with the actual leader. Time how long
another member takes to report `leader`.

**Observed behavior:**

The current leader, Keeper-0, was discovered dynamically and killed. Keeper-2
became the new leader in about four seconds. Three batches succeeded with zero errors,
and the old leader rejoined as a follower.

Result: **PASS** — Keeper leader election was automatic and safe.

### Chaos#9: Lose Keeper Quorum

Save this YAML as `tests/09-keeper-quorum-loss.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-09
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
        - clickhouse-chaos-v2-keeper-0
        - clickhouse-chaos-v2-keeper-1
```

What this chaos does: Holds two of the three Keeper members failed for 45
seconds, removing the majority required for coordination.

**Expected behavior:** Existing reads may continue, but coordination-dependent
writes or replication can stall or fail. After quorum returns, queued work
should settle and both replicas of every shard should converge without data
repair.

Do not repair a brief replica mismatch while queues are still moving. Require
two consecutive equal checks within ten minutes.

**Observed behavior:**

Two Keeper members were failed for 45 seconds. The surviving Keeper-2 still
answered `mntr` as `leader`, but one member cannot form a majority; the role
label alone therefore did not prove quorum. KubeDB still reported `Ready`.
Four batches succeeded before or around session loss and two attempts failed.

After quorum returned, both replicas converged without repair. The gate
required two equal count-and-checksum checks five seconds apart and passed in
30 seconds. This prevents a brief part-visibility race from being mistaken for
permanent divergence. The gate passed in 23 seconds.

Result: **PASS** — automatic convergence occurred, but KubeDB `Ready` alone did
not describe Keeper availability.

### Chaos#10: Fail All Keeper Members

Save this YAML as `tests/10-full-keeper-outage.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-10
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
        - clickhouse-chaos-v2-keeper-0
        - clickhouse-chaos-v2-keeper-1
        - clickhouse-chaos-v2-keeper-2
```

What this chaos does: Holds all three Keeper members unavailable for 45
seconds.

**Expected behavior:** ClickHouse processes can remain reachable, but Keeper
operations should be unavailable and replicated writes may fail. Recovery
requires a new one-leader/two-follower quorum and drained replica queues.

During injection, check `mntr` directly even if KubeDB still reports `Ready`.

**Observed behavior:**

All three Keeper pods were held failed for 45 seconds. ClickHouse processes
remained reachable and KubeDB still showed `Ready`, but Keeper-dependent work
stalled. Two batches succeeded around the session boundary and two attempts
failed. Quorum reformed with one leader and the full gate passed in 26 seconds.

Result: **PASS** — coordination outage was safe and recoverable.

## Network Chaos

### Chaos#11: Add Network Delay

Save this YAML as `tests/11-network-delay.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-v2-exp-11
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
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

**Observed behavior:**

The suite added 500ms latency with 50ms jitter to traffic reaching one
replica. Seventeen batches succeeded without errors, KubeDB stayed `Ready`, and
the final replication queue was empty. The gate passed in 14 seconds.

Result: **PASS** — latency reduced responsiveness but did not break integrity.

### Chaos#12: Drop 30 Percent of Packets

Save this YAML as `tests/12-network-loss.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-v2-exp-12
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
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

**Observed behavior:**

Thirty percent packet loss was applied to one replica. Twenty-two batches
succeeded and none failed. TCP retries and the healthy replica absorbed the
fault; the target later showed zero queue and zero delay, and the gate passed
in 12 seconds.

Result: **PASS** — packet loss caused no lasting replication damage.

### Chaos#13: Duplicate 50 Percent of Packets

Save this YAML as `tests/13-network-duplicate.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-v2-exp-13
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
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

**Observed behavior:**

Half of the selected replica's incoming packets were duplicated. Thirty-eight
batches succeeded. The final equality `count() == uniqExact(id)` proved that
network duplication did not create duplicate database rows.

Result: **PASS** — ClickHouse and TCP tolerated duplicated packets.

### Chaos#14: Limit Bandwidth to 1 Mbps

Save this YAML as `tests/14-bandwidth.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-v2-exp-14
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
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

**Observed behavior:**

Traffic reaching one replica was restricted to 1 Mbps for 45 seconds.
Thirty-eight batches still completed without error. Replica queues were empty
after the limit was removed.

Result: **PASS** — this small workload fit within the constrained bandwidth.

### Chaos#15: Partition One Replica from Data Peers

Save this YAML as `tests/15-data-partition.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-v2-exp-15
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
  direction: both
  target:
    mode: all
    selector:
      namespaces:
        - demo
      pods:
        demo:
          - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1
          - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0
          - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1
```

What this chaos does: Isolates shard-0 replica-0 in both directions from
the other three ClickHouse data pods.

**Expected behavior:** The isolated replica can fall behind while its sibling
serves the shard. After reconnection it should fetch missing parts and match
the sibling without deleting its pod or PVC.

**Observed behavior:**

Shard-0 replica-0 was isolated in both directions from the other three data
pods. The workload completed eight batches and saw three transient failures.
KubeDB stayed `Ready`; after reconnection, the isolated replica matched its
shard sibling and the gate passed in 18 seconds.

Result: **PASS** — the replica rejoined without manual synchronization.

### Chaos#16: Partition One Replica from Keeper

Save this YAML as `tests/16-keeper-partition.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: clickhouse-chaos-v2-exp-16
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
  direction: both
  target:
    mode: all
    selector:
      namespaces:
        - demo
      pods:
        demo:
          - clickhouse-chaos-v2-keeper-0
          - clickhouse-chaos-v2-keeper-1
          - clickhouse-chaos-v2-keeper-2
```

What this chaos does: Blocks one data replica from communicating with all
three Keeper members for 45 seconds.

**Expected behavior:** An existing Keeper session may mask a short partition.
If the session expires, replicated-table operations on that replica should
stop safely rather than accept uncoordinated state. It should become writable
again after reconnection.

**Observed behavior:**

One ClickHouse replica was isolated from all three Keeper members. At ten
seconds it still reported writable because its existing Keeper session had not
expired. It switched to read-only at about 15 seconds. Four batches succeeded
and two attempts failed during the measured
window. After reconnection, the replica was writable with two active replicas,
and the gate passed in 24 seconds.

Result: **PASS** — short Keeper isolation was tolerated and fully recovered.

## Resource Stress

### Chaos#17: Stress CPU

Save this YAML as `tests/17-cpu-stress.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: clickhouse-chaos-v2-exp-17
  namespace: demo
spec:
  mode: one
  duration: 60s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
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

**Observed behavior:**

Two stress workers requested 80 percent CPU load inside one ClickHouse
container for 60 seconds. Fifty batches succeeded, none failed, and the
pod did not restart. The database stayed `Ready`, and the gate passed in 12
seconds.

Result: **PASS** — the tested CPU pressure caused no availability loss.

### Chaos#18: Stress Memory

Save this YAML as `tests/18-memory-stress.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: clickhouse-chaos-v2-exp-18
  namespace: demo
spec:
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0
  containerNames:
    - clickhouse
  stressors:
    memory:
      workers: 1
      size: 256MiB
```

What this chaos does: Allocates an additional 256 MiB in a ClickHouse pod
whose memory limit is 1 GiB.

**Expected behavior:** This test should create strong pressure without
intentionally forcing an OOM kill. The process should remain alive and cgroup
usage should fall after cleanup. A restart or sustained near-limit usage would
fail the test.

While the fault is injected, compare `memory.current` with `memory.max`:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 -c clickhouse -- bash -c '
  printf "memory_current="; cat /sys/fs/cgroup/memory.current
  printf "memory_max="; cat /sys/fs/cgroup/memory.max
'
```

Output during the fault:

```text
memory_current=1061523456
memory_max=1073741824
```

For this experiment, use the following commands for cleanup instead of the
generic deletion step. Delete the fault, wait 60 seconds, and measure again:

```bash
kubectl delete -f tests/18-memory-stress.yaml
```

```text
stresschaos.chaos-mesh.org "clickhouse-chaos-v2-exp-18" deleted from demo namespace
```

Wait for cgroup usage to settle:

```bash
sleep 60
```

Output: none. `sleep` returned after 60 seconds.

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 -c clickhouse -- bash -c '
  printf "memory_current="; cat /sys/fs/cgroup/memory.current
  printf "memory_max="; cat /sys/fs/cgroup/memory.max
'
```

Output after recovery:

```text
memory_current=852787200
memory_max=1073741824
```

**Observed behavior:**

One worker allocated 256 MiB inside a pod limited to 1 GiB. Cgroup usage rose
to 1,061,523,456 bytes against a 1,073,741,824-byte maximum—about 98.9%.
During the fault and the 60-second cooldown observation, 88 batches succeeded,
none failed, and no OOM restart occurred. After 60 seconds, usage settled to
852,787,200 bytes.

Result: **PASS** — memory pressure recovered, but the narrow peak headroom is
an operational warning.

## IO Chaos

### Chaos#19: Add Filesystem Latency

Save this YAML as `tests/19-io-latency.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: clickhouse-chaos-v2-exp-19
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
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

Before running the full experiment, use the same manifest for a 10-second
capability probe with `delay: 10ms` and `percent: 5`. Continue only if it
reaches `AllInjected`, reaches `AllRecovered`, and leaves no FUSE mount.

During the full fault, confirm that IOChaos installed its FUSE layer:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  findmnt -T /var/lib/clickhouse
```

Output during injection:

```text
TARGET              SOURCE FSTYPE OPTIONS
/var/lib/clickhouse toda   fuse   rw,nosuid,nodev,relatime,user_id=0,group_id=0,default_permissions,allow_other
```

After `AllRecovered`, check both the mount and PID 1:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  mount | grep /var/lib/clickhouse
```

Output immediately after our fault recovered:

```text
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

The normal `ext4` mount had returned. Next, check the process:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

Output immediately after our fault recovered:

```text
    PID STAT COMMAND
      1 Tsl  clickhouse-serv
```

The mount should be the normal filesystem and the process state should not
contain `T`. If the mount is clean but PID 1 is stopped, resume the existing
process and rerun the complete gate:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  kill -CONT 1
```

`kill -CONT` prints nothing when it succeeds. Check PID 1 again:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

Output after `SIGCONT`:

```text
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

**Observed behavior:**

IOChaos mounted its `toda` FUSE layer over `/var/lib/clickhouse` and delayed
half of file operations by 100ms for 45 seconds. Three batches were
acknowledged and three attempts failed or became ambiguous. Chaos Mesh reported
`AllRecovered`, removed `toda`, and restored ext4, but left ClickHouse PID 1 in
stopped state `Tsl`. `kill -CONT 1` resumed the same process; both replicas
then matched and the full gate passed.

Result: **PASS WITH MANUAL CLEANUP** — ClickHouse data remained correct and
the FUSE mount was removed, but this IOChaos cleanup required `SIGCONT`.

### Chaos#20: Return Recoverable EIO

Save this YAML as `tests/20-io-fault.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: clickhouse-chaos-v2-exp-20
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0
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
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 -c clickhouse | \
  grep -E 'Input/output error|CANNOT_STATVFS' | wc -l
```

Output from our run:

```text
337
```

**Observed behavior:**

Ten percent of selected data-volume operations returned errno 5 for 30
seconds. ClickHouse logged messages including `Input/output error` and
`CANNOT_STATVFS`. In simple terms, ClickHouse temporarily could not read file
attributes or calculate free disk space. It did not corrupt bytes.

The workload had ten successes and 15 failures. The pod did not restart, the
normal ext4 mount returned automatically, and the gate passed in 18 seconds.
The logs contained 337 matching storage-error entries during the injection.

Result: **PASS** — explicit temporary disk errors were visible and recoverable.

## DNS and Time Chaos

### Chaos#21: Return Keeper DNS Errors

Save this YAML as `tests/21-keeper-dns-error.yaml`. Use the complete
`.svc.cluster.local` names; a pattern ending at `.svc` does not match the DNS
query made by the resolver.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: DNSChaos
metadata:
  name: clickhouse-chaos-v2-exp-21
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
  containerNames:
    - clickhouse
  patterns:
    - clickhouse-chaos-v2-keeper-0.clickhouse-chaos-v2-keeper-pods.demo.svc.cluster.local
    - clickhouse-chaos-v2-keeper-1.clickhouse-chaos-v2-keeper-pods.demo.svc.cluster.local
    - clickhouse-chaos-v2-keeper-2.clickhouse-chaos-v2-keeper-pods.demo.svc.cluster.local
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
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  getent hosts \
  clickhouse-chaos-v2-keeper-2.clickhouse-chaos-v2-keeper-pods.demo.svc.cluster.local
```

Output during injection:

```text
command terminated with exit code 2
```

No address was returned. After `AllRecovered`, run the same command again:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  getent hosts \
  clickhouse-chaos-v2-keeper-2.clickhouse-chaos-v2-keeper-pods.demo.svc.cluster.local
```

Output after recovery:

```text
10.42.0.232     clickhouse-chaos-v2-keeper-2.clickhouse-chaos-v2-keeper-pods.demo.svc.cluster.local
```

**Observed behavior:**

The first attempt used names ending at `.svc`; the resolver actually queried
the full `.svc.cluster.local` names, so a direct lookup still succeeded. That
attempt was treated as inconclusive and not counted.

The test was rerun with exact full Keeper FQDNs. A direct `getent` probe failed
during injection and resolved again after recovery. Existing Keeper sessions
and cached addresses allowed 38 batches to succeed without errors.

Result: **PASS** — DNS failure was proved, while established coordination
connections masked application impact.

### Chaos#22: Skew the Clock Back Two Hours

Save this YAML as `tests/22-clock-skew.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: TimeChaos
metadata:
  name: clickhouse-chaos-v2-exp-22
  namespace: demo
spec:
  mode: one
  duration: 45s
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1
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
workload_pod=$(kubectl get pod -n demo \
  -l app=clickhouse-chaos-v2-workload \
  -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n demo "$workload_pod" -- bash -c '
  clickhouse-client \
    --host clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1.clickhouse-chaos-v2-chaos-v2-cluster-shard-1-pods.demo.svc \
    --user "$CH_USER" \
    --password "$CH_PASSWORD" \
    --query "SELECT nowInBlock64(3), sleepEachRow(0.5)
             FROM numbers(120)
             SETTINGS max_block_size=1 FORMAT TSV"
'
```

Captured output excerpt from terminal 1:

```text
2026-09-02 07:06:03.694    0
2026-09-02 05:06:04.252    0
```

The adjacent timestamps prove that the same ClickHouse query moved backward
two hours. In terminal 2, validate and apply the fault:

```bash
kubectl apply --dry-run=server -f tests/22-clock-skew.yaml
```

```text
timechaos.chaos-mesh.org/clickhouse-chaos-v2-exp-22 created (server dry run)
```

```bash
kubectl apply -f tests/22-clock-skew.yaml
```

```text
timechaos.chaos-mesh.org/clickhouse-chaos-v2-exp-22 created
```

```bash
kubectl wait -n demo --for=condition=AllInjected \
  -f tests/22-clock-skew.yaml --timeout=90s
```

```text
timechaos.chaos-mesh.org/clickhouse-chaos-v2-exp-22 condition met
```

Wait for the 45-second fault to finish:

```bash
kubectl wait -n demo --for=condition=AllRecovered \
  -f tests/22-clock-skew.yaml --timeout=150s
```

```text
timechaos.chaos-mesh.org/clickhouse-chaos-v2-exp-22 condition met
```

```bash
kubectl delete -f tests/22-clock-skew.yaml
```

```text
timechaos.chaos-mesh.org "clickhouse-chaos-v2-exp-22" deleted from demo namespace
```

After `AllRecovered`, inspect PID 1:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 -c clickhouse -- \
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
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 -c clickhouse -- \
  kill -CONT 1
```

`kill -CONT` printed nothing. Checking PID 1 again produced:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

```text
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

**Observed behavior:**

The pre-existing ClickHouse timestamp stream moved backward by two hours as
soon as TimeChaos attached. The control replica kept the correct time. Twelve
batches succeeded during the measured fault window without an error.

Chaos Mesh reported recovery but left ClickHouse PID 1 in stopped state `Tsl`.
The documented `kill -CONT 1` action resumed it. The target and control then
matched the host UTC time, and the full gate passed in 11 seconds. The stopped
process was a TimeChaos cleanup problem: Chaos Mesh paused
the process while removing the clock injection but did not resume it. It does
not mean that ClickHouse stops simply because its timezone is wrong, and it is
not a general consequence of running ClickHouse with Chaos Mesh installed.

Result: **PASS WITH MANUAL CLEANUP** — ClickHouse data remained correct;
the tested clock skew was tolerated, but this TimeChaos run required
`SIGCONT` to finish cleanup.

## Combined Chaos and Recovery Soak

### Chaos#23: Combine I/O Latency with Sibling Failure

Save both resources in `tests/23-io-plus-sibling-failure.yaml` and prove that
both reach `AllInjected`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: clickhouse-chaos-v2-exp-23-io
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
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
  name: clickhouse-chaos-v2-exp-23-pod
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
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-1
```

What this chaos does: Adds 100 ms storage latency to shard-0 replica-0
while holding its sibling replica-1 failed.

**Expected behavior:** This removes the healthy fast path for the shard, so
errors and `Critical` are acceptable. Once both faults clear, the siblings
must converge and the IOChaos FUSE mount must disappear without PVC changes.

For this experiment, use the following commands instead of the generic
`kubectl delete -f` step. Delete `PodChaos` first and `IOChaos` second. If the
I/O target still shows a `toda` mount or `Transport endpoint is not connected`,
recreate only that pod while preserving its PVC, then rerun the complete gate:

```bash
kubectl delete podchaos -n demo clickhouse-chaos-v2-exp-23-pod
```

```text
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-23-pod" deleted from demo namespace
```

```bash
kubectl delete iochaos -n demo clickhouse-chaos-v2-exp-23-io
```

```text
iochaos.chaos-mesh.org "clickhouse-chaos-v2-exp-23-io" deleted from demo namespace
```

Check the I/O target's mount before running the gate:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  mount | grep /var/lib/clickhouse
```

Output from our run:

```text
/dev/vda1 on /var/lib/clickhouse type ext4 (rw,relatime,discard,errors=remount-ro,commit=30)
```

The mount recovered, but the process still needed to be checked:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

Output from our run:

```text
    PID STAT COMMAND
      1 Tsl  clickhouse-serv
```

If its state contains `T`, resume the existing ClickHouse process:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  kill -CONT 1
```

The signal command printed nothing. The next PID check showed:

```bash
kubectl exec -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 -c clickhouse -- \
  ps -o pid,stat,comm -p 1
```

```text
    PID STAT COMMAND
      1 Ssl  clickhouse-serv
```

If cleanup did not restore the normal mount, use:

```bash
kubectl delete pod -n demo \
  clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
kubectl wait -n demo --for=condition=Ready \
  pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 --timeout=5m
```

We did not run this fallback because the mount had already returned to ext4;
there is therefore no deletion output for this conditional command.

**Observed behavior:**

Shard-0 replica-0 received 100ms I/O latency while replica-1 was failed. This
made the shard degraded enough for KubeDB to report `Critical`. Five batches
succeeded and two failed or became ambiguous. Both faults recovered and
`toda` unmounted cleanly, but Chaos Mesh left the I/O target's PID 1 stopped in
state `Tsl`. `kill -CONT 1` resumed the existing process; the complete gate
then passed in 21 seconds without pod recreation.

Result: **PASS WITH MANUAL CLEANUP** — the shard recovered from simultaneous
storage and sibling failure, but IOChaos cleanup required `SIGCONT`.

### Chaos#24: Run Three Recovery-Soak Cycles

Save the first cycle as `tests/24-1-recovery-soak.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-24-1
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0
  gracePeriod: 0
```

Save the second cycle as `tests/24-2-recovery-soak.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-24-2
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0
  gracePeriod: 0
```

Save the third cycle as `tests/24-3-recovery-soak.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: clickhouse-chaos-v2-exp-24-3
  namespace: demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    pods:
      demo:
        - clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1
  gracePeriod: 0
```

What this chaos does: Repeats one-shot pod kills across three replicas,
running the complete recovery gate between cycles.

**Expected behavior:** Every cycle should return to the same healthy baseline.
No replication backlog, checksum difference, stopped process, stale FUSE
mount, or restart instability may accumulate across cycles.

Apply the files in numeric order. For each cycle, record the old UID, inject
the kill, require a new UID, delete the `PodChaos`, and pass the full recovery
gate before continuing.

```console
$ kubectl get pod -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 \
    -o jsonpath='{.metadata.uid}{"\n"}'
2faaa154-3abb-40cb-9b02-28b05aa04dbe
$ kubectl exec -n demo "$workload_pod" -- rm -f /state/pause
$ kubectl apply -f tests/24-1-recovery-soak.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-24-1 created
$ kubectl wait -n demo --for=condition=AllInjected \
    -f tests/24-1-recovery-soak.yaml --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-24-1 condition met
$ kubectl delete -f tests/24-1-recovery-soak.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-24-1" deleted from demo namespace
$ kubectl wait -n demo --for=create \
    pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 --timeout=5m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 condition met
$ kubectl wait -n demo --for=condition=Ready \
    pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 --timeout=5m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 condition met
$ kubectl get pod -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-1-0 \
    -o jsonpath='{.metadata.uid}{"\n"}'
11fac46f-2b3f-45fb-a37f-08072c7de6b3
```

Run the complete recovery gate, then perform cycle 2:

```console
$ kubectl get pod -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 \
    -o jsonpath='{.metadata.uid}{"\n"}'
e715653c-f75c-48ad-95b9-ec2991d298f4
$ kubectl exec -n demo "$workload_pod" -- rm -f /state/pause
$ kubectl apply -f tests/24-2-recovery-soak.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-24-2 created
$ kubectl wait -n demo --for=condition=AllInjected \
    -f tests/24-2-recovery-soak.yaml --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-24-2 condition met
$ kubectl delete -f tests/24-2-recovery-soak.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-24-2" deleted from demo namespace
$ kubectl wait -n demo --for=create \
    pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 --timeout=5m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 condition met
$ kubectl wait -n demo --for=condition=Ready \
    pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 --timeout=5m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 condition met
$ kubectl get pod -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-0-0 \
    -o jsonpath='{.metadata.uid}{"\n"}'
4070efd2-d645-4d21-8c6c-aac4a758766d
```

Run the complete recovery gate again, then perform cycle 3:

```console
$ kubectl get pod -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 \
    -o jsonpath='{.metadata.uid}{"\n"}'
0b5061db-202e-4965-85fc-1e591a4c043a
$ kubectl exec -n demo "$workload_pod" -- rm -f /state/pause
$ kubectl apply -f tests/24-3-recovery-soak.yaml
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-24-3 created
$ kubectl wait -n demo --for=condition=AllInjected \
    -f tests/24-3-recovery-soak.yaml --timeout=90s
podchaos.chaos-mesh.org/clickhouse-chaos-v2-exp-24-3 condition met
$ kubectl delete -f tests/24-3-recovery-soak.yaml
podchaos.chaos-mesh.org "clickhouse-chaos-v2-exp-24-3" deleted from demo namespace
$ kubectl wait -n demo --for=create \
    pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 --timeout=5m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 condition met
$ kubectl wait -n demo --for=condition=Ready \
    pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 --timeout=5m
pod/clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 condition met
$ kubectl get pod -n demo clickhouse-chaos-v2-chaos-v2-cluster-shard-1-1 \
    -o jsonpath='{.metadata.uid}{"\n"}'
3642c136-96d4-49b1-a175-6a8270972a35
```

Run the complete recovery gate one final time.

**Observed behavior:**

The final test killed and recovered three replicas one at a time, requiring
the complete gate after every cycle. Gate times were 14, 12, and 11 seconds.
Across all cycles, 29 batches succeeded and two failed or became ambiguous.
Every target got a new UID; no backlog, checksum drift, or unhealthy process
accumulated.

Result: **PASS** — repeated recovery remained stable.

### Chaos#25: Delete One Shard Replica and Its PVC

This test is different from `PodChaos`. Killing a pod leaves its Persistent
VolumeClaim intact, so the replacement pod simply mounts the same data again.
Here we deliberately remove both the pod and its PVC to simulate permanent
loss of one replica's disk. Chaos Mesh does not delete Kubernetes PVCs, so the
fault is injected with `kubectl delete`.

What this chaos does: Permanently removes the local metadata and data files of
shard-0 replica-1. Shard-0 replica-0 remains available as its donor, while
both replicas of shard 1 remain untouched as a control.

**Expected behavior:** KubeDB should create a new pod and PVC, detect that the
new ClickHouse process has no local schema, remove the lost replica's stale
Keeper registrations, recreate its schema from the sibling replica, and let
`ReplicatedMergeTree` fetch all shard-0 parts. The rebuilt replica must have a
new pod UID, PVC UID, and PV name, but its row count and checksum must match
the sibling. No manual `CREATE TABLE`, `ATTACH PART`, or data copy is allowed
after fault injection.

This experiment requires a KubeDB ClickHouse operator with replica recovery
support. It was executed separately from the earlier workload so that the
previous campaign's counters remained unchanged.

Save this manifest as `clickhouse-replica-recovery.yaml`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-replica-recovery
  namespace: demo
  labels:
    chaos-test.kubedb.com/suite: replica-pvc-loss
spec:
  version: 26.2.6
  clusterTopology:
    cluster:
      name: recovery-cluster
      shards: 2
      replicas: 2
      storageType: Durable
      storage:
        storageClassName: local-path
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
      podTemplate:
        spec:
          containers:
            - name: clickhouse
              resources:
                requests:
                  cpu: 500m
                  memory: 1Gi
                limits:
                  cpu: "2"
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

Validate and create the recovery cluster:

```bash
kubectl apply --dry-run=server -f clickhouse-replica-recovery.yaml
```

```text
clickhouse.kubedb.com/clickhouse-replica-recovery created (server dry run)
```

```bash
kubectl apply -f clickhouse-replica-recovery.yaml
```

```text
clickhouse.kubedb.com/clickhouse-replica-recovery created
```

```bash
kubectl wait -n demo clickhouse/clickhouse-replica-recovery \
  --for=jsonpath='{.status.phase}'=Ready --timeout=10m
```

```text
clickhouse.kubedb.com/clickhouse-replica-recovery condition met
```

Confirm that the two-shard, two-replica data plane and three-member Keeper
quorum are running:

```bash
kubectl get pods -n demo \
  -l app.kubernetes.io/instance=clickhouse-replica-recovery
```

```text
NAME                                                     READY   STATUS    RESTARTS   AGE
clickhouse-replica-recovery-keeper-0                     1/1     Running   0          68s
clickhouse-replica-recovery-keeper-1                     1/1     Running   0          62s
clickhouse-replica-recovery-keeper-2                     1/1     Running   0          57s
clickhouse-replica-recovery-recovery-cluster-shard-0-0   1/1     Running   0          66s
clickhouse-replica-recovery-recovery-cluster-shard-0-1   1/1     Running   0          61s
clickhouse-replica-recovery-recovery-cluster-shard-1-0   1/1     Running   0          64s
clickhouse-replica-recovery-recovery-cluster-shard-1-1   1/1     Running   0          60s
```

Create a replicated local table and a Distributed table. The local table owns
the data on each shard; the Distributed table routes inserts and reads across
both shards:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --multiquery --query "
      CREATE DATABASE IF NOT EXISTS recovery_test ON CLUSTER '\''{cluster}'\'';
      CREATE TABLE IF NOT EXISTS recovery_test.events_local
      ON CLUSTER '\''{cluster}'\'' (
        id UInt64,
        payload UInt64
      )
      ENGINE = ReplicatedMergeTree(
        '\''/clickhouse/{installation}/{cluster}/tables/{shard}/{database}/{table}'\'',
        '\''{replica}'\''
      )
      ORDER BY id;
      CREATE TABLE IF NOT EXISTS recovery_test.events
      ON CLUSTER '\''{cluster}'\'' AS recovery_test.events_local
      ENGINE = Distributed(
        '\''{cluster}'\'', recovery_test, events_local, id
      );
    "
'
```

Each DDL returned one successful status row from every data pod. The three
DDL statements therefore produced twelve rows:

```text
clickhouse-replica-recovery-recovery-cluster-shard-0-0.clickhouse-replica-recovery-pods  9000  0    3  0
clickhouse-replica-recovery-recovery-cluster-shard-1-1.clickhouse-replica-recovery-pods  9000  0    2  0
clickhouse-replica-recovery-recovery-cluster-shard-1-0.clickhouse-replica-recovery-pods  9000  0    1  0
clickhouse-replica-recovery-recovery-cluster-shard-0-1.clickhouse-replica-recovery-pods  9000  0    0  0
clickhouse-replica-recovery-recovery-cluster-shard-0-0.clickhouse-replica-recovery-pods  9000  0    3  0
clickhouse-replica-recovery-recovery-cluster-shard-1-1.clickhouse-replica-recovery-pods  9000  0    2  0
clickhouse-replica-recovery-recovery-cluster-shard-1-0.clickhouse-replica-recovery-pods  9000  0    1  0
clickhouse-replica-recovery-recovery-cluster-shard-0-1.clickhouse-replica-recovery-pods  9000  0    0  0
clickhouse-replica-recovery-recovery-cluster-shard-0-0.clickhouse-replica-recovery-pods  9000  0    3  0
clickhouse-replica-recovery-recovery-cluster-shard-1-1.clickhouse-replica-recovery-pods  9000  0    2  0
clickhouse-replica-recovery-recovery-cluster-shard-1-0.clickhouse-replica-recovery-pods  9000  0    1  0
clickhouse-replica-recovery-recovery-cluster-shard-0-1.clickhouse-replica-recovery-pods  9000  0    0  0
```

Insert 100,000 deterministic rows. `id` is both unique and the sharding key;
`payload` is derived from `id`, so both count and checksum can be reproduced:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "INSERT INTO recovery_test.events
      SELECT number, sipHash64(number)
      FROM numbers(100000)
      SETTINGS insert_distributed_sync=1"
'
```

The insert printed nothing on success. Check the whole cluster:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events FORMAT TSV"
'
```

```text
100000  100000  6234091558740710151
```

Run `SYSTEM SYNC REPLICA recovery_test.events_local` once on each data pod.
Each successful command prints nothing:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SYSTEM SYNC REPLICA recovery_test.events_local"
'
```

Output: none.

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SYSTEM SYNC REPLICA recovery_test.events_local"
'
```

Output: none.

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SYSTEM SYNC REPLICA recovery_test.events_local"
'
```

Output: none.

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SYSTEM SYNC REPLICA recovery_test.events_local"
'
```

Output: none.

Before deleting anything, query the local table on each pod separately. Use
the same query for every pod:

```sql
SELECT count(), uniqExact(id), sum(cityHash64(id, payload))
FROM recovery_test.events_local FORMAT TSV
```

Shard-0 replica-0:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50000  50000  13247772413203435930
```

Shard-0 replica-1, which will be deleted:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50000  50000  13247772413203435930
```

Shard-1 replica-0:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50000  50000  11433063219246825837
```

Shard-1 replica-1:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50000  50000  11433063219246825837
```

Record the donor identities:

```bash
kubectl get pod -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -o jsonpath='donor_pod_uid={.metadata.uid}{"\n"}'
```

```text
donor_pod_uid=a66cf33b-4ac1-4ab9-90d1-3402a381bb16
```

```bash
kubectl get pvc -n demo \
  data-clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -o jsonpath='donor_pvc_uid={.metadata.uid}{" donor_pv="}{.spec.volumeName}{"\n"}'
```

```text
donor_pvc_uid=20e36c93-de9f-4a84-9f8d-57d92b445308 donor_pv=pvc-20e36c93-de9f-4a84-9f8d-57d92b445308
```

Record the target identities:

```bash
kubectl get pod -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -o jsonpath='target_pod_uid={.metadata.uid}{"\n"}'
```

```text
target_pod_uid=0feb827f-f657-42b8-8806-ecba0635d44e
```

```bash
kubectl get pvc -n demo \
  data-clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -o jsonpath='target_pvc_uid={.metadata.uid}{" target_pv="}{.spec.volumeName}{"\n"}'
```

```text
target_pvc_uid=c0ee4206-2c61-4c6f-a2de-5b6c8ded81aa target_pv=pvc-c0ee4206-2c61-4c6f-a2de-5b6c8ded81aa
```

Delete the PVC first without waiting. Kubernetes keeps it protected while the
pod is still using it:

```bash
kubectl delete pvc -n demo \
  data-clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  --wait=false
```

```text
persistentvolumeclaim "data-clickhouse-replica-recovery-recovery-cluster-shard-0-1" deleted from demo namespace
```

Now force-delete only the target pod. This releases the old PVC and allows the
PetSet to create a new disk for the same replica ordinal:

```bash
kubectl delete pod -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  --grace-period=0 --force --wait=false
```

```text
Warning: Immediate deletion does not wait for confirmation that the running resource has been terminated.
pod "clickhouse-replica-recovery-recovery-cluster-shard-0-1" force deleted from demo namespace
```

Wait for the replacement pod and PVC objects:

```bash
kubectl wait -n demo --for=create \
  pod/clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  --timeout=2m
```

```text
pod/clickhouse-replica-recovery-recovery-cluster-shard-0-1 condition met
```

```bash
kubectl wait -n demo --for=create \
  pvc/data-clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  --timeout=2m
```

```text
persistentvolumeclaim/data-clickhouse-replica-recovery-recovery-cluster-shard-0-1 condition met
```

Prove that Kubernetes supplied new storage rather than reattaching the old
disk:

```bash
kubectl get pod -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -o jsonpath='new_target_pod_uid={.metadata.uid}{"\n"}'
```

```text
new_target_pod_uid=04b5455f-a021-45fd-bb1e-b678ed7e8a60
```

```bash
kubectl get pvc -n demo \
  data-clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -o jsonpath='new_target_pvc_uid={.metadata.uid}{" new_target_pv="}{.spec.volumeName}{"\n"}'
```

```text
new_target_pvc_uid=2e3c7b23-7ba7-4ed1-a022-c775ecd0f9f8 new_target_pv=pvc-2e3c7b23-7ba7-4ed1-a022-c775ecd0f9f8
```

```bash
kubectl get pv pvc-c0ee4206-2c61-4c6f-a2de-5b6c8ded81aa
```

```text
Error from server (NotFound): persistentvolumes "pvc-c0ee4206-2c61-4c6f-a2de-5b6c8ded81aa" not found
```

Wait for the replacement process to start, then immediately test for the
local table:

```bash
kubectl wait -n demo \
  pod/clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  --for=condition=Ready --timeout=2m
```

```text
pod/clickhouse-replica-recovery-recovery-cluster-shard-0-1 condition met
```

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "EXISTS TABLE recovery_test.events_local"
'
```

```text
0
```

The zero is important: it proves that the new PVC was empty. The pod process
being Ready only means ClickHouse can answer a request; it does not mean the
lost schema and data have already returned. At this point KubeDB reported:

```bash
kubectl get clickhouse -n demo clickhouse-replica-recovery
```

```text
NAME                          VERSION   STATUS     AGE
clickhouse-replica-recovery   26.2.6    Critical   3m
```

Do not create the missing table manually. The recovery controller waits until
the new pod has remained Ready for 60 seconds, then treats the missing schema
as a wiped replica instead of a slow startup. Wait for KubeDB recovery:

```bash
kubectl wait -n demo clickhouse/clickhouse-replica-recovery \
  --for=jsonpath='{.status.phase}'=Ready --timeout=10m
```

```text
clickhouse.kubedb.com/clickhouse-replica-recovery condition met
```

Confirm that the operator—not a manual command—performed the repair:

```bash
kubectl logs -n kubedb kubedb-kubedb-provisioner-0 \
  -c operator --since=10m | \
  grep 'replica recovery: repaired clickhouse-replica-recovery-recovery-cluster-shard-0-1'
```

```text
I0903 03:46:14.066272       1 replica_recovery.go:283] replica recovery: repaired clickhouse-replica-recovery-recovery-cluster-shard-0-1 from clickhouse-replica-recovery-recovery-cluster-shard-0-0, created 3 object(s)
```

The rebuilt target must now match its donor. Check the donor first:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50000  50000  13247772413203435930
```

Check the rebuilt target:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50000  50000  13247772413203435930
```

The matching count, unique-ID count, and checksum prove that the replacement
downloaded the shard's data from its sibling.

Check replication state on the rebuilt target:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT is_readonly, is_session_expired, queue_size,
      inserts_in_queue, merges_in_queue, total_replicas, active_replicas
      FROM system.replicas
      WHERE database='\''recovery_test'\'' AND table='\''events_local'\''
      FORMAT TSV"
'
```

```text
0  0  0  0  0  2  2
```

Read left to right, the rebuilt table is writable, its Keeper session is
active, every queue is empty, and both replicas are active. Run the same check
on shard-0 replica-0:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT is_readonly, is_session_expired, queue_size,
      inserts_in_queue, merges_in_queue, total_replicas, active_replicas
      FROM system.replicas
      WHERE database='\''recovery_test'\'' AND table='\''events_local'\''
      FORMAT TSV"
'
```

```text
0  0  0  0  0  2  2
```

Run it on shard-1 replica-0:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT is_readonly, is_session_expired, queue_size,
      inserts_in_queue, merges_in_queue, total_replicas, active_replicas
      FROM system.replicas
      WHERE database='\''recovery_test'\'' AND table='\''events_local'\''
      FORMAT TSV"
'
```

```text
0  0  0  0  0  2  2
```

Run it on shard-1 replica-1:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT is_readonly, is_session_expired, queue_size,
      inserts_in_queue, merges_in_queue, total_replicas, active_replicas
      FROM system.replicas
      WHERE database='\''recovery_test'\'' AND table='\''events_local'\''
      FORMAT TSV"
'
```

```text
0  0  0  0  0  2  2
```

Finally, prove that the recovered cluster accepts and replicates new data.
Insert 100 new deterministic rows:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "INSERT INTO recovery_test.events
      SELECT number, sipHash64(number)
      FROM numbers(100000, 100)
      SETTINGS insert_distributed_sync=1"
'
```

The insert printed nothing on success. The final Distributed result was:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events FORMAT TSV"
'
```

```text
100100  100100  8550420753814024808
```

Synchronize the four replicas again after the insert:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SYSTEM SYNC REPLICA recovery_test.events_local"
'
```

Output: none.

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SYSTEM SYNC REPLICA recovery_test.events_local"
'
```

Output: none.

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SYSTEM SYNC REPLICA recovery_test.events_local"
'
```

Output: none.

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SYSTEM SYNC REPLICA recovery_test.events_local"
'
```

Output: none.

Check shard-0 replica-0:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50050  50050  2504112362217261381
```

Check the rebuilt shard-0 replica-1:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-0-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50050  50050  2504112362217261381
```

Check shard-1 replica-0:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-0 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50050  50050  6046308391596763427
```

Check shard-1 replica-1:

```bash
kubectl exec -n demo \
  clickhouse-replica-recovery-recovery-cluster-shard-1-1 \
  -c clickhouse -- bash -c '
  clickhouse-client \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --query "SELECT count(), uniqExact(id),
      sum(cityHash64(id, payload))
      FROM recovery_test.events_local FORMAT TSV"
'
```

```text
50050  50050  6046308391596763427
```

Check Keeper-0:

```bash
kubectl exec -n demo clickhouse-replica-recovery-keeper-0 \
  -c clickhouse-keeper -- bash -c '
  exec 3<>/dev/tcp/127.0.0.1/9181
  printf "mntr\n" >&3
  timeout 3 cat <&3
' | awk '$1=="zk_server_state" {print $2}'
```

```text
leader
```

Check Keeper-1:

```bash
kubectl exec -n demo clickhouse-replica-recovery-keeper-1 \
  -c clickhouse-keeper -- bash -c '
  exec 3<>/dev/tcp/127.0.0.1/9181
  printf "mntr\n" >&3
  timeout 3 cat <&3
' | awk '$1=="zk_server_state" {print $2}'
```

```text
follower
```

Check Keeper-2:

```bash
kubectl exec -n demo clickhouse-replica-recovery-keeper-2 \
  -c clickhouse-keeper -- bash -c '
  exec 3<>/dev/tcp/127.0.0.1/9181
  printf "mntr\n" >&3
  timeout 3 cat <&3
' | awk '$1=="zk_server_state" {print $2}'
```

```text
follower
```

The leader may be a different Keeper when you run the test. The required
result is exactly one leader and two followers.

Check the final KubeDB phase:

```bash
kubectl get clickhouse -n demo clickhouse-replica-recovery
```

```text
NAME                          VERSION   STATUS   AGE
clickhouse-replica-recovery   26.2.6    Ready    6m
```

Check the complete pod set:

```bash
kubectl get pods -n demo \
  -l app.kubernetes.io/instance=clickhouse-replica-recovery
```

```text
NAME                                                     READY   STATUS    RESTARTS   AGE
clickhouse-replica-recovery-keeper-0                     1/1     Running   0          5m43s
clickhouse-replica-recovery-keeper-1                     1/1     Running   0          5m37s
clickhouse-replica-recovery-keeper-2                     1/1     Running   0          5m32s
clickhouse-replica-recovery-recovery-cluster-shard-0-0   1/1     Running   0          5m41s
clickhouse-replica-recovery-recovery-cluster-shard-0-1   1/1     Running   0          3m10s
clickhouse-replica-recovery-recovery-cluster-shard-1-0   1/1     Running   0          5m39s
clickhouse-replica-recovery-recovery-cluster-shard-1-1   1/1     Running   0          5m35s
```

The replacement pod was created at `03:44:53Z`; KubeDB returned to `Ready` at
`03:46:33Z`, 100 seconds later. This includes the deliberate 60-second
ready-pod grace period and subsequent schema and part synchronization.

**Observed behavior:** The target pod initially started on a genuinely empty
new disk and had no `recovery_test.events_local` table. KubeDB reported
`Critical`, selected shard-0 replica-0 as the donor, removed the stale Keeper
registrations, recreated three objects, and allowed `ReplicatedMergeTree` to
download all 50,000 original rows. The rebuilt replica's checksum matched its
sibling exactly, all queues returned to zero, and 100 post-recovery rows were
accepted and replicated. The other shard remained unchanged.

Result: **PASS** — one complete shard-replica disk loss recovered
automatically from its sibling without data loss or manual database repair.

## Chaos Testing Results Summary

| # | Fault | Workload success/error delta | Recovery observation | Verdict |
| ---: | --- | ---: | --- | --- |
| 1 | Single replica pod kill | +5 / +0 | Pod UID changed; full gate passed | PASS |
| 2 | Replica pod failure, 45s | +27 / +10 | Status still `Ready` at 15s; gate in 26s | PASS |
| 3 | ClickHouse container kill | +1 / +0 | Restart count increased; full gate passed | PASS |
| 4 | Three alternating pod kills | +36 / +1 | All three UIDs changed; no cumulative backlog | PASS |
| 5 | Both replicas of shard 0 failed | +0 / +39 | `Ready` → `Critical` → `NotReady`; gate in 25s | PASS |
| 6 | All four data pods failed | +0 / +39 | Client outage; gate in 16s | PASS |
| 7 | Keeper follower kill | +3 / +0 | Quorum retained; full gate passed | PASS |
| 8 | Keeper leader kill | +3 / +0 | New leader in about 4s | PASS |
| 9 | Keeper quorum loss | +4 / +2 | Survivor reported leader without quorum; gate in 23s | PASS |
| 10 | All Keeper members failed | +2 / +2 | Quorum reformed; gate in 26s | PASS |
| 11 | 500ms network delay | +17 / +0 | Stayed `Ready`; gate in 14s | PASS |
| 12 | 30% packet loss | +22 / +0 | Stayed `Ready`; gate in 12s | PASS |
| 13 | 50% packet duplication | +38 / +0 | No duplicate IDs; gate in 12s | PASS |
| 14 | 1 Mbps bandwidth limit | +38 / +0 | No write errors; gate in 11s | PASS |
| 15 | Data-replica partition | +8 / +3 | Rejoined and converged in 18s | PASS |
| 16 | Replica-to-Keeper partition | +4 / +2 | Read-only after about 15s; gate in 24s | PASS |
| 17 | CPU stress | +50 / +0 | No restart; gate in 12s | PASS |
| 18 | 256 MiB memory stress | +88 / +0 | Includes cooldown; no restart | PASS |
| 19 | 100ms I/O latency | +3 / +3 | ext4 restored; manual `SIGCONT` needed | PASS WITH MANUAL CLEANUP |
| 20 | 10% EIO | +10 / +15 | 337 storage errors; gate in 18s | PASS |
| 21 | Keeper DNS errors | +38 / +0 | DNS error proved; cached session kept writes alive | PASS |
| 22 | Clock skew −2h | +12 / +0 | Skew proved; manual `SIGCONT` needed | PASS WITH MANUAL CLEANUP |
| 23 | I/O latency plus sibling failure | +5 / +2 | `Critical`; manual `SIGCONT`; gate in 21s | PASS WITH MANUAL CLEANUP |
| 24 | Three-cycle recovery soak | +29 / +2 | Gates: 14s, 12s, 11s | PASS |
| 25 | Shard replica and PVC deletion | One-shot 100,000-row dataset | New pod/PVC/PV; schema and 50,000 shard rows restored in 100s | PASS |

The table records the counters captured around each fault window. Six other
successful batches completed during the short setup or recovery transitions,
so the success deltas are not intended to sum to the final client counter.

## Final Integrity Evidence

The original 24-experiment workload finished with:

```text
ClickHouse phase: Ready
Ready database pods: 7/7
Distributed rows: 50,987
Unique IDs: 50,987

Shard 0 replica 0: 25,968 rows, checksum 1308421401945657900
Shard 0 replica 1: 25,968 rows, checksum 1308421401945657900
Shard 1 replica 0: 25,019 rows, checksum 1106137736089666342
Shard 1 replica 1: 25,019 rows, checksum 1106137736089666342

Every replica: writable=1, queue=0, total=2, active=2, lost_parts=0, delay=0
Keeper: 1 leader, 2 followers
Data mounts: ext4 on all four pods
ClickHouse PID 1 state: running on all four pods
Remaining test chaos objects: 0
```

Here `writable=1` summarizes `is_readonly=0`.

The separate replica-and-PVC-loss experiment finished with:

```text
ClickHouse phase: Ready
Ready database pods: 7/7
Distributed rows: 100,100
Unique IDs: 100,100

Shard 0 replica 0: 50,050 rows, checksum 2504112362217261381
Shard 0 replica 1: 50,050 rows, checksum 2504112362217261381
Shard 1 replica 0: 50,050 rows, checksum 6046308391596763427
Shard 1 replica 1: 50,050 rows, checksum 6046308391596763427

Every replica: is_readonly=0, queue_size=0, total_replicas=2, active_replicas=2
Keeper: 1 leader, 2 followers
Replacement pod restarts: 0
Old PV: deleted
Replacement PVC: Bound
```

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
- **PID state `Tsl`:** two IOChaos runs and the TimeChaos run left ClickHouse
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
   the EIO case cleanly. The standalone I/O-latency run, combined I/O run, and
   TimeChaos each required manual `SIGCONT` after cleanup reported success.
6. Memory pressure reached 98.9% of the limit without OOM; production limits
   should retain more safety headroom.
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

Experiments 19, 22, and 23 also separate database resilience from chaos-tool
cleanup. ClickHouse data remained correct under storage latency and a two-hour
clock offset, but Chaos Mesh 2.8.4 did not resume the affected process
automatically in those three runs. Therefore the database-integrity checks
passed, while the automatic cleanup expectation did not.

## What Next?

Repeat the same suite on a multi-node cluster with production-equivalent
storage, workload volume, resource limits, and monitoring. NodeChaos can then
be added safely because losing one worker will not also remove the Kubernetes
control plane and the chaos controller.

## Cleanup

Save the final evidence before removing anything. Pause the workload, record
its counters, then scale it down:

```bash
workload_pod=$(kubectl get pod -n demo \
  -l app=clickhouse-chaos-v2-workload \
  -o jsonpath='{.items[0].metadata.name}')
```

The variable assignment prints nothing. Pause and read the final counters:

```bash
kubectl exec -n demo "$workload_pod" -- touch /state/pause
kubectl exec -n demo "$workload_pod" -- bash -c '
  printf "attempts="; cat /state/attempt_batches
  printf "success="; cat /state/success_batches
  printf "failed="; cat /state/failed_batches
'
```

```text
attempts=611
success=491
failed=120
```

```bash
kubectl scale deployment -n demo \
  clickhouse-chaos-v2-workload --replicas=0
```

```text
deployment.apps/clickhouse-chaos-v2-workload scaled
```

Repeat the complete recovery gate directly from a ClickHouse pod. Record the
Distributed and local counts, unique-ID counts, checksums, `system.replicas`
state, Keeper roles, pod state, PID states, mounts, and
remaining Chaos Mesh resources.

Delete only resources belonging to this disposable campaign:

```bash
kubectl get podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos \
  -n demo
```

```text
No resources found in demo namespace.
```

```bash
kubectl delete deployment -n demo clickhouse-chaos-v2-workload \
  --ignore-not-found
```

```text
deployment.apps "clickhouse-chaos-v2-workload" deleted
```

```bash
kubectl delete configmap -n demo clickhouse-chaos-v2-workload \
  --ignore-not-found
```

```text
configmap "clickhouse-chaos-v2-workload" deleted
```

```bash
kubectl delete clickhouse -n demo clickhouse-chaos-v2 \
  --ignore-not-found
```

```text
clickhouse.kubedb.com "clickhouse-chaos-v2" deleted
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
  grep clickhouse-chaos-v2
```

Output from our final name check:

```text
(no output)
```

The ClickHouse, its PetSets, pods, PVCs, generated
auth Secret, and Chaos Mesh resources no longer existed. Unrelated resources
in `demo` remained unchanged.
