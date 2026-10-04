# CatX Product Reality Audit — post-Sponsors

Status: `AUDIT READY — DRAFT PR FOR MAINTAINER REVIEW`

Audited develop: `6207a2d6fc94e91b96006d0ac68f350c355879a8`

Audit branch: `chore/catx-product-reality-audit`

Accepted Sponsors Management head: `3cc4e111fcf6d9161f02d540cbf07a37e271388f`
Sponsors Management merge: `6207a2d6fc94e91b96006d0ac68f350c355879a8`

This audit records what the integrated product actually exposes and what has
representative runtime evidence. It does not treat a route, source package,
roadmap item, or `DONE` label as proof of a working end-to-end capability.

## Evidence and method

Evidence was collected from:

- the integrated source tree, fork registry, feature settings, migrations, and
  route registration;
- package and integration tests, including the exact Sponsors SQLite and
  PostgreSQL acceptance suite;
- exact-SHA hosted checks for the accepted Sponsors head: [CI run 37216324289](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37216324289),
  [Fork Verification run 37216324233](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37216324233),
  [Docs CI run 37216324242](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37216324242),
  and [Release run 37216324232](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37216324232);
- the prior public stable observer evidence recorded in
  [docs/31_V0_2_0_RELEASE_QUALIFICATION.md](./31_V0_2_0_RELEASE_QUALIFICATION.md),
  including real panel/Xray restart, feature-off, traffic, policy, and
  multi-node observations;
- disposable synthetic data only. No production sponsor content, credentials,
  cookies, decrypted bodies, or Authorization headers were collected.

Evidence labels in the matrix:

- `YES` means the capability is present and has direct evidence for that
  column;
- `PARTIAL` means source/package or focused tests prove only part of the
  claim;
- `NO EVIDENCE` means the audit did not find a representative proof;
- `N/A` means the column does not apply to that capability.

## Capability reality matrix

