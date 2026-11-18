extends RefCounted
class_name EconomyView

const PRODUCTION_LABELS := {
	"produced": "Партия выпущена",
	"retooling": "Переналадка прошла на последнем тике",
	"inputs": "Не хватило сырья",
	"labor": "Не хватило работников",
	"cash": "Не хватило средств на выпуск",
	"deposit": "Нет доступной залежи",
	"research": "Требуется исследование",
	"no_recipe": "Нет производственного рецепта",
	"inactive": "Производство отключено",
	"unavailable": "Производство недоступно",
}

static func recipe_text(recipe: Dictionary, catalog: Dictionary) -> String:
	if recipe.is_empty():
		return ""
	var ingredients := PackedStringArray()
	var inputs = recipe.get("inputs")
	if inputs is Array:
		for input in inputs:
			ingredients.append(quantity_text(input, catalog))
	var source := ", ".join(ingredients)
	if source.is_empty():
		source = "Залежь" if not str(recipe.get("extract_mode", "")).is_empty() else "Без сырья"
	var lines := PackedStringArray(["%s → %s" % [source, quantity_text(recipe["output"], catalog)]])
	var labor: Dictionary = recipe.get("labor", {})
	var costs := PackedStringArray()
	for role in [["worker", "раб."], ["engineer", "инж."], ["manager", "упр."]]:
		var count := int(labor.get(role[0], 0))
		if count:
			costs.append("%d %s" % [count, role[1]])
	var energy := int(recipe.get("energy", 0))
	if energy:
		costs.append("энергия %d" % energy)
	var water := float(recipe.get("water_cost", 0))
	if water:
		costs.append("вода %.2f ед." % water)
	var utilities := float(recipe.get("utility_cost", 0))
	if utilities:
		costs.append("коммунальные %.2f ₽ по базовому тарифу" % utilities)
	if not costs.is_empty():
		lines.append("На партию · " + ", ".join(costs))
	return "\n".join(lines)

static func production_text(building: Dictionary, catalog: Dictionary) -> String:
	var report: Dictionary = building.get("production", {})
	var remaining := int(building.get("retool_weeks", 0))
	var line := str(PRODUCTION_LABELS.get(report.get("status", ""), "Ожидается первый производственный тик"))
	if str(building.get("recipe_id", "")).is_empty():
		line = PRODUCTION_LABELS["no_recipe"]
	elif remaining > 0:
		line = "До завершения переналадки: %d тик." % remaining
	elif not report.is_empty():
		line = "Тик %d · %s" % [int(report["tick"]), line]
	if int(report.get("qty", 0)) > 0:
		line += "\nПоследний выпуск · %s · тик %d" % [
			quantity_text(report, catalog), int(report["output_tick"]),
		]
	return line

static func quantity_text(entry: Dictionary, catalog: Dictionary) -> String:
	return "%s ×%d" % [CatalogUI.entry_name(catalog, "items", entry["item_id"]), int(entry["qty"])]

static func add_item_row(list: ItemList, catalog: Dictionary, id: String, text: String) -> void:
	var entry := CatalogUI.find(catalog, "items", id)
	var index := list.add_item(text, CatalogUI.icon(entry, "items"), false)
	list.set_item_tooltip(index, text)

static func list_view() -> ItemList:
	var list := ItemList.new()
	list.fixed_icon_size = Vector2i(34, 34)
	list.auto_height = true
	list.add_theme_constant_override("v_separation", 8)
	return list


static func touch_height() -> int:
	return 88 if OS.has_feature("mobile") else 52
