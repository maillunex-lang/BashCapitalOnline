# BashCapital — этапы разработки (для любого AI)

> **Как пользоваться этим файлом:** возьми **один** открытый этап (`status: open` / первую задачу `next` внутри этапа). Не прыгай через этапы. После работы обнови [`STATUS.json`](./STATUS.json) и [`ROADMAP.md`](../ROADMAP.md), затем `commit` + `push` в `main`.

## Правила для любого AI (обязательно)

1. Прочитай [`AGENTS.md`](../AGENTS.md) и этот файл целиком один раз.
2. Источник истины по статусу задач — `STATUS.json`. Этот файл — **как делать**.
3. Одна сессия = один этап или одна задача этапа. Не рефакторь всё подряд.
4. Не завышай статус: `done` только если выполнены **критерии приёмки** ниже.
5. `maturity`: `shipped` | `mvp` | `scaffold` | `unverified` — пиши честно.
6. Перед коммитом:
   ```bash
   cd server && go test ./... && go build -o /tmp/bashcapital ./cmd/server
   python3 scripts/validate_chains.py   # если трогал каталог/рецепты
   ```
7. Ветки: `cursor/<short-name>-c48e` → merge/push в **`main`**, если человек просит main или это дефолт репо.
8. **Не ломай инварианты** из AGENTS.md (building_type, shop allowlists, min 1 tick delivery, сервер авторитетен).

### Как поднять игру локально (smoke)

```bash
# терминал 1 — API
rm -f data/saves/bashcapital.db   # только для чистого мира
cd server && go run ./cmd/server -data ../data -dev -addr :8080

# терминал 2 — web
cd web && python3 -m http.server 8090 --bind 127.0.0.1
# открыть http://127.0.0.1:8090/
```

Проверка: `GET /api/catalog` → items≈1000; `GET /api/world` → города `ufa`, `beloretsk`, `sterlitamak`.

---

## Карта этапов

| Этап | Цель | Статус сейчас | Игрок получает |
|------|------|---------------|----------------|
| **S0** Foundation MVP | Сервер+клиенты+тики | ✅ mvp | Можно открыть фирму и тикать |
| **S1** Catalog & art | Каталог, цепочки, иконки | ✅ mvp | Видны товары/здания |
| **S2** Three cities | Уфа/Белорецк/Стерлитамак | ✅ mvp | 3 города, логистика |
| **S3** Playable loop | Сессия 30–60 мин без поломок | 🔓 **ТЕКУЩИЙ** | «Уже можно поиграть» |
| **S4** CapLab depth | Спрос/бренд/качество | ⏳ next | Глубокая экономика |
| **S5** Contracts & exchange UI | Телефонный UI биржи | ⏳ next | Торговля руками |
| **S6** Monetization live | Реальная подписка | ⏳ next | PRO/заместитель за деньги |
| **S7** Android ship | APK на устройстве | ⏳ next | Игра на телефоне |
| **S8** Scale & polish | Postgres, load-test, geo-map, i18n | 📦 backlog | Прод-масштаб |

---

## S0 — Foundation MVP (готово)

**Цель:** авторитетный Go-сервер, web/Godot клиенты, тики, фирмы, здания, дороги.

**Уже есть:** `server/`, `web/`, `client/`, базовый sim.

**Не делать заново.** Только чинить регрессии.

---

## S1 — Catalog & art (готово, mvp)

**Цель:** CapLab-масштаб каталога + иконки.

**Команды поддержки:**
```bash
python3 scripts/gen_full_catalog.py
python3 scripts/validate_chains.py
python3 scripts/rebalance_economy.py
python3 scripts/gen_catalog_icons.py
python3 scripts/build_art_pack.py
```

**Критерий «не сломать»:** `validate_chains.py` → `holes=0`, negative margins=0 после rebalance.

---

## S2 — Three cities (готово, mvp)

**Цель:** три активных города Башкортостана.

**Факт:** enabled = `ufa`, `beloretsk`, `sterlitamak`. Остальные в `data/cities.json` с `enabled:false`.

**Координаты** — схематичные (игровая проекция), не GIS.

---

## S3 — Playable loop ⭐ ТЕКУЩИЙ ЭТАП

**Цель:** игрок может 30–60 минут вести фирму без тупиков: ферма → завод → магазин → деньги → тики.

### S3.1 Первый запуск без боли — `RM-FIRST-RUN`
**Статус задачи:** создать/вести как `next` → `done`.

**Сделать:**
1. В `README.md` — блок «Как запустить за 2 минуты» (команды выше + скрин ожидаемого UI).
2. Web: если нет `bc_firm` — понятный онбординг (название, отрасль, город по отрасли).
3. После register сразу показать карту своего города и 1 подсказку «построй ферму на залежи».
4. Dev-кнопка «Тик» уже есть — оставь; в `-dev` автотик 30с ок.

