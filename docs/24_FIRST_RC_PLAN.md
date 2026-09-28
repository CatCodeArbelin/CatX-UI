# Proposed first CatX-UI RC plan

This is a plan only. It deliberately creates no tag or GitHub release.

## Proposed identity

- version: `0.1.0`
- tag: `v0.1.0`
- title: `CatX-UI v0.1.0 — First Release Candidate`
- channel: stable-compatible release identity

The current updater accepts stable tags only in `vMAJOR.MINOR.PATCH` form and
the checked-in fork version is `0.1.0`. A tag such as `v0.1.0-rc.1` would
currently be rejected by stable-channel parsing. Changing that convention is
an owner decision and must be implemented and tested before use.

## Planned release body

Use the verified contents of `docs/23_RC2_RELEASE_NOTES.md`, the exact
upstream base version, migration notes, rollback instructions, known
limitations, and the final CI/staging evidence. Do not copy an unverified
checklist into the public release.

## Artifact list

Publish only assets emitted by the release workflow:

- `catx-ui-linux-amd64.tar.gz`
- `catx-ui-linux-arm64.tar.gz`
- `catx-ui-linux-armv7.tar.gz`
- `catx-ui-linux-armv6.tar.gz`
- `catx-ui-linux-386.tar.gz`
- `catx-ui-linux-armv5.tar.gz`
- `catx-ui-linux-s390x.tar.gz`
- `catx-ui-windows-amd64.zip`
- the CatX installer/updater control assets and their checksum sidecars

The final publication step must include the checksum list generated from the
exact release assets and must be preceded by the release identity and artifact
qualification jobs.

## Upgrade and rollback text

Tell operators to back up the database and installation directory, use the
CatX installer/update path, wait for panel and Xray health checks, and retain
the update evidence (`runId`, state, exit code, finish time, rollback flags).
If rollback is unhealthy, stop further retries and use the recorded recovery
state rather than manually replacing files on a production node.
