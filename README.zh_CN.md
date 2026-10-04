[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

# CatX-UI

CatX-UI 是一个持续维护的 [3x-ui](https://github.com/MHSanaei/3x-ui) 下游 fork，用于管理 Xray。它由 CatCodeArbelin 在独立仓库和发布渠道中维护，不是 upstream 的官方项目，也不代表获得上游维护者的认可。

稳定版 **v0.2.0** 基于 **MHSanaei/3x-ui v3.9.0**，内置 **Xray 26.9.30**。为了保持兼容性，内部 Go 模块路径 `github.com/mhsanaei/3x-ui/v3` 不变。

## CatX 能力

- **Policy** — 面向组和客户端的策略、计划、域名/类别决策、模拟和解释。
- **Insight** — Activity、流量历史、DNS Intelligence、目的地元数据、分类、保留和隐私控制。
- **Traffic Control** — 基于能力检测的配额、时间窗口、速率和 soft-throttle。
- **Operations** — 审计、webhook、指标、self-service、fleet、备份/恢复和带恢复机制的事务式更新器。

适用的模块可以独立关闭。通用 Xray 用户内核归因不受支持，因此 Traffic Control 必须明确显示 `supported`、`degraded` 或 `unsupported`，不能宣称无法证明的逐用户内核限速。

## 安装

```bash
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh)
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh) v0.2.0
```

`dev-latest` 是可选的滚动开发渠道，RC 标签仅用于测试。`x-ui update` 使用 stable，`x-ui update-dev` 使用 dev。更新器会校验 checksum；激活失败时会尝试恢复之前的可用程序、配置和数据库。

## Docker

```bash
docker run -d --name catx-ui --cap-add=NET_ADMIN --cap-add=NET_RAW \
  -v "$PWD/db:/etc/x-ui" -v "$PWD/cert:/root/cert" -p 2053:2053 \
  --restart unless-stopped ghcr.io/catcodearbelin/catx-ui:v0.2.0
```

稳定别名 `v0.2.0`、`0.2.0` 和 `latest` 支持 `linux/amd64`、`linux/arm64`、`linux/arm/v7`、`linux/arm/v6` 和 `linux/386`。默认数据库为 SQLite；PostgreSQL 使用 `XUI_DB_TYPE` 和 `XUI_DB_DSN` 配置。

## 隐私边界

CatX 只收集所需的元数据：DNS/域名、可见的 SNI、目的地 IP/端口、协议、流量、节点、inbound/outbound 和策略决策。不会使用 TLS MITM，不会解密 HTTPS，也不会收集 HTTP body、cookie、Authorization、凭据、密码或消息。风险信号是启发式的，不会依据单一信号自动封禁用户。CatX 不是计费平台，也不是通用插件生态。

## 链接

- [仓库内文档](docs/README.md)
- [稳定版 v0.2.0](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.2.0)
- [Issues](https://github.com/CatCodeArbelin/CatX-UI/issues) · [Security](SECURITY.md) · [Contributing](CONTRIBUTING.md)

CatX 保留 GPLv3 义务和上游版权/许可声明。完整信息请查看 [LICENSE](LICENSE) 和[英文 README](/README.md)。
