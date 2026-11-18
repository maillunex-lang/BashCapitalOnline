#!/usr/bin/env python3
"""
Bottom-up balance for full catalog — Bashkortostan 2026 realism.
Targets (after labor+energy at mid corridor):
  extract/raw ops: 12–28%
  intermediate: 10–22%
  finished retail-gate: 15–30%
Corridors wide enough for quality/brand but mid is cost-based.
"""
from __future__ import annotations

import json
from collections import defaultdict, deque
from pathlib import Path

ROOT = Path("/workspace/data")

# Hourly tick wage proxies (₽) — roughly 1/170 of monthly Bashkortostan 2026 wages.
WAGE = {"worker": 18.0, "engineer": 32.0, "manager": 48.0}
ENERGY_EXTRACT = 10.0
ENERGY_FACTORY = 18.0
MARGIN = {"raw": 0.18, "intermediate": 0.16, "finished": 0.22}

# Bashkortostan-2026-ish mid anchors for base raw (₽ / game unit).
RAW_MID = {
    "wheat": 14, "rye": 13, "barley": 12, "oats": 11, "buckwheat": 22,
    "sunflower": 28, "rapeseed": 26, "sugar_beet": 9,
    "potato": 16, "carrot": 15, "cabbage": 12, "onion": 14, "cucumber": 20, "tomato": 24, "pepper": 26,
    "flax": 20, "cotton": 24, "hemp": 18,
    "apple": 22, "cherry": 35, "pear": 24, "plum": 23, "strawberry": 45, "raspberry": 50, "currant": 32,
    "honey": 95,
    "cattle": 140, "sheep": 90, "goat": 85, "poultry": 35, "pig": 110, "horse": 160,
    "milk": 28, "sheep_milk": 36, "goat_milk": 40, "egg": 12, "koumiss_base": 30,
    "sand": 6, "quartz": 30, "limestone": 8, "clay": 7, "gravel": 7,
    "iron_ore": 22, "copper_ore": 28, "zinc_ore": 26, "gold_ore": 180, "bauxite": 24,
    "coal": 16, "salt": 10, "oil": 55, "gas": 40, "chemicals": 48,
    "wood": 18, "firewood": 10, "birch": 17, "pine": 16, "linden": 19,
}


def labor_cost(labor: dict) -> float:
    return sum(WAGE.get(k, 0) * float(v) for k, v in (labor or {}).items())


def energy_cost(recipe: dict) -> float:
    e = float(recipe.get("energy", 0))
    if not recipe.get("inputs"):
        return e * ENERGY_EXTRACT
    return e * ENERGY_FACTORY


def corridor(mid: float, tier: str):
    if tier == "raw":
        lo, hi = mid * 0.55, mid * 1.45
    elif tier == "intermediate":
        lo, hi = mid * 0.6, mid * 1.4
    else:
        lo, hi = mid * 0.65, mid * 1.4
    return {"min": round(max(1.0, lo), 1), "max": round(max(lo + 2, hi), 1)}


def base_id(iid: str) -> str:
    for s in ("_export", "_prem", "_eco", "_std", "_ufa", "_sterli", "_beloretsk"):
        if iid.endswith(s):
            return iid[: -len(s)]
    return iid


