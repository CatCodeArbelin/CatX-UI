# CURRENT TASK

## Work Package

`WP-5B — Shared Group Quota / Accounting Extensions`

## Status

WP-5A is DONE and merged into `develop`. WP-5B is authorized for full
implementation on the current feature branch. Do not merge into `develop`, do
not start WP-6A, and do not modify `main`.

## Scope

The package is shared group quota and accounting extensions using the approved
accounting design below.

Required scope after the stopping point: shared group quota, atomic accounting,
and optional fixed-point traffic multiplier, reusing upstream counters.

## Required workflow

- Implement WP-5B on the current focused feature branch.
- Run all relevant verification and migration/concurrency coverage.
- Create focused commits and push the feature branch.
- Monitor canonical CI, inspect exact failures, and fix genuine defects until
  both required workflows are green.
- Do not merge into `develop`, start WP-6A, touch `main`, or modify preview/demo
  data.

## Authorized documents

- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/06_DATABASE_MIGRATIONS.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/10_POLICY_ENGINE.md`

## Authorized skills

- `skills/database-migration/SKILL.md`
- `skills/feature-implementer/SKILL.md`
- `skills/security-privacy-review/SKILL.md`

## Hard constraints

- reuse existing traffic accounting and client/group models;
- no duplicate counters, parallel accounting store, or unrelated refactor;
- upstream `client_traffics` remain the only authoritative traffic counters;
- use per-membership bases and group carry for reset/removal/rebaseline safety;
- use fixed-point active/pending multipliers; changes apply only at group
  reset/rebaseline boundaries;
- persist quota-disable ownership and reconcile only group-owned disables;
- while a group remains depleted, lifecycle/manual/API enable operations must
  not bypass the group quota; ownership remains until reset/recovery;
- lock deterministically and keep runtime/network/Xray/node calls outside DB
  transactions;
- WP-5A history is read-only and never participates in quota accounting;
- no sensitive payloads, cookies, credentials, or decrypted message storage;
- no runtime schema mutation outside migrations;
- no changes to `main` and no preview/demo data changes;
- keep the working tree clean at handoff.

## Approved accounting model

For each current group membership:

```text
member_up   = max(client.up   - membership.base_up,   0)
member_down = max(client.down - membership.base_down, 0)
raw_up      = group.carry_up   + sum(member_up)
raw_down    = group.carry_down + sum(member_down)
charged     = ceil(raw_up * active_multiplier_ppm / 1_000_000)
            + ceil(raw_down * active_multiplier_ppm / 1_000_000)
```

There are no period-base fields. Client reset/removal/deletion and real counter
decrease/rebaseline first transfer consumed deltas into group carry, then set
membership bases to the post-operation counter values. Group reset/recovery
zeros carry, rebases all current memberships, and activates the pending
multiplier.

Group depletion disables all currently enabled affected members using existing
disable semantics. Only transitions caused by this group quota are marked as
group-owned; any other disable/enable reason clears that ownership. Reset or
recovery re-enables only current members still owned by this group quota.
