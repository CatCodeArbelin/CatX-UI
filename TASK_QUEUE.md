# TASK QUEUE — WORK PACKAGE MODE

The agent executes one **work package** at a time, not one tiny task per branch.

A work package may contain several checkpoints/commits, but all work must stay inside the package scope.

The agent must stop before entering the next work package.

---

# PHASE 0 — Fork Safety Foundation

## WP-0A — Foundation Guardrails — DONE

Recommended model: `Luna Medium`

Branch:

```text
feature/wp-0a-foundation-guardrails
```

Includes the former tasks:

```text
TASK-002 Baseline Compatibility Suite
TASK-003 Fixed Fork Hook Contracts & Bootstrap
TASK-004 make verify-fork + CI Gate
TASK-007 Fork Settings Facade & Feature Flags
TASK-010 Empty API / OpenAPI / Frontend Registries
```

### Goal

Build the low-risk structural foundation required before product features.

### Checkpoints

#### CP-0A.1 — Baseline Compatibility

- map/reuse existing upstream tests;
- add only missing compatibility tests/fixtures;
- document coverage in `docs/20_BASELINE_COMPATIBILITY.md`;
- no product behavior changes.

#### CP-0A.2 — Fixed No-op Fork Hooks

- add the smallest verified first-party hook boundaries;
- no generic plugin framework;
- all hooks default to exact no-ops;
- baseline behavior must remain unchanged.

#### CP-0A.3 — verify-fork + CI

- add `make verify-fork`;
- keep upstream `make verify` unchanged;
- add fork-specific CI gate.

#### CP-0A.4 — Fork Settings / Feature Flags

Initial flags:

```text
analytics.enabled
dns_intelligence.enabled
policies.enabled
traffic_control.enabled
security_anomaly.enabled
```

All default OFF.

#### CP-0A.5 — Empty API / OpenAPI / Frontend Registries

Add only minimal registries/descriptors for future fork features.

No product feature pages yet.

### Merge gate

WP-0A may merge only when the available test environment can verify the package.

If local Windows cannot run required tests, use an approved alternative such as WSL/CI/container-based verification. Do not claim green tests when they were not executed.

---

## WP-0B — Release Identity & Updater Safety — DONE

Recommended model: `Sol High`

Branch:

```text
feature/wp-0b-release-safety
```

Includes:

```text
TASK-005 CatX-UI Identity & Updater Isolation
TASK-006 Transactional Updater Staging & Automatic Rollback
```

### Goal

Ensure CatX-UI cannot be overwritten by official upstream releases and failed updates restore a known-good installation.

### Checkpoints

1. CatX-UI release identity.
2. Fork/upstream/Xray version separation.
3. Retarget all updater/install paths.
4. Integrity validation.
5. Staged update.
6. Healthcheck.
7. Automatic rollback.
8. Update/rollback tests.

---

## WP-0C — Xray / Database Recovery Safety — DONE

Recommended model: `Sol High`

Branch:

```text
feature/wp-0c-runtime-recovery
```

Includes:

```text
TASK-008 Xray Candidate Validation & Known-Good Rollback
TASK-009 DB Import / Xray Start Recovery Gap
```

### Goal

Make runtime-affecting changes recoverable before Policy Engine or traffic-control features exist.

---

# PHASE 1 — Analytics

## WP-1A — Analytics Data Foundation — DONE

Recommended model: `Luna Medium`

Includes:

- analytics persistence schema;
- destination event model;
- Xray online/traffic adapter;
- historical aggregation skeleton;
- retention foundation.

Reuse upstream clients/groups/nodes/traffic counters.

Do not create parallel upstream models.

---

## WP-1B — access.log & Session Pipeline — DONE

Recommended model: `Sol High`

Includes:

- incremental access.log collector;
- rotation/truncate/restart handling;
- bounded batching;
- parser fuzz/race tests;
- session correlation.

---

## WP-1C — Activity API/UI — DONE

Recommended model: `Luna Medium`

Includes:

- client activity API;
- client activity UI;
- basic historical views;
- source/confidence display.

---

