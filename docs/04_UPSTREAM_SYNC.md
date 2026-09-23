# Upstream Synchronization

## Goal

Keep custom work continuously integrable with `MHSanaei/3x-ui`.

## Remotes

```bash
git remote add upstream https://github.com/MHSanaei/3x-ui.git
git fetch upstream --tags
```

## Production policy

Production follows stable upstream tags.

Optional CI may test latest upstream `main` for early conflict detection.

## Sync workflow

```bash
git fetch upstream --tags
git checkout develop
git checkout -b sync/upstream-vX.Y.Z
git merge vX.Y.Z
```

Resolve conflicts semantically.

Then run:

```bash
make verify
make verify-fork
go test -race ./...
```

Plus migration/integration/smoke tests.

## Merge vs rebase

Do not routinely rewrite long-lived downstream history.

Rebase is fine for short-lived local feature branches.

## Semantic conflicts

A conflict-free Git merge does not guarantee semantic compatibility.

Review upstream changes around:

- DB/model/migrations;
- Xray config generation;
- route registration;
- client/group models;
- traffic stats;
- subscriptions;
- updater;
- node APIs;
- event bus;
- jobs;
- frontend routes/forms.

## Compatibility CI

Recommended scheduled check:

```text
fetch upstream/main
→ temporary merge/compare
→ build
→ make verify
→ make verify-fork
→ report
```

Do not auto-merge.

## git rerere

Recommended:

```bash
git config rerere.enabled true
```

Still review every conflict resolution.

## Upstream-first rule

If upstream implements an equivalent feature:

1. compare behavior;
2. prefer upstream implementation;
3. migrate fork data/settings if needed;
4. remove duplicate fork code;
5. keep compatibility adapter only where required;
6. document deprecation.

Do not maintain two competing implementations permanently.

## Upstream contributions

Generic bugfixes that are not fork-specific should be proposed upstream when practical.
