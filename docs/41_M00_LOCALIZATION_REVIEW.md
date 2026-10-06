# M00 Localization Review and Correction

## Scope and evidence boundary

This record covers only:

`M00 — Core Lifecycle / Feature Settings`

Reviewed surfaces are the CatX Feature Settings page, its settings navigation
entry, the M00 title and `DEVS` maturity marker, the desired-setting and active
runtime-state presentation, dependency/restart/error messages, empty state, and
the feature names/descriptions shown inside Feature Settings.

Business pages for M01–M11, their navigation labels, and broad UI/CSS review are
outside this package.

Source language: `en-US`
Qualification translation: `ru-RU`

## Implementation trace

The visible M00 path is:

```text
AppSidebar settings submenu
  → /settings#catx-features
  → SettingsPage
  → CatxFeaturesTab
  → GET/PUT /panel/api/fork/settings/features
```

`ForkModuleTitle` adds the centralized M00 maturity marker. The backend owns
the machine values and the active/desired distinction; the frontend owns their
localized presentation.

Production M00 runtime states are exactly:

`feature_off`, `active`, `initializing`, `restart_required`, `error`.

The API flag keys and runtime-state values remain unchanged by this review.

## Pre-correction inventory

Every key below was traced to the M00 surface before edits. `OK` means the
existing copy was retained. Other classifications describe the correction made
in this work package.

