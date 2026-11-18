# AGENTS.md — инструкции для AI-исполнителей

## Цель проекта

Онлайн economy-игра **BashCapital** (Bashkortostan × Capitalism Lab): Go-сервер + Godot/Android + web-превью.

**Сейчас приоритет:** этап **S5** закрыт по web (контракты); дальше **S6 payments** / **S7 APK** по `priority_order`. План: [`roadmap/STAGES.md`](./roadmap/STAGES.md).

## Источник истины

| Файл | Роль |
|------|------|
| [`roadmap/STAGES.md`](./roadmap/STAGES.md) | **Как делать** — этапы, шаги, критерии приёмки |
| [`roadmap/STATUS.json`](./roadmap/STATUS.json) | **Что в каком статусе** (`done`/`next`/…) + `maturity` |
| [`ROADMAP.md`](./ROADMAP.md) | Короткое человекочитаемое зеркало |
| Код на `main` | Факт |

**Не** изобретай параллельную TODO-систему. Не помечай `scaffold` как «готово к проду».

### maturity (обязательно честно)

- `shipped` — работает end-to-end в обычной игре  
- `mvp` — играет локально, баланс/UX ещё сырой  
- `scaffold` — каркас/API есть, поведение неполное  
- `unverified` — код есть, на устройстве/проде не проверено  

## Куда пушить

- Feature-ветка: `cursor/<short-name>-c48e`
- **Всегда сливай итог в `main`** (`git checkout main && git merge … && git push origin main`). PR можно открывать для ревью, но не оставляй готовое только в feature-ветке.
- Не force-push в `main`

## Как брать задачу (каждый раз)

1. Открой `STAGES.md` → текущий этап (сейчас **S5**/далее по STATUS)
2. Открой `STATUS.json` → `priority_order` → первый `status: "next"` без `claim`
3. Поставь `"status": "in_progress"`, `"claim": {"owner": "<agent>", "at": "<ISO-date>"}`
4. Делай **только** эту задачу (+ явные depends_on)
5. Прогони проверки из задачи / минимум:
   ```bash
   cd server && go test ./... && go build -o /tmp/bashcapital ./cmd/server
   ```
6. `status: "done"`, честные `notes` + `maturity` + `files`
7. Обнови `ROADMAP.md` (секции Сейчас / Сделано)
8. Commit + **merge/push в `main`**

Шаблон отчёта — в конце `STAGES.md`.

## Доменные инварианты (не ломать)

- Рецепт завода только своего `building_type` (мебель ≠ хлеб)
- Тип магазина ограничивает ассортимент (`data/shops.json`)
- `sell_external=false` — продажа только себе
- Закупка входов: `own` | `city` | `market` | `firm`
- Доставка: тики = f(дистанция городов), **минимум 1 тик**
- Тик: 1 реальный час = 1/24 игрового месяца (в `-dev` автотик ~30с)
- Экономика в `data/*.json`, **сервер авторитетен**
- Координаты городов — **схематичные**, не GIS-карта РБ

## Структура репо

| Путь | Роль |
|------|------|
| `server/` | Симуляция + API (`cmd/server`) |
| `data/` | Каталог, города, economy, packs, saves |
| `web/` | Mobile web UI → обычно `:8090` |
| `client/` | Godot 4 клиент |
| `scripts/` | Каталог, баланс, арт, sync roadmap |
| `roadmap/` | STAGES + STATUS |

## Быстрый smoke

```bash
cd server && go run ./cmd/server -data ../data -dev -addr :8080
# другой терминал:
cd web && python3 -m http.server 8090 --bind 127.0.0.1
```

## GitHub Issues

```bash
./scripts/sync_roadmap_to_github.sh
```

Токен агента часто read-only — тогда это делает владелец/CI.
