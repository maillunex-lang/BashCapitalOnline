# AGENTS.md — инструкции для AI-исполнителей

## Цель проекта

Онлайн economy-игра **BashCapital** (Bashkortostan × Capitalism Lab): Go-сервер + Godot/Android + web-превью.

## Источник истины по задачам

1. [`roadmap/STATUS.json`](./roadmap/STATUS.json) — **машинный** статус (`done` / `in_progress` / `next` / `backlog` / `blocked`)
2. [`ROADMAP.md`](./ROADMAP.md) — человекочитаемое зеркало
3. Открытый PR / `main` — код

**Не** изобретай параллельную TODO-систему. Обновляй `STATUS.json` в том же коммите, что и код.

## Куда пушить

- Работай в feature-ветке `cursor/<name>-c48e` при необходимости, но **итоговые изменения всегда доводи до `main`** (`git checkout main && git merge … && git push origin main`), если человек просит пуш в main или это дефолт репозитория.
- Не force-push в `main`.

## Как брать задачу

1. Прочитай `priority_order` в `STATUS.json`
2. Возьми первый пункт со `status: "next"` без `claim`
3. Поставь `"status": "in_progress"`, `"claim": {"owner": "<agent-or-model>", "at": "<ISO-date>"}`
4. Реализуй **только** эту задачу (и явные зависимости)
5. По завершении: `status: "done"`, краткие `notes`, список `files`
6. Обнови таблицы в `ROADMAP.md`
7. Запусти проверки: `cd server && go build ./...`
8. Commit + push в `main` (и синхронизируй PR-ветку при наличии)

## Доменные инварианты (не ломать)

- Рецепт завода только своего `building_type` (мебель ≠ хлеб)
- Тип магазина ограничивает ассортимент (`data/shops.json`)
- `sell_external=false` — продажа только себе
- Закупка входов: `own` | `city` | `market` | `firm`
- Доставка по расстоянию городов: тики = f(дистанция), **минимум 1 тик** (даже в одном городе)
- Тик: 1 реальный час = 1/24 игрового месяца
- Данные экономики в `data/*.json`, сервер авторитетен

## Структура

| Путь | Роль |
|------|------|
| `server/` | Авторитетная симуляция + API |
| `data/` | Каталог, города, economy, shops |
| `web/` | Mobile web UI |
| `client/` | Godot 4 клиент |
| `scripts/` | Генераторы ассетов / sync roadmap |

## Синхронизация Issues на GitHub

Токен агента часто **read-only**. Владелец или CI с write-токеном:

```bash
./scripts/sync_roadmap_to_github.sh
```

Создаёт/обновляет labels `roadmap:*` и issues с маркером `<!-- roadmap-id: RM-XXX -->`.
