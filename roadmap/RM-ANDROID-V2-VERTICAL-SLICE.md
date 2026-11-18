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

### 2. Map and three cities
- Ufa, Beloretsk, Sterlitamak; parcels, deposits, roads, factories, stores, other companies.
- Implement incremental dirty-tile redraw in Godot and web if still used for admin/debug, not a full redraw per state update.
- Validate parcel selection, construction placement, deposit rules, and server-authoritative state.
- Profile frame time, memory and draw calls on low-end Android.

### 3. One complete playable economic chain
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

## Immediate next action

Phase 2: validate the three cities and implement incremental dirty-tile rendering in Godot.
