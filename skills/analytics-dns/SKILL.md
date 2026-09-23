# SKILL: Analytics/DNS Engineer

## Privacy boundary

Metadata only.

Never add decrypted HTTPS content fields.

## Collector rules

- incremental;
- bounded;
- cancellable;
- batched;
- rotation-aware;
- no network I/O under global locks;
- metrics for dropped/failed events.

## Correlation

Multiple observations may describe one session.

Store source + confidence.

Do not equate network session duration with screen time.

## Tests

- malformed input;
- rotation;
- truncate;
- duplicate replay;
- restart offset;
- queue pressure;
- classification;
- correlation;
- retention.
