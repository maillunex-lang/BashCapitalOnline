# BashCapital — дорожная карта

> **Для AI:** сначала [`AGENTS.md`](./AGENTS.md) → подробные этапы [`roadmap/STAGES.md`](./roadmap/STAGES.md) → статус [`roadmap/STATUS.json`](./roadmap/STATUS.json).  
> `STATUS.json` — источник истины. `STAGES.md` — как выполнять. Этот файл — краткое зеркало.

Обновлено 2026-11-19 · Ветка **`main`** · Разработка сервера V2

## Честность статусов

`done` = код есть и запускается. Это **не** значит «прод / APK проверен / live-оплата / load-test».  
Смотри поле `maturity` в STATUS: `shipped` | `mvp` | `scaffold` | `unverified`.

## Правило слияния

Готовое **всегда** сливать в **`main`** (`merge` + `git push origin main`). Feature-ветки `cursor/*-c48e` — только рабочий буфер.

## Сейчас

Android Preview 0.4.4 доступен. Полная проверка на устройстве и релизная подпись отложены. На сервере завершены локальные прогоны до 100 клиентов, включая сценарий со 100 работающими заводами. [План Android V2](roadmap/RM-ANDROID-V2-VERTICAL-SLICE.md)

## Этапы (сводка)

| Этап | Статус | Суть |
|------|--------|------|
| S0–S4 | mvp ✅ | Foundation → demand |
| S5 Contracts UI | mvp ✅ | Биржа/контракты в web |
| **S6 Payments live** | 🔓 mvp | ЮKassa create+webhook; mock gated |
| S7 Android verify | в работе | APK на устройстве |
| S8 Scale | в работе | Нагрузочный тест локально, PostgreSQL по необходимости, geo-map и i18n в бэклоге |

## Недавно сделано

| ID | maturity | Что |
|----|----------|-----|
| RM-LOGISTICS-DIST | mvp | Расчёт тарифа, склад и касса перед отправкой, отслеживание груза и история поступлений в Godot и web |
| RM-LOAD-TEST | mvp | REST, WebSocket и SQLite проверены при 100 клиентах со своими заводами |
| Android V2 · интерфейс | mvp | Склад, витрина, грузы и состояние производства проверены в Godot headless |
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

RM-POSTGRES · RM-GEO-MAP · RM-I18N

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
