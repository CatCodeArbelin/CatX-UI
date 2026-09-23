# Policy Engine

## Purpose

Provide human-friendly per-client/group control without replacing upstream routing.

## Entities

### Policy

Reusable policy definition.

### Assignment

Binds policy to group/client.

### Override

Client-specific deviation.

### TemporaryOverride

Time-bounded exception.

### Schedule

Recurring policy windows.

## Inheritance

Recommended:

```text
global upstream defaults
→ group policy
→ client policy/override
→ temporary override
```

Define exact precedence and test it.

## Compilation

```text
Policy
→ compiler
→ minimal Xray routing representation
→ validator
→ apply
```

Do not create a second full Xray config generator.

## Per-user routing

Use existing Xray user/client routing semantics where supported.

Avoid creating separate inbound per user.

## Categories

Stable product categories map through replaceable providers such as:

```text
geosite
custom domain sets
maintained local/remote lists
```

## Schedules

Use configured panel timezone.

Test DST/time boundaries.

## Temporary overrides

Fields:

```text
client/group
scope
action
created_at
expires_at
created_by
optional reason
```

Expiration must work correctly after restart.

## Quarantine

Use explicit allowlists for management/subscription paths that must remain available.

## Policy Simulator

Must reuse production decision/compiler logic.

## Explainability

Every decision should be traceable to source policy/rule.

## Rollback

Risky policy application uses the shared snapshot/apply/healthcheck pipeline.

## Tests

Table-driven tests for:

- inheritance;
- priority;
- allow override;
- block;
- schedule boundary;
- temporary expiry;
- timezone;
- unknown category;
- feature disabled;
- generated routing.
