extends RefCounted
class_name BreadGuide

const CITY := "ufa"
const STEPS := [
	{"type": "grain_farm", "recipe": "extract_wheat", "output": "wheat", "title": "Ферма · пшеница", "hint": "Поставьте ферму на свободный участок с пшеницей."},
	{"type": "flour_mill", "recipe": "food_01", "output": "wheat_flour", "title": "Мельница · пшеничная мука", "hint": "Постройте мельницу в Уфе. Она может автоматически брать пшеницу со склада вашей фермы."},
	{"type": "bakery", "recipe": "food_11", "output": "wheat_bread", "title": "Пекарня · пшеничный хлеб", "hint": "Постройте пекарню в Уфе. Она может автоматически забирать муку с мельницы."},
	{"type": "grocery", "item": "wheat_bread", "title": "Продуктовый · пшеничный хлеб", "hint": "Арендуйте витрину в Уфе для пшеничного хлеба. Магазин автоматически возьмёт хлеб с вашего склада в этом городе."},
]

static func available(catalog: Dictionary) -> bool:
	return int(catalog.get("economy_catalog_version", 0)) >= 2 and not CatalogUI.find(catalog, "recipes", "food_11").is_empty()

static func asset_for(step: Dictionary, world: Dictionary, firm_id: String) -> Dictionary:
	var shops := step.has("item")
	var fallback: Dictionary = {}
	for asset in world.get("shops" if shops else "buildings", {}).values():
		if asset.get("firm_id") != firm_id or asset.get("city_id") != CITY:
			continue
		if asset.get("shop_type" if shops else "type") != step["type"]:
			continue
		if matches(step, asset):
			return asset
		fallback = asset
	return fallback

static func matches(step: Dictionary, asset: Dictionary) -> bool:
	if step.has("item"):
		return asset.get("item_id") == step["item"] and asset.get("stock_source", "own") == "own" and asset.get("auto_restock", true)
	if asset.get("recipe_id") != step["recipe"]:
		return false
	return step["type"] == "grain_farm" or (
		asset.get("auto_procure", false) and asset.get("default_input", {}).get("mode") == "own"
	)

static func suggest_parcel(step: Dictionary, world: Dictionary, firm_id: String) -> Dictionary:
	var city: Dictionary = world.get("cities", {}).get(CITY, {})
	var wheat := {}
	var deposits = city.get("deposits")
	if step["type"] == "grain_farm" and deposits is Array:
		for deposit in deposits:
			if deposit["item_id"] == "wheat" and int(deposit["amount"]) > 0:
				wheat[Vector2i(int(deposit["x"]), int(deposit["y"]))] = true
	var candidate: Dictionary = {}
	for parcel in city.get("parcels", []):
		var owner := str(parcel.get("owner_id", ""))
		if not owner.is_empty() and owner != firm_id:
			continue
		if not str(parcel.get("building_id", "")).is_empty() or not str(parcel.get("rented_by", "")).is_empty():
			continue
		if step.has("item"):
			if int(parcel.get("retail_slots", 0)) <= 0:
				continue
		elif step["type"] == "grain_farm":
			var coord := Vector2i(int(parcel["x"]), int(parcel["y"]))
			if not wheat.has(coord):
				continue
		elif str(parcel.get("zone")) not in ["industrial", "empty"]:
			continue
		if parcel.get("has_road", false):
			return parcel
		if candidate.is_empty():
			candidate = parcel
	return candidate
