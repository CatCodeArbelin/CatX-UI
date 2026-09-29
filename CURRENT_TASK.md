# CURRENT TASK

## Work Package

`RC-5 — Final Second Public RC Qualification & Cut`

## Base and branch

WP-9D is DONE and merged into `develop` with merge SHA
`17588b4ed90798979ee2a3953ed5d1d8430afd6e`. Work on branch
`feature/rc5-final-second-public-rc`, created from that exact post-WP-9D
`develop` state. `main` must remain untouched.

## Required reading

For RC-5, read only these canonical documents in addition to `AGENTS.md` and
this file:

- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/22_RC2_STAGING_RELEASE_QUALIFICATION.md`
- `docs/27_SECOND_RC_PLAN.md`

## Scope

Re-run the complete second public RC qualification on the exact post-WP-9D
state and, only after every pre-tag gate passes, create and publish the
CatX-owned `v0.1.0-rc.2` prerelease. Preserve the RC-4 release-engineering
improvements, explicit RC channel semantics, artifact identity and checksum
verification, Linux/PostgreSQL/upgrade coverage, rollback evidence, and the
documented disposable `rc.1 → rc.2` transition.

Do not touch stable `v0.1.0`, `main`, or unrelated product areas. Do not start
repository productization. No `v0.1.0-rc.2` tag or release exists yet.

## Mandatory qualification gates

- exact-SHA `make verify` and `make verify-fork`;
- frontend, Go, i18n, accessibility, generated, artifact, Linux, PostgreSQL,
  upgrade, checksum, release-identity, and rollback coverage;
- disposable clean-install and `v0.1.0-rc.1 → v0.1.0-rc.2` transition smoke;
- panel/Xray health, database preservation, Russian and one RTL locale checks;
- evidence of successful restore after rollback exercise.

Do not create, tag, or publish `v0.1.0-rc.2` until the full pre-tag matrix is
green. Do not begin stable qualification in this package.
