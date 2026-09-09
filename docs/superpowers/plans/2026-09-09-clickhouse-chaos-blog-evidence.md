# ClickHouse Chaos Blog Evidence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the ClickHouse chaos article with one chronological, fully evidenced execution of all 25 experiments against a fresh cluster.

**Architecture:** Recreate the existing `clickhouse-chaos` topology in `demo`, deploy one stateful workload counter, and execute every experiment sequentially. Each experiment records a baseline, direct fault-specific evidence, workload impact, recovery, and the common integrity gate before the next experiment starts.

**Tech Stack:** K3s, kubectl, KubeDB ClickHouse 26.2.6, ClickHouse Keeper, Chaos Mesh 2.8.4, Bash, Hugo Markdown.

**Spec:** `docs/superpowers/specs/2026-09-09-clickhouse-chaos-blog-evidence-design.md`

## Global Constraints

- Use namespace `demo` and resource name `clickhouse-chaos`.
- Use two shards, two replicas per shard, three Keeper members, 4Gi data PVCs, and a 1 CPU ClickHouse limit.
- Run tests 1 through 25 in strict numeric order against one dataset.
- Never start the next test until the recovery gate passes.
- Do not publish shell loops or workload-pod variables.
- Put each command's output immediately after that command.
- Do not manually write a probe file under `/var/lib/clickhouse`.
- Treat timed-out writes as failed or ambiguous, never acknowledged.
- Preserve unrelated work already present on branch `ch-chaos`.

---

### Task 1: Reset and provision the test environment

**Files:**
- Read: `content/post/chaos-testing-clickhouse/index.md`
- Read: `/home/shuvo/blog/clickhouse-chaos-testing-fresh-runbook.md`
- Modify later: `content/post/chaos-testing-clickhouse/index.md`

**Interfaces:**
- Consumes: the approved destructive-reset authorization and manifests already embedded in the article/runbook.
- Produces: a fresh Ready `clickhouse-chaos` cluster with zero workload counters and no active chaos resources.

- [ ] **Step 1: Record and remove active test objects**

Run:

```bash
kubectl get podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos -n demo
kubectl delete podchaos,networkchaos,stresschaos,iochaos,dnschaos,timechaos -n demo --all
```

Expected: no test fault remains.

- [ ] **Step 2: Delete workload before deleting credentials**

Run:

```bash
kubectl delete deployment,configmap -n demo clickhouse-chaos-workload --ignore-not-found
```

Expected: the Deployment and ConfigMap are absent.

- [ ] **Step 3: Delete the existing ClickHouse resource and its WipeOut PVCs**

Run:

```bash
kubectl delete clickhouse -n demo clickhouse-chaos --wait=true --timeout=15m
kubectl get pods,pvc -n demo | grep clickhouse-chaos
```

Expected: the second command returns no ClickHouse or Keeper pod/PVC.

- [ ] **Step 4: Apply the exact `setup/clickhouse-chaos.yaml` manifest from the article**

Run the manifest with `kubectl apply -f -`, then:

```bash
kubectl wait -n demo --for=jsonpath='{.status.phase}'=Ready clickhouse/clickhouse-chaos --timeout=15m
kubectl get clickhouse,petset,pods,pvc -n demo
```

Expected: ClickHouse is `Ready`; four ClickHouse and three Keeper pods are Running; four 4Gi data PVCs and three 1Gi Keeper PVCs are Bound.

- [ ] **Step 5: Create schema and workload from the exact article manifests**

Run the schema multiquery from the article, apply `setup/clickhouse-workload.yaml`, discover the generated workload pod name, and pause it after ten acknowledged batches.

Expected: at least ten attempts and ten successful batches, with zero failed
batches. Capture the actual values rather than forcing an exact stopping
point. For example:

```text
attempted=12
successful=12
failed=0
```

- [ ] **Step 6: Capture the initial full recovery gate**

Run the article's commands separately for ClickHouse status, seven pods, distributed integrity, four local-replica integrity values, four `system.replicas` rows, three Keeper roles, four PID states, and four mounts.

