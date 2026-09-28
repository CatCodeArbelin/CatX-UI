# CURRENT TASK

## Work Package

`WP-8C — Multi-node Update Orchestration`

## Status

WP-8A is DONE: merged into `develop` at `0a6dab0f`, pushed, and its feature
branch was deleted after Fork verification `36347039861` and Release CatX-UI
`36347039863` reported green. Do not modify `main` or preview/demo data.

WP-8B is DONE: final head `fa030309d2628f764cff477896627e57adeed65f` passed
Fork verification `36357830273` and Release CatX-UI `36357830282`, was merged
into `develop` at `d9b09764`, pushed, and its feature branch was deleted.
WP-8C Stage A is authorized on `feature/wp-8c-multi-node-update-orchestration`.
Do not merge WP-8C, enable production fleet mutation, modify `main`, or touch
preview/demo data.

Stage A is orchestration foundation only. It may plan, preflight, reconcile,
and exercise a fake executor, but production reconciliation must not call
`runtime.Remote.UpdatePanel` yet. Do not use SSH, create another node agent,
bypass `runtime.Remote`, or update transitive/read-only nodes.

Required Stage A behavior includes persistent campaigns and immutable target
snapshots; stable/dev release resolution and checksum-trust preflight;
dry-run planning; direct-node eligibility and explicit blocked reasons;
deterministic campaign states; canary/batch/parallel/health/soak/stop-on-
failure controls; durable leases and duplicate-dispatch protection;
restart/crash recovery, abort/retry/reconcile, concurrent-campaign exclusion;
WP-8A audit events; protected APIs; Fleet Update UI; OpenAPI/generated types;
i18n; SQLite/PostgreSQL migration coverage; and feature-disabled no-op
behavior. Health convergence requires fresh heartbeat, `PanelVersion`, and
node/Xray health. Rollback is reported only from explicit node evidence; node-
local rollback remains the existing single-node updater responsibility.

## Fixed decisions

- Self-service authentication is completely separate from admin, API, node,
  and monitor authentication.
- Subscription credentials and subscription URLs are never portal credentials.
- Administrators create cryptographically random 256-bit portal access tokens.
  The token is shown exactly once at creation/rotation and only its hash is
  persisted. The token is used only to establish a dedicated portal session.
- Portal sessions use a separate cookie/session namespace with HttpOnly,
  Secure, SameSite, fixation prevention, CSRF/origin protection, and server-
  side client identity resolution.
- Token rotation/revocation increments credential version and immediately
  invalidates all existing portal sessions.
- Self-service v1 permits only: own profile/subscription status; own assigned
  inbounds and sanitized connection/host information; own traffic, quota and
  expiry; own HWID/device listing; rename own device; revoke own device through
  an existing ownership-safe authoritative service; and rotate/revoke own
  portal access.
- Self-service cannot reset subscription/client credentials, edit policies or
  inbounds, CRUD hosts, change groups, control nodes/runtime, delete arbitrary
  history, or access admin settings.
- Fleet dashboard is admin-only. Self-service users never receive node/fleet
  health or runtime state.
- Default host projection is existing
  `ClientRecord → ClientInbound → Inbound → enabled/non-hidden Host`, exposing
  only connection-required allowlisted fields. Never expose private node
  addresses, credentials, certificate material/pins, raw TLS/runtime config, or
  admin metadata.
- Explicit positive visibility grants may map client/group subjects to existing
  hosts. Client groups and host groups are never implicitly equivalent.
- Extend WP-8A audit actors with `actor_type = client_portal`, stable
  `ClientRecord` identity, and a sanitized display snapshot only. Never store or
  emit portal token material.
- Prefer existing models/services. Do not create duplicate client, device,
  host, group, node, or runtime models.
- Feature-disabled behavior must remain no-op/upstream-compatible.

## Authorized implementation scope

1. Portal credential persistence and explicit migrations.
2. Explicit client/group-to-existing-host visibility grants where required.
3. Dedicated portal authentication/session middleware.
4. Rate limits for login, rotation, and device mutation.
5. `/portal/me`, devices, hosts, traffic/status, and safe mutation APIs.
6. Ownership-safe adapters over existing models/services.
7. Admin portal-access management.
8. Admin host-visibility grant management.
9. Admin fleet dashboard extensions using existing node/runtime/heartbeat data.
10. Portal frontend pages.
11. OpenAPI/generated types/i18n.
12. WP-8A audit integration.
13. SQLite/PostgreSQL migration and upgrade coverage.
14. Comprehensive auth-isolation, replay/revoke/rotation, CSRF, rate-limit,
    host-field leakage, client isolation, HWID, stale/remote node, audit,
    concurrency, and feature-disabled tests.

Do not persist passwords, portal tokens, subscription credentials, cookies,
Authorization headers, request bodies, decrypted content, or webhook response
bodies.

## Required workflow

- Keep the working tree clean and use focused commits.
- Preserve upstream architecture and minimize upstream touch points.
- Run relevant tests after each implementation stage.
- Run `make verify`, `make verify-fork`, and `go test -race ./...` where
  applicable.
- Push this feature branch and run canonical Fork verification and Release
  CatX-UI CI. Fix exact failures until both are green.
- Perform a final WP-8B scope and security audit.
- Stop on the fully green pushed branch. Do not merge WP-8B or start WP-8C.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`
