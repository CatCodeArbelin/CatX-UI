# RC-2 Staging Qualification & Release Engineering

RC-2 is a disposable-environment qualification package. It does not authorize
production access, a `main` update, a release tag, or a GitHub release.

## Canonical checks

- `make verify`
- `make verify-fork`
- release workflow artifact qualification on the exact commit
- Linux install/update rehearsal using the release workflow artifacts
- SQLite and PostgreSQL migration coverage where the canonical environment is
  available
- real node `StartUpdate`/`GetUpdateStatus` HTTP boundary coverage
- feature-off compatibility and secret-leak checks

## Release asset matrix

The release workflow emits seven Linux archives (`amd64`, `arm64`, `armv7`,
`armv6`, `386`, `armv5`, `s390x`), one Windows archive (`amd64`), and the
CatX control assets (`update`, `install`, menu, updater library, metadata, and
changelog), each with a matching SHA-256 sidecar. The matrix is intentionally
validated from the workflow output rather than inferred from a local build.

## Staging topology

The Linux rehearsal uses disposable Ubuntu 24.04 containers, a local synthetic
release server, one panel process, and two synthetic legacy panel binaries:

- production baseline: `7ef22f94c950ff09f0870e2295fa65ad5968742c`;
- pre-RC develop/WP-8C generation: `1c88ed18`.

The database fixture is generated inside the disposable environment and uses
only synthetic settings and a synthetic marker row. No real addresses,
credentials, tokens, or production database files are used.

## Failure policy

An artifact with a modified archive/checksum, missing sidecar, wrong asset name,
foreign repository identity, or foreign release URL must fail closed. Update
status is accepted only when its string `runId` matches the originating run.
An ambiguous POST result is retained as unknown until node evidence resolves
it; it is never blindly dispatched again.

The final audit must record which checks ran and which remained unavailable,
especially PostgreSQL, privileged Xray enforcement, and long-duration soak
coverage when the runner does not provide those capabilities.
