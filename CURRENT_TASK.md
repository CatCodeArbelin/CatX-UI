# CURRENT TASK

## Work Package

`WP-6A — Shaping Core`

## Status

WP-5B is DONE and merged into `develop`. WP-6A is authorized for full
implementation on the current feature branch. Do not merge WP-6A, start WP-6B,
modify `main`, or touch preview/demo data.

## Scope

The package implements the shaping substrate, not WP-6B speed-limit policy.
Implement capability detection, the Shaper contract, Linux tc/nft/IFB state,
deterministic ownership/naming, mark allocation and collision detection, local
reconciliation, the status API, restart/drift reconciliation, rollback, and
unsupported-platform exact no-op behavior.

Do not add generic Xray user-to-marked-outbound routing in WP-6A; actual Xray
user-to-mark injection remains an explicit integration boundary for WP-6B.
Linux uses CatX-owned tc/nft state, conntrack mark propagation, and IFB for
reverse-path shaping. Never replace an incompatible or admin-owned qdisc.
Remote shaping uses the existing authenticated runtime.Remote/CatX API
architecture; no new agent and no SSH fallback. Do not add traffic counters or
a shaping-policy table.

## Required workflow

- Read the authorized documents and skills, inspect the actual current
  integration points, and implement the scoped substrate with focused commits.
- Add deterministic fake-executor coverage for apply, drift, and rollback;
  add privileged Linux/network-namespace coverage when supported.
- Run relevant tests and both canonical CI workflows; fix genuine failures.
- Do not merge WP-6A, start WP-6B, modify `main`, or touch preview/demo data.
- Keep the working tree clean at handoff.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/11_QOS_TRAFFIC_CONTROL.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Authorized skills

- `skills/feature-implementer/SKILL.md`
- `skills/security-privacy-review/SKILL.md`

## Hard constraints

- reuse existing Xray, client, inbound, and node boundaries;
- no second traffic/accounting engine;
- no sensitive payloads, cookies, credentials, or decrypted message storage;
- no runtime schema mutation outside migrations;
- no changes to `main` and no preview/demo data changes;
- keep the working tree clean at handoff.

## Stop condition

Stop on the green pushed WP-6A feature branch. Do not merge WP-6A or start
WP-6B.
