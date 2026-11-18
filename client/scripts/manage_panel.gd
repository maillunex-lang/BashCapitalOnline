extends PanelContainer
class_name ManagePanel

signal closed
signal save_building(payload: Dictionary)
signal save_shop(payload: Dictionary)
signal rent_shop(payload: Dictionary)
signal stock_shop(shop_id: String, qty: int)
signal transfer_stock(payload: Dictionary)

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
	title.text = CatalogUI.entry_name(catalog, "buildings", str(b.get("type", "")))
	sub.text = "Производство"
	hint.text = ""

	_recipe = OptionButton.new()
	var btype := str(b.get("type", ""))
	for r in catalog.get("recipes", []):
		if str(r.get("building_type", "")) != btype:
			continue
		var output := CatalogUI.find(catalog, "items", r["output"]["item_id"])
		_recipe.add_icon_item(CatalogUI.icon(output, "items"), r["name"])
		_recipe.set_item_metadata(_recipe.item_count - 1, str(r.get("id")))
		if str(r.get("id")) == str(b.get("recipe_id", "")):
			_recipe.select(_recipe.item_count - 1)
	_recipe.add_theme_constant_override("icon_max_width", 28)
	_recipe.get_popup().add_theme_constant_override("icon_max_width", 28)
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
	var inventory: Dictionary = b.get("inventory", {}) if b.get("inventory") is Dictionary else {}
	if inventory.values().any(func(quantity): return int(quantity) > 0):
		var transfer := Button.new()
		transfer.text = "Склад и отправка груза"
		transfer.pressed.connect(_open_transfer.bind(b, catalog, world, my_firm, inventory))
		body.add_child(transfer)
	visible = true

func _open_transfer(b: Dictionary, catalog: Dictionary, world: Dictionary, my_firm: String, inventory: Dictionary) -> void:
	_clear_body()
	sub.text = "Склад и отправка груза"
	hint.text = "Доставка занимает от одного тика и оплачивается при отправке."
	var item := OptionButton.new()
	for id in inventory:
		if int(inventory[id]) > 0:
			item.add_item("%s ×%d" % [CatalogUI.entry_name(catalog, "items", id), int(inventory[id])])
			item.set_item_metadata(item.item_count - 1, id)
	_add_field("Товар на складе", item)

	var destination := OptionButton.new()
	destination.fit_to_longest_item = false
	destination.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS
	for other in world.get("buildings", {}).values():
		if other["firm_id"] != my_firm or other["id"] == b["id"]:
			continue
		var city := CatalogUI.entry_name(catalog, "cities", other["city_id"])
		var name := CatalogUI.entry_name(catalog, "buildings", other["type"])
		destination.add_item("%s · %s (%d, %d)" % [city, name, int(other["x"]), int(other["y"])])
		destination.set_item_metadata(destination.item_count - 1, other["id"])
	_add_field("Здание назначения", destination)

	var qty := SpinBox.new()
	qty.min_value = 1
	qty.max_value = maxi(1, int(inventory.get(CatalogUI.selected_id(item), 0)))
	item.item_selected.connect(func(_index): qty.max_value = int(inventory[CatalogUI.selected_id(item)]))
	_add_field("Количество", qty)

	var send := Button.new()
	send.text = "Отправить"
	send.disabled = item.item_count == 0 or destination.item_count == 0
	send.pressed.connect(func():
		visible = false
		transfer_stock.emit({
			"from_building_id": b["id"],
			"to_building_id": CatalogUI.selected_id(destination),
			"item_id": CatalogUI.selected_id(item),
			"qty": int(qty.value),
		})
	)
	body.add_child(send)

func open_shop(s: Dictionary, catalog: Dictionary) -> void:
	kind = "shop"
	target_id = str(s.get("id", ""))
	_clear_body()
	title.text = CatalogUI.entry_name(catalog, "shops", str(s.get("shop_type", "")))
	sub.text = "Ассортимент ограничен типом магазина"
	hint.text = ""
	_shop_fields(catalog, s)

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
	save.disabled = _shop_item.item_count == 0
	_shop_type.item_selected.connect(func(_index): save.disabled = _shop_item.item_count == 0)
	body.add_child(save)

	var stock := Button.new()
	stock.text = "Пополнить +10 со склада"
	stock.pressed.connect(func(): stock_shop.emit(target_id, 10))
	body.add_child(stock)
	visible = true

func open_rental(catalog: Dictionary, parcel: Dictionary) -> void:
	kind = "rental"
	_clear_body()
	title.text = "Аренда магазина"
	sub.text = "Участок (%s, %s)" % [parcel["x"], parcel["y"]]
	hint.text = ""
	_shop_fields(catalog, {})
	var confirm := Button.new()
	confirm.text = "Арендовать"
	confirm.disabled = _shop_item.item_count == 0
	_shop_type.item_selected.connect(func(_index): confirm.disabled = _shop_item.item_count == 0)
	confirm.pressed.connect(func():
		var payload := parcel.duplicate()
		payload["shop_type"] = CatalogUI.selected_id(_shop_type)
		payload["item_id"] = CatalogUI.selected_id(_shop_item)
		payload["price"] = _price.value
		visible = false
		rent_shop.emit(payload)
	)
	body.add_child(confirm)
	visible = true

func _shop_fields(catalog: Dictionary, shop: Dictionary) -> void:
	_shop_type = OptionButton.new()
	for entry in catalog.get("shops", []):
		CatalogUI.add_option(_shop_type, entry, "shops", str(shop.get("shop_type", "")))
	_add_field("Тип магазина", _shop_type)
	_shop_item = OptionButton.new()
	_add_field("Товар на витрине", _shop_item)
	CatalogUI.fill_shop_items(_shop_item, catalog, CatalogUI.selected_id(_shop_type), str(shop.get("item_id", "")))
	_shop_type.item_selected.connect(func(_index):
		CatalogUI.fill_shop_items(_shop_item, catalog, CatalogUI.selected_id(_shop_type))
	)
	_price = SpinBox.new()
	_price.min_value = 1
	_price.max_value = 100000
	_price.value = float(shop.get("price", 55))
	_add_field("Розничная цена", _price)

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
		"shop_type": CatalogUI.selected_id(_shop_type),
		"item_id": CatalogUI.selected_id(_shop_item),
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
		body.remove_child(c)
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
