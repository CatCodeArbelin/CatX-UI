# CatX-UI v0.2.0 Release Notes

Status: `FINAL — v0.2.0 published and qualified`

These notes describe the integrated CatX v0.2.0 behavior that is being
qualified. They do not replace the immutable historical v0.1.0 notes.

## Foundation

- CatX-UI now uses the upstream `MHSanaei/3x-ui v3.9.0` baseline.
- The bundled Xray version is `26.9.30`.
- CatX release identity, updater, installer, checksums, rollback, and release
  assets remain owned by `CatCodeArbelin/CatX-UI`.
- SQLite and PostgreSQL remain supported, with the upstream 3.9.0 migration
  behavior integrated alongside CatX state preservation and recovery checks.

## Upstream-integrated behavior

The release qualification covers the upstream changes present in the exact
integrated source, including:

- client and inbound lifecycle/reset improvements;
- renewal and reset behavior for client traffic and related state;
- the native in-process TUIC implementation and its accounting/lifecycle;
- subscription generation and protocol-specific subscription updates;
- AmneziaWG, WireGuard, endpoint, and related network behavior;
- database, migration, import, and runtime recovery changes from upstream
  v3.9.0.

The public RC qualification must validate these paths with the actual emitted
release artifacts rather than only with CI build outputs.

## CatX capabilities retained

The integrated release retains the existing CatX Policy, Insight, Traffic
Control, and Operations surfaces, including Activity, DNS Intelligence, Policy
and explain/simulation flows, Traffic History, Group Quota, Risk, Audit,
Webhooks, Portal, Fleet, Fleet Update, backup/restore, and transactional
updater behavior. Qualification records whether each surface is active,
degraded, or intentionally disabled rather than treating UI presence as proof
of runtime support.

## Limitations

- Generic Xray-user attribution is unsupported unless a specific upstream or
  Xray capability is explicitly proven for that path.
- Generic per-user kernel shaping is capability-dependent. Traffic Control
  must expose unsupported or degraded capability state and must not advertise
  universal per-user enforcement.
- DNS and destination intelligence uses observable metadata only. It does not
  decrypt HTTPS, collect request bodies, cookies, Authorization headers,
  credentials, passwords, or messages.
- DNS policy cannot promise complete control of every DoH or other encrypted
  resolver path.
- Classification, enrichment, risk, and sharing signals remain evidence or
  heuristics; a single IP or heuristic signal is not proof of account sharing
  and does not auto-ban a user.

## Qualification status

The public `v0.2.0-rc.1` is the first planned and only preauthorized RC. A
second RC was not required: public RC observation was green, and stable reused
the same proven product/runtime source without a stable-only product patch.

Stable `v0.2.0` is published from final main SHA
`2b1760e98e665bde388e94c420bd22aa182fe2b0`. The annotated immutable tag,
canonical release workflow, CatX-owned multi-arch GHCR image, and actual-public
stable observer all passed. The public observer also passed the real
`v0.1.0 → v0.2.0` upgrade/rollback path and enabled/disabled `restartPanel`
lifecycle checks.
