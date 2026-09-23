# Architecture

## 1. Architectural rule

**Upstream 3x-ui remains the application. Fork modules extend it.**

Do not build a parallel control plane.

## 2. Proposed package layout

Names are recommendations and must be adapted to the real upstream tree.

```text
internal/
├── forkext/
│   ├── registry.go
│   ├── flags.go
│   └── lifecycle.go
│
├── policy/
│   ├── model/
│   ├── repository/
│   ├── service/
│   ├── compiler/
│   ├── schedule/
│   ├── category/
│   └── enforcement/
│
├── analytics/
│   ├── collector/
│   │   ├── accesslog/
│   │   ├── dns/
│   │   ├── xrayapi/
│   │   └── destination/
│   ├── normalizer/
│   ├── correlator/
│   ├── classifier/
│   ├── aggregator/
│   ├── repository/
│   └── service/
│
├── trafficcontrol/
│   ├── model/
│   ├── quota/
│   ├── shaping/
│   ├── adapter/
│   └── service/
│
├── forkops/
│   ├── snapshot/
│   ├── rollback/
│   ├── updater/
│   ├── audit/
│   └── health/
│
└── web/
    └── upstream...
```

Frontend:

```text
frontend/src/features/
├── policies/
├── activity/
├── dns-intelligence/
├── traffic-control/
├── fork-operations/
└── audit/
```

## 3. Minimal integration boundary

Prefer one registration area:

```go
type Extension interface {
    RegisterRoutes(...)
    RegisterJobs(...)
    RegisterEventSubscribers(...)
    RegisterMigrations(...)
}
```

Do not build a generic plugin framework in v1.

## 4. Xray boundary

```text
upstream config builder
        │
        ▼
fork decorator/compiler
        │
        ▼
candidate Xray config
        │
    validate/apply
        ▼
       Xray
```

Policy logic must not become a second full Xray config generator.

## 5. Analytics data flow

```text
Xray API ───────┐
access.log ─────┤
DNS observer ───┤
destination ────┤
TLS/SNI meta ───┤
                ▼
            Normalizer
                │
                ▼
        DestinationEvent
                │
          Correlator/Dedupe
                │
                ▼
              Session
            ┌───┴────┐
            ▼        ▼
        Analytics   Policy/Audit
```

## 6. Real-time vs history

Use Xray online/stat APIs for real-time state where available.

Use `access.log` for history/analytics, not as the primary online truth.

## 7. Event model

Conceptual:

```go
type DestinationEvent struct {
    Timestamp     time.Time
    ClientKey     string
    SourceIP      netip.Addr
    Domain        string
    DestinationIP netip.Addr
    Port          uint16
    Network       string
    Protocol      string
    Observation   string
    InboundTag    string
    OutboundTag   string
    Service       string
    Category      string
    Confidence    uint8
}
```

No sensitive HTTP-content fields.

## 8. Session correlation

One browsing action may produce:

```text
DNS
TLS SNI
access.log
routing metadata
```

Do not count each observation as a separate visit.

Use bounded correlation windows based on:

```text
client
domain/destination
port
network
time
```

Expose confidence when uncertain.

## 9. Database strategy

Use fork-specific tables and repositories.

Do not overload unrelated upstream columns with large fork JSON blobs.

## 10. Failure isolation

Analytics/enrichment/webhook failures must not stop Xray unless explicit enforcement depends on them.

Examples:

```text
analytics DB error → Xray continues
GeoIP unavailable → store event without GeoIP
webhook fails → retry/log, no traffic outage
```

## 11. No architectural transplant

Do not make baseline operation depend on:

- microservices;
- mandatory Redis/Kafka/ClickHouse;
- a new node-agent architecture;
- a new subscription control plane;
- Remnawave Squad/Profile concepts.
