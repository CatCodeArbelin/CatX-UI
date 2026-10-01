[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

# CatX-UI

CatX-UI — поддерживаемый downstream-форк [3x-ui](https://github.com/MHSanaei/3x-ui) для управления Xray. Это отдельный проект CatCodeArbelin, а не официальный проект upstream и не продукт, одобренный авторами 3x-ui.

Стабильная версия **v0.1.0** основана на **MHSanaei/3x-ui v3.8.5** и включает **Xray 26.9.9**. Внутренний Go-путь `github.com/mhsanaei/3x-ui/v3` сохранён ради совместимости.

## Что добавляет CatX

- **Policy** — политики для групп и клиентов, расписания, правила доменов и категорий, симулятор и объяснение решений.
- **Insight** — Activity, история трафика, DNS Intelligence, метаданные назначений, классификация, хранение и приватность.
- **Traffic Control** — квоты, окна, скорости и soft-throttle с явным определением возможностей.
- **Operations** — аудит, webhooks, метрики, self-service, fleet, backup/restore и транзакционный updater с восстановлением.

Модули отключены по умолчанию, где это применимо. Универсальная kernel-настройка скорости по пользователям Xray не поддерживается: при отсутствии атрибуции панель показывает `supported`, `degraded` или `unsupported`, а не заявляет непроверенное enforcement.

## Установка

```bash
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh)
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh) v0.1.0
```

`dev-latest` — отдельный rolling dev-канал, а RC-теги предназначены только для тестирования. Команды `x-ui update` и `x-ui update-dev` выбирают stable и dev соответственно. Обновление использует snapshot → validate → stage → healthcheck и пытается автоматически вернуть предыдущую рабочую версию при ошибке.

## Docker

```bash
docker run -d --name catx-ui --cap-add=NET_ADMIN --cap-add=NET_RAW \
  -v "$PWD/db:/etc/x-ui" -v "$PWD/cert:/root/cert" -p 2053:2053 \
  --restart unless-stopped ghcr.io/catcodearbelin/catx-ui:v0.1.0
```

Доступны алиасы `v0.1.0`, `0.1.0` и `latest` для `linux/amd64`, `linux/arm64`, `linux/arm/v7`, `linux/arm/v6` и `linux/386`. SQLite используется по умолчанию; PostgreSQL задаётся через `XUI_DB_TYPE` и `XUI_DB_DSN`.

## Ограничения приватности

CatX использует только метаданные: домены/DNS, SNI когда виден, IP/порт назначения, протокол, трафик, узел, inbound/outbound и решения политики. TLS MITM, расшифровка HTTPS, тела HTTP, cookies, Authorization, пароли, учётные данные и сообщения не собираются. Risk-сигналы являются эвристиками и не выполняют автоматический ban по одному признаку. CatX не является биллинговой системой или универсальной plugin-платформой.

## Документация и ссылки

- [Документация в репозитории](docs/README.md)
- [Стабильный релиз v0.1.0](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.1.0)
- [Issues](https://github.com/CatCodeArbelin/CatX-UI/issues) · [Security](SECURITY.md) · [Contributing](CONTRIBUTING.md)

CatX сохраняет GPLv3 и upstream-уведомления. См. [LICENSE](LICENSE) и [английский README](/README.md) для полной информации.
