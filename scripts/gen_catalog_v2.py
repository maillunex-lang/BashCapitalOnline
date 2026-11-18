#!/usr/bin/env python3
"""Generate data/catalog_v2 for BashCapital economy_catalog_version=2."""
from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
OUT = ROOT / "data" / "catalog_v2"
LEGACY = ROOT / "data"

# --- 40 raw ---
RAW = [
    ("iron_ore", "Железная руда", "ore", "mine", "metal"),
    ("copper_ore", "Медная руда", "ore", "mine", "metal"),
    ("bauxite", "Бокситы", "ore", "mine", "metal"),
    ("limestone", "Известняк", "mineral", "quarry", "mining"),
    ("sand", "Песок", "mineral", "quarry", "mining"),
    ("clay", "Глина", "mineral", "quarry", "mining"),
    ("quartz", "Кварц", "mineral", "quarry", "mining"),
    ("stone", "Камень", "mineral", "quarry", "mining"),
    ("salt", "Соль", "mineral", "mine", "mining"),
    ("crude_oil", "Нефть", "fuel", "oil_well", "oil_chem"),
    ("natural_gas", "Природный газ", "fuel", "gas_well", "oil_chem"),
    ("coal", "Уголь", "fuel", "mine", "mining"),
    ("peat", "Торф", "fuel", "peat_field", "mining"),
    ("wheat", "Пшеница", "grain", "grain_farm", "agro"),
    ("rye", "Рожь", "grain", "grain_farm", "agro"),
    ("barley", "Ячмень", "grain", "grain_farm", "agro"),
    ("oats", "Овёс", "grain", "grain_farm", "agro"),
    ("buckwheat", "Гречиха", "grain", "grain_farm", "agro"),
    ("potato", "Картофель", "veg", "crop_farm", "agro"),
    ("sugar_beet", "Сахарная свёкла", "veg", "crop_farm", "agro"),
    ("sunflower", "Подсолнечник", "oilseed", "oilseed_farm", "agro"),
    ("flax", "Лён", "fiber", "fiber_farm", "textile"),
    ("vegetables", "Овощи", "veg", "vegetable_farm", "agro"),
    ("milk", "Сырое молоко", "dairy", "dairy_farm", "dairy"),
    ("beef", "Говядина", "food", "livestock_farm", "dairy"),
    ("pork", "Свинина", "food", "livestock_farm", "dairy"),
    ("poultry", "Мясо птицы", "food", "poultry_farm", "dairy"),
    ("eggs", "Яйца", "food", "poultry_farm", "dairy"),
    ("wool", "Шерсть", "textile", "sheep_farm", "textile"),
    ("raw_hide", "Кожа-сырец", "textile", "livestock_farm", "textile"),
    ("softwood", "Хвойная древесина", "timber", "logging_camp", "timber"),
    ("hardwood", "Лиственная древесина", "timber", "logging_camp", "timber"),
    ("wood_chips", "Древесная щепа", "timber", "logging_camp", "timber"),
    ("forest_berries", "Лесные ягоды", "fruit", "forest_gathering", "bashkir"),
    ("mushrooms", "Грибы", "food", "forest_gathering", "bashkir"),
    ("honey", "Мёд", "food", "apiary", "bashkir"),
    ("apples", "Яблоки", "fruit", "orchard", "agro"),
    ("cherries", "Вишня", "fruit", "orchard", "agro"),
    ("horse_meat", "Конина", "food", "horse_farm", "bashkir"),
    ("mare_milk", "Кобылье молоко", "dairy", "horse_farm", "bashkir"),
]

# --- 55 intermediate ---
INTER = [
    ("pig_iron", "Чугун", "metal", "metal"),
    ("steel", "Сталь", "metal", "metal"),
    ("copper", "Медь", "metal", "metal"),
    ("aluminum", "Алюминий", "metal", "metal"),
    ("steel_sheet", "Стальной лист", "metal", "metal"),
    ("steel_pipe", "Стальная труба", "metal", "metal"),
    ("copper_wire", "Медный провод", "metal", "metal"),
    ("aluminum_profile", "Алюминиевый профиль", "metal", "metal"),
    ("metal_parts", "Металлические детали", "metal", "metal"),
    ("gasoline_base", "Бензиновая фракция", "fuel", "oil_chem"),
    ("diesel_base", "Дизельная фракция", "fuel", "oil_chem"),
    ("fuel_oil", "Мазут", "fuel", "oil_chem"),
    ("bitumen", "Битум", "fuel", "oil_chem"),
    ("polymer_granules", "Полимерные гранулы", "chem", "oil_chem"),
    ("synthetic_rubber", "Синтетический каучук", "chem", "oil_chem"),
    ("industrial_solvent", "Растворитель", "chem", "oil_chem"),
    ("chemical_reagent", "Химический реагент", "chem", "oil_chem"),
    ("cement", "Цемент", "mineral", "mining"),
    ("concrete", "Бетон", "mineral", "mining"),
    ("brick", "Кирпич", "mineral", "mining"),
    ("glass_sheet", "Листовое стекло", "mineral", "packaging"),
    ("insulation", "Утеплитель", "mineral", "mining"),
    ("ceramic_tile", "Керамическая плитка", "mineral", "mining"),
    ("paint_base", "Основа краски", "chem", "oil_chem"),
    ("dry_mix", "Сухая строительная смесь", "mineral", "mining"),
    ("wheat_flour", "Пшеничная мука", "grain", "agro"),
    ("rye_flour", "Ржаная мука", "grain", "agro"),
    ("groats", "Крупа", "grain", "agro"),
    ("sugar", "Сахар", "food", "agro"),
    ("sunflower_oil", "Растительное масло", "food", "agro"),
    ("cream", "Сливки", "dairy", "dairy"),
    ("butter_base", "Масляная основа", "dairy", "dairy"),
    ("minced_meat", "Фарш", "food", "dairy"),
    ("fruit_puree", "Фруктовое пюре", "food", "agro"),
    ("fermented_milk", "Кисломолочная основа", "dairy", "dairy"),
    ("flax_fiber", "Льняное волокно", "textile", "textile"),
    ("yarn", "Пряжа", "textile", "textile"),
    ("woven_fabric", "Ткань", "textile", "textile"),
    ("knitted_fabric", "Трикотажное полотно", "textile", "textile"),
    ("tanned_leather", "Выделанная кожа", "textile", "textile"),
    ("textile_filling", "Текстильный наполнитель", "textile", "textile"),
    ("dyed_fabric", "Окрашенная ткань", "textile", "textile"),
    ("lumber", "Доски", "timber", "timber"),
    ("plywood", "Фанера", "timber", "timber"),
    ("particle_board", "ДСП", "timber", "timber"),
    ("paper", "Бумага", "packaging", "timber"),
    ("cardboard", "Картон", "packaging", "timber"),
    ("furniture_parts", "Мебельные детали", "timber", "timber"),
    ("engine", "Двигатель", "tech", "metal"),
    ("wiring_harness", "Жгут проводки", "tech", "tech"),
    ("circuit_board", "Печатная плата", "tech", "tech"),
    ("electronic_module", "Электронный модуль", "tech", "tech"),
    ("battery_cell", "Аккумуляторная ячейка", "tech", "tech"),
    ("plastic_parts", "Пластиковые детали", "tech", "tech"),
    ("mechanical_assembly", "Механический узел", "tech", "metal"),
]