# PHASE 2 — DNS / Destination Intelligence

## WP-2A — DNS & Evidence Fusion — DONE

Recommended model: `Sol High`

Includes:

- DNS observer;
- destination evidence fusion;
- SNI/TLS/QUIC metadata where observable;
- source/confidence semantics.

No TLS MITM.

---

## WP-2B — Enrichment & Classification — DONE

Recommended model: `Luna Medium`

Includes:

- ASN/GeoIP;
- service recognition;
- category classifier;
- first-seen/new-domain intelligence;
- relationship graph.

Completed and merged into `develop` (`a1895d77`).

---

## WP-2C — DNS Intelligence UI / Privacy / Retention — DONE

Recommended model: `Luna Medium`

Includes:

- DNS dashboard;
- privacy dashboard;
- retention controls;
- delete-history controls.

Completed and merged into `develop`.

---

# PHASE 3 — Policy Engine v1

## WP-3A — Policy Data/API — DONE

Recommended model: `Luna Medium`

Includes:

- policy schema/repository;
- CRUD API;
- group assignment;
- client override.

Reuse upstream `ClientGroup` and normalized clients.

Completed and merged into `develop`.

---

## WP-3B — Policy Decision Engine & Xray Compiler — DONE

Recommended model: `Sol High`

Includes:

- precedence engine;
- deterministic final-config decorator;
- route/order preservation;
- policy fixtures;
- Xray validation;
- rollback integration.

Completed and merged into `develop`.

---

## WP-3C — Simulator / Explain / UI — DONE

Recommended model: `Luna Medium`

Includes:

- Policy Simulator;
- Explain Decision;
- Explain Route;
- policy UI.

Completed and merged into `develop` (`15c7e310`).

Reuse upstream route-test capabilities where possible.

---

# PHASE 4 — Policy Engine v2

## WP-4A — Categories / Schedules / Temporary Overrides — DONE

Recommended model: `Sol Medium`

Includes:

- categories;
- schedules;
- timezone/DST tests;
- temporary overrides.

Completed and merged into `develop` (`0de64062`).

---

## WP-4B — Quarantine / Managed DNS / SafeSearch — DONE

Recommended model: `Sol High`

Includes:

- quarantine;
- managed DNS policy;
- SafeSearch where supported;
- documented DoH limitations.

---

# PHASE 5 — Traffic History & Quotas

## WP-5A — Historical Traffic — DONE

Recommended model: `Luna Medium`

Includes:

- custom date ranges;
- service/category breakdown;
- reuse upstream per-client/per-inbound/per-node accounting.

---

## WP-5B — Shared Group Quota / Accounting Extensions — DONE

Recommended model: `Sol Medium`

Includes:

- shared group quota;
- atomic accounting;
- optional fixed-point traffic multiplier.

---

# PHASE 6 — QoS / Traffic Control

## WP-6A — Shaping Core — DONE

Recommended model: `Sol High`

Includes:

- capability detection;
- shaping adapter;
- Linux implementation;
- reconciliation.

---

## WP-6B — Speed / Rolling Quota / Soft Throttle — DONE

Recommended model: `Sol High`

Includes:

- per-client upload/download speed;
- rolling/window quota;
- post-quota/post-expiry throttle;
- optional category cap when classification is reliable.

---

# PHASE 7 — Security / Anomaly

## WP-7A — Risk Intelligence

Status: DONE

Recommended model: `Sol Medium`

Includes:

- extended IP/session history;
- sharing-risk engine;
- ASN/country alerts;
- DNS anomaly heuristics.

No automatic ban by default.

---

# PHASE 8 — Operations / Product Polish

## WP-8A — Audit / Webhooks / Metrics

Status: DONE

Recommended model: `Luna Medium`

Includes:

- durable fork audit UI;
- webhooks;
- Prometheus export.

Reuse upstream event bus for notifications, not as durable audit storage.

---

## WP-8B — Self-service / Host Visibility / Fleet UI

Status: DONE

Recommended model: `Luna Medium`, with security review for self-service.

Includes:

