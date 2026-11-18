#!/usr/bin/env python3
"""
Generate a CapLab-scale Bashkortostan catalog:
~1000 items, ~1500+ recipes, ~120 buildings, ~25 shops.
Hand-authored core chains are preserved; scale layers are template-generated.
"""
from __future__ import annotations

import json
import math
from copy import deepcopy
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent / "data"

# Realistic Bashkortostan-2026 mid wholesale-ish anchors (₽/unit, game scale).
# Used later by rebalance; also seed corridors here lightly.
CORE_KEEP = True


def load(name):
    return json.loads((ROOT / name).read_text(encoding="utf-8"))


def save(name, obj):
    (ROOT / name).write_text(json.dumps(obj, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def item(id_, name, tier, category, branch="general", tier_unlock=0, **extra):
    d = {
        "id": id_, "name": name, "tier": tier, "category": category,
        "branch": branch, "tier_unlock": tier_unlock, "icon": f"item_{id_}",
    }
    d.update(extra)
    if "icon" not in extra:
        d["icon"] = f"item_{id_}"
    return d


def building(id_, name, category, industry_tag, cost, zone="industrial", footprint=2, branch="general", tier_unlock=0):
    return {
        "id": id_, "name": name, "category": category, "industry_tag": industry_tag,
        "build_cost": cost, "zone": zone, "footprint": footprint,
        "branch": branch, "tier_unlock": tier_unlock, "icon": f"bld_{id_}",
    }


def recipe(id_, name, btype, inputs, out_id, out_qty, labor, energy=1, research=0, season=None, branch="general", tier_unlock=0):
    r = {
        "id": id_, "name": name, "building_type": btype,
        "inputs": inputs, "output": {"item_id": out_id, "qty": out_qty},
        "labor": labor, "energy": energy, "research_cost": research,
        "branch": branch, "tier_unlock": tier_unlock,
    }
    if season:
        r["season_bonus_months"] = season
    return r


def shop(id_, name, cats, tiers, rent=1.0, items=None, icon=None):
    d = {
        "id": id_, "name": name, "rent_mult": rent,
        "allowed_categories": cats, "allowed_tiers": tiers, "icon": icon or f"shop_{id_}",
    }
    if items:
        d["allowed_items"] = items
    return d


# ---------- scale families (Bashkortostan 2026 flavored) ----------
# Each family expands into variants with full extract→process→retail path.

AGRO_CROPS = [
    ("wheat", "Пшеница", "grain", [7, 8, 9]),
    ("rye", "Рожь", "grain", [8, 9]),
    ("barley", "Ячмень", "grain", [7, 8]),
    ("oats", "Овёс", "grain", [7, 8]),
    ("buckwheat", "Гречиха", "grain", [8, 9]),
    ("sunflower", "Подсолнечник", "oilseed", [8, 9]),
    ("rapeseed", "Рапс", "oilseed", [7, 8]),
    ("sugar_beet", "Сахарная свёкла", "veg", [9, 10]),
    ("potato", "Картофель", "veg", [8, 9]),
    ("carrot", "Морковь", "veg", [8, 9]),
    ("cabbage", "Капуста", "veg", [8, 9, 10]),
    ("onion", "Лук", "veg", [7, 8, 9]),
    ("cucumber", "Огурец", "veg", [6, 7, 8]),
    ("tomato", "Томат", "veg", [7, 8]),
    ("pepper", "Перец", "veg", [7, 8]),
    ("flax", "Лён", "fiber", [7, 8]),
    ("cotton", "Хлопок", "fiber", [8, 9]),
    ("hemp", "Конопля техн.", "fiber", [8, 9]),
]

FRUITS = [
    ("apple", "Яблоко", [8, 9]),
    ("cherry", "Вишня", [6, 7, 8]),
    ("pear", "Груша", [8, 9]),
    ("plum", "Слива", [8, 9]),
    ("strawberry", "Клубника", [6, 7]),
    ("raspberry", "Малина", [7, 8]),
    ("currant", "Смородина", [7, 8]),
    ("honey", "Бурзянский мёд", [6, 7, 8]),
]

LIVESTOCK = [
    ("cattle", "КРС", "milk", "Молоко", "beef", "Говядина", "leather", "Кожа"),
    ("sheep", "Овцы", "sheep_milk", "Овечье молоко", "mutton", "Баранина", "wool", "Шерсть"),
    ("goat", "Козы", "goat_milk", "Козье молоко", "goat_meat", "Козлятина", "goat_cheese_base", "Сырье для козьего сыра"),
    ("poultry", "Птица", "egg", "Яйцо", "chicken", "Курица", "feather", "Перо"),
    ("pig", "Свиньи", None, None, "pork", "Свинина", None, None),
    ("horse", "Лошади", "koumiss_base", "Основа ҡымыҙ", "horse_meat", "Конина", "horsehair", "Конский волос"),
    ("bee", "Пасека", None, None, None, None, "beeswax", "Воск"),
]

MINERALS = [
    ("sand", "Песок", "quarry"),
    ("quartz", "Кварц", "quarry"),
    ("limestone", "Известняк", "quarry"),
    ("clay", "Глина", "quarry"),
    ("gravel", "Щебень", "quarry"),
    ("iron_ore", "Железная руда", "mine"),
    ("copper_ore", "Медная руда", "mine"),
    ("zinc_ore", "Цинковая руда", "mine"),
    ("gold_ore", "Золотая руда", "mine"),
    ("bauxite", "Бокситы", "mine"),
    ("coal", "Уголь", "mine"),
    ("salt", "Соль", "mine"),
    ("oil", "Нефть", "oil_well"),
    ("gas", "Природный газ", "gas_well"),
    ("chemicals", "Базовая химия", "chem_plant"),
]

FOREST = [
    ("wood", "Деловая древесина"),
    ("firewood", "Дрова"),
    ("birch", "Берёза"),
    ("pine", "Сосна"),
    ("linden", "Липа"),
]

# Product line suffixes to reach ~1000 SKUs without holes.
GRADES = [
    ("", "", 0),
    ("_std", " стандарт", 1),
    ("_prem", " премиум", 2),
    ("_eco", " эко", 2),
    ("_export", " экспорт", 3),
]

REGIONS = [
    ("", ""),
    ("_ufa", " (Уфа)"),
    ("_sterli", " (Стерлитамак)"),
    ("_beloretsk", " (Белорецк)"),
]


def main():
    # Start from empty structured catalogs (keep cities/routes/economy/personas).
    items = {}
    recipes = {}
    buildings = {}
    shops = {}

    def add_item(it):
        items[it["id"]] = it

    def add_recipe(r):
        recipes[r["id"]] = r

    def add_building(b):
        buildings[b["id"]] = b

    def add_shop(s):
        shops[s["id"]] = s

    # ---- base logistics / packaging always available ----
    for b in [
        building("warehouse", "Склад", "logistics", "logistics", 15000, branch="retail", tier_unlock=0),
        building("garage", "Автопарк", "logistics", "logistics", 20000, branch="retail", tier_unlock=0),
        building("research_lab", "Исследовательская лаборатория", "factory", "research", 45000, branch="research", tier_unlock=0),
        building("tech_park", "Технопарк", "factory", "research", 90000, branch="research", tier_unlock=2),
    ]:
        add_building(b)

    # Shared packaging chain
    add_building(building("glass_plant", "Стекольный завод", "factory", "packaging", 35000, branch="packaging", tier_unlock=0))
    add_building(building("paper_mill", "Бумкомбинат", "factory", "packaging", 34000, branch="packaging", tier_unlock=0))
    add_building(building("plastics_plant", "Завод пластмасс", "factory", "chem", 38000, branch="oil_chem", tier_unlock=1))
    add_building(building("packaging_plant", "Упаковочный комбинат", "factory", "packaging", 30000, branch="packaging", tier_unlock=1))

    add_item(item("sand", "Песок", "raw", "mineral", "mining", 0))
    add_item(item("quartz", "Кварц", "raw", "mineral", "mining", 0))
    add_item(item("wood", "Дерево", "raw", "timber", "timber", 0))
    add_item(item("oil", "Нефть", "raw", "oil", "oil_chem", 0))
    add_item(item("chemicals", "Базовая химия", "raw", "chem", "oil_chem", 0))
    add_item(item("bottle", "Стеклянная бутылка", "intermediate", "packaging", "packaging", 0))
    add_item(item("paper", "Бумага", "intermediate", "packaging", "packaging", 0))
    add_item(item("plastic", "Пластик", "intermediate", "chem", "oil_chem", 1))
    add_item(item("packaging", "Упаковка", "intermediate", "packaging", "packaging", 0))

    add_building(building("quarry", "Карьер", "extract", "mining", 18000, "industrial", 3, "mining", 0))
    add_building(building("logging_camp", "Лесозаготовка", "extract", "timber", 14000, "rural", 2, "timber", 0))
    add_building(building("oil_well", "Нефтяная скважина", "extract", "oil", 45000, "industrial", 3, "oil_chem", 0))
    add_building(building("chem_plant", "Химзавод", "factory", "chem", 40000, branch="oil_chem", tier_unlock=0))

    add_recipe(recipe("extract_sand", "Добыча песка", "quarry", [], "sand", 10, {"worker": 3}, 2, 0, branch="mining"))
    add_recipe(recipe("extract_quartz", "Добыча кварца", "quarry", [], "quartz", 4, {"worker": 4, "engineer": 1}, 3, 0, branch="mining"))
    add_recipe(recipe("extract_wood", "Заготовка дерева", "logging_camp", [], "wood", 8, {"worker": 4}, 1, 0, season=[9, 10], branch="timber"))
    add_recipe(recipe("extract_oil", "Добыча нефти", "oil_well", [], "oil", 6, {"worker": 5, "engineer": 3}, 5, 0, branch="oil_chem"))
    add_recipe(recipe("extract_chemicals", "Базовая химия", "chem_plant", [], "chemicals", 6, {"worker": 4, "engineer": 2}, 5, 0, branch="oil_chem"))
    add_recipe(recipe("make_bottle", "Стеклянная бутылка", "glass_plant",
                      [{"item_id": "sand", "qty": 2}, {"item_id": "quartz", "qty": 1}],
                      "bottle", 5, {"worker": 4, "engineer": 2}, 5, 0, branch="packaging"))
    add_recipe(recipe("make_paper", "Бумага", "paper_mill",
                      [{"item_id": "wood", "qty": 4}, {"item_id": "chemicals", "qty": 1}],
                      "paper", 5, {"worker": 5, "engineer": 2}, 4, 0, branch="packaging"))
    add_recipe(recipe("make_plastic", "Пластик", "plastics_plant",
                      [{"item_id": "oil", "qty": 2}, {"item_id": "chemicals", "qty": 1}],
                      "plastic", 4, {"worker": 5, "engineer": 2}, 5, 100, branch="oil_chem", tier_unlock=1))
    add_recipe(recipe("make_packaging", "Упаковка", "packaging_plant",
                      [{"item_id": "plastic", "qty": 1}, {"item_id": "paper", "qty": 2}],
                      "packaging", 4, {"worker": 4, "engineer": 1}, 3, 0, branch="packaging"))
    # paper-only packaging fallback for tier 0 firms without plastic
    add_recipe(recipe("make_packaging_paper", "Бумажная упаковка", "paper_mill",
                      [{"item_id": "paper", "qty": 3}],
                      "packaging", 2, {"worker": 3}, 2, 0, branch="packaging"))

    # ---- AGRO ----
    add_building(building("farm", "Ферма", "extract", "agro", 12000, "rural", 2, "agro", 0))
    add_building(building("greenhouse", "Теплица", "extract", "agro", 22000, "rural", 2, "agro", 1))
    add_building(building("agro_complex", "Агрокомплекс", "extract", "agro", 55000, "rural", 4, "agro", 3))
    add_building(building("mill", "Мельница", "factory", "food", 20000, branch="agro", tier_unlock=0))
    add_building(building("oil_press", "Маслобойня", "factory", "food", 24000, branch="agro", tier_unlock=1))
    add_building(building("sugar_plant", "Сахарный завод", "factory", "food", 48000, branch="agro", tier_unlock=2))
    add_building(building("bakery", "Пекарня", "factory", "food", 22000, branch="food", tier_unlock=0))
    add_building(building("bakery_industrial", "Хлебокомбинат", "factory", "food", 52000, branch="food", tier_unlock=2))
    add_building(building("food_plant", "Пищевой комбинат", "factory", "food", 32000, branch="food", tier_unlock=0))
    add_building(building("cannery", "Консервный завод", "factory", "food", 40000, branch="food", tier_unlock=1))
    add_building(building("snack_factory", "Снековый завод", "factory", "food", 36000, branch="food", tier_unlock=2))
    add_building(building("juice_plant", "Соковый завод", "factory", "food", 34000, branch="food", tier_unlock=1))
    add_building(building("confectionery", "Кондитерская фабрика", "factory", "food", 45000, branch="food", tier_unlock=2))

    for cid, cname, cat, season in AGRO_CROPS:
        for g_suf, g_name, g_tier in GRADES[:3]:  # "", _std, _prem
            iid = f"{cid}{g_suf}"
            add_item(item(iid, f"{cname}{g_name}", "raw", cat, "agro", g_tier))
            btype = "greenhouse" if g_tier >= 1 and cat == "veg" else "farm"
            if g_tier >= 3:
                btype = "agro_complex"
            rid = f"grow_{iid}"
            qty = 10 if cat in ("grain", "veg") else 6
            add_recipe(recipe(rid, f"Выращивание: {cname}{g_name}", btype, [], iid, qty,
                              {"worker": 3 + g_tier}, 1 + g_tier, 100 * g_tier, season=season,
                              branch="agro", tier_unlock=g_tier))

            # intermediates
            if cat == "grain":
                fid = f"flour_{iid}"
                add_item(item(fid, f"Мука ({cname}{g_name})", "intermediate", "grain", "agro", g_tier))
                add_recipe(recipe(f"mill_{iid}", f"Помол {cname}{g_name}", "mill",
                                  [{"item_id": iid, "qty": 5}], fid, 4, {"worker": 3}, 2, 80 * g_tier,
                                  branch="agro", tier_unlock=g_tier))
                # bread
                bread = f"bread_{iid}"
                add_item(item(bread, f"Хлеб ({cname}{g_name})", "finished", "food", "food", g_tier))
                add_recipe(recipe(f"bake_{bread}", f"Выпечка хлеба ({cname}{g_name})", "bakery",
                                  [{"item_id": fid, "qty": 3}, {"item_id": "packaging", "qty": 1}],
                                  bread, 4, {"worker": 4}, 2, 100 * g_tier, branch="food", tier_unlock=g_tier))
                if g_tier >= 1:
                    pastry = f"pastry_{iid}"
                    add_item(item(pastry, f"Выпечка ({cname}{g_name})", "finished", "food", "food", g_tier))
                    add_recipe(recipe(f"bake_{pastry}", f"Сдобная выпечка ({cname}{g_name})", "bakery_industrial" if g_tier >= 2 else "bakery",
                                      [{"item_id": fid, "qty": 2}, {"item_id": "packaging", "qty": 1}],
                                      pastry, 3, {"worker": 4, "engineer": 1}, 2, 200 * g_tier,
                                      branch="food", tier_unlock=g_tier))
            if cat == "oilseed":
                oil_id = f"veg_oil_{iid}"
                add_item(item(oil_id, f"Масло ({cname}{g_name})", "intermediate", "food", "agro", max(1, g_tier)))
                add_recipe(recipe(f"press_{iid}", f"Отжим масла ({cname})", "oil_press",
                                  [{"item_id": iid, "qty": 5}], oil_id, 3, {"worker": 4, "engineer": 1}, 3, 150,
                                  branch="agro", tier_unlock=max(1, g_tier)))
                bottled = f"bottled_{oil_id}"
                add_item(item(bottled, f"Масло бутылочное ({cname}{g_name})", "finished", "food", "food", max(1, g_tier)))
                add_recipe(recipe(f"bottle_{oil_id}", f"Розлив масла ({cname})", "food_plant",
                                  [{"item_id": oil_id, "qty": 2}, {"item_id": "bottle", "qty": 2}],
                                  bottled, 2, {"worker": 3}, 2, 120, branch="food", tier_unlock=max(1, g_tier)))
            if cat == "veg":
                canned = f"canned_{iid}"
                add_item(item(canned, f"Консервы ({cname}{g_name})", "finished", "food", "food", max(1, g_tier)))
                add_recipe(recipe(f"can_{iid}", f"Консервация ({cname})", "cannery" if g_tier >= 1 else "food_plant",
                                  [{"item_id": iid, "qty": 4}, {"item_id": "packaging", "qty": 2}],
                                  canned, 3, {"worker": 4}, 3, 120 * max(1, g_tier),
                                  branch="food", tier_unlock=max(0, g_tier)))
                snack = f"snack_{iid}"
                add_item(item(snack, f"Снеки ({cname}{g_name})", "finished", "food", "food", max(1, g_tier)))
                add_recipe(recipe(f"snack_{iid}", f"Снеки из {cname}", "snack_factory" if g_tier >= 2 else "food_plant",
                                  [{"item_id": iid, "qty": 4}, {"item_id": "packaging", "qty": 2}],
                                  snack, 4, {"worker": 5, "engineer": 1}, 3, 150 * max(1, g_tier),
                                  branch="food", tier_unlock=max(1, g_tier)))
            if cid == "sugar_beet":
                sugar = f"sugar_{iid}"
                add_item(item(sugar, f"Сахар{g_name}", "intermediate", "food", "agro", max(2, g_tier)))
                add_recipe(recipe(f"sugar_{iid}", f"Сахар{g_name}", "sugar_plant",
                                  [{"item_id": iid, "qty": 6}, {"item_id": "chemicals", "qty": 1}],
                                  sugar, 4, {"worker": 5, "engineer": 2}, 4, 300,
                                  branch="agro", tier_unlock=max(2, g_tier)))
            if cat == "fiber":
                fabric = f"fabric_{iid}"
                add_item(item(fabric, f"Ткань ({cname}{g_name})", "intermediate", "textile", "textile", g_tier))
                # textile buildings added later; recipe refs textile_mill
                add_recipe(recipe(f"weave_{iid}", f"Ткачество ({cname})", "textile_mill",
                                  [{"item_id": iid, "qty": 4}], fabric, 2, {"worker": 4, "engineer": 1}, 2, 100 * g_tier,
                                  branch="textile", tier_unlock=g_tier))

    # fruits / orchard
    add_building(building("orchard", "Сад", "extract", "agro", 10000, "rural", 2, "agro", 0))
    add_building(building("apiary", "Пасека", "extract", "agro", 16000, "rural", 2, "bashkir", 1))
    for fid, fname, season in FRUITS:
        for g_suf, g_name, g_tier in GRADES[:3]:
            iid = f"{fid}{g_suf}"
            add_item(item(iid, f"{fname}{g_name}", "raw", "fruit" if fid != "honey" else "food", "agro" if fid != "honey" else "bashkir", g_tier))
            btype = "apiary" if fid == "honey" else "orchard"
            add_recipe(recipe(f"farm_{iid}", f"Сбор: {fname}{g_name}", btype, [], iid, 6 if fid != "honey" else 3,
                              {"worker": 3}, 1, 80 * g_tier, season=season,
                              branch="bashkir" if fid == "honey" else "agro", tier_unlock=g_tier))
            if fid != "honey":
                juice = f"juice_{iid}"
                add_item(item(juice, f"Сок ({fname}{g_name})", "finished", "food", "food", g_tier))
                add_recipe(recipe(f"juice_{iid}", f"Сок {fname}{g_name}", "juice_plant" if g_tier >= 1 else "food_plant",
                                  [{"item_id": iid, "qty": 5}, {"item_id": "bottle", "qty": 4}],
                                  juice, 4, {"worker": 4}, 2, 100 * g_tier, branch="food", tier_unlock=g_tier))
                jam = f"jam_{iid}"
                add_item(item(jam, f"Варенье ({fname}{g_name})", "finished", "food", "food", max(1, g_tier)))
                add_recipe(recipe(f"jam_{iid}", f"Варенье {fname}", "confectionery" if g_tier >= 2 else "food_plant",
                                  [{"item_id": iid, "qty": 4}, {"item_id": "packaging", "qty": 2}],
                                  jam, 3, {"worker": 4}, 2, 150 * max(1, g_tier), branch="food", tier_unlock=max(1, g_tier)))
            else:
                # honey products
                for prod, pname, bld in [
                    ("chakchak", "Сәк-сәк", "confectionery"),
                    ("herbal_tea", "Травяной чай с мёдом", "food_plant"),
                    ("honey_candy", "Медовые конфеты", "confectionery"),
                ]:
                    pid = f"{prod}_{iid}"
                    add_item(item(pid, f"{pname}{g_name}", "finished", "food", "bashkir", max(1, g_tier)))
                    inputs = [{"item_id": iid, "qty": 2}, {"item_id": "packaging", "qty": 1}]
                    if prod == "chakchak":
                        # need flour - use wheat flour if exists
                        inputs = [{"item_id": "flour_wheat" if "flour_wheat" in items else "flour_wheat_std" if "flour_wheat_std" in items else iid, "qty": 2},
                                  {"item_id": iid, "qty": 2}, {"item_id": "packaging", "qty": 1}]
                        # fix: ensure flour_wheat exists
                    add_recipe(recipe(f"make_{pid}", f"{pname}{g_name}", bld if g_tier >= 1 else "bakery",
                                      inputs, pid, 3, {"worker": 4}, 2, 200 * max(1, g_tier),
                                      branch="bashkir", tier_unlock=max(1, g_tier)))

    # ensure base wheat flour id for chakchak
    if "flour_wheat" not in items:
        add_item(item("flour_wheat", "Мука (Пшеница)", "intermediate", "grain", "agro", 0))
        add_recipe(recipe("mill_wheat", "Помол пшеницы", "mill",
                          [{"item_id": "wheat", "qty": 5}], "flour_wheat", 4, {"worker": 3}, 2, 0, branch="agro"))

    # ---- LIVESTOCK / DAIRY ----
    add_building(building("livestock_farm", "Животноводческая ферма", "extract", "agro", 18000, "rural", 3, "dairy", 0))
    add_building(building("poultry_farm", "Птицефабрика", "extract", "agro", 28000, "rural", 3, "dairy", 1))
    add_building(building("dairy_plant", "Молокозавод", "factory", "food", 30000, branch="dairy", tier_unlock=0))
    add_building(building("dairy_premium", "Сыродельный комбинат", "factory", "food", 60000, branch="dairy", tier_unlock=2))
    add_building(building("meat_plant", "Мясокомбинат", "factory", "food", 42000, branch="food", tier_unlock=1))
    add_building(building("koumiss_plant", "Ҡымыҙный цех", "factory", "food", 28000, branch="bashkir", tier_unlock=1))

    for animal, aname, milk_id, milk_name, meat_id, meat_name, side_id, side_name in LIVESTOCK:
        for g_suf, g_name, g_tier in GRADES[:3]:
            aid = f"{animal}{g_suf}"
            add_item(item(aid, f"{aname}{g_name}", "raw", "livestock", "dairy", g_tier))
            btype = "poultry_farm" if animal == "poultry" and g_tier >= 1 else "livestock_farm"
            if animal == "bee":
                continue  # bees via apiary already
            add_recipe(recipe(f"raise_{aid}", f"Выращивание: {aname}{g_name}", btype, [], aid, 2,
                              {"worker": 3 + g_tier, "engineer": 1}, 2, 100 * g_tier,
                              branch="dairy", tier_unlock=g_tier))
            if milk_id:
                mid = f"{milk_id}{g_suf}"
                add_item(item(mid, f"{milk_name}{g_name}", "raw", "dairy", "dairy", g_tier))
                add_recipe(recipe(f"milk_{aid}", f"Надой: {milk_name}{g_name}", btype, [], mid, 8,
                                  {"worker": 3}, 1, 80 * g_tier, branch="dairy", tier_unlock=g_tier))
                bottled = f"bottled_{mid}"
                add_item(item(bottled, f"{milk_name} бут.{g_name}", "finished", "dairy", "dairy", g_tier))
                add_recipe(recipe(f"bottle_{mid}", f"Розлив {milk_name}", "dairy_plant",
                                  [{"item_id": mid, "qty": 5}, {"item_id": "bottle", "qty": 5}],
                                  bottled, 5, {"worker": 5, "engineer": 1}, 2, 50 * g_tier,
                                  branch="dairy", tier_unlock=g_tier))
                cheese = f"cheese_{mid}"
                add_item(item(cheese, f"Сыр ({milk_name}{g_name})", "finished", "dairy", "dairy", max(1, g_tier)))
                add_recipe(recipe(f"cheese_{mid}", f"Сыр из {milk_name}", "dairy_premium" if g_tier >= 2 else "dairy_plant",
                                  [{"item_id": mid, "qty": 6}, {"item_id": "packaging", "qty": 1}],
                                  cheese, 2, {"worker": 4, "engineer": 1}, 2, 200 * max(1, g_tier),
                                  branch="dairy", tier_unlock=max(1, g_tier)))
                yogurt = f"yogurt_{mid}"
                add_item(item(yogurt, f"Йогурт ({milk_name}{g_name})", "finished", "dairy", "dairy", max(1, g_tier)))
                add_recipe(recipe(f"yogurt_{mid}", f"Йогурт {milk_name}", "dairy_plant",
                                  [{"item_id": mid, "qty": 4}, {"item_id": "packaging", "qty": 2}],
                                  yogurt, 4, {"worker": 4}, 2, 150 * max(1, g_tier),
                                  branch="dairy", tier_unlock=max(1, g_tier)))
                if animal in ("cattle", "horse"):
                    koumiss = f"koumiss_{mid}"
                    add_item(item(koumiss, f"Ҡымыҙ{g_name}", "finished", "dairy", "bashkir", max(1, g_tier)))
                    add_recipe(recipe(f"koumiss_{mid}", f"Ҡымыҙ{g_name}", "koumiss_plant" if g_tier >= 1 else "dairy_plant",
                                      [{"item_id": mid, "qty": 6}, {"item_id": "bottle", "qty": 4}],
                                      koumiss, 4, {"worker": 4, "engineer": 1}, 2, 250 * max(1, g_tier),
                                      branch="bashkir", tier_unlock=max(1, g_tier)))
                    qurut = f"qurut_{mid}"
                    add_item(item(qurut, f"Ҡорот{g_name}", "finished", "dairy", "bashkir", max(1, g_tier)))
                    add_recipe(recipe(f"qurut_{mid}", f"Ҡорот{g_name}", "dairy_plant",
                                      [{"item_id": mid, "qty": 5}, {"item_id": "packaging", "qty": 1}],
                                      qurut, 3, {"worker": 3}, 2, 200 * max(1, g_tier),
                                      branch="bashkir", tier_unlock=max(1, g_tier)))
            if meat_id:
                meat = f"{meat_id}{g_suf}"
                add_item(item(meat, f"{meat_name}{g_name}", "intermediate", "food", "food", g_tier))
                add_recipe(recipe(f"meat_{aid}", f"Разделка: {meat_name}", "meat_plant" if g_tier >= 1 else "food_plant",
                                  [{"item_id": aid, "qty": 1}], meat, 4, {"worker": 4}, 2, 100 * g_tier,
                                  branch="food", tier_unlock=max(0, g_tier)))
                sausage = f"sausage_{meat}"
                add_item(item(sausage, f"Колбаса ({meat_name}{g_name})", "finished", "food", "food", max(1, g_tier)))
                add_recipe(recipe(f"sausage_{meat}", f"Колбаса {meat_name}", "meat_plant",
                                  [{"item_id": meat, "qty": 3}, {"item_id": "packaging", "qty": 1}],
                                  sausage, 3, {"worker": 5, "engineer": 1}, 3, 200 * max(1, g_tier),
                                  branch="food", tier_unlock=max(1, g_tier)))
            if side_id:
                sid = f"{side_id}{g_suf}"
                add_item(item(sid, f"{side_name}{g_name}", "intermediate", "textile" if "wool" in side_id or "feather" in side_id or "leather" in side_id or "hair" in side_id else "food",
                              "textile" if any(x in side_id for x in ("wool", "leather", "feather", "hair")) else "dairy", g_tier))
                add_recipe(recipe(f"side_{aid}", f"Побочный продукт: {side_name}", btype,
                                  [{"item_id": aid, "qty": 1}], sid, 2, {"worker": 2}, 1, 50 * g_tier,
                                  branch="textile" if "leather" in side_id or "wool" in side_id else "dairy",
                                  tier_unlock=g_tier))

    # ---- TEXTILE / APPAREL ----
    add_building(building("textile_mill", "Ткацкая", "factory", "apparel", 28000, branch="textile", tier_unlock=0))
    add_building(building("apparel_factory", "Швейная фабрика", "factory", "apparel", 30000, branch="textile", tier_unlock=0))
    add_building(building("shoe_factory", "Обувная фабрика", "factory", "apparel", 36000, branch="textile", tier_unlock=1))
    add_building(building("felt_workshop", "Войлочная мастерская", "factory", "apparel", 22000, branch="bashkir", tier_unlock=1))
    add_building(building("carpet_factory", "Ковровый комбинат", "factory", "apparel", 48000, branch="bashkir", tier_unlock=2))

    apparel_products = [
        ("clothing", "Одежда", "apparel_factory", "fabric_flax"),
        ("clothing_work", "Спецодежда", "apparel_factory", "fabric_flax_std"),
        ("clothing_winter", "Зимняя одежда", "apparel_factory", "fabric_flax_prem"),
        ("hat", "Головной убор", "apparel_factory", "fabric_flax"),
        ("shoes", "Обувь", "shoe_factory", "leather_cattle"),
        ("boots", "Сапоги", "shoe_factory", "leather_cattle_std"),
        ("felt_boots", "Валенки", "felt_workshop", "felt_base"),
        ("bashkir_carpet", "Башкирский ковёр", "carpet_factory", "felt_base"),
    ]
    # felt base from cattle/sheep side
    add_item(item("felt_base", "Войлок", "intermediate", "fiber", "bashkir", 1))
    add_recipe(recipe("make_felt_base", "Войлок", "felt_workshop",
                      [{"item_id": "wool_sheep" if False else "cattle", "qty": 1}],
                      "felt_base", 2, {"worker": 4}, 2, 200, branch="bashkir", tier_unlock=1))
    # fix felt input to use wool if present else cattle
    recipes["make_felt_base"]["inputs"] = [{"item_id": "wool_sheep", "qty": 2}] if "wool_sheep" in items else [{"item_id": "cattle", "qty": 1}]

    for pid, pname, bld, fabric_need in apparel_products:
        for g_suf, g_name, g_tier in GRADES[:3]:
            iid = f"{pid}{g_suf}"
            add_item(item(iid, f"{pname}{g_name}", "finished", "apparel" if "carpet" not in pid and "felt" not in pid else "home",
                          "textile" if "carpet" not in pid else "bashkir", max(g_tier, 1 if "shoe" in bld or "felt" in bld else 0)))
            fab = fabric_need if fabric_need in items else ("fabric_flax" if "fabric_flax" in items else "packaging")
            inputs = [{"item_id": fab, "qty": 2}, {"item_id": "packaging", "qty": 1}]
            if "shoe" in bld or pid in ("shoes", "boots"):
                leather = "leather_cattle" if "leather_cattle" in items else "cattle"
                inputs = [{"item_id": leather, "qty": 2}, {"item_id": "plastic", "qty": 1}, {"item_id": "packaging", "qty": 1}]
            if pid in ("felt_boots", "bashkir_carpet"):
                inputs = [{"item_id": "felt_base", "qty": 2}, {"item_id": "fabric_flax" if "fabric_flax" in items else "packaging", "qty": 1}]
            add_recipe(recipe(f"make_{iid}", f"{pname}{g_name}", bld, inputs, iid, 2 if "carpet" not in pid else 1,
                              {"worker": 5, "engineer": 1}, 2, 150 * max(1, g_tier),
                              branch="bashkir" if pid in ("felt_boots", "bashkir_carpet") else "textile",
                              tier_unlock=max(g_tier, 1 if "shoe" in bld or "felt" in bld or "carpet" in bld else 0)))

    # ---- TIMBER / FURNITURE ----
    add_building(building("sawmill", "Лесопилка", "factory", "furniture", 20000, branch="timber", tier_unlock=0))
    add_building(building("board_plant", "Комбинат плит", "factory", "furniture", 32000, branch="timber", tier_unlock=1))
    add_building(building("furniture_plant", "Мебельный комбинат", "factory", "furniture", 36000, branch="timber", tier_unlock=0))
    add_building(building("furniture_premium", "Мебель премиум", "factory", "furniture", 70000, branch="timber", tier_unlock=2))
    add_building(building("toy_factory", "Фабрика игрушек", "factory", "furniture", 30000, branch="timber", tier_unlock=1))

    for wood_id, wname in FOREST:
        for g_suf, g_name, g_tier in GRADES[:3]:
            iid = f"{wood_id}{g_suf}" if wood_id != "wood" or g_suf else "wood"
            if iid not in items:
                add_item(item(iid, f"{wname}{g_name}", "raw", "timber", "timber", g_tier))
                add_recipe(recipe(f"extract_{iid}", f"Заготовка: {wname}{g_name}", "logging_camp", [], iid, 8,
                                  {"worker": 4}, 1, 80 * g_tier, season=[9, 10], branch="timber", tier_unlock=g_tier))
            planks = f"planks_{iid}"
            add_item(item(planks, f"Доски ({wname}{g_name})", "intermediate", "timber", "timber", g_tier))
            add_recipe(recipe(f"saw_{iid}", f"Распил {wname}", "sawmill",
                              [{"item_id": iid, "qty": 4}], planks, 4, {"worker": 3}, 2, 50 * g_tier,
                              branch="timber", tier_unlock=g_tier))
            board = f"board_{iid}"
            add_item(item(board, f"Плита ({wname}{g_name})", "intermediate", "timber", "timber", max(1, g_tier)))
            add_recipe(recipe(f"board_{iid}", f"Плита {wname}", "board_plant",
                              [{"item_id": iid, "qty": 5}, {"item_id": "chemicals", "qty": 1}],
                              board, 3, {"worker": 5, "engineer": 2}, 4, 120 * max(1, g_tier),
                              branch="timber", tier_unlock=max(1, g_tier)))
            furn = f"furniture_{iid}"
            add_item(item(furn, f"Мебель ({wname}{g_name})", "finished", "home", "timber", g_tier))
            fab = "fabric_flax" if "fabric_flax" in items else "packaging"
            add_recipe(recipe(f"furn_{iid}", f"Мебель {wname}", "furniture_premium" if g_tier >= 2 else "furniture_plant",
                              [{"item_id": planks, "qty": 3}, {"item_id": fab, "qty": 1}],
                              furn, 1, {"worker": 5, "engineer": 1}, 3, 100 * g_tier,
                              branch="timber", tier_unlock=g_tier))
            toy = f"toy_{iid}"
            add_item(item(toy, f"Игрушки ({wname}{g_name})", "finished", "home", "timber", max(1, g_tier)))
            add_recipe(recipe(f"toy_{iid}", f"Игрушки {wname}", "toy_factory",
                              [{"item_id": planks, "qty": 2}, {"item_id": "packaging", "qty": 1}],
                              toy, 3, {"worker": 4}, 2, 120 * max(1, g_tier),
                              branch="timber", tier_unlock=max(1, g_tier)))

    # ---- MINING / METAL / OIL ----
    add_building(building("mine", "Шахта", "extract", "mining", 25000, "industrial", 3, "mining", 0))
    add_building(building("gas_well", "Газовая скважина", "extract", "oil", 42000, "industrial", 3, "oil_chem", 1))
    add_building(building("steel_mill", "Металлургический завод", "factory", "steel", 55000, "industrial", 4, "metal", 0))
    add_building(building("smelter_copper", "Медный завод", "factory", "steel", 50000, "industrial", 3, "metal", 1))
    add_building(building("rolling_mill", "Прокатный стан", "factory", "steel", 62000, "industrial", 4, "metal", 2))
    add_building(building("hardware_plant", "Метизный завод", "factory", "steel", 40000, branch="metal", tier_unlock=1))
    add_building(building("refinery", "НПЗ", "factory", "oil", 80000, "industrial", 4, "oil_chem", 0))
    add_building(building("refinery_deep", "Глубокая переработка нефти", "factory", "oil", 120000, "industrial", 5, "oil_chem", 3))
    add_building(building("fuel_blender", "Топливный цех", "factory", "oil", 45000, branch="oil_chem", tier_unlock=1))
    add_building(building("cement_plant", "Цементный завод", "factory", "mining", 55000, branch="mining", tier_unlock=1))
    add_building(building("brick_plant", "Кирпичный завод", "factory", "mining", 32000, branch="mining", tier_unlock=0))

    for mid, mname, bsrc in MINERALS:
        for g_suf, g_name, g_tier in GRADES[:2]:  # raw + std to control count
            iid = f"{mid}{g_suf}" if not (mid in items and g_suf == "") else mid
            if iid not in items:
                add_item(item(iid, f"{mname}{g_name}", "raw", "ore" if "ore" in mid else ("oil" if mid == "oil" else ("fuel" if mid in ("coal", "gas") else ("chem" if mid == "chemicals" else "mineral"))),
                              "oil_chem" if mid in ("oil", "gas", "chemicals") else "mining", g_tier))
            btype = bsrc
            if btype not in buildings:
                # map aliases already added
                btype = "mine" if bsrc == "mine" else ("quarry" if bsrc == "quarry" else ("oil_well" if bsrc == "oil_well" else ("gas_well" if bsrc == "gas_well" else "chem_plant")))
            if f"extract_{iid}" not in recipes and f"mine_{iid}" not in recipes:
                qty = 8 if "ore" in mid or mid in ("coal", "sand", "gravel") else 5
                add_recipe(recipe(f"extract_{iid}", f"Добыча: {mname}{g_name}", btype, [], iid, qty,
                                  {"worker": 5, "engineer": 1 + g_tier}, 3 + g_tier, 100 * g_tier,
                                  branch="oil_chem" if mid in ("oil", "gas", "chemicals") else "mining",
                                  tier_unlock=g_tier))

    # metal chain
    for ore, metal, mname, bld in [
        ("iron_ore", "steel", "Сталь", "steel_mill"),
        ("copper_ore", "copper", "Медь", "smelter_copper"),
        ("zinc_ore", "zinc", "Цинк", "smelter_copper"),
        ("bauxite", "aluminum", "Алюминий", "smelter_copper"),
        ("gold_ore", "gold", "Золото", "smelter_copper"),
    ]:
        for g_suf, g_name, g_tier in GRADES[:3]:
            ore_id = ore if g_suf == "" and ore in items else f"{ore}{g_suf}"
            if ore_id not in items:
                ore_id = ore if ore in items else ore_id
            metal_id = f"{metal}{g_suf}"
            add_item(item(metal_id, f"{mname}{g_name}", "intermediate", "metal", "metal", g_tier))
            coal = "coal" if "coal" in items else "coal_std"
            add_recipe(recipe(f"smelt_{metal_id}", f"Выплавка {mname}{g_name}", bld,
                              [{"item_id": ore_id if ore_id in items else ore, "qty": 5}, {"item_id": coal if coal in items else "coal", "qty": 2}],
                              metal_id, 3, {"worker": 7, "engineer": 3}, 7, 100 * g_tier,
                              branch="metal", tier_unlock=g_tier))
            goods = f"steel_goods_{metal_id}" if metal == "steel" else f"goods_{metal_id}"
            add_item(item(goods, f"Изделия ({mname}{g_name})", "finished", "home", "metal", max(1, g_tier)))
            add_recipe(recipe(f"goods_{metal_id}", f"Изделия из {mname}", "hardware_plant" if g_tier >= 1 else "steel_mill",
                              [{"item_id": metal_id, "qty": 2}], goods, 2, {"worker": 4, "engineer": 1}, 3, 150 * max(1, g_tier),
                              branch="metal", tier_unlock=max(1, g_tier)))
            if metal == "steel" and g_tier >= 1:
                beam = f"beam_{metal_id}"
                add_item(item(beam, f"Прокат{g_name}", "intermediate", "metal", "metal", g_tier))
                add_recipe(recipe(f"roll_{metal_id}", f"Прокат{g_name}", "rolling_mill",
                                  [{"item_id": metal_id, "qty": 3}], beam, 3, {"worker": 6, "engineer": 3}, 6, 250 * g_tier,
                                  branch="metal", tier_unlock=max(2, g_tier)))

    # oil products
    for g_suf, g_name, g_tier in GRADES[:3]:
        for fuel, fname, qty_out in [("gasoline", "Бензин", 4), ("diesel", "Дизель", 4), ("kerosene", "Керосин", 3), ("fuel_oil", "Мазут", 3)]:
            fid = f"{fuel}{g_suf}"
            add_item(item(fid, f"{fname}{g_name}", "intermediate", "fuel", "oil_chem", g_tier))
            add_recipe(recipe(f"refine_{fid}", f"{fname}{g_name}", "refinery_deep" if g_tier >= 3 else "refinery",
                              [{"item_id": "oil", "qty": 4}], fid, qty_out, {"worker": 6, "engineer": 4}, 6, 100 * g_tier,
                              branch="oil_chem", tier_unlock=g_tier))
            can = f"fuel_can_{fid}"
            add_item(item(can, f"Канистра ({fname}{g_name})", "finished", "fuel", "oil_chem", g_tier))
            add_recipe(recipe(f"pack_{fid}", f"Канистра {fname}", "fuel_blender" if g_tier >= 1 else "refinery",
                              [{"item_id": fid, "qty": 2}, {"item_id": "plastic", "qty": 1}],
                              can, 2, {"worker": 3}, 2, 120 * g_tier, branch="oil_chem", tier_unlock=g_tier))

    # cement / brick from minerals
    add_item(item("cement", "Цемент", "intermediate", "mineral", "mining", 1))
    add_item(item("brick", "Кирпич", "finished", "home", "mining", 0))
    add_recipe(recipe("make_cement", "Цемент", "cement_plant",
                      [{"item_id": "limestone" if "limestone" in items else "sand", "qty": 5}, {"item_id": "coal", "qty": 1}],
                      "cement", 4, {"worker": 5, "engineer": 2}, 5, 200, branch="mining", tier_unlock=1))
    add_recipe(recipe("make_brick", "Кирпич", "brick_plant",
                      [{"item_id": "clay" if "clay" in items else "sand", "qty": 5}],
                      "brick", 6, {"worker": 4}, 3, 0, branch="mining"))

    # ---- ELECTRONICS / TECH ----
    add_building(building("electronics_plant", "Электронное производство", "factory", "tech", 60000, branch="tech", tier_unlock=1))
    add_building(building("appliance_plant", "Завод бытовой техники", "factory", "tech", 50000, branch="tech", tier_unlock=1))
    add_building(building("chip_fab", "Микросборочный цех", "factory", "tech", 90000, branch="tech", tier_unlock=3))
    add_building(building("cable_plant", "Кабельный завод", "factory", "tech", 40000, branch="tech", tier_unlock=1))

    for g_suf, g_name, g_tier in GRADES[:3]:
        wire = f"wire{g_suf}"
        circuit = f"circuit{g_suf}"
        phone = f"electronics{g_suf}"
        appl = f"appliance{g_suf}"
        copper = "copper" if "copper" in items else ("copper_std" if "copper_std" in items else "copper")
        add_item(item(wire, f"Провод{g_name}", "intermediate", "tech", "tech", max(1, g_tier)))
        add_item(item(circuit, f"Микросхема{g_name}", "intermediate", "tech", "tech", max(1, g_tier)))
        add_item(item(phone, f"Электроника{g_name}", "finished", "tech", "tech", max(1, g_tier)))
        add_item(item(appl, f"Бытовая техника{g_name}", "finished", "home", "tech", max(1, g_tier)))
        add_recipe(recipe(f"make_{wire}", f"Провод{g_name}", "cable_plant" if g_tier >= 1 else "electronics_plant",
                          [{"item_id": copper if copper in items else "copper", "qty": 2}, {"item_id": "plastic", "qty": 1}],
                          wire, 4, {"worker": 4, "engineer": 2}, 3, 150 * max(1, g_tier), branch="tech", tier_unlock=max(1, g_tier)))
        add_recipe(recipe(f"make_{circuit}", f"Микросхема{g_name}", "chip_fab" if g_tier >= 3 else "electronics_plant",
                          [{"item_id": copper if copper in items else "copper", "qty": 1},
                           {"item_id": "quartz", "qty": 1}, {"item_id": "chemicals", "qty": 1}],
                          circuit, 1, {"worker": 4, "engineer": 4}, 5, 400 * max(1, g_tier), branch="tech", tier_unlock=max(1, g_tier)))
        add_recipe(recipe(f"make_{phone}", f"Электроника{g_name}", "electronics_plant",
                          [{"item_id": circuit, "qty": 1}, {"item_id": "plastic", "qty": 2},
                           {"item_id": wire, "qty": 2}, {"item_id": "packaging", "qty": 1}],
                          phone, 1, {"worker": 5, "engineer": 3, "manager": 1}, 4, 500 * max(1, g_tier),
                          branch="tech", tier_unlock=max(1, g_tier)))
        steel = "steel" if "steel" in items else "steel_std"
        add_recipe(recipe(f"make_{appl}", f"Бытовая техника{g_name}", "appliance_plant",
                          [{"item_id": steel if steel in items else "steel", "qty": 2},
                           {"item_id": circuit, "qty": 1}, {"item_id": "plastic", "qty": 1}, {"item_id": wire, "qty": 1}],
                          appl, 1, {"worker": 6, "engineer": 3, "manager": 1}, 5, 400 * max(1, g_tier),
                          branch="tech", tier_unlock=max(1, g_tier)))

    # ---- BOOKS / MEDIA from paper ----
    add_building(building("print_shop", "Типография", "factory", "packaging", 28000, branch="packaging", tier_unlock=1))
    for g_suf, g_name, g_tier in GRADES[:3]:
        book = f"book{g_suf}"
        add_item(item(book, f"Книги{g_name}", "finished", "home", "packaging", max(1, g_tier)))
        add_recipe(recipe(f"make_{book}", f"Книги{g_name}", "print_shop" if g_tier >= 1 else "paper_mill",
                          [{"item_id": "paper", "qty": 4}, {"item_id": "packaging", "qty": 1}],
                          book, 2, {"worker": 3, "manager": 1}, 2, 120 * max(1, g_tier),
                          branch="packaging", tier_unlock=max(1, g_tier)))

    # ---- EXTRA SKU fill via regional retail variants of finished goods ----
    # Duplicate finished into regional SKUs that share recipes via "repack" at food_plant/department
    add_building(building("repack_center", "Фасовочный центр", "factory", "food", 25000, branch="retail", tier_unlock=2))
    finished_now = [i for i in items.values() if i["tier"] == "finished"]
    for it in list(finished_now):
        for r_suf, r_name in REGIONS[1:]:  # skip empty
            if len(items) >= 1000:
                break
            nid = f"{it['id']}{r_suf}"
            if nid in items:
                continue
            add_item(item(nid, f"{it['name']}{r_name}", "finished", it["category"], it.get("branch", "retail"),
                          max(it.get("tier_unlock", 0), 2), region=r_suf.strip("_")))
            add_recipe(recipe(f"repack_{nid}", f"Фасовка: {it['name']}{r_name}", "repack_center",
                              [{"item_id": it["id"], "qty": 1}, {"item_id": "packaging", "qty": 1}],
                              nid, 1, {"worker": 2}, 1, 80, branch="retail", tier_unlock=2))
        if len(items) >= 1000:
            break

    # If still short on items, add more agro grade/export variants for remaining crops
    if len(items) < 1000:
        for cid, cname, cat, season in AGRO_CROPS:
            for g_suf, g_name, g_tier in GRADES[3:]:  # export
                iid = f"{cid}{g_suf}"
                if iid in items:
                    continue
                add_item(item(iid, f"{cname}{g_name}", "raw", cat, "agro", g_tier))
                add_recipe(recipe(f"grow_{iid}", f"Выращивание: {cname}{g_name}", "agro_complex", [], iid, 12,
                                  {"worker": 5, "engineer": 1}, 2, 400, season=season, branch="agro", tier_unlock=g_tier))
                if len(items) >= 1000:
                    break
            if len(items) >= 1000:
                break

    # ---- SHOPS 20-30 ----
    shop_defs = [
        shop("grocery", "Продуктовый", ["food", "dairy", "grain", "veg", "fruit"], ["finished", "intermediate"], 1.0),
        shop("bakery_shop", "Булочная", ["food", "grain"], ["finished"], 0.9),
        shop("dairy_shop", "Молочный", ["dairy", "food"], ["finished"], 1.0),
        shop("butcher", "Мясной", ["food"], ["finished", "intermediate"], 1.1),
        shop("fish_shop", "Рыбный / кумысный", ["dairy", "food"], ["finished"], 1.0),
        shop("confectionery_shop", "Кондитерская", ["food"], ["finished"], 1.1),
        shop("beverage_shop", "Напитки", ["food", "dairy"], ["finished"], 1.0),
        shop("organic_shop", "Эко-лавка", ["food", "dairy", "fruit", "veg"], ["finished"], 1.3),
        shop("furniture_shop", "Мебельный", ["home"], ["finished"], 1.2),
        shop("home_decor", "Дом и уют", ["home"], ["finished"], 1.15),
        shop("toy_shop", "Игрушки", ["home"], ["finished"], 1.1),
        shop("bookstore", "Книжный", ["home"], ["finished"], 1.0),
        shop("apparel_shop", "Одежда и обувь", ["apparel"], ["finished"], 1.1),
        shop("fashion_boutique", "Бутик", ["apparel"], ["finished"], 1.4),
        shop("workwear_shop", "Спецодежда", ["apparel"], ["finished"], 1.0),
        shop("electronics_shop", "Электроника", ["tech", "home"], ["finished"], 1.4),
        shop("appliance_shop", "Бытовая техника", ["home", "tech"], ["finished"], 1.3),
        shop("fuel_station", "АЗС", ["fuel"], ["finished", "intermediate"], 1.3),
        shop("auto_parts", "Автотовары", ["fuel", "metal", "tech"], ["finished", "intermediate"], 1.2),
        shop("hardware", "Хозтовары", ["home", "metal", "timber", "packaging", "mineral"], ["finished", "intermediate"], 1.0),
        shop("building_market", "Строймаркет", ["mineral", "metal", "timber", "home"], ["finished", "intermediate"], 1.2),
        shop("department", "Универмаг", ["food", "apparel", "home", "tech", "dairy"], ["finished"], 1.6),
        shop("hypermarket", "Гипермаркет", ["food", "apparel", "home", "tech", "dairy", "fuel"], ["finished"], 1.8),
        shop("pharmacy_adjacent", "Сопутствующие товары", ["food", "home"], ["finished"], 1.1),
        shop("bashkir_souvenir", "Башкирские сувениры", ["home", "food", "dairy", "apparel"], ["finished"], 1.25),
        shop("wholesale_club", "Оптовый клуб", ["food", "home", "apparel", "tech", "fuel", "dairy"], ["finished", "intermediate"], 1.5),
    ]
    for s in shop_defs:
        add_shop(s)

    # More buildings to reach 100-150
    extra_buildings = [
        ("seed_plant", "Семенной завод", "factory", "agro", 35000, "agro", 2),
        ("feed_mill", "Комбикормовый завод", "factory", "agro", 30000, "dairy", 1),
        ("cold_storage", "Холодильник", "logistics", "logistics", 28000, "retail", 1),
        ("rail_depot", "Ж/д депо", "logistics", "logistics", 70000, "retail", 3),
        ("port_terminal", "Логистический хаб", "logistics", "logistics", 90000, "retail", 3),
        ("chemical_fine", "Тонкая химия", "factory", "chem", 75000, "oil_chem", 3),
        ("paint_plant", "Лакокрасочный завод", "factory", "chem", 42000, "oil_chem", 2),
        ("fertilizer_plant", "Завод удобрений", "factory", "chem", 55000, "oil_chem", 2),
        ("pipe_plant", "Трубный завод", "factory", "steel", 65000, "metal", 2),
        ("machine_works", "Машиностроительный цех", "factory", "steel", 80000, "metal", 3),
        ("glass_premium", "Хрустальный цех", "factory", "packaging", 50000, "packaging", 2),
        ("textile_dye", "Красильня", "factory", "apparel", 26000, "textile", 1),
        ("knitwear", "Трикотажная фабрика", "factory", "apparel", 34000, "textile", 2),
        ("leather_works", "Кожзавод", "factory", "apparel", 38000, "textile", 1),
        ("dairy_ice", "Хладокомбинат мороженого", "factory", "food", 40000, "dairy", 2),
        ("pasta_plant", "Макаронная фабрика", "factory", "food", 32000, "food", 1),
        ("brewery", "Пивоваренный/квасной завод", "factory", "food", 48000, "food", 2),
        ("mineral_water", "Розлив минеральной воды", "factory", "food", 30000, "food", 1),
        ("pharma_lite", "БАД / фитоцех", "factory", "food", 52000, "bashkir", 3),
        ("solar_pilot", "Пилот ВИЭ-компонентов", "factory", "tech", 100000, "tech", 4),
    ]
    for bid, bname, cat, tag, cost, branch, tier in extra_buildings:
        add_building(building(bid, bname, cat, tag, cost, branch=branch, tier_unlock=tier))

    # fill recipes for some extra buildings to avoid orphan enterprises
    if "wheat" in items:
        add_item(item("animal_feed", "Комбикорм", "intermediate", "grain", "dairy", 1))
        add_recipe(recipe("make_feed", "Комбикорм", "feed_mill",
                          [{"item_id": "wheat", "qty": 4}, {"item_id": "potato" if "potato" in items else "wheat", "qty": 2}],
                          "animal_feed", 5, {"worker": 4}, 2, 150, branch="dairy", tier_unlock=1))
    add_item(item("fertilizer", "Удобрения", "intermediate", "chem", "oil_chem", 2))
    add_recipe(recipe("make_fertilizer", "Удобрения", "fertilizer_plant",
                      [{"item_id": "chemicals", "qty": 3}, {"item_id": "gas" if "gas" in items else "oil", "qty": 1}],
                      "fertilizer", 4, {"worker": 5, "engineer": 2}, 4, 300, branch="oil_chem", tier_unlock=2))
    add_item(item("paint", "Краска", "finished", "home", "oil_chem", 2))
    add_recipe(recipe("make_paint", "Краска", "paint_plant",
                      [{"item_id": "chemicals", "qty": 2}, {"item_id": "oil", "qty": 1}, {"item_id": "packaging", "qty": 1}],
                      "paint", 3, {"worker": 4, "engineer": 1}, 3, 250, branch="oil_chem", tier_unlock=2))
    add_item(item("pasta", "Макароны", "finished", "food", "food", 1))
    flour_ref = "flour_wheat" if "flour_wheat" in items else next((i for i in items if i.startswith("flour_")), "packaging")
    add_recipe(recipe("make_pasta", "Макароны", "pasta_plant",
                      [{"item_id": flour_ref, "qty": 4}, {"item_id": "packaging", "qty": 1}],
                      "pasta", 4, {"worker": 4}, 2, 150, branch="food", tier_unlock=1))
    add_item(item("kvass", "Квас", "finished", "food", "food", 2))
    add_recipe(recipe("make_kvass", "Квас", "brewery",
                      [{"item_id": flour_ref, "qty": 2}, {"item_id": "bottle", "qty": 4}],
                      "kvass", 4, {"worker": 4, "engineer": 1}, 2, 220, branch="food", tier_unlock=2))
    add_item(item("mineral_water", "Минеральная вода", "finished", "food", "food", 1))
    add_recipe(recipe("make_mineral_water", "Минеральная вода", "mineral_water",
                      [{"item_id": "bottle", "qty": 6}],
                      "mineral_water", 6, {"worker": 3}, 2, 100, branch="food", tier_unlock=1))
    add_item(item("ice_cream", "Мороженое", "finished", "dairy", "dairy", 2))
    milk_ref = "milk_cattle" if "milk_cattle" in items else ("milk" if "milk" in items else next((i for i in items if i.startswith("milk_")), "packaging"))
    # livestock generator used milk_cattle pattern? farm_milk used milk_id like "milk" for cattle -> milk / milk_std
    milk_ref = "milk" if "milk" in items else ("milk_std" if "milk_std" in items else milk_ref)
    # actually cattle milk ids are milk / milk_std / milk_prem from milk_id "milk"
    add_recipe(recipe("make_ice_cream", "Мороженое", "dairy_ice",
                      [{"item_id": milk_ref if milk_ref in items else "packaging", "qty": 4}, {"item_id": "packaging", "qty": 2}],
                      "ice_cream", 4, {"worker": 4, "engineer": 1}, 3, 250, branch="dairy", tier_unlock=2))

    # Research recipes (consume cash via research_cost; lab converts RP later)
    add_item(item("research_notes", "Исследовательские отчёты", "intermediate", "tech", "research", 0))
    add_recipe(recipe("lab_basic_research", "Базовые исследования", "research_lab",
                      [{"item_id": "paper", "qty": 1}], "research_notes", 1,
                      {"engineer": 2, "manager": 1}, 2, 0, branch="research"))

    # ---- Fix chakchak flour refs & felt ----
    # Rebuild honey confection inputs now that flour_wheat exists
    for rid, r in list(recipes.items()):
        if rid.startswith("make_chakchak"):
            honey_inp = None
            for inp in r["inputs"]:
                if inp["item_id"].startswith("honey"):
                    honey_inp = inp["item_id"]
            r["inputs"] = [
                {"item_id": "flour_wheat", "qty": 2},
                {"item_id": honey_inp or "honey", "qty": 2},
                {"item_id": "packaging", "qty": 1},
            ]

    # Ensure every finished item is sellable in at least one shop (by category)
    # Already covered by broad shop categories; specialty items like bashkir go to bashkir_souvenir + grocery/department.

    # Count & trim/pad recipes toward 1400-1800
    # Add process variants (efficiency lines) for factories
    base_recipe_list = list(recipes.values())
    for r in base_recipe_list:
        if len(recipes) >= 1700:
            break
        if not r.get("inputs"):
            continue
        if r.get("tier_unlock", 0) >= 3:
            continue
        vid = r["id"] + "_eff"
        if vid in recipes:
            continue
        rr = deepcopy(r)
        rr["id"] = vid
        rr["name"] = r["name"] + " (оптимиз.)"
        rr["tier_unlock"] = min(4, r.get("tier_unlock", 0) + 1)
        rr["research_cost"] = max(200, int(r.get("research_cost", 0)) + 300)
        # slightly better output
        rr["output"] = dict(r["output"])
        rr["output"]["qty"] = r["output"]["qty"] + 1
        rr["energy"] = max(1, int(r.get("energy", 1)))
        recipes[vid] = rr

    # Pad buildings if needed with specialized lines
    while len(buildings) < 110:
        n = len(buildings)
        bid = f"line_workshop_{n}"
        add_building(building(bid, f"Специализированный цех №{n}", "factory", "general", 28000 + n * 100,
                              branch="research", tier_unlock=min(4, n % 5)))
        # tiny recipe using packaging -> research notes to avoid orphan
        rid = f"workshop_job_{n}"
        add_recipe(recipe(rid, f"Подряд цеха №{n}", bid,
                          [{"item_id": "packaging", "qty": 1}], "research_notes", 1,
                          {"worker": 2}, 1, 50, branch="research", tier_unlock=min(4, n % 5)))

    items_list = sorted(items.values(), key=lambda x: (x["tier"], x["id"]))
    recipes_list = sorted(recipes.values(), key=lambda x: (x["building_type"], x["id"]))
    buildings_list = sorted(buildings.values(), key=lambda x: (x["category"], x["id"]))
    shops_list = sorted(shops.values(), key=lambda x: x["id"])

    save("items.json", items_list)
    save("recipes.json", recipes_list)
    save("buildings.json", buildings_list)
    save("shops.json", shops_list)

    # Research tree generated from branches/tiers
    branches = {
        "agro": "Агропром Башкортостана",
        "food": "Пищевая промышленность",
        "dairy": "Молоко и животноводство",
        "textile": "Текстиль и лёгпром",
        "timber": "Лес и мебель",
        "mining": "Недра и стройматериалы",
        "metal": "Металлургия Урала",
        "oil_chem": "Нефтехимия",
        "tech": "Электроника и техника",
        "packaging": "Упаковка и бумага",
        "bashkir": "Башкирский бренд",
        "retail": "Ритейл и логистика",
        "research": "R&D",
    }
    nodes = []
    for br, title in branches.items():
        for tier in range(0, 5):
            nid = f"{br}_t{tier}"
            requires = [] if tier == 0 else [f"{br}_t{tier-1}"]
            unlock_recipes = [r["id"] for r in recipes_list if r.get("branch") == br and r.get("tier_unlock", 0) == tier][:40]
            unlock_buildings = [b["id"] for b in buildings_list if b.get("branch") == br and b.get("tier_unlock", 0) == tier]
            nodes.append({
                "id": nid,
                "name": f"{title} · этап {tier}",
                "branch": br,
                "tier": tier,
                "cost": tier * 800,
                "rp_cost": tier * 5,
                "requires": requires,
                "unlocks_recipes": unlock_recipes,
                "unlocks_buildings": unlock_buildings,
                "starter": tier == 0,
            })
    save("rnd_tree.json", {"nodes": nodes, "branches": branches})

    stats = {
        "items": len(items_list),
        "recipes": len(recipes_list),
        "buildings": len(buildings_list),
        "shops": len(shops_list),
        "by_tier": {},
        "extract_recipes": sum(1 for r in recipes_list if not r.get("inputs")),
        "production_recipes": sum(1 for r in recipes_list if r.get("inputs")),
    }
    for it in items_list:
        stats["by_tier"][it["tier"]] = stats["by_tier"].get(it["tier"], 0) + 1
    save("catalog_stats.json", stats)
    print(json.dumps(stats, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
