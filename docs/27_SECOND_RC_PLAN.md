# Second public CatX-UI RC plan

This document qualifies the second public prerelease. It does not authorize a
production update or a stable release.

RC-4 release engineering was qualified, but the public `v0.1.0-rc.2` cut was
intentionally deferred after human review found product UX and localization
blockers. WP-9D is the blocking product-quality package. The final pre-tag
qualification and cut is planned as RC-5; its exact final post-WP-9D/RC-5 SHA,
not the historical RC-4 SHA, must pass every pre-tag gate.

WP-9D introduced a narrow protected CatX feature-settings control surface. It
did not change database schema, Xray compiler semantics, stable-channel
semantics, or the updater transaction model.

## Identity and base

- target tag: `v0.1.0-rc.2`;
- stable base: `v0.1.0`;
- fork version: `0.1.0`;
- exact upstream base: `MHSanaei/3x-ui v3.8.5`;
- bundled Xray: `26.9.9`;
- release channel: explicit CatX RC, with `prerelease=true` and `latest=false`.

Since `v0.1.0-rc.1`, this RC contains the completed CatX frontend polish:
unified styling, all 13 locale bundles, RTL foundations and fixes,
responsive/mobile corrections, accessibility hardening, and confirmation
safeguards. There is no new backend feature scope. Translation review does not
constitute native-speaker certification.

## Migration and recovery

The release uses the existing SQLite/PostgreSQL migration path and preserves
the database during updates. It does not introduce a schema migration in RC-4.
The update transaction remains:

```text
snapshot → validate → stage → backup → install → migrate → start
→ panel/Xray healthcheck → commit
```

If activation or health checks fail, the updater restores the previous binary,
service files, configuration, and database snapshot, restarts as needed, and
checks panel/Xray health. An unhealthy rollback is reported as a distinct
failure; it is not an ordinary successful update. All testing is disposable and
must retain evidence of the restore result.

## RC discovery and the rc.1 transition

The already-published `v0.1.0-rc.1` binary resolves the exact checked-in
`v0.1.0-rc.1` tag. It therefore does not automatically discover `rc.2`, and
`rc.1` is not retagged or replaced.

RC-4 adds a narrow discovery improvement for binaries built with the updated
RC identity: they inspect only the CatX release collection, reject drafts,
non-RC releases, malformed or unapproved RC tags, foreign repository metadata,
and older SemVer candidates, then choose the highest approved RC at or above
the running RC. Stable still uses only `/releases/latest`; `dev-latest` is
unchanged. Checksum and candidate identity verification remain mandatory.

The supported one-time `rc.1 → rc.2` transition is explicit and verified:

1. Back up the database and `/usr/local/x-ui` configuration.
2. Download `catx-ui-install.sh` and its matching `.sha256` sidecar from the
   `v0.1.0-rc.2` CatX release and verify the sidecar before execution.
3. Run that verified RC-2 installer with the explicit `0.1.0-rc.2` argument on
   the disposable RC-1 installation. The RC-2 installer downloads the matching
   RC-2 updater, verifies its checksum and candidate identity, and performs the
   transactional update.
4. Confirm panel and Xray health, release identity, and database preservation.

This is a manual operator transition, not automatic channel discovery by the
old RC-1 binary. No RC path silently downgrades or moves a stable installation
onto the RC channel.

## Channel semantics

SemVer ordering is `0.1.0-rc.1 < 0.1.0-rc.2 < 0.1.0`.

- stable clients resolve the stable-only latest pointer and reject prereleases;
- RC clients resolve approved CatX RC metadata and never use the stable latest
  pointer as an RC target;
- dev clients remain opt-in and resolve the rolling `dev-latest` prerelease;
- RC publication is prerelease and never the GitHub latest release;
- the stable `v0.1.0` pointer is not moved by this RC.

## Known limitations

- `v0.1.0-rc.1` remains pinned by design and needs the explicit verified
  transition above;
- Xray remains an external runtime and optional capabilities may be reported as
  unsupported;
- PostgreSQL, privileged Linux enforcement, and long-duration soak coverage
  depend on the canonical CI runner;
- no production node is part of RC qualification.

## Pre-tag and post-publication gate

The exact final post-WP-9D/RC-5 SHA must pass the full Go/frontend/i18n/accessibility,
artifact, Linux, PostgreSQL, upgrade, checksum, identity, and rollback matrix
before the annotated tag is created. After publication, use the actual RC-2
assets for a disposable clean install, the documented RC-1 transition, panel
and Xray startup, Russian and one RTL locale, and a rollback exercise.
