# Feature Catalog

Legend:

- **P0** foundation/blocker
- **P1** core product
- **P2** important enhancement
- **P3** later/experimental

# P0 — Fork Foundation

## F-001 Fork identity

Track:

```text
Fork version
Upstream base
Upstream latest
Xray version
```

## F-002 Own updater

- fork-owned release source;
- stable/dev channels;
- backup;
- integrity verification;
- healthcheck;
- rollback.

## F-003 Feature flags

Large modules independently disableable.

## F-004 Config snapshots

Snapshot before risky config changes.

## F-005 Validate / apply / rollback

For Xray config, policy, routing, DNS enforcement, updater.

## F-006 Fork audit trail

Record administrative mutations and rollback/update events.

# P1 — Analytics / DNS / Destination Intelligence

## F-101 Incremental access.log tailer

Must:

- keep offset;
- detect truncate;
- detect rotation/replacement;
- resume safely;
- avoid whole-file reread;
- batch events;
- support graceful shutdown;
- use bounded backpressure.

## F-102 Xray API observer

Use current Xray APIs for online/stat data where available.

## F-103 DNS observer

Collect where attributable:

- timestamp;
- client;
- query;
- query type;
- result metadata;
- allow/block decision;
- resolver path/source.

Do not claim complete DNS visibility.

## F-104 Destination observer

Fuse:

- Xray requested destination;
- DNS evidence;
- SNI where visible;
- plaintext HTTP Host;
- destination IP;
- TLS/QUIC metadata;
- inbound/outbound tags.

## F-105 ASN / GeoIP

Optional enrichment.

Failure must not block traffic.

## F-106 Service classifier

Map:

```text
evidence
→ service
→ category
```

Examples:

```text
googlevideo.com → YouTube → Streaming
cdninstagram.com → Instagram → Social
```

Keep confidence/evidence.

## F-107 Session correlator

Deduplicate multiple observations into logical sessions.

## F-108 Client Activity page

Show:

- recent services/destinations;
- allow/block action;
- first/last seen;
- traffic;
- session count;
- source/confidence.

## F-109 Traffic history

Filters:

```text
Today
7 days
30 days
custom range
```

Breakdowns:

- client;
- inbound;
- node;
- outbound where reliable;
- service;
- category.

History persists across upstream counter reset.

## F-110 DNS Dashboard

- query count;
- blocked;
- unique domains;
- new domains;
- NXDOMAIN;
- top services/categories;
- clients.

## F-111 Privacy Dashboard

Display:

```text
HTTPS content      NOT COLLECTED
Cookies            NOT COLLECTED
Passwords          NOT COLLECTED
Request bodies     NOT COLLECTED
```

Also show collected metadata categories.

## F-112 Retention controls

Configurable retention + delete history.

## F-113 Service/DNS relationship graph

Group multiple infrastructure domains under a logical service.

## F-114 New-domain / first-seen intelligence

Per-client/global first seen, last seen, category, confidence.

# P1 — Policy Engine

## F-201 Policy entity

Reusable human-friendly policy.

Potential fields:

- allowed inbounds;
- traffic quota;
- IP limit;
- allowed/blocked domains;
- categories;
- outbound preference;
- DNS policy;
- schedule;
- speed profile.

## F-202 Group policy assignment

Group → Policy.

## F-203 Client override

Support:

```text
Inherited
Override
Return to Policy
```

## F-204 Policy precedence

Recommended:

```text
1 emergency/system
2 temporary client allow
3 explicit client allow
4 explicit client block
5 group allow
6 group block
7 category policy
8 global/upstream routing
9 default outbound
```

Final precedence must be documented and tested.

## F-205 Domain allow/block

Per-client/group rules compiled to existing Xray mechanisms.

## F-206 Category policy

Logical categories:

```text
Social
Adult
Gambling
Ads
Trackers
Malware
Crypto
Gaming
Streaming
Messaging
AI
```

