# CURRENT TASK

## Work Package

`RC-4 — Second Public RC Qualification & Cut`

## Completed predecessor

WP-9C was completed at implementation SHA
`954a1536a3c018eec48eb9b7a6146f8e1328fae0` and Fork Verification
`36510971071` (`success`). It was merged into `develop` with merge SHA
`8435466fe004f2488ad0be241fab7d898da84147`. The RC-4 branch starts from
that exact post-merge `develop` state. `main` must remain untouched.

## Branch and release target

Work only on `feature/rc4-second-public-rc`.

- stable version: `0.1.0` / `v0.1.0`;
- RC version: `0.1.0-rc.2` / `v0.1.0-rc.2`;
- dev channel: `dev-latest`;
- no stable release and no production update.

## Scope

Audit and qualify the existing release, installer, updater, and channel
semantics. Keep the implementation narrow and fork-owned. Do not add product
features, rewrite upstream services, modify the database schema, or touch
`main`.

Required work:

- prove the old `rc.1` binary is pinned to `v0.1.0-rc.1` and document the
  explicit verified `rc.1 → rc.2` path without claiming automatic discovery;
- make RC discovery reusable for `rc.2` and future approved CatX RC tags while
  preserving stable `/releases/latest`, unchanged `dev-latest`, trusted
  repository/asset identity, draft and malformed-tag rejection, SemVer order,
  checksum validation, no downgrade, and no stable→RC cross-grade;
- generalize release workflow qualification to `feature/rc*-*`; branch builds
  must never publish releases; retain Linux and PostgreSQL staging gates;
- update the mirrored release identity and tests to `0.1.0-rc.2`;
- add `docs/27_SECOND_RC_PLAN.md` with scope, base/Xray versions, migration,
  recovery, limitations, channel semantics, and upgrade instructions.

## Mandatory qualification before the public tag

Qualify one exact final SHA with:

- `make verify` and `make verify-fork`;
- frontend lint/typecheck/tests, i18n contracts, RTL/accessibility checks,
  generated artifacts, release identity, artifact inspection;
- Linux and PostgreSQL staging;
- Linux amd64/arm64/armv7/armv6/386/armv5/s390x and Windows amd64 assets;
- clean install, supported upgrades, `rc.1 → rc.2`, rollback, checksum and
  identity failure cases in disposable environments.

Do not create or push `v0.1.0-rc.2` until every applicable automated gate is
green on the exact final SHA and the working tree is clean. Push only that tag,
wait for the tagged workflow, verify the published release and assets, then
run the post-publication disposable smoke test. Preserve all run IDs and
evidence before deleting the RC-4 branch or merging it according to the
repository release convention.

## Authorized documents

- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/22_RC2_STAGING_RELEASE_QUALIFICATION.md`
- `docs/23_RC2_RELEASE_NOTES.md`
- `docs/24_FIRST_RC_PLAN.md`
- `docs/27_SECOND_RC_PLAN.md`
