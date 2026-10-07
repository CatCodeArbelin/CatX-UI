# M00 UI / CSS / Product Visual Review

## Scope and decision

This record covers only:

`M00 — Core Lifecycle / Feature Settings`

The review covers the CatX Feature Settings page, its navigation entry, the
`DEVS` maturity presentation, desired-versus-active runtime state presentation,
localized copy, theme behavior, and responsive layout. M01–M11 business pages
were not opened or qualified.

The automated UI review passed. The governed workflow is now:

`WAITING_HUMAN_REVIEW`

M00 remains `DEVS`. Screenshots, browser automation, and CI cannot provide the
required maintainer Human Review approval.

## Candidate and hosted evidence

- Branch: `qualify/v0.3-m00-core`
- Candidate commit: `c61c6b0a6092928fca87102018f6f4e0b0e49cbf`
- Hosted workflow: [M00 module qualification run 37589404472](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37589404472)
- Uploaded artifact: `catx-m00-ui-review-37589404472`
- Artifact digest: `sha256:fb44b972041a43149a8ab256bd8f6daaf19781519a26693166cd8fbc9a6c0e77`

All three hosted jobs passed:

- `m00`: deterministic M00 tests, the real `restartPanel` boundary test, and
  both M00 race suites;
- `m00-frontend`: focused M00 localization/regression tests, lint, format,
  typecheck, and production frontend build;
- `m00-real-ui`: disposable real panel, Xray runtime, Playwright review, state
  transition, screenshots, metrics, and console capture.

## Product contract exercised

The real-panel review verified:

- English (`en-US`) and Russian (`ru-RU`) M00 presentation;
- expanded desktop navigation and the collapsed/mobile navigation drawer;
- viewports `1920x1080`, `1366x768`, `1024x768`, `768x768`, `430x900`, and
  `375x812`;
- dark, ultra-dark, and light theme states;
- the M00 navigation entry, title, and exactly one visible `DEVS` heading badge;
- all 10 managed feature rows and exactly one labelled switch per row;
- localized desired-state and active-runtime-state labels;
- no raw machine-state values leaked into the Russian UI;
- no horizontal overflow and controlled feature-row height variation at the
  compact and mobile widths;
- no browser console errors or page errors;
- saving a changed desired flag produces a visible restart-required state;
- after the real panel restart, the selected feature reaches `active` and is
  visibly distinct from its desired state; and
- restoration of the disposable panel's original feature flags after the
  transition test.

## Review findings and corrections

The first real-panel passes found two genuine M00 responsive defects at
`375px`: feature rows could overflow horizontally and long dependency text
could produce uncontrolled row heights. The M00-only UI correction:

- adds scoped classes and CSS to `CatxFeaturesTab`;
- allows metadata and dependency messages to wrap;
- stacks the row action area at mobile widths; and
- preserves the feature-row metadata width so the desktop and compact layouts
  remain unchanged.

The final responsive evidence is green at both `430px` and `375px`.

An intermediate diagnostic run reported two `1366px` findings because the
review harness logged in but did not navigate that context to the M00 settings
route before inspecting it. The harness was corrected in `c61c6b0a`; the final
run navigates the compact context explicitly and passes. This was a test-flow
defect, not a product finding.

The only product files changed for this visual review are:

- `frontend/src/pages/settings/CatxFeaturesTab.tsx`
- `frontend/src/pages/settings/CatxFeaturesTab.css`

The hosted review harness and workflow changes are limited to M00 qualification
evidence and do not alter M01–M11 product behavior.

## Human Review handoff

The maintainer should review the uploaded screenshots and the M00 Feature
Settings page in the hosted artifact, with particular attention to the compact
and mobile layouts, the localized Russian copy, and the desired/active state
distinction.

Until explicit maintainer `HUMAN REVIEW PASS` is recorded, M00 must remain
`DEVS`, no module branch may merge into `develop` or `main`, and no release or
tag may be created.