| Capability                              | Implemented | UI exposed                                      | Runtime tested | Single node | Multi-node  | Feature off | Restart tested | Limitation / evidence boundary                                                                                                   |
| --------------------------------------- | ----------- | ----------------------------------------------- | -------------- | ----------- | ----------- | ----------- | -------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Login and session auth                  | YES         | YES                                             | YES            | YES         | N/A         | YES         | YES            | Upstream path; covered by release qualification.                                                                                 |
| Dashboard                               | YES         | YES                                             | YES            | YES         | YES         | YES         | YES            | Core upstream surface.                                                                                                           |
| Clients, inbounds, subscriptions        | YES         | YES                                             | YES            | YES         | YES         | YES         | YES            | Core upstream behavior; no CatX replacement model.                                                                               |
| Xray lifecycle and config               | YES         | YES                                             | YES            | YES         | YES         | YES         | YES            | Xray is external; optional API capability differences remain.                                                                    |
| Activity                                | YES         | YES (`/activity`)                               | YES            | YES         | PARTIAL     | YES         | YES            | Real stable observer plus analytics tests; no fresh multi-node Activity run in this audit.                                       |
| Traffic History                         | YES         | PARTIAL                                         | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Repository/API coverage exists; end-to-end visual and multi-node proof is thinner than the claim.                                |
| DNS Intelligence                        | YES         | YES                                             | YES            | YES         | PARTIAL     | YES         | YES            | Real observer and metadata-only tests; enrichment provider availability is environment-dependent.                                |
| ASN / GeoIP enrichment                  | YES         | PARTIAL                                         | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Enrichment is consumed by analytics/risk; no separately verified product dashboard.                                              |
| Analytics retention and deletion        | YES         | YES                                             | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Persistence and retention tests pass; long-running retention soak not repeated here.                                             |
| Policy entities and assignment          | YES         | YES (`/policies`)                               | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | CRUD/compiler tests are strong; live multi-node policy application needs a dedicated run.                                        |
| Policy precedence and overrides         | YES         | YES                                             | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Deterministic compiler and schedule tests; live Xray route proof is limited.                                                     |
| Categories and schedules                | YES         | YES                                             | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Timezone/DST tests exist; no fresh operator flow across DST and restart.                                                         |
| Managed DNS and SafeSearch              | YES         | YES                                             | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Capability-aware APIs and explainability tests; provider/core support remains bounded.                                           |
| Policy Simulator and Explain            | YES         | YES                                             | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Read-only simulator/compiler tests; no multi-node or live traffic proof.                                                         |
| Traffic-control capability detection    | YES         | PARTIAL                                         | YES            | YES         | PARTIAL     | YES         | YES            | Stable observer and backend tests; Linux privileges/capabilities are required.                                                   |
| Group Quota                             | YES         | PARTIAL                                         | YES            | YES         | PARTIAL     | YES         | YES            | Accounting and reset tests plus release observation; no full multi-node soak here.                                               |
| Fixed/rolling windows and soft throttle | YES         | PARTIAL                                         | YES            | YES         | NO EVIDENCE | YES         | PARTIAL        | Lifecycle/math tests are present; generic per-user kernel attribution is unsupported.                                            |
| Traffic apply/remove/no-clobber         | YES         | PARTIAL                                         | YES            | YES         | PARTIAL     | YES         | YES            | Ownership-safe Linux backend and prior observer evidence; unsupported backends degrade safely.                                   |
| Risk Intelligence                       | YES         | YES (API/UI entry)                              | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Explainable metadata signals; no automatic ban and no multi-node product smoke.                                                  |
| Durable Audit                           | YES         | YES (`/audit`)                                  | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | SQLite/transaction/sanitization tests; full operator lifecycle not freshly exercised.                                            |
| Webhooks                                | YES         | YES (`/webhooks`)                               | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Delivery/security code and tests exist; runtime enablement is bundled with Audit.                                                |
| Prometheus / Metrics                    | YES         | API exposed                                     | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Bounded-label handler exists; no deployed scrape/alert run in this audit.                                                        |
| Self-service Portal                     | YES         | YES (`/portal-access`)                          | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Portal ownership/session tests exist; no external-client end-to-end run here.                                                    |
| Fleet dashboard                         | YES         | YES (`/fleet`)                                  | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Inventory and capability surfaces exist; multi-node live evidence is not complete.                                               |
| Fleet Update                            | YES         | YES (`/fleet-updates`)                          | PARTIAL        | YES         | NO EVIDENCE | YES         | PARTIAL        | Planning/reconciliation is implemented; mutation is separately gated and not assumed active.                                     |
| Sponsors feature flag                   | YES         | YES                                             | YES            | YES         | N/A         | YES         | PARTIAL        | Exact acceptance suite proves disabled/no-fetch and reload boundary; isolated real panel restart is still a v0.3.0 evidence gap. |
| Sponsors local provider                 | YES         | YES (`/catx/sponsors`, `/catx/sponsors/manage`) | YES            | YES         | YES         | YES         | PARTIAL        | SQLite/PostgreSQL CRUD, audit, PUT, and feature-off tests pass; no multi-node replication by design.                             |
| Sponsors remote provider                | YES         | YES                                             | YES            | YES         | N/A         | YES         | PARTIAL        | HTTPS/SSRF/redirect/size/MIME tests pass; remote mode is read-only and operator-source-only.                                     |
| Sponsors scheduling, slots, preview     | YES         | YES                                             | YES            | YES         | N/A         | YES         | PARTIAL        | Deterministic windows/slot caps and frontend preview tests pass; login slot is intentionally unsupported by CatX.                |
| Sponsors audit and persistence          | YES         | YES                                             | YES            | YES         | N/A         | YES         | PARTIAL        | Metadata-only atomic mutation audit and startup/reload migration instrumentation pass.                                           |

## Reality findings by classification

### Product source defects

No new source-level blocker was proven during this audit. The accepted
Sponsors correction remains low-divergence: the upstream Sanaei implementation
is retained in upstream-owned source, while its legacy public routes are not
registered in CatX composition.

### Migration defects

