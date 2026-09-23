# Analytics and DNS Intelligence

## Goal

Understand client network activity without decrypting HTTPS.

## Evidence sources

### Xray API

Primary for current online state/statistics when available.

### access.log

History/analytics source.

### DNS

Domain intent/evidence when observable.

### Destination metadata

Original requested destination where exposed.

### TLS/SNI

Metadata only.

### HTTP Host

Plaintext HTTP only.

## access.log collector state

Track:

```text
path
file identity/inode
offset
last size
partial-line buffer
```

Behavior:

- tail incrementally;
- checkpoint offset;
- recover after restart;
- detect `size < offset`;
- detect file replacement;
- handle rotation;
- avoid expensive regex-heavy parsing when simpler parsing works;
- bounded queue;
- batch DB writes.

## Normalized event

Every event records its source:

```text
SourceDNS
SourceAccessLog
SourceTLS
SourceProxyTarget
SourceXrayAPI
```

## Classification

```text
evidence
→ provider
→ service
→ category
→ confidence
```

Do not hard-bind product categories to one provider format.

## Session model

Network session duration is estimated activity, not screen time.

Fields may include:

```text
client
service/domain
first_seen
last_seen
connections
up
down
confidence
```

## DNS bypass limitations

Optional policy may restrict:

- direct port 53;
- DoT port 853;
- known public DoH endpoints.

Never promise complete DoH control.

## Retention

Delete in bounded batches.

## Metrics

Collect:

- events/sec;
- parse errors;
- dropped events;
- queue depth;
- batch latency;
- correlation rate;
- classification hit rate;
- unknown-service ratio.
