# ClickHouse Chaos Blog Evidence Design

## Objective

Rebuild the ClickHouse chaos-testing article from one chronological execution
against a fresh KubeDB-managed `clickhouse-chaos` cluster, replacing every
unsupported, contradictory, stale, or out-of-order observation with captured
command output.

## Scope

- Recreate `clickhouse-chaos` in namespace `demo` with two shards, two
  replicas per shard, three Keeper members, 4Gi data PVCs, and a 1 CPU limit.
- Reset and deploy the continuous workload once.
- Execute experiments 1 through 25 in numeric order.
- Update only
  `content/post/chaos-testing-clickhouse/index.md` with publishable results.
- Preserve the existing PostgreSQL-inspired structure and explicit workload
  resume/pause handling for every experiment.

## Evidence Contract

Every experiment must include:

1. A relevant healthy baseline captured before `kubectl apply` or destructive
   action.
2. Confirmation that Chaos Mesh reached `AllInjected`, except experiment 25,
   which is an intentional Kubernetes pod/PVC deletion.
3. Direct evidence of the configured fault while it is active. Resource
   creation, `AllInjected`, or an unchanged `Ready` phase alone is not proof.
4. Workload counters before and after the measured window.
5. `AllRecovered`, fault-object deletion, and ClickHouse recovery in the
   correct chronological order.
6. Fault-specific recovery evidence plus replica integrity evidence.
7. One command followed immediately by its output. Commands for different
   replicas or Keeper members remain separate; no shell loops are published.

## Fault-Specific Proof

- Pod kill/failure: target UID, readiness, restart, or direct unavailability
  before/during/after as appropriate.
- Keeper failures: roles before injection, unavailable member/quorum during
  injection, and exactly one leader plus two followers after recovery.
- Network delay: measured request latency before and during injection.
- Packet loss: packet or request results that visibly show loss during
  injection.
- Packet duplication: a network probe that observes duplicates, followed by
  database unique-ID verification.
- Bandwidth: transfer duration or throughput before and during the 1 Mbps cap.
- Network partition: a direct connection that succeeds before, fails during,
  and succeeds after the partition.
- CPU stress: cgroup CPU/throttling values before and during injection.
- Memory stress: cgroup memory before, during, and after injection, captured in
  chronological order.
- I/O latency: the `toda` mount plus measured read-only operation duration
  before and during injection.
- EIO: a read-only existing-file probe that succeeds before, returns EIO
  during injection, and creates no file.
- DNS: the same Keeper lookup succeeds before, fails during, and succeeds
  after injection; an unrelated name remains resolvable during the fault.
- Time: one already-running ClickHouse timestamp stream shows normal, skewed,
  and restored time.
- Combined fault: prove both components independently while simultaneously
  active.
- Recovery soak: show UID changes and the complete recovery gate between
  cycles.
- PVC loss: show old/new pod, PVC, and PV identities, missing local table,
  operator repair, equal donor/rebuilt data, and successful new writes.

## Data and Safety

- The approved reset permanently removes the current ClickHouse data PVCs.
- No chaos proof manually writes files under `/var/lib/clickhouse`.
- Experiment 25 deletes only the explicitly named shard-0 replica-1 PVC and
  pod after recording their identities.
- Workload timeouts remain failed or ambiguous; they are never counted as
  acknowledged writes.
- Workload counters and row totals must never decrease between experiments.

## Acceptance Criteria

- All 25 experiment sections satisfy the evidence contract.
- No observation contradicts the command output beside it.
- Test and final-summary counters form one monotonic chronological history.
- Every replica pair converges to matching `count()`, `uniqExact(id)`, and
  `sum(payload)` values.
- Keeper ends with exactly one leader and two followers.
- No Chaos Mesh experiment remains, all expected PVCs are Bound, all database
  pods are Ready, and ClickHouse reports `Ready`.
- Markdown fences are balanced and `git diff --check` passes.
