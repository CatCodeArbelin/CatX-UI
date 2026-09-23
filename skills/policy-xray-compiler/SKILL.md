# SKILL: Policy/Xray Compiler Engineer

## Rules

- upstream Xray config generation remains authoritative;
- compiler is deterministic;
- compiler emits decision trace;
- feature OFF changes nothing;
- do not create per-user inbound unless unavoidable;
- use existing Xray user/routing semantics;
- validate before reload;
- rollback on failed healthcheck.

## Tests

Table-driven:

- client allow;
- client block;
- group rules;
- category;
- precedence;
- schedule;
- temporary override;
- outbound;
- unknown target;
- feature disabled.

Golden config fixtures required.
