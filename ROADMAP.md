# BashCapital — дорожная карта

> **Для AI-исполнителей:** сначала прочитай [`AGENTS.md`](./AGENTS.md) и [`roadmap/STATUS.json`](./roadmap/STATUS.json).  
> Статус в `STATUS.json` — источник истины. Этот файл — человекочитаемое зеркало.

Обновлено: 2026-11-18 · Каноническая ветка: **`main`** · PR: [#1](https://github.com/maillunex-lang/BashCapitalOnline/pull/1)

> AI и люди: после завершения задачи **пушить в `main`**. Feature-ветки — только промежуточно.

## Как читать статусы

| Статус | Значение |
|--------|----------|
| `done` | Сделано и смержено/в ветке PR — **не трогать**, если нет бага |
| `in_progress` | Уже взято — не параллелить без координации |
| `next` | Следующий приоритет — **бери отсюда** |
| `backlog` | Запланировано, но не сейчас |
| `blocked` | Ждёт зависимости / секретов / решения человека |

## Сейчас (next)

1. **RM-ADMIN** — Админ-портал мира (цены, рецепты, тики, города)
2. **RM-ASSETS-APK** — Remote assets + самообновление APK
3. **RM-MAP-DIRTY** — Инкрементальная перерисовка карты (dirty tiles) до конца
4. **RM-GODOT-MOBILE** — Godot UI: панели завода/магазина как в web, телефонные layout’ы

## Сделано (done)

| ID | Что |
|----|-----|
| RM-MVP-CORE | Go-сервер, тики, города Уфа/Белорецк, фирмы, здания, дороги |
| RM-CATALOG | Разнообразный каталог items/recipes/buildings как Capitalism Lab |
| RM-SHOP-SPEC | Типы магазинов + запрет чужого ассортимента |
| RM-FACTORY-SPEC | Рецепты только своего типа здания |
| RM-FACTORY-OPS | Источники закупки (own/city/market/firm), sell_external |
| RM-CITY-MARKET | Городской опт low-quality + коридоры цен |
| RM-AI-SEED | Несколько AI-цепочек (молочка, хлеб, мебель, текстиль, сталь, нефть) |
| RM-AI-CHAT | Персоны + событийный чат с лимитами |
| RM-MAP-SPRITES | Iso-спрайты зданий/залежей/тайлов |
| RM-WEB-MOBILE | Web UI: mobile tabs, manage sheet завода/магазина, иконки |
| RM-ROADMAP-DOCS | ROADMAP.md + STATUS.json + AGENTS.md + sync script |

## Backlog

| ID | Что | Заметки |
|----|-----|---------|
| RM-STERLITAMAK | Включить 3-й город (Стерлитамак) | После админки городов |
| RM-DEMAND | Спрос по сегментам качества/бренда глубже CapLab | |
| RM-CONTRACTS-UI | UI контрактов и биржи на телефоне | |
| RM-PAYMENT | Реальная оплата подписки (не mock) | Нужны секреты |
| RM-PERSIST | Сохранение мира (Postgres/SQLite) | Сейчас in-memory |
| RM-AUTH | Аккаунты игроков / сессии | |
| RM-BALANCE | Баланс экономики под hourly ticks | |
| RM-I18N | Башкирский / английский UI | |

## Правило обновления

После любой значимой задачи AI **обязан**:

1. Обновить `roadmap/STATUS.json` (`status`, `updated`, `notes`, `files`)
2. Синхронизировать таблицу выше
3. Закоммитить в той же PR-ветке
4. При наличии write-токена: `./scripts/sync_roadmap_to_github.sh`
