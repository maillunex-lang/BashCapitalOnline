extends PanelContainer
class_name ManagePanel

signal closed
signal save_building(payload: Dictionary)
signal save_shop(payload: Dictionary)
signal rent_shop(payload: Dictionary)
signal stock_shop(shop_id: String, qty: int)
signal transfer_stock(payload: Dictionary)
signal asset_chosen(asset_kind: String, asset_id: String)
signal bread_parcel_requested(step: Dictionary)

@onready var title: Label = $Margin/VBox/Header/Title
@onready var icon: TextureRect = $Margin/VBox/Header/Icon
@onready var sub: Label = $Margin/VBox/Sub
@onready var scroll: ScrollContainer = $Margin/VBox/Scroll
@onready var body: VBoxContainer = $Margin/VBox/Scroll/Body
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
var _snapshot: Label
var _bread_rows: Array[Label] = []
var _bread_buttons: Array[Button] = []
var _inventory_rows: ItemList
var _inventory_heading: Label
var _transfer: Button
var _current_world: Dictionary = {}
var api: ApiClient

func _ready() -> void:
	$Margin/VBox/CloseBtn.custom_minimum_size.y = EconomyView.touch_height()
	visible = false

func open_building(b: Dictionary, catalog: Dictionary, world: Dictionary, my_firm: String) -> void:
	kind = "building"
	target_id = str(b.get("id", ""))
	_clear_body()
	var definition := CatalogUI.find(catalog, "buildings", str(b.get("type", "")))
	title.text = str(definition.get("name", b.get("type", "")))
	icon.texture = CatalogUI.icon(definition, "buildings")
	sub.text = "Производство и склад"
	hint.text = ""
	_snapshot = _add_summary()
	_inventory_heading = _add_summary()
	_inventory_heading.text = "Склад пуст"
	_inventory_rows = EconomyView.list_view()
	_inventory_rows.hide()
	body.add_child(_inventory_rows)
	_current_world = world

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
	var recipe_info := _add_summary()
	var show_recipe := func():
		var recipe := CatalogUI.find(catalog, "recipes", CatalogUI.selected_id(_recipe))
		recipe_info.text = EconomyView.recipe_text(recipe, catalog)
		recipe_info.visible = not recipe_info.text.is_empty()
	_recipe.item_selected.connect(func(_index): show_recipe.call())
	show_recipe.call()

	_input = OptionButton.new()
	for pair in [["own", "Со своего склада"], ["city", "У города"], ["market", "У других (биржа)"], ["firm", "У конкретной фирмы"]]:
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
	_add_touch_control(_auto)

	_sell = CheckBox.new()
	_sell.text = "Разрешить продажу другим игрокам"
	_sell.button_pressed = b.get("sell_external", true) != false
	_add_touch_control(_sell)

	var save := Button.new()
	save.text = "Сохранить"
	save.pressed.connect(_on_save_building)
	_add_touch_control(save)
	_transfer = Button.new()
	_transfer.text = "Склад и отправка груза"
	_transfer.pressed.connect(func():
		var latest: Dictionary = _current_world.get("buildings", {}).get(target_id, b)
		_open_transfer(latest, catalog, _current_world, my_firm, latest.get("inventory", {}))
	)
	_transfer.disabled = true
	_add_touch_control(_transfer)
	_show_building_snapshot(b, catalog)
	visible = true