Expected: matching replica pairs, writable empty queues, one Keeper leader, two followers, PID state without `T`, and ext4 data mounts.

### Task 2: Re-evidence pod and data-plane failures (Tests 1–6)

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md:968`

**Interfaces:**
- Consumes: fresh cluster and paused workload from Task 1.
- Produces: chronological evidence for tests 1–6 and a healthy gate after each test.

- [ ] **Step 1: Test 1 single replica pod kill**

Record workload counters and target UID; resume workload; apply `tests/01-pod-kill.yaml`; wait for `AllInjected`; show target absence/recreation, the ClickHouse phase during replacement, and new UID; delete the object; pause workload; capture counters and the full gate.

- [ ] **Step 2: Test 2 timed replica pod failure**

Record target UID, readiness, restart count, and counters; apply the 45-second fault; show `AllInjected=True`, target NotReady/restarts, and ClickHouse transition; wait for `AllRecovered`; pause, delete, and show the same UID, final restart delta, counters, and full gate.

- [ ] **Step 3: Test 3 container kill**

Record pod UID and ClickHouse-container restart count before applying the fault. After `AllInjected`, show the same UID and an increased restart count, then delete, pause, and run the gate.

- [ ] **Step 4: Test 4 three alternating pod kills**

For each manifest `04-a`, `04-b`, and `04-c`, print the selected pod UID before applying, wait for `AllInjected`, delete the one-shot resource, wait for Ready, and print the new UID. Run the full gate and capture counters between each cycle.

- [ ] **Step 5: Test 5 complete shard-0 outage**

Before injection show both shard-0 pods Ready and a successful Distributed query. During `AllInjected`, show both targets unavailable and capture a Distributed query returning `ALL_CONNECTION_TRIES_FAILED`. After `AllRecovered`, show both pods Ready and equal shard-0 integrity values.

- [ ] **Step 6: Test 6 complete data-plane outage**

Before injection show four Ready data pods and successful SQL. During `AllInjected`, show all four targets unavailable and capture a failed SQL connection from the workload pod. After `AllRecovered`, show the same PVC identities, all pods Ready, and the full gate.

### Task 3: Re-evidence Keeper failures (Tests 7–10)

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md:1857`

**Interfaces:**
- Consumes: healthy post-Test-6 Keeper quorum.
- Produces: role-accurate Keeper failure evidence with no hard-coded stale leader assumption.

- [ ] **Step 1: Capture all Keeper roles immediately before Tests 7 and 8**

Run `mntr` separately on Keeper-0, Keeper-1, and Keeper-2. Edit the YAML shown in the article so Test 7 targets an observed follower and Test 8 targets the observed leader.

- [ ] **Step 2: Test 7 follower kill**

Record target UID, apply and wait for `AllInjected`, show the target UID changes while the existing leader still answers, then delete and run the full gate including all three roles.

- [ ] **Step 3: Test 8 leader kill**

Record the current leader and its UID, apply the manifest targeting that exact pod, show a different surviving Keeper becomes leader, show the replacement UID, then verify exactly one leader and two followers after recovery.

- [ ] **Step 4: Test 9 quorum loss**

Show a healthy leader before injection. During the two-member failure, show both targets unavailable and the survivor returning `This instance is not currently serving requests`; capture workload counters. After recovery print all three roles and empty replication queues.

- [ ] **Step 5: Test 10 full Keeper outage**

Before injection show all three `mntr` responses. During injection show direct 4lw/connection failures to each Keeper separately and a replicated-write failure or ambiguity. After recovery show one leader, two followers, and the full gate.

