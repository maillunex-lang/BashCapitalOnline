# BashCapital — дорожная карта

> **Для AI-исполнителей:** сначала прочитай [`AGENTS.md`](./AGENTS.md) и [`roadmap/STATUS.json`](./roadmap/STATUS.json).  
> Статус в `STATUS.json` — источник истины. Этот файл — человекочитаемое зеркало.

Обновлено: 2026-11-18 · Каноническая ветка: **`main`**

> AI и люди: после завершения задачи **пушить в `main`**. Feature-ветки — только промежуточно.

## Сейчас (next)

1. **RM-STERLITAMAK** — Включить 3-й город
2. **RM-DEMAND** — Спрос/качество/бренд глубже CapLab
3. **RM-CONTRACTS-UI** — UI контрактов и биржи на телефоне

## Сделано (done)

| ID | Что |
|----|-----|
| RM-MVP-CORE | Go-сервер, тики, города Уфа/Белорецк, фирмы, здания, дороги |
| RM-CATALOG | Каталог items/recipes/buildings |
| RM-SHOP-SPEC / RM-FACTORY-SPEC / RM-FACTORY-OPS | Специализация и ops |
| RM-CITY-MARKET | Городской опт + коридоры |
| RM-AI-SEED / RM-AI-CHAT | AI-цепочки и чат |
| RM-MAP-SPRITES | Iso-спрайты |
| RM-WEB-MOBILE / RM-GODOT-MOBILE | Mobile manage sheets |
| RM-ROADMAP-DOCS | STATUS + AGENTS |
| RM-ADMIN | Админ API + web/admin.html |
| RM-PERSIST / RM-AUTH / RM-PAYMENT | SQLite, сессии, YooKassa stub |
| RM-LOGISTICS-DIST | Доставка по расстоянию, min 1 тик |
| RM-BASHKIR-FLAVOR | Национальные продукты, season, rnd_tree |
| RM-BALANCE | Bottom-up баланс рецептов/коридоров (`scripts/rebalance_economy.py`, `data/chains.json`) |
| RM-ASSETS-APK | Remote packs + APK check (`data/packs`, Godot RemoteAssets) |
| RM-MAP-DIRTY | Dirty tiles + map_revision на клиенте |

## Backlog

| ID | Что |
|----|-----|
| RM-I18N | Башкирский / английский UI |

## Баланс

Пересчёт после правок рецептов:

```bash
python3 scripts/rebalance_economy.py
```

Цепочки: [`data/chains.json`](./data/chains.json) · отчёт: [`data/balance_report.json`](./data/balance_report.json)

## Арт

Полный каталог иконок (товары / здания / магазины):

```bash
python3 scripts/gen_catalog_icons.py
python3 scripts/build_art_pack.py
```

Пак `base_art_v3` + атласы `items_atlas.png` / `buildings_atlas.png`.

## Правило обновления

1. Обновить `roadmap/STATUS.json`
2. Синхронизировать ROADMAP.md
3. Commit + push в `main`
