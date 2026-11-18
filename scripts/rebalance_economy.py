#!/usr/bin/env python3
"""
Bottom-up CapLab-style rebalance:
raw mid prices → recipe cost + labor/energy → output mid with tier margin → corridors.
Also patches a few structurally broken recipe ratios.
"""
from __future__ import annotations

import json
from collections import defaultdict, deque
from pathlib import Path

ROOT = Path("/workspace/data")
WAGE = {"worker": 12.0, "engineer": 24.0, "manager": 36.0}
ENERGY_FACTORY = 15.0
ENERGY_EXTRACT = 8.0
MARGIN = {"raw": 0.0, "intermediate": 0.22, "finished": 0.28}

# Anchor mid prices for extractable raw / base (₽ per unit at quality~3)
RAW_MID = {
    # Anchors include room for labor+energy on extract recipes (~15–30% margin).
    "wheat": 10, "potato": 9, "apple": 14, "cherry": 20, "flax": 16, "cotton": 18,
    "wood": 12, "cattle": 120, "milk": 14, "sand": 8, "quartz": 28, "iron_ore": 18,
    "copper_ore": 22, "coal": 13, "oil": 42, "gas": 34, "chemicals": 42, "honey": 50,
}

# Structural recipe fixes (id -> patch fields)
RECIPE_FIXES = {
    # electronics chain was deeply negative
    "make_electronics": {
        "inputs": [
            {"item_id": "circuit", "qty": 1},
            {"item_id": "plastic", "qty": 2},
            {"item_id": "wire", "qty": 2},
            {"item_id": "packaging", "qty": 1},
        ],
        "output": {"item_id": "electronics", "qty": 1},
    },
    "make_appliance": {
        "inputs": [
            {"item_id": "steel", "qty": 2},
            {"item_id": "circuit", "qty": 1},
            {"item_id": "plastic", "qty": 1},
            {"item_id": "wire", "qty": 1},
        ],
        "output": {"item_id": "appliance", "qty": 1},
    },
    "refine_gasoline": {
        "inputs": [{"item_id": "oil", "qty": 4}],
        "output": {"item_id": "gasoline", "qty": 4},
    },
    "refine_diesel": {
        "inputs": [{"item_id": "oil", "qty": 4}],
        "output": {"item_id": "diesel", "qty": 4},
    },
    "make_plastic": {
        "inputs": [{"item_id": "oil", "qty": 2}, {"item_id": "chemicals", "qty": 1}],
        "output": {"item_id": "plastic", "qty": 4},
    },
    "make_butter": {
        "inputs": [{"item_id": "milk", "qty": 4}, {"item_id": "packaging", "qty": 1}],
        "output": {"item_id": "butter", "qty": 2},
    },
    "make_furniture": {
        "inputs": [
            {"item_id": "fabric", "qty": 1},
            {"item_id": "particle_board", "qty": 2},
        ],
        "output": {"item_id": "furniture", "qty": 1},
    },
    "make_furniture_wood": {
        "inputs": [
            {"item_id": "planks", "qty": 3},
            {"item_id": "fabric", "qty": 1},
        ],
        "output": {"item_id": "furniture", "qty": 1},
    },
    "make_sugar": {
        "inputs": [{"item_id": "apple", "qty": 3}, {"item_id": "chemicals", "qty": 1}],
        "output": {"item_id": "sugar", "qty": 4},
    },
    "make_bottle": {
        "inputs": [{"item_id": "sand", "qty": 2}, {"item_id": "quartz", "qty": 1}],
        "output": {"item_id": "bottle", "qty": 5},
    },
    "harvest_honey": {"output": {"item_id": "honey", "qty": 3}},
    "raise_cattle": {
        "output": {"item_id": "cattle", "qty": 2},
        "labor": {"worker": 3, "engineer": 1},
        "energy": 2,
    },
    "extract_oil": {"output": {"item_id": "oil", "qty": 6}},
    "extract_gas": {"output": {"item_id": "gas", "qty": 6}},
    "extract_chemicals": {
        "output": {"item_id": "chemicals", "qty": 6},
        "labor": {"worker": 4, "engineer": 2},
        "energy": 5,
    },
    "make_circuit": {
        "inputs": [
            {"item_id": "copper", "qty": 1},
            {"item_id": "quartz", "qty": 1},
            {"item_id": "chemicals", "qty": 1},
        ],
        "output": {"item_id": "circuit", "qty": 1},
    },
}


def labor_cost(labor: dict) -> float:
    return sum(WAGE.get(k, 0) * float(v) for k, v in labor.items())


def energy_cost(recipe: dict) -> float:
    e = float(recipe.get("energy", 0))
    if not recipe.get("inputs"):
        return e * ENERGY_EXTRACT
    return e * ENERGY_FACTORY


