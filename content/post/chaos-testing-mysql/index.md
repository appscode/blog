---
title: "Chaos Engineering Results: KubeDB MySQL Group Replication Achieves Zero Data Loss Across 48 Experiments"
date: "2026-04-08"
weight: 14
authors:
- SK Ali Arman
tags:
- chaos-engineering
- chaos-mesh
- database
- kubedb
- kubernetes
- mysql
- group-replication
- high-availability
---

## Overview

We conducted **48 chaos experiments** across **3 MySQL versions** (8.0.36, 8.4.8, 9.6.0) and **2 Group Replication topologies** (Single-Primary and Multi-Primary) on KubeDB-managed 3-node clusters. The goal: validate that KubeDB MySQL delivers **zero data loss**, **automatic failover**, and **self-healing recovery** under realistic failure conditions with production-level write loads.

**The result: every experiment passed with zero data loss, zero split-brain, and zero errant GTIDs.**

This post summarizes the methodology, results, and key findings from the most comprehensive chaos testing effort we have run on KubeDB MySQL to date.

## Why Chaos Testing?

Running databases on Kubernetes introduces failure modes that traditional infrastructure does not have — pods can be evicted, nodes can go down, network policies can partition traffic, and resource limits can trigger OOMKills at any time. Chaos engineering deliberately injects these failures to verify that the system recovers correctly **before** they happen in production.

For a MySQL Group Replication cluster managed by KubeDB, we needed to answer:

- Does the cluster **lose data** when a primary is killed mid-transaction?
- Does **automatic failover** work under network partitions?
- Can the cluster **self-heal** after a full outage with no manual intervention?
- Are **GTIDs consistent** across all nodes after recovery?
- Does the cluster survive **combined failures** (CPU + memory + load simultaneously)?

## Test Environment

| Component | Details |
|---|---|
| Cluster Topology | 3-node Group Replication (Single-Primary & Multi-Primary) |
| MySQL Versions | 8.0.36, 8.4.8, 9.6.0 |
| Storage | 2Gi PVC per node (Durable, ReadWriteOnce) |
| Memory Limit | 1.5Gi per MySQL pod |
| CPU Request | 500m per pod |
| Chaos Engine | Chaos Mesh |
| Load Generator | sysbench `oltp_write_only`, 4-12 tables, 4-16 threads |
| Baseline TPS | ~2,400 (Single-Primary) / ~1,150 (Multi-Primary) |

All experiments were run under **sustained sysbench write load** to simulate production traffic during failures. The load generator ran as a Kubernetes Deployment inside the same namespace as the MySQL cluster.

## The 12-Experiment Matrix

Every MySQL version and topology was tested against the same 12-experiment matrix covering single-node failures, resource exhaustion, network degradation, and complex multi-fault scenarios:

| # | Experiment | Chaos Type | What It Tests |
|---|---|---|---|
| 1 | Pod Kill | PodChaos | Ungraceful termination (grace-period=0) |
| 2 | OOMKill | StressChaos / Load | Memory exhaustion beyond pod limits |
| 3 | Network Partition | NetworkChaos | Isolate a node from the cluster |
| 4 | CPU Stress (98%) | StressChaos | Extreme CPU pressure on nodes |
| 5 | IO Latency (100ms) | IOChaos | Disk I/O delays on a node |
| 6 | Network Latency (1s) | NetworkChaos | Replication traffic delays |
| 7 | Packet Loss (30%) | NetworkChaos | Unreliable network across cluster |
| 8 | Combined Stress | StressChaos x3 | Memory + CPU + load simultaneously |
| 9 | Full Cluster Kill | kubectl delete | All 3 pods deleted at once |
| 10 | OOMKill Natural | Load | 128-thread queries to exhaust memory |
| 11 | Scheduled Pod Kill | Schedule | Repeated kills every 30s-1min |
| 12 | Degraded Failover | Workflow | IO latency + pod kill in sequence |

## Data Integrity Validation

Every experiment verified data integrity through **4 checks** across all 3 nodes:

1. **GTID Consistency** — `SELECT @@gtid_executed` must match on all nodes after recovery
2. **Checksum Verification** — `CHECKSUM TABLE` on all sysbench tables must match across nodes
3. **Row Count Validation** — Cumulative tracking table row counts must be preserved
4. **Errant GTID Detection** — No local `server_uuid` GTIDs outside the group UUID

