# Economy V2: fresh-world launch policy

## Decision (2026-11-18)

The project owner confirms there are no real players to migrate. **Economy V2 launches as a new world from scratch.** Do not spend development time migrating V1 player inventories, buildings, shops, contracts, research, or money. V1 migration code and mappings are optional legacy references, not a release requirement.

## Required implementation and deployment behavior

1. Treat `data/catalog_v2/` as the authoritative economy catalog: 170 items, 200 recipes, 92 buildings, 14 shop types. Do not fall back silently to V1 when V2 is required; fail startup with an actionable error if V2 is incomplete.
2. Create a fresh world using the current V2 catalog and the three enabled cities. Initialize plots, deposits, AI firms, market, treasury, shops and all starting parameters using V2 rules. No old world state or IDs should be loaded into the new world.
3. Provide an **explicit, one-time reset command** or deployment procedure. It must locate the actual persistence path from running configuration, stop the game server before touching persistence, create a timestamped backup outside the live data directory, verify the backup, then clear **only the BashCapitalOnline game-world persistence** and start the server again. Do not delete source code, catalog JSON, assets, logs, credentials, unrelated databases, Docker volumes, or other projects.
4. Make reset opt-in with an explicit operator confirmation and an unambiguous target path. **Never wipe the world automatically on every restart, deployment, or Git pull.** Do not run the reset against the live VPS merely because this document exists.
5. On first boot after reset, check that `economy_catalog_version=2`, the correct cities are enabled, there are no legacy player firms/inventories/buildings, and extraction/production/retail work in a smoke test. Preserve normal persistence across subsequent restarts.
6. Update the deployment/runbook and ROAD/STATUS to mark V1 migration as intentionally skipped in favor of a fresh launch. Keep V1 snapshots for reference and rollback until V2 is validated.

## Acceptance criteria

- Clean start works without any V1 save or migration mapping.
- The game server loads only the V2 catalog and creates a new world.
- Restart without reset preserves the new world.
- Reset affects only this game's identified persistence path and creates a verified backup first.
- The reset is never implicit; execution on the live VPS requires a separate explicit action.
- Go tests, catalog validator and end-to-end production smoke tests pass.

**Status:** Decision recorded; this document does not itself execute a reset or prove that the deployment scripts implement the procedure. Implementation and live reset remain separate tasks.