def topo_recipes(recipes, items):
    produced_by = defaultdict(list)
    for r in recipes:
        produced_by[r["output"]["item_id"]].append(r["id"])
    # Kahn on items
    indeg = {i: 0 for i in items}
    graph = defaultdict(set)
    for r in recipes:
        out = r["output"]["item_id"]
        for inp in r.get("inputs", []):
            graph[inp["item_id"]].add(out)
    for a, outs in graph.items():
        for b in outs:
            indeg[b] = indeg.get(b, 0) + 1
    q = deque([i for i, d in indeg.items() if d == 0])
    order = []
    while q:
        n = q.popleft()
        order.append(n)
        for m in graph.get(n, []):
            indeg[m] -= 1
            if indeg[m] == 0:
                q.append(m)
    return order, produced_by


def corridor(mid: float, tier: str):
    if tier == "raw":
        lo, hi = mid * 0.45, mid * 1.55
    elif tier == "intermediate":
        lo, hi = mid * 0.55, mid * 1.5
    else:
        lo, hi = mid * 0.6, mid * 1.45
    return {"min": round(max(1.0, lo), 1), "max": round(max(lo + 2, hi), 1)}


def main():
    items = {i["id"]: i for i in json.loads((ROOT / "items.json").read_text())}
    recipes = json.loads((ROOT / "recipes.json").read_text())
    eco = json.loads((ROOT / "economy.json").read_text())

    # apply recipe fixes
    by_id = {r["id"]: r for r in recipes}
    for rid, patch in RECIPE_FIXES.items():
        if rid not in by_id:
            continue
        by_id[rid].update(patch)
    recipes = list(by_id.values())
    # keep stable order roughly by original file — rewrite sorted by building then id
    recipes.sort(key=lambda r: (r["building_type"], r["id"]))

    mid = dict(RAW_MID)
    for iid, it in items.items():
        if it["tier"] == "raw" and iid not in mid:
            mid[iid] = 15.0

    order, produced_by = topo_recipes(recipes, items)
    report = []
    for iid in order:
        it = items.get(iid)
        if not it:
            continue
        if it["tier"] == "raw" and iid in RAW_MID:
            continue
        cands = produced_by.get(iid, [])
        if not cands:
            continue
        # Price floor = costliest viable recipe so every alternate recipe stays non-negative.
        best = None  # max unit cost
        for rid in cands:
            r = by_id[rid]
            ok = True
            inp_cost = 0.0
            for inp in r.get("inputs", []):
                if inp["item_id"] not in mid:
                    ok = False
                    break
                inp_cost += mid[inp["item_id"]] * inp["qty"]
            if not ok:
                continue
            unit_cost = (inp_cost + labor_cost(r.get("labor", {})) + energy_cost(r)) / max(1, r["output"]["qty"])
            if best is None or unit_cost > best[0]:
                best = (unit_cost, r, inp_cost)
        if best is None:
            continue
        unit_cost, r, inp_cost = best
        margin = MARGIN.get(it["tier"], 0.2)
        price = unit_cost * (1.0 + margin)
        # don't collapse below raw-ish floors for finished
        mid[iid] = round(price, 2)
        report.append({
            "item": iid,
            "setter_recipe": r["id"],
            "unit_cost": round(unit_cost, 2),
            "mid": mid[iid],
            "margin": margin,
            "input_cost_batch": round(inp_cost, 2),
        })

    corridors = {iid: corridor(mid[iid], items[iid]["tier"]) for iid in items if iid in mid}
    eco["price_corridors"] = dict(sorted(corridors.items()))
    eco["balance"] = {
        "version": 2,
        "wage_per_tick": WAGE,
        "energy_cost_extract": ENERGY_EXTRACT,
        "energy_cost_factory": ENERGY_FACTORY,
        "target_margin": MARGIN,
        "notes": "Corridors derived bottom-up from recipe costs; re-run scripts/rebalance_economy.py after recipe edits.",
    }
    # city market ~ low quality wholesaler: ~55-70% of mid for raw/inter
    for offer in eco.get("city_market", []):
        m = mid.get(offer["item_id"])
        if m:
            offer["price"] = round(m * 0.62, 1)

    (ROOT / "recipes.json").write_text(json.dumps(recipes, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    (ROOT / "economy.json").write_text(json.dumps(eco, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    (ROOT / "balance_report.json").write_text(json.dumps({
        "mids": mid,
        "setters": report,
        "sample_margins": _verify(recipes, items, mid),
    }, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"rebalanced {len(corridors)} corridors, {len(RECIPE_FIXES)} recipe fixes")
    vm = _verify(recipes, items, mid)
    print("negative margins:", [x for x in vm if x["profit"] < 0][:10])
    print("top profits:", sorted(vm, key=lambda x: -x["profit"])[:5])


def _verify(recipes, items, mid):
    rows = []
    for r in recipes:
        out = r["output"]
        if out["item_id"] not in mid:
            continue
        rev = mid[out["item_id"]] * out["qty"]
        cost = sum(mid.get(i["item_id"], 0) * i["qty"] for i in r.get("inputs", []))
        cost += labor_cost(r.get("labor", {})) + energy_cost(r)
        rows.append({
            "id": r["id"],
            "profit": round(rev - cost, 1),
            "margin_pct": round(100 * (rev - cost) / rev, 1) if rev else 0,
        })
    return sorted(rows, key=lambda x: x["profit"])


if __name__ == "__main__":
    main()
