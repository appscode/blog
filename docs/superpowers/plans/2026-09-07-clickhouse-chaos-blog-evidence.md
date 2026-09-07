# ClickHouse Chaos Blog Evidence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rerun all 25 ClickHouse chaos experiments and publish a single article that demonstrates injection, impact, recovery, and integrity with one command followed immediately by its output.

**Architecture:** One KubeDB ClickHouse named `clickhouse-chaos` runs throughout the campaign with two shards, two replicas per shard, three managed Keeper members, 4Gi data PVCs, and a 1-CPU ClickHouse limit. A continuous UUID-keyed workload records acknowledged and ambiguous batches. Every fault must pass the shared recovery gate before the next fault starts, and experiment 25 reuses the same cluster and data.

**Tech Stack:** Kubernetes, KubeDB ClickHouse 26.2.6, ClickHouse Keeper, Chaos Mesh 2.8.4, kubectl, Bash, Hugo Markdown

**Spec:** `docs/superpowers/specs/2026-09-07-clickhouse-chaos-blog-evidence-design.md`

## Global Constraints

- Use `demo` and the resource name `clickhouse-chaos`.
- Use two shards, two replicas, and three managed Keeper members.
- Use 4Gi data PVCs and a 1-CPU ClickHouse limit.
- Do not update operator or database images.
- Use the same cluster for experiments 1 through 25.
- Preserve unrelated resources in `demo`.
- Put every command's output immediately after that command.

### Task 1: Create the fresh cluster and workload

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Update the embedded ClickHouse manifest to request 4Gi data storage, 1 CPU, and enough memory to avoid the earlier workload-driven OOM.
- [ ] Apply the manifest and wait for KubeDB `Ready`.
- [ ] Capture `kubectl get clickhouse`, pod, PVC, and Keeper-role output separately.
- [ ] Create the replicated local table, Distributed table, and workload Deployment.
- [ ] Capture the initial workload counters and data-integrity baseline.

### Task 2: Run pod and Keeper experiments 1–10

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Run each pod/container/shard/data-plane experiment individually and record baseline, `AllInjected`, live status, `AllRecovered`, recovery gate, and cleanup.
- [ ] Rediscover Keeper roles immediately before role-specific tests.
- [ ] Run follower, leader, quorum-loss, and full-Keeper-loss experiments and record role transitions and workload impact.

### Task 3: Run network experiments 11–16

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Run delay, loss, duplication, bandwidth, data-peer partition, and Keeper partition experiments.
- [ ] Record client counters, ClickHouse phase, replica state, and recovery for every fault.

### Task 4: Run resource experiments 17–18

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Run CPU stress and capture throttling, pod restart count, workload impact, and recovery.
- [ ] Run bounded memory stress and compare `memory.current` with `memory.max` before, during, and after cleanup.

### Task 5: Run storage experiments 19–20

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Run filesystem latency and recoverable EIO faults.
- [ ] Capture mount state, ClickHouse PID state, injected storage errors, cleanup behavior, and the full recovery gate.
- [ ] Use `SIGCONT` only when PID 1 is demonstrably stopped after Chaos Mesh cleanup.

### Task 6: Run DNS and time experiments 21–22

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Run Keeper DNS failure and prove the DNS error independently from cached Keeper sessions.
- [ ] Run two-hour clock skew, prove shifted timestamps, and record process state during cleanup.

### Task 7: Run combined, soak, and PVC-loss experiments 23–25

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Run combined I/O latency and sibling failure, then pass the full gate.
- [ ] Run three sequential recovery-soak pod kills with a full gate after each cycle.
- [ ] On the same cluster, record the selected replica pod/PVC/PV identities, delete its PVC and pod, and prove new identities, sibling restoration, replica equality, and successful new writes.

### Task 8: Normalize the complete article

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Give every experiment the approved twelve-part evidence sequence.
- [ ] Replace aggregate console transcripts with one command and one immediately following output block.
- [ ] Explain expected errors and status transitions in plain language.
- [ ] Update the results table and final integrity evidence from the fresh campaign.

### Task 9: Verify the publication artifact

**Files:**
- Verify: `content/post/chaos-testing-clickhouse/index.md`

- [ ] Confirm there are exactly 25 experiment headings and 25 result rows.
- [ ] Confirm Markdown fences are balanced and `git diff --check` succeeds.
- [ ] Confirm no stale resource names or second ClickHouse deployment remain.
- [ ] Run the Hugo build when Hugo is installed; otherwise report the unavailable tool without claiming a site-build pass.