# --- 75 finished (oatmeal → horse_meat_pack per ALT-06 normalization) ---
FINISHED = [
    ("wheat_bread", "Хлеб пшеничный", "food", "agro"),
    ("rye_bread", "Хлеб ржаной", "food", "agro"),
    ("bun", "Булочки", "food", "agro"),
    ("cookies", "Печенье", "food", "agro"),
    ("pasta", "Макароны", "food", "agro"),
    ("buckwheat_pack", "Гречка фасованная", "food", "agro"),
    ("horse_meat_pack", "Конина фасованная", "food", "bashkir"),  # replaces oatmeal
    ("potato_pack", "Картофель фасованный", "food", "agro"),
    ("vegetable_mix", "Овощная смесь", "food", "agro"),
    ("sunflower_oil_bottle", "Масло бутылочное", "food", "agro"),
    ("sugar_pack", "Сахар фасованный", "food", "agro"),
    ("milk_bottle", "Молоко питьевое", "dairy", "dairy"),
    ("cream_pack", "Сливки фасованные", "dairy", "dairy"),
    ("butter", "Масло сливочное", "dairy", "dairy"),
    ("cheese", "Сыр", "dairy", "dairy"),
    ("yogurt", "Йогурт", "dairy", "dairy"),
    ("beef_pack", "Говядина фасованная", "food", "dairy"),
    ("pork_pack", "Свинина фасованная", "food", "dairy"),
    ("chicken_pack", "Курица фасованная", "food", "dairy"),
    ("eggs_pack", "Яйца фасованные", "food", "dairy"),
    ("bottled_water", "Вода питьевая", "food", "agro"),
    ("apple_juice", "Яблочный сок", "food", "agro"),
    ("cherry_juice", "Вишнёвый сок", "food", "agro"),
    ("berry_drink", "Ягодный морс", "food", "bashkir"),
    ("kvass", "Квас", "food", "agro"),
    ("koumiss", "Кумыс", "dairy", "bashkir"),
    ("dairy_drink", "Кисломолочный напиток", "dairy", "dairy"),
    ("tshirt", "Футболка", "apparel", "textile"),
    ("shirt", "Рубашка", "apparel", "textile"),
    ("trousers", "Брюки", "apparel", "textile"),
    ("jacket", "Куртка", "apparel", "textile"),
    ("knit_sweater", "Свитер", "apparel", "textile"),
    ("workwear", "Спецодежда", "apparel", "textile"),
    ("leather_shoes", "Кожаная обувь", "apparel", "textile"),
    ("textile_shoes", "Текстильная обувь", "apparel", "textile"),
    ("winter_hat", "Зимняя шапка", "apparel", "textile"),
    ("table", "Стол", "home", "timber"),
    ("chair", "Стул", "home", "timber"),
    ("wardrobe", "Шкаф", "home", "timber"),
    ("bed", "Кровать", "home", "timber"),
    ("sofa", "Диван", "home", "timber"),
    ("detergent", "Моющее средство", "home", "oil_chem"),
    ("dishware", "Посуда", "home", "mining"),
    ("pillow", "Подушка", "home", "textile"),
    ("cement_bag", "Цемент мешковой", "mineral", "mining"),
    ("brick_pack", "Кирпич упаковка", "mineral", "mining"),
    ("concrete_mix", "Бетон товарный", "mineral", "mining"),
    ("window_unit", "Оконный блок", "mineral", "mining"),
    ("paint_can", "Краска", "mineral", "oil_chem"),
    ("tile_pack", "Плитка упаковка", "mineral", "mining"),
    ("insulation_pack", "Утеплитель упаковка", "mineral", "mining"),
    ("pipe_set", "Комплект труб", "metal", "metal"),
    ("tire", "Шина", "tech", "oil_chem"),
    ("car_battery", "Автомобильный аккумулятор", "tech", "tech"),
    ("engine_oil", "Моторное масло", "fuel", "oil_chem"),
    ("brake_parts", "Тормозные детали", "tech", "metal"),
    ("filter_set", "Фильтры", "tech", "metal"),
    ("car_body_parts", "Кузовные детали", "tech", "metal"),
    ("car_spare_kit", "Комплект запчастей", "tech", "metal"),
    ("smartphone", "Смартфон", "tech", "tech"),
    ("television", "Телевизор", "tech", "tech"),
    ("refrigerator", "Холодильник", "tech", "tech"),
    ("washing_machine", "Стиральная машина", "tech", "tech"),
    ("electric_kettle", "Электрочайник", "tech", "tech"),
    ("home_router", "Домашний роутер", "tech", "tech"),
    ("bashkir_honey_pack", "Башкирский мёд фасованный", "food", "bashkir"),
    ("qurut", "Ҡорот", "dairy", "bashkir"),
    ("chak_chak", "Чак-чак", "food", "bashkir"),
    ("honey_gift_set", "Подарочный набор мёда", "food", "bashkir"),
    ("bashkir_souvenir", "Сувенир Башкортостана", "home", "bashkir"),
    ("ai92", "Бензин АИ-92", "fuel", "oil_chem"),
    ("ai95", "Бензин АИ-95", "fuel", "oil_chem"),
    ("diesel_fuel", "Дизельное топливо", "fuel", "oil_chem"),
    ("lpg", "Автогаз", "fuel", "oil_chem"),
    ("heating_fuel", "Печное топливо", "fuel", "oil_chem"),
]

assert len(RAW) == 40, len(RAW)
assert len(INTER) == 55, len(INTER)
assert len(FINISHED) == 75, len(FINISHED)

GEOLOGICAL = {
    "mine", "quarry", "oil_well", "gas_well", "peat_field",
}
RENEWABLE = {
    "grain_farm", "crop_farm", "oilseed_farm", "fiber_farm", "vegetable_farm",
    "dairy_farm", "livestock_farm", "poultry_farm", "sheep_farm", "logging_camp",
    "forest_gathering", "apiary", "orchard", "horse_farm", "greenhouse",
}

EXTRACT_BLD = {
    "mine": ("Шахта", "industrial", 45000, "metal"),
    "quarry": ("Карьер", "industrial", 28000, "mining"),
    "oil_well": ("Нефтяная скважина", "industrial", 60000, "oil_chem"),
    "gas_well": ("Газовая скважина", "industrial", 55000, "oil_chem"),
    "peat_field": ("Торфоразработка", "rural", 18000, "mining"),
    "grain_farm": ("Зерновое хозяйство", "rural", 14000, "agro"),
    "crop_farm": ("Полевое хозяйство", "rural", 12000, "agro"),
    "oilseed_farm": ("Масличная ферма", "rural", 13000, "agro"),
    "fiber_farm": ("Льноводческая ферма", "rural", 12000, "textile"),
    "vegetable_farm": ("Овощная ферма", "rural", 11000, "agro"),
    "dairy_farm": ("Молочная ферма", "rural", 16000, "dairy"),
    "livestock_farm": ("Мясная ферма", "rural", 18000, "dairy"),
    "poultry_farm": ("Птицефабрика", "rural", 14000, "dairy"),
    "sheep_farm": ("Овцеводческая ферма", "rural", 13000, "textile"),
    "logging_camp": ("Лесозаготовка", "rural", 15000, "timber"),
    "forest_gathering": ("Лесной промысел", "rural", 9000, "bashkir"),
    "apiary": ("Пасека", "rural", 10000, "bashkir"),
    "orchard": ("Плодовый сад", "rural", 14000, "agro"),
    "horse_farm": ("Коневодческая ферма", "rural", 20000, "bashkir"),
}

FACTORY_BLD = {
    "blast_furnace": ("Доменная печь", 70000, "metal"),
    "steel_plant": ("Сталеплавильный завод", 80000, "metal"),
    "smelter": ("Медеплавильный цех", 55000, "metal"),
    "aluminum_plant": ("Алюминиевый завод", 85000, "metal"),
    "rolling_mill": ("Прокатный стан", 60000, "metal"),
    "pipe_plant": ("Трубный завод", 55000, "metal"),
    "wire_plant": ("Кабельный цех", 40000, "metal"),
    "profile_plant": ("Профильный цех", 42000, "metal"),
    "machine_works": ("Машиностроительный цех", 75000, "metal"),
    "refinery": ("НПЗ", 120000, "oil_chem"),
    "refinery_deep": ("Глубокая переработка", 140000, "oil_chem"),
    "petrochemical_plant": ("Нефтехимия", 100000, "oil_chem"),
    "chemical_plant": ("Химзавод", 70000, "oil_chem"),
    "cement_plant": ("Цементный завод", 50000, "mining"),
    "concrete_plant": ("Бетонный узел", 35000, "mining"),
    "brickworks": ("Кирпичный завод", 30000, "mining"),
    "glass_plant": ("Стекольный завод", 45000, "packaging"),
    "insulation_plant": ("Завод утеплителя", 40000, "mining"),
    "ceramics_plant": ("Керамический завод", 38000, "mining"),
    "paint_plant": ("Лакокрасочный завод", 36000, "oil_chem"),
    "dry_mix_plant": ("Завод сухих смесей", 32000, "mining"),
    "flour_mill": ("Мельница", 22000, "agro"),
    "grain_mill": ("Крупозавод", 20000, "agro"),
    "sugar_plant": ("Сахарный завод", 48000, "agro"),
    "oil_press": ("Маслобойня", 25000, "agro"),
    "dairy_plant": ("Молокозавод", 35000, "dairy"),
    "meat_plant": ("Мясокомбинат", 40000, "dairy"),
    "fruit_plant": ("Плодоперерабока", 28000, "agro"),
    "fiber_plant": ("Волокноцех", 24000, "textile"),
    "spinning_mill": ("Прядильня", 26000, "textile"),
    "textile_mill": ("Ткацкая фабрика", 32000, "textile"),
    "knitwear": ("Трикотажная фабрика", 30000, "textile"),
    "leather_works": ("Кожзавод", 34000, "textile"),
    "textile_dye": ("Красильня", 22000, "textile"),
    "sawmill": ("Лесопилка", 20000, "timber"),
    "wood_plant": ("Деревообработка", 28000, "timber"),
    "paper_mill": ("Бумкомбинат", 45000, "timber"),
    "furniture_plant": ("Мебельная фабрика", 38000, "timber"),
    "engine_plant": ("Моторный завод", 90000, "metal"),
    "electronics_plant": ("Электронное производство", 70000, "tech"),
    "chip_fab": ("Микросборочный цех", 95000, "tech"),
    "battery_plant": ("Аккумуляторный завод", 65000, "tech"),
    "plastics_plant": ("Пластмассовый завод", 40000, "oil_chem"),
    "bakery": ("Пекарня", 18000, "agro"),
    "confectionery": ("Кондитерская", 22000, "agro"),
    "pasta_plant": ("Макаронная фабрика", 24000, "agro"),
    "packaging_plant": ("Фасовочный цех", 16000, "packaging"),
    "food_plant": ("Пищекомбинат", 30000, "agro"),
    "bottling_plant": ("Розлив", 20000, "agro"),
    "cheese_plant": ("Сырзавод", 32000, "dairy"),
    "water_plant": ("Водоочистка / розлив воды", 25000, "agro"),
    "juice_plant": ("Соковый завод", 28000, "agro"),
    "beverage_plant": ("Напитковый завод", 26000, "agro"),
    "fermentation_plant": ("Бродильный цех", 24000, "bashkir"),
    "garment_factory": ("Швейная фабрика", 28000, "textile"),
    "workwear_factory": ("Фабрика спецодежды", 26000, "textile"),
    "shoe_factory": ("Обувная фабрика", 30000, "textile"),
    "building_packaging": ("Стройфасовка", 18000, "mining"),
    "window_plant": ("Оконный завод", 42000, "mining"),
    "tire_plant": ("Шинный завод", 70000, "oil_chem"),
    "lubricant_plant": ("Маслозавод", 35000, "oil_chem"),
    "auto_parts_plant": ("Автокомпоненты", 55000, "metal"),
    "electronics_assembly": ("Сборка электроники", 80000, "tech"),
    "appliance_plant": ("Бытовая техника", 85000, "tech"),
    "honey_packaging": ("Фасовка мёда", 14000, "bashkir"),
    "gift_packaging": ("Подарочная упаковка", 12000, "bashkir"),
    "souvenir_workshop": ("Сувенирная мастерская", 15000, "bashkir"),
    "gas_processing": ("Газопереработка", 90000, "oil_chem"),
}

