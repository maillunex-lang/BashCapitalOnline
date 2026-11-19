extends SceneTree

var failed := false

func _initialize() -> void:
	_run.call_deferred()

func _run() -> void:
	var args := OS.get_cmdline_user_args()
	if args.size() != 1:
		printerr("Pass the URL of a fresh V2 server after --")
		quit(2)
		return

	var game = await _start_game(args[0], true)
	if not _check(not game.api.firm_id.is_empty() and not game.api.session_token.is_empty(), "registered account and firm"):
		quit(1)
		return
	var id: String = game.api.firm_id
	var farm := await _build(game, "ufa", "grain_farm", "extract_wheat", "wheat")
	if not _check(farm.has("id"), "built farm"):
		quit(1)
		return
	game._on_asset_chosen("buildings", farm["id"])
	var initial_summary: Label = game.manage_panel._snapshot
	_check(initial_summary.text.contains("Ожидается"), "new farm waits for first tick")
	await game._on_tick_pressed()
	_check(game.manage_panel._snapshot == initial_summary, "tick refresh preserves the open form")
	_check(initial_summary.text.contains("Партия выпущена") and initial_summary.text.contains("Последний выпуск"), "live panel shows the server's successful batch")
	_check(game.manage_panel._inventory_rows.item_count == 1, "live panel shows the new wheat stock")
	_check(game.manage_panel._inventory_rows.get_item_icon(0) != null, "stock has a visible icon")
	_check(game.manage_panel._transfer.custom_minimum_size.y >= 48, "transfer has a large touch target")
	var warehouse := await _build(game, "sterlitamak", "warehouse", "", "")
	if not _check(farm.has("id") and warehouse.has("id"), "built farm and warehouse"):
		quit(1)
		return

	var world: Dictionary = game.api.world
	var live_farm: Dictionary = world["buildings"][farm["id"]]
	game.manage_panel.open_building(live_farm, game.api.catalog, world, id)
	var open_button: Button = game.manage_panel.body.get_child(-1)
	if not _check(open_button.text == "Склад и отправка груза", "farm shows a stock transfer control"):
		quit(1)
		return
	open_button.pressed.emit()
	var send: Button = game.manage_panel.body.get_child(-1)
	if not _check(await _wait_for_quote(send), "shipment form received the route quote"):
		quit(1)
		return
	var estimate: Label = game.manage_panel.body.get_child(-3)
	_check(estimate.text.contains("Наёмная доставка"), "estimate shows the hired transport tariff")
	var regular: Dictionary = await game.api.route_quote("ufa", "sterlitamak")
	_check(estimate.text.begins_with("%.2f ₽" % float(regular["cost_per_unit"])), "regular one-unit price matches the API")
	var truck: Dictionary = await game.api.action("buy_truck", {})
	_check(not truck.has("error"), "firm buys a truck")
	world = game.api.world
	live_farm = world["buildings"][farm["id"]]
	game.manage_panel.open_building(live_farm, game.api.catalog, world, id)
	open_button = game.manage_panel.body.get_child(-1)
	open_button.pressed.emit()
	send = game.manage_panel.body.get_child(-1)
	if not _check(await _wait_for_quote(send), "shipment form received the fleet quote"):
		quit(1)
		return
	estimate = game.manage_panel.body.get_child(-3)
	var qty: SpinBox = game.manage_panel.body.find_children("*", "SpinBox", false, false)[0]
	qty.value = 2
	var stock_world := game.api.world.duplicate(true) as Dictionary
	stock_world["buildings"][farm["id"]]["inventory"]["wheat"] = 1
	game.manage_panel.refresh_snapshot(game.api.catalog, stock_world, id)
	_check(send.disabled and estimate.text.contains("недостаточно товара"), "a live stock change blocks dispatch")
	var low_cash_world := game.api.world.duplicate(true) as Dictionary
	low_cash_world["firms"][id]["cash"] = 0
	game.manage_panel.refresh_snapshot(game.api.catalog, low_cash_world, id)
	_check(send.disabled and estimate.text.contains("не хватает"), "an unaffordable dispatch is blocked")
	game.manage_panel.refresh_snapshot(game.api.catalog, game.api.world, id)
	var sent_estimate := estimate.text
	_check(not send.disabled and sent_estimate.contains("на складе получателя 0 ед."), "stock and dispatch recover on the next snapshot")
	_check(sent_estimate.contains("Своя фура") and sent_estimate.contains("за 2 ед."), "quote updates when quantity changes")
	send.pressed.emit()
	await game.api.world_updated
	await process_frame
	var shipment: Dictionary = {}
	for row in game.api.world["shipments"].values():
		if row["firm_id"] == id:
			shipment = row
	if _check(not shipment.is_empty(), "send control queued a shipment"):
		_check(sent_estimate.begins_with("%.2f ₽ за 2 ед. · %d тик." % [float(shipment["cost_paid"]), int(shipment["ticks_left"])]), "dispatch matches the displayed fleet estimate")
		_check(shipment["to_building_id"] == warehouse["id"] and shipment["item_id"] == "wheat", "form sent wheat to the selected warehouse")
		game._on_company_pressed()
		_check(game.manage_panel.kind == "company" and game.manage_panel._snapshot.text.contains("Уфа") and game.manage_panel._snapshot.text.contains("Стерлитамак") and game.manage_panel._snapshot.text.contains("прибытие"), "company shows the shipment route")
		await process_frame
		await process_frame
		await _check_company_layout(game)
		for button in game.manage_panel.body.find_children("*", "Button", false, false):
			if button.text == "История доставок":
				button.pressed.emit()
				break
		_check(game.manage_panel.kind == "shipment_history" and game.manage_panel._snapshot.text.contains("Поступлений пока нет"), "history is empty while the shipment is in transit")
		for _tick in int(shipment["ticks_left"]):
			await game._on_tick_pressed()
		var target: Dictionary = game.api.world["buildings"][warehouse["id"]]
		_check(target["inventory"]["wheat"] == shipment["qty"], "shipment arrived in the warehouse")
		_check(game.manage_panel._snapshot.text.contains("Доставлено") and game.manage_panel._snapshot.text.contains("Своя фура"), "open history updates when cargo arrives")

	game.queue_free()
	await process_frame
	var reconnected = await _start_game(args[0], false)
	_check(reconnected.api.firm_id == id and reconnected.api.world["buildings"].has(warehouse["id"]), "login reopened the same firm and warehouse")
	print("Economy controls smoke: %s" % ("failed" if failed else "passed"))
	quit(1 if failed else 0)