Do not permanently bind DB semantics to one `geosite:*` token.

## F-207 Schedules

Weekly schedules.

## F-208 Temporary overrides

Examples:

```text
Allow YouTube 15 min
30 min
1 hour
until 22:00
```

Require `expires_at`, cleanup, and audit.

## F-209 Quarantine

Restrict Internet while preserving explicitly required management/subscription paths.

## F-210 Managed DNS policy

Optional:

- managed resolver;
- direct port 53 restriction;
- optional DoT restriction;
- optional known DoH endpoint restrictions.

Do not claim perfect DoH prevention.

## F-211 SafeSearch

Optional provider-supported DNS/routing enforcement.

## F-212 Policy Simulator

Input:

```text
client
destination
protocol
```

Output:

```text
matched policy
matched category
matched rule
action
outbound
reason
```

## F-213 Explain Decision / Explain Route

Show actual decision trace.

# P1/P2 — Traffic Control

## F-301 Shared group quota — P1

One traffic pool for a group.

## F-302 Rolling/window quota — P2

Examples:

```text
10 GB / 2 hours
50 GB / day
```

Action:

```text
disable
or throttle
```

## F-303 Post-quota/expiry soft throttle — P2

Default remains current upstream disable behavior.

Optional throttle state.

## F-304 Per-client speed limit — P2

Upload/download limits.

Must use isolated enforcement adapter.

## F-305 Per-inbound traffic breakdown — P1

Per client and aggregate.

## F-306 Per-node traffic breakdown — P1

For multi-node deployments.

## F-307 Traffic multiplier — P3

Accounting-only fixed-point multiplier.

## F-308 Per-category quota/cap — P3

Only when classification confidence is high enough.

# P1/P2 — Security / Anomaly

## F-401 IP/session history — P1

## F-402 Account-sharing risk engine — P2

Signals may include:

- unique IP count;
- ASN count;
- country changes;
- simultaneous sessions;
- datacenter ASN;
- rapid geography changes;
- device/HWID changes.

Never:

```text
IP > N → automatic ban
```

## F-403 New ASN/country alerts — P2

## F-404 DNS anomaly heuristics — P3

Examples:

- high NXDOMAIN;
- entropy;
- burst of new domains;
- beacon-like frequency;
- possible DGA.

Always label as heuristic.

## F-405 Outbound/node health notifications — P2

Extend upstream health/events.

# P2 — Client / Subscription Usability

## F-501 Group subscription link

Only if reusable from existing subscription engine.

## F-502 Resource visibility

Allow selected Hosts/resources to be visible to:

```text
all
groups
clients
```

Initial scope: Hosts.

## F-503 Self-service portal

Safe actions only:

- quota/expiry display;
- device/HWID display;
- self-service old-device reset with rate limits;
- QR/subscription;
- optional token rotation with confirmation.

## F-504 Small routing/subscription improvements

Expose existing Xray capabilities without a second configuration system.

# P2/P3 — Routing / Observatory Enhancements

## F-550 Outbound quality dashboard — P2

Extend upstream observatory/health.

## F-551 Smart failover policy — P3

Build only on upstream-supported mechanisms.

## F-552 Route quality scoring — P3

Derived explainable operator score.

# P2 — Operations

## F-601 Backup validation

Verify backup readability/restorability where practical.

## F-602 Config diff

Show candidate vs current.

## F-603 Webhooks

Events:

- client enabled/disabled;
- quota exhausted;
- policy change;
- anomaly;
- node up/down;
- outbound change;
- update success/failure;
- rollback.

## F-604 Prometheus

Fork metrics plus safe upstream metrics.

## F-605 Fleet dashboard

Extend multi-node:

- fork version;
- Xray version;
- health;
- CPU/RAM;
- traffic;
- clients;
- update state.

## F-606 Remote Xray version control — later

Only if cleanly supported through upstream APIs.

## F-607 Logging level/retention controls

Community-driven quality-of-life enhancement.
