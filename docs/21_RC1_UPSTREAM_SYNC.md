# RC-1 Upstream Sync Record

## Pinned merge

RC-1 merges the exact upstream commit
`17d7dd46b512d0a9c22921a6094f30c672e436c9` into the CatX-UI branch based on
`develop` after WP-8C. The merge is intentionally non-rebased.

## Nontrivial conflict resolutions

- `.github/workflows/release.yml`: retained CatX-UI pull-request path gating
  and release identity/control-asset behavior while accepting upstream’s
  surrounding workflow changes. This keeps fork release checks active without
  changing the upstream build matrix semantics.
- `.github/workflows/smoke.yml`: retained the CatX smoke workflow because the
  upstream side deleted it, while the fork’s smoke coverage remains an
  independent safety boundary.
- `go.mod` and `go.sum`: retained the upstream newer `go-brrr` and `perfstat`
  versions and preserved fork-only requirements. The dependency update is
  carried through the checksums rather than pinning the older fork versions.
- `internal/web/routes_contract_test.go`: combined upstream’s isolated
  `dbtest` fixture with CatX audit-route enablement, so the test remains
  deterministic and continues to exercise the documented fork route surface.
- `internal/web/service/client_traffic.go`: retained CatX group-quota reset
  accounting and bulk lifecycle reconciliation; the reset still re-enables
  only the captured disabled clients and allows quota enforcement to veto
  unsafe re-enablement.
- `frontend/src/components/command-palette/CommandPalette.tsx`,
  `frontend/src/layouts/AppSidebar.tsx`, and `frontend/src/routes.tsx`:
  combined upstream sponsor navigation with CatX fork navigation and portal
  routes. Neither navigation surface is silently dropped.
- Generated frontend files and `frontend/public/openapi.json`: treated as
  derived artifacts. They were regenerated after resolving the merged Go
  schemas and endpoint registries, rather than preserving a stale side of the
  conflict.
- Frontend formatter compatibility: kept the fork’s existing `oxfmt` 0.68
  toolchain while accepting the upstream application dependency updates. The
  upstream formatter bump reformats the whole existing source tree; pinning
  the established formatter keeps RC-1 low-divergence and avoids an unrelated
  repository-wide formatting rewrite.
- RC-1 push validation: broadened Fork verification to `feature/*` and the
  release build workflow to `feature/rc1-*`. Release publication remains
  restricted to the existing main/tag conditions, so RC validation does not
  publish or update production artifacts.
- Concurrency hardening: reconciliation now uses the existing durable
  campaign lease fields for cross-process ownership; only the short planning
  snapshot is serialized in memory. Network-bound node executor calls no
  longer run while holding a service-wide mutex.

## Invariants checked during resolution

The merge preserves CatX updater ownership and transactional rollback,
feature-disabled behavior, the Xray external-runtime boundary, strict privacy
limits, authentication/API contracts, and modular fork registration. No TLS
MITM, decrypted HTTP-body collection, Remnawave source, release tag, `main`
update, or production update was introduced.
