#!/usr/bin/env python3
"""Validate resource → recipe → shop completeness; report holes."""
from __future__ import annotations

import json
from collections import defaultdict
from pathlib import Path

ROOT = Path("/workspace/data")


def main():
    items = {i["id"]: i for i in json.loads((ROOT / "items.json").read_text())}
    recipes = json.loads((ROOT / "recipes.json").read_text())
    buildings = {b["id"]: b for b in json.loads((ROOT / "buildings.json").read_text())}
    shops = json.loads((ROOT / "shops.json").read_text())

    produced = defaultdict(list)
    consumed = set()
    missing_inputs = []
    missing_buildings = []
    for r in recipes:
        out = r["output"]["item_id"]
        produced[out].append(r["id"])
        if r["building_type"] not in buildings:
            missing_buildings.append((r["id"], r["building_type"]))
        for inp in r.get("inputs", []):
            consumed.add(inp["item_id"])
            if inp["item_id"] not in items:
                missing_inputs.append((r["id"], inp["item_id"]))

    never_produced = sorted(set(items) - set(produced))
    # raw may be extract-only; finished/intermediate must be produced
    holes_unproduced = [i for i in never_produced if items[i]["tier"] != "raw"]
    consumed_missing = sorted(consumed - set(items))
    produced_missing = sorted(set(produced) - set(items))

    def shop_ok(item_id):
        it = items[item_id]
        for s in shops:
            if s.get("allowed_items") and item_id in s["allowed_items"]:
                return s["id"]
            if it["category"] in s.get("allowed_categories", []) and it["tier"] in s.get("allowed_tiers", []):
                return s["id"]
        return None

    finished = [i for i, v in items.items() if v["tier"] == "finished"]
    no_shop = [(i, items[i]["name"], items[i]["category"]) for i in finished if not shop_ok(i)]

    # orphan extract: raw never used and not sold
    unused_raw = []
    for i, v in items.items():
        if v["tier"] != "raw":
            continue
        if i not in consumed and v["category"] not in ("fruit", "veg", "dairy", "grain", "livestock"):
            # raw livestock/crops can be intermediate path only
            if i not in produced:
                unused_raw.append(i)

    report = {
        "counts": {
            "items": len(items),
            "recipes": len(recipes),
            "buildings": len(buildings),
            "shops": len(shops),
        },
        "missing_input_refs": missing_inputs[:50],
        "missing_building_refs": missing_buildings[:50],
        "unproduced_non_raw": holes_unproduced[:80],
        "consumed_not_in_items": consumed_missing[:50],
        "produced_not_in_items": produced_missing[:50],
        "finished_without_shop": no_shop[:80],
        "ok": not any([
            missing_inputs, missing_buildings, holes_unproduced,
            consumed_missing, produced_missing, no_shop,
        ]),
    }
    (ROOT / "chain_validation.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
    print("ok" if report["ok"] else "HOLES FOUND")
    print("unproduced_non_raw", len(holes_unproduced))
    print("finished_without_shop", len(no_shop))
    print("missing_inputs", len(missing_inputs))
    print("missing_buildings", len(missing_buildings))
    if no_shop[:10]:
        print("sample no_shop", no_shop[:10])
    if holes_unproduced[:10]:
        print("sample unproduced", holes_unproduced[:10])
    if missing_inputs[:10]:
        print("sample bad inputs", missing_inputs[:10])
    if missing_buildings[:10]:
        print("sample bad buildings", missing_buildings[:10])


if __name__ == "__main__":
    main()
