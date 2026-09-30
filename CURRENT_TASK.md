# CURRENT TASK

## Work Package

`RC-6 — Real RC Observation & Stable Readiness`

## Base and branch

RC-5 was merged into `develop` with merge SHA
`a0cd0499d1423025ae2b8085fa3fd15db71380c3`. Work on branch
`feature/rc6-rc2-observation-stable-readiness`, created from that merge commit.
`main` remains untouched.

The immutable public RC tag `v0.1.0-rc.2` remains exactly at
`4d8feae2e62db914d3146504340d9b9f802088b2`.

## Required reading

For RC-6, read only these canonical documents in addition to `AGENTS.md` and
this file:

- `docs/03_ROADMAP.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/27_SECOND_RC_PLAN.md`

## Scope

Observe the already-published `v0.1.0-rc.2` on an isolated disposable or test
Linux host, establish the feature-off upstream baseline first, then exercise
representative CatX capabilities incrementally. If observation confirms a
source blocker, this same RC-6 package carries only the focused fix, hosted
candidate proof, exact-SHA qualification, `v0.1.0-rc.3` publication, and
post-publication observation. There is no separate RC-7 work package.

This package must not:

- add new product features;
- publish stable `v0.1.0`;
- touch `main`;
- start repository productization;
- silently place unobserved source changes into a stable release.

The Docker release path is part of RC-6 stable-readiness isolation. It must
not publish to upstream-owned namespaces and must preserve multi-architecture
behavior where practical.

## Observation decision

If a source-level stable blocker is found, complete the focused fix and
fully qualify/publish `v0.1.0-rc.3`, then repeat real observation with the
actual RC-3 assets. When that observation is green, close RC-6 as ready to
start Stable Qualification and stop. Stable `v0.1.0` is outside this package.
