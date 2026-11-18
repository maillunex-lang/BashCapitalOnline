extends SceneTree

var failures := 0

func _initialize() -> void:
	_run.call_deferred()

func _run() -> void:
	var args := OS.get_cmdline_user_args()
	if args.size() != 1:
		printerr("Usage: godot --headless --path client --script res://tests/catalog_v2_smoke.gd -- http://127.0.0.1:8080")
		quit(2)
		return

	var game = load("res://scenes/main.tscn").instantiate()
	root.add_child(game)
	game.url_edit.text = args[0]
	game.name_edit.text = "Каталог V2 smoke"
	await game._on_connect_pressed()
	if game.api.firm_id.is_empty():
		_check(false, "connect and create a firm: " + game.status_label.text)
		quit(1)
		return

	var catalog: Dictionary = game.api.catalog
	_check(catalog["economy_catalog_version"] == 2, "expected server catalog V2")
	game._on_catalog_pressed()
	var panel: CatalogPanel = game.catalog_panel
	for section in CatalogPanel.KINDS.size():
		var kind: String = CatalogPanel.KINDS[section]
		var source_path := ProjectSettings.globalize_path("res://").path_join("../data/catalog_v2/" + kind + ".json")
		var source: Array = JSON.parse_string(FileAccess.get_file_as_string(source_path))
		var expected := {}
		for entry in source:
			expected[entry["id"]] = entry["name"]
		panel.section.select(section)
		panel.refresh()
		_check(panel.rows.item_count == source.size(), kind + ": displayed count")
		for index in panel.rows.item_count:
			var id: String = panel.rows.get_item_metadata(index)["id"]
			_check(panel.rows.get_item_text(index) == expected.get(id), kind + ": label and ID for " + id)
			expected.erase(id)
			var texture := panel.rows.get_item_icon(index)
			_check(texture != null and texture.get_width() > 0, kind + ": texture for " + id)
			panel._show_entry(index)
			_check(not panel.details.text.is_empty(), kind + ": details for " + id)
		_check(expected.is_empty(), kind + ": every catalog ID must be discoverable")
		print("%s: %s rows" % [kind, panel.rows.item_count])

	panel.section.select(0)
	panel.search.text = "wheat_bread"
	panel.refresh()
	_check(panel.rows.item_count == 1, "item search by V2 ID")
	panel.search.text = "ПШЕНИЧ"
	panel.refresh()
	_check(panel.rows.item_count > 0, "case-insensitive Russian item search")
	panel.hide()

	var build: OptionButton = game.build_option
	_check(build.item_count == catalog["buildings"].size(), "every building in the construction menu")
	for index in build.item_count:
		var entry: Dictionary = catalog["buildings"][index]
		_check(build.get_item_text(index) == entry["name"] and build.get_item_metadata(index) == entry["id"], "construction option " + entry["id"])

	var manage: ManagePanel = game.manage_panel
	var recipes_by_building := {}
	for recipe in catalog["recipes"]:
		var id: String = recipe["building_type"]
		if not recipes_by_building.has(id):
			recipes_by_building[id] = []
		recipes_by_building[id].append(recipe["id"])
	for building in catalog["buildings"]:
		var id: String = building["id"]
		manage.open_building({"type": id}, catalog, {}, "")
		_check(_ids(manage._recipe) == recipes_by_building.get(id, []), "recipe choices for " + id)

	var city: Dictionary = game.api.world["cities"][game.current_city]
	var parcel: Dictionary = {}
	for candidate in city["parcels"]:
		if candidate["retail_slots"] > 0 and candidate.get("rented_by", "") == "" and candidate.get("has_road", false):
			parcel = candidate
			break
	_check(not parcel.is_empty(), "fresh world has an available roadside retail slot")
	if not parcel.is_empty():
		game._on_tile_clicked(int(parcel["x"]), int(parcel["y"]), parcel)
		game._on_rent_pressed()
		_check(manage._shop_type.item_count == 14, "every shop type in the rent menu")
		for shop in catalog["shops"]:
			manage.open_shop({"shop_type": shop["id"]}, catalog)
			var actual := _ids(manage._shop_item)
			var allowed: Array = shop["allowed_items"].duplicate()
			actual.sort()
			allowed.sort()
			_check(actual == allowed, "assortment for " + shop["id"])

		game._on_rent_pressed()
		var item_index := _ids(manage._shop_item).find("wheat_bread")
		_check(item_index >= 0, "V2 bread available in grocery store")
		if item_index >= 0:
			manage._shop_item.select(item_index)
			var confirm: Button = manage.body.get_child(-1)
			confirm.pressed.emit()
			await game.api.world_updated
			var rented := false
			for shop in game.api.world["shops"].values():
				if shop["firm_id"] == game.api.firm_id and shop["shop_type"] == "grocery" and shop["item_id"] == "wheat_bread":
					rented = true
					manage.open_shop(shop, catalog)
					_check(manage._inventory_rows.item_count == 1 and manage._inventory_rows.get_item_icon(0) != null, "leased shelf shows a product icon")
					_check(manage._inventory_rows.get_item_text(0).ends_with("×0"), "leased shelf shows actual stock")
			_check(rented, "game rented the V2 grocery slot through the server")
			await process_frame
			await game.api.world_updated
			await create_timer(0.1).timeout
	print("V2 catalog UI and rental: %s failures" % failures)
	quit(0 if failures == 0 else 1)

func _ids(button: OptionButton) -> Array:
	var ids := []
	for index in button.item_count:
		ids.append(button.get_item_metadata(index))
	return ids

func _check(ok: bool, message: String) -> void:
	if not ok:
		failures += 1
		printerr("FAIL: " + message)
