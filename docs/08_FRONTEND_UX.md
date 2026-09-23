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
