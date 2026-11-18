extends SceneTree

var failed := false

func _initialize() -> void:
	_run.call_deferred()

func _run() -> void:
	var args := OS.get_cmdline_user_args()
	if args.size() != 1:
		quit(2)
		return
	var game = load("res://scenes/main.tscn").instantiate()
	root.add_child(game)
	game.url_edit.text = args[0]
	game.name_edit.text = "Хлебная цепочка smoke"
	game.industry_option.select(1)
	await game._on_connect_pressed()
	if not _check(not game.api.firm_id.is_empty(), "firm created"):
		return
	await game._on_city_selected("beloretsk")
	game._on_company_pressed()
	var first: Button = game.manage_panel.body.get_child(0)
	first.pressed.emit()
	_check(game.manage_panel.kind == "bread_guide", "company opens the bread guide")
	_check(game.manage_panel._snapshot.text.begins_with("0 из 4"), "new firm starts with an empty checklist")

	for i in BreadGuide.STEPS.size():
		var step: Dictionary = BreadGuide.STEPS[i]
		if not _check(not game.manage_panel._bread_buttons[i].disabled, "guide found a place for " + step["type"]):
			return
		game.manage_panel._bread_buttons[i].pressed.emit()
		while game.current_city != "ufa" or game.selected_parcel.is_empty() or game.manage_panel.kind == "bread_guide" and game.manage_panel.visible:
			await process_frame
		_check(game.current_city == "ufa", "shortcut opens Ufa")
		if step.has("recipe"):
			_check(CatalogUI.selected_id(game.build_option) == step["type"], "correct construction type")
			_check(CatalogUI.selected_id(game.build_recipe) == step["recipe"], "correct construction recipe")
			await game._on_build_pressed()
		else:
			_check(CatalogUI.selected_id(game.manage_panel._shop_type) == "grocery", "shop defaults to grocery")
			_check(CatalogUI.selected_id(game.manage_panel._shop_item) == "wheat_bread", "shop defaults to wheat bread")
			var confirm: Button = game.manage_panel.body.get_child(-1)
			confirm.pressed.emit()
			await game.api.world_updated
			await process_frame
		game.manage_panel.open_bread_guide(game.api.catalog, game.api.world, game.api.firm_id)
		if not _check(game.manage_panel._snapshot.text.begins_with("%d из 4" % (i + 1)), "checklist advances after the server action"):
			return

	var sold := false
	for _tick in 8:
		await game._on_tick_pressed()
		for line in game.api.world.get("last_tick_log", []):
			if "Розница Хлебная цепочка smoke" in line and "sold" in line and "wheat_bread" in line:
				sold = true
		if sold:
			break
	_check(sold, "the guided wheat chain led to a retail sale")
	print("Bread guide smoke: 4/4 configured, retail sale = %s" % sold)
	quit(1 if failed else 0)

func _check(ok: bool, label: String) -> bool:
	if not ok:
		failed = true
		printerr("FAIL: " + label)
		quit(1)
	return ok