## Results — Single-Primary Mode

### MySQL 9.6.0 — All 12 PASSED

| # | Experiment | Failover | Data Loss | Errant GTIDs | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill Primary | Yes | Zero | 0 | **PASS** |
| 2 | OOMKill Natural | Yes | Zero | 0 | **PASS** |
| 3 | Network Partition | Yes | Zero | 0 | **PASS** |
| 4 | IO Latency (100ms) | No | Zero | 0 | **PASS** |
| 5 | Network Latency (1s) | No | Zero | 0 | **PASS** |
| 6 | CPU Stress (98%) | No | Zero | 0 | **PASS** |
| 7 | Packet Loss (30%) | Yes | Zero | 0 | **PASS** |
| 8 | Combined Stress | Yes (OOMKill) | Zero | 0 | **PASS** |
| 9 | Full Cluster Kill | Yes | Zero | 0 | **PASS** |
| 10 | OOMKill Retry | No (survived) | Zero | 0 | **PASS** |
| 11 | Scheduled Replica Kill | Multiple | Zero | 0 | **PASS** |
| 12 | Degraded Failover | Yes | Zero | 0 | **PASS** |

### MySQL 8.4.8 — All 12 PASSED

| # | Experiment | Failover | Data Loss | Errant GTIDs | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill Primary | Yes | Zero | 0 | **PASS** |
| 2 | OOMKill Stress | No (survived) | Zero | 0 | **PASS** |
| 3 | Network Partition | Yes | Zero | 0 | **PASS** |
| 4 | IO Latency (100ms) | No | Zero | 0 | **PASS** |
| 5 | Network Latency (1s) | No | Zero | 0 | **PASS** |
| 6 | CPU Stress (98%) | No | Zero | 0 | **PASS** |
| 7 | Packet Loss (30%) | Yes | Zero | 0 | **PASS** |
| 8 | Combined Stress | Yes (OOMKill) | Zero | 0 | **PASS** |
| 9 | Full Cluster Kill | Yes | Zero | 0 | **PASS** |
| 10 | OOMKill Natural | No (survived) | Zero | 0 | **PASS** |
| 11 | Scheduled Replica Kill | Multiple | Zero | 0 | **PASS** |
| 12 | Degraded Failover | Yes | Zero | 0 | **PASS** |

### MySQL 8.0.36 — All 12 PASSED

| # | Experiment | Failover | Data Loss | Errant GTIDs | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill Primary | Yes | Zero | 0 | **PASS** |
| 2 | OOMKill Natural | No (survived) | Zero | 0 | **PASS** |
| 3 | Network Partition | Yes | Zero | 0 | **PASS** |
| 4 | IO Latency (100ms) | No | Zero | 0 | **PASS** |
| 5 | Network Latency (1s) | No | Zero | 0 | **PASS** |
| 6 | CPU Stress (98%) | No | Zero | 0 | **PASS** |
| 7 | Packet Loss (30%) | Yes | Zero | 0 | **PASS** |
| 8 | Combined Stress | Yes (OOMKill) | Zero | 0 | **PASS** |
| 9 | Full Cluster Kill | Yes | Zero | 0 | **PASS** |
| 10 | OOMKill Natural (retry) | Yes | Zero | 0 | **PASS** |
| 11 | Scheduled Replica Kill | Multiple | Zero | 0 | **PASS** |
| 12 | Degraded Failover | Yes | Zero | 0 | **PASS** |

## Results — Multi-Primary Mode (MySQL 8.4.8)

In Multi-Primary mode, **all 3 nodes accept writes** — there is no primary/replica distinction. This changes the failure dynamics significantly: no failover election is needed, but Paxos consensus must be maintained across all writable nodes.

