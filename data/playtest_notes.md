# Playtest notes (S3.2)

## 2026-11-18
### Fixed
- AI shop prices snap to corridor mid.
- Seed finished goods + higher initial shop stock.
- Retail per-tick cap (5–8) so shelves survive restock.
- AI restock reserves inventory for shops; exchange dump uses corridor prices.
- AI build fallback `findAIBuildSite` when hardcoded coords fail.
- City wholesale fallback for AI restock.

### Verify
- After 24 ticks: multiple shops with stock>0 (esp. food/fuel).
- Player can grow wheat on deposit.

### Still open
- CapLab demand/brand depth (RM-DEMAND).
- Contracts phone UI.