LOGISTICS = {
    "warehouse": ("Склад", 15000, "retail"),
    "garage": ("Автопарк", 20000, "retail"),
    "cold_storage": ("Холодильник", 28000, "retail"),
    "port_terminal": ("Логистический хаб", 90000, "retail"),
    "rail_depot": ("Ж/д депо", 70000, "retail"),
}


def item(iid, name, tier, category, branch):
    return {
        "id": iid, "name": name, "tier": tier, "category": category,
        "branch": branch, "tier_unlock": 0, "icon": f"item_{iid}",
        "quality_model": "batch_1_5",
    }


def recipe(rid, name, bld, inputs, out_id, qty, labor, energy, branch, tier=0, research=0,
           season=None, water=0.0, utility=0.0, extract_mode=""):
    r = {
        "id": rid, "name": name, "building_type": bld,
        "inputs": [{"item_id": a, "qty": b} for a, b in inputs],
        "output": {"item_id": out_id, "qty": qty},
        "labor": labor, "energy": energy, "research_cost": research,
        "branch": branch, "tier_unlock": tier,
        "season_bonus_months": season or [],
        "water_cost": water, "utility_cost": utility,
    }
    if extract_mode:
        r["extract_mode"] = extract_mode  # depleting | renewable
    return r


def L(w=2, e=0, m=0):
    d = {"worker": w}
    if e:
        d["engineer"] = e
    if m:
        d["manager"] = m
    return d