- self-service HWID/device portal;
- host visibility per group/client if upstream still lacks it;
- fleet dashboard extensions.

---

## WP-8C — Multi-node Update Orchestration

Status: DONE

Recommended model: `Sol High`

Only after single-node updater/rollback is proven.

# RELEASE CANDIDATES

## RC-1 — Upstream Sync & Release Candidate Hardening — DONE

Merged into `develop` with merge commit `ec6fbc3b`. The exact pinned upstream
commit and RC-1 verification records are retained in
`docs/21_RC1_UPSTREAM_SYNC.md`.

## RC-2 — Staging Qualification & Release Engineering — DONE

Qualify release artifacts, Linux installation/upgrades, populated migrations,
the real node-local updater, rollback, disposable fleet rollout, feature-off
compatibility, secret boundaries, resource soak, and first-RC documentation.

---

## RC-3 — Prerelease Versioning & First Public RC Cut — DONE

The first public release candidate `0.1.0-rc.1` was cut successfully at
`09f5432aacfb2c45f3df79cdca2e15e77bc9a8a6`.

## WP-9A — Frontend Design System & i18n Foundation — DONE

Branch: `feature/wp-9a-frontend-design-i18n-foundation`

Align CatX frontend surfaces with the upstream 3x-ui design system, stabilize
fork-owned i18n namespaces, establish RTL and responsive foundations, and add
contract tests. Completed on the feature branch with green Fork Verification
run `36504762871` at SHA
`d986386987a5645a2da12250595306f0d8b9578a`.

## WP-9B — Full CatX Localization — DONE

Branch: `feature/wp-9b-full-localization`

Completed at `4a02a2edfefe48fa30a81acc9e47f521f3e2c8d1` after green Fork
Verification run `36508672070`. The final translation audit confirmed 238
CatX keys in English and every supported locale, with matching key trees,
value types, placeholders, and only the documented technical-term
English-identical allowlist.

Translate the stabilized CatX `fork.*` namespace across all 12 non-English
locales while preserving the WP-9A i18n contract, behavior, RTL foundation,
and low-divergence upstream boundary. Do not merge until the package
verification gate is green.

## WP-9C — Accessibility, RTL, Responsive & Visual QA — DONE

Branch: `feature/wp-9c-visual-accessibility-qa`

Prove that completed CatX surfaces remain one coherent 3x-ui panel across
themes, supported viewport sizes, LTR/RTL locales, and long-string locales.
Fix only confirmed overflow, responsive, directionality, accessibility,
contrast, focus, labeling, confirmation, or semantic translation defects.
Do not modify backend, database, Xray, updater, release semantics, or `main`.

Completed at `954a1536a3c018eec48eb9b7a6146f8e1328fae0` after green Fork
Verification run `36510971071`; merged into `develop` with merge SHA
`8435466fe004f2488ad0be241fab7d898da84147`.

## RC-4 — Second Public RC Qualification & Cut — DONE

Branch: `feature/rc4-second-public-rc`

Qualify and cut `v0.1.0-rc.2` without touching `main` or publishing stable
`v0.1.0`. Preserve the explicit RC channel, verify the one-time `rc.1 → rc.2`
operator path, generalize future RC discovery and branch qualification, and
retain the full artifact, Linux, PostgreSQL, upgrade, rollback, and publication
evidence. After publication the next phase is release observation and stable
qualification, not feature development.

RC-4 release engineering was technically qualified at
`e780d9b87e7bb2fcb3cdeab6e33f11b893d9e68b`; Fork Verification
`36514812058` and Release CatX-UI `36514812043` were green. The public
`v0.1.0-rc.2` cut was intentionally deferred because manual smoke testing
found product UX/localization blockers. No public tag or release was created.
RC-4 merged into `develop` with `b66ac15f79d411b5a0b2386e8770b05e88254bc2`,
and its feature branch was deleted. `main` was not changed.

## WP-9D — Product UX Integration & Human QA — DONE

Branch: `feature/wp-9d-product-ux-integration`

