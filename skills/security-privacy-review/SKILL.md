# SKILL: Security and Privacy Reviewer

## Mandatory checks

- no TLS MITM;
- no cookie/password/token capture;
- secret redaction;
- path traversal;
- command injection;
- SQL injection;
- authorization;
- CSRF where relevant;
- rate limiting;
- audit;
- least privilege;
- backup secrets;
- updater integrity.

## Risk engine

Heuristic signals must not be presented as certainty.

## Output

Findings sorted by:

```text
Critical
High
Medium
Low
Informational
```

Include concrete remediation and affected files.
