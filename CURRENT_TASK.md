# CURRENT TASK

## Work Package

`WP-0B — Release Identity & Updater Safety`

## Recommended model

**GPT-5.6 Sol High**

## Mandatory reads

- `AGENTS.md`
- `CURRENT_TASK.md`
- `TASK_QUEUE.md`
- `docs/19_REPOSITORY_MAP.md`
- `docs/12_UPDATER_IDENTITY.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/04_UPSTREAM_SYNC.md`
- `adr/0004-own-updater.md`
- `skills/test-release-guardian/SKILL.md`
- `skills/security-privacy-review/SKILL.md`
- `skills/feature-implementer/SKILL.md`

## Scope

### CP-0B.1 — CatX-UI identity and version model

- fork version;
- upstream base version;
- Xray version remains separate;
- stable/dev release channels;
- no fake or hardcoded version ambiguity.

### CP-0B.2 — Updater isolation

- audit and retarget all official-upstream update/install paths;
- verified hotspots include:
  - `internal/web/service/panel/panel.go`
  - `update.sh`
  - `x-ui.sh`
  - `install.sh`
  - `.github/workflows/release.yml`
- CatX-UI must never accidentally install an official `MHSanaei/3x-ui` binary.

### CP-0B.3 — Release integrity

- CatX-UI-owned release source;
- consistent artifact naming;
- checksum/integrity verification;
- tests that reject wrong repository/assets;
- no arbitrary unverified update payload.

### CP-0B.4 — Transactional updater

Required flow:

```text
download → verify → stage → backup → install candidate → migrate → start → healthcheck → commit
```

Failure flow:

```text
restore previous binary/config/state → recover DB as required → start → healthcheck → audit/report
```

### CP-0B.5 — Verification

- unit tests for version/release selection;
- fake release provider/server where appropriate;
- updater failure-path tests;
- rollback tests;
- GitHub Actions verification;
- preserve existing `make verify-fork`;
- no weakening of WP-0A gates.

## Critical rules

- Preserve upstream 3x-ui architecture.
- Do not rewrite the panel updater into a separate control plane.
- Minimize permanent upstream-touch points.
- Do not touch Policy Engine, Analytics, DNS Intelligence, or QoS.
- Do not fork xray-core.
- Do not change `main`.
- Do not begin WP-0C.
- Do not merge into `develop` during implementation.
- Every dangerous updater action must have recovery.
- Tests must actually run; green claims without execution are forbidden.

## Administrative boundary

This task activates WP-0B only. Do not begin implementation until the work-package branch is prepared and reviewed.
