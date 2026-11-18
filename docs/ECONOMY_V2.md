# Economy Catalog V2

**Version:** `economy_catalog_version = 2`  
**Location:** `data/catalog_v2/` (preferred by `store.ResolveCatalogDir`)  
**Legacy snapshot:** `data/legacy_v1/`

## Counts

| Layer | Count |
|-------|------:|
| Raw resources | 40 |
| Intermediates | 55 |
| Finished goods | 75 |
| **SKU total** | **170** |
| Recipes (base + alts) | 200 |
| Extract buildings | 19 |
| Factory buildings | 68 |
| Logistics | 5 |
| **Buildings total** | **92** |
| Retail formats | 14 |

No `_std` / `_prem` SKU clones — quality is batch parameter 1–5 on buildings/shipments.

## Design rules

- Water / electricity / packaging are **utility costs** on recipes (`water_cost`, `utility_cost`), not free hidden items.
- Geological deposits **deplete**; agro/forest/apiary/orchard plots are **renewable**.
- Bitumen is a real road material (cash premium if missing).
- B2B: city market + building_market / fuel_depot allowlists for intermediates.
- `horse_meat_pack` replaces `oatmeal` among the 75 finished SKUs (ALT-06 normalization).

## Cities

| City | Role |
|------|------|
| Ufa | Oil/gas, agro, livestock, forest, bashkir specialties, tech assembly |
| Beloretsk | Ores, coal, stone, timber |
| Sterlitamak | Salt, peat, building minerals, chemistry |

## Validation

```bash
python3 scripts/validate_catalog_v2.py
cd server && go test ./internal/sim/ -run CatalogV2 -count=1
```

See also: `docs/ECONOMY_V2_MIGRATION.md`, `docs/ECONOMY_V2_AUDIT.md`.
