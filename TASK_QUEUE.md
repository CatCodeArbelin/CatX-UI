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

## Stable Qualification — v0.1.0 — CURRENT

Branch: `feature/stable-v0.1.0-qualification`

Starting SHA: `ecd8ee1afbee4996f9a099d67014c4ab2236ae44`.

Qualify the already-observed RC3 code for stable without adding features,
starting Repository Productization, or touching `main` before all stable gates
pass. The first mandatory gate is a persistent disposable Linux regression
using actual public `v0.1.0-rc.3` assets. It must save
`analytics.enabled` and `dns_intelligence.enabled`, invoke the real
`POST /panel/api/setting/restartPanel` endpoint, verify recovery and Activity /
DNS Intelligence health, then disable both flags and repeat the real restart.

The workflow must retain sanitized evidence. A source-level defect blocks
stable and requires focused `v0.1.0-rc.4` qualification and observation; an
infrastructure-only failure must follow the documented retry policy.

After every exact-SHA stable gate passes, merge the qualified branch into
`main`, create annotated `v0.1.0` without moving any RC tag, verify the actual
published stable assets and release identity, run the narrow post-publication
stable smoke, then stop before Repository Productization.

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