Group CatX destinations into existing upstream sidebar submenus, use semantic
Ant Design icons, remove Sponsors from primary navigation without deleting its
implementation, improve shared page composition and empty states, make
feature-off states explicit and localized, and complete the Russian and
narrow non-Russian leakage review. The current package includes the native
themed CatX admin shell, restored upstream submenu destinations, contextual
client access, CatX feature settings, actionable feature-off UX, semantic
localization QA, and the mandatory human visual gate. Preserve direct routes.

Final human functional smoke was approved at implementation SHA
`52f70d832b5062dda495b9b6e406075b2ce3ae0a`. The apparent restart failure was
environment-only: the disposable preview used automatic container removal and
exited cleanly on SIGINT; a corrected persistent-container smoke passed with
Analytics and DNS Analytics enabled, including save, restart, flag persistence,
Activity, `/activity`, `/clients`, `/nodes`, and `/routing` HTTP 200 responses,
and no panic, fatal, or migration errors. WP-9D is merge-ready. The next
package is RC-5; no `v0.1.0-rc.2` tag or release exists.

WP-9D was merged into `develop` with merge SHA
`17588b4ed90798979ee2a3953ed5d1d8430afd6e`; its feature branch was deleted
locally and remotely.

## RC-5 — Final Second Public RC Qualification & Cut — DONE

Branch: `feature/rc5-final-second-public-rc`

RC-5 merged into `develop` with merge SHA
`a0cd0499d1423025ae2b8085fa3fd15db71380c3` after the exact-SHA Fork
Verification and Release CatX-UI gates passed. The public prerelease tag is
`v0.1.0-rc.2`, and its immutable release SHA is
`4d8feae2e62db914d3146504340d9b9f802088b2`. The canonical tagged Release
workflow `36674096974` passed. Post-publication clean install, actual
published-asset RC1-to-RC2 transition, rollback, checksum, panel/Xray, and
Russian/RTL smoke qualification passed. Stable `v0.1.0` was not published and
`main` was not changed.

## RC-6 — Real RC Observation & Stable Readiness — DONE

Branch: `feature/rc6-rc2-observation-stable-readiness`

Observe the already-published `v0.1.0-rc.2` on isolated/test infrastructure.
Establish a feature-off upstream baseline before incrementally exercising
representative CatX Analytics, DNS Intelligence, Policy, Traffic Control, and
Operations flows. If the source blocker is confirmed, carry only its focused
fix, qualify and publish `v0.1.0-rc.3`, then repeat observation with the actual
RC-3 assets. This remains one RC-6 package; do not create an RC-7 branch or
work package. Audit and isolate the Docker release path from upstream image
namespaces, do not add product features, do not publish stable `v0.1.0`, and
keep `main` untouched.

If a source-level blocker is found, finish the focused fix, exact-SHA RC-3
qualification/publication, and actual RC-3 observation within RC-6. If that
observation is green, record `RC-6 COMPLETE — READY TO START STABLE
QUALIFICATION` and stop before Stable Qualification.

## RC-6 completion record

RC-6 completed on exact SHA `437d5e2f5bb835118ba6628ea62059b27721cfa3`.
Hosted candidate qualification, Fork verification, public `v0.1.0-rc.3`
release qualification, and post-publication Linux observation passed. Stable
Qualification is the next roadmap step. No RC-7 branch or work package was
created; stable `v0.1.0` and `main` remain untouched.

RC-6 was merged into `develop` with merge SHA
`ecd8ee1afbee4996f9a099d67014c4ab2236ae44`.

## Stable Qualification — v0.1.0 — DONE

Branch: `feature/stable-v0.1.0-qualification`

Starting SHA: `ecd8ee1afbee4996f9a099d67014c4ab2236ae44`.

Qualify the already-observed RC3 code for stable without adding features,
starting Repository Productization, or touching `main` before all stable gates
pass. The first mandatory gate is a persistent disposable Linux regression
using actual public `v0.1.0-rc.3` assets. It must save
`analytics.enabled` and `dns_intelligence.enabled`, invoke the real
`POST /panel/api/setting/restartPanel` endpoint, verify recovery and Activity /
DNS Intelligence health, then disable both flags and repeat the real restart.

