package sim

import "bashcapital/server/internal/model"

func (e *Engine) markDirty(cityID string, x, y int, reason string) {
	e.World.MapRevision++
	e.World.DirtyTiles = append(e.World.DirtyTiles, model.DirtyTile{
		CityID: cityID, X: x, Y: y, Reason: reason,
	})
	// keep buffer small for clients
	if len(e.World.DirtyTiles) > 200 {
		e.World.DirtyTiles = e.World.DirtyTiles[len(e.World.DirtyTiles)-200:]
	}
}

func (e *Engine) clearDirty() {
	e.World.DirtyTiles = nil
}

// ClearDirty exports clear for admin/API ack.
func (e *Engine) ClearDirty() { e.clearDirty() }
