# CatX-UI Upstream v3.9.0 Sync Evidence

Status: `CURRENT`

This document records evidence for the actual upstream integration. It is
not a duplicate of the general maintenance procedure in
`docs/04_UPSTREAM_SYNC.md`.

## Candidate and ancestry

| Item                          | Value                                      |
| ----------------------------- | ------------------------------------------ |
| CatX integration base         | `65957a8988b5b51c4ac01dbd1121f04b55af6c8a` |
| Strategy merge SHA            | `7cb97e13def59280e92e5667d3f43ac23db8d470` |
| Develop reconciliation SHA    | `65957a8988b5b51c4ac01dbd1121f04b55af6c8a` |
| Upstream old tag              | `v3.8.5`                                   |
| Upstream old SHA              | `7ef22f94c950ff09f0870e2295fa65ad5968742c` |
| Upstream new tag              | `v3.9.0`                                   |
| Upstream new SHA              | `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb` |
| Upstream commit count         | `99`                                       |
| Changed-path count            | `563`                                      |
| Sensitive-path count          | `224`                                      |
| CatX-overlap count            | `64`                                       |
| Sync branch                   | `sync/upstream-v3.9.0`                     |
| Upstream merge SHA            | pending conflict resolution                |
| Final qualified candidate SHA | pending                                    |
| Sync PR                       | pending                                    |
| Develop merge SHA             | pending                                    |
| Final develop SHA             | pending                                    |

## Conflict inventory

The dry run reported 12 conflicts. The actual conflict list is captured from
`git diff --name-only --diff-filter=U` immediately after the real merge and
replaced here if it differs.

| Path                                  | Category                  | Resolution | Evidence                                      |
| ------------------------------------- | ------------------------- | ---------- | --------------------------------------------- |
| `.github/workflows/release.yml`       | RELEASE/UPDATER           | kept CatX variable-based Xray URL; identity correction follows separately | CatX release ownership preserved |
| `frontend/package-lock.json`          | GENERATED / FRONTEND/UX   | resolved package source, then regenerated with npm | no hand-merged lockfile |
| `frontend/package.json`               | FRONTEND/UX               | adopted upstream toolchain versions; retained CatX overrides and script policy | package source is upstream plus CatX safety constraints |
| `go.mod`                              | UPSTREAM-ONLY / DEPENDENCY | adopted upstream Go/Xray/quic requirements and retained CatX Prometheus dependency | `go mod tidy` completed |
| `internal/config/version`             | RELEASE/UPDATER           | kept deleted; CatX `forkrelease` remains authoritative | no upstream release identity reintroduced |
| `internal/web/service/client_traffic.go` | SEMANTIC-CONFLICT       | combined upstream TUIC/node reset delivery with CatX quota-preserving reset and re-enable | quota ownership remains in CatX layer |
| `internal/web/service/inbound_node.go` | SEMANTIC-CONFLICT        | adopted upstream `clientFrozen` reset semantics while preserving CatX master lifecycle rules | node lifecycle boundary reviewed |
| `internal/web/service/inbound_traffic_apply.go` | SEMANTIC-CONFLICT | retained both CatX group-quota reset ownership and upstream renewed-email handling | reset ordering preserved |
| `internal/web/service/panel/panel.go` | RELEASE/UPDATER           | retained CatX prerelease-aware `forkrelease` comparator | upstream numeric helper remains available elsewhere |
| `internal/web/service/panel/panel_test.go` | GENERATED / RELEASE/UPDATER | retained CatX strict and prerelease version tests | release identity tests remain CatX-owned |
| `internal/web/web.go`                 | CATX-HOOK / SEMANTIC-CONFLICT | combined fork stop/reload lifecycle with upstream HTTP and TUIC graceful shutdown | Restart Panel keeps Xray running |
| `update.sh`                           | RELEASE/UPDATER           | retained transactional CatX updater with snapshot, validation, healthcheck, and rollback | upstream direct updater not adopted |

## Resolution summary

The 12 conflicts were resolved semantically using the base, CatX side, upstream
side, and relevant upstream commits. No repository-wide `ours` or `theirs`
resolution was used. CatX release ownership, updater transactionality, fork
runtime lifecycle, quota ownership, and feature registration boundaries remain
explicit after the merge.

High-risk review areas are database/startup lifecycle, migrations, client and
inbound runtime, subscriptions, TUIC, AmneziaWG, Xray identity/version,
frontend/generated contracts, updater/release ownership, and feature-off
compatibility.

## Verification

| Check                               | Result  | Evidence                                 |
| ----------------------------------- | ------- | ---------------------------------------- |
| `make verify`                       | pending | hosted/local result to be recorded       |
| `make verify-fork`                  | pending | hosted/local result to be recorded       |
| Go tests                            | pending | hosted/local result to be recorded       |
| Go race tests                       | pending | hosted/local result to be recorded       |
| SQLite migration qualification      | pending | fresh and supported upgrade              |
| PostgreSQL migration qualification  | pending | fresh and supported upgrade              |
| Generated drift                     | pending | authoritative generators                 |
| Frontend lint/typecheck/tests/build | pending | hosted result                            |
| Frontend visual gate                | pending | required only if declared by affected UX |
| Xray config/start/reload            | pending | candidate and runtime smoke              |
| Restart Panel regression            | pending | real endpoint lifecycle                  |
| Multi-node/runtime                  | pending | hosted/disposable evidence               |
| Traffic Control                     | pending | capability/apply/reconcile/rollback      |
| Feature-off compatibility           | pending | integrated upstream baseline             |
| Non-publishing release matrix       | pending | exact candidate                          |
| Disposable Linux smoke              | pending | synthetic data only                      |
| Native TUIC smoke                   | pending | if active in candidate                   |
| Dependency audit                    | pending | production dependency graph              |
| `git diff --check`                  | pending | final candidate                          |

## Final record

- final sync merge SHA: pending;
- post-merge correction commits: pending;
- final exact qualified candidate SHA: pending;
- final develop integration SHA: pending;
- Xray version after sync: `26.9.30`;
- migration/recovery result: pending;
- release publication: explicitly not performed.
