package sim

import (
	"fmt"

	"bashcapital/server/internal/model"
)

const roadCostPerTile = 1500.0

func seedCityRoads(city *model.CityState) {
	// Главный крест от центра — стартовая дорожная сеть
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.X == 0 || p.Y == 0 {
			p.HasRoad = true
			if p.Traffic < 0.5 {
				p.Traffic = 0.55
			}
		}
	}
}

// ensureRoadAccess прокладывает дорогу от (x,y) до ближайшей существующей дороги.
// Стоимость новых тайлов списывается со строителя (если free=false).
// Возвращает число новых тайлов дороги и стоимость.
func (e *Engine) ensureRoadAccess(firm *model.Firm, city *model.CityState, x, y int, free bool) (int, float64, error) {
	parcel := parcelAt(city, x, y)
	if parcel == nil {
		return 0, 0, fmt.Errorf("участок не найден")
	}
	if connectedToRoad(city, x, y) {
		if !parcel.HasRoad {
			parcel.HasRoad = true
			parcel.Traffic = maxf(parcel.Traffic, 0.45)
		}
		return 0, 0, nil
	}

	path := shortestPathToRoad(city, x, y)
	if path == nil {
		return 0, 0, fmt.Errorf("нельзя проложить дорогу до сети")
	}

	newTiles := 0
	for _, pt := range path {
		p := parcelAt(city, pt[0], pt[1])
		if p == nil {
			continue
		}
		if !p.HasRoad {
			newTiles++
		}
	}
	cost := float64(newTiles) * roadCostPerTile
	if !free {
		if firm.Cash < cost {
			return 0, 0, fmt.Errorf("не хватает средств на дорогу (%.0f ₽ за %d кл.)", cost, newTiles)
		}
		firm.Cash -= cost
	}
	for _, pt := range path {
		p := parcelAt(city, pt[0], pt[1])
		if p == nil {
			continue
		}
		p.HasRoad = true
		p.Traffic = maxf(p.Traffic, 0.5)
	}
	return newTiles, cost, nil
}

func parcelAt(city *model.CityState, x, y int) *model.Parcel {
	for i := range city.Parcels {
		if city.Parcels[i].X == x && city.Parcels[i].Y == y {
			return &city.Parcels[i]
		}
	}
	return nil
}

func connectedToRoad(city *model.CityState, x, y int) bool {
	p := parcelAt(city, x, y)
	if p != nil && p.HasRoad {
		return true
	}
	for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		n := parcelAt(city, x+d[0], y+d[1])
		if n != nil && n.HasRoad {
			return true
		}
	}
	return false
}

func shortestPathToRoad(city *model.CityState, sx, sy int) [][2]int {
	type node struct{ x, y int }
	exist := map[[2]int]bool{}
	for _, p := range city.Parcels {
		exist[[2]int{p.X, p.Y}] = true
	}
	start := node{sx, sy}
	queue := []node{start}
	prev := map[[2]int]*[2]int{}
	seen := map[[2]int]bool{ [2]int{sx, sy}: true }
	var end *[2]int

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		// цель — любая клетка, соседняя к дороге, или сама дорога (кроме старта без дороги — ищем клетку с HasRoad)
		p := parcelAt(city, cur.x, cur.y)
		if p != nil && p.HasRoad && !(cur.x == sx && cur.y == sy) {
			end = &[2]int{cur.x, cur.y}
			break
		}
		// также если сосед уже дорога — можно остановиться на текущей (проложим до cur включительно)
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := cur.x+d[0], cur.y+d[1]
			key := [2]int{nx, ny}
			if !exist[key] || seen[key] {
				continue
			}
			seen[key] = true
			prev[key] = &[2]int{cur.x, cur.y}
			nparcel := parcelAt(city, nx, ny)
			if nparcel != nil && nparcel.HasRoad {
				end = &key
				queue = nil
				break
			}
			queue = append(queue, node{nx, ny})
		}
	}
	if end == nil {
		return nil
	}
	// reconstruct path start -> end
	var rev [][2]int
	cur := *end
	for {
		rev = append(rev, cur)
		if cur[0] == sx && cur[1] == sy {
			break
		}
		p := prev[cur]
		if p == nil {
			break
		}
		cur = *p
	}
	// reverse
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}
