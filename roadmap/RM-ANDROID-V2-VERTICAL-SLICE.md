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

### 1. Catalog V2 integration (FIRST)
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

## Immediate next action
Audit current Godot catalog/UI against V2 and implement phase 1. Do not start by deploying or wiping any server.
