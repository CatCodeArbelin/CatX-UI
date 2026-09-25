# CURRENT TASK

## Work Package

`WP-4B — Quarantine / Managed DNS / SafeSearch`

## Status

Implementation authorized on the existing feature branch. Do not merge this
work package into `develop` or `main`.

## Scope

- quarantine behavior;
- dedicated bounded quarantine-release exceptions;
- managed DNS policy;
- SafeSearch where supported;
- simulator, explain, compiler, API/OpenAPI, UI, and test integration.

## Reviewed decisions

- Distinguish Xray's `dns` outbound from Xray built-in DNS. The `dns` outbound
  handles intercepted plaintext UDP/TCP DNS; built-in DNS may use supported
  upstream transports including DoH, DOHL, and DOQL. Reuse native Xray
  mechanisms where appropriate.
- `quarantine-release` is a dedicated exception to quarantine state. Generic
  allow, temporary, domain, service, category, or SafeSearch overrides must
  never bypass quarantine.
- Reuse `policies.enabled` and capability-level opt-in data. Add no new global
  feature flags unless an independent kill-switch is demonstrably required.
- No TLS MITM, HTTPS decryption, payload/header inspection, or false claims of
  enforcing client-controlled encrypted DNS.
- Reuse the existing policy engine, decision/explain path, Xray decorator,
  recovery path, schedules, overrides, category vocabulary, and upstream
  models. Do not create parallel policy/DNS/config subsystems.

## Required workflow

- Keep the working tree focused and clean apart from WP-4B changes.
- Run relevant unit/integration/race/Xray-validation tests, `make verify`, and
  `make verify-fork`.
- Make focused commits, push the feature branch, monitor CI, and fix genuine
  defects until both required workflows are green.
- Stop on the green feature branch for final review. Do not start WP-5A.

## Authorized documents

- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/06_DATABASE_MIGRATIONS.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/10_POLICY_ENGINE.md`

## Authorized skills

- `skills/security-privacy-review/SKILL.md`
- `skills/policy-xray-compiler/SKILL.md`
- `skills/database-migration/SKILL.md`
- `skills/feature-implementer/SKILL.md`
- documented DoH limitations.

## Hard constraints

- preserve existing upstream-compatible behavior when WP-4B capability data is
  absent or `policies.enabled` is false;
- no TLS MITM, HTTPS decryption, payload inspection, or main-branch changes.
