# ClickHouse Chaos Blog Evidence Design

## Goal

Publish one reproducible AppsCode blog article that demonstrates the impact and
recovery of all 25 ClickHouse chaos experiments with the same level of evidence
as the PostgreSQL chaos article.

## Test Environment

- Namespace: `demo`
- ClickHouse resource: `clickhouse-chaos`
- Topology: two shards, two replicas per shard, three managed Keeper members
- ClickHouse data storage: 4Gi per pod
- ClickHouse CPU limit: 1 CPU per pod
- Workload: continuous 100-row UUID-keyed synchronous Distributed inserts
- Chaos Mesh resources: created in `demo`
- Existing operator and container images remain unchanged

The campaign uses disposable test data. Any existing ClickHouse resource used
by the earlier recovery experiment is removed before creating the fresh
`clickhouse-chaos` resource. Unrelated resources in `demo` are not modified.

## Evidence Format

Every experiment contains these sections in this order:

1. Purpose and real-world failure being simulated.
2. Expected ClickHouse, replica, Keeper, workload, and KubeDB-status behavior.
3. A named manifest file under `tests/`.
4. Baseline commands and their captured outputs.
5. One injection command followed immediately by its output.
6. A separate `AllInjected` command followed immediately by its output.
7. Commands and outputs showing the fault's live impact.
8. Commands and outputs showing the KubeDB status transition.
9. A separate `AllRecovered` command followed immediately by its output.
10. Recovery checks for rows, unique IDs, replica equality, replication queues,
    Keeper roles, pod state, process state, and mounts as applicable.
11. One cleanup command followed immediately by its output.
12. A plain-language explanation and verdict.

No code block combines unrelated commands and aggregate output. A command that
is intrinsically one shell script, such as the workload container entrypoint,
may contain multiple shell statements.

## Recovery Gate

After every experiment, testing stops until all applicable checks pass:

- ClickHouse reports `Ready`.
- Four ClickHouse pods and three Keeper pods are Ready.
- The Distributed table has `count() == uniqExact(id)`.
- Both replicas of each shard have matching row counts and payload checksums in
  two stable observations five seconds apart.
- `system.replicas` reports writable replicas, an empty queue, two total
  replicas, and two active replicas.
- Keeper reports exactly one leader and two followers.
- ClickHouse PID 1 is running and data mounts have no stale Chaos Mesh FUSE
  layer.
- The test's Chaos Mesh object no longer exists.

Timeouts are recorded as failed or ambiguous writes because ClickHouse may have
accepted rows before the client lost the response. They are not called data
loss or duplication without ID and checksum evidence.

## Experiment Scope

The article demonstrates all 25 existing scenarios: data pod and container
failures, shard and data-plane outages, Keeper member and quorum failures,
network delay/loss/duplication/bandwidth/partitions, CPU and bounded-memory
stress, filesystem latency and EIO, Keeper DNS failure, clock skew, combined
faults, recovery soak cycles, and deletion of one shard replica with its PVC.

The PVC-loss experiment additionally proves that the pod UID, PVC UID, and PV
change; the replacement initially has no local schema; the operator repairs it
from its sibling; the old PV disappears; and new writes succeed after recovery.

## Safety and Failure Handling

- Run only against the disposable `clickhouse-chaos` resource.
- Capture the target pod and PVC before destructive commands.
- Never delete both replicas of a shard during the PVC-loss experiment.
- Pause the workload before final equality checks to avoid observing an active
  replication visibility race.
- For the known Chaos Mesh cleanup issue, document `SIGCONT` only when PID 1 is
  actually in a stopped state.
- If a test does not reach `AllInjected`, report it as not executed and fix the
  selector before retrying.
- If recovery fails, preserve logs and object status in the article rather than
  describing an unobserved successful recovery.

## Deliverable and Validation

The publishable file is
`content/post/chaos-testing-clickhouse/index.md`. It retains AppsCode front
matter and contains all manifests, commands, captured outputs, explanations,
the results summary, and cleanup instructions.

Validation requires 25 experiment headings, 25 summary rows, balanced Markdown
fences, no stale `clickhouse-chaos-v2` or `clickhouse-replica-recovery` names,
and a Hugo build when the required Hugo tooling is available.
