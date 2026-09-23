# Definition of Done

A feature is DONE only when all applicable conditions are true.

## Product

- requirement works;
- behavior documented;
- limitations documented;
- feature flag exists where isolation is needed;
- disabled behavior preserves upstream semantics.

## Backend

- service/repository boundaries respected;
- errors handled;
- no sensitive logs;
- concurrency safe;
- cancellation/shutdown handled.

## Database

- migration exists;
- SQLite tested;
- PostgreSQL tested;
- populated upgrade tested;
- indexes reviewed;
- retention impact reviewed.

## Xray

- candidate config validated;
- feature-off compatibility checked;
- Xray start/reload smoke passed;
- policy trace available where relevant.

## Frontend

- loading/error/empty states;
- i18n;
- typecheck;
- important interactions tested;
- dangerous actions confirmed;
- unsupported capability clearly shown.

## Tests

- unit;
- integration;
- regression;
- race where concurrent;
- fuzz where parser;
- `make verify`;
- `make verify-fork`.

## Operations

- rollback/recovery defined;
- audit event for sensitive actions;
- metrics/logging sufficient for debugging;
- release notes updated;
- no undocumented manual recovery required for normal failure.

## Git

- focused PR;
- no unrelated formatting/refactor;
- generated files updated through proper mechanism;
- upstream-touch budget reviewed.
