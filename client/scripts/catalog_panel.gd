extends AcceptDialog
class_name CatalogPanel

const KINDS := ["items", "buildings", "recipes", "shops"]
var catalog: Dictionary = {}
var section := OptionButton.new()
var search := LineEdit.new()
var rows := ItemList.new()
var details := RichTextLabel.new()

func _ready() -> void:
	title = "Каталог экономики"
	ok_button_text = "Закрыть"
	var content := VBoxContainer.new()
	add_child(content)
	content.custom_minimum_size = Vector2(520, 420)
	section.custom_minimum_size.y = EconomyView.touch_height()
	section.get_popup().add_theme_constant_override("v_separation", EconomyView.touch_height() - 28)
	search.custom_minimum_size.y = EconomyView.touch_height()
	for label in ["Товары", "Здания", "Рецепты", "Магазины"]:
		section.add_item(label)
	content.add_child(section)
	search.placeholder_text = "Поиск по названию или ID"
	content.add_child(search)
	rows.size_flags_vertical = Control.SIZE_EXPAND_FILL
	rows.fixed_icon_size = Vector2i(38, 38)
	rows.add_theme_constant_override("v_separation", EconomyView.touch_height() - 38)
	get_ok_button().custom_minimum_size.y = EconomyView.touch_height()
	content.add_child(rows)
	details.custom_minimum_size.y = 126
	details.add_theme_font_size_override("normal_font_size", 18)
	content.add_child(details)
	section.item_selected.connect(func(_index): refresh())
	search.text_changed.connect(func(_query): refresh())
	rows.item_selected.connect(_show_entry)

func open(data: Dictionary) -> void:
	catalog = data
	title = "Каталог экономики V%s" % catalog["economy_catalog_version"]
	search.clear()
	refresh()
	popup_centered_ratio(0.8)

func refresh() -> void:
	rows.clear()
	details.clear()
	var kind: String = KINDS[section.selected]
	var query := search.text.strip_edges().to_lower()
	for entry in catalog.get(kind, []):
		var label := str(entry["name"])
		if not query.is_empty() and query not in label.to_lower() and query not in str(entry["id"]).to_lower():
			continue
		var texture: Texture2D
		if kind == "recipes":
			texture = CatalogUI.icon(CatalogUI.find(catalog, "items", entry["output"]["item_id"]), "items")
		else:
			texture = CatalogUI.icon(entry, kind)
		var index := rows.add_item(label, texture)
		rows.set_item_metadata(index, entry)
	if rows.item_count > 0:
		rows.select(0)
		_show_entry(0)
	else:
		details.text = "Ничего не найдено"

func _show_entry(index: int) -> void:
	var entry: Dictionary = rows.get_item_metadata(index)
	match KINDS[section.selected]:
		"items":
			var producers := PackedStringArray()
			var consumers := PackedStringArray()
			for recipe in catalog["recipes"]:
				if recipe["output"]["item_id"] == entry["id"]:
					producers.append(recipe["name"])
				var inputs = recipe.get("inputs")
				if inputs is Array and inputs.any(func(input): return input["item_id"] == entry["id"]):
					consumers.append(recipe["name"])
			details.text = "%s · %s" % [entry["name"], {"raw": "сырьё", "intermediate": "полуфабрикат", "finished": "готовый товар"}.get(entry["tier"], entry["tier"])]
			if not producers.is_empty():
				details.text += "\nПолучение · " + ", ".join(producers)
			if not consumers.is_empty():
				details.text += "\nИспользование · " + ", ".join(consumers)
		"buildings":
			details.text = "Стоимость строительства: %s ₽" % int(entry["build_cost"])
		"recipes":
			details.text = "%s\n%s" % [
				CatalogUI.entry_name(catalog, "buildings", entry["building_type"]),
				EconomyView.recipe_text(entry, catalog),
			]
		"shops":
			var items := PackedStringArray()
			for item in catalog["items"]:
				if CatalogUI.shop_allows(entry, item):
					items.append(item["name"])
			details.text = "Ассортимент: %s" % ", ".join(items)
