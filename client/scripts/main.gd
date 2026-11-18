extends Control

@onready var gfx: GraphicsSettings = $GraphicsSettings
@onready var api: ApiClient = $ApiClient
@onready var backdrop: TextureRect = $Backdrop
@onready var logo: TextureRect = $SafeArea/UI/TopBar/Logo
@onready var company_btn: Button = $SafeArea/UI/TopBar/CompanyBtn
@onready var status_label: Label = $SafeArea/UI/TopBar/Status
@onready var tick_label: Label = $SafeArea/UI/TopBar/TickLabel
@onready var cash_label: Label = $SafeArea/UI/TopBar/CashLabel
@onready var login_panel: PanelContainer = $SafeArea/UI/LoginPanel
@onready var username: LineEdit = $SafeArea/UI/LoginPanel/Margin/VBox/Username
@onready var password: LineEdit = $SafeArea/UI/LoginPanel/Margin/VBox/Password
@onready var name_edit: LineEdit = $SafeArea/UI/LoginPanel/Margin/VBox/NameEdit
@onready var industry_option: OptionButton = $SafeArea/UI/LoginPanel/Margin/VBox/Industry
const CONNECTION_PREFS := "user://connection.cfg"

@onready var connection_message: Label = $SafeArea/UI/LoginPanel/Margin/VBox/ConnectionMessage
@onready var connect_btn: Button = $SafeArea/UI/LoginPanel/Margin/VBox/ConnectBtn
@onready var register_btn: Button = $SafeArea/UI/LoginPanel/Margin/VBox/RegisterBtn
@onready var url_edit: LineEdit = $SafeArea/UI/LoginPanel/Margin/VBox/UrlEdit
@onready var city_host: Node2D = $GameLayer/CityHost
@onready var city_map: CityMap = $GameLayer/CityHost/CityMap
@onready var side: PanelContainer = $SafeArea/UI/SidePanel
@onready var side_title: Label = $SafeArea/UI/SidePanel/Margin/VBox/Title
@onready var side_body: RichTextLabel = $SafeArea/UI/SidePanel/Margin/VBox/Body
@onready var build_option: OptionButton = $SafeArea/UI/SidePanel/Margin/VBox/BuildType
@onready var build_recipe: OptionButton = $SafeArea/UI/SidePanel/Margin/VBox/BuildRecipe
@onready var log_box: RichTextLabel = $SafeArea/UI/BottomPanel/Margin/Log
@onready var rankings_box: ItemList = $SafeArea/UI/RankPanel/Margin/VBox/List
@onready var city_tabs: HBoxContainer = $SafeArea/UI/CityTabs
@onready var settings_panel: SettingsPanel = $SafeArea/UI/SettingsPanel
@onready var manage_panel: ManagePanel = $SafeArea/UI/ManagePanel
@onready var game_layer: Node2D = $GameLayer
@onready var bottom_panel: PanelContainer = $SafeArea/UI/BottomPanel
@onready var rank_panel: PanelContainer = $SafeArea/UI/RankPanel

var connecting := false
var polling_world := false
var current_city: String = "ufa"
var selected_parcel: Dictionary = {}
var grade_material: ShaderMaterial
var remote_assets: RemoteAssets
var catalog_panel := CatalogPanel.new()
var season_names: PackedStringArray = ["", "Ҡыш", "Ҡыш", "Яҙ", "Яҙ", "Яҙ", "Йәй", "Йәй", "Йәй", "Көҙ", "Көҙ", "Көҙ", "Ҡыш"]

