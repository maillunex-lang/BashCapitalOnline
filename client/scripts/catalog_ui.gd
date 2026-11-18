extends RefCounted
class_name CatalogUI

const FALLBACKS := {"items": "icon_market", "buildings": "icon_factory", "shops": "icon_shop"}
const ICON_PATHS := preload("res://assets/catalog_v2/icons.gd").PATHS
static var _textures: Dictionary = {}

static func find(catalog: Dictionary, kind: String, id: String) -> Dictionary:
	for entry in catalog.get(kind, []):
		if entry["id"] == id:
			return entry
	return {}

static func entry_name(catalog: Dictionary, kind: String, id: String) -> String:
	return str(find(catalog, kind, id).get("name", id))

static func icon(entry: Dictionary, kind: String) -> Texture2D:
	var fallback := "ui/%s.png" % FALLBACKS[kind]
	var path := "res://assets/" + str(ICON_PATHS.get(entry.get("icon", ""), fallback))
	if not _textures.has(path):
		var texture: Texture2D = load(path) if ResourceLoader.exists(path) else load("res://assets/" + fallback)
		_textures[path] = texture
	return _textures[path]

static func add_option(button: OptionButton, entry: Dictionary, kind: String, selected := "") -> void:
	var label := str(entry["name"])
	var index := button.item_count
	if index == 0:
		button.fit_to_longest_item = false
		button.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS
		button.add_theme_constant_override("icon_max_width", 28)
		button.get_popup().add_theme_constant_override("icon_max_width", 28)
	button.add_icon_item(icon(entry, kind), label)
	button.set_item_metadata(index, entry["id"])
	if entry["id"] == selected:
		button.select(index)

static func selected_id(button: OptionButton) -> String:
	if button == null or button.selected < 0:
		return ""
	return str(button.get_item_metadata(button.selected))

static func fill_shop_items(button: OptionButton, catalog: Dictionary, shop_id: String, selected := "") -> void:
	button.clear()
	var shop := find(catalog, "shops", shop_id)
	if shop.is_empty():
		return
	for item in catalog["items"]:
		if shop_allows(shop, item):
			add_option(button, item, "items", selected)

static func shop_allows(shop: Dictionary, item: Dictionary) -> bool:
	var allowed = shop.get("allowed_items")
	if allowed is Array and not allowed.is_empty():
		return item["id"] in allowed
	var categories = shop.get("allowed_categories")
	var tiers = shop.get("allowed_tiers")
	return categories is Array and item["category"] in categories and (
		tiers == null or tiers.is_empty() or item["tier"] in tiers
	)