| Translation key | en-US value | ru-RU value | Surface/context | Classification |
|---|---|---|---|---|
| `fork.settings.title` | CatX-UI Features | Функции CatX-UI | navigation, page title | OK |
| `fork.settings.intro` | Enable only the fork modules you need. Dependencies are checked explicitly and changes require a panel restart. | Включайте только нужные модули форка. Зависимости проверяются явно; после изменения флагов требуется перезапуск панели. | page introduction | UNCLEAR_TRANSLATION |
| `fork.settings.save` | Save feature settings | Сохранить функции | apply action | OK |
| `fork.settings.saved` | Feature settings saved. Restart the panel to apply module changes. | Настройки функций сохранены. Перезапустите панель для применения изменений. | success notification | OK |
| `fork.settings.loadFailed` | Feature settings could not be loaded. | Не удалось загрузить настройки функций. | load failure | OK |
| `fork.settings.saveFailed` | Feature settings could not be saved. | Не удалось сохранить настройки функций. | save failure | OK |
| `fork.settings.restartRequired` | Panel restart required | Требуется перезапуск панели | pending/restart alert | OK |
| `fork.settings.restartRequiredDescription` | Most CatX modules initialize at startup. Save the flags, then restart the panel before using a changed module. | Большинство модулей CatX запускаются при старте. После сохранения флагов перезапустите панель перед использованием изменённого модуля. | pending/restart explanation | UNCLEAR_TRANSLATION |
| `fork.settings.dependencyWarning` | Enable the required dependency first. Dependencies are never enabled automatically. | Сначала включите требуемую зависимость. Зависимости не включаются автоматически. | dependency alert | OK |
| `fork.settings.runtimeState` | Runtime state | Состояние среды выполнения | active-state label | UNCLEAR_TRANSLATION |
| `fork.settings.desiredState` | Desired state | Желаемое состояние | saved/desired switch label | MISSING_EN, MISSING_RU |
| `fork.settings.noFeatures` | No CatX features are available. | Функции CatX недоступны. | empty list | MISSING_EN, MISSING_RU |
| `fork.settings.unknownFeature` | Unknown feature | Неизвестная функция | defensive unknown-key fallback | MISSING_EN, MISSING_RU |
| `fork.settings.unknownFeatureDetails` | This feature is not recognized by this panel version. | Эта функция не распознана в текущей версии панели. | defensive unknown-key description | MISSING_EN, MISSING_RU |
| `fork.settings.dependencyRequired` | Enable {dependency} before enabling {feature}. | Для функции «{feature}» сначала включите «{dependency}». | dependency detail | OK |
| `fork.settings.features.analytics.name` | Analytics | Аналитика | feature name | OK |
| `fork.settings.features.analytics.details` | Collect and show metadata-only activity, DNS, session, and traffic insights. | Собирает и показывает сведения об активности, DNS, сессиях и трафике только по метаданным. | feature description | OK |
| `fork.settings.features.dnsIntelligence.name` | DNS intelligence | DNS-аналитика | feature name | OK |
| `fork.settings.features.dnsIntelligence.details` | Enrich activity with DNS observations and resolver intelligence. | Дополняет активность наблюдениями DNS и сведениями о резолвере. | feature description | OK |
| `fork.settings.features.policies.name` | Policy engine | Движок политик | feature name | OK |
| `fork.settings.features.policies.details` | Create and assign policies and preview their non-mutating decisions. | Позволяет создавать и назначать политики, а также просматривать решения без изменений. | feature description | OK |
| `fork.settings.features.trafficControl.name` | Traffic control / QoS | Управление трафиком / QoS | feature name | OK |
| `fork.settings.features.trafficControl.details` | Apply client traffic limits and fixed-window QoS controls. | Применяет лимиты трафика клиентов и управление QoS в фиксированных окнах. | feature description | OK |
| `fork.settings.features.riskIntelligence.name` | Risk intelligence | Анализ рисков | feature name | INCORRECT_SEMANTICS |
| `fork.settings.features.riskIntelligence.details` | Detect and review metadata-based security and anomaly signals. | Выявляет и помогает просматривать сигналы безопасности и аномалии по метаданным. | feature description | UNCLEAR_TRANSLATION |
| `fork.settings.features.audit.name` | Audit | Аудит | feature name | OK |
| `fork.settings.features.audit.details` | Record and review metadata-only administrative events. | Записывает и позволяет просматривать административные события только на основе метаданных. | feature description | UNCLEAR_TRANSLATION |
| `fork.settings.features.clientPortal.name` | Client portal | Клиентский портал | feature name | OK |
| `fork.settings.features.clientPortal.details` | Provide clients with controlled access to their portal and devices. | Даёт клиентам контролируемый доступ к порталу и устройствам. | feature description | UNCLEAR_TRANSLATION |
| `fork.settings.features.fleetUpdates.name` | Fleet updates | Обновление узлов | feature name | OK |
| `fork.settings.features.fleetUpdates.details` | Plan and reconcile updates across managed nodes. | Планирует и согласует обновления управляемых узлов. | feature description | OK |
| `fork.settings.features.sponsors.name` | Sponsors | Спонсоры | feature name | OK |
| `fork.settings.features.sponsors.details` | Show operator-configured sponsor metadata through authenticated CatX surfaces. | Показывает настроенные оператором метаданные спонсоров в аутентифицированных интерфейсах CatX. | feature description | UNCLEAR_TRANSLATION |
| `fork.settings.features.productionFleetUpdates.name` | Production fleet updates | Обновление узлов в рабочей среде | feature name | OK |
| `fork.settings.features.productionFleetUpdates.details` | Allow dispatching fleet updates to production nodes. | Разрешает отправку обновлений узлам в рабочей среде. | feature description | OK |
| `fork.settings.states.feature_off` | Disabled | Отключена | M00 runtime state tag | MISSING_EN, MISSING_RU |
| `fork.settings.states.active` | Active | Активна | M00 runtime state tag | MISSING_EN, MISSING_RU |
| `fork.settings.states.initializing` | Starting | Запускается | M00 runtime state tag | MISSING_EN, MISSING_RU |
| `fork.settings.states.restart_required` | Restart required | Требуется перезапуск | M00 runtime state tag | MISSING_EN, MISSING_RU |
| `fork.settings.states.error` | Activation error | Ошибка активации | M00 activation/configuration error tag | MISSING_EN, MISSING_RU |
| `fork.settings.states.unknown` | Unknown state | Неизвестное состояние | M00 defensive unknown-state fallback | MISSING_EN, MISSING_RU |
| `fork.maturity.modules.m00` | Core Lifecycle / Feature Settings | Основной жизненный цикл / Функции CatX-UI | M00 title and navigation | UNCLEAR_TRANSLATION |
| `fork.maturity.explanation` | This module is under development and has not yet completed release qualification. | Модуль находится в разработке и ещё не прошёл квалификацию для релиза. | DEVS explanation | INCORRECT_SEMANTICS |
| `fork.maturity.tooltip` | Product maturity: {explanation} | Зрелость продукта: {explanation} | DEVS tooltip | OK |
| `fork.maturity.accessible` | {module}: {maturity}. {explanation} | Модуль «{module}»: {maturity}. {explanation} | DEVS accessible label | OK |
| `DEVS` | generated from the centralized maturity registry | exact `DEVS` marker | badge | OK; must remain untranslated |

