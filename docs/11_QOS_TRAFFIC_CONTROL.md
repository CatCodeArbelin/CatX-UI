# QoS and Traffic Control

## Status

High-value but technically risky.

Do not ship UI controls that cannot be enforced reliably.

## Goals

- per-client upload/download speed;
- shared group quota;
- rolling quotas;
- post-quota throttle;
- optional accounting multiplier.

## Architecture

Use an adapter:

```go
type Shaper interface {
    Capabilities() Capabilities
    ApplyClientLimit(...)
    RemoveClientLimit(...)
    Reconcile(...)
}
```

Possible Linux implementation may use kernel facilities such as `tc`/nftables where attribution is reliable.

Do not scatter shell execution throughout services.

## Capability detection

UI must clearly show unsupported/degraded enforcement.

## Reconciliation

Desired state must survive:

- panel restart;
- Xray restart;
- reconnect;
- client enable/disable;
- node restart where applicable.

## Quota accounting

Shared group quota updates must be atomic.

Avoid floating-point accounting.

## Rolling windows

Define fixed vs sliding windows explicitly.

Prefer fixed/tumbling windows first unless sliding is truly required.

## Soft throttle

State model:

```text
ACTIVE
→ quota reached
→ THROTTLED
→ renewal/reset
→ ACTIVE
```

Current upstream disable behavior remains default.

## Failure behavior

If shaping fails:

- mark degraded;
- alert/log;
- do not silently claim the limit is active.

## Privileges

Use minimum OS privilege required.

## Tests

- apply/remove;
- restart reconcile;
- multiple clients;
- limit changes;
- client disable;
- quota transition;
- group contention;
- failure cleanup.
