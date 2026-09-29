# CatX Localization Glossary

WP-9B uses this glossary for consistent operator-panel terminology. Localized
terms should be natural for the target language while preserving the meaning
of the English source. Product names, protocol names, route paths, enum values,
JSON field names, version strings, node IDs, run IDs, and technical examples
remain unchanged.

## Product and network identifiers

Keep `CatX-UI`, `3x-ui`, `Xray`, `DNS`, `SNI`, `ASN`, `IP`, `CIDR`, `API`,
`HTTP`, `HTTPS`, `UUID`, and `SafeSearch` unchanged unless normal target-locale
punctuation requires surrounding text to change.

## Operator terminology

| English | Guidance |
| --- | --- |
| Policy / Policy engine | Access and routing policy; use the established administrative term, not a legal-policy synonym. |
| Assignment | Binding a policy to a client or group. |
| Override / Temporary override | A deliberate precedence exception; retain the temporary/time-bounded distinction. |
| Quarantine | Restricted safety state; do not soften it to a generic pause. |
| Managed DNS / SafeSearch | DNS control and filtering features; preserve product meaning. |
| Activity / Session / Traffic | Event activity, connected session, and transferred traffic respectively. |
| Quota / Throttle | Usage allowance and deliberate rate limitation respectively. |
| Risk / Audit / Webhook | Risk signal, recorded administrative event, and HTTP callback respectively. |
| Fleet / Fleet update | Managed node group and an update operation across that group. |
| Campaign / Canary / Batch / Parallelism | Coordinated update plan, limited early target, grouped work, and concurrency limit. |
| Dry run / Dispatch / Reconcile / Soak / Rollback | Simulation without mutation, send/apply operation, compare and converge state, observation period, and restore previous state. |
| Portal / Host visibility / Device / Expiry / Retention | Client portal, permitted host display, registered client device, end time, and data holding period. |
| Evidence / Confidence | Supporting signal and certainty level; do not imply proof when the source says evidence. |
| Unsupported / Degraded / Unknown | Capability unavailable, operating with reduced capability, and state not established. |
| Feature / Module | A user-visible capability / its independently configured subsystem. |
| Feature disabled | The capability is intentionally unavailable; explain how to enable it where the operator can act. |
| Restart required | Saved configuration needs a panel restart before the module uses the new state. |

Operational words such as update, rollback, abort, retry, delete, revoke,
quarantine, reset, and dry run must preserve their actual action semantics in
every locale.
