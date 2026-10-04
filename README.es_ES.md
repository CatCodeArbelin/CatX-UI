[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

# CatX-UI

CatX-UI es un fork downstream mantenido de [3x-ui](https://github.com/MHSanaei/3x-ui) para administrar Xray. Tiene repositorio y canal de releases propios; no es el proyecto oficial de upstream ni cuenta con su aprobación.

La versión estable **v0.2.0** está basada en **MHSanaei/3x-ui v3.9.0** e incluye **Xray 26.9.30**. El módulo Go interno `github.com/mhsanaei/3x-ui/v3` se conserva para compatibilidad.

## Capacidades CatX

- **Policy**: políticas para grupos y clientes, horarios, dominios/categorías, simulación y explicación de decisiones.
- **Insight**: Activity, historial de tráfico, DNS Intelligence, metadatos de destino, clasificación, retención y privacidad.
- **Traffic Control**: cuotas, ventanas, velocidades y soft-throttle con detección explícita de capacidades.
- **Operations**: auditoría, webhooks, métricas, self-service, fleet, backup/restore y actualizaciones transaccionales con recuperación.

Las funciones CatX se pueden desactivar de forma independiente cuando corresponde. La atribución genérica de usuarios Xray para shaping en el kernel no está soportada; Traffic Control debe indicar `supported`, `degraded` o `unsupported`.

## Instalación

```bash
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh)
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh) v0.2.0
```

`dev-latest` es un canal rolling de desarrollo y los tags RC son solo para pruebas. `x-ui update` usa stable y `x-ui update-dev` usa dev. El updater verifica checksums y, si falla la activación, intenta restaurar la versión, configuración y base de datos anteriores.

## Docker

```bash
docker run -d --name catx-ui --cap-add=NET_ADMIN --cap-add=NET_RAW \
  -v "$PWD/db:/etc/x-ui" -v "$PWD/cert:/root/cert" -p 2053:2053 \
  --restart unless-stopped ghcr.io/catcodearbelin/catx-ui:v0.2.0
```

Los alias `v0.2.0`, `0.2.0` y `latest` publican `linux/amd64`, `linux/arm64`, `linux/arm/v7`, `linux/arm/v6` y `linux/386`. SQLite es el valor predeterminado y PostgreSQL se configura con `XUI_DB_TYPE` y `XUI_DB_DSN`.

## Privacidad y límites

CatX observa metadatos permitidos como DNS/dominios, SNI visible, IP/puerto de destino, protocolo, tráfico, nodo, inbound/outbound y decisiones de política. No usa TLS MITM ni captura cuerpos HTTPS/HTTP, cookies, Authorization, credenciales, contraseñas o mensajes. Las alertas de riesgo son heurísticas y no bloquean automáticamente por una sola señal. CatX no es una plataforma de facturación ni un ecosistema genérico de plugins.

## Enlaces

- [Documentación local del repositorio](docs/README.md)
- [Release estable v0.2.0](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.2.0)
- [Issues](https://github.com/CatCodeArbelin/CatX-UI/issues) · [Security](SECURITY.md) · [Contributing](CONTRIBUTING.md)

CatX conserva las obligaciones GPLv3 y los avisos de upstream. Consulta [LICENSE](LICENSE) y el [README en inglés](/README.md) para los detalles completos.
