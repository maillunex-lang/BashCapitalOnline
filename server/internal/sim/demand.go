package sim

import (
	"math"

	"bashcapital/server/internal/model"
)

func (e *Engine) demandCfg() model.DemandConfig {
	d := e.Catalog.Economy.Demand
	// Sensible defaults if economy.json omit/partial
	if d.QualityWeight == 0 && d.BrandWeight == 0 && d.PriceWeight == 0 {
		d.ClassShares.Budget, d.ClassShares.Mid, d.ClassShares.Premium = 0.45, 0.4, 0.15
		d.QualityWeight, d.BrandWeight, d.PriceWeight, d.DistanceWeight = 0.38, 0.18, 0.28, 0.16
		d.BrandDecay, d.BrandFloor, d.BrandGainOnSale = 0.002, 0.03, 0.006
		d.TooExpensiveAt = 1.35
		d.PopDemandPerDay = map[string]float64{
			"food": 0.08, "home": 0.02, "apparel": 0.015, "fuel": 0.01, "tech": 0.008, "textile": 0.01,
		}
	}
	if d.TooExpensiveAt <= 0 {
		d.TooExpensiveAt = 1.35
	}
	if d.PopDemandPerDay == nil {
		d.PopDemandPerDay = map[string]float64{"food": 0.08, "home": 0.02}
	}
	sum := d.ClassShares.Budget + d.ClassShares.Mid + d.ClassShares.Premium
	if sum <= 0 {
		d.ClassShares.Budget, d.ClassShares.Mid, d.ClassShares.Premium = 0.45, 0.4, 0.15
	}
	return d
}

func (e *Engine) shopShelfQuality(shop *model.ShopLease) float64 {
	if shop.Quality >= 1 {
		return clamp(shop.Quality, 1, 5)
	}
	q := 3.0
	for _, b := range e.World.Buildings {
		if b.FirmID == shop.FirmID && b.Inventory[shop.ItemID] >= 0 {
			if b.Quality > 0 {
				q = b.Quality
				break
			}
		}
	}
	for _, b := range e.World.Buildings {
		if b.FirmID == shop.FirmID && b.Quality > 0 {
			return clamp(b.Quality, 1, 5)
		}
	}
	return clamp(q, 1, 5)
}

func (e *Engine) corridorMid(itemID string) float64 {
	if c, ok := e.Catalog.Economy.PriceCorridors[itemID]; ok && c.Max > 0 {
		return (c.Min + c.Max) / 2
	}
	return 100
}

// retailAttract — 0..~1.5 score from quality, brand, price vs corridor, distance.
// Class mix: budget cares more about price, premium about quality.
func (e *Engine) retailAttract(quality, brand, price, mid, distFactor float64) (attract float64, tooExpensive bool) {
	d := e.demandCfg()
	if mid <= 0 {
		mid = 100
	}
	ratio := price / mid
	if ratio > d.TooExpensiveAt {
		return 0, true
	}
	priceScore := clamp(1.0-(ratio-0.7)/0.8, 0, 1) // sweet spot near ~0.7–1.1× mid
	qScore := clamp(quality/5.0, 0, 1)
	bScore := clamp(brand, 0, 1)

	budget := d.ClassShares.Budget * (priceScore*0.55 + qScore*0.2 + bScore*0.1 + distFactor*0.15)
	midC := d.ClassShares.Mid * (priceScore*0.3 + qScore*0.35 + bScore*0.2 + distFactor*0.15)
	prem := d.ClassShares.Premium * (priceScore*0.15 + qScore*0.5 + bScore*0.25 + distFactor*0.1)
	classMix := budget + midC + prem

	weighted := d.QualityWeight*qScore + d.BrandWeight*bScore + d.PriceWeight*priceScore + d.DistanceWeight*distFactor
	attract = 0.55*classMix + 0.45*weighted
	return attract, false
}

func (e *Engine) decayBrands() {
	d := e.demandCfg()
	if d.BrandDecay <= 0 {
		return
	}
	for _, f := range e.World.Firms {
		if f.BrandAwareness == nil {
			continue
		}
		for cat, v := range f.BrandAwareness {
			nv := v * (1.0 - d.BrandDecay)
			if nv < d.BrandFloor {
				nv = d.BrandFloor
			}
			f.BrandAwareness[cat] = nv
		}
	}
}

func (e *Engine) categoryDemandPool(city *model.CityState, category string) float64 {
	d := e.demandCfg()
	perDay := d.PopDemandPerDay[category]
	if perDay <= 0 {
		perDay = 0.01
	}
	base := city.DemandSeg[category]
	hourly := float64(city.Population) * perDay / 24.0
	return base + hourly
}

func softTake(pool, attract float64, stockLeft, maxSold int) int {
	if pool <= 0 || attract <= 0 || stockLeft <= 0 || maxSold <= 0 {
		return 0
	}
	raw := pool * attract * 0.42
	take := int(math.Floor(raw + 1e-9))
	if take < 1 && raw >= 0.35 {
		take = 1
	}
	if take > stockLeft {
		take = stockLeft
	}
	if take > maxSold {
		take = maxSold
	}
	return take
}