**Файлы:** `README.md`, `web/index.html`, `web/app.js`, при необходимости `web/styles.css`.

**Приёмка:**
- [ ] С нуля: поднять server+web, открыть фирму, увидеть карту/залежи/AI.
- [ ] Нет пустого экрана / молчаливых 404 на иконки ключевых товаров.

### S3.2 Живой playtest экономики — `RM-BALANCE-PLAYTEST`
**Сделать:**
1. Чистый мир, 24–48 тиков (`POST /api/tick` или `-dev`).
2. Зафиксировать в `data/playtest_notes.md`: банкротства AI, пустые магазины, отрицательный cash у игрока без причины, рецепты без входов.
3. Починить **конкретные** поломки (сид AI координат, цены магазина, стартовый cash, unlock рецептов).
4. Не «переписывать всю экономику» — точечные фиксы.

**Приёмка:**
- [ ] После 24 тиков ≥1 AI-магазин с stock>0 в Уфе.
- [ ] Игрок food-цепочки может произвести `bread_wheat` и выставить в магазин в коридоре цен.
- [ ] `validate_chains.py` всё ещё ок.

### S3.3 Заместитель и задания видны — `RM-DEPUTY-UX`
**Сделать:** во вкладке «Ещё» показать статус подписки, имя зама, список заданий + `last_note` после тиков.

**Приёмка:** hire_deputy + set_deputy_tasks → после тика видны `last_note`.

### S3.4 Карта показывает продукцию — `RM-MAP-ICONS-PLAY`
**Сделать:** на web/Godot карте залежи = `item_*`/`res_*`, магазины = иконка товара на витрине.

**Приёмка:** скрин/ручная проверка — пшеница/молоко/бензин узнаваемы на карте.

---

## S4 — CapLab depth (спрос)

### S4.1 — `RM-DEMAND`
**Сделать:**
1. Прочитай текущий `DemandSeg` в `server/internal/sim` (population×constants).
2. Добавь в `data/economy.json` параметры: class shares, quality sensitivity, brand decay.
3. В тике розницы: спрос зависит от quality товара в магазине + brand фирмы + цена vs corridor mid.
4. Логируй в tick log: sold qty, stockout, too_expensive.

**Не делать:** нейросеть, ML, внешние API.

**Приёмка:**
- [ ] Два магазина с разной quality одним item — лучше quality продаёт больше при близкой цене.
- [ ] Юнит-тест или скриптовая проверка на движке.

---

## S5 — Contracts & exchange UI

### S5.1 — `RM-CONTRACTS-UI`
**Сделать:**
1. Найди API: orders/contracts в `server/internal/api` + sim.
2. Web-вкладка «Опт»: список ордеров с иконками, кнопка создать buy/sell, принять контракт.
3. Godot: минимальный паритет или web-first (если Godot не успеваешь — явно напиши в notes).

**Приёмка:** игрок создаёт sell-ордер на `bread_wheat`, AI или город его закрывает/частично заполняет за N тиков.

---

## S6 — Live payments

### S6.1 — `RM-PAYMENT-LIVE`
**Сделать:** реальный YooKassa Payment create; webhook с проверкой `YOOKASSA_SECRET`; mock-pay только под `-dev`/env `ALLOW_MOCK_PAY=1`.

**Приёмка:** без секрета — безопасный отказ в non-dev; с секретом — подписка активируется только после webhook success.

---

## S7 — Android

### S7.1 — `RM-APK-VERIFY`
**Сделать:** export Godot APK, поставить на устройство/эмулятор, проверить remote packs + геймплей.

**Приёмка:** чеклист в `roadmap/apk_verify.md` с датой и результатом. `maturity` у RM-ASSETS-APK → `mvp` или `shipped`.

---

## S8 — Backlog (не трогать, пока S3–S5 не зелёные)

| ID | Что |
|----|-----|
| `RM-MAP-INCREMENTAL` | Реальный incremental redraw по dirty tiles |
| `RM-POSTGRES` | Postgres вместо/рядом с SQLite |
| `RM-LOAD-TEST` | Soak/concurrent players |
| `RM-GEO-MAP` | Карта республики, не локальные iso-сетки |
| `RM-I18N` | ba / en UI |

---

## Шаблон отчёта AI после задачи

```text
TASK: RM-XXXX
STAGE: S3
DONE:
- ...
NOT DONE / OUT OF SCOPE:
- ...
VERIFY:
- go test: PASS/FAIL
- validate_chains: PASS/FAIL (если релевантно)
- manual: ...
FILES:
- ...
MATURITY: mvp|shipped|scaffold|unverified
STATUS.json: updated
Pushed: main / PR url
```

## Порядок работ «хочу поиграть скорее»

1. **S3.1** First-run  
2. **S3.2** Playtest fixes  
3. **S3.3** Deputy UX  
4. **S3.4** Map icons  
5. Потом S4/S5  

Не начинай S6–S8, пока человек явно не попросит.
