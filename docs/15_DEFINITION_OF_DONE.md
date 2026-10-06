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
- expected feature-disabled states are explicit, localized, and actionable;
- raw backend errors are not used as normal feature-off UX;
- user-facing frontend changes receive manual visual QA across applicable
  theme, desktop/mobile, and LTR/RTL cases;
- when a work package declares a human approval gate, merge is forbidden until
  that approval is recorded.

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

## Module Qualification READY gate

`Implementation complete` does not mean `Module READY`.

A CatX module may become `READY` only when all applicable qualification gates
have passed:

- architecture/code review;
- frozen module contract;
- unit tests;
- synthetic integration tests;
- SQLite tests;
- PostgreSQL tests;
- migration/recovery tests;
- runtime acceptance;
- effective Xray-state validation where applicable;
- feature-disabled compatibility;
- restart/persistence behavior;
- failure/error-state semantics;
- en-US qualification;
- ru-RU qualification;
- frontend interaction tests;
- UI/visual review;
- explicit maintainer Human Review PASS;
- known limitations documented.

A skipped, unavailable, or blocked mandatory gate is not PASS.
