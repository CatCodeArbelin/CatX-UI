# Testing, Rollback, and Release

## 1. Verification targets

### Upstream

```bash
make verify
```

### Fork

```bash
make verify-fork
```

It should include, as applicable:

```text
make verify
fork unit tests
SQLite integration
PostgreSQL integration
migration tests
parser fuzz
race tests
Xray config fixtures
frontend typecheck/tests
Docker/package smoke
```

## 2. Test pyramid

### Unit

- policy precedence;
- schedules;
- parser;
- classifier;
- correlator;
- quota math;
- updater version parsing;
- snapshot state machine.

### Integration

- repositories;
- SQLite/PostgreSQL migrations;
- Xray config decoration;
- API/routes;
- retention;
- events;
- updater staging.

### Smoke

- fresh install;
- upgrade from previous custom release;
- Xray boot;
- panel login;
- create client;
- generate subscription;
- all fork features OFF;
- enable one fork feature;
- rollback.

## 3. Golden compatibility fixtures

Maintain fixtures proving:

```text
upstream input
→ fork with features OFF
→ semantically equivalent output
```

Especially for:

- Xray config;
- subscriptions;
- client behavior.

## 4. Parser tests

Cover:

- valid lines;
- malformed lines;
- partial final line;
- huge line;
- truncate;
- rename rotation;
- copytruncate;
- restart with saved offset;
- offset beyond file length;
- duplicate replay;
- backpressure.

Fuzz parsers.

## 5. Race tests

Mandatory for concurrent custom modules:

```bash
go test -race ./internal/analytics/... ./internal/policy/... ./internal/trafficcontrol/...
```

Adapt paths to actual implementation.

## 6. Migration tests

At minimum:

```text
upstream baseline schema → current fork
previous custom release → current fork
empty DB → current fork
```

Both SQLite and PostgreSQL.

## 7. Config apply transaction

```text
generate candidate
→ validate
→ create snapshot
→ apply
→ reload/restart
→ healthcheck
→ mark committed
```

Failure:

```text
restore previous
→ reload/restart
→ healthcheck
→ audit/report
```

## 8. Update transaction

```text
download
→ verify checksum/signature
→ stage
→ backup
→ install
→ migrate
→ start
→ healthcheck
→ commit
```

Failure requires binary/config/DB recovery strategy.

## 9. Release candidate

Use RC for changes touching:

- migrations;
- updater;
- policy compiler;
- shaping;
- node protocol;
- core subscription generation.

## 10. Merge gate

No merge to `main` if:

- tests fail;
- migration behavior is unknown;
- dangerous rollback is untested;
- isolation feature flag is missing where needed;
- generated files are stale;
- race issues remain;
- privacy boundary is violated.

## Synthetic Module Acceptance

Between unit testing and full runtime smoke, CatX uses deterministic synthetic
module acceptance tests.

The goal is to prove complete module behavior through production integration
boundaries.

Recommended deterministic fixture vocabulary:

- user: `catx-test-user`
- group: `catx-test-group`
- node: `catx-test-node`
- inbound: `catx-test-inbound`
- domain: `video.example.test`
- secondary domain: `social.example.test`
- SNI: `video.example.test`
- upload traffic: `100 MiB`
- download traffic: `250 MiB`
- deterministic fixed clock
- deterministic fixed seed

Use reserved test names/domains and ephemeral state.

Randomized/fuzz testing is separate from deterministic acceptance.

### Test layers

1. UNIT
2. SYNTHETIC INTEGRATION
3. DATABASE
4. REAL RUNTIME
5. FRONTEND
6. HUMAN REVIEW

A test must assert the actual promised effect.

Examples of insufficient proof:

- HTTP request returned 200;
- a form value was saved;
- a DOM element exists;
- a database row exists.

Examples of stronger proof:

- policy changes effective Xray configuration;
- disabling policy removes the CatX routing delta;
- synthetic traffic changes exact quota/accounting state;
- DNS/SNI evidence appears through the production analytics path;
- a revoked portal credential can no longer authenticate;
- an audit action creates the expected durable event.
