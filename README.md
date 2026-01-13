# Verdis

Versioned Redis: a key value store with multi version concurrency control (MVCC) and historical data access.

## Why Verdis Exists

Most key value stores only know the present. Overwrite a key and the old value is gone. Need to know what changed, when it changed, or roll back a bad write? You are on your own.

That is a problem when history matters. Audit trails, debugging production issues, tracking state changes over time, recovering from mistakes. These are common needs that end up handled with workarounds like timestamped keys or separate audit logs.

Verdis treats history as a first class stuff. Every write creates a version. Every version stays accessible. Rollback is a single command.

## Roadmap

This is the current plan. It might change as things progress.

1. Foundation: RESP protocol, basic MVCC engine, in memory operations
2. Version Control Commands: historical reads, rollback, diff, blame
3. Persistence: write ahead log, LSM tree storage, compaction
4. Transaction Support: snapshot isolation, atomic commits
5. Operations and Observability: metrics, logging, admin commands
6. Performance Optimization: caching, parallel compaction, connection pooling

That is the war that we are creating for ourselves. Let's see how it goes and how we become older quickly :)

- Mushie was here - If you are here gang, you gotta also suffer :)