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
