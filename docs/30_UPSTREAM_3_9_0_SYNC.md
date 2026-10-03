# CatX-UI Upstream v3.9.0 Sync Evidence

Status: `CURRENT`

This document records evidence for the actual upstream integration. It is
not a duplicate of the general maintenance procedure in
`docs/04_UPSTREAM_SYNC.md`.

## Candidate and ancestry

| Item                          | Value                                                  |
| ----------------------------- | ------------------------------------------------------ |
| CatX integration base         | `65957a8988b5b51c4ac01dbd1121f04b55af6c8a`             |
| Strategy merge SHA            | `7cb97e13def59280e92e5667d3f43ac23db8d470`             |
| Develop reconciliation SHA    | `65957a8988b5b51c4ac01dbd1121f04b55af6c8a`             |
| Upstream old tag              | `v3.8.5`                                               |
| Upstream old SHA              | `7ef22f94c950ff09f0870e2295fa65ad5968742c`             |
| Upstream new tag              | `v3.9.0`                                               |
| Upstream new SHA              | `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb`             |
| Upstream commit count         | `99`                                                   |
| Changed-path count            | `563`                                                  |
| Sensitive-path count          | `224`                                                  |
| CatX-overlap count            | `64`                                                   |
| Sync branch                   | `sync/upstream-v3.9.0`                                 |
| Upstream merge SHA            | `9335eb4af5ab4c73976fdcbeda875926400e53f8`             |
| Final qualified candidate SHA | `3b25757f` (qualification rerun pending)               |
| Sync PR                       | [#4](https://github.com/CatCodeArbelin/CatX-UI/pull/4) |
| Develop merge SHA             | pending                                                |
| Final develop SHA             | pending                                                |

## Conflict inventory

The dry run reported 12 conflicts. The actual conflict list is captured from
`git diff --name-only --diff-filter=U` immediately after the real merge and
replaced here if it differs.

| Path                                            | Category                      | Resolution                                                                                   | Evidence                                                |
| ----------------------------------------------- | ----------------------------- | -------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| `.github/workflows/release.yml`                 | RELEASE/UPDATER               | kept CatX variable-based Xray URL; identity correction follows separately                    | CatX release ownership preserved                        |
| `frontend/package-lock.json`                    | GENERATED / FRONTEND/UX       | resolved package source, then regenerated with npm                                           | no hand-merged lockfile                                 |
| `frontend/package.json`                         | FRONTEND/UX                   | adopted upstream toolchain versions; retained CatX overrides and script policy               | package source is upstream plus CatX safety constraints |
| `go.mod`                                        | UPSTREAM-ONLY / DEPENDENCY    | adopted upstream Go/Xray/quic requirements and retained CatX Prometheus dependency           | `go mod tidy` completed                                 |
| `internal/config/version`                       | RELEASE/UPDATER               | kept deleted; CatX `forkrelease` remains authoritative                                       | no upstream release identity reintroduced               |
| `internal/web/service/client_traffic.go`        | SEMANTIC-CONFLICT             | combined upstream TUIC/node reset delivery with CatX quota-preserving reset and re-enable    | quota ownership remains in CatX layer                   |
| `internal/web/service/inbound_node.go`          | SEMANTIC-CONFLICT             | adopted upstream `clientFrozen` reset semantics while preserving CatX master lifecycle rules | node lifecycle boundary reviewed                        |
| `internal/web/service/inbound_traffic_apply.go` | SEMANTIC-CONFLICT             | retained both CatX group-quota reset ownership and upstream renewed-email handling           | reset ordering preserved                                |
| `internal/web/service/panel/panel.go`           | RELEASE/UPDATER               | retained CatX prerelease-aware `forkrelease` comparator                                      | upstream numeric helper remains available elsewhere     |
| `internal/web/service/panel/panel_test.go`      | GENERATED / RELEASE/UPDATER   | retained CatX strict and prerelease version tests                                            | release identity tests remain CatX-owned                |
| `internal/web/web.go`                           | CATX-HOOK / SEMANTIC-CONFLICT | combined fork stop/reload lifecycle with upstream HTTP and TUIC graceful shutdown            | Restart Panel keeps Xray running                        |
| `update.sh`                                     | RELEASE/UPDATER               | retained transactional CatX updater with snapshot, validation, healthcheck, and rollback     | upstream direct updater not adopted                     |

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

| Check                               | Result       | Evidence                               |
| ----------------------------------- | ------------ | -------------------------------------- |
| `make verify`                       | success      | hosted `verify-fork`                   |
| `make verify-fork`                  | success      | hosted `verify-fork`                   |
| Go tests                            | success      | hosted `go-test`                       |
| Go race tests                       | success      | hosted `race`                          |
| SQLite migration qualification      | success      | hosted fork verification               |
| PostgreSQL migration qualification  | success      | hosted `postgres-durable-first`        |
| Generated drift                     | success      | hosted `codegen` and fork verification |
| Frontend lint/typecheck/tests/build | success      | hosted `frontend`                      |
| Frontend visual gate                | not required | no visual-approval gate declared       |
| Xray config/start/reload            | success      | hosted fork verification               |
| Restart Panel regression            | success      | hosted fork verification               |
| Multi-node/runtime                  | success      | hosted fork verification               |
| Traffic Control                     | success      | hosted fork verification               |
| Feature-off compatibility           | success      | hosted fork verification               |
| Non-publishing release matrix       | success      | hosted build/artifact qualification    |
| Disposable Linux smoke              | skipped      | workflow gate not enabled for this PR  |
| Native TUIC smoke                   | success      | hosted Go, race, and fork verification |
| Dependency audit                    | success      | hosted `govulncheck`                   |
| `git diff --check`                  | success      | local and hosted verification          |

## Final record

- final sync merge SHA: `9335eb4af5ab4c73976fdcbeda875926400e53f8`;
- post-merge correction commits: `c38b640a`, `f3168d86`, `a46aa3fd`, `7b64e878`, `b2627c61`, `3b25757f`;
- final exact qualified candidate SHA: pending final hosted rerun;
- final develop integration SHA: pending;
- Xray version after sync: `26.9.30`;
- migration/recovery result: hosted migration, updater transaction, and rollback checks passed;
- release publication: explicitly not performed.