def build_recipes():
    rs = []
    # 40 extract
    for iid, name, cat, bld, branch in RAW:
        mode = "depleting" if bld in GEOLOGICAL else "renewable"
        season = []
        if bld in ("grain_farm", "crop_farm", "oilseed_farm", "fiber_farm", "vegetable_farm", "orchard", "apiary"):
            season = [6, 7, 8] if bld != "grain_farm" else [7, 8, 9]
        qty = 4 if mode == "renewable" else 3
        if iid in ("crude_oil", "natural_gas", "iron_ore", "coal"):
            qty = 4
        energy = 3 if mode == "depleting" else 1
        labor = L(4, 1) if mode == "depleting" else L(3)
        rs.append(recipe(
            f"extract_{iid}", f"Добыча: {name}", bld, [], iid, qty, labor, energy, branch,
            season=season, extract_mode=mode, utility=5.0 if mode == "renewable" else 8.0,
        ))

    # 55 intermediates
    mid = [
        ("met_01", "Чугун", "blast_furnace", [("iron_ore", 4), ("coal", 2)], "pig_iron", 3, L(4, 2), 6, "metal"),
        ("met_02", "Сталь", "steel_plant", [("pig_iron", 3), ("limestone", 1)], "steel", 3, L(4, 2), 5, "metal"),
        ("met_03", "Медь", "smelter", [("copper_ore", 4), ("coal", 1)], "copper", 2, L(3, 1), 5, "metal"),
        ("met_04", "Алюминий", "aluminum_plant", [("bauxite", 4)], "aluminum", 2, L(3, 2), 7, "metal", 1, 0, 0, 12.0),
        ("met_05", "Стальной лист", "rolling_mill", [("steel", 3)], "steel_sheet", 3, L(3, 1), 4, "metal"),
        ("met_06", "Стальная труба", "pipe_plant", [("steel", 2)], "steel_pipe", 2, L(3, 1), 4, "metal"),
        ("met_07", "Медный провод", "wire_plant", [("copper", 2)], "copper_wire", 4, L(2, 1), 3, "metal"),
        ("met_08", "Алюминиевый профиль", "profile_plant", [("aluminum", 2)], "aluminum_profile", 3, L(2, 1), 3, "metal"),
        ("met_09", "Металлические детали", "machine_works", [("steel", 2), ("copper", 1)], "metal_parts", 3, L(3, 2), 4, "metal"),
        ("oil_01", "Бензиновая фракция", "refinery", [("crude_oil", 5)], "gasoline_base", 3, L(4, 2), 6, "oil_chem"),
        ("oil_02", "Дизельная фракция", "refinery", [("crude_oil", 5)], "diesel_base", 3, L(4, 2), 6, "oil_chem"),
        ("oil_03", "Мазут", "refinery", [("crude_oil", 4)], "fuel_oil", 3, L(3, 2), 5, "oil_chem"),
        ("oil_04", "Битум", "refinery", [("crude_oil", 3), ("limestone", 1)], "bitumen", 2, L(3, 1), 5, "oil_chem"),
        ("oil_05", "Полимеры", "petrochemical_plant", [("crude_oil", 3), ("natural_gas", 2)], "polymer_granules", 3, L(3, 2), 6, "oil_chem", 1),
        ("oil_06", "Каучук", "petrochemical_plant", [("polymer_granules", 2), ("natural_gas", 1)], "synthetic_rubber", 2, L(3, 2), 5, "oil_chem", 1),
        ("oil_07", "Растворитель", "chemical_plant", [("crude_oil", 2), ("natural_gas", 1)], "industrial_solvent", 3, L(2, 2), 4, "oil_chem"),
        ("oil_08", "Химреагент", "chemical_plant", [("natural_gas", 2), ("salt", 1)], "chemical_reagent", 3, L(2, 2), 4, "oil_chem"),
        ("bld_01", "Цемент", "cement_plant", [("limestone", 3), ("clay", 1), ("coal", 1)], "cement", 4, L(3, 1), 5, "mining"),
        ("bld_02", "Бетон", "concrete_plant", [("cement", 2), ("sand", 2), ("stone", 2)], "concrete", 5, L(3), 3, "mining", 0, 0, 4.0),
        ("bld_03", "Кирпич", "brickworks", [("clay", 3), ("coal", 1)], "brick", 4, L(3), 4, "mining"),
        ("bld_04", "Стекло", "glass_plant", [("sand", 3), ("quartz", 1), ("limestone", 1)], "glass_sheet", 4, L(3, 1), 5, "packaging"),
        ("bld_05", "Утеплитель", "insulation_plant", [("glass_sheet", 2), ("polymer_granules", 1)], "insulation", 3, L(2, 1), 4, "mining", 1),
        ("bld_06", "Плитка", "ceramics_plant", [("clay", 2), ("quartz", 1)], "ceramic_tile", 3, L(2, 1), 3, "mining"),
        ("bld_07", "Основа краски", "paint_plant", [("chemical_reagent", 2), ("industrial_solvent", 1)], "paint_base", 3, L(2, 1), 3, "oil_chem"),
        ("bld_08", "Сухая смесь", "dry_mix_plant", [("cement", 2), ("sand", 1), ("limestone", 1)], "dry_mix", 4, L(2), 3, "mining"),
        ("food_01", "Пшеничная мука", "flour_mill", [("wheat", 4)], "wheat_flour", 3, L(2), 2, "agro"),
        ("food_02", "Ржаная мука", "flour_mill", [("rye", 4)], "rye_flour", 3, L(2), 2, "agro"),
        ("food_03", "Крупа", "grain_mill", [("buckwheat", 4)], "groats", 3, L(2), 2, "agro"),
        ("food_04", "Сахар", "sugar_plant", [("sugar_beet", 5)], "sugar", 3, L(3, 1), 4, "agro"),
        ("food_05", "Растительное масло", "oil_press", [("sunflower", 4)], "sunflower_oil", 3, L(2), 3, "agro"),
        ("food_06", "Сливки", "dairy_plant", [("milk", 4)], "cream", 3, L(2), 2, "dairy"),
        ("food_07", "Масляная основа", "dairy_plant", [("cream", 3)], "butter_base", 2, L(2), 2, "dairy"),
        ("food_08", "Фарш", "meat_plant", [("beef", 2), ("pork", 2)], "minced_meat", 3, L(3), 3, "dairy"),
        ("food_09", "Фруктовое пюре", "fruit_plant", [("apples", 3), ("cherries", 2)], "fruit_puree", 4, L(2), 2, "agro"),
        ("food_10", "Кисломолочная основа", "dairy_plant", [("milk", 4)], "fermented_milk", 3, L(2), 2, "dairy"),
        ("tex_01", "Льняное волокно", "fiber_plant", [("flax", 4)], "flax_fiber", 3, L(2), 2, "textile"),
        ("tex_02", "Пряжа", "spinning_mill", [("flax_fiber", 3), ("wool", 1)], "yarn", 3, L(2, 1), 3, "textile"),
        ("tex_03", "Ткань", "textile_mill", [("yarn", 3)], "woven_fabric", 3, L(3, 1), 3, "textile"),
        ("tex_04", "Трикотаж", "knitwear", [("yarn", 3)], "knitted_fabric", 3, L(3, 1), 3, "textile"),
        ("tex_05", "Выделанная кожа", "leather_works", [("raw_hide", 4), ("chemical_reagent", 1)], "tanned_leather", 3, L(3, 1), 4, "textile"),
        ("tex_06", "Наполнитель", "textile_mill", [("wool", 2), ("wood_chips", 1)], "textile_filling", 3, L(2), 2, "textile"),
        ("tex_07", "Окрашенная ткань", "textile_dye", [("woven_fabric", 3), ("chemical_reagent", 1)], "dyed_fabric", 3, L(2, 1), 3, "textile"),
        ("wood_01", "Доски", "sawmill", [("softwood", 4)], "lumber", 3, L(3), 3, "timber"),
        ("wood_02", "Фанера", "wood_plant", [("lumber", 3), ("chemical_reagent", 1)], "plywood", 3, L(2, 1), 3, "timber"),
        ("wood_03", "ДСП", "wood_plant", [("wood_chips", 2), ("chemical_reagent", 1)], "particle_board", 3, L(2, 1), 3, "timber"),
        ("wood_04", "Бумага", "paper_mill", [("wood_chips", 3)], "paper", 3, L(3, 1), 4, "timber", 0, 0, 3.0),
        ("wood_05", "Картон", "paper_mill", [("paper", 3)], "cardboard", 3, L(2), 2, "timber"),
        ("wood_06", "Мебельные детали", "furniture_plant", [("lumber", 2), ("particle_board", 1)], "furniture_parts", 3, L(3, 1), 3, "timber"),
        ("tech_01", "Двигатель", "engine_plant", [("steel", 2), ("metal_parts", 2), ("aluminum", 1)], "engine", 1, L(3, 3, 1), 6, "metal", 2),
        ("tech_02", "Жгут проводки", "electronics_plant", [("copper_wire", 2), ("polymer_granules", 1)], "wiring_harness", 3, L(2, 2), 3, "tech", 1),
        ("tech_03", "Печатная плата", "chip_fab", [("copper", 1), ("quartz", 1), ("polymer_granules", 1)], "circuit_board", 2, L(1, 3), 5, "tech", 2),
        ("tech_04", "Электронный модуль", "electronics_plant", [("circuit_board", 2), ("copper_wire", 1)], "electronic_module", 2, L(2, 2), 4, "tech", 2),
        ("tech_05", "Аккумуляторная ячейка", "battery_plant", [("aluminum", 1), ("copper", 1), ("chemical_reagent", 1)], "battery_cell", 2, L(2, 2), 4, "tech", 1),
        ("tech_06", "Пластиковые детали", "plastics_plant", [("polymer_granules", 3)], "plastic_parts", 4, L(2, 1), 3, "oil_chem", 1),
        ("tech_07", "Механический узел", "machine_works", [("metal_parts", 2), ("wiring_harness", 1)], "mechanical_assembly", 2, L(3, 2), 4, "metal", 1),
    ]
    for row in mid:
        rid, name, bld, inputs, out, qty, labor, energy, branch = row[:9]
        tier = row[9] if len(row) > 9 else 0
        research = row[10] if len(row) > 10 else 0
        water = row[11] if len(row) > 11 else 0.0
        utility = row[12] if len(row) > 12 else 2.0
        # unpack carefully - I mixed optional positions
        # Fix: use explicit kwargs style in list instead
        pass

    # Rebuild mid with consistent tuples: id,name,bld,inputs,out,qty,labor,energy,branch,tier,research,water,utility
    mid_fixed = [
        ("met_01", "Чугун", "blast_furnace", [("iron_ore", 4), ("coal", 2)], "pig_iron", 3, L(4, 2), 6, "metal", 0, 0, 0, 2),
        ("met_02", "Сталь", "steel_plant", [("pig_iron", 3), ("limestone", 1)], "steel", 3, L(4, 2), 5, "metal", 0, 0, 0, 2),
        ("met_03", "Медь", "smelter", [("copper_ore", 4), ("coal", 1)], "copper", 2, L(3, 1), 5, "metal", 0, 0, 0, 2),
        ("met_04", "Алюминий", "aluminum_plant", [("bauxite", 4)], "aluminum", 2, L(3, 2), 7, "metal", 1, 0, 0, 12),
        ("met_05", "Стальной лист", "rolling_mill", [("steel", 3)], "steel_sheet", 3, L(3, 1), 4, "metal", 0, 0, 0, 2),
        ("met_06", "Стальная труба", "pipe_plant", [("steel", 2)], "steel_pipe", 2, L(3, 1), 4, "metal", 0, 0, 0, 2),
        ("met_07", "Медный провод", "wire_plant", [("copper", 2)], "copper_wire", 4, L(2, 1), 3, "metal", 0, 0, 0, 2),
        ("met_08", "Алюминиевый профиль", "profile_plant", [("aluminum", 2)], "aluminum_profile", 3, L(2, 1), 3, "metal", 0, 0, 0, 2),
        ("met_09", "Металлические детали", "machine_works", [("steel", 2), ("copper", 1)], "metal_parts", 3, L(3, 2), 4, "metal", 0, 0, 0, 2),
        ("oil_01", "Бензиновая фракция", "refinery", [("crude_oil", 5)], "gasoline_base", 3, L(4, 2), 6, "oil_chem", 0, 0, 0, 3),
        ("oil_02", "Дизельная фракция", "refinery", [("crude_oil", 5)], "diesel_base", 3, L(4, 2), 6, "oil_chem", 0, 0, 0, 3),
        ("oil_03", "Мазут", "refinery", [("crude_oil", 4)], "fuel_oil", 3, L(3, 2), 5, "oil_chem", 0, 0, 0, 3),
        ("oil_04", "Битум", "refinery", [("crude_oil", 3), ("limestone", 1)], "bitumen", 2, L(3, 1), 5, "oil_chem", 0, 0, 0, 3),
        ("oil_05", "Полимеры", "petrochemical_plant", [("crude_oil", 3), ("natural_gas", 2)], "polymer_granules", 3, L(3, 2), 6, "oil_chem", 1, 0, 0, 3),
        ("oil_06", "Каучук", "petrochemical_plant", [("polymer_granules", 2), ("natural_gas", 1)], "synthetic_rubber", 2, L(3, 2), 5, "oil_chem", 1, 0, 0, 3),
        ("oil_07", "Растворитель", "chemical_plant", [("crude_oil", 2), ("natural_gas", 1)], "industrial_solvent", 3, L(2, 2), 4, "oil_chem", 0, 0, 0, 2),
        ("oil_08", "Химреагент", "chemical_plant", [("natural_gas", 2), ("salt", 1)], "chemical_reagent", 3, L(2, 2), 4, "oil_chem", 0, 0, 0, 2),
        ("bld_01", "Цемент", "cement_plant", [("limestone", 3), ("clay", 1), ("coal", 1)], "cement", 4, L(3, 1), 5, "mining", 0, 0, 0, 2),
        ("bld_02", "Бетон", "concrete_plant", [("cement", 2), ("sand", 2), ("stone", 2)], "concrete", 5, L(3), 3, "mining", 0, 0, 4, 2),
        ("bld_03", "Кирпич", "brickworks", [("clay", 3), ("coal", 1)], "brick", 4, L(3), 4, "mining", 0, 0, 0, 2),
        ("bld_04", "Стекло", "glass_plant", [("sand", 3), ("quartz", 1), ("limestone", 1)], "glass_sheet", 4, L(3, 1), 5, "packaging", 0, 0, 0, 2),
        ("bld_05", "Утеплитель", "insulation_plant", [("glass_sheet", 2), ("polymer_granules", 1)], "insulation", 3, L(2, 1), 4, "mining", 1, 0, 0, 2),
        ("bld_06", "Плитка", "ceramics_plant", [("clay", 2), ("quartz", 1)], "ceramic_tile", 3, L(2, 1), 3, "mining", 0, 0, 0, 2),
        ("bld_07", "Основа краски", "paint_plant", [("chemical_reagent", 2), ("industrial_solvent", 1)], "paint_base", 3, L(2, 1), 3, "oil_chem", 0, 0, 0, 2),
        ("bld_08", "Сухая смесь", "dry_mix_plant", [("cement", 2), ("sand", 1), ("limestone", 1)], "dry_mix", 4, L(2), 3, "mining", 0, 0, 0, 2),
        ("food_01", "Пшеничная мука", "flour_mill", [("wheat", 4)], "wheat_flour", 3, L(2), 2, "agro", 0, 0, 0, 1),
        ("food_02", "Ржаная мука", "flour_mill", [("rye", 4)], "rye_flour", 3, L(2), 2, "agro", 0, 0, 0, 1),
        ("food_03", "Крупа", "grain_mill", [("buckwheat", 4)], "groats", 3, L(2), 2, "agro", 0, 0, 0, 1),
        ("food_04", "Сахар", "sugar_plant", [("sugar_beet", 5)], "sugar", 3, L(3, 1), 4, "agro", 0, 0, 0, 2),
        ("food_05", "Растительное масло", "oil_press", [("sunflower", 4)], "sunflower_oil", 3, L(2), 3, "agro", 0, 0, 0, 1),
        ("food_06", "Сливки", "dairy_plant", [("milk", 4)], "cream", 3, L(2), 2, "dairy", 0, 0, 0, 1),
        ("food_07", "Масляная основа", "dairy_plant", [("cream", 3)], "butter_base", 2, L(2), 2, "dairy", 0, 0, 0, 1),
        ("food_08", "Фарш", "meat_plant", [("beef", 2), ("pork", 2)], "minced_meat", 3, L(3), 3, "dairy", 0, 0, 0, 1),
        ("food_09", "Фруктовое пюре", "fruit_plant", [("apples", 3), ("cherries", 2)], "fruit_puree", 4, L(2), 2, "agro", 0, 0, 0, 1),
        ("food_10", "Кисломолочная основа", "dairy_plant", [("milk", 4)], "fermented_milk", 3, L(2), 2, "dairy", 0, 0, 0, 1),
        ("tex_01", "Льняное волокно", "fiber_plant", [("flax", 4)], "flax_fiber", 3, L(2), 2, "textile", 0, 0, 0, 1),
        ("tex_02", "Пряжа", "spinning_mill", [("flax_fiber", 3), ("wool", 1)], "yarn", 3, L(2, 1), 3, "textile", 0, 0, 0, 1),
        ("tex_03", "Ткань", "textile_mill", [("yarn", 3)], "woven_fabric", 3, L(3, 1), 3, "textile", 0, 0, 0, 1),
        ("tex_04", "Трикотаж", "knitwear", [("yarn", 3)], "knitted_fabric", 3, L(3, 1), 3, "textile", 0, 0, 0, 1),
        ("tex_05", "Выделанная кожа", "leather_works", [("raw_hide", 4), ("chemical_reagent", 1)], "tanned_leather", 3, L(3, 1), 4, "textile", 0, 0, 0, 2),
        ("tex_06", "Наполнитель", "textile_mill", [("wool", 2), ("wood_chips", 1)], "textile_filling", 3, L(2), 2, "textile", 0, 0, 0, 1),
        ("tex_07", "Окрашенная ткань", "textile_dye", [("woven_fabric", 3), ("chemical_reagent", 1)], "dyed_fabric", 3, L(2, 1), 3, "textile", 0, 0, 0, 1),
        ("wood_01", "Доски", "sawmill", [("softwood", 4)], "lumber", 3, L(3), 3, "timber", 0, 0, 0, 1),
        ("wood_02", "Фанера", "wood_plant", [("lumber", 3), ("chemical_reagent", 1)], "plywood", 3, L(2, 1), 3, "timber", 0, 0, 0, 2),
        ("wood_03", "ДСП", "wood_plant", [("wood_chips", 2), ("chemical_reagent", 1)], "particle_board", 3, L(2, 1), 3, "timber", 0, 0, 0, 2),
        ("wood_04", "Бумага", "paper_mill", [("wood_chips", 3)], "paper", 3, L(3, 1), 4, "timber", 0, 0, 3, 2),
        ("wood_05", "Картон", "paper_mill", [("paper", 3)], "cardboard", 3, L(2), 2, "timber", 0, 0, 0, 1),
        ("wood_06", "Мебельные детали", "furniture_plant", [("lumber", 2), ("particle_board", 1)], "furniture_parts", 3, L(3, 1), 3, "timber", 0, 0, 0, 1),
        ("tech_01", "Двигатель", "engine_plant", [("steel", 2), ("metal_parts", 2), ("aluminum", 1)], "engine", 1, L(3, 3, 1), 6, "metal", 2, 100, 0, 4),
        ("tech_02", "Жгут проводки", "electronics_plant", [("copper_wire", 2), ("polymer_granules", 1)], "wiring_harness", 3, L(2, 2), 3, "tech", 1, 0, 0, 2),
        ("tech_03", "Печатная плата", "chip_fab", [("copper", 1), ("quartz", 1), ("polymer_granules", 1)], "circuit_board", 2, L(1, 3), 5, "tech", 2, 150, 0, 3),
        ("tech_04", "Электронный модуль", "electronics_plant", [("circuit_board", 2), ("copper_wire", 1)], "electronic_module", 2, L(2, 2), 4, "tech", 2, 0, 0, 2),
        ("tech_05", "Аккумуляторная ячейка", "battery_plant", [("aluminum", 1), ("copper", 1), ("chemical_reagent", 1)], "battery_cell", 2, L(2, 2), 4, "tech", 1, 0, 0, 2),
        ("tech_06", "Пластиковые детали", "plastics_plant", [("polymer_granules", 3)], "plastic_parts", 4, L(2, 1), 3, "oil_chem", 1, 0, 0, 2),
        ("tech_07", "Механический узел", "machine_works", [("metal_parts", 2), ("wiring_harness", 1)], "mechanical_assembly", 2, L(3, 2), 4, "metal", 1, 0, 0, 2),
    ]
    assert len(mid_fixed) == 55
    for rid, name, bld, inputs, out, qty, labor, energy, branch, tier, research, water, utility in mid_fixed:
        rs.append(recipe(rid, name, bld, inputs, out, qty, labor, energy, branch, tier, research, water=water, utility=utility))

    fin = [
        ("food_11", "Хлеб пшеничный", "bakery", [("wheat_flour", 3)], "wheat_bread", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("food_12", "Хлеб ржаной", "bakery", [("rye_flour", 3)], "rye_bread", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("food_13", "Булочки", "bakery", [("wheat_flour", 2), ("sugar", 1)], "bun", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("food_14", "Печенье", "confectionery", [("wheat_flour", 2), ("sugar", 1), ("butter_base", 1)], "cookies", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("food_15", "Макароны", "pasta_plant", [("wheat_flour", 3), ("eggs", 1)], "pasta", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("food_16", "Гречка фасованная", "packaging_plant", [("groats", 3)], "buckwheat_pack", 3, L(1), 1, "agro", 0, 0, 0, 4),
        ("food_17", "Конина фасованная", "meat_plant", [("horse_meat", 3)], "horse_meat_pack", 3, L(2), 2, "bashkir", 0, 0, 0, 3),
        ("food_18", "Картофель фасованный", "packaging_plant", [("potato", 4)], "potato_pack", 4, L(1), 1, "agro", 0, 0, 0, 3),
        ("food_19", "Овощная смесь", "food_plant", [("vegetables", 4), ("sunflower_oil", 1)], "vegetable_mix", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("food_20", "Масло бутылочное", "bottling_plant", [("sunflower_oil", 3)], "sunflower_oil_bottle", 3, L(2), 2, "agro", 0, 0, 0, 4),
        ("food_21", "Сахар фасованный", "packaging_plant", [("sugar", 3)], "sugar_pack", 3, L(1), 1, "agro", 0, 0, 0, 3),
        ("food_22", "Молоко питьевое", "dairy_plant", [("milk", 4)], "milk_bottle", 4, L(2), 2, "dairy", 0, 0, 0, 3),
        ("food_23", "Сливки фасованные", "dairy_plant", [("cream", 3)], "cream_pack", 3, L(2), 2, "dairy", 0, 0, 0, 3),
        ("food_24", "Масло сливочное", "dairy_plant", [("butter_base", 2)], "butter", 2, L(2), 2, "dairy", 0, 0, 0, 3),
        ("food_25", "Сыр", "cheese_plant", [("milk", 4), ("salt", 1)], "cheese", 3, L(2, 1), 3, "dairy", 0, 0, 0, 3),
        ("food_26", "Йогурт", "dairy_plant", [("fermented_milk", 3), ("fruit_puree", 1)], "yogurt", 4, L(2), 2, "dairy", 0, 0, 0, 3),
        ("food_27", "Говядина фасованная", "meat_plant", [("beef", 3)], "beef_pack", 3, L(2), 2, "dairy", 0, 0, 0, 3),
        ("food_28", "Свинина фасованная", "meat_plant", [("pork", 3)], "pork_pack", 3, L(2), 2, "dairy", 0, 0, 0, 3),
        ("food_29", "Курица фасованная", "meat_plant", [("poultry", 3)], "chicken_pack", 3, L(2), 2, "dairy", 0, 0, 0, 3),
        ("food_30", "Яйца фасованные", "packaging_plant", [("eggs", 4)], "eggs_pack", 4, L(1), 1, "dairy", 0, 0, 0, 3),
        ("drink_01", "Вода питьевая", "water_plant", [], "bottled_water", 4, L(2, 1), 3, "agro", 0, 0, 8, 5),
        ("drink_02", "Яблочный сок", "juice_plant", [("apples", 4)], "apple_juice", 3, L(2), 2, "agro", 0, 0, 1, 3),
        ("drink_03", "Вишнёвый сок", "juice_plant", [("cherries", 4)], "cherry_juice", 3, L(2), 2, "agro", 0, 0, 1, 3),
        ("drink_04", "Ягодный морс", "juice_plant", [("forest_berries", 3)], "berry_drink", 4, L(2), 2, "bashkir", 0, 0, 2, 3),
        ("drink_05", "Квас", "beverage_plant", [("rye_flour", 2), ("sugar", 1)], "kvass", 4, L(2), 2, "agro", 0, 0, 3, 3),
        ("drink_06", "Кумыс", "fermentation_plant", [("mare_milk", 4)], "koumiss", 3, L(2), 2, "bashkir", 0, 0, 0, 3),
        ("drink_07", "Кисломолочный напиток", "dairy_plant", [("fermented_milk", 3), ("honey", 1)], "dairy_drink", 4, L(2), 2, "dairy", 0, 0, 0, 3),
        ("app_01", "Футболка", "garment_factory", [("knitted_fabric", 2)], "tshirt", 3, L(2), 2, "textile", 0, 0, 0, 2),
        ("app_02", "Рубашка", "garment_factory", [("woven_fabric", 2)], "shirt", 2, L(2), 2, "textile", 0, 0, 0, 2),
        ("app_03", "Брюки", "garment_factory", [("dyed_fabric", 3)], "trousers", 2, L(2), 2, "textile", 0, 0, 0, 2),
        ("app_04", "Куртка", "garment_factory", [("dyed_fabric", 2), ("textile_filling", 1)], "jacket", 2, L(3), 3, "textile", 0, 0, 0, 2),
        ("app_05", "Свитер", "knitwear", [("knitted_fabric", 3)], "knit_sweater", 2, L(2), 2, "textile", 0, 0, 0, 2),
        ("app_06", "Спецодежда", "workwear_factory", [("dyed_fabric", 3), ("tanned_leather", 1)], "workwear", 2, L(3), 3, "textile", 0, 0, 0, 2),
        ("app_07", "Кожаная обувь", "shoe_factory", [("tanned_leather", 2), ("synthetic_rubber", 1)], "leather_shoes", 2, L(3, 1), 3, "textile", 1, 0, 0, 2),
        ("app_08", "Текстильная обувь", "shoe_factory", [("woven_fabric", 2), ("synthetic_rubber", 1)], "textile_shoes", 2, L(2), 2, "textile", 0, 0, 0, 2),
        ("app_09", "Зимняя шапка", "knitwear", [("knitted_fabric", 2), ("wool", 1)], "winter_hat", 3, L(2), 2, "textile", 0, 0, 0, 2),
        ("home_01", "Стол", "furniture_plant", [("furniture_parts", 2), ("lumber", 1)], "table", 1, L(3), 3, "timber", 0, 0, 0, 2),
        ("home_02", "Стул", "furniture_plant", [("furniture_parts", 2)], "chair", 2, L(2), 2, "timber", 0, 0, 0, 2),
        ("home_03", "Шкаф", "furniture_plant", [("furniture_parts", 3), ("plywood", 1)], "wardrobe", 1, L(3, 1), 3, "timber", 0, 0, 0, 2),
        ("home_04", "Кровать", "furniture_plant", [("furniture_parts", 3), ("plywood", 1)], "bed", 1, L(3, 1), 3, "timber", 0, 0, 0, 2),
        ("home_05", "Диван", "furniture_plant", [("furniture_parts", 2), ("textile_filling", 2), ("dyed_fabric", 1)], "sofa", 1, L(3, 1), 4, "timber", 1, 0, 0, 3),
        ("home_06", "Моющее средство", "chemical_plant", [("chemical_reagent", 2), ("industrial_solvent", 1)], "detergent", 3, L(2), 2, "oil_chem", 0, 0, 0, 2),
        ("home_07", "Посуда", "ceramics_plant", [("ceramic_tile", 2), ("glass_sheet", 1)], "dishware", 3, L(2), 2, "mining", 0, 0, 0, 2),
        ("home_08", "Подушка", "textile_mill", [("textile_filling", 2), ("woven_fabric", 1)], "pillow", 3, L(2), 2, "textile", 0, 0, 0, 2),
        ("build_01", "Цемент мешковой", "building_packaging", [("cement", 3)], "cement_bag", 3, L(1), 1, "mining", 0, 0, 0, 3),
        ("build_02", "Кирпич упаковка", "building_packaging", [("brick", 4)], "brick_pack", 4, L(1), 1, "mining", 0, 0, 0, 3),
        ("build_03", "Бетон товарный", "concrete_plant", [("concrete", 4)], "concrete_mix", 4, L(2), 2, "mining", 0, 0, 0, 2),
        ("build_04", "Оконный блок", "window_plant", [("glass_sheet", 2), ("aluminum_profile", 2)], "window_unit", 2, L(3, 1), 3, "mining", 1, 0, 0, 2),
        ("build_05", "Краска", "paint_plant", [("paint_base", 3)], "paint_can", 3, L(2), 2, "oil_chem", 0, 0, 0, 3),
        ("build_06", "Плитка упаковка", "building_packaging", [("ceramic_tile", 3)], "tile_pack", 3, L(1), 1, "mining", 0, 0, 0, 3),
        ("build_07", "Утеплитель упаковка", "building_packaging", [("insulation", 3)], "insulation_pack", 3, L(1), 1, "mining", 0, 0, 0, 3),
        ("build_08", "Комплект труб", "pipe_plant", [("steel_pipe", 3)], "pipe_set", 3, L(2), 2, "metal", 0, 0, 0, 2),
        ("auto_01", "Шина", "tire_plant", [("synthetic_rubber", 2), ("steel", 1)], "tire", 2, L(3, 1), 4, "oil_chem", 1, 0, 0, 2),
        ("auto_02", "Автоаккумулятор", "battery_plant", [("battery_cell", 3), ("plastic_parts", 1)], "car_battery", 1, L(2, 2), 3, "tech", 1, 0, 0, 2),
        ("auto_03", "Моторное масло", "lubricant_plant", [("fuel_oil", 2), ("chemical_reagent", 1)], "engine_oil", 3, L(2), 2, "oil_chem", 0, 0, 0, 3),
        ("auto_04", "Тормозные детали", "auto_parts_plant", [("steel", 2), ("metal_parts", 1)], "brake_parts", 3, L(3, 1), 3, "metal", 1, 0, 0, 2),
        ("auto_05", "Фильтры", "auto_parts_plant", [("paper", 1), ("plastic_parts", 1), ("metal_parts", 1)], "filter_set", 3, L(2), 2, "metal", 0, 0, 0, 2),
        ("auto_06", "Кузовные детали", "auto_parts_plant", [("steel_sheet", 2), ("plastic_parts", 1)], "car_body_parts", 2, L(3, 1), 3, "metal", 1, 0, 0, 2),
        ("auto_07", "Комплект запчастей", "auto_parts_plant", [("mechanical_assembly", 2), ("wiring_harness", 1)], "car_spare_kit", 2, L(3, 2), 4, "metal", 1, 0, 0, 2),
        ("elec_01", "Смартфон", "electronics_assembly", [("electronic_module", 2), ("battery_cell", 1), ("glass_sheet", 1), ("plastic_parts", 1)], "smartphone", 1, L(2, 3, 1), 5, "tech", 2, 200, 0, 4),
        ("elec_02", "Телевизор", "electronics_assembly", [("electronic_module", 2), ("glass_sheet", 2), ("plastic_parts", 1)], "television", 1, L(2, 3), 5, "tech", 2, 150, 0, 4),
        ("elec_03", "Холодильник", "appliance_plant", [("mechanical_assembly", 2), ("steel_sheet", 2), ("electronic_module", 1)], "refrigerator", 1, L(3, 2), 5, "tech", 2, 100, 0, 3),
        ("elec_04", "Стиральная машина", "appliance_plant", [("mechanical_assembly", 2), ("steel_sheet", 2), ("electronic_module", 1)], "washing_machine", 1, L(3, 2), 5, "tech", 2, 100, 0, 3),
        ("elec_05", "Электрочайник", "appliance_plant", [("mechanical_assembly", 1), ("steel_sheet", 1), ("plastic_parts", 1)], "electric_kettle", 2, L(2, 1), 3, "tech", 1, 0, 0, 2),
        ("elec_06", "Домашний роутер", "electronics_assembly", [("circuit_board", 2), ("electronic_module", 1), ("plastic_parts", 1)], "home_router", 2, L(2, 2), 3, "tech", 2, 0, 0, 2),
        ("bash_01", "Башкирский мёд", "honey_packaging", [("honey", 3)], "bashkir_honey_pack", 3, L(1), 1, "bashkir", 0, 0, 0, 3),
        ("bash_02", "Ҡорот", "dairy_plant", [("fermented_milk", 3), ("salt", 1)], "qurut", 3, L(2), 2, "bashkir", 0, 0, 0, 2),
        ("bash_03", "Чак-чак", "confectionery", [("wheat_flour", 2), ("honey", 2), ("sugar", 1)], "chak_chak", 3, L(2), 2, "bashkir", 0, 0, 0, 3),
        ("bash_04", "Подарочный набор мёда", "gift_packaging", [("bashkir_honey_pack", 2), ("cardboard", 1)], "honey_gift_set", 1, L(1), 1, "bashkir", 0, 0, 0, 4),
        ("bash_05", "Сувенир", "souvenir_workshop", [("lumber", 1), ("dyed_fabric", 1), ("paper", 1)], "bashkir_souvenir", 3, L(2), 2, "bashkir", 0, 0, 0, 2),
        ("fuel_01", "АИ-92", "refinery", [("gasoline_base", 3)], "ai92", 3, L(2, 1), 3, "oil_chem", 0, 0, 0, 2),
        ("fuel_02", "АИ-95", "refinery_deep", [("gasoline_base", 3), ("chemical_reagent", 1)], "ai95", 3, L(2, 2), 4, "oil_chem", 1, 50, 0, 2),
        ("fuel_03", "Дизель", "refinery", [("diesel_base", 3)], "diesel_fuel", 3, L(2, 1), 3, "oil_chem", 0, 0, 0, 2),
        ("fuel_04", "Автогаз", "gas_processing", [("natural_gas", 3)], "lpg", 3, L(2, 1), 3, "oil_chem", 0, 0, 0, 2),
        ("fuel_05", "Печное топливо", "refinery", [("fuel_oil", 3)], "heating_fuel", 3, L(2), 2, "oil_chem", 0, 0, 0, 2),
    ]
    assert len(fin) == 75, len(fin)
    for rid, name, bld, inputs, out, qty, labor, energy, branch, tier, research, water, utility in fin:
        rs.append(recipe(rid, name, bld, inputs, out, qty, labor, energy, branch, tier, research, water=water, utility=utility))

    # 30 alternatives (normalized ALT-06/11/12/13)
    alts = [
        ("alt_01", "Ячневая крупа", "grain_mill", [("barley", 4)], "groats", 3, L(2), 2, "agro", 0, 0, 0, 1),
        ("alt_02", "Овсяная крупа", "grain_mill", [("oats", 4)], "groats", 3, L(2), 2, "agro", 0, 0, 0, 1),
        ("alt_03", "Смешанный хлеб", "bakery", [("wheat_flour", 2), ("rye_flour", 1)], "wheat_bread", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("alt_04", "Медовый ржаной", "bakery", [("rye_flour", 3), ("honey", 1)], "rye_bread", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("alt_05", "Грибная смесь", "food_plant", [("mushrooms", 2), ("vegetables", 2)], "vegetable_mix", 3, L(2), 2, "agro", 0, 0, 0, 3),
        # alt_06 horse_meat_pack already food_17
        ("alt_07", "Лиственные доски", "sawmill", [("hardwood", 4)], "lumber", 3, L(3), 3, "timber", 0, 0, 0, 1),
        ("alt_08", "Берёзовая фанера", "wood_plant", [("hardwood", 3), ("chemical_reagent", 1)], "plywood", 3, L(2, 1), 3, "timber", 0, 0, 0, 2),
        ("alt_09", "Цемент на торфе", "cement_plant", [("limestone", 2), ("clay", 1), ("peat", 1)], "cement", 3, L(3, 1), 5, "mining", 0, 0, 0, 2),
        ("alt_10", "Кирпич на торфе", "brickworks", [("clay", 3), ("peat", 1)], "brick", 4, L(3), 4, "mining", 0, 0, 0, 2),
        # alt_11: cement blend bag from dry_mix (named as cementitious mix sold as cement_bag)
        ("alt_11", "Цементная смесь в мешках", "building_packaging", [("dry_mix", 2), ("cement", 1)], "cement_bag", 3, L(1), 1, "mining", 0, 0, 0, 3),
        # alt_12: dry_mix wholesale pack as insulation? NO — use concrete_mix from dry_mix+water utility as construction bag
        ("alt_12", "Смесь для стяжки", "concrete_plant", [("dry_mix", 3), ("sand", 1)], "concrete_mix", 3, L(2), 2, "mining", 0, 0, 2, 2),
        # alt_13: road mix — bitumen + stone → heating_fuel? NO. Create pipe_set? Spec: bitumen road demand.
        # Represent as sellable road binder pack using bitumen+stone → concrete_mix FORBIDDEN.
        # Instead produce insulation_pack? Better: paint_can base from bitumen? Use engine_oil alt? 
        # Documented: bitumen consumed by road_works recipe producing heating_fuel? No.
        # Use: bitumen + stone → (utility road) we add recipe road_binder as dry_mix alternative sold B2B via building_market as dry_mix? 
        # Spec: if roads can't consume, B2B demand. We'll add recipe that produces dry_mix from bitumen+stone as "дорожная смесь" wrong.
        # Correct replacement: lubricant from bitumen path already exists. alt_13 = asphalt_patch as cement_bag? No.
        # Use car_body_parts? Simplest valid: bitumen + stone → brick? No.
        # Valid: bitumen used in paint_base alt
        ("alt_13", "Битумная основа краски", "paint_plant", [("bitumen", 2), ("industrial_solvent", 1)], "paint_base", 3, L(2, 1), 3, "oil_chem", 0, 0, 0, 2),
        ("alt_14", "Ремкомплект двигателя", "engine_plant", [("engine", 1), ("mechanical_assembly", 1)], "car_spare_kit", 2, L(3, 2), 4, "metal", 2, 0, 0, 2),
        ("alt_15", "Стиралка с мотором", "appliance_plant", [("engine", 1), ("steel_sheet", 1), ("electronic_module", 1)], "washing_machine", 1, L(3, 2), 5, "tech", 2, 0, 0, 3),
        ("alt_16", "Стол из массива", "furniture_plant", [("hardwood", 2), ("furniture_parts", 1)], "table", 1, L(3), 3, "timber", 0, 0, 0, 2),
        ("alt_17", "Стул с обивкой", "furniture_plant", [("plywood", 2), ("dyed_fabric", 1)], "chair", 1, L(2), 2, "timber", 0, 0, 0, 2),
        ("alt_18", "Тканая футболка", "garment_factory", [("woven_fabric", 2)], "tshirt", 3, L(2), 2, "textile", 0, 0, 0, 2),
        ("alt_19", "Демисезонная куртка", "garment_factory", [("knitted_fabric", 2), ("dyed_fabric", 1)], "jacket", 2, L(3), 3, "textile", 0, 0, 0, 2),
        ("alt_20", "Комбинированная обувь", "shoe_factory", [("tanned_leather", 1), ("woven_fabric", 1), ("synthetic_rubber", 1)], "leather_shoes", 2, L(3, 1), 3, "textile", 1, 0, 0, 2),
        ("alt_21", "Медовый йогурт", "dairy_plant", [("milk", 4), ("honey", 1)], "yogurt", 4, L(2), 2, "dairy", 0, 0, 0, 3),
        ("alt_22", "Сливочный сыр", "cheese_plant", [("milk", 3), ("cream", 1)], "cheese", 2, L(2, 1), 3, "dairy", 0, 0, 0, 3),
        ("alt_23", "Медовое печенье", "confectionery", [("wheat_flour", 2), ("sugar", 1), ("honey", 1)], "cookies", 4, L(2), 2, "agro", 0, 0, 0, 3),
        ("alt_24", "Ягодно-яблочный морс", "juice_plant", [("apples", 2), ("forest_berries", 2)], "berry_drink", 3, L(2), 2, "bashkir", 0, 0, 1, 3),
        ("alt_25", "Ячменный квас", "beverage_plant", [("barley", 2), ("sugar", 1)], "kvass", 4, L(2), 2, "agro", 0, 0, 3, 3),
        ("alt_26", "Газохимические полимеры", "petrochemical_plant", [("crude_oil", 2), ("natural_gas", 2)], "polymer_granules", 2, L(3, 2), 6, "oil_chem", 1, 0, 0, 3),
        ("alt_27", "Сталь из лома", "steel_plant", [("pig_iron", 2), ("metal_parts", 1)], "steel", 3, L(4, 2), 5, "metal", 1, 0, 0, 2),
        ("alt_28", "Высокопрозрачное стекло", "glass_plant", [("sand", 2), ("quartz", 2)], "glass_sheet", 3, L(3, 1), 5, "packaging", 1, 80, 0, 2),
        ("alt_29", "Автономный роутер", "electronics_assembly", [("circuit_board", 2), ("electronic_module", 1), ("battery_cell", 1)], "home_router", 1, L(2, 2), 3, "tech", 2, 0, 0, 2),
        ("alt_30", "Зимний дизель", "refinery_deep", [("diesel_base", 3), ("chemical_reagent", 1)], "diesel_fuel", 3, L(2, 2), 4, "oil_chem", 1, 40, 0, 2),
    ]
    # 29 alts (alt_06 folded into food_17) → total recipes = 40+55+75+29 = 199
    for rid, name, bld, inputs, out, qty, labor, energy, branch, tier, research, water, utility in alts:
        rs.append(recipe(rid, name, bld, inputs, out, qty, labor, energy, branch, tier, research, water=water, utility=utility))
    return rs


def build_shops():
    food20 = [x[0] for x in FINISHED[:20]]
    drinks = ["bottled_water", "apple_juice", "cherry_juice", "berry_drink", "kvass", "koumiss", "dairy_drink"]
    apparel = ["tshirt", "shirt", "trousers", "jacket", "knit_sweater", "workwear", "leather_shoes", "textile_shoes", "winter_hat"]
    home = ["table", "chair", "wardrobe", "bed", "sofa", "detergent", "dishware", "pillow"]
    build = ["cement_bag", "brick_pack", "concrete_mix", "window_unit", "paint_can", "tile_pack", "insulation_pack", "pipe_set"]
    auto = ["tire", "car_battery", "engine_oil", "brake_parts", "filter_set", "car_body_parts", "car_spare_kit"]
    elec = ["smartphone", "television", "refrigerator", "washing_machine", "electric_kettle", "home_router"]
    bash = ["bashkir_honey_pack", "qurut", "chak_chak", "honey_gift_set", "bashkir_souvenir"]
    fuel = ["ai92", "ai95", "diesel_fuel", "lpg"]
    return [
        {"id": "grocery", "name": "Продуктовый", "rent_mult": 1.0, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": food20 + ["bottled_water", "kvass", "apple_juice"], "icon": "shop_grocery", "markup": 1.25, "capacity": 80},
        {"id": "bakery_shop", "name": "Булочная", "rent_mult": 0.9, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": ["wheat_bread", "rye_bread", "bun", "cookies", "chak_chak"], "icon": "shop_bakery", "markup": 1.35, "capacity": 40},
        {"id": "dairy_shop", "name": "Молочный", "rent_mult": 1.05, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": ["milk_bottle", "cream_pack", "butter", "cheese", "yogurt", "dairy_drink", "koumiss", "qurut"], "icon": "shop_dairy", "markup": 1.3, "capacity": 50},
        {"id": "butcher", "name": "Мясной", "rent_mult": 1.1, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": ["beef_pack", "pork_pack", "chicken_pack", "eggs_pack", "horse_meat_pack"], "icon": "shop_butcher", "markup": 1.28, "capacity": 45},
        {"id": "beverage_shop", "name": "Напитки", "rent_mult": 1.0, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": drinks, "icon": "shop_beverage", "markup": 1.4, "capacity": 55},
        {"id": "clothing_shop", "name": "Одежда и обувь", "rent_mult": 1.2, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": apparel, "icon": "shop_clothing", "markup": 1.55, "capacity": 60},
        {"id": "furniture_home", "name": "Дом и мебель", "rent_mult": 1.15, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": home, "icon": "shop_furniture", "markup": 1.45, "capacity": 35},
        {"id": "building_market", "name": "Строймаркет", "rent_mult": 1.1, "allowed_categories": [], "allowed_tiers": ["finished", "intermediate"],
         "allowed_items": build + ["cement", "brick", "dry_mix", "concrete", "bitumen"], "icon": "shop_building", "markup": 1.22, "capacity": 100},
        {"id": "auto_parts_shop", "name": "Автотовары", "rent_mult": 1.15, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": auto, "icon": "shop_auto", "markup": 1.35, "capacity": 50},
        {"id": "electronics_shop", "name": "Электроника", "rent_mult": 1.35, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": elec, "icon": "shop_electronics", "markup": 1.5, "capacity": 40},
        {"id": "bashkir_shop", "name": "Башкирские товары", "rent_mult": 0.95, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": bash, "icon": "shop_bashkir", "markup": 1.6, "capacity": 30},
        {"id": "fuel_station", "name": "АЗС", "rent_mult": 1.4, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": fuel, "icon": "shop_fuel", "markup": 1.18, "capacity": 120},
        {"id": "fuel_depot", "name": "Топливный склад", "rent_mult": 1.0, "allowed_categories": [], "allowed_tiers": ["finished", "intermediate"],
         "allowed_items": ["heating_fuel", "ai92", "ai95", "diesel_fuel", "lpg", "fuel_oil", "gasoline_base", "diesel_base"], "icon": "shop_fuel_depot", "markup": 1.12, "capacity": 200},
        {"id": "hypermarket", "name": "Гипермаркет", "rent_mult": 1.5, "allowed_categories": [], "allowed_tiers": ["finished"],
         "allowed_items": food20[:12] + drinks[:4] + apparel[:4] + home[:4] + elec[:3], "icon": "shop_hyper", "markup": 1.2, "capacity": 150},
    ]


def write(name, obj):
    path = OUT / name
    path.write_text(json.dumps(obj, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    return path


def main():
    OUT.mkdir(parents=True, exist_ok=True)
    items = []
    for iid, name, cat, bld, branch in RAW:
        items.append(item(iid, name, "raw", cat, branch))
    for iid, name, cat, branch in INTER:
        items.append(item(iid, name, "intermediate", cat, branch))
    for iid, name, cat, branch in FINISHED:
        items.append(item(iid, name, "finished", cat, branch))

    buildings = []
    for bid, (name, zone, cost, branch) in EXTRACT_BLD.items():
        buildings.append({
            "id": bid, "name": name, "category": "extract", "industry_tag": branch,
            "build_cost": cost, "zone": zone, "footprint": 2, "branch": branch,
            "tier_unlock": 0, "icon": f"bld_{bid}",
            "resource_mode": "depleting" if bid in GEOLOGICAL else "renewable",
        })
    for bid, (name, cost, branch) in FACTORY_BLD.items():
        buildings.append({
            "id": bid, "name": name, "category": "factory", "industry_tag": branch,
            "build_cost": cost, "zone": "industrial", "footprint": 2, "branch": branch,
            "tier_unlock": 0, "icon": f"bld_{bid}",
        })
    for bid, (name, cost, branch) in LOGISTICS.items():
        buildings.append({
            "id": bid, "name": name, "category": "logistics", "industry_tag": "logistics",
            "build_cost": cost, "zone": "industrial", "footprint": 2, "branch": branch,
            "tier_unlock": 0, "icon": f"bld_{bid}",
        })

    recipes = build_recipes()
    shops = build_shops()

    cities = [
        {
            "id": "ufa", "name": "Уфа", "x": 420, "y": 310, "focus": "capital_refining_tech",
            "start_population": 12000, "start_radius": 4,
            "resources": [
                "crude_oil", "natural_gas", "sand", "clay",
                "wheat", "rye", "barley", "oats", "buckwheat", "potato", "sugar_beet", "sunflower", "flax", "vegetables",
                "milk", "beef", "pork", "poultry", "eggs", "wool", "raw_hide",
                "softwood", "hardwood", "wood_chips", "forest_berries", "mushrooms",
                "honey", "apples", "cherries", "horse_meat", "mare_milk",
            ],
            "labor": {"worker": 4000, "engineer": 900, "manager": 350}, "enabled": True,
        },
        {
            "id": "beloretsk", "name": "Белорецк", "x": 620, "y": 380, "focus": "metallurgy",
            "start_population": 5000, "start_radius": 3,
            "resources": [
                "iron_ore", "copper_ore", "bauxite", "coal", "limestone", "quartz", "stone", "sand", "clay",
                "softwood", "hardwood", "wood_chips",
            ],
            "labor": {"worker": 2200, "engineer": 500, "manager": 120}, "enabled": True,
        },
        {
            "id": "sterlitamak", "name": "Стерлитамак", "x": 480, "y": 420, "focus": "chemistry_building",
            "start_population": 7000, "start_radius": 3,
            "resources": ["salt", "limestone", "clay", "sand", "stone", "peat", "quartz"],
            "labor": {"worker": 2800, "engineer": 700, "manager": 180}, "enabled": True,
        },
    ]

    # Simple corridors (rebalance later)
    corridors = {}
    raw_mid = {
        "iron_ore": 22, "copper_ore": 28, "bauxite": 24, "limestone": 8, "sand": 6, "clay": 7,
        "quartz": 30, "stone": 7, "salt": 10, "crude_oil": 55, "natural_gas": 40, "coal": 16, "peat": 9,
        "wheat": 14, "rye": 13, "barley": 12, "oats": 11, "buckwheat": 22, "potato": 16, "sugar_beet": 9,
        "sunflower": 28, "flax": 20, "vegetables": 18, "milk": 28, "beef": 90, "pork": 70, "poultry": 45,
        "eggs": 12, "wool": 35, "raw_hide": 40, "softwood": 16, "hardwood": 19, "wood_chips": 8,
        "forest_berries": 40, "mushrooms": 38, "honey": 95, "apples": 22, "cherries": 35,
        "horse_meat": 85, "mare_milk": 42,
    }
    for it in items:
        if it["tier"] == "raw":
            mid = raw_mid.get(it["id"], 20)
            corridors[it["id"]] = {"min": round(mid * 0.55, 1), "max": round(mid * 1.45, 1)}
        elif it["tier"] == "intermediate":
            mid = 60
            corridors[it["id"]] = {"min": round(mid * 0.6, 1), "max": round(mid * 1.4, 1)}
        else:
            mid = 140
            corridors[it["id"]] = {"min": round(mid * 0.65, 1), "max": round(mid * 1.4, 1)}

    economy = {
        "economy_catalog_version": 2,
        "price_corridors": corridors,
        "city_market": [
            {"item_id": "crude_oil", "price": 48, "quality": 2.0, "stock_per_tick": 25},
            {"item_id": "natural_gas", "price": 35, "quality": 2.0, "stock_per_tick": 20},
            {"item_id": "wheat", "price": 10, "quality": 1.5, "stock_per_tick": 40},
            {"item_id": "coal", "price": 12, "quality": 1.8, "stock_per_tick": 30},
            {"item_id": "cement", "price": 40, "quality": 2.0, "stock_per_tick": 15},
            {"item_id": "dry_mix", "price": 35, "quality": 2.0, "stock_per_tick": 12},
            {"item_id": "bitumen", "price": 45, "quality": 2.0, "stock_per_tick": 10},
        ],
        "chat": {"max_messages": 120, "ai_messages_per_tick_cap": 2, "min_ticks_between_ai_msg": 2},
        "logistics": {"units_per_tick": 80, "cost_per_unit_dist": 0.012, "local_cost_per_unit": 0.8, "base_cost_per_unit": 1.2},
        "balance": {
            "version": 2, "region": "Bashkortostan", "year": 2026,
            "wage_per_tick": {"worker": 18.0, "engineer": 32.0, "manager": 48.0},
            "energy_cost_extract": 10.0, "energy_cost_factory": 18.0,
            "water_cost_per_unit": 2.5, "utility_cost_scale": 1.0,
            "target_margin": {"raw": 0.18, "intermediate": 0.16, "finished": 0.22},
            "notes": "Catalog V2 — quality is batch 1–5, not SKU clones.",
        },
        "demand": {
            "class_shares": {"budget": 0.35, "mid": 0.45, "premium": 0.2},
            "quality_weight": 0.35, "brand_weight": 0.2, "price_weight": 0.35, "distance_weight": 0.1,
            "brand_decay": 0.98, "brand_floor": 0.2, "brand_gain_on_sale": 0.02,
            "pop_demand_per_day": {"food": 0.08, "dairy": 0.03, "apparel": 0.015, "home": 0.02, "tech": 0.008, "fuel": 0.01, "mineral": 0.01},
            "too_expensive_at": 1.65,
        },
    }

    meta = {
        "economy_catalog_version": 2,
        "items": len(items),
        "raw": 40, "intermediate": 55, "finished": 75,
        "recipes": len(recipes),
        "buildings": len(buildings),
        "shops": len(shops),
        "notes": "oatmeal replaced by horse_meat_pack (ALT-06). ALT-12/13 normalized. Target ~200 recipes.",
    }

    # Migration maps (heuristic)
    item_map = {}
    legacy_items = json.loads((LEGACY / "items.json").read_text())
    v2_ids = {i["id"] for i in items}
    synonyms = {
        "oil": "crude_oil", "gas": "natural_gas", "wood": "softwood", "birch": "hardwood",
        "pine": "softwood", "firewood": "wood_chips", "apple": "apples", "cherry": "cherries",
        "egg": "eggs", "cattle": "beef", "pig": "pork", "sheep": "wool", "leather": "tanned_leather",
        "fabric_flax": "woven_fabric", "flour_wheat": "wheat_flour", "flour_rye": "rye_flour",
        "steel_pipe": "steel_pipe", "plastic": "plastic_parts", "circuit": "circuit_board",
        "chemicals": "chemical_reagent", "packaging": None, "bottle": None,
    }
    for it in legacy_items:
        lid = it["id"]
        base = lid
        for s in ("_export", "_prem", "_eco", "_std", "_ufa", "_sterli", "_beloretsk"):
            if base.endswith(s):
                base = base[: -len(s)]
        if base in v2_ids:
            item_map[lid] = {"action": "map_to_new", "new_id": base}
        elif base in synonyms:
            tgt = synonyms[base]
            if tgt is None:
                item_map[lid] = {"action": "convert_to_compensation", "reason": "utility_not_sku"}
            else:
                item_map[lid] = {"action": "map_to_new", "new_id": tgt}
        elif lid.startswith("bee"):
            item_map[lid] = {"action": "remove_only_if_unused", "reason": "misnamed_apiary"}
        else:
            item_map[lid] = {"action": "convert_to_compensation", "reason": "no_v2_equivalent"}

    write("items.json", items)
    write("buildings.json", buildings)
    write("recipes.json", recipes)
    write("shops.json", shops)
    write("cities.json", cities)
    write("economy.json", economy)
    write("catalog_meta.json", meta)
    write("migration_items.json", item_map)
    # keep routes from legacy
    if (LEGACY / "routes.json").exists():
        write("routes.json", json.loads((LEGACY / "routes.json").read_text()))

    print(json.dumps(meta, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