Hosted restart regression run `36804879634` used actual public RC3 assets. The
archive, checksums, release identity, clean Linux install, login, SQLite,
Xray, persisted feature flags, real `POST /panel/api/setting/restartPanel`,
same-process recovery, and web/sub server restart all passed. The source-level
gate failed because after the real panel restart the analytics and DNS runtime
reported disabled even though the persisted `analytics.enabled` and
`dns_intelligence.enabled` flags remained enabled. The panel-only restart path
does not re-run the fork runtime configuration performed during process
startup.

The focused generic fork-runtime lifecycle fix was published as immutable
`v0.1.0-rc.4` at exact SHA
`283abaa401d60b4374a4b618a5a5f8a72ec26cb4`. Final Fork Verification
`36807215175`, final hosted restart/Traffic Control observation `36807215104`,
the non-publishing complete RC matrix `36807225858`, and canonical tagged
Release CatX-UI workflow `36807931375` passed.

Post-publication run `36808874388` downloaded and verified the actual public
RC4 Linux archive and metadata, then passed enabled and disabled transitions
through the real `POST /panel/api/setting/restartPanel` endpoint. Analytics,
DNS Intelligence, Activity, panel, Xray, and representative generic managed
feature runtimes matched persisted state after both restarts.

Stable Qualification completed on exact product/runtime SHA
`fd28ea7144147d9164b70810d4a24872a3d48b4f`. Exact-SHA Fork Verification
`36811795614`, stable candidate lifecycle `36811795653`, and the complete
non-publishing stable release matrix `36811937162` passed. The matrix covered
Linux/Windows artifacts and identity, SQLite, PostgreSQL, clean install,
actual RC4-to-stable upgrade, database preservation, restart health, and
rollback to RC4. Reconnect, node-restart, multi-node fanout, remote capability,
Traffic Control, quota/window, feature-off, and no-clobber coverage remained
green.

`main` and annotated `v0.1.0` resolve to the qualified SHA. Canonical tagged
Release run `36813723610` passed; the public final/latest release has 28
expected assets. Corrected actual-public-assets observer `36815136071` passed
updater identity, Russian LTR, Persian RTL, enabled and disabled real
`restartPanel` lifecycles, generic managed-feature reload, panel/SQLite/Xray
health, live traffic, Traffic Control apply/remove/no-clobber, authoritative
quota/window behavior, and feature-disable cleanup.

Observer attempts `36814708161` and `36814901784` remain recorded. The first
was initially classified as latest-release visibility lag; the retry proved
the qualification assertion incorrectly expected an unprefixed latest version
instead of tag `v0.1.0`. Harness-only commit
`35a03bb9d4407ebe343d9d8c0f8c62024dbdc4bb` corrected the assertion on the
qualification branch and was not merged into the already-published stable
product.

Docker run `36813723615` passed. CatX-only aliases `v0.1.0`, `0.1.0`, and
`latest` resolve to
`sha256:3e94b98560e8261c315b9f9ae4d97a0e8ea5977e00072f9b310088eb9fac1322`
with amd64, arm64, arm/v7, arm/v6, and 386 images plus provenance/SBOM
attestations; the OCI revision is the qualified SHA.

Non-canonical follow-ups remain explicit: the post-freeze npm audit found a
high `brace-expansion` advisory and low DOMPurify advisory after all source
build/test/lint/typecheck gates passed; Docs format drift and disabled GitHub
Pages are tooling/configuration issues; and the secondary release-install
smoke's no-systemd container cannot perform transactional updater service
control, while the dedicated clean-install/upgrade/rollback matrix passed.
None established a stable binary runtime defect during qualification.

All RC tags remain immutable, public stable assets remain unchanged, no
post-qualification product/runtime source entered `main`, and Repository
Productization has not started. Stop before Productization.

## Repository Productization — COMPLETE

Branch: `feature/repository-productization`

Starting stable product/runtime SHA: `fd28ea7144147d9164b70810d4a24872a3d48b4f`.

