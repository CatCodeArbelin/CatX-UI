[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

# CatX-UI

CatX-UI یک fork پایین‌دستیِ تحت نگهداری از [3x-ui](https://github.com/MHSanaei/3x-ui) برای مدیریت Xray است. این پروژه در مخزن و کانال انتشار مستقل CatCodeArbelin نگهداری می‌شود؛ پروژهٔ رسمی upstream نیست و تأیید maintainerهای 3x-ui را ادعا نمی‌کند.

نسخهٔ پایدار **v0.1.0** بر پایهٔ **MHSanaei/3x-ui v3.8.5** است و **Xray 26.9.9** را همراه دارد. مسیر داخلی Go یعنی `github.com/mhsanaei/3x-ui/v3` برای سازگاری حفظ شده است.

## لایه‌های CatX

- **Policy** — سیاست‌های گروه و کاربر، زمان‌بندی، تصمیم‌های دامنه/دسته، شبیه‌سازی و توضیح تصمیم.
- **Insight** — Activity، تاریخچهٔ ترافیک، DNS Intelligence، فرادادهٔ مقصد، دسته‌بندی، نگهداری و حریم خصوصی.
- **Traffic Control** — سهمیه، پنجرهٔ زمانی، سرعت و soft-throttle با تشخیص قابلیت.
- **Operations** — audit، webhook، metrics، self-service، fleet، پشتیبان‌گیری/بازیابی و updater تراکنشی با recovery.

ماژول‌ها در صورت امکان مستقل خاموش می‌شوند. نسبت‌دادن عمومی کاربر Xray برای shaping هسته پشتیبانی نمی‌شود؛ بنابراین وضعیت Traffic Control به‌صورت `supported`، `degraded` یا `unsupported` نمایش داده می‌شود.

## نصب

```bash
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh)
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh) v0.1.0
```

`dev-latest` کانال rolling توسعه است و tagهای RC فقط برای آزمایش هستند. دستور `x-ui update` کانال stable و `x-ui update-dev` کانال dev را انتخاب می‌کند. updater checksum را بررسی می‌کند و در صورت شکست فعال‌سازی، بازگردانی نسخه، تنظیمات و دیتابیس قبلی را امتحان می‌کند.

## Docker

```bash
docker run -d --name catx-ui --cap-add=NET_ADMIN --cap-add=NET_RAW \
  -v "$PWD/db:/etc/x-ui" -v "$PWD/cert:/root/cert" -p 2053:2053 \
  --restart unless-stopped ghcr.io/catcodearbelin/catx-ui:v0.1.0
```

aliasهای `v0.1.0`، `0.1.0` و `latest` برای `linux/amd64`، `linux/arm64`، `linux/arm/v7`، `linux/arm/v6` و `linux/386` منتشر شده‌اند. SQLite پیش‌فرض است و PostgreSQL با `XUI_DB_TYPE` و `XUI_DB_DSN` تنظیم می‌شود.

## مرز حریم خصوصی

CatX فقط metadata لازم مانند DNS/دامنه، SNI قابل مشاهده، IP/port مقصد، پروتکل، ترافیک، node، inbound/outbound و تصمیم سیاست را جمع‌آوری می‌کند. TLS MITM، رمزگشایی HTTPS، بدنهٔ HTTP، cookie، Authorization، اعتبارنامه، رمز عبور یا پیام جمع‌آوری نمی‌شود. سیگنال‌های ریسک heuristic هستند و با یک سیگنال کاربر را خودکار ban نمی‌کنند. CatX پلتفرم billing یا اکوسیستم عمومی plugin نیست.

## پیوندها

- [مستندات داخل مخزن](docs/README.md)
- [انتشار پایدار v0.1.0](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.1.0)
- [Issues](https://github.com/CatCodeArbelin/CatX-UI/issues) · [Security](SECURITY.md) · [Contributing](CONTRIBUTING.md)

CatX تعهدات GPLv3 و اطلاعیه‌های upstream را حفظ می‌کند. برای جزئیات [LICENSE](LICENSE) و [README انگلیسی](/README.md) را ببینید.
