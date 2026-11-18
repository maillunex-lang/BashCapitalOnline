# BashCapital — дорожная карта

> **Для AI:** сначала [`AGENTS.md`](./AGENTS.md) → подробные этапы [`roadmap/STAGES.md`](./roadmap/STAGES.md) → статус [`roadmap/STATUS.json`](./roadmap/STATUS.json).  
> `STATUS.json` — источник истины. `STAGES.md` — как выполнять. Этот файл — краткое зеркало.

Обновлено: 2026-11-18 · Ветка: **`main`** · Текущий этап: **S3 Playable loop**

## Честность статусов

`done` = код есть и запускается. Это **не** значит «прод / APK проверен / live-оплата / load-test».  
Смотри поле `maturity` в STATUS: `shipped` | `mvp` | `scaffold` | `unverified`.

## Сейчас (этап S3 почти закрыт → играть можно)

S3.1–S3.4 сделаны (mvp). Дальше по приоритету:

1. **RM-DEMAND** — спрос/качество/бренд глубже CapLab  
2. **RM-CONTRACTS-UI** — UI биржи/контрактов на телефоне  
3. **RM-PAYMENT-LIVE** — живая ЮKassa  
4. **RM-APK-VERIFY** — APK на устройстве  

**Поиграть сейчас:** см. README «Запуск за 2 минуты».

## Этапы (сводка)

| Этап | Статус | Суть |
|------|--------|------|
| S0 Foundation | mvp ✅ | Go + web + Godot, тики, фирмы |
| S1 Catalog & art | mvp ✅ | ~1000 товаров, цепочки, иконки, art pack v3 |
| S2 Three cities | mvp ✅ | Уфа, Белорецк, Стерлитамак |
| **S3 Playable loop** | 🔓 current | Сессия 30–60 мин без тупиков |
| S4 Demand depth | next | Спрос/качество/бренд |
| S5 Contracts UI | next | Биржа/контракты на телефоне |
| S6 Payments live | next | Настоящая ЮKassa |
| S7 Android verify | next | APK на устройстве |
| S8 Scale | backlog | Postgres, load-test, geo-map, i18n |

## Уже сделано (mvp / scaffold — см. STATUS)

| ID | maturity | Что есть / чего нет |
|----|----------|---------------------|
| RM-MVP-CORE | mvp | Сервер и клиенты; не load-tested |
| RM-CATALOG / SCALE / ICONS | mvp | Генераторный каталог + иконки; не ручной CapLab-полиш |
| RM-SHOP/FACTORY-* | shipped/mvp | Специализация заводов и магазинов |
| RM-PERSIST | mvp | **Только SQLite** (Postgres = backlog) |
| RM-PAYMENT | scaffold | mock-pay + форма webhook; **не live** |
| RM-ASSETS-APK | unverified | Packs + prompt в коде; **APK на девайсе не проверен** |
| RM-MAP-DIRTY | scaffold | dirty list есть; **full redraw** всё ещё |
| RM-STERLITAMAK | mvp | 3-й город chemistry + AI |
| RM-BALANCE | mvp | Скрипт коридоров; нужен live playtest |
| RM-LOGISTICS-DIST | mvp | Min 1 tick; coords схематичные |

## Backlog

RM-MAP-INCREMENTAL · RM-POSTGRES · RM-LOAD-TEST · RM-GEO-MAP · RM-I18N

## Запуск

```bash
cd server && go run ./cmd/server -data ../data -dev -addr :8080
cd web && python3 -m http.server 8090 --bind 127.0.0.1
# http://127.0.0.1:8090/
```

Каталог/арт/баланс:

```bash
python3 scripts/validate_chains.py
python3 scripts/rebalance_economy.py
python3 scripts/gen_catalog_icons.py && python3 scripts/build_art_pack.py
```

## Правило обновления

1. Обновить `roadmap/STATUS.json` (+ `maturity`, notes)
2. При смене этапа — строку в `roadmap/STAGES.md` / этот файл
3. Commit + push в `main`