No migration defect was proven. Sponsors schema creation is in the canonical
startup migration path, SQLite and PostgreSQL tests pass, repeated migration is
idempotent, and runtime reload does not invoke `sponsors.Migrate`. Feature
schemas that are enabled after startup require the existing restart-required
configuration flow; that is intentional lifecycle behavior, not a runtime
schema mutation path.

### Frontend defects

No blocking frontend defect was proven. The management page has direct title
coverage, complete PUT submission, feature-off handling, and local/remote
read-only behavior. The reusable upstream Sponsor schema still contains the
historical login-slot and `/sponsors/logo` example shape; CatX production
responses do not advertise login slots and the legacy route is unreachable.
This is a stale upstream schema/documentation boundary, not an active fetch
path.

### Harness and infrastructure findings

- The Windows host does not provide the CGO SQLite toolchain required by the
  native SQLite backup/limit APIs and race tests. Linux container and hosted
  runs supplied the valid verification environment.
- The repository’s release workflow intentionally skips disposable Linux and
  PostgreSQL staging on ordinary feature PRs. The canonical PR CI PostgreSQL
  job was strengthened to run the Sponsors PostgreSQL test and passed on the
  accepted exact SHA; release staging remains a policy-gated follow-up.
- Existing roadmap/task documents used `DONE` broadly without a per-capability
  runtime evidence matrix. This audit supplies that missing reality record.

### Intentional limitations

- Xray remains an external runtime; unsupported optional APIs and generic
  per-user kernel attribution are reported rather than fabricated.
- Remote Sponsors mode is read-only, has no implicit local merge, and has no
  automatic Sanaei fallback or remote import.
- Sponsors do not replicate or synchronize local records across nodes.
- Webhooks and metrics are currently operationally bundled with the Audit
  runtime gate; their inventory flags are reserved rather than independently
  exposed in the feature settings UI.
- CatX supports Sponsor slots `dashboard`, `sidebar`, and `page`; the upstream
  `login` slot remains intentionally unavailable to unauthenticated CatX
  rendering.

## v0.3.0 scope decision

### MUST FIX BEFORE v0.3.0

1. Add a disposable end-to-end Sponsors smoke that uses the real panel,
   authenticated management API, local CRUD, provider persistence, feature
   disable/re-enable, and the real `POST /panel/api/setting/restartPanel`,
   asserting data survives and no schema migration runs during restart.
2. Run a representative multi-node reality qualification for Policy, Traffic,
   Risk, Fleet, Portal, and Sponsors, or narrow their product claims to the
   single-node evidence that actually exists.

### SHOULD FIX BEFORE v0.3.0

1. Give Webhooks and Metrics independent feature lifecycle controls if they are
   intended to be independently disableable; otherwise document the Audit
   coupling in the operator UI and API descriptions.
2. Remove or clearly namespace the historical upstream Sponsor schema examples
   from generated API documentation without deleting the retained upstream
   implementation.
3. Add a deployed Prometheus scrape and Webhook delivery/retry smoke, plus a
   dedicated Traffic History visual flow test.

### SAFE TO DEFER

- remote Sponsor import and multi-node Sponsor replication;
- generic per-user traffic attribution on Xray backends that do not expose the
  required capability;
- richer standalone ASN/GeoIP and service-classification dashboards;
- longer retention/soak and cross-node risk analytics qualification.

### DOCUMENTATION ONLY

- Replace broad `DONE` claims with the matrix categories above as each
  capability receives direct runtime evidence.
- Keep the explicit feature-off, restart-required, provider-authority, and
  unsupported-capability wording next to UI/API claims.

## Corrections made during this audit

Only the audit work-package records were updated: `CURRENT_TASK.md`,
`TASK_QUEUE.md`, and this document. No major feature, runtime behavior,
database schema, version, release, tag, or `main` change was made during the
audit.

## Final invariants

- `main` unchanged;
- `v0.2.0` and `v0.2.0-rc.1` unchanged;
- no release published and no upstream sync performed;
- Sponsors PR #10 merged only after complete exact-SHA acceptance;
- this audit branch will be pushed as a Draft PR and left open/unmerged;
- no major feature was added during the audit.
