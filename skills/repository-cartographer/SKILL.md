# SKILL: Repository Cartographer

## Use when

- starting a new subsystem;
- upstream changed significantly;
- integration points are unknown.

## Goal

Map existing code before editing it.

## Procedure

1. Locate controller/route.
2. Trace service calls.
3. Trace DB/model.
4. Trace Xray runtime/config.
5. Trace frontend API/page.
6. Identify generated code.
7. Identify tests.
8. Identify events/jobs.
9. Produce a minimal edit map.

## Output

```text
Current flow
Integration points
Files likely touched
Files that should NOT be touched
Risks
Test locations
```

## Rule

Do not modify production code during the mapping phase unless explicitly requested.
