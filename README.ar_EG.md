[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

# CatX-UI

CatX-UI هو fork downstream تتم صيانته من [3x-ui](https://github.com/MHSanaei/3x-ui) لإدارة Xray. للمشروع مستودع وقناة إصدارات مستقلان لدى CatCodeArbelin؛ وهو ليس مشروع upstream الرسمي ولا يدّعي تأييد مطوري 3x-ui.

الإصدار المستقر **v0.2.0** مبني على **MHSanaei/3x-ui v3.9.0** ويتضمن **Xray 26.9.30**. تم الحفاظ على مسار Go الداخلي `github.com/mhsanaei/3x-ui/v3` للتوافق.

## طبقات CatX

- **Policy** — سياسات المجموعات والعملاء، الجداول، قرارات النطاق/الفئة، المحاكاة والشرح.
- **Insight** — Activity، سجلّ حركة المرور، DNS Intelligence، بيانات الوجهة، التصنيف والخصوصية.
- **Traffic Control** — الحصص والنوافذ والسرعات وsoft-throttle مع اكتشاف القدرات.
- **Operations** — التدقيق، webhooks، المقاييس، self-service، fleet، النسخ/الاستعادة والتحديث التبادلي مع الاسترجاع.

يمكن تعطيل الوحدات بشكل مستقل عند الحاجة. لا يدعم المشروع إسناد مستخدم Xray العام لتشكيل حركة المرور في النواة؛ لذلك يعرض Traffic Control حالات `supported` أو `degraded` أو `unsupported` بوضوح.

## التثبيت

```bash
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh)
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh) v0.2.0
```

`dev-latest` قناة تطوير rolling، ووسوم RC للاختبار فقط. يستخدم `x-ui update` قناة stable ويستخدم `x-ui update-dev` قناة dev. يتحقق updater من checksum ويحاول استعادة النسخة والإعدادات وقاعدة البيانات السابقة إذا فشل التفعيل.

## Docker

```bash
docker run -d --name catx-ui --cap-add=NET_ADMIN --cap-add=NET_RAW \
  -v "$PWD/db:/etc/x-ui" -v "$PWD/cert:/root/cert" -p 2053:2053 \
  --restart unless-stopped ghcr.io/catcodearbelin/catx-ui:v0.2.0
```

الوسوم `v0.2.0` و`0.2.0` و`latest` متاحة للمنصات `linux/amd64` و`linux/arm64` و`linux/arm/v7` و`linux/arm/v6` و`linux/386`. SQLite هو الافتراضي، وPostgreSQL يُضبط بواسطة `XUI_DB_TYPE` و`XUI_DB_DSN`.

## حدود الخصوصية

يجمع CatX metadata فقط مثل DNS/النطاقات، وSNI الظاهر، وIP/port الوجهة، والبروتوكول، وحركة المرور، والعقدة، وinbound/outbound، وقرارات السياسة. لا يوجد TLS MITM أو فك HTTPS أو جمع لأجسام HTTP أو cookies أو Authorization أو كلمات المرور أو الرسائل. إشارات المخاطر heuristic ولا تؤدي إلى ban تلقائي من إشارة واحدة. CatX ليس نظام فوترة ولا منظومة plugins عامة.

## الروابط

- [توثيق المستودع المحلي](docs/README.md)
- [الإصدار المستقر v0.2.0](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.2.0)
- [Issues](https://github.com/CatCodeArbelin/CatX-UI/issues) · [Security](SECURITY.md) · [Contributing](CONTRIBUTING.md)

يحافظ CatX على التزامات GPLv3 وإشعارات upstream. راجع [LICENSE](LICENSE) و[README الإنجليزي](/README.md) للتفاصيل.