func _ready() -> void:
	DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)
	if ResourceLoader.exists("res://assets/bg/city_backdrop.jpg"):
		backdrop.texture = load("res://assets/bg/city_backdrop.jpg")
	if ResourceLoader.exists("res://assets/ui/logo_bashcapital.png"):
		logo.texture = load("res://assets/ui/logo_bashcapital.png")
	elif ResourceLoader.exists("res://assets/ui/logo_bashcapital.jpg"):
		logo.texture = load("res://assets/ui/logo_bashcapital.jpg")
	remote_assets = RemoteAssets.new()
	add_child(remote_assets)
	remote_assets.setup(api)
	remote_assets.progress.connect(_show_connection_message)
	_restore_connection_preferences()
	remote_assets.packs_ready.connect(CatalogUI.clear_texture_cache)
	remote_assets.update_available.connect(_on_apk_update)
	industry_option.clear()
	var industries := [
		["Молочка (Уфа)", "dairy"],
		["Еда / пекарня (Уфа)", "food"],
		["Мебель (Уфа)", "furniture"],
		["Одежда (Уфа)", "apparel"],
		["Металл (Белорецк)", "steel"],
		["Нефть / химия (Стерлитамак)", "oil"],
	]
	for i in industries.size():
		industry_option.add_item(str(industries[i][0]), i)
		industry_option.set_item_metadata(i, industries[i][1])
	add_child(catalog_panel)
	build_option.item_selected.connect(func(_index): _fill_build_recipes())
	build_recipe.get_popup().add_theme_constant_override("icon_max_width", 28)
	_fill_build_options()
	city_map.tile_clicked.connect(_on_tile_clicked)
	get_viewport().size_changed.connect(_fit_city_map)
	gfx.quality_changed.connect(_on_quality_changed)
	_on_quality_changed(gfx.quality)
	settings_panel.quality_picked.connect(func(q): gfx.set_quality(q))
	settings_panel.fullscreen_toggled.connect(func(on): gfx.set_fullscreen(on))
	settings_panel.closed.connect(func(): settings_panel.visible = false)
	settings_panel.subscribe_pressed.connect(_on_subscribe)
	settings_panel.hire_deputy_pressed.connect(_on_hire_deputy)
	settings_panel.deputy_tasks_pressed.connect(_on_deputy_tasks)
	manage_panel.save_building.connect(_on_save_building)
	manage_panel.save_shop.connect(_on_save_shop)
	manage_panel.rent_shop.connect(_on_rent_shop)
	manage_panel.stock_shop.connect(_on_stock_shop)
	manage_panel.transfer_stock.connect(_on_transfer_stock)
	manage_panel.asset_chosen.connect(_on_asset_chosen)
	manage_panel.bread_parcel_requested.connect(_on_bread_parcel_requested)
	var refresh_timer := Timer.new()
	refresh_timer.wait_time = 8.0
	refresh_timer.autostart = true
	refresh_timer.timeout.connect(_poll_world)
	add_child(refresh_timer)
	manage_panel.closed.connect(func(): manage_panel.visible = false)
	side.visible = false
	game_layer.visible = false
	bottom_panel.visible = false
	rank_panel.visible = false
	city_tabs.visible = false
	settings_panel.visible = false
	manage_panel.visible = false

func _fill_build_options() -> void:
	build_option.clear()
	for entry in api.catalog.get("buildings", []):
		CatalogUI.add_option(build_option, entry, "buildings")
	_fill_build_recipes()

func _fill_build_recipes() -> void:
	build_recipe.clear()
	var building := CatalogUI.selected_id(build_option)
	var deposit := _parcel_deposit(selected_parcel)
	for recipe in api.catalog.get("recipes", []):
		if recipe["building_type"] != building:
			continue
		var item_id: String = recipe["output"]["item_id"]
		var output := CatalogUI.find(api.catalog, "items", item_id)
		var index := build_recipe.item_count
		build_recipe.add_icon_item(CatalogUI.icon(output, "items"), recipe["name"])
		build_recipe.set_item_metadata(index, recipe["id"])
		if item_id == deposit.get("item_id") and not str(recipe.get("extract_mode", "")).is_empty():
			build_recipe.select(index)
	build_recipe.visible = build_recipe.item_count > 1

func _on_catalog_pressed() -> void:
	catalog_panel.open(api.catalog)

func _on_quality_changed(level: int) -> void:
	city_map.detail_level = level
	# Ultra color grade on backdrop
	if level >= GraphicsSettings.QUALITY_ULTRA:
		if grade_material == null:
			var sh: Shader = load("res://shaders/ultra_grade.gdshader")
			grade_material = ShaderMaterial.new()
			grade_material.shader = sh
		backdrop.material = grade_material
		backdrop.modulate = Color(0.85, 0.9, 0.88, 1)
	else:
		backdrop.material = null
		backdrop.modulate = Color(0.75, 0.82, 0.8, 1)
	status_label.text = "Графика: %s" % gfx.quality_name()

