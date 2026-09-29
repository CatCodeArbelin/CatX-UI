# Frontend / UX

## Principle

The fork should still feel like 3x-ui.

Do not replace upstream navigation or introduce a separate admin-console philosophy.

## New areas

Prefer a small set:

```text
Policies
Activity
Traffic
Operations
```

## Client UI additions

Possible sections:

```text
Overview
Policy
Activity
Connections
Traffic
Devices
```

## Policy UX

Example:

```text
Policy: Child

Traffic quota      100 GB
Speed              inherited
IP limit           3

Categories
Social             BLOCK
Adult              BLOCK
Gambling           BLOCK
Streaming          ALLOW

Schedule
Mon-Fri 07:00-21:30

Overrides
YouTube            ALLOW until 22:00
```

Inherited values must look different from overrides.

## Activity UX

Default to logical services, not raw DNS noise:

```text
YouTube
14 sessions
890 MB
last seen 2 min ago
```

Expandable details can show evidence/domains.

## Confidence UX

Example:

```text
Service: Meta
Confidence: 62%
Evidence: destination ASN
```

## Policy Simulator

Example:

```text
Client      Eva-iPhone
Destination instagram.com
Protocol    TCP/443

Result      BLOCK
Policy      Child
Rule        Social Networks
Reason      Category block
```

## Explain Route

```text
client
→ policy
→ category/rule
→ upstream routing
→ outbound
→ node
```

## Dangerous actions

Require clear confirmation for:

- apply routing;
- update;
- rollback;
- quarantine;
- delete history;
- reset devices.

## i18n

Every new visible string must use upstream i18n.

No hardcoded UI strings.

## Generated API/frontend code

Regenerate through upstream mechanisms.

Do not hand-edit generated schemas as the source of truth.

## CatX frontend design-system contract

CatX pages remain visually part of 3x-ui. They reuse the existing Ant Design
components, panel layout, theme tokens, CSS variables, spacing scale, card
styles, responsive breakpoints, and loading/error/empty-state conventions.
CatX must not introduce a separate theme, UI framework, or page shell.

Grouping navigation must never hide the original upstream destination. Fork
wrappers such as `ForkAdminPageShell` are allowed only when they compose the
existing upstream `Layout`, `AppSidebar`, theme configuration, tokens, and
responsive behavior; they must not create an independent visual shell.

Client-specific operations should be reachable from client context where
practical, while global administration remains on global pages. An expected
disabled feature renders an actionable localized state, not a raw backend
error; where appropriate, its action leads to CatX feature settings. Human
visual review is required for release-facing frontend changes.

All visible CatX text, including navigation, command-palette entries, API-doc
section names, headings, labels, placeholders, alerts, confirmations, and
empty/error states, uses the `fork.*` i18n namespace. English is the source of
truth; locale parity is checked automatically. RTL locales use logical CSS
properties and a document/component direction supplied by the active locale.

Light, dark, and ultra-dark themes must remain supported. Unsupported or
degraded capabilities stay explicit in the UI. Dangerous actions require a
clear confirmation. Pages must remain usable at 375, 430, 768, and 1024+
pixel widths; tables may scroll horizontally when their data requires it.
