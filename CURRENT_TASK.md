# CURRENT TASK

## Work Package

`Repository Productization — COMPLETE`

## Base and branch

This package starts from the published stable product/runtime SHA
`fd28ea7144147d9164b70810d4a24872a3d48b4f` (`main` and tag `v0.1.0`) on
branch `feature/repository-productization`.

The final Stable Qualification evidence commit
`2ec8d6c3661dd37c81cc057066781217a5f26faf` was inspected and confirmed to be
documentation-only (`CURRENT_TASK.md`, `TASK_QUEUE.md`). Its changes were
carried onto this branch as commit `05479906` without importing unrelated
post-release qualification harness work.

Immutable public tags remain unchanged:

- `v0.1.0-rc.1` — `09f5432aacfb2c45f3df79cdca2e15e77bc9a8a6`;
- `v0.1.0-rc.2` — `4d8feae2e62db914d3146504340d9b9f802088b2`;
- `v0.1.0-rc.3` — `437d5e2f5bb835118ba6628ea62059b27721cfa3`;
- `v0.1.0-rc.4` — `283abaa401d60b4374a4b618a5a5f8a72ec26cb4`;
- `v0.1.0` — `fd28ea7144147d9164b70810d4a24872a3d48b4f`.

## Required reading

Before changing implementation or public repository state, read:

- `AGENTS.md`;
- `CURRENT_TASK.md` and `TASK_QUEUE.md`;
- `docs/00_MASTER_SPEC.md`;
- `docs/01_ARCHITECTURE.md`;
- `docs/02_FEATURE_CATALOG.md`;
- `docs/03_ROADMAP.md`;
- `docs/04_UPSTREAM_SYNC.md`;
- `docs/05_TESTING_ROLLBACK_RELEASE.md`;
- `docs/08_FRONTEND_UX.md`;
- `docs/12_UPDATER_IDENTITY.md`;
- `docs/14_NON_GOALS.md`;
- `docs/15_DEFINITION_OF_DONE.md`;
- `docs/16_LEGAL_GPL_NOTES.md`;
- `docs/17_AI_WORKFLOW.md`;
- `docs/19_REPOSITORY_MAP.md`;
- `docs/23_RC2_RELEASE_NOTES.md`;
- `docs/28_RC6_REAL_RC_OBSERVATION.md`.

## Scope

Productize the CatX-UI repository and its public-facing documentation while
preserving the released v0.1.0 runtime. This is not a new application-feature
phase. Do not begin Upstream Maintenance Strategy, sync a newer upstream
release, create v0.1.1, or change runtime behavior.

Allowed changes are limited to repository identity, documentation, release
notes, public issue/contribution/security routing, documentation-site identity,
public assets/metadata, and corrections to public install/update/Docker
examples. Compatibility-sensitive internal identifiers remain unchanged,
including the Go module/import path, package paths, database schema names,
CLI/API routes, and Xray boundaries.

Public claims must match stable v0.1.0. In particular, generic Xray-user
kernel attribution and universal per-user kernel shaping are unsupported;
Traffic Control must use capability-dependent supported/degraded/unsupported
language; CatX does not use TLS MITM or collect decrypted bodies, cookies,
Authorization headers, credentials, passwords, or messages; and CatX is not a
billing platform or generic plugin ecosystem.

## Required productization outcomes

1. Replace the upstream-oriented root README with a coherent CatX README that
   identifies the downstream relationship, release/update sources, supported
   platforms/databases, Docker image, architecture, caveats, privacy boundary,
   rollback expectations, contribution routes, and GPL attribution.
2. Audit and update linked README translations without claiming that CatX is
   the official upstream project.
3. Retarget CatX install/update/Docker examples to `CatCodeArbelin/CatX-UI`
   and `ghcr.io/catcodearbelin/catx-ui`, while keeping legitimate upstream
   attribution and technical references.
4. Productize `CONTRIBUTING.md`, `SECURITY.md`, issue/PR routing, cloud-init
   notes, and local docs-site identity.
5. Add stable v0.1.0 release notes covering the shipped capabilities,
   limitations, upgrade, and rollback expectations.
6. Perform a stale public-link and product-reality audit, classify remaining
   upstream references, and document any repository metadata that cannot be
   changed with available permissions.

## Verification gate

Run checks appropriate to the final diff and report exact results. At minimum:

- Markdown/link and image/asset-reference sanity;
- stale CatX install/update/Docker-link audit;
- docs build/format checks when docs-site files change;
- `make verify` and `make verify-fork` if required by the final diff or
  repository governance.

Do not claim checks that were not run. Inspect the final diff for runtime/source
drift and confirm the stable and RC tag targets remain unchanged.

## Stop condition

After Repository Productization is complete, stop. Do not start Upstream
Maintenance Strategy or any new product feature work in this package.

## Completion record

The prerequisite frontend dependency remediation was isolated in security PR
#2 and merged first:

- dependency fix commit: `960138303cc5b8e2bfd84471dfb0b43eec9d53e7`;
- dependency-fix merge commit: `e046c282fc8ea770dfe35c0d2b00d31341d28f4c`;
- Productization branch prerequisite merge commit: `606d31e5de1dcc93f025a7d86487d21119dacf62`;
- final Productization head before merge: `e39b3dbf6c780518cb77a209e6d968875fec2680`;
- Productization PR #1 merge commit and final `main`: `db610e986862bbe902d0fc8f1e4aa52cd707095f`.

The remediation is lockfile-only and changes no runtime source. Final hosted
evidence: Docs CI `36911212458`, CI `36911212478`, Fork Verification
`36911212331`, Release CatX-UI `36911212551`, and Deploy Smoke
`36911212337` all passed. No v0.1.1 was created, no upstream release was
synchronized, and the stable/RC tags remain immutable.