func _open_transfer(b: Dictionary, catalog: Dictionary, world: Dictionary, my_firm: String, inventory: Dictionary) -> void:
	_clear_body()
	sub.text = "Склад и отправка груза"
	hint.text = ""
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
	qty.value = 1
	qty.max_value = maxi(1, int(inventory.get(CatalogUI.selected_id(item), 0)))
	item.item_selected.connect(func(_index): qty.max_value = int(inventory[CatalogUI.selected_id(item)]))
	_add_field("Количество", qty)

	var estimate := _add_summary()
	var retry := Button.new()
	retry.text = "Повторить расчёт"
	_add_touch_control(retry)
	retry.hide()
	var send := Button.new()
	send.text = "Отправить"
	send.disabled = true
	send.pressed.connect(func():
		visible = false
		transfer_stock.emit({
			"from_building_id": b["id"],
			"to_building_id": CatalogUI.selected_id(destination),
			"item_id": CatalogUI.selected_id(item),
			"qty": int(qty.value),
		})
	)
	_add_touch_control(send)
	var route := {"request": 0, "quote": {}}
	var via_truck := int(world.get("firms", {}).get(my_firm, {}).get("trucks", 0)) > 0
	var display_quote := func():
		var quote: Dictionary = route["quote"]
		if quote.is_empty():
			return
		estimate.text = "%.2f ₽ за %d ед. · %d тик. · %s" % [
			float(quote["cost_per_unit"]) * qty.value, int(qty.value), int(quote["ticks"]),
			"Своя фура" if via_truck else "Наёмная доставка",
		]
		send.disabled = false
	var update_quote := func():
		route["request"] += 1
		var seq: int = route["request"]
		route["quote"] = {}
		send.disabled = true
		retry.hide()
		var building: Dictionary = world.get("buildings", {}).get(CatalogUI.selected_id(destination), {})
		if item.item_count == 0 or building.is_empty():
			estimate.text = "Для отправки нужен товар и второе здание вашей фирмы."
			return
		estimate.text = "Рассчитываем доставку…"
		var quote := await api.route_quote(b["city_id"], building["city_id"])
		if not is_instance_valid(estimate) or not estimate.is_inside_tree() or seq != route["request"]:
			return
		if not quote.has("cost_per_unit") or not quote.has("ticks"):
			estimate.text = "Сервер не вернул расчёт доставки."
			retry.show()
			return
		if via_truck:
			# Older V2 servers publish only the regular tariff.
			quote = quote.get("fleet", {
				"ticks": maxi(1, int(quote["ticks"]) - 1),
				"cost_per_unit": float(quote["cost_per_unit"]) * 0.55,
			})
		route["quote"] = quote
		display_quote.call()
	destination.item_selected.connect(func(_index): update_quote.call())
	qty.value_changed.connect(func(_value): display_quote.call())
	retry.pressed.connect(update_quote)
	update_quote.call()

func open_shop(s: Dictionary, catalog: Dictionary) -> void:
	kind = "shop"
	target_id = str(s.get("id", ""))
	_clear_body()
	var definition := CatalogUI.find(catalog, "shops", str(s.get("shop_type", "")))
	title.text = str(definition.get("name", s.get("shop_type", "")))
	icon.texture = CatalogUI.icon(definition, "shops")
	sub.text = "Розничная торговля"
	hint.text = ""
	_snapshot = _add_summary()
	_inventory_rows = EconomyView.list_view()
	_inventory_rows.add_item("", null, false)
	body.add_child(_inventory_rows)
	_show_shop_snapshot(s, catalog)
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
	_add_touch_control(_restock)

	var save := Button.new()
	save.text = "Сохранить"
	save.pressed.connect(_on_save_shop)
	save.disabled = _shop_item.item_count == 0
	_shop_type.item_selected.connect(func(_index): save.disabled = _shop_item.item_count == 0)
	_add_touch_control(save)

	var stock := Button.new()
	stock.text = "Пополнить на 10 ед."
	stock.pressed.connect(func(): stock_shop.emit(target_id, 10))
	_add_touch_control(stock)
	visible = true

func open_rental(catalog: Dictionary, parcel: Dictionary, initial := {}) -> void:
	kind = "rental"
	_clear_body()
	title.text = "Аренда магазина"
	icon.texture = CatalogUI.icon({}, "shops")
	sub.text = "Участок (%s, %s)" % [parcel["x"], parcel["y"]]
	hint.text = ""
	_shop_fields(catalog, initial)
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
	_add_touch_control(confirm)
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
	_add_field("Розничная цена, ₽", _price)
	var price_info := _add_summary()
	var show_price := func():
		var id := CatalogUI.selected_id(_shop_item)
		var corridor: Dictionary = catalog.get("economy", {}).get("price_corridors", {}).get(id, {})
		price_info.text = ""
		if not corridor.is_empty():
			price_info.text = "Допустимая цена %.2f–%.2f ₽" % [float(corridor["min"]), float(corridor["max"])]
		price_info.visible = not price_info.text.is_empty()
	_shop_item.item_selected.connect(func(_index): show_price.call())
	_shop_type.item_selected.connect(func(_index): show_price.call())
	show_price.call()

