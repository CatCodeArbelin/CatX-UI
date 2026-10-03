# Upstream Maintenance Strategy

This is the canonical operating procedure for maintaining CatX-UI as a low-
divergence downstream fork. It applies to upstream release maintenance only;
it does not authorize a sync, release, or production deployment.

## Source and branch model

| Role | Ref | Policy |
| --- | --- | --- |
| Upstream | `MHSanaei/3x-ui` | Read-only source of upstream stable tags and history. |
| Downstream | `CatCodeArbelin/CatX-UI` | Product, fork modules, qualification, and releases. |
| `upstream` remote | `https://github.com/MHSanaei/3x-ui.git` | Fetch only; its push URL is deliberately disabled. |
| `origin` remote | `https://github.com/CatCodeArbelin/CatX-UI.git` | Downstream branches and review. |
| `main` | production | Only qualified CatX releases. |
| `develop` | integration | Receives reviewed syncs and CatX work after qualification. |
| `sync/upstream-vX.Y.Z` | disposable sync branch | Starts at the selected CatX integration baseline and merges one exact upstream stable tag. Never deploy directly. |
| `feature/*` | focused work | CatX changes, including this strategy package. |
| `release/*` | optional short-lived qualification | Only when a release candidate needs an isolated release gate. |

CatX-specific work lands in focused feature branches, then `develop`. A sync
branch is created from the current integration baseline (`develop` when it is
the intended next integration point; otherwise the explicitly recorded CatX
baseline), never by rebasing a long-lived branch. Stable production follows
CatX tags, not upstream branch heads. CatX tags and RC tags remain immutable.

## Detecting an upstream release

Fetch metadata first:

```bash
git fetch upstream --tags --prune
git tag --list 'v*' --sort=-version:refname
```

Consider only tags matching `vMAJOR.MINOR.PATCH` as stable candidates. Tags
with a prerelease suffix, `upstream/main`, and moving development heads are not
production inputs. An Xray version change is recorded separately from the
upstream application tag. Security-only and documentation-only upstream
changes still use the same review boundary, but their qualification scope may
be smaller. Discovery never merges or publishes anything.

The current CatX recorded base is `MHSanaei/3x-ui v3.8.5`; the strategy dry
run assessed the newer stable `v3.9.0` separately from product integration.

## Canonical sync procedure

1. Confirm a clean worktree, current remotes, the selected exact stable tag,
   and the CatX baseline SHA. Enable `rerere` only as a conflict-memory aid;
   every resolution still needs review.
2. Create `sync/upstream-vX.Y.Z` from the recorded CatX integration baseline.
3. Merge the exact tag with a merge commit. Do not rebase long-lived CatX
   history and do not merge upstream directly into `main`.
4. Classify every conflict and every sensitive-path delta below. Inspect
   conflict-free changes too; a clean merge is not semantic proof.
5. Preserve CatX identity, updater isolation, privacy boundaries, and fixed
   integration hooks with the smallest possible downstream change.
6. Run `make verify`, `make verify-fork`, migration checks when applicable,
   frontend/generated checks, Xray/config checks, feature-off checks, CatX
   regression checks, and disposable install/upgrade smoke.
7. Review the sync branch and its evidence. Integrate to `develop` only after
   the relevant gates pass; use an RC for the high-risk cases below. A stable
   release is a later, separately authorized operation.

Useful commands:

```bash
git config rerere.enabled true
git fetch upstream --tags --prune
git switch develop
git switch -c sync/upstream-vX.Y.Z
git merge --no-ff vX.Y.Z
bash scripts/upstream-maintenance-report.sh --compare vX.Y.Z
make verify
make verify-fork
```

## Conflict classification and resolution

| Class | Resolution | Required evidence |
| --- | --- | --- |
| A. UPSTREAM-ONLY | Accept upstream behavior after review. | Relevant upstream tests and regression scope. |
| B. CATX-HOOK | Preserve the fixed hook at the new upstream boundary; keep fork logic in fork-owned modules. | Hook existence and feature-off tests. |
| C. SEMANTIC-CONFLICT | Compare intended behavior, choose explicitly, and document the decision. | Targeted tests and reviewer sign-off. |
| D. GENERATED | Regenerate from authoritative sources; never hand-merge generated output. | Generator/build verification and diff review. |
| E. MIGRATION/SCHEMA | Review ordering, SQLite/PostgreSQL behavior, indexes, retention, and recovery. | Fresh, upgrade, rollback, and both-engine migration evidence. |
| F. RELEASE/UPDATER | Preserve CatX identity and reject upstream-owned assets. | Release identity, checksum, updater, and rollback tests. |
| G. FRONTEND/UX | Keep the upstream shell and adapt CatX routes/navigation at descriptors or adapters. | Type/build, route, locale, and visual checks where declared. |
| H. SECURITY/PRIVACY | Preserve no-MITM and no-sensitive-body collection boundaries; stop on ambiguity. | Security/privacy review and negative tests. |