| # | Experiment | Data Loss | GTIDs | Checksums | Verdict |
|---|---|---|---|---|---|
| 1 | Pod Kill (random) | Zero | MATCH | MATCH | **PASS** |
| 2 | OOMKill (1200MB stress) | Zero | MATCH | MATCH | **PASS** |
| 3 | Network Partition (3 min) | Zero | MATCH | MATCH | **PASS** |
| 4 | CPU Stress (98%, 3 min) | Zero | MATCH | MATCH | **PASS** |
| 5 | IO Latency (100ms, 3 min) | Zero | MATCH | MATCH | **PASS** |
| 6 | Network Latency (1s, 3 min) | Zero | MATCH | MATCH | **PASS** |
| 7 | Packet Loss (30%, 3 min) | Zero | MATCH | MATCH | **PASS** |
| 8 | Combined Stress (mem+cpu+load) | Zero | MATCH | MATCH | **PASS** |
| 9 | Full Cluster Kill | Zero | MATCH | MATCH | **PASS** |
| 10 | OOMKill Natural (90 JOINs) | Zero | MATCH | MATCH | **PASS** |
| 11 | Scheduled Pod Kill (every 1 min) | Zero | MATCH | MATCH | **PASS** |
| 12 | Degraded Failover (IO + Kill) | Zero | MATCH | MATCH | **PASS** |

**All 12 experiments PASSED with zero data loss.**

## Failover Performance (Single-Primary)

| Scenario | Failover Time | Full Recovery Time |
|---|---|---|
| Pod Kill Primary | ~2-3 seconds | ~30-33 seconds |
| OOMKill Primary | ~2-3 seconds | ~30 seconds |
| Network Partition | ~3 seconds | ~3 minutes |
| Packet Loss (30%) | ~30 seconds | ~2 minutes |
| Full Cluster Kill | ~10 seconds | ~1-2 minutes |
| Combined Stress (OOMKill) | ~3 seconds | ~4 minutes |

## Performance Impact Under Chaos

### Single-Primary Mode

| Chaos Type | TPS During Chaos | Reduction from Baseline (~2,400) |
|---|---|---|
| IO Latency (100ms) | 2-3.5 | 99.9% |
| Network Latency (1s) | 1.2-1.4 | 99.9% |
| CPU Stress (98%) | 1,300-1,370 | ~46% |
| Packet Loss (30%) | Variable | Triggers failover |

### Multi-Primary Mode

| Chaos Type | TPS During Chaos | Impact |
|---|---|---|
| IO Latency (100ms) | 272 | ~73% drop |
| Network Latency (1s) | 1.57 | 99.9% drop |
| CPU Stress (98%) | 0 (writes blocked) | Paxos consensus fails |
| Packet Loss (30%) | 4.98 | 99.6% drop |
| Combined Stress | ~530 then OOMKill | ~44% drop |

## Multi-Primary vs Single-Primary

| Aspect | Multi-Primary | Single-Primary |
|---|---|---|
| Failover needed | No (all primaries) | Yes (election ~2-3s) |
| Write availability | All nodes writable | Only primary writable |
| CPU stress 98% | All writes blocked (Paxos fails) | ~46% TPS reduction |
| IO latency impact | ~73% TPS drop | ~99.9% TPS drop |
| Packet loss 30% | 4.98 TPS (stayed ONLINE) | Triggers failover |
| High concurrency | GR certification conflicts possible | No conflicts (single writer) |
| Recovery mechanism | Rejoin as PRIMARY | Election + rejoin |

## Version Compatibility

| Capability | 8.0.36 | 8.4.8 | 9.6.0 |
|---|---|---|---|
| Pod Kill Recovery | Yes | Yes | Yes |
| OOMKill Recovery | Yes | Yes | Yes |
| Network Partition Recovery | Yes | Yes | Yes |
| CLONE Plugin | Yes | Yes | Yes |
| Single-Primary (12 tests) | **12/12** | **12/12** | **12/12** |
| Multi-Primary (12 tests) | Not tested | **12/12** | Not tested |

## Key Takeaways

1. **KubeDB MySQL achieves zero data loss** across all 48 chaos experiments in both Single-Primary and Multi-Primary topologies.

2. **Automatic failover works reliably** — primary election completes in 2-3 seconds, full recovery in under 4 minutes for all scenarios.

3. **Multi-Primary mode is production-ready** — all 12 experiments passed on MySQL 8.4.8. Be aware that multi-primary has higher sensitivity to CPU stress and network issues due to Paxos consensus requirements on all writable nodes.

4. **Transient GTID mismatches are normal** — brief mismatches (15-30 seconds) during recovery are expected and resolve automatically via GR distributed recovery.

## What's Next

- **Multi-Primary testing on additional MySQL versions** — extend chaos testing to MySQL 8.0.36 and 9.6.0 in Multi-Primary mode
- **Long-duration soak testing** — extended chaos runs (hours/days) to validate stability under sustained failure injection

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