func open_company(catalog: Dictionary, world: Dictionary, firm_id: String) -> void:
	kind = "company"
	_clear_body()
	var firm: Dictionary = world.get("firms", {}).get(firm_id, {})
	title.text = str(firm.get("name", "Компания"))
	icon.texture = load("res://assets/ui/icon_firm.png")
	sub.text = "Имущество и грузы"
	hint.text = ""
	if BreadGuide.available(catalog):
		var guide := Button.new()
		guide.text = "Первая цепочка · хлеб"
		guide.pressed.connect(func(): open_bread_guide(catalog, _current_world, firm_id))
		_add_touch_control(guide)
	_snapshot = _add_summary()
	for section in [["buildings", "Здания"], ["shops", "Магазины"]]:
		var label := _add_summary()
		label.text = section[1]
		var entries: Array = world.get(section[0], {}).values()
		var count := 0
		for asset in entries:
			if asset["firm_id"] != firm_id:
				continue
			count += 1
			var type_id: String = asset.get("type", asset.get("shop_type", ""))
			var definition := CatalogUI.find(catalog, section[0], type_id)
			var city := CatalogUI.entry_name(catalog, "cities", asset["city_id"])
			var btn := Button.new()
			btn.text = "%s · %s (%d, %d)" % [definition.get("name", type_id), city, int(asset["x"]), int(asset["y"])]
			btn.icon = CatalogUI.icon(definition, section[0])
			btn.add_theme_constant_override("icon_max_width", 30)
			btn.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS
			btn.tooltip_text = btn.text
			btn.pressed.connect(asset_chosen.emit.bind(section[0], asset["id"]))
			_add_touch_control(btn)
		if count == 0:
			label.text += " · нет"
	refresh_snapshot(catalog, world, firm_id)
	visible = true

func open_bread_guide(catalog: Dictionary, world: Dictionary, firm_id: String) -> void:
	kind = "bread_guide"
	_clear_body()
	title.text = "Первая цепочка"
	sub.text = "Уфа · пшеница → мука → хлеб → магазин"
	hint.text = ""
	icon.texture = CatalogUI.icon(CatalogUI.find(catalog, "items", "wheat_bread"), "items")
	_snapshot = _add_summary()
	for step in BreadGuide.STEPS:
		_bread_rows.append(_add_summary())
		var action := Button.new()
		action.pressed.connect(func():
			var asset := BreadGuide.asset_for(step, _current_world, firm_id)
			if asset.is_empty():
				bread_parcel_requested.emit(step)
			else:
				asset_chosen.emit("shops" if step.has("item") else "buildings", asset["id"])
		)
		_bread_buttons.append(action)
		_add_touch_control(action)
	var back := Button.new()
	back.text = "Назад к компании"
	back.pressed.connect(func(): open_company(catalog, _current_world, firm_id))
	_add_touch_control(back)
	refresh_snapshot(catalog, world, firm_id)
	show()

func _show_bread_progress(world: Dictionary, firm_id: String) -> void:
	var ready := 0
	for i in BreadGuide.STEPS.size():
		var step: Dictionary = BreadGuide.STEPS[i]
		var asset := BreadGuide.asset_for(step, world, firm_id)
		var complete := BreadGuide.matches(step, asset)
		ready += int(complete)
		var heading := ("Готово · " if complete else "%d. " % (i + 1)) + str(step["title"])
		var details := str(step["hint"])
		if not asset.is_empty():
			if complete and step.has("item"):
				details = "На витрине: %d ед. хлеба. Продажи идут по игровым тикам." % int(asset.get("stock", 0))
			elif complete:
				var batch: Dictionary = asset.get("production", {})
				if batch.get("item_id") == step["output"] and int(batch.get("qty", 0)) > 0:
					details += " Последний выпуск: %d ед." % int(batch["qty"])
			else:
				details = "Откройте объект и выберите указанный товар и источник «Со своего склада». Включите автообеспечение."
		var no_parcel := asset.is_empty() and BreadGuide.suggest_parcel(step, world, firm_id).is_empty()
		if no_parcel:
			details += " Сейчас подходящих свободных участков нет."
		_bread_rows[i].text = heading + "\n" + details
		_bread_buttons[i].text = "Выбрать участок" if asset.is_empty() else "Открыть объект"
		_bread_buttons[i].disabled = no_parcel
	_snapshot.text = "%d из %d объектов готовы" % [ready, BreadGuide.STEPS.size()]
	if ready == BreadGuide.STEPS.size():
		_snapshot.text += " · магазин может продавать хлеб по тикам"