func _wait_for_quote(send: Button) -> bool:
	var deadline := Time.get_ticks_msec() + 3000
	while send.disabled and Time.get_ticks_msec() < deadline:
		await process_frame
	return not send.disabled

func _check_company_layout(game) -> void:
	var panel: ManagePanel = game.manage_panel
	var safe: Control = game.get_node("SafeArea/UI")
	var close_button: Button = panel.get_node("Margin/VBox/CloseBtn")
	var settings_button: Button = game.get_node("SafeArea/UI/TopBar/SettingsBtn")
	var previous_scale_size := root.content_scale_size
	var previous_window_size := root.size
	for layout in [Vector2i(1600, 900), Vector2i(2000, 900)]:
		root.content_scale_size = layout
		root.size = layout
		await process_frame
		await process_frame
		_check(game.get_viewport_rect().size == Vector2(layout), "viewport uses the requested landscape size")
		_check(safe.get_global_rect().encloses(panel.get_global_rect()), "panel fits inside the safe area")
		_check(panel.get_global_rect().encloses(close_button.get_global_rect()), "close button remains inside the panel")
		_check(safe.get_global_rect().encloses(game.company_btn.get_global_rect()), "company button fits inside the safe area")
		_check(not game.company_btn.get_global_rect().intersects(settings_button.get_global_rect()), "top-bar buttons do not overlap")
		print("Economy panel layout: %dx%d" % [layout.x, layout.y])
	_check(panel.scroll.horizontal_scroll_mode == ScrollContainer.SCROLL_MODE_DISABLED, "form scrolls vertically")
	root.content_scale_size = previous_scale_size
	root.size = previous_window_size

func _start_game(url: String, register: bool):
	var game = load("res://scenes/main.tscn").instantiate()
	root.add_child(game)
	if register:
		game.url_edit.text = "ftp://invalid"
		await game._on_connect_pressed()
		_check(game.login_panel.visible and game.connection_message.visible, "invalid server error appears in the login form")
		game.url_edit.text = url
		game.username.text = "godot_economy_smoke"
	else:
		_check(game.url_edit.text == url, "reopening the client restores the working server")
		_check(game.username.text == "godot_economy_smoke", "reopening the client restores the login")
		_check(game.password.text.is_empty(), "password is not remembered")
	game.password.text = "test-password"
	game.name_edit.text = "Фирма Godot"
	game.industry_option.select(1)
	await game._on_connect_pressed(register)
	return game

func _build(game, city_id: String, building: String, recipe: String, resource: String) -> Dictionary:
	var city: Dictionary = game.api.world["cities"][city_id]
	for parcel in city["parcels"]:
		if parcel.has("building_id"):
			continue
		var eligible: bool = parcel["zone"] in ["industrial", "empty"]
		if not resource.is_empty():
			eligible = city["deposits"].any(func(d):
				return d["x"] == parcel["x"] and d["y"] == parcel["y"] and d["item_id"] == resource and d["amount"] > 0
			)
		if eligible:
			return await game.api.action("build", {
				"city_id": city_id, "building_type": building, "recipe_id": recipe,
				"x": parcel["x"], "y": parcel["y"],
			})
	return {}

func _check(ok: bool, label: String) -> bool:
	if not ok:
		failed = true
		printerr("FAIL: " + label)
	return ok