No normal M00 product copy was hardcoded in the traced components. The
`DEVS` badge is an intentional exact governance marker, while API paths,
feature keys, and machine state values are technical constants and are not
rendered as normal labels.

The backend feature-settings endpoint contains English API `msg` diagnostics,
but the M00 page requests it silently and renders localized UI messages from
the `fork.settings.*` catalog. Those diagnostics are not user-visible M00
surface copy and were not changed in this localization-only package.

## Corrections

- Reworded English and Russian copy to identify CatX modules/settings rather
  than internal “fork”/“flags” terminology.
- Added an explicit localized `Desired state` label beside each persisted
  feature switch and retained the separate localized active runtime-state
  label.
- Replaced `Unavailable`/`Недоступна` for the M00 error state with precise
  `Activation error`/`Ошибка активации`.
- Kept the shared `fork.common.states.*` catalog unchanged so later module
  state presentations are outside this package; M00 uses its scoped
  `fork.settings.states.*` labels.
- Corrected the DEVS explanation to mean implemented but not release-qualified;
  it does not imply a runtime failure or disabled feature.
- Aligned the Russian Risk feature name with the existing `Аналитика рисков`
  CatX terminology and improved several descriptions for natural product UI.
- Added a localized empty state for a missing feature list.
- Added localized safe fallbacks so unexpected feature keys or runtime-state
  values or descriptions cannot be rendered as machine identifiers.
- Kept all machine keys, enum values, runtime semantics, maturity, and
  non-M00 locales unchanged.

## Focused localization proof

The focused frontend test checks:

- complete M00 key presence in en-US and ru-RU;
- matching value types and placeholders;
- all five production M00 states have localized labels in both languages;
- the exact `DEVS` marker and correct explanation;
- no hardcoded normal M00 copy in the settings component;
- no raw feature/state machine-value rendering;
- en-US → ru-RU changes labels without changing API payload machine values or
  desired/active semantics;
- the localized empty state and desired-state label.

The existing CatX i18n contract remains the broader parity check for all
locales; this package does not require full translation of other locales.

## TEST CHANGE JUSTIFICATION

The existing `fork-maturity.test.tsx` expected:

- en-US: `This module is under development and has not yet completed release qualification.`
- ru-RU: `Модуль находится в разработке и ещё не прошёл квалификацию для релиза.`

Those expectations contradicted the frozen qualification contract: `DEVS`
means the module is implemented but has not completed release qualification;
it does not describe an unfinished implementation or runtime state. The
corrected expectation is:

- en-US: `This module is implemented but has not yet completed release qualification.`
- ru-RU: `Модуль реализован, но ещё не прошёл квалификацию перед выпуском.`

The test remains equally strict about the exact `DEVS` marker and both locale
explanations; it was corrected to encode the approved product contract rather
than weakened to accept either wording.

## Qualification result

Functional M00 regression remains governed by the existing M00-T01–T20 suite.
The localization candidate must rerun all twenty tests after the frontend
changes. M00 remains `DEVS`; UI/CSS review and Human Review are separate
stages.