func _on_settings_pressed() -> void:
	var firm: Dictionary = {}
	if api.firm_id != "" and api.world.has("firms"):
		firm = api.world["firms"].get(api.firm_id, {})
	var deputy: Dictionary = firm.get("deputy", {}) if firm.get("deputy", null) != null else {}
	settings_panel.bind_state(
		gfx.quality,
		gfx.fullscreen,
		bool(firm.get("subscription_active", false)),
		int(firm.get("subscription_until_tick", 0)),
		deputy
	)
	settings_panel.visible = true

func _on_subscribe(months: int) -> void:
	var res := await api.action("subscribe", {"months": months})
	if res.has("error"):
		status_label.text = str(res["error"])
	else:
		status_label.text = "Подписка оформлена"
	await _refresh()
	_on_settings_pressed()

func _on_hire_deputy(deputy_name: String) -> void:
	var res := await api.action("hire_deputy", {"name": deputy_name})
	if res.has("error"):
		status_label.text = str(res["error"])
	await _refresh()
	_on_settings_pressed()

func _on_deputy_tasks(kinds: PackedStringArray) -> void:
	var arr: Array = []
	for k in kinds:
		arr.append(k)
	var res := await api.action("set_deputy_tasks", {"tasks": arr})
	if res.has("error"):
		status_label.text = str(res["error"])
	else:
		status_label.text = "Задания выданы заместителю"
	await _refresh()

func _restore_connection_preferences() -> void:
	var prefs := ConfigFile.new()
	if prefs.load(CONNECTION_PREFS) == OK:
		url_edit.text = str(prefs.get_value("connection", "server", url_edit.text))
		username.text = str(prefs.get_value("connection", "username", ""))

func _save_connection_preferences(login: String) -> void:
	var prefs := ConfigFile.new()
	prefs.set_value("connection", "server", api.base_url)
	prefs.set_value("connection", "username", login)
	var error := prefs.save(CONNECTION_PREFS)
	if error != OK:
		push_warning("Could not save connection preferences: %s" % error_string(error))

func _show_connection_message(message: String) -> void:
	if login_panel.visible:
		connection_message.text = message
		connection_message.show()
	else:
		status_label.text = message

func _on_connect_pressed(register_account := false) -> void:
	if connecting:
		return
	connecting = true
	connect_btn.disabled = true
	register_btn.disabled = true
	var fields := [url_edit, username, password, name_edit]
	for field in fields:
		field.editable = false
	industry_option.disabled = true
	await _connect_to_world(register_account)
	for field in fields:
		field.editable = true
	industry_option.disabled = false
	connect_btn.disabled = false
	register_btn.disabled = false
	connecting = false

func _connect_to_world(register_account: bool) -> void:
	var address := url_edit.text.strip_edges()
	if address.is_empty():
		_show_connection_message("Введите адрес игрового сервера.")
		return
	if not "://" in address:
		address = "https://" + address
	var host := address.get_slice("://", 1).get_slice("/", 0)
	if not (address.begins_with("http://") or address.begins_with("https://")) or host.is_empty() or "@" in address or "?" in address or "#" in address:
		_show_connection_message("Укажите адрес игрового сервера с https:// или http://.")
		return
	api.configure(address)
	_show_connection_message("Подключение к серверу…")
	if not await api.fetch_catalog():
		_show_connection_message("Сервер не отдал каталог. Проверьте адрес и подключение к сети.")
		return
	city_map.bind_catalog(api.catalog)
	await remote_assets.check_and_fetch()
	_fill_build_options()
	var industry: String = str(industry_option.get_item_metadata(industry_option.selected))
	var login := username.text.strip_edges()
	if register_account or not login.is_empty() or not password.text.is_empty():
		var auth: Dictionary
		if register_account:
			auth = await api.auth_register(login, password.text)
		else:
			auth = await api.auth_login(login, password.text)
		if not auth.has("session_token"):
			_show_connection_message(str(auth.get("error", "Не удалось войти в аккаунт")).strip_edges().left(240))
			return
	if api.firm_id.is_empty():
		var firm := await api.register_firm(name_edit.text.strip_edges(), industry)
		if not firm.has("id"):
			_show_connection_message("Не удалось открыть фирму: %s" % str(firm.get("error", "сервер недоступен")).strip_edges().left(200))
			return
	else:
		await api.fetch_world()
		var firm: Dictionary = api.world.get("firms", {}).get(api.firm_id, {})
		if firm.is_empty():
			_show_connection_message("Фирма аккаунта недоступна. Попробуйте подключиться снова.")
			return
		industry = str(firm.get("industry", industry))
		api.firm = firm
	if api.world.is_empty():
		await api.fetch_world()
	if api.world.is_empty():
		_show_connection_message("Сервер не отдал состояние мира. Попробуйте подключиться снова.")
		return
	_save_connection_preferences(login)
	password.clear()
	status_label.text = "Фирма: %s" % api.firm.get("name", "")
	login_panel.visible = false
	company_btn.show()
	game_layer.visible = true
	bottom_panel.visible = true
	rank_panel.visible = false
	city_tabs.visible = true
	side.visible = true
	if industry == "steel":
		current_city = "beloretsk"
	elif industry == "oil":
		current_city = "sterlitamak" if _city_enabled("sterlitamak") else "beloretsk"
	else:
		current_city = "ufa"
	_rebuild_city_tabs()
	await _refresh()

