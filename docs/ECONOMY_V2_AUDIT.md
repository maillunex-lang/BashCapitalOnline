# Economy V2 machine audit

{
  "counts": {
    "items": 170,
    "recipes": 200,
    "buildings": 92,
    "shops": 14
  },
  "sku": [
    {
      "id": "iron_ore",
      "tier": "raw",
      "name": "Железная руда",
      "sources": [
        "extract_iron_ore"
      ],
      "consumers": [
        "met_01"
      ],
      "shops": [],
      "cities": [
        "beloretsk"
      ]
    },
    {
      "id": "copper_ore",
      "tier": "raw",
      "name": "Медная руда",
      "sources": [
        "extract_copper_ore"
      ],
      "consumers": [
        "met_03"
      ],
      "shops": [],
      "cities": [
        "beloretsk"
      ]
    },
    {
      "id": "bauxite",
      "tier": "raw",
      "name": "Бокситы",
      "sources": [
        "extract_bauxite"
      ],
      "consumers": [
        "met_04"
      ],
      "shops": [],
      "cities": [
        "beloretsk"
      ]
    },
    {
      "id": "limestone",
      "tier": "raw",
      "name": "Известняк",
      "sources": [
        "extract_limestone"
      ],
      "consumers": [
        "met_02",
        "oil_04",
        "bld_01",
        "bld_04",
        "bld_08",
        "alt_09"
      ],
      "shops": [],
      "cities": [
        "beloretsk",
        "sterlitamak"
      ]
    },
    {
      "id": "sand",
      "tier": "raw",
      "name": "Песок",
      "sources": [
        "extract_sand"
      ],
      "consumers": [
        "bld_02",
        "bld_04",
        "bld_08",
        "alt_12",
        "alt_28"
      ],
      "shops": [],
      "cities": [
        "ufa",
        "beloretsk",
        "sterlitamak"
      ]
    },
    {
      "id": "clay",
      "tier": "raw",
      "name": "Глина",
      "sources": [
        "extract_clay"
      ],
      "consumers": [
        "bld_01",
        "bld_03",
        "bld_06",
        "alt_09",
        "alt_10"
      ],
      "shops": [],
      "cities": [
        "ufa",
        "beloretsk",
        "sterlitamak"
      ]
    },
    {
      "id": "quartz",
      "tier": "raw",
      "name": "Кварц",
      "sources": [
        "extract_quartz"
      ],
      "consumers": [
        "bld_04",
        "bld_06",
        "tech_03",
        "alt_28"
      ],
      "shops": [],
      "cities": [
        "beloretsk",
        "sterlitamak"
      ]
    },
    {
      "id": "stone",
      "tier": "raw",
      "name": "Камень",
      "sources": [
        "extract_stone"
      ],
      "consumers": [
        "bld_02"
      ],
      "shops": [],
      "cities": [
        "beloretsk",
        "sterlitamak"
      ]
    },
    {
      "id": "salt",
      "tier": "raw",
      "name": "Соль",
      "sources": [
        "extract_salt"
      ],
      "consumers": [
        "oil_08",
        "food_25",
        "bash_02"
      ],
      "shops": [],
      "cities": [
        "sterlitamak"
      ]
    },
    {
      "id": "crude_oil",
      "tier": "raw",
      "name": "Нефть",
      "sources": [
        "extract_crude_oil"
      ],
      "consumers": [
        "oil_01",
        "oil_02",
        "oil_03",
        "oil_04",
        "oil_05",
        "oil_07",
        "alt_26"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "natural_gas",
      "tier": "raw",
      "name": "Природный газ",
      "sources": [
        "extract_natural_gas"
      ],
      "consumers": [
        "oil_05",
        "oil_06",
        "oil_07",
        "oil_08",
        "fuel_04",
        "alt_26"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "coal",
      "tier": "raw",
      "name": "Уголь",
      "sources": [
        "extract_coal"
      ],
      "consumers": [
        "met_01",
        "met_03",
        "bld_01",
        "bld_03"
      ],
      "shops": [],
      "cities": [
        "beloretsk"
      ]
    },
    {
      "id": "peat",
      "tier": "raw",
      "name": "Торф",
      "sources": [
        "extract_peat"
      ],
      "consumers": [
        "alt_09",
        "alt_10"
      ],
      "shops": [],
      "cities": [
        "sterlitamak"
      ]
    },
    {
      "id": "wheat",
      "tier": "raw",
      "name": "Пшеница",
      "sources": [
        "extract_wheat"
      ],
      "consumers": [
        "food_01"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "rye",
      "tier": "raw",
      "name": "Рожь",
      "sources": [
        "extract_rye"
      ],
      "consumers": [
        "food_02"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "barley",
      "tier": "raw",
      "name": "Ячмень",
      "sources": [
        "extract_barley"
      ],
      "consumers": [
        "alt_01",
        "alt_25"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "oats",
      "tier": "raw",
      "name": "Овёс",
      "sources": [
        "extract_oats"
      ],
      "consumers": [
        "alt_02"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "buckwheat",
      "tier": "raw",
      "name": "Гречиха",
      "sources": [
        "extract_buckwheat"
      ],
      "consumers": [
        "food_03"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "potato",
      "tier": "raw",
      "name": "Картофель",
      "sources": [
        "extract_potato"
      ],
      "consumers": [
        "food_18"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "sugar_beet",
      "tier": "raw",
      "name": "Сахарная свёкла",
      "sources": [
        "extract_sugar_beet"
      ],
      "consumers": [
        "food_04"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "sunflower",
      "tier": "raw",
      "name": "Подсолнечник",
      "sources": [
        "extract_sunflower"
      ],
      "consumers": [
        "food_05"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "flax",
      "tier": "raw",
      "name": "Лён",
      "sources": [
        "extract_flax"
      ],
      "consumers": [
        "tex_01"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "vegetables",
      "tier": "raw",
      "name": "Овощи",
      "sources": [
        "extract_vegetables"
      ],
      "consumers": [
        "food_19",
        "alt_05"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "milk",
      "tier": "raw",
      "name": "Сырое молоко",
      "sources": [
        "extract_milk"
      ],
      "consumers": [
        "food_06",
        "food_10",
        "food_22",
        "food_25",
        "alt_21",
        "alt_22"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "beef",
      "tier": "raw",
      "name": "Говядина",
      "sources": [
        "extract_beef"
      ],
      "consumers": [
        "food_08",
        "food_27"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "pork",
      "tier": "raw",
      "name": "Свинина",
      "sources": [
        "extract_pork"
      ],
      "consumers": [
        "food_08",
        "food_28"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "poultry",
      "tier": "raw",
      "name": "Мясо птицы",
      "sources": [
        "extract_poultry"
      ],
      "consumers": [
        "food_29"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "eggs",
      "tier": "raw",
      "name": "Яйца",
      "sources": [
        "extract_eggs"
      ],
      "consumers": [
        "food_15",
        "food_30"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "wool",
      "tier": "raw",
      "name": "Шерсть",
      "sources": [
        "extract_wool"
      ],
      "consumers": [
        "tex_02",
        "tex_06",
        "app_09"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "raw_hide",
      "tier": "raw",
      "name": "Кожа-сырец",
      "sources": [
        "extract_raw_hide"
      ],
      "consumers": [
        "tex_05"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "softwood",
      "tier": "raw",
      "name": "Хвойная древесина",
      "sources": [
        "extract_softwood"
      ],
      "consumers": [
        "wood_01"
      ],
      "shops": [],
      "cities": [
        "ufa",
        "beloretsk"
      ]
    },
    {
      "id": "hardwood",
      "tier": "raw",
      "name": "Лиственная древесина",
      "sources": [
        "extract_hardwood"
      ],
      "consumers": [
        "alt_07",
        "alt_08",
        "alt_16"
      ],
      "shops": [],
      "cities": [
        "ufa",
        "beloretsk"
      ]
    },
    {
      "id": "wood_chips",
      "tier": "raw",
      "name": "Древесная щепа",
      "sources": [
        "extract_wood_chips"
      ],
      "consumers": [
        "tex_06",
        "wood_03",
        "wood_04"
      ],
      "shops": [],
      "cities": [
        "ufa",
        "beloretsk"
      ]
    },
    {
      "id": "forest_berries",
      "tier": "raw",
      "name": "Лесные ягоды",
      "sources": [
        "extract_forest_berries"
      ],
      "consumers": [
        "drink_04",
        "alt_24"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "mushrooms",
      "tier": "raw",
      "name": "Грибы",
      "sources": [
        "extract_mushrooms"
      ],
      "consumers": [
        "alt_05"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "honey",
      "tier": "raw",
      "name": "Мёд",
      "sources": [
        "extract_honey"
      ],
      "consumers": [
        "drink_07",
        "bash_01",
        "bash_03",
        "alt_04",
        "alt_21",
        "alt_23"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "apples",
      "tier": "raw",
      "name": "Яблоки",
      "sources": [
        "extract_apples"
      ],
      "consumers": [
        "food_09",
        "drink_02",
        "alt_24"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "cherries",
      "tier": "raw",
      "name": "Вишня",
      "sources": [
        "extract_cherries"
      ],
      "consumers": [
        "food_09",
        "drink_03"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "horse_meat",
      "tier": "raw",
      "name": "Конина",
      "sources": [
        "extract_horse_meat"
      ],
      "consumers": [
        "food_17"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "mare_milk",
      "tier": "raw",
      "name": "Кобылье молоко",
      "sources": [
        "extract_mare_milk"
      ],
      "consumers": [
        "drink_06"
      ],
      "shops": [],
      "cities": [
        "ufa"
      ]
    },
    {
      "id": "pig_iron",
      "tier": "intermediate",
      "name": "Чугун",
      "sources": [
        "met_01"
      ],
      "consumers": [
        "met_02",
        "alt_27"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "steel",
      "tier": "intermediate",
      "name": "Сталь",
      "sources": [
        "met_02",
        "alt_27"
      ],
      "consumers": [
        "met_05",
        "met_06",
        "met_09",
        "tech_01",
        "auto_01",
        "auto_04"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "copper",
      "tier": "intermediate",
      "name": "Медь",
      "sources": [
        "met_03"
      ],
      "consumers": [
        "met_07",
        "met_09",
        "tech_03",
        "tech_05"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "aluminum",
      "tier": "intermediate",
      "name": "Алюминий",
      "sources": [
        "met_04"
      ],
      "consumers": [
        "met_08",
        "tech_01",
        "tech_05"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "steel_sheet",
      "tier": "intermediate",
      "name": "Стальной лист",
      "sources": [
        "met_05"
      ],
      "consumers": [
        "auto_06",
        "elec_03",
        "elec_04",
        "elec_05",
        "alt_15"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "steel_pipe",
      "tier": "intermediate",
      "name": "Стальная труба",
      "sources": [
        "met_06"
      ],
      "consumers": [
        "build_08"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "copper_wire",
      "tier": "intermediate",
      "name": "Медный провод",
      "sources": [
        "met_07"
      ],
      "consumers": [
        "tech_02",
        "tech_04"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "aluminum_profile",
      "tier": "intermediate",
      "name": "Алюминиевый профиль",
      "sources": [
        "met_08"
      ],
      "consumers": [
        "build_04"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "metal_parts",
      "tier": "intermediate",
      "name": "Металлические детали",
      "sources": [
        "met_09"
      ],
      "consumers": [
        "tech_01",
        "tech_07",
        "auto_04",
        "auto_05",
        "alt_27"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "gasoline_base",
      "tier": "intermediate",
      "name": "Бензиновая фракция",
      "sources": [
        "oil_01"
      ],
      "consumers": [
        "fuel_01",
        "fuel_02"
      ],
      "shops": [
        "fuel_depot"
      ],
      "cities": []
    },
    {
      "id": "diesel_base",
      "tier": "intermediate",
      "name": "Дизельная фракция",
      "sources": [
        "oil_02"
      ],
      "consumers": [
        "fuel_03",
        "alt_30"
      ],
      "shops": [
        "fuel_depot"
      ],
      "cities": []
    },
    {
      "id": "fuel_oil",
      "tier": "intermediate",
      "name": "Мазут",
      "sources": [
        "oil_03"
      ],
      "consumers": [
        "auto_03",
        "fuel_05"
      ],
      "shops": [
        "fuel_depot"
      ],
      "cities": []
    },
    {
      "id": "bitumen",
      "tier": "intermediate",
      "name": "Битум",
      "sources": [
        "oil_04"
      ],
      "consumers": [
        "alt_13"
      ],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "polymer_granules",
      "tier": "intermediate",
      "name": "Полимерные гранулы",
      "sources": [
        "oil_05",
        "alt_26"
      ],
      "consumers": [
        "oil_06",
        "bld_05",
        "tech_02",
        "tech_03",
        "tech_06"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "synthetic_rubber",
      "tier": "intermediate",
      "name": "Синтетический каучук",
      "sources": [
        "oil_06"
      ],
      "consumers": [
        "app_07",
        "app_08",
        "auto_01",
        "alt_20"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "industrial_solvent",
      "tier": "intermediate",
      "name": "Растворитель",
      "sources": [
        "oil_07"
      ],
      "consumers": [
        "bld_07",
        "home_06",
        "alt_13"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "chemical_reagent",
      "tier": "intermediate",
      "name": "Химический реагент",
      "sources": [
        "oil_08"
      ],
      "consumers": [
        "bld_07",
        "tex_05",
        "tex_07",
        "wood_02",
        "wood_03",
        "tech_05",
        "home_06",
        "auto_03",
        "fuel_02",
        "alt_08",
        "alt_30"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "cement",
      "tier": "intermediate",
      "name": "Цемент",
      "sources": [
        "bld_01",
        "alt_09"
      ],
      "consumers": [
        "bld_02",
        "bld_08",
        "build_01",
        "alt_11"
      ],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "concrete",
      "tier": "intermediate",
      "name": "Бетон",
      "sources": [
        "bld_02"
      ],
      "consumers": [
        "build_03"
      ],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "brick",
      "tier": "intermediate",
      "name": "Кирпич",
      "sources": [
        "bld_03",
        "alt_10"
      ],
      "consumers": [
        "build_02"
      ],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "glass_sheet",
      "tier": "intermediate",
      "name": "Листовое стекло",
      "sources": [
        "bld_04",
        "alt_28"
      ],
      "consumers": [
        "bld_05",
        "home_07",
        "build_04",
        "elec_01",
        "elec_02"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "insulation",
      "tier": "intermediate",
      "name": "Утеплитель",
      "sources": [
        "bld_05"
      ],
      "consumers": [
        "build_07"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "ceramic_tile",
      "tier": "intermediate",
      "name": "Керамическая плитка",
      "sources": [
        "bld_06"
      ],
      "consumers": [
        "home_07",
        "build_06"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "paint_base",
      "tier": "intermediate",
      "name": "Основа краски",
      "sources": [
        "bld_07",
        "alt_13"
      ],
      "consumers": [
        "build_05"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "dry_mix",
      "tier": "intermediate",
      "name": "Сухая строительная смесь",
      "sources": [
        "bld_08"
      ],
      "consumers": [
        "alt_11",
        "alt_12"
      ],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "wheat_flour",
      "tier": "intermediate",
      "name": "Пшеничная мука",
      "sources": [
        "food_01"
      ],
      "consumers": [
        "food_11",
        "food_13",
        "food_14",
        "food_15",
        "bash_03",
        "alt_03",
        "alt_23",
        "alt_minced_pasta"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "rye_flour",
      "tier": "intermediate",
      "name": "Ржаная мука",
      "sources": [
        "food_02"
      ],
      "consumers": [
        "food_12",
        "drink_05",
        "alt_03",
        "alt_04"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "groats",
      "tier": "intermediate",
      "name": "Крупа",
      "sources": [
        "food_03",
        "alt_01",
        "alt_02"
      ],
      "consumers": [
        "food_16"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "sugar",
      "tier": "intermediate",
      "name": "Сахар",
      "sources": [
        "food_04"
      ],
      "consumers": [
        "food_13",
        "food_14",
        "food_21",
        "drink_05",
        "bash_03",
        "alt_23",
        "alt_25"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "sunflower_oil",
      "tier": "intermediate",
      "name": "Растительное масло",
      "sources": [
        "food_05"
      ],
      "consumers": [
        "food_19",
        "food_20"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "cream",
      "tier": "intermediate",
      "name": "Сливки",
      "sources": [
        "food_06"
      ],
      "consumers": [
        "food_07",
        "food_23",
        "alt_22"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "butter_base",
      "tier": "intermediate",
      "name": "Масляная основа",
      "sources": [
        "food_07"
      ],
      "consumers": [
        "food_14",
        "food_24"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "minced_meat",
      "tier": "intermediate",
      "name": "Фарш",
      "sources": [
        "food_08"
      ],
      "consumers": [
        "alt_minced_pasta"
      ],
      "shops": [
        "butcher"
      ],
      "cities": []
    },
    {
      "id": "fruit_puree",
      "tier": "intermediate",
      "name": "Фруктовое пюре",
      "sources": [
        "food_09"
      ],
      "consumers": [
        "food_26"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "fermented_milk",
      "tier": "intermediate",
      "name": "Кисломолочная основа",
      "sources": [
        "food_10"
      ],
      "consumers": [
        "food_26",
        "drink_07",
        "bash_02"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "flax_fiber",
      "tier": "intermediate",
      "name": "Льняное волокно",
      "sources": [
        "tex_01"
      ],
      "consumers": [
        "tex_02"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "yarn",
      "tier": "intermediate",
      "name": "Пряжа",
      "sources": [
        "tex_02"
      ],
      "consumers": [
        "tex_03",
        "tex_04"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "woven_fabric",
      "tier": "intermediate",
      "name": "Ткань",
      "sources": [
        "tex_03"
      ],
      "consumers": [
        "tex_07",
        "app_02",
        "app_08",
        "home_08",
        "alt_18",
        "alt_20"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "knitted_fabric",
      "tier": "intermediate",
      "name": "Трикотажное полотно",
      "sources": [
        "tex_04"
      ],
      "consumers": [
        "app_01",
        "app_05",
        "app_09",
        "alt_19"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "tanned_leather",
      "tier": "intermediate",
      "name": "Выделанная кожа",
      "sources": [
        "tex_05"
      ],
      "consumers": [
        "app_06",
        "app_07",
        "alt_20"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "textile_filling",
      "tier": "intermediate",
      "name": "Текстильный наполнитель",
      "sources": [
        "tex_06"
      ],
      "consumers": [
        "app_04",
        "home_05",
        "home_08"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "dyed_fabric",
      "tier": "intermediate",
      "name": "Окрашенная ткань",
      "sources": [
        "tex_07"
      ],
      "consumers": [
        "app_03",
        "app_04",
        "app_06",
        "home_05",
        "bash_05",
        "alt_17",
        "alt_19"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "lumber",
      "tier": "intermediate",
      "name": "Доски",
      "sources": [
        "wood_01",
        "alt_07"
      ],
      "consumers": [
        "wood_02",
        "wood_06",
        "home_01",
        "bash_05"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "plywood",
      "tier": "intermediate",
      "name": "Фанера",
      "sources": [
        "wood_02",
        "alt_08"
      ],
      "consumers": [
        "home_03",
        "home_04",
        "alt_17"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "particle_board",
      "tier": "intermediate",
      "name": "ДСП",
      "sources": [
        "wood_03"
      ],
      "consumers": [
        "wood_06"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "paper",
      "tier": "intermediate",
      "name": "Бумага",
      "sources": [
        "wood_04"
      ],
      "consumers": [
        "wood_05",
        "auto_05",
        "bash_05"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "cardboard",
      "tier": "intermediate",
      "name": "Картон",
      "sources": [
        "wood_05"
      ],
      "consumers": [
        "bash_04"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "furniture_parts",
      "tier": "intermediate",
      "name": "Мебельные детали",
      "sources": [
        "wood_06"
      ],
      "consumers": [
        "home_01",
        "home_02",
        "home_03",
        "home_04",
        "home_05",
        "alt_16"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "engine",
      "tier": "intermediate",
      "name": "Двигатель",
      "sources": [
        "tech_01"
      ],
      "consumers": [
        "alt_14",
        "alt_15"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "wiring_harness",
      "tier": "intermediate",
      "name": "Жгут проводки",
      "sources": [
        "tech_02"
      ],
      "consumers": [
        "tech_07",
        "auto_07"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "circuit_board",
      "tier": "intermediate",
      "name": "Печатная плата",
      "sources": [
        "tech_03"
      ],
      "consumers": [
        "tech_04",
        "elec_06",
        "alt_29"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "electronic_module",
      "tier": "intermediate",
      "name": "Электронный модуль",
      "sources": [
        "tech_04"
      ],
      "consumers": [
        "elec_01",
        "elec_02",
        "elec_03",
        "elec_04",
        "elec_06",
        "alt_15",
        "alt_29"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "battery_cell",
      "tier": "intermediate",
      "name": "Аккумуляторная ячейка",
      "sources": [
        "tech_05"
      ],
      "consumers": [
        "auto_02",
        "elec_01",
        "alt_29"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "plastic_parts",
      "tier": "intermediate",
      "name": "Пластиковые детали",
      "sources": [
        "tech_06"
      ],
      "consumers": [
        "auto_02",
        "auto_05",
        "auto_06",
        "elec_01",
        "elec_02",
        "elec_05",
        "elec_06"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "mechanical_assembly",
      "tier": "intermediate",
      "name": "Механический узел",
      "sources": [
        "tech_07"
      ],
      "consumers": [
        "auto_07",
        "elec_03",
        "elec_04",
        "elec_05",
        "alt_14"
      ],
      "shops": [],
      "cities": []
    },
    {
      "id": "wheat_bread",
      "tier": "finished",
      "name": "Хлеб пшеничный",
      "sources": [
        "food_11",
        "alt_03"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "bakery_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "rye_bread",
      "tier": "finished",
      "name": "Хлеб ржаной",
      "sources": [
        "food_12",
        "alt_04"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "bakery_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "bun",
      "tier": "finished",
      "name": "Булочки",
      "sources": [
        "food_13"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "bakery_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "cookies",
      "tier": "finished",
      "name": "Печенье",
      "sources": [
        "food_14",
        "alt_23"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "bakery_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "pasta",
      "tier": "finished",
      "name": "Макароны",
      "sources": [
        "food_15",
        "alt_minced_pasta"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "buckwheat_pack",
      "tier": "finished",
      "name": "Гречка фасованная",
      "sources": [
        "food_16"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "horse_meat_pack",
      "tier": "finished",
      "name": "Конина фасованная",
      "sources": [
        "food_17"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "butcher",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "potato_pack",
      "tier": "finished",
      "name": "Картофель фасованный",
      "sources": [
        "food_18"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "vegetable_mix",
      "tier": "finished",
      "name": "Овощная смесь",
      "sources": [
        "food_19",
        "alt_05"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "sunflower_oil_bottle",
      "tier": "finished",
      "name": "Масло бутылочное",
      "sources": [
        "food_20"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "sugar_pack",
      "tier": "finished",
      "name": "Сахар фасованный",
      "sources": [
        "food_21"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "milk_bottle",
      "tier": "finished",
      "name": "Молоко питьевое",
      "sources": [
        "food_22"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "dairy_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "cream_pack",
      "tier": "finished",
      "name": "Сливки фасованные",
      "sources": [
        "food_23"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "dairy_shop"
      ],
      "cities": []
    },
    {
      "id": "butter",
      "tier": "finished",
      "name": "Масло сливочное",
      "sources": [
        "food_24"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "dairy_shop"
      ],
      "cities": []
    },
    {
      "id": "cheese",
      "tier": "finished",
      "name": "Сыр",
      "sources": [
        "food_25",
        "alt_22"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "dairy_shop"
      ],
      "cities": []
    },
    {
      "id": "yogurt",
      "tier": "finished",
      "name": "Йогурт",
      "sources": [
        "food_26",
        "alt_21"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "dairy_shop"
      ],
      "cities": []
    },
    {
      "id": "beef_pack",
      "tier": "finished",
      "name": "Говядина фасованная",
      "sources": [
        "food_27"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "butcher"
      ],
      "cities": []
    },
    {
      "id": "pork_pack",
      "tier": "finished",
      "name": "Свинина фасованная",
      "sources": [
        "food_28"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "butcher"
      ],
      "cities": []
    },
    {
      "id": "chicken_pack",
      "tier": "finished",
      "name": "Курица фасованная",
      "sources": [
        "food_29"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "butcher"
      ],
      "cities": []
    },
    {
      "id": "eggs_pack",
      "tier": "finished",
      "name": "Яйца фасованные",
      "sources": [
        "food_30"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "butcher"
      ],
      "cities": []
    },
    {
      "id": "bottled_water",
      "tier": "finished",
      "name": "Вода питьевая",
      "sources": [
        "drink_01"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "beverage_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "apple_juice",
      "tier": "finished",
      "name": "Яблочный сок",
      "sources": [
        "drink_02"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "beverage_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "cherry_juice",
      "tier": "finished",
      "name": "Вишнёвый сок",
      "sources": [
        "drink_03"
      ],
      "consumers": [],
      "shops": [
        "beverage_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "berry_drink",
      "tier": "finished",
      "name": "Ягодный морс",
      "sources": [
        "drink_04",
        "alt_24"
      ],
      "consumers": [],
      "shops": [
        "beverage_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "kvass",
      "tier": "finished",
      "name": "Квас",
      "sources": [
        "drink_05",
        "alt_25"
      ],
      "consumers": [],
      "shops": [
        "grocery",
        "beverage_shop"
      ],
      "cities": []
    },
    {
      "id": "koumiss",
      "tier": "finished",
      "name": "Кумыс",
      "sources": [
        "drink_06"
      ],
      "consumers": [],
      "shops": [
        "dairy_shop",
        "beverage_shop"
      ],
      "cities": []
    },
    {
      "id": "dairy_drink",
      "tier": "finished",
      "name": "Кисломолочный напиток",
      "sources": [
        "drink_07"
      ],
      "consumers": [],
      "shops": [
        "dairy_shop",
        "beverage_shop"
      ],
      "cities": []
    },
    {
      "id": "tshirt",
      "tier": "finished",
      "name": "Футболка",
      "sources": [
        "app_01",
        "alt_18"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "shirt",
      "tier": "finished",
      "name": "Рубашка",
      "sources": [
        "app_02"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "trousers",
      "tier": "finished",
      "name": "Брюки",
      "sources": [
        "app_03"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "jacket",
      "tier": "finished",
      "name": "Куртка",
      "sources": [
        "app_04",
        "alt_19"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "knit_sweater",
      "tier": "finished",
      "name": "Свитер",
      "sources": [
        "app_05"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop"
      ],
      "cities": []
    },
    {
      "id": "workwear",
      "tier": "finished",
      "name": "Спецодежда",
      "sources": [
        "app_06"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop"
      ],
      "cities": []
    },
    {
      "id": "leather_shoes",
      "tier": "finished",
      "name": "Кожаная обувь",
      "sources": [
        "app_07",
        "alt_20"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop"
      ],
      "cities": []
    },
    {
      "id": "textile_shoes",
      "tier": "finished",
      "name": "Текстильная обувь",
      "sources": [
        "app_08"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop"
      ],
      "cities": []
    },
    {
      "id": "winter_hat",
      "tier": "finished",
      "name": "Зимняя шапка",
      "sources": [
        "app_09"
      ],
      "consumers": [],
      "shops": [
        "clothing_shop"
      ],
      "cities": []
    },
    {
      "id": "table",
      "tier": "finished",
      "name": "Стол",
      "sources": [
        "home_01",
        "alt_16"
      ],
      "consumers": [],
      "shops": [
        "furniture_home",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "chair",
      "tier": "finished",
      "name": "Стул",
      "sources": [
        "home_02",
        "alt_17"
      ],
      "consumers": [],
      "shops": [
        "furniture_home",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "wardrobe",
      "tier": "finished",
      "name": "Шкаф",
      "sources": [
        "home_03"
      ],
      "consumers": [],
      "shops": [
        "furniture_home",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "bed",
      "tier": "finished",
      "name": "Кровать",
      "sources": [
        "home_04"
      ],
      "consumers": [],
      "shops": [
        "furniture_home",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "sofa",
      "tier": "finished",
      "name": "Диван",
      "sources": [
        "home_05"
      ],
      "consumers": [],
      "shops": [
        "furniture_home"
      ],
      "cities": []
    },
    {
      "id": "detergent",
      "tier": "finished",
      "name": "Моющее средство",
      "sources": [
        "home_06"
      ],
      "consumers": [],
      "shops": [
        "furniture_home"
      ],
      "cities": []
    },
    {
      "id": "dishware",
      "tier": "finished",
      "name": "Посуда",
      "sources": [
        "home_07"
      ],
      "consumers": [],
      "shops": [
        "furniture_home"
      ],
      "cities": []
    },
    {
      "id": "pillow",
      "tier": "finished",
      "name": "Подушка",
      "sources": [
        "home_08"
      ],
      "consumers": [],
      "shops": [
        "furniture_home"
      ],
      "cities": []
    },
    {
      "id": "cement_bag",
      "tier": "finished",
      "name": "Цемент мешковой",
      "sources": [
        "build_01",
        "alt_11"
      ],
      "consumers": [],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "brick_pack",
      "tier": "finished",
      "name": "Кирпич упаковка",
      "sources": [
        "build_02"
      ],
      "consumers": [],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "concrete_mix",
      "tier": "finished",
      "name": "Бетон товарный",
      "sources": [
        "build_03",
        "alt_12"
      ],
      "consumers": [],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "window_unit",
      "tier": "finished",
      "name": "Оконный блок",
      "sources": [
        "build_04"
      ],
      "consumers": [],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "paint_can",
      "tier": "finished",
      "name": "Краска",
      "sources": [
        "build_05"
      ],
      "consumers": [],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "tile_pack",
      "tier": "finished",
      "name": "Плитка упаковка",
      "sources": [
        "build_06"
      ],
      "consumers": [],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "insulation_pack",
      "tier": "finished",
      "name": "Утеплитель упаковка",
      "sources": [
        "build_07"
      ],
      "consumers": [],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "pipe_set",
      "tier": "finished",
      "name": "Комплект труб",
      "sources": [
        "build_08"
      ],
      "consumers": [],
      "shops": [
        "building_market"
      ],
      "cities": []
    },
    {
      "id": "tire",
      "tier": "finished",
      "name": "Шина",
      "sources": [
        "auto_01"
      ],
      "consumers": [],
      "shops": [
        "auto_parts_shop"
      ],
      "cities": []
    },
    {
      "id": "car_battery",
      "tier": "finished",
      "name": "Автомобильный аккумулятор",
      "sources": [
        "auto_02"
      ],
      "consumers": [],
      "shops": [
        "auto_parts_shop"
      ],
      "cities": []
    },
    {
      "id": "engine_oil",
      "tier": "finished",
      "name": "Моторное масло",
      "sources": [
        "auto_03"
      ],
      "consumers": [],
      "shops": [
        "auto_parts_shop"
      ],
      "cities": []
    },
    {
      "id": "brake_parts",
      "tier": "finished",
      "name": "Тормозные детали",
      "sources": [
        "auto_04"
      ],
      "consumers": [],
      "shops": [
        "auto_parts_shop"
      ],
      "cities": []
    },
    {
      "id": "filter_set",
      "tier": "finished",
      "name": "Фильтры",
      "sources": [
        "auto_05"
      ],
      "consumers": [],
      "shops": [
        "auto_parts_shop"
      ],
      "cities": []
    },
    {
      "id": "car_body_parts",
      "tier": "finished",
      "name": "Кузовные детали",
      "sources": [
        "auto_06"
      ],
      "consumers": [],
      "shops": [
        "auto_parts_shop"
      ],
      "cities": []
    },
    {
      "id": "car_spare_kit",
      "tier": "finished",
      "name": "Комплект запчастей",
      "sources": [
        "auto_07",
        "alt_14"
      ],
      "consumers": [],
      "shops": [
        "auto_parts_shop"
      ],
      "cities": []
    },
    {
      "id": "smartphone",
      "tier": "finished",
      "name": "Смартфон",
      "sources": [
        "elec_01"
      ],
      "consumers": [],
      "shops": [
        "electronics_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "television",
      "tier": "finished",
      "name": "Телевизор",
      "sources": [
        "elec_02"
      ],
      "consumers": [],
      "shops": [
        "electronics_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "refrigerator",
      "tier": "finished",
      "name": "Холодильник",
      "sources": [
        "elec_03"
      ],
      "consumers": [],
      "shops": [
        "electronics_shop",
        "hypermarket"
      ],
      "cities": []
    },
    {
      "id": "washing_machine",
      "tier": "finished",
      "name": "Стиральная машина",
      "sources": [
        "elec_04",
        "alt_15"
      ],
      "consumers": [],
      "shops": [
        "electronics_shop"
      ],
      "cities": []
    },
    {
      "id": "electric_kettle",
      "tier": "finished",
      "name": "Электрочайник",
      "sources": [
        "elec_05"
      ],
      "consumers": [],
      "shops": [
        "electronics_shop"
      ],
      "cities": []
    },
    {
      "id": "home_router",
      "tier": "finished",
      "name": "Домашний роутер",
      "sources": [
        "elec_06",
        "alt_29"
      ],
      "consumers": [],
      "shops": [
        "electronics_shop"
      ],
      "cities": []
    },
    {
      "id": "bashkir_honey_pack",
      "tier": "finished",
      "name": "Башкирский мёд фасованный",
      "sources": [
        "bash_01"
      ],
      "consumers": [
        "bash_04"
      ],
      "shops": [
        "bashkir_shop"
      ],
      "cities": []
    },
    {
      "id": "qurut",
      "tier": "finished",
      "name": "Ҡорот",
      "sources": [
        "bash_02"
      ],
      "consumers": [],
      "shops": [
        "dairy_shop",
        "bashkir_shop"
      ],
      "cities": []
    },
    {
      "id": "chak_chak",
      "tier": "finished",
      "name": "Чак-чак",
      "sources": [
        "bash_03"
      ],
      "consumers": [],
      "shops": [
        "bakery_shop",
        "bashkir_shop"
      ],
      "cities": []
    },
    {
      "id": "honey_gift_set",
      "tier": "finished",
      "name": "Подарочный набор мёда",
      "sources": [
        "bash_04"
      ],
      "consumers": [],
      "shops": [
        "bashkir_shop"
      ],
      "cities": []
    },
    {
      "id": "bashkir_souvenir",
      "tier": "finished",
      "name": "Сувенир Башкортостана",
      "sources": [
        "bash_05"
      ],
      "consumers": [],
      "shops": [
        "bashkir_shop"
      ],
      "cities": []
    },
    {
      "id": "ai92",
      "tier": "finished",
      "name": "Бензин АИ-92",
      "sources": [
        "fuel_01"
      ],
      "consumers": [],
      "shops": [
        "fuel_station",
        "fuel_depot"
      ],
      "cities": []
    },
    {
      "id": "ai95",
      "tier": "finished",
      "name": "Бензин АИ-95",
      "sources": [
        "fuel_02"
      ],
      "consumers": [],
      "shops": [
        "fuel_station",
        "fuel_depot"
      ],
      "cities": []
    },
    {
      "id": "diesel_fuel",
      "tier": "finished",
      "name": "Дизельное топливо",
      "sources": [
        "fuel_03",
        "alt_30"
      ],
      "consumers": [],
      "shops": [
        "fuel_station",
        "fuel_depot"
      ],
      "cities": []
    },
    {
      "id": "lpg",
      "tier": "finished",
      "name": "Автогаз",
      "sources": [
        "fuel_04"
      ],
      "consumers": [],
      "shops": [
        "fuel_station",
        "fuel_depot"
      ],
      "cities": []
    },
    {
      "id": "heating_fuel",
      "tier": "finished",
      "name": "Печное топливо",
      "sources": [
        "fuel_05"
      ],
      "consumers": [],
      "shops": [
        "fuel_depot"
      ],
      "cities": []
    }
  ]
}
