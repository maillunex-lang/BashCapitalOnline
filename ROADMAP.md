# BashCapital — дорожная карта

> **Для AI-исполнителей:** сначала прочитай [`AGENTS.md`](./AGENTS.md) и [`roadmap/STATUS.json`](./roadmap/STATUS.json).  
> Статус в `STATUS.json` — источник истины. Этот файл — человекочитаемое зеркало.

Обновлено: 2026-11-18 · Каноническая ветка: **`main`**

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

1. **RM-ASSETS-APK** — Remote assets + самообновление APK
2. **RM-MAP-DIRTY** — Инкрементальная перерисовка карты (dirty tiles) до конца

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
| RM-ADMIN | Админ API + `web/admin.html` (pause/tick/cities/economy) |
| RM-GODOT-MOBILE | Godot ManagePanel завода/магазина |
| RM-PERSIST | SQLite autosave мира + accounts/payments |
| RM-AUTH | Аккаунты / сессии (`X-Session-Token`) |
| RM-PAYMENT | YooKassa-ready create + webhook + mock-pay |
| RM-LOGISTICS-DIST | Доставка по расстоянию, минимум 1 тик |
| RM-BASHKIR-FLAVOR | Ҡымыҙ/ҡорот/сәк-сәк/мёд, season, rnd_tree, elevation |

## Backlog

| ID | Что | Заметки |
|----|-----|---------|
| RM-STERLITAMAK | Включить 3-й город (Стерлитамак) | Админка городов готова |
| RM-DEMAND | Спрос по сегментам качества/бренда глубже CapLab | |
| RM-CONTRACTS-UI | UI контрактов и биржи на телефоне | |
| RM-BALANCE | Баланс экономики под hourly ticks | |
| RM-I18N | Башкирский / английский UI | |

## Правило обновления

После любой значимой задачи AI **обязан**:

1. Обновить `roadmap/STATUS.json` (`status`, `updated`, `notes`, `files`)
2. Синхронизировать таблицу выше
3. Закоммитить в той же PR-ветке
4. При наличии write-токена: `./scripts/sync_roadmap_to_github.sh`