The final Stable Qualification evidence commit
`2ec8d6c3661dd37c81cc057066781217a5f26faf` was documentation-only and was
carried onto this branch as `05479906`. Productization is limited to public
repository identity, documentation, release notes, public issue/contribution/
security routing, docs-site identity, and install/update/Docker examples. Do
not change v0.1.0 runtime semantics, compatibility-sensitive internal names,
the stable tag, or any RC tag.

Required completion checks:

- product claims match the shipped stable product and capability model;
- CatX-owned install, update, release, issue, and Docker links are used;
- required upstream attribution and GPL notices remain intact;
- translated READMEs and local docs are not presented as official upstream
  CatX documentation;
- stale public upstream-link audit, Markdown/link sanity, and relevant docs
  checks are recorded;
- final diff contains no Go or frontend runtime-source change;
- stable/RC tag targets remain unchanged and no v0.1.1 or upstream maintenance
  work starts.

Stop after this work package and wait for separate authorization for Upstream
Maintenance Strategy.

### Completion record

Security PR #2 (`fix/frontend-audit-dependencies`) was merged first at
`e046c282fc8ea770dfe35c0d2b00d31341d28f4c` from focused fix commit
`960138303cc5b8e2bfd84471dfb0b43eec9d53e7`. Its lockfile-only change cleared
the hosted frontend audit while preserving runtime source. The Productization
branch includes that prerequisite through merge commit
`606d31e5de1dcc93f025a7d86487d21119dacf62`.

Productization PR #1 final head was `e39b3dbf6c780518cb77a209e6d968875fec2680`
and merged at `db610e986862bbe902d0fc8f1e4aa52cd707095f`. Docs CI
`36911212458`, CI `36911212478`, Fork Verification `36911212331`, Release
CatX-UI `36911212551`, and Deploy Smoke `36911212337` passed. No v0.1.1,
upstream maintenance, tag movement, or runtime-source change is part of this
package.

# Upstream Maintenance Strategy — COMPLETE

Branch: `feature/upstream-maintenance-strategy`

Starting `main`: `8f63afc6f1fdbac0f50d3bfd4f6f0bb8da4255fb`.

Recorded upstream base: `MHSanaei/3x-ui v3.8.5`. The current fetched stable
tag is `v3.9.0` at `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb`.

This package establishes the canonical maintenance process in
`docs/04_UPSTREAM_SYNC.md`, inventories CatX/upstream touchpoints and sensitive
paths, defines conflict/migration/rollback/feature-off/RC policy, and adds the
advisory `scripts/upstream-maintenance-report.sh`. It must not add product
features, publish a release, move any tag, or merge upstream into `main` or
`develop`.

The `v3.9.0` assessment is a non-production dry run on
`sync/upstream-v3.9.0`. Any actual upgrade is a separate work package.

# Upstream Sync — 3x-ui v3.9.0 — COMPLETE

Branch: `sync/upstream-v3.9.0`

Integration base: `65957a8988b5b51c4ac01dbd1121f04b55af6c8a`.

Integrate exact upstream `v3.9.0` at
`3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb` from the recorded
`v3.8.5` base `7ef22f94c950ff09f0870e2295fa65ad5968742c`. Preserve upstream
ancestry, CatX ownership, fixed fork hooks, migration/recovery invariants,
feature-off compatibility, and the no-release boundary. This package has no
new product feature scope and must not merge the sync directly into `main`.

Known dry-run evidence is 99 upstream commits, 563 changed paths, 224
sensitive paths, 64 CatX touchpoint overlaps, and 12 conflicts. The actual
merge conflict list, semantic resolutions, verification, exact candidate SHA,
sync PR, and `develop` merge SHA belong in
`docs/30_UPSTREAM_3_9_0_SYNC.md` and must be updated as evidence is produced.

The exact qualified candidate was merged into `develop`. `v0.1.0`, RC tags,
`main`, and release publication remain unchanged.

Completion record:

