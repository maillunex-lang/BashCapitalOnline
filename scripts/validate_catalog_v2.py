#!/usr/bin/env python3
"""Validate economy catalog V2 acceptance criteria (§11)."""
from __future__ import annotations

import json
import sys
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent / "data" / "catalog_v2"


def main() -> int:
    items = {i["id"]: i for i in json.loads((ROOT / "items.json").read_text())}
    recipes = json.loads((ROOT / "recipes.json").read_text())
    buildings = {b["id"]: b for b in json.loads((ROOT / "buildings.json").read_text())}
    shops = json.loads((ROOT / "shops.json").read_text())
    cities = json.loads((ROOT / "cities.json").read_text())
    eco = json.loads((ROOT / "economy.json").read_text())

    raw = [i for i, v in items.items() if v["tier"] == "raw"]
    inter = [i for i, v in items.items() if v["tier"] == "intermediate"]
    fin = [i for i, v in items.items() if v["tier"] == "finished"]
    errors = []

    if len(raw) != 40:
        errors.append(f"raw={len(raw)} want 40")
    if len(inter) != 55:
        errors.append(f"intermediate={len(inter)} want 55")
    if len(fin) != 75:
        errors.append(f"finished={len(fin)} want 75")
    if any("_std" in i or "_prem" in i for i in items):
        errors.append("grade clone IDs present")

    produced = defaultdict(list)
    consumed = defaultdict(list)
    for r in recipes:
        if r["building_type"] not in buildings:
            errors.append(f"recipe {r['id']} bad building {r['building_type']}")
        out = r["output"]["item_id"]
        if out not in items:
            errors.append(f"recipe {r['id']} bad out {out}")
        produced[out].append(r["id"])
        for inp in r.get("inputs", []):
            if inp["item_id"] not in items:
                errors.append(f"recipe {r['id']} bad in {inp['item_id']}")
            consumed[inp["item_id"]].append(r["id"])
        if out not in eco.get("price_corridors", {}):
            errors.append(f"no corridor {out}")

    for iid in items:
        if not produced.get(iid) and items[iid]["tier"] != "raw":
            # raw must have extract
            errors.append(f"unproduced {iid}")
        if items[iid]["tier"] == "raw" and not produced.get(iid):
            errors.append(f"raw without extract {iid}")

    # intermediates must be consumed or B2B (city_market / building_market allow)
    b2b = {m["item_id"] for m in eco.get("city_market", [])}
    for s in shops:
        b2b.update(s.get("allowed_items") or [])
    for iid in inter:
        if not consumed.get(iid) and iid not in b2b:
            errors.append(f"dead intermediate {iid}")

    def shop_ok(iid: str) -> bool:
        for s in shops:
            if iid in (s.get("allowed_items") or []):
                return True
        return False

    for iid in fin:
        if not shop_ok(iid):
            errors.append(f"finished without shop {iid}")

    # deposits coverage
    covered = set()
    for c in cities:
        if c.get("enabled", True):
            covered.update(c.get("resources", []))
    missing_dep = [i for i in raw if i not in covered]
    # salt etc must be in some city — all raw should appear
    for i in missing_dep:
        errors.append(f"raw not in city resources {i}")

    # empty non-logistics buildings
    have = {r["building_type"] for r in recipes}
    for b in buildings.values():
        if b["category"] != "logistics" and b["id"] not in have:
            errors.append(f"building without recipe {b['id']}")

    report = {
        "ok": not errors,
        "counts": {
            "items": len(items), "raw": len(raw), "intermediate": len(inter), "finished": len(fin),
            "recipes": len(recipes), "buildings": len(buildings), "shops": len(shops),
        },
        "errors": errors[:80],
        "error_count": len(errors),
    }
    (ROOT / "validation.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
    print("ok" if report["ok"] else "FAIL", report["counts"], "errors", report["error_count"])
    if errors[:15]:
        print("\n".join(errors[:15]))
    return 0 if report["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
