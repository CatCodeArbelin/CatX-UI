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
