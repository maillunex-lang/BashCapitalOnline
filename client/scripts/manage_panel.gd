extends PanelContainer
class_name ManagePanel

signal closed
signal save_building(payload: Dictionary)
signal save_shop(payload: Dictionary)
signal stock_shop(shop_id: String, qty: int)

@onready var title: Label = $Margin/VBox/Title
@onready var sub: Label = $Margin/VBox/Sub
@onready var body: VBoxContainer = $Margin/VBox/Body
@onready var hint: Label = $Margin/VBox/Hint

var kind: String = ""
var target_id: String = ""
var _recipe: OptionButton
var _input: OptionButton
var _firm: OptionButton
var _auto: CheckBox
var _sell: CheckBox
var _shop_type: OptionButton
var _shop_item: OptionButton
var _price: SpinBox
var _stock_src: OptionButton
var _restock: CheckBox

func _ready() -> void:
	visible = false

func open_building(b: Dictionary, catalog: Dictionary, world: Dictionary, my_firm: String) -> void:
	kind = "building"
	target_id = str(b.get("id", ""))
	_clear_body()
	title.text = _building_name(str(b.get("type", "")), catalog)
	sub.text = "Специализация: только рецепты «%s»" % b.get("type", "")
	hint.text = "Как в Capitalism Lab: выберите продукт, источник входов и можно ли продавать другим."

	_recipe = OptionButton.new()
	var btype := str(b.get("type", ""))
	for r in catalog.get("recipes", []):
		if str(r.get("building_type", "")) != btype:
			continue
		var label := "%s → %s" % [r.get("name", r.get("id")), r.get("output", {}).get("item_id", "")]
		_recipe.add_item(label)
		_recipe.set_item_metadata(_recipe.item_count - 1, str(r.get("id")))
		if str(r.get("id")) == str(b.get("recipe_id", "")):
			_recipe.select(_recipe.item_count - 1)
	_add_field("Что производить", _recipe)

	_input = OptionButton.new()
	for pair in [["own", "Со своего склада"], ["city", "У города (low quality)"], ["market", "У других (биржа)"], ["firm", "У конкретной фирмы"]]:
		_input.add_item(pair[1])
		_input.set_item_metadata(_input.item_count - 1, pair[0])
		if str(b.get("default_input", {}).get("mode", "own")) == pair[0]:
			_input.select(_input.item_count - 1)
	_add_field("Закупка входов", _input)

	_firm = OptionButton.new()
	_firm.add_item("—")
	_firm.set_item_metadata(0, "")
	var prefer := str(b.get("default_input", {}).get("firm_id", ""))
	for fid in world.get("firms", {}).keys():
		if str(fid) == my_firm:
			continue
		var f: Dictionary = world["firms"][fid]
		_firm.add_item(str(f.get("name", fid)))
		_firm.set_item_metadata(_firm.item_count - 1, str(fid))
		if str(fid) == prefer:
			_firm.select(_firm.item_count - 1)
	_add_field("Фирма-поставщик", _firm)

	_auto = CheckBox.new()
	_auto.text = "Автозакупка перед производством"
	_auto.button_pressed = bool(b.get("auto_procure", false))
	body.add_child(_auto)

	_sell = CheckBox.new()
	_sell.text = "Разрешить продажу другим игрокам"
	_sell.button_pressed = b.get("sell_external", true) != false
	body.add_child(_sell)

	var save := Button.new()
	save.text = "Сохранить"
	save.pressed.connect(_on_save_building)
	body.add_child(save)
	visible = true

