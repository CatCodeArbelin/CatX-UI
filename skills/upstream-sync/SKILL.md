# SKILL: Upstream Sync Guardian

## Procedure

1. Fetch upstream tags.
2. Record old/new upstream base.
3. Review release notes.
4. Diff semantic hotspots.
5. Create sync branch.
6. Merge upstream.
7. Resolve conflicts favoring upstream behavior.
8. Verify custom hooks.
9. Detect upstream implementations that duplicate fork features.
10. Run full tests.
11. Run migration upgrades.
12. Smoke Xray/subscriptions.
13. Produce a sync report.

## Required report

```text
Old/new upstream base
Conflicts
Semantic risks
Fork features affected
Upstream duplicates
Migrations
Test results
Rollback plan
```

## Never

Auto-merge directly to production.
