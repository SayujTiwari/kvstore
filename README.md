# kvstore

A tiny, Redis-style in-memory key–value database built in Go.  
It focuses on **speed**, **concurrency**, and **crash recovery** (AOF + binary snapshots) to showcase real systems concepts you’d use on infra teams.

## Why
Most student projects are CRUD apps. `kvstore` is an **infrastructure** project: it teaches sharding, synchronization, durability, and performance measurement—the skills big-tech SWE teams care about.

## What it does
- Accepts simple text commands over TCP (one command per line):  
  `PING`, `SET key value`, `GET key`, `DEL key`
- Stores data **in memory** with **sharded RW-locks** for concurrency
- Persists data via **Append-Only Log (AOF)**
- Fast restarts using **binary snapshots + AOF rotation**
- Comes with a **CLI**, **load generator**, **bench tool**, and a **startup timer** for repeatable metrics

## When you’d use it (examples)
- **Caching** computed results to reduce DB load
- **Session storage** for web apps
- **Learning systems**: locks, persistence, file formats, and benchmarking

## Quick Start

### 1) Run the server (PowerShell)
```powershell
# in-memory only (fastest)
$env:KV_AOF="off"; $env:KV_SNAPSHOT="off"
go run .\cmd\kv-server

# or with durability
$env:KV_AOF="on";  $env:KV_SNAPSHOT="on"; $env:KV_FSYNC="everysec"
go run .\cmd\kv-server



## Restart Time (1,000,000 keys, 64-byte values)
| Mode            | startup time (ms) |
|-----------------|-------------------|
| AOF only        | 447               |
| Snapshot + AOF  | 221               |  **(-50.6%)**

**Config:** `KV_AOF=on`, `KV_SNAPSHOT=on|off`, `KV_FSYNC=everysec`

## Throughput (bench: 200 clients × 2k iters)
| Config          | Ops/sec |
|-----------------|---------|
| memory only     | 303k    |
| AOF             | 285k    |
| AOF + Snapshot  | 299k    |

**Server:** Windows (local), Go 1.22, sharded RW-locks  
**Bench cmd:** `go run ./cmd/kv-bench`
