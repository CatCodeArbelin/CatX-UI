[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

# CatX-UI

CatX-UI, Xray yönetimi için [3x-ui](https://github.com/MHSanaei/3x-ui) tabanlı, sürdürülen bir downstream fork'tur. CatCodeArbelin tarafından ayrı bir depoda ve release kanalında geliştirilir; upstream'in resmi projesi veya onaylı ürünü değildir.

Kararlı **v0.1.0**, **MHSanaei/3x-ui v3.8.5** tabanlıdır ve **Xray 26.9.9** içerir. Uyumluluk için dahili Go modül yolu `github.com/mhsanaei/3x-ui/v3` korunmuştur.

## CatX katmanları

- **Policy** — grup/istemci politikaları, zamanlamalar, alan adı/kategori kararları, simülasyon ve açıklama.
- **Insight** — Activity, trafik geçmişi, DNS Intelligence, hedef metadatası, sınıflandırma, retention ve gizlilik kontrolleri.
- **Traffic Control** — yetenek algılama ile kota, pencere, hız ve soft-throttle akışları.
- **Operations** — audit, webhook, metrik, self-service, fleet, backup/restore ve geri dönüşlü transactional updater.

Uygun modüller bağımsız olarak kapatılabilir. Genel Xray kullanıcı-kernel attribution desteklenmez; Traffic Control destek durumunu `supported`, `degraded` veya `unsupported` olarak açıkça belirtir.

## Kurulum

```bash
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh)
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh) v0.1.0
```

`dev-latest` rolling geliştirme kanalıdır; RC etiketleri yalnızca test içindir. `x-ui update` stable, `x-ui update-dev` dev kanalını kullanır. Güncelleyici checksum doğrular ve aktivasyon başarısız olursa önceki çalışan binary/config/database durumunu geri getirmeyi dener.

## Docker

```bash
docker run -d --name catx-ui --cap-add=NET_ADMIN --cap-add=NET_RAW \
  -v "$PWD/db:/etc/x-ui" -v "$PWD/cert:/root/cert" -p 2053:2053 \
  --restart unless-stopped ghcr.io/catcodearbelin/catx-ui:v0.1.0
```

`v0.1.0`, `0.1.0` ve `latest` etiketleri `linux/amd64`, `linux/arm64`, `linux/arm/v7`, `linux/arm/v6` ve `linux/386` için yayımlanır. Varsayılan veritabanı SQLite'tır; PostgreSQL `XUI_DB_TYPE` ve `XUI_DB_DSN` ile seçilir.

## Gizlilik sınırları

CatX yalnızca DNS/alan adı, görünür SNI, hedef IP/port, protokol, trafik, node/inbound/outbound ve politika kararları gibi metadata toplar. TLS MITM, HTTPS çözme, HTTP body, cookie, Authorization, kimlik bilgileri, parola veya mesaj toplanmaz. Risk sinyalleri sezgiseldir ve tek bir sinyalle otomatik ban uygulamaz. CatX bir billing platformu veya genel plugin ekosistemi değildir.

## Bağlantılar

- [Depodaki yerel dokümantasyon](docs/README.md)
- [Kararlı v0.1.0 release](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.1.0)
- [Issues](https://github.com/CatCodeArbelin/CatX-UI/issues) · [Security](SECURITY.md) · [Contributing](CONTRIBUTING.md)

CatX GPLv3 yükümlülüklerini ve upstream bildirimlerini korur. Ayrıntılar için [LICENSE](LICENSE) ve [İngilizce README](/README.md) dosyasına bakın.
