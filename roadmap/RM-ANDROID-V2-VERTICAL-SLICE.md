# RM-ANDROID-V2-VERTICAL-SLICE - playable Android economy V2

Status: IN PROGRESS (plan approved 2026-11-18). Target: real signed Android APK, not merely a report.

## Constraints
- Main branch is source of truth. No separate HTML demo. Previous temporary demo server has been stopped.
- Use authoritative data/catalog_v2: 170 items (40 raw, 55 intermediate, 75 finished), 200 recipes, 92 buildings, 14 shops.
- Start with a NEW game world; no V1 player-save migration. Do not reset databases automatically.
- Keep live VPS services of unrelated projects untouched. Any deployment must use a distinct service, directory, ports and database.
- Existing game state must persist across ordinary restarts.
- Do not claim Android verification without device evidence. Build success alone is insufficient.
- Prioritize low-end Android performance and incremental map updates.

## Execution order and acceptance gates

### 1. Catalog V2 integration (verified headless)
- Audit Godot client for all V1 hard-coded item/building/shop IDs, recipe names, icons and filters.
- Load actual server catalog version 2 and render item/building/recipe/shop lists.
- Ensure every one of 170 items, 200 recipes, 92 buildings and 14 shops can be discovered in client UI where relevant; no broken names, icons or missing mappings.
- Produce deterministic automated client/server catalog consistency checks.
- Regenerate V2 icon manifest and asset pack with safe fallback; don't ship thousands of obsolete V1 icons unnecessarily.
- Confirm Go catalog validator and tests green.
- Deliver report with actual counts, defects, screenshots or headless verification evidence.

### 2. Map and three cities (verified locally; device profiling pending)
- Ufa, Beloretsk, Sterlitamak; parcels, deposits, roads, factories, stores, other companies.
- Implement incremental dirty-tile redraw in Godot and web if still used for admin/debug, not a full redraw per state update.
- Validate parcel selection, construction placement, deposit rules, and server-authoritative state.
- Profile frame time, memory and draw calls on low-end Android.

### 3. One complete playable economic chain (verified locally)
- Account registration -> company creation -> select city/parcel -> extract -> process -> transport -> shop -> sell -> profit -> reinvest.
- Test wheat -> flour -> bread -> grocery as canonical path, plus at least one metallurgy or fuel chain.
- Include purchase of inputs, money deductions, labor/utility constraints, inventory updates, logistics, prices and sales.
- Test from zero player state without injecting resources, cash or research in the test itself except normal starter grants.
- Automated integration tests must confirm positive and negative cases and persistence after restart.

### 4. UI and graphics
- Update V2 item/building/shop icons, inventory, recipe tree, store catalog, company panels, tooltips and production progress.
- Android touch hit targets, safe areas, orientation, readable text, minimal animation stutter.
- Missing assets should have intentional fallbacks, not blank or broken sprites.

### 5. Android APK
- Export signed installable APK using existing authorized signing pipeline without exposing secrets.
- Confirm APK boots, downloads optional remote assets, connects to V2 server, survives reconnect, and supports canonical economic chain.
- Verify on actual Android device when available; otherwise label device verification as pending.
- Publish verifiable APK path/artifact, build SHA, server catalog version, smoke-test results and known defects.

## Working protocol
1. Inspect HEAD and existing ROAD/STATUS before each task.
2. Implement only current phase, run tests, commit to main, update STATUS.json and this document.
3. Explicitly distinguish done, partial, unverified. Never mark device verification as done based solely on headless export.
4. Do not expand catalog, add PostgreSQL, languages, or load testing until the vertical slice is playable.
5. Stop on errors, report blockers precisely, do not silently substitute mock gameplay.

## Phase 1 verification (2026-11-18)

