# Fork Identity and Updater

## Problem

Official upstream updater is designed for official 3x-ui releases.

The fork must never update itself back to an official upstream binary.

## Version model

Keep separate:

```text
Fork version
Upstream base
Upstream latest
Xray version
```

## Release source

Use a fork-owned release repository/source.

Do not rely on hidden DB edits to override upstream updater behavior.

The authoritative CatX-UI source is `CatCodeArbelin/CatX-UI`. Release assets
use the `catx-ui-*` prefix. Go reads the identity and version metadata from
`internal/forkrelease`; standalone shell entry points mirror the identity and
`scripts/test-release-identity.sh` rejects drift or official-upstream URLs.

## Channels

```text
stable
dev
```

Optional later:

```text
rc
```

## Release metadata

Publish:

- fork version;
- upstream base;
- checksums;
- changelog;
- migration notes;
- supported architectures;
- rollback notes.

## Update flow

```text
check
→ download
→ verify integrity
→ stage
→ backup
→ install
→ migrate
→ start
→ healthcheck
→ commit
```

Failure:

```text
restore binary/config
→ recover DB as defined
→ start
→ healthcheck
→ audit
```

The current implementation stages below the installation parent, verifies the
candidate's embedded CatX-UI identity, stops the existing service, and snapshots
the installation, CLI, service unit, environment files, and database. SQLite
state is copied while the service is stopped. PostgreSQL uses `pg_dump` and is
restored with `pg_restore --single-transaction`.

The candidate is then activated, migrated, started, and required to pass two
consecutive service-health probes. Any install, migration, start, healthcheck,
or interruption failure restores the known-good files and database, starts the
old service, healthchecks it, and appends a mode-0600 audit record. A failed
rollback healthcheck is reported distinctly and leaves the recovery snapshot
in place for diagnosis.

## Xray updater

Panel updater and Xray updater remain independent.

Do not force "latest Xray" automatically.

## Supply chain

Prefer:

- GitHub Releases;
- checksums;
- optional signatures;
- pinned download domains;
- no arbitrary internal `curl | sh`.

## Tests

Use fake release provider/server.

Unit tests must not depend on live GitHub.
