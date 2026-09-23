# SKILL: Modular Feature Implementer

## Procedure

1. Read `AGENTS.md`.
2. Read `CURRENT_TASK.md`.
3. Read task/module spec.
4. Confirm feature flag if required.
5. Implement fork module first.
6. Add minimal upstream registration hook.
7. Add migration if needed.
8. Add API/OpenAPI/i18n/frontend.
9. Add tests.
10. Run relevant verification.
11. Summarize upstream files touched.

## Stop and redesign if

- fork logic spreads through many upstream files;
- unrelated upstream semantics need changes;
- dangerous config has no rollback;
- feature-off compatibility cannot be tested.

## Never

- unrelated refactor;
- repository-wide formatting;
- silent API break.
