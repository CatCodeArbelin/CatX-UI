# Community-driven Backlog

This file captures current 3x-ui community demand that fits the fork philosophy.

Re-check each issue before implementation because upstream may implement it first.

## Strong fits

- Client Group Policies — `#5900`
- Shared Traffic Quota for Client Groups — `#6534`
- Traffic Usage History by Custom Date Range — `#6353`
- Per-Inbound Traffic Statistics per Client — `#5649`
- Post-expiry throttle / window quotas / per-client speed limit — `#6622`
- Different Hosts for different clients/groups — `#6618`
- Self-service HWID management — `#6422`
- Outbound-change notifications — `#6425`
- Remote Xray version selection for subnodes — `#5948`
- Logging level/storage controls — `#6584`

## Conditional fits

- Group subscription link — `#5694`
- Named routing/subscription profiles — `#6322`
- Richer Clash/Mihomo profile generation — `#5889`

Conditional items must not cause architectural drift.

## Decision rule

A community feature belongs in the fork only if:

```text
useful
AND fits 3x-ui architecture
AND maintainable downstream
AND not already solved upstream
AND testable/rollback-safe
```

## Issue links

```text
https://github.com/MHSanaei/3x-ui/issues/5900
https://github.com/MHSanaei/3x-ui/issues/6534
https://github.com/MHSanaei/3x-ui/issues/6353
https://github.com/MHSanaei/3x-ui/issues/5649
https://github.com/MHSanaei/3x-ui/issues/6622
https://github.com/MHSanaei/3x-ui/issues/6618
https://github.com/MHSanaei/3x-ui/issues/6422
https://github.com/MHSanaei/3x-ui/issues/6425
https://github.com/MHSanaei/3x-ui/issues/5948
https://github.com/MHSanaei/3x-ui/issues/6584
https://github.com/MHSanaei/3x-ui/issues/5694
https://github.com/MHSanaei/3x-ui/issues/6322
https://github.com/MHSanaei/3x-ui/issues/5889
```