func open_shop(s: Dictionary, catalog: Dictionary) -> void:
	kind = "shop"
	target_id = str(s.get("id", ""))
	_clear_body()
	title.text = _shop_name(str(s.get("shop_type", "")), catalog)
	sub.text = "Ассортимент ограничен типом магазина"
	hint.text = "Продуктовый не продаст электронику — тип витрины как в Capitalism Lab."

	_shop_type = OptionButton.new()
	for sh in catalog.get("shops", []):
		_shop_type.add_item(str(sh.get("name", sh.get("id"))))
		_shop_type.set_item_metadata(_shop_type.item_count - 1, str(sh.get("id")))
		if str(sh.get("id")) == str(s.get("shop_type", "")):
			_shop_type.select(_shop_type.item_count - 1)
	_shop_type.item_selected.connect(func(_i): _refill_shop_items(catalog, ""))
	_add_field("Тип магазина", _shop_type)

	_shop_item = OptionButton.new()
	_add_field("Товар на витрине", _shop_item)
	_refill_shop_items(catalog, str(s.get("item_id", "")))

	_price = SpinBox.new()
	_price.min_value = 1
	_price.max_value = 100000
	_price.value = float(s.get("price", 50))
	_add_field("Розничная цена", _price)

	_stock_src = OptionButton.new()
	for pair in [["own", "Со своего склада"], ["city", "У города"], ["market", "С биржи"]]:
		_stock_src.add_item(pair[1])
		_stock_src.set_item_metadata(_stock_src.item_count - 1, pair[0])
		if str(s.get("stock_source", "own")) == pair[0]:
			_stock_src.select(_stock_src.item_count - 1)
	_add_field("Откуда пополнять", _stock_src)

	_restock = CheckBox.new()
	_restock.text = "Автопополнение"
	_restock.button_pressed = s.get("auto_restock", true) != false
	body.add_child(_restock)

	var save := Button.new()
	save.text = "Сохранить"
	save.pressed.connect(_on_save_shop)
	body.add_child(save)

	var stock := Button.new()
	stock.text = "Пополнить +10 со склада"
	stock.pressed.connect(func(): stock_shop.emit(target_id, 10))
	body.add_child(stock)
	visible = true

func _refill_shop_items(catalog: Dictionary, selected: String) -> void:
	if _shop_item == null or _shop_type == nil:
		return
	_shop_item.clear()
	var typ := str(_shop_type.get_item_metadata(_shop_type.selected))
	var shop_def := {}
	for sh in catalog.get("shops", []):
		if str(sh.get("id")) == typ:
			shop_def = sh
			break
	for it in catalog.get("items", []):
		if not _shop_allows(shop_def, it):
			continue
		_shop_item.add_item(str(it.get("name", it.get("id"))))
		_shop_item.set_item_metadata(_shop_item.item_count - 1, str(it.get("id")))
		if str(it.get("id")) == selected:
			_shop_item.select(_shop_item.item_count - 1)

func _shop_allows(shop_def: Dictionary, item: Dictionary) -> bool:
	var allowed_items: Array = shop_def.get("allowed_items", [])
	if not allowed_items.is_empty():
		return str(item.get("id")) in allowed_items
	var cats: Array = shop_def.get("allowed_categories", [])
	var tiers: Array = shop_def.get("allowed_tiers", [])
	if not cats.is_empty() and not (str(item.get("category")) in cats):
		return false
	if not tiers.is_empty() and not (str(item.get("tier")) in tiers):
		return false
	return true

func _on_save_building() -> void:
	var recipe_id := ""
	if _recipe and _recipe.item_count > 0:
		recipe_id = str(_recipe.get_item_metadata(_recipe.selected))
	var mode := "own"
	if _input and _input.item_count > 0:
		mode = str(_input.get_item_metadata(_input.selected))
	var firm_id := ""
	if _firm and _firm.item_count > 0:
		firm_id = str(_firm.get_item_metadata(_firm.selected))
	save_building.emit({
		"building_id": target_id,
		"recipe_id": recipe_id,
		"default_input": mode,
		"firm_id": firm_id,
		"auto_procure": _auto.button_pressed if _auto else false,
		"sell_external": _sell.button_pressed if _sell else true,
	})
	visible = false

func _on_save_shop() -> void:
	save_shop.emit({
		"shop_id": target_id,
		"shop_type": str(_shop_type.get_item_metadata(_shop_type.selected)) if _shop_type else "",
		"item_id": str(_shop_item.get_item_metadata(_shop_item.selected)) if _shop_item and _shop_item.item_count > 0 else "",
		"price": _price.value if _price else 50,
		"stock_source": str(_stock_src.get_item_metadata(_stock_src.selected)) if _stock_src else "own",
		"auto_restock": _restock.button_pressed if _restock else true,
	})
	visible = false

func _on_close() -> void:
	visible = false
	closed.emit()

func _clear_body() -> void:
	for c in body.get_children():
		c.queue_free()
	_recipe = null
	_input = null
	_firm = null
	_auto = null
	_sell = null
	_shop_type = null
	_shop_item = null
	_price = null
	_stock_src = null
	_restock = null

func _add_field(label: String, control: Control) -> void:
	var l := Label.new()
	l.text = label
	body.add_child(l)
	body.add_child(control)

func _building_name(id: String, catalog: Dictionary) -> String:
	for b in catalog.get("buildings", []):
		if str(b.get("id")) == id:
			return str(b.get("name", id))
	return id

func _shop_name(id: String, catalog: Dictionary) -> String:
	for s in catalog.get("shops", []):
		if str(s.get("id")) == id:
			return str(s.get("name", id))
	return id
