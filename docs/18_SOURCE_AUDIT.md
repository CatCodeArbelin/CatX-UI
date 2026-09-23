# Source / Requirement Audit

## Baseline checked at requirement freeze

```text
Repository: MHSanaei/3x-ui
Stable baseline: v3.8.5
Release date: 2026-09-16
```

Recent upstream commits existed after the stable tag.

Always re-check before coding.

## Important upstream facts

- backend uses Go/Gin/GORM;
- frontend uses React/Ant Design;
- SQLite and PostgreSQL are supported;
- Xray is managed by the panel;
- online client-IP state uses Xray stats rather than access.log fallback;
- official updater is upstream-specific and must be replaced/parameterized for a fork;
- Xray routing supports user/client selectors, so per-client routing should reuse existing mechanisms.

## Community requests incorporated

```text
#5900  Client Group Policies
#6534  Shared group quota
#6353  Traffic history by date range
#5649  Per-inbound traffic per client
#6622  Post-expiry throttle / rolling quota / per-client speed
#6618  Different Hosts by client/group
#6422  Self-service HWID
#6425  Outbound-change notification
#5948  Remote Xray version on subnodes
#6584  Logging controls
#5694  Group subscription link
#6322  Named routing profiles
#5889  Richer Clash/Mihomo profiles
```

A community issue is a demand signal, not an automatic implementation requirement.

## Explicitly rejected

### TLS MITM

Rejected as a product direction.

### Remnawave architecture

Rejected because it changes the 3x-ui product model too much.

## Decision rule

A feature belongs only if it is:

```text
useful
AND compatible with 3x-ui architecture
AND maintainable downstream
AND not already solved upstream
AND testable/rollback-safe
```