Upstream removals and renames are handled by understanding the replacement
design, adapting CatX at the smallest supported boundary, removing obsolete
fork glue, and retaining the CatX product contract where practical. A shim
must have a documented removal condition; dead upstream architecture is not
preserved indefinitely.

## Sensitive-path review

Changes in these areas require elevated review even without a Git conflict:

```text
main.go
internal/database/**
internal/web/web.go
internal/web/runtime/**
internal/web/service/{xray.go,client*.go,inbound*.go,node*.go,subscription*.go}
internal/sub/**
internal/xray/**
internal/web/controller/{api.go,client.go,inbound.go,node.go,subscription.go}
frontend/src/{routes.tsx,layouts/**,components/command-palette/**}
frontend/src/forkext/**
internal/forkext/**
internal/forkrelease/**
update.sh
install.sh
x-ui.sh
scripts/catx-update-transaction.sh
.github/workflows/{release.yml,smoke.yml,fork-verify.yml}
```

The advisory report identifies deltas in these paths and overlaps with the
known CatX touchpoint inventory. It is a review aid, not a semantic proof.

## CatX touchpoint inventory

Fork-owned code is concentrated in `internal/forkext/**`,
`internal/forkrecovery/**`, `internal/forkrelease/**`, `frontend/src/forkext/**`,
and the additive `scripts/` and workflow verification assets. The current
upstream-touch inventory is:

| Boundary | Current touchpoint | Rule |
| --- | --- | --- |
| Bootstrap/lifecycle | `main.go`; `internal/web/web.go` | Keep fixed install/start/stop calls; no feature logic in composition roots. |
| Database/migrations | `internal/database/db.go` | One fork registration boundary; migrations remain fork-owned and ordered. |
| Runtime/node dispatch | `internal/web/runtime/**`, node services | Reuse upstream local/remote dispatch; preserve capability degradation. |
| Protected APIs | `internal/web/controller/api.go` and `frontend/src/forkext/registry.ts` | Mount under the existing authenticated boundary and shared descriptors. |
| Xray compiler/apply | `internal/web/service/xray.go`, `internal/xray/**` | Decorate the completed upstream config; preserve order and rollback. |
| Subscriptions | `main.go`, `internal/sub/**`, subscription service | Extend existing providers; do not create a parallel subscription engine. |
| Frontend routes/navigation | `frontend/src/routes.tsx`, `AppSidebar.tsx`, command palette, endpoint catalog | Use fork-owned descriptors/adapters and preserve the upstream shell. |
| Release/updater identity | `internal/forkrelease/**`, panel updater, `install.sh`, `update.sh`, `x-ui.sh` | CatX-owned sources only; never self-update into upstream. |
| Verification | `Makefile`, `.github/workflows/fork-verify.yml`, `scripts/*` tests | Keep `make verify` unchanged and additive fork checks. |

The implementation remains authoritative. This inventory is intentionally
review-oriented; the report script derives the changed-file evidence and can
flag drift in the named boundaries.

## Feature-off compatibility

With all CatX feature flags disabled, generated Xray configuration, clients,
inbounds, subscriptions, panel lifecycle, node operations, routing, and
upstream API behavior must remain as close as possible to the recorded
upstream-compatible baseline. The proof includes ordered/config fixture tests,
no fork route/job/subscriber registration, SQLite and PostgreSQL migration
checks, and release/update tests that reject official-upstream assets.

This compatibility exception is explicit: CatX release identity and updater
source remain CatX-owned even when fork product features are disabled.

## Database and recovery policy

Any upstream model or migration delta is reviewed for SQLite and PostgreSQL,
indexes, retention, migration order, fork-owned state preservation, and
upgrade from the supported previous CatX release. No runtime schema mutation
is allowed. For Xray, routing, updater, node, traffic, or DNS-affecting
changes use:

```text
snapshot → validate → apply → healthcheck → commit known-good state
FAIL → restore previous state → restart/reload → healthcheck → audit failure
```

If restoration or healthcheck cannot be proven, stop the sync and do not
integrate it.

## Version and RC policy

Do not bump versions automatically. A sync is a PATCH when behavior and
compatibility remain unchanged and only maintenance/security/documentation
corrections are included; a MINOR release is required for backward-compatible
CatX behavior or user-visible capability; a MAJOR release is reserved for
breaking API, configuration, migration, updater, or product-contract changes.
An upstream sync that changes migrations, updater, Xray compiler/config,
Traffic Control, node runtime/protocol, subscriptions, lifecycle/restart, or
security-sensitive behavior requires a public RC. Documentation-only and
strictly non-runtime upstream changes need no unnecessary RC cycle.

## Stop conditions

Stop and report a blocker when the exact tag cannot be identified, remotes do
not match this model, a sensitive boundary lacks a test or rollback path,
fork-owned state may be lost, CatX identity could be overwritten, privacy
behavior is ambiguous, or the change grows into an upstream refactor. Never
auto-merge, auto-resolve conflicts, auto-publish, deploy, or create a CatX
release as part of discovery or strategy work.