- upstream merge commit: `9335eb4af5ab4c73976fdcbeda875926400e53f8`;
- final qualified code candidate: `196ca986414bcc8e6ec7af4a9d807669546faaee`;
- sync PR: [#4](https://github.com/CatCodeArbelin/CatX-UI/pull/4);
- `develop` merge commit: `c5f2e4e0165d96577af3f0e9d3b40810a98e7764`;
- `main` remains `7cb97e13def59280e92e5667d3f43ac23db8d470`;
- Xray version after sync: `26.9.30`;
- release publication and tag movement: not performed.

# CatX v0.2.0 Release Qualification — CURRENT

Branch: `feature/v0.2.0-release-qualification`

Release baseline: `31a442cef876669aeae97ba9cb1e3e8888f42613`.

Starting `origin/main` is `7cb97e13def59280e92e5667d3f43ac23db8d470` and is an
ancestor of the release baseline. The package qualifies the existing CatX
v3.9.0 integration for one public `v0.2.0-rc.1`, actual-public-artifact
observation, and stable `v0.2.0` only if that exact RC passes. It is not an
upstream sync and must not add product features or unrelated refactors.

Release identity:

- previous stable: `v0.1.0` at runtime SHA `fd28ea7144147d9164b70810d4a24872a3d48b4f`;
- fork version: `0.2.0`;
- first RC: `0.2.0-rc.1`;
- frozen RC candidate source: `f9178025d79e8bc403cf7f3f77f83aab3028965a`;
- immutable RC tag: `v0.2.0-rc.1 → f9178025d79e8bc403cf7f3f77f83aab3028965a`;
- public RC release: [v0.2.0-rc.1](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.2.0-rc.1);
- actual-public RC observer: [run 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602) — PASS, including public `v0.1.0` upgrade/rollback;
- RC Docker qualification: [run 37149015510](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149015510), job `111295304468` — PASS on attempt 4; both GHCR aliases resolve to manifest `sha256:cc7021156c971ac1f23d3b9728a5720afcd5a3d04e76ea62495efd4eaf762adf` with OCI revision `f9178025d79e8bc403cf7f3f77f83aab3028965a`;
- upstream: `MHSanaei/3x-ui v3.9.0` at `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb`;
- upstream integration merge: `c5f2e4e0165d96577af3f0e9d3b40810a98e7764`;
- Xray: `26.9.30`.

Required evidence is maintained in
`docs/31_V0_2_0_RELEASE_QUALIFICATION.md`: exact candidate SHAs, release
identity, all verification categories, SQLite/PostgreSQL upgrades, backup and
restore, updater transaction, rollback, real `restartPanel`, live Xray/TUIC,
subscriptions, AmneziaWG, multi-node, Traffic Control/quota, policy,
Activity/DNS, localization, non-publishing matrix, public RC observation,
stable publication, and immutable-tag verification.

Do not create `rc.2` unless public RC observation finds a genuine source,
release, migration/update/rollback, artifact, or runtime defect. Do not mutate
an immutable tag. Stable must reuse the proven product/runtime source and may
contain only the required RC-to-stable identity/evidence change. The optional
Claude review workflow is not a release gate unless repository governance
explicitly requires it.

The package is complete only after stable qualification and publication, or
with an exact documented blocker and the required final status phrase.

# Operating Rules

## Branching

One branch per work package, not per checkpoint.

Example:

```text
feature/wp-0a-foundation-guardrails
```

Inside that branch the agent may create multiple focused commits:

```text
test: establish upstream compatibility baseline
feat: add no-op fork hook contracts
ci: add verify-fork gate
feat: add fork feature flags
feat: add empty fork registries
```

## Git automation

The agent may:

- create/switch the work-package branch;
- commit checkpoints;
- push the branch.

The agent must NOT merge into `develop` unless the package merge gate is satisfied.

A merge is forbidden if required tests were not executed successfully.

## Human interaction

Normal workflow should require human review only at work-package boundaries, not after every microtask.

## Expensive model usage

Prioritize Sol for:

```text
WP-0B
WP-0C
WP-1B
WP-2A
WP-3B
WP-4B
WP-6A
WP-6B
WP-8C
upstream synchronization
security/recovery blockers
```

Use Luna Medium for the majority of implementation work.