### Task 4: Re-evidence network faults (Tests 11–16)

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md:2393`

**Interfaces:**
- Consumes: healthy data and Keeper topology.
- Produces: direct before/during/after network measurements rather than `Ready`-only evidence.

- [ ] **Step 1: Test 11 delay**

Resolve the selected pod IP with `kubectl get pod -n demo clickhouse-chaos-chaos-cluster-shard-0-0 -o jsonpath='{.status.podIP}{"\n"}'`, then run `ping -c 5 -W 2` from the workload pod and capture round-trip statistics. Apply 500ms delay, wait for `AllInjected`, repeat the identical ping and capture the increased RTT, then repeat after recovery using the concrete IP printed by the first command.

- [ ] **Step 2: Test 12 packet loss**

Resolve the selected pod IP, then run `ping -c 20 -W 1` from the workload pod. Capture zero loss before injection, observable loss during `AllInjected`, and zero loss after recovery using the same concrete IP.

- [ ] **Step 3: Test 13 packet duplication**

Resolve the selected pod IP and run `ping -c 20 -W 1` from the workload pod before, during, and after injection. Require the active output to contain `DUP!` or a positive duplicate count. If no duplicate is observable, record the experiment as unsupported in this environment rather than declaring PASS. Afterward prove `count() == uniqExact(id)`.

- [ ] **Step 4: Test 14 1 Mbps bandwidth**

Set the published manifest direction to `both`. From the workload pod, run `/usr/bin/time -f 'elapsed=%e' clickhouse-client` directly against shard-0 replica-0 with `SELECT repeat('x', 2097152) FORMAT TSV`, redirecting query output to `/dev/null`. Record elapsed time before and during the 1 Mbps cap, calculate throughput for the 2 MiB response, and repeat after recovery.

- [ ] **Step 5: Test 15 data-peer partition**

From the selected replica, run `timeout 3 bash -c 'exec 3<>/dev/tcp/clickhouse-chaos-chaos-cluster-shard-0-1.clickhouse-chaos-pods.demo.svc/9000'; echo exit_code=$?`. Require exit code 0 before, 124 during, and 0 after the partition. Capture replication state and matching shard data after recovery.

- [ ] **Step 6: Test 16 Keeper partition**

From the selected replica, use `timeout 3 bash -c 'exec 3<>/dev/tcp/clickhouse-chaos-keeper-0.clickhouse-chaos-keeper-pods.demo.svc/9181'; echo exit_code=$?` separately for the connectivity transition. Require exit code 0 before, 124 during, and 0 after; also capture `is_readonly=1`/`is_session_expired=1` during and `is_readonly=0`, empty queue, and two active replicas after recovery.

### Task 5: Re-evidence stress, storage, DNS, and time faults (Tests 17–22)

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md:3248`

**Interfaces:**
- Consumes: healthy post-network cluster.
- Produces: correctly ordered quantitative resource, storage, DNS, and time evidence.

- [ ] **Step 1: Test 17 CPU stress**

Capture `cpu.stat`, restart count, and counters before injection. During `AllInjected`, capture a second `cpu.stat` and show increased throttling. After recovery show unchanged restart count, Ready status, counters, and full gate.

- [ ] **Step 2: Test 18 memory stress**

Capture `memory.current`, `memory.max`, restart count, and `oom_kill` before injection. Apply and wait for `AllInjected`, capture higher memory while active, wait for `AllRecovered`, delete, then capture reduced memory, unchanged restart/`oom_kill`, and full gate. Never delete before the live measurement.

- [ ] **Step 3: Test 19 filesystem latency**

Run `/usr/bin/time -f 'elapsed=%e' sh -c 'find /var/lib/clickhouse/store -type f | head -n 20 | xargs stat >/dev/null'` before injection and while the mount is `toda`; require a measurable increase consistent with 100ms delays across the bounded 20-file read-only probe. After recovery show ext4 and PID state. Run `kill -CONT 1` only when the captured PID state contains `T`, then prove equal replicas.

- [ ] **Step 4: Test 20 EIO regression**

Repeat `find /var/lib/clickhouse/store -type f -exec stat {} +` before and during injection, update its chronological counters, and verify ext4, PID, and equal replicas after recovery. Do not create or delete a probe file.

- [ ] **Step 5: Test 21 DNS error**

Resolve the exact Keeper FQDN before applying. During `AllInjected`, show the same lookup exits 2 while `kubernetes.default.svc.cluster.local` still resolves. After `AllRecovered`, show the exact Keeper FQDN resolves again. Remove the existing pre-apply failure/recovery block.

- [ ] **Step 6: Test 22 clock skew**