func refresh_snapshot(catalog: Dictionary, world: Dictionary, firm_id: String) -> void:
	if _snapshot == null:
		return
	_current_world = world
	match kind:
		"bread_guide":
			_show_bread_progress(world, firm_id)
		"building":
			var building: Dictionary = world.get("buildings", {}).get(target_id, {})
			if not building.is_empty():
				_show_building_snapshot(building, catalog)
		"shop":
			var shop: Dictionary = world.get("shops", {}).get(target_id, {})
			if not shop.is_empty():
				_show_shop_snapshot(shop, catalog)
		"company":
			var firm: Dictionary = world.get("firms", {}).get(firm_id, {})
			var lines := PackedStringArray(["Касса %.0f ₽ · Долг %.0f ₽ · Фуры %d" % [
				float(firm.get("cash", 0)), float(firm.get("credit", 0)), int(firm.get("trucks", 0)),
			]])
			for shipment in world.get("shipments", {}).values():
				if shipment["firm_id"] != firm_id:
					continue
				lines.append("%s · %s → %s · %d тик." % [
					EconomyView.quantity_text(shipment, catalog),
					CatalogUI.entry_name(catalog, "cities", shipment["from_city"]),
					CatalogUI.entry_name(catalog, "cities", shipment["to_city"]),
					int(shipment["ticks_left"]),
				])
			if lines.size() == 1:
				lines.append("Грузов в пути нет")
			_snapshot.text = "\n".join(lines)

func _show_building_snapshot(building: Dictionary, catalog: Dictionary) -> void:
	_snapshot.text = EconomyView.production_text(building, catalog)
	var inventory: Dictionary = building.get("inventory", {}) if building.get("inventory") is Dictionary else {}
	if inventory == _inventory_rows.get_meta("snapshot", {}):
		return
	_inventory_rows.set_meta("snapshot", inventory.duplicate())
	_inventory_rows.clear()
	for id in inventory:
		var qty := int(inventory[id])
		if qty > 0:
			EconomyView.add_item_row(_inventory_rows, catalog, id, EconomyView.quantity_text({"item_id": id, "qty": qty}, catalog))
	_inventory_rows.visible = _inventory_rows.item_count > 0
	_transfer.disabled = not _inventory_rows.visible
	_inventory_heading.text = "На складе" if _inventory_rows.visible else "Склад пуст"

func _show_shop_snapshot(shop: Dictionary, catalog: Dictionary) -> void:
	var id := str(shop.get("item_id", ""))
	var product := CatalogUI.entry_name(catalog, "items", id)
	_snapshot.text = "Витрина · %.2f ₽ за ед." % float(shop.get("price", 0))
	var item_icon := CatalogUI.icon(CatalogUI.find(catalog, "items", id), "items")
	_inventory_rows.set_item_icon(0, item_icon)
	_inventory_rows.set_item_text(0, EconomyView.quantity_text({"item_id": id, "qty": shop.get("stock", 0)}, catalog))
	_inventory_rows.set_item_tooltip(0, product)

func _add_summary() -> Label:
	var label := Label.new()
	label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	label.add_theme_font_size_override("font_size", 18)
	label.add_theme_color_override("font_color", Color("#bfdad1"))
	body.add_child(label)
	return label

func _add_touch_control(control: Control) -> void:
	control.custom_minimum_size.y = EconomyView.touch_height()
	if control is OptionButton:
		control.fit_to_longest_item = false
		control.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS
		control.get_popup().add_theme_constant_override("v_separation", EconomyView.touch_height() - 28)
	body.add_child(control)

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
	scroll.scroll_vertical = 0
	_bread_rows.clear()
	_bread_buttons.clear()
	_snapshot = null
	_inventory_rows = null
	_transfer = null
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
	l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	body.add_child(l)
	_add_touch_control(control)
