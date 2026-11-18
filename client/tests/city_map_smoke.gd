extends SceneTree

var failures := 0
var draws: Dictionary = {}

func _initialize() -> void:
	_run.call_deferred()

func _run() -> void:
	var args := OS.get_cmdline_user_args()
	if args.size() != 1:
		printerr("Pass the URL of a fresh V2 server after --")
		quit(2)
		return
	var game = load("res://scenes/main.tscn").instantiate()
	root.add_child(game)
	game.url_edit.text = args[0]
	game.name_edit.text = "Карта V2 smoke"
	game.industry_option.select(4)
	await game._on_connect_pressed()
	if game.api.firm_id.is_empty():
		_check(false, "connect and create a firm: " + game.status_label.text)
		quit(1)
		return

	var map: CityMap = game.city_map
	var cities := ["ufa", "sterlitamak", "beloretsk"]
	for city_id in cities:
		await game._on_city_selected(city_id)
		await process_frame
		var city: Dictionary = game.api.world["cities"][city_id]
		_check(map._tiles.size() == city["parcels"].size(), city_id + ": all parcels retained")
		_check(map.selected == CityMap.NO_TILE, city_id + ": selection cleared on city switch")
		_check(not city["deposits"].is_empty(), city_id + ": visible deposits")
		var road_count := 0
		for tile in map._tiles.values():
			road_count += int(tile.visual.road)
		_check(road_count > 0, city_id + ": mapped roads")
		_check(not map._buildings.is_empty(), city_id + ": other companies' buildings appear")
		print("%s: %d parcels, %d road tiles, %d buildings, %d shops" % [
			city_id, map._tiles.size(), road_count, map._buildings.size(), map._shops.size(),
		])

	for coord in map._tiles:
		map._tiles[coord].draw.connect(_record_draw.bind(coord))
	draws.clear()
	var snapshot: Dictionary = game.api.world.duplicate(true)
	game._render_world()
	await process_frame
	await process_frame
	_check(draws.is_empty(), "identical snapshot rebuilds zero parcel draw commands")

	var city: Dictionary = snapshot["cities"]["beloretsk"]
	var deposit_coords := {}
	for deposit in city["deposits"]:
		if deposit["amount"] > 0:
			var coord := _coord(deposit)
			if not deposit_coords.has(coord):
				deposit_coords[coord] = deposit
	var wrong := CityMap.NO_TILE
	var ore := CityMap.NO_TILE
	var mine_items := ["copper_ore", "bauxite", "coal"]
	for parcel in city["parcels"]:
		if not str(parcel.get("building_id", "")).is_empty():
			continue
		var coord := _coord(parcel)
		if not deposit_coords.has(coord) or deposit_coords[coord]["item_id"] in ["limestone", "clay"]:
			wrong = coord
		elif deposit_coords[coord]["item_id"] in mine_items and (ore == CityMap.NO_TILE or not parcel["has_road"]):
			ore = coord
	_check(wrong != CityMap.NO_TILE and ore != CityMap.NO_TILE, "fresh world exposes suitable and unsuitable mine sites")
	if wrong == CityMap.NO_TILE or ore == CityMap.NO_TILE:
		quit(1)
		return

	# Resource amounts can change without map_revision/dirty_tiles.
	var item_id: String = deposit_coords[ore]["item_id"]
	deposit_coords[ore]["amount"] -= 1
	map.set_city_data(city, snapshot["buildings"], snapshot["shops"], game.api.firm_id)
	await process_frame
	await process_frame
	_check(draws.keys() == [ore], "one changed deposit redraws exactly one parcel")
	print("Deposit-only update: %d/%d parcels redrawn" % [draws.size(), map._tiles.size()])

	for index in game.build_option.item_count:
		if game.build_option.get_item_metadata(index) == "mine":
			game.build_option.select(index)
			break
	draws.clear()
	_click(map, wrong)
	await process_frame
	await process_frame
	_check(draws.keys() == [wrong], "selecting a parcel redraws just its highlight")
	_check(map.selected == wrong and _coord(game.selected_parcel) == wrong, "input selects the elevated parcel")
	await game._on_build_pressed()
	_check(game.status_label.text.contains("нужна залежь"), "server rejects a mine without a matching deposit: " + game.status_label.text)
	_check(map.building_at(wrong).is_empty(), "rejected action does not place a building")

	_click(map, ore)
	await process_frame
	_check(map.selected == ore and _coord(game.selected_parcel) == ore, "input moves selection to the deposit")
	_check(CatalogUI.selected_id(game.build_recipe) == "extract_" + item_id, "construction form chooses the matching recipe")
	await process_frame
	draws.clear()
	await game._on_build_pressed()
	await process_frame
	await process_frame
	var building := map.building_at(ore)
	_check(building.get("type") == "mine" and building.get("firm_id") == game.api.firm_id, "server accepted and placed the mine")
	_check(building.get("recipe_id") == "extract_" + item_id, "the server used the chosen deposit recipe")
	_check(map.parcel_at(ore)["has_road"] and map._tiles[ore].visual.road, "road state is visible from the server snapshot")
	_check(draws.has(ore) and draws.size() < map._tiles.size(), "construction redraws only changed parcels")
	print("Mine and road update (%s): %d/%d parcels redrawn" % [item_id, draws.size(), map._tiles.size()])
	await game._on_city_selected("ufa")
	_check(game.selected_parcel.is_empty() and map.building_at(ore).get("id") != building.get("id"), "no selection or building leaks between cities")
	print("Map smoke: %d failures" % failures)
	quit(1 if failures else 0)

func _record_draw(coord: Vector2i) -> void:
	draws[coord] = true

func _coord(entry: Dictionary) -> Vector2i:
	return Vector2i(int(entry["x"]), int(entry["y"]))

func _click(map: CityMap, coord: Vector2i) -> void:
	var event := InputEventMouseButton.new()
	event.button_index = MOUSE_BUTTON_LEFT
	event.pressed = true
	event.position = map.get_global_transform_with_canvas() * map._tiles[coord].position
	root.push_input(event, true)
	event = event.duplicate()
	event.pressed = false
	root.push_input(event, true)

func _check(ok: bool, message: String) -> void:
	if not ok:
		failures += 1
		printerr("FAIL: " + message)