def main():
    items = {i["id"]: i for i in json.loads((ROOT / "items.json").read_text())}
    recipes = json.loads((ROOT / "recipes.json").read_text())
    eco = json.loads((ROOT / "economy.json").read_text())
    by_id = {r["id"]: r for r in recipes}

    mid = {}
    for iid, it in items.items():
        if it["tier"] != "raw":
            continue
        b = base_id(iid)
        anchor = RAW_MID.get(iid, RAW_MID.get(b))
        if anchor is None:
            # category fallbacks Bashkortostan 2026
            cat = it.get("category", "")
            anchor = {
                "grain": 14, "veg": 16, "fruit": 28, "fiber": 20, "oilseed": 26,
                "livestock": 100, "dairy": 30, "mineral": 12, "ore": 25, "fuel": 20,
                "oil": 55, "chem": 48, "timber": 17, "food": 40,
            }.get(cat, 20)
        # grade multipliers
        if iid.endswith("_prem") or iid.endswith("_eco"):
            anchor *= 1.25
        elif iid.endswith("_export"):
            anchor *= 1.4
        elif iid.endswith("_std"):
            anchor *= 1.15
        mid[iid] = round(anchor, 2)

    produced_by = defaultdict(list)
    for r in recipes:
        produced_by[r["output"]["item_id"]].append(r["id"])

    # Ensure every extract recipe is profitable at raw mid (real ops aren't run at a loss).
    for r in recipes:
        if r.get("inputs"):
            continue
        out = r["output"]["item_id"]
        if out not in mid:
            continue
        unit_floor = (labor_cost(r.get("labor", {})) + energy_cost(r)) * (1.0 + MARGIN["raw"]) / max(1, r["output"]["qty"])
        if mid[out] < unit_floor:
            mid[out] = round(unit_floor, 2)

    # topo on items
    indeg = {i: 0 for i in items}
    graph = defaultdict(set)
    for r in recipes:
        out = r["output"]["item_id"]
        for inp in r.get("inputs", []):
            graph[inp["item_id"]].add(out)
    for a, outs in graph.items():
        for b in outs:
            if b in indeg:
                indeg[b] += 1
    q = deque([i for i, d in indeg.items() if d == 0])
    order = []
    seen = set()
    while q:
        n = q.popleft()
        if n in seen:
            continue
        seen.add(n)
        order.append(n)
        for m in graph.get(n, []):
            indeg[m] -= 1
            if indeg[m] <= 0:
                q.append(m)
    for i in items:
        if i not in seen:
            order.append(i)

    setters = []
    for iid in order:
        it = items.get(iid)
        if not it or it["tier"] == "raw":
            continue
        cands = produced_by.get(iid, [])
        if not cands:
            continue
        best = None  # max unit cost → all recipes non-negative
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
            # fallback: average of known inputs missing — skip temporarily
            continue
        unit_cost, r, inp_cost = best
        # regional finished premium
        mult = 1.0
        if iid.endswith("_ufa"):
            mult = 1.05
        elif iid.endswith("_sterli"):
            mult = 1.03
        elif iid.endswith("_beloretsk"):
            mult = 1.04
        price = unit_cost * (1.0 + MARGIN.get(it["tier"], 0.18)) * mult
        mid[iid] = round(max(price, 2.0), 2)
        setters.append({"item": iid, "recipe": r["id"], "mid": mid[iid], "unit_cost": round(unit_cost, 2)})

    # fill any remaining without mid
    for iid, it in items.items():
        if iid not in mid:
            mid[iid] = 25.0 if it["tier"] == "raw" else (60.0 if it["tier"] == "intermediate" else 120.0)

    corridors = {iid: corridor(mid[iid], items[iid]["tier"]) for iid in items}
    eco["price_corridors"] = dict(sorted(corridors.items()))
    eco["balance"] = {
        "version": 3,
        "region": "Bashkortostan",
        "year": 2026,
        "wage_per_tick": WAGE,
        "energy_cost_extract": ENERGY_EXTRACT,
        "energy_cost_factory": ENERGY_FACTORY,
        "target_margin": MARGIN,
        "notes": "Corridors from Bashkortostan-2026-ish raw anchors + recipe costs. Re-run after catalog edits.",
    }
    # city market: low-quality wholesale ~58% of mid for common raw/inter
    market_ids = [
        "wheat", "potato", "milk", "wood", "sand", "flax", "iron_ore", "coal",
        "bottle", "chemicals", "plastic", "packaging", "flour_wheat", "paper", "oil", "gas",
    ]
    eco["city_market"] = []
    for iid in market_ids:
        if iid not in mid:
            continue
        eco["city_market"].append({
            "item_id": iid,
            "price": round(mid[iid] * 0.58, 1),
            "quality": 1.4,
            "stock_per_tick": 40 if items[iid]["tier"] == "raw" else 25,
        })

    (ROOT / "economy.json").write_text(json.dumps(eco, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    # verify margins
    rows = []
    for r in recipes:
        out = r["output"]
        if out["item_id"] not in mid:
            continue
        rev = mid[out["item_id"]] * out["qty"]
        cost = sum(mid.get(i["item_id"], 0) * i["qty"] for i in r.get("inputs", []))
        cost += labor_cost(r.get("labor", {})) + energy_cost(r)
        profit = rev - cost
        rows.append({
            "id": r["id"],
            "profit": round(profit, 1),
            "margin_pct": round(100 * profit / rev, 1) if rev else 0,
            "building": r["building_type"],
        })
    rows.sort(key=lambda x: x["profit"])
    neg = [x for x in rows if x["profit"] < -0.05]
    report = {
        "mids_sample": {k: mid[k] for k in list(sorted(mid))[:40]},
        "negative_count": len(neg),
        "negative": neg[:30],
        "worst": rows[:15],
        "best": rows[-15:],
        "median_margin_pct": sorted(x["margin_pct"] for x in rows)[len(rows) // 2] if rows else 0,
        "recipe_count": len(rows),
        "setters": len(setters),
    }
    (ROOT / "balance_report.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"corridors={len(corridors)} recipes_verified={len(rows)} negative={len(neg)} median%={report['median_margin_pct']}")
    if neg[:5]:
        print("worst neg", neg[:5])


if __name__ == "__main__":
    main()
