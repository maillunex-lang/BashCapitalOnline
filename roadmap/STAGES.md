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
7. Ветки: `cursor/<short-name>-c48e` → **всегда** merge/push в **`main`** после готовности.
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
| **S3** Playable loop | Сессия 30–60 мин без поломок | ✅ mvp | «Уже можно поиграть» |
| **S4** CapLab depth | Спрос/бренд/качество | ✅ mvp | Глубокая экономика |
| **S5** Contracts & exchange UI | Телефонный UI биржи | ✅ mvp | Торговля руками |
| **S6** Monetization live | Реальная подписка | 🔓 mvp | PRO/заместитель за деньги |
| **S7** Android ship | APK на устройстве | 🔓 **ТЕКУЩИЙ** | Игра на телефоне |
| **S8** Scale & polish | Postgres, load-test, geo-map, i18n | 📦 backlog | Прод-масштаб |

Приоритет текущей работы описан в [плане Android V2](RM-ANDROID-V2-VERTICAL-SLICE.md).

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

## S3 — Playable loop (готово, mvp)

**Цель:** игрок может 30–60 минут вести фирму без тупиков: ферма → завод → магазин → деньги → тики.

**Закрыто 2026-11-18:** persist до тика, аренда/автопополнение, atomic configure, кабинет собственника, playtest AI shops.

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

## S4 — CapLab depth (готово, mvp)

### S4.1 — `RM-DEMAND` ✅ mvp
**Сделано:** `economy.demand` (class shares, weights, brand decay); розница сортирует полки по quality и делит городской бюджет категории; цена vs corridor mid; лог `sold` / `stockout` / `too_expensive`; тесты `TestHigherQualitySellsMore`, `TestTooExpensiveBlocksSales`.

---

## S5 — Contracts & exchange UI (готово, mvp web-first)

### S5.1 — `RM-CONTRACTS-UI` ✅ mvp (web-first)
**Сделано:** вкладка «Опт» — форма buy/sell, ордера, «В контракт», контракты. Godot parity — нет.

---

## S6 — Live payments

### S6.1 — `RM-PAYMENT-LIVE` ✅ mvp
**Сделано:** `YOOKASSA_SHOP_ID`+`YOOKASSA_SECRET_KEY` → live Payment create; webhook gate + live API re-fetch; mock-pay только `-dev`/`ALLOW_MOCK_PAY=1`; тесты api; web кнопка оплаты.

**Приёмка:**
- [x] без секрета вне -dev — webhook 503; mock-pay 403
- [x] с webhook secret — подписка после succeeded
- [ ] реальный магазин ЮKassa в проде (нужны ключи владельца) — не unverified device, но money path не прогонялся здесь → maturity **mvp**

---

## S7 — Android

### S7.1 — `RM-APK-VERIFY`
**Сделать:** export Godot APK, поставить на устройство/эмулятор, проверить remote packs + геймплей.

**Приёмка:** чеклист в `roadmap/apk_verify.md` с датой и результатом. `maturity` у RM-ASSETS-APK → `mvp` или `shipped`.

---

## S8 — Масштабирование

| ID | Что |
|----|-----|
| `RM-MAP-INCREMENTAL` | Реальный incremental redraw по dirty tiles |
| `RM-POSTGRES` | Postgres вместо/рядом с SQLite |
| `RM-LOAD-TEST` | Локальный прогон 10, 50 и 100 игроков, включая 30 секунд со 100 игроками |
| `RM-GEO-MAP` | Карта республики, не локальные iso-сетки |
| `RM-I18N` | ba / en UI |


### S8.1 — Нагрузочный тест V2

Тест поднимает локальный HTTP-сервер со свежей временной SQLite-базой. Клиенты регистрируются, создают фирмы, читают снимок мира, оплачивают рекламу и получают события тиков по WebSocket. Тест проверяет последовательность событий, ошибки сохранения и баланс каждой фирмы после повторного открытия базы.

```bash
cd server
GOMAXPROCS=2 BASHCAPITAL_LOAD=1 go test ./internal/api -run TestV2LocalLoad -count=1 -v
GOMAXPROCS=2 BASHCAPITAL_LOAD=1 BASHCAPITAL_LOAD_TICKS=300 go test ./internal/api -run 'TestV2LocalLoad/100_players$' -count=1 -v
```

Замеры 2026-11-18 получены на Linux с Go 1.24.10 при `GOMAXPROCS=2`. В коротком прогоне со 100 клиентами p95 HTTP-запроса тика снизился с 981 до 406 мс после выноса сетевой записи из замка мира и перехода на компактный JSON снимка. В каждом коротком прогоне измерено 15 тиков. В WebSocket у медленного получателя есть собственная очередь и дедлайн записи, остальные получают события независимо.

В 30-секундном прогоне со 100 клиентами обработано по 6 766 чтений мира и игровых действий, а также 300 тиков. Все 30 000 событий WebSocket доставлены. Сохранено 7 066 снимков SQLite без ошибок, баланс каждой фирмы проверен после открытия базы. Пул SQLite зарегистрировал 1 074 ожидания общей длительностью 1,985 с. Вызов сохранения, включая сериализацию, занял в среднем 3,416 мс.

| Операция | p95 | Максимум |
| --- | ---: | ---: |
| Чтение мира | 9 мс | 191 мс |
| Оплата рекламы | 501 мс | 1 265 мс |
| HTTP-запрос тика | 437 мс | 1 077 мс |
| От запроса тика до события WebSocket | 437 мс | 1 077 мс |

Длительность и данные теста отражают нагрузку на HTTP, WebSocket и SQLite. Для оценки продакшена ещё нужна репрезентативная нагрузка с собственными фабриками игроков, реальной сетью и более длительным прогоном.

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
