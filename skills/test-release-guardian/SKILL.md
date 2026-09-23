# SKILL: Test and Release Guardian

## Gate

No green tests = no merge.

## Checklist

- upstream verify;
- fork verify;
- race tests;
- parser fuzz;
- SQLite/PostgreSQL migrations;
- feature-off compatibility;
- Xray config validation;
- startup smoke;
- upgrade smoke;
- rollback smoke;
- privacy/logging review.

## Failure

If a critical test fails:

1. do not merge;
2. identify regression;
3. restore known-good state if environment changed;
4. document the result.

## Output

```text
PASS / FAIL
blocking findings
non-blocking findings
rollback verified yes/no
```