func _refresh() -> void:
	await api.fetch_world()
	_render_world()
	await _render_rankings()

func _render_world() -> void:
	var w: Dictionary = api.world
	var month: int = int(w.get("month", 1))
	var season: String = season_names[month] if month >= 1 and month <= 12 else ""
	tick_label.text = "Тик %s · %s.%s · %s" % [w.get("tick", 0), month, w.get("year", 2026), season]
	if api.firm_id != "" and w.has("firms"):
		var f: Dictionary = w["firms"].get(api.firm_id, {})
		var sub := " · PRO" if f.get("subscription_active", false) else ""
		cash_label.text = "Касса: %s ₽ · Фуры: %s%s" % [_money(f.get("cash", 0)), f.get("trucks", 0), sub]
	var cities: Dictionary = w.get("cities", {})
	if cities.has(current_city):
		city_map.set_city_data(cities[current_city], w.get("buildings", {}), w.get("shops", {}), api.firm_id)
		_fit_city_map()
		if not selected_parcel.is_empty():
			selected_parcel = city_map.parcel_at(city_map.selected)
			_render_parcel_details(selected_parcel)
		side_title.text = _city_name(current_city)
	var logs = w.get("last_tick_log")
	log_box.clear()
	if logs is Array:
		for line in logs:
			log_box.append_text(str(line) + "\n")
	if manage_panel.visible:
		manage_panel.refresh_snapshot(api.catalog, w, api.firm_id)

func _fit_city_map() -> void:
	var left := 20.0
	if rank_panel.visible:
		left = rank_panel.get_global_rect().end.x + 20.0
	var right := side.get_global_rect().position.x - 16.0
	var top := city_tabs.get_global_rect().end.y + 10.0
	var bottom := bottom_panel.get_global_rect().position.y - 8.0
	city_map.frame_map(Rect2(Vector2(left, top), Vector2(right - left, bottom - top)))

func _on_rank_toggle() -> void:
	rank_panel.visible = not rank_panel.visible
	_fit_city_map()

func _render_rankings() -> void:
	var rows: Array = await api.rankings()
	rankings_box.clear()
	rows.sort_custom(func(a, b): return float(a.get("capital", 0)) > float(b.get("capital", 0)))
	for r in rows:
		var tag := "AI" if r.get("is_ai", false) else "Вы"
		rankings_box.add_item("%s · %s — %s ₽" % [tag, r.get("name", ""), _money(r.get("capital", 0))])

func _on_tile_clicked(x: int, y: int, parcel: Dictionary) -> void:
	selected_parcel = parcel
	_render_parcel_details(parcel)
	_fill_build_recipes()
	_open_parcel_management(x, y)

func _render_parcel_details(parcel: Dictionary) -> void:
	if parcel.is_empty():
		side_body.text = "Выберите участок на карте"
		return
	var x := int(parcel["x"])
	var y := int(parcel["y"])
	var zone := str(parcel.get("zone", "?"))
	var slots := int(parcel.get("retail_slots", 0))
	side_body.clear()
	side_body.append_text("[b]Участок (%s, %s)[/b]\nЗона: %s\nДорога: %s\nСлоты аренды: %s\nТрафик: %.2f\nВысота: %.2f\n" % [
		x, y, zone, "да" if bool(parcel.get("has_road", false)) else "нет (подведём за ваш счёт)", slots, float(parcel.get("traffic", 0)), float(parcel.get("elevation", 0))
	])
	var deposit := _parcel_deposit(parcel)
	if not deposit.is_empty():
		side_body.append_text("\n[color=#c2b280]Залежь: %s · %s/%s[/color]\n" % [
			CatalogUI.entry_name(api.catalog, "items", str(deposit["item_id"])),
			deposit["amount"], deposit["max_amount"],
		])

	var coord := Vector2i(x, y)
	var building := city_map.building_at(coord)
	if not building.is_empty():
		_append_business(CatalogUI.entry_name(api.catalog, "buildings", building["type"]), building["firm_id"])
	var shop := city_map.shop_at(coord)
	if not shop.is_empty():
		_append_business(CatalogUI.entry_name(api.catalog, "shops", shop["shop_type"]), shop["firm_id"])

