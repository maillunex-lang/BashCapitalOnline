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
	await game.api.manual_tick()
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
	if not _check(not send.disabled, "shipment form has a product and a destination"):
		quit(1)
		return
	send.pressed.emit()
	await game.api.world_updated
	await process_frame
	var shipment: Dictionary = {}
	for row in game.api.world["shipments"].values():
		if row["firm_id"] == id:
			shipment = row
	if _check(not shipment.is_empty(), "send control queued a shipment"):
		_check(shipment["to_building_id"] == warehouse["id"] and shipment["item_id"] == "wheat", "form sent wheat to the selected warehouse")
		for _tick in int(shipment["ticks_left"]):
			await game.api.manual_tick()
		var target: Dictionary = game.api.world["buildings"][warehouse["id"]]
		_check(target["inventory"]["wheat"] == shipment["qty"], "shipment arrived in the warehouse")

	game.queue_free()
	await process_frame
	var reconnected = await _start_game(args[0], false)
	_check(reconnected.api.firm_id == id and reconnected.api.world["buildings"].has(warehouse["id"]), "login reopened the same firm and warehouse")
	print("Economy controls smoke: %s" % ("failed" if failed else "passed"))
	quit(1 if failed else 0)

func _start_game(url: String, register: bool):
	var game = load("res://scenes/main.tscn").instantiate()
	root.add_child(game)
	game.url_edit.text = url
	game.username.text = "godot_economy_smoke"
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
