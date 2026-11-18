# BashCapital — дорожная карта

> **Для AI:** сначала [`AGENTS.md`](./AGENTS.md) → подробные этапы [`roadmap/STAGES.md`](./roadmap/STAGES.md) → статус [`roadmap/STATUS.json`](./roadmap/STATUS.json).  
> `STATUS.json` — источник истины. `STAGES.md` — как выполнять. Этот файл — краткое зеркало.

Обновлено: 2026-11-18 · Ветка: **`main`** · Текущий этап: **Android V2** (следующий шаг — интерфейс и графика)

## Честность статусов

`done` = код есть и запускается. Это **не** значит «прод / APK проверен / live-оплата / load-test».  
Смотри поле `maturity` в STATUS: `shipped` | `mvp` | `scaffold` | `unverified`.

## Правило слияния

Готовое **всегда** сливать в **`main`** (`merge` + `git push origin main`). Feature-ветки `cursor/*-c48e` — только рабочий буфер.

## Сейчас

[RM-ANDROID-V2-VERTICAL-SLICE](roadmap/RM-ANDROID-V2-VERTICAL-SLICE.md) в работе. Локально проверены хлебная и топливная цепочки, доставка между городами, прибыль от продажи хлеба и восстановление фирмы после перезапуска. Следующий шаг — доработать интерфейс Android для склада, рецептов, магазинов и производства. Профилирование на устройстве и подписанный APK ещё впереди.

## Этапы (сводка)

| Этап | Статус | Суть |
|------|--------|------|
| S0–S4 | mvp ✅ | Foundation → demand |
| S5 Contracts UI | mvp ✅ | Биржа/контракты в web |
| **S6 Payments live** | 🔓 mvp | ЮKassa create+webhook; mock gated |
| S7 Android verify | в работе | APK на устройстве |
| S8 Scale | backlog | Postgres, load-test, geo-map, i18n |

## Недавно сделано

| ID | maturity | Что |
|----|----------|-----|
| Android V2 · экономика | mvp | Хлеб и АИ‑92 продаются, доставка и фирма сохраняются в SQLite |
| RM-MAP-INCREMENTAL | mvp | Частичная перерисовка Godot и web, проверена стройка на подходящей залежи |
| Android V2 · каталог | mvp | 170 товаров, 200 рецептов, 92 здания и 14 типов магазинов в Godot. Проверена аренда с товаром V2 |
| RM-PAYMENT-LIVE | mvp | Live YooKassa create + verify webhook; mock-pay только -dev/ALLOW_MOCK_PAY |
| RM-CONTRACTS-UI | mvp | Web ордера + контракты |
| RM-DEMAND | mvp | class shares, quality/brand/corridor |
| RM-OWNER-DESKTOP | mvp | Кабинет собственника |
| RM-BUGFIX-PLAY | mvp | Persist/rent/restock/atomic configure |
| RM-PAYMENT | scaffold | soft-cash subscribe для -dev |
| RM-ASSETS-APK | unverified | packs в коде; APK на девайсе нет |

## Backlog

RM-POSTGRES · RM-LOAD-TEST · RM-GEO-MAP · RM-I18N

## Запуск

```bash
cd server && go run ./cmd/server -data ../data -dev -addr :8080
cd web && python3 -m http.server 8090 --bind 127.0.0.1
# http://127.0.0.1:8090/
```

## Правило обновления

1. Обновить `roadmap/STATUS.json` (+ `maturity`, notes)
2. При смене этапа — строку в `roadmap/STAGES.md` / этот файл
3. Commit + **push в `main`**