- Godot 4.6.3 connected to the Go server with a fresh SQLite world. The catalog browser displayed 170 items (40 raw, 55 intermediate, 75 finished), 92 buildings, 200 recipes and 14 shops. The smoke test matched each displayed ID and name to the source catalog, checked icons and search, and verified recipe choices and shop assortments.
- The Godot rental form opened a grocery store with `wheat_bread`. The server returned the lease in world state. This also checks that the form submits catalog IDs independent of translated labels.
- The V2 pack contains 276 icon mappings, backed by 84 existing matching textures and three shared fallbacks (items, buildings and shops). The legacy icon directory is excluded from Godot import. Run `python3 scripts/build_catalog_v2_icons.py` to refresh the mappings and pack.
- The audit found V1 item/shop IDs in rental, unlocalized building and recipe labels, map sprites tied to a fixed building table, and a startup error when the new world had an empty tick log. Those paths now use the catalog or handle the empty state.
- Run `python3 scripts/validate_catalog_v2.py`, then `go test ./...` and `go build -o /tmp/bashcapital ./cmd/server` from `server/`.
- To repeat the Godot check from the repository root, start a scratch server with `/tmp/bashcapital -data "$PWD/data" -db "$(mktemp -d)/world.db" -addr 127.0.0.1:18081 -tick-hours 0`. Run `godot --headless --path client --editor --import`, then `godot --headless --path client --script res://tests/catalog_v2_smoke.gd -- http://127.0.0.1:18081`.
- Device performance, signed APK, and loading remote pack textures on Android remain pending.

## Phase 2 verification (2026-11-18)

- A fresh V2 world contained 41 parcels in Ufa and 25 each in Beloretsk and Sterlitamak. Godot checked roads, deposits, other companies' buildings, and clearing selection on city switches.
- Godot reuses one set of draw commands per parcel. An unchanged snapshot redrew zero parcels; changing one deposit at the same map revision redrew one of 25. Selecting a parcel redrew its highlight. In the smoke run, the server refused a mine with a mismatched deposit, then placed a mine on coal and confirmed its road access. The successful action redrew one of 25 parcels.
- Chromium checked that an unchanged snapshot skips painting. Selection, changes to a road, and removal of a building replayed only overlapping parcels in the damaged canvas region. The pixel comparison against a full paint allows for rounding at antialiased clip edges.
- Go tests and server build passed. For the Godot smoke, use the fresh server and import commands above, then run `godot --headless --path client --script res://tests/city_map_smoke.gd -- http://127.0.0.1:18081`. For web, serve `web/` on port 8090 and run `python3 scripts/test_web_map.py --api http://127.0.0.1:18081` with Playwright and Chromium installed.
- Frame time, memory, and GPU draw calls on a low-end Android device still need measurement.

## Phase 3 verification (2026-11-18)

- HTTP integration starts with account registration and normal starter grants, with AI competitors present. The bread firm grows wheat, mills it, bakes bread, ships to its warehouse in Sterlitamak, auto-restocks a grocery and sells there. After the pipeline fills, operating cash rises by at least 5,000 ₽ and the firm buys a new parcel.
- The fuel firm pays for crude oil, refines gasoline base, retools for AI-92, stocks a fuel station and sells the output. Tests check the refinery's inputs, output, wages and utility deductions, and reject fuel in a grocery.
- The tests reject a city purchase without a receiving building, dispatch with insufficient stock and stocking a shop before the shipment arrives. An outstanding shipment, lease, balance and inventory are checked after reopening SQLite and signing back in.
- Fresh V2 worlds keep at least one open parcel for each resource already generated when seeding AI competitors. This prevents the AI from occupying the last playable wheat deposit during initial setup.
- Godot's account and warehouse controls passed a headless smoke against the built server. The server process was also restarted with the same scratch database, and the account, firm, tick, cash and delivered wheat were restored.
- Run `go test ./... -run TestPlayableV2 -v` from `server/`. To check the Godot controls on a fresh local server, run `godot --headless --path client --script res://tests/economy_actions_smoke.gd -- http://127.0.0.1:18081` after importing the project.

## Immediate next action

Phase 4: polish the Android UI for inventory, recipes, retail and production progress, and verify touch layout.
