# Verdis

Versioned Redis: a key value store with multi version concurrency control (MVCC) and historical data access.

## Why Verdis Exists

Most key value stores only know the present. Overwrite a key and the old value is gone. Need to know what changed, when it changed, or roll back a bad write? You are on your own.

That is a problem when history matters. Audit trails, debugging production issues, tracking state changes over time, recovering from mistakes. These are common needs that end up handled with workarounds like timestamped keys or separate audit logs.

Verdis treats history as a first class stuff. Every write creates a version. Every version stays accessible. Rollback is a single command.

## Roadmap

This is the current plan. It might change as things progress.

### Phase 1: Foundation

| Component | Status |
|-----------|--------|
| RESP protocol parser | Done |
| TCP server with connection handling | Done |
| MVCC engine core | in WIP (We are working on it :)) |
| Basic commands (GET, SET, DEL, EXISTS, PING) | Done |

### Phase 2: Version Control Commands

| Component | Status |
|-----------|--------|
| Historical reads (GET key@version) | WIP |
| HISTORY command | WIP |
| ROLLBACK command | WIP |

We are searching on it and we will see what happens

### Phase 3: Persistence (Planned)

Write ahead log, LSM tree storage engine, compaction. Data survives restarts


That is the war that we are creating for ourselves. Let's see how it goes and how we become older quickly :)

Mushie was here. If you are here gang, you gotta also suffer :)