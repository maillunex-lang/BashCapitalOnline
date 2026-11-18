# BashCapital — дорожная карта

> **Для AI:** сначала [`AGENTS.md`](./AGENTS.md) → подробные этапы [`roadmap/STAGES.md`](./roadmap/STAGES.md) → статус [`roadmap/STATUS.json`](./roadmap/STATUS.json).  
> `STATUS.json` — источник истины. `STAGES.md` — как выполнять. Этот файл — краткое зеркало.

Обновлено: 2026-11-18 · Ветка: **`main`** · Текущий этап: **S5 Contracts UI** (S3–S4 закрыты)

## Честность статусов

`done` = код есть и запускается. Это **не** значит «прод / APK проверен / live-оплата / load-test».  
Смотри поле `maturity` в STATUS: `shipped` | `mvp` | `scaffold` | `unverified`.

## Правило слияния

Готовое **всегда** сливать в **`main`** (`merge` + `git push origin main`). Feature-ветки `cursor/*-c48e` — только рабочий буфер.

## Сейчас (этап S5 → дальше S6/S7)

S3–S4 и web-contracts закрыты. Дальше по приоритету:

1. **RM-PAYMENT-LIVE** — живая ЮKassa  
2. **RM-APK-VERIFY** — APK на устройстве  
3. **RM-MAP-INCREMENTAL** — частичная перерисовка карты  

**Поиграть сейчас:** README → кабинет → стройка; вкладка «Опт» — ордера и контракты.

## Этапы (сводка)

| Этап | Статус | Суть |
|------|--------|------|
| S0 Foundation | mvp ✅ | Go + web + Godot, тики, фирмы |
| S1 Catalog & art | mvp ✅ | ~1000 товаров, цепочки, иконки, art pack v3 |
| S2 Three cities | mvp ✅ | Уфа, Белорецк, Стерлитамак |
| S3 Playable loop | mvp ✅ | Сессия 30–60 мин, кабинет, persist/restock |
| S4 Demand depth | mvp ✅ | Спрос/качество/бренд |
| **S5 Contracts UI** | 🔓 current | Биржа/контракты в web (Godot — нет) |
| S6 Payments live | next | Настоящая ЮKassa |
| S7 Android verify | next | APK на устройстве |
| S8 Scale | backlog | Postgres, load-test, geo-map, i18n |

## Недавно сделано

| ID | maturity | Что |
|----|----------|-----|
| RM-BUGFIX-PLAY | mvp | Persist до тика, snap аренды, автопополнение игрока, atomic configure, тесты |
| RM-OWNER-DESKTOP | mvp | Кабинет собственника → город → карта размещения |
| RM-DEMAND | mvp | class shares, quality/brand/corridor, contested budget, тесты |
| RM-CONTRACTS-UI | mvp | Web ордера + контракты (web-first) |
| RM-BALANCE-PLAYTEST | mvp | 24-tick AI shops stocked |
| RM-STERLITAMAK | mvp | 3-й город |
| RM-PERSIST | mvp | SQLite + Persist() на actions |
| RM-PAYMENT | scaffold | mock-pay; **не live** |
| RM-ASSETS-APK | unverified | packs в коде; APK на девайсе нет |

## Backlog

RM-MAP-INCREMENTAL · RM-POSTGRES · RM-LOAD-TEST · RM-GEO-MAP · RM-I18N

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