func _append_business(name: String, firm_id: String) -> void:
	var firm: Dictionary = api.world.get("firms", {}).get(firm_id, {})
	side_body.add_text("\n%s · %s" % [name, firm.get("name", firm_id)])

func _parcel_deposit(parcel: Dictionary) -> Dictionary:
	var city: Dictionary = api.world.get("cities", {}).get(current_city, {})
	var deposits = city.get("deposits")
	if parcel.is_empty() or not deposits is Array:
		return {}
	for deposit in deposits:
		if deposit["x"] == parcel["x"] and deposit["y"] == parcel["y"] and deposit["amount"] > 0:
			return deposit
	return {}

func _open_parcel_management(x: int, y: int) -> void:
	var coord := Vector2i(x, y)
	var building := city_map.building_at(coord)
	var shop := city_map.shop_at(coord)
	if building.get("firm_id", "") == api.firm_id and not building.is_empty():
		manage_panel.open_building(building, api.catalog, api.world, api.firm_id)
	elif shop.get("firm_id", "") == api.firm_id and not shop.is_empty():
		manage_panel.open_shop(shop, api.catalog)
	else:
		manage_panel.hide()

func _on_build_pressed() -> void:
	if selected_parcel.is_empty():
		return
	var typ := CatalogUI.selected_id(build_option)
	if typ.is_empty():
		return
	var res := await api.action("build", {
		"city_id": current_city,
		"building_type": typ,
		"recipe_id": CatalogUI.selected_id(build_recipe),
		"x": selected_parcel.get("x", 0),
		"y": selected_parcel.get("y", 0),
	})
	status_label.text = str(res.get("error", "Здание построено"))
	_render_world()

func _on_rent_pressed() -> void:
	if selected_parcel.is_empty():
		return
	manage_panel.open_rental(api.catalog, {
		"city_id": current_city,
		"x": selected_parcel["x"],
		"y": selected_parcel["y"],
	})

func _on_rent_shop(payload: Dictionary) -> void:
	var res := await api.action("rent_shop", payload)
	status_label.text = str(res.get("error", "Магазин арендован"))
	await _refresh()

func _on_buy_parcel_pressed() -> void:
	if selected_parcel.is_empty():
		return
	await api.action("buy_parcel", {"city_id": current_city, "x": selected_parcel.get("x", 0), "y": selected_parcel.get("y", 0)})
	await _refresh()

func _on_develop_pressed() -> void:
	if selected_parcel.is_empty():
		return
	await api.action("develop_housing", {"city_id": current_city, "x": selected_parcel.get("x", 0), "y": selected_parcel.get("y", 0)})
	await _refresh()

func _on_truck_pressed() -> void:
	await api.action("buy_truck", {})
	await _refresh()

func _on_tick_pressed() -> void:
	await api.manual_tick()
	await _refresh()

func _city_enabled(id: String) -> bool:
	for c in api.catalog.get("cities", []):
		if str(c.get("id")) == id:
			return bool(c.get("enabled", true))
	return false

func _rebuild_city_tabs() -> void:
	if city_tabs == null:
		return
	for child in city_tabs.get_children():
		if not child.has_meta("dynamic_city_tab"):
			continue
		city_tabs.remove_child(child)
		child.queue_free()
	var tick_btn: Node = city_tabs.get_node_or_null("TickBtn")
	for c in api.catalog.get("cities", []):
		if not bool(c.get("enabled", false)):
			continue
		var btn := Button.new()
		btn.set_meta("dynamic_city_tab", true)
		btn.text = str(c.get("name", c.get("id")))
		btn.custom_minimum_size.y = 52
		var cid := str(c.get("id"))
		btn.pressed.connect(_on_city_selected.bind(cid))
		if tick_btn:
			city_tabs.add_child(btn)
			city_tabs.move_child(btn, tick_btn.get_index())
		else:
			city_tabs.add_child(btn)