Start one timestamp stream before applying TimeChaos. Capture normal, minus-two-hour, and restored timestamps from that single stream; inspect PID after recovery, run `kill -CONT 1` only when the captured state contains `T`, then run the full gate.

### Task 6: Re-evidence combined, soak, and PVC-loss scenarios (Tests 23–25)

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md:4486`

**Interfaces:**
- Consumes: one continuous dataset through Test 22.
- Produces: final chronological dataset and destructive replica-rebuild proof.

- [ ] **Step 1: Test 23 combined I/O latency and sibling failure**

Capture baseline read-only probe duration and sibling readiness. Apply both resources, wait for both `AllInjected`, show increased probe duration on replica-0 and direct unavailability of replica-1, then capture counters and ClickHouse status. After both `AllRecovered`, show ext4, PID, sibling readiness, equal shard data, and full gate.

- [ ] **Step 2: Test 24 three-cycle soak**

For each cycle capture UID and workload counters, inject, show new UID, delete, pause, and run the complete gate including checksums and queues before the next cycle. Preserve separate commands and outputs for every pod.

- [ ] **Step 3: Test 25 pod/PVC loss**

Pause workload; record donor/target integrity plus target pod/PVC/PV identities; delete only shard-0 replica-1 PVC and pod; show new identities and missing local table; capture operator recovery logs; wait for automatic schema/data repair; prove donor/rebuilt equality; resume and pause workload to prove new writes.

### Task 7: Rewrite the article from the evidence ledger

**Files:**
- Modify: `content/post/chaos-testing-clickhouse/index.md`

**Interfaces:**
- Consumes: captured outputs from Tasks 1–6.
- Produces: the final publishable article.

- [ ] **Step 1: Replace setup and workload outputs**

Use only fresh cluster identities, pod names, ages, versions, and initial counters.

- [ ] **Step 2: Replace every Test 1–25 transition**

For each test, order content as create manifest, explain fault, expected behavior, resume workload, baseline, apply, `AllInjected`, direct impact, `AllRecovered`, pause, delete, recovery checks, observed behavior, and result.

- [ ] **Step 3: Recalculate the result summary**

Every table row must quote evidence printed in its experiment section. Remove PASS for any experiment whose configured fault cannot be observed in this environment.

- [ ] **Step 4: Replace final counters and integrity evidence**

Use the final gate after Test 25. Verify counters and row totals never decrease when reading the article from Test 1 through Test 25.

### Task 8: Final verification

**Files:**
- Verify: `content/post/chaos-testing-clickhouse/index.md`

**Interfaces:**
- Consumes: rewritten article and recovered cluster.
- Produces: evidence-backed completion report.

- [ ] **Step 1: Run structural checks**

Run:

```bash
git diff --check
awk '/^### Chaos#[0-9]+:/{c++} /^#### Resume the workload$/{r++} /^#### Pause the workload$/{p++} /^```/{f++} END {printf "chaos=%d resume=%d pause=%d fences=%d balanced=%s\n",c,r,p,f,(f%2==0?"yes":"no")}' content/post/chaos-testing-clickhouse/index.md
```

Expected: 25 chaos headings, the intentional workload headings, and balanced fences.

- [ ] **Step 2: Scan for stale contradictions**

Run `rg` for every old UID, stale counter, stale IP, `261`, `239202`, `59500`, `2461`, `590 to 645`, and removed write probe. Expected: none remains unless explicitly explained as historical invalid evidence.

- [ ] **Step 3: Run the live final gate**

Run all ClickHouse, pod, chaos-resource, distributed/local integrity, `system.replicas`, Keeper, PID, mount, and PVC commands separately.

Expected: ClickHouse Ready, seven Ready database pods, zero chaos objects, matching replicas, writable empty queues, one leader/two followers, runnable PID 1, ext4 mounts, and Bound PVCs.

- [ ] **Step 4: Render if Hugo is available**

Run:

```bash
hugo --config=config.dev.yaml --buildDrafts --buildFuture -d public/blog
```

Expected: exit code 0. If Hugo is unavailable, report that limitation without claiming a rendered validation.
