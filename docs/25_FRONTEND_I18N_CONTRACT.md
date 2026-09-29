# CatX Frontend i18n Contract

WP-9A owns the frontend translation contract for fork-specific surfaces. The
English locale is the source of truth. Fork keys are namespaced below `fork`:

- `fork.common.*`
- `fork.activity.*`
- `fork.policy.*`
- `fork.audit.*`
- `fork.webhooks.*`
- `fork.portal.*`
- `fork.fleet.*`
- `fork.fleetUpdate.*`
- `fork.apiDocs.*`
- `fork.settings.*`

Every supported locale must contain the same CatX key tree as English. Tests
reject missing keys, empty values, interpolation placeholder drift, JSON
errors, and object/string shape changes. A non-English value equal to English
is recorded as untranslated but is not a WP-9A failure; full translation is
WP-9B scope.

Supported locales are `ar-EG`, `en-US`, `es-ES`, `fa-IR`, `id-ID`, `ja-JP`,
`pt-BR`, `ru-RU`, `tr-TR`, `uk-UA`, `vi-VN`, `zh-CN`, and `zh-TW`.

`ar-EG` and `fa-IR` set document and Ant Design direction to RTL. All other
supported locales remain LTR. Components use logical CSS properties where
layout styling is required.

Go-i18n reserves `description` as message metadata. Do not introduce
`description` as a normal sibling translation key in a Go-loaded locale tree;
use a non-reserved semantic key such as `intro`, `details`, or `helpText`.
CatX-visible strings outside `fork.*` are not exempt from semantic localization
review; integrated client surfaces such as risk UI remain in scope. The
13-locale parity and placeholder rules apply to every CatX namespace.