func _on_city_selected(cid: String) -> void:
	if current_city != cid:
		selected_parcel = {}
		manage_panel.hide()
		side_body.text = "Выберите участок на карте"
	current_city = cid
	_fill_build_recipes()
	_render_world()
	await _refresh()

func _city_name(id: String) -> String:
	for c in api.catalog.get("cities", []):
		if str(c.get("id")) == id:
			return str(c.get("name"))
	return id

func _money(v) -> String:
	return str(int(float(v)))

func _on_bread_parcel_requested(step: Dictionary) -> void:
	if current_city != BreadGuide.CITY:
		await _on_city_selected(BreadGuide.CITY)
	var parcel := BreadGuide.suggest_parcel(step, api.world, api.firm_id)
	if parcel.is_empty():
		status_label.text = "Подходящего свободного участка нет"
		return
	city_map.focus_parcel(Vector2i(int(parcel["x"]), int(parcel["y"])))
	if step.has("item"):
		var corridor: Dictionary = api.catalog.get("economy", {}).get("price_corridors", {}).get(step["item"], {})
		manage_panel.open_rental(api.catalog, {
			"city_id": current_city, "x": parcel["x"], "y": parcel["y"],
		}, {"shop_type": step["type"], "item_id": step["item"], "price": floorf(float(corridor.get("max", 55)))})
	else:
		_select_catalog_option(build_option, step["type"])
		_fill_build_recipes()
		_select_catalog_option(build_recipe, step["recipe"])
		status_label.text = "Участок выбран. Нажмите «Построить»."

func _select_catalog_option(option: OptionButton, id: String) -> void:
	for index in option.item_count:
		if option.get_item_metadata(index) == id:
			option.select(index)
			return

func _on_company_pressed() -> void:
	manage_panel.open_company(api.catalog, api.world, api.firm_id)

func _on_asset_chosen(asset_kind: String, asset_id: String) -> void:
	var asset: Dictionary = api.world.get(asset_kind, {}).get(asset_id, {})
	if asset.is_empty() or asset["firm_id"] != api.firm_id:
		return
	if asset_kind == "buildings":
		manage_panel.open_building(asset, api.catalog, api.world, api.firm_id)
	elif asset_kind == "shops":
		manage_panel.open_shop(asset, api.catalog)

func _poll_world() -> void:
	if login_panel.visible or polling_world:
		return
	polling_world = true
	await api.fetch_world()
	polling_world = false
	_render_world()

func _on_save_building(payload: Dictionary) -> void:
	var res := await api.action("configure_building", payload)
	if res.has("error"):
		status_label.text = str(res["error"])
	else:
		status_label.text = "Завод настроен"
	await _refresh()

func _on_save_shop(payload: Dictionary) -> void:
	var res := await api.action("configure_shop", payload)
	if res.has("error"):
		status_label.text = str(res["error"])
	else:
		status_label.text = "Магазин настроен"
	await _refresh()

func _on_stock_shop(shop_id: String, qty: int) -> void:
	var res := await api.action("stock_shop", {"shop_id": shop_id, "qty": qty})
	if res.has("error"):
		status_label.text = str(res["error"])
	await _refresh()

func _on_apk_update(info: Dictionary) -> void:
	status_label.text = "Доступно обновление APK %s" % str(info.get("version_name", ""))
	# Android: open confirmation_url / apk.url via OS.shell_open when exported.
	var url := str(info.get("url", ""))
	if url.begins_with("http"):
		OS.shell_open(url)
	elif url != "" and api != null:
		OS.shell_open(api.base_url + url)


func _on_transfer_stock(payload: Dictionary) -> void:
	var res := await api.action("transfer_stock", payload)
	if res.has("error"):
		status_label.text = str(res["error"])
	else:
		status_label.text = "Груз отправлен · %d тик." % int(res.get("ticks_left", 0))
	_render_world()

func _on_zoom_in() -> void:
	city_map.zoom_center(1.3)

func _on_zoom_out() -> void:
	city_map.zoom_center(1.0 / 1.3)

func _on_zoom_fit() -> void:
	city_map.reset_camera()
