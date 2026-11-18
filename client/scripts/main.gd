extends Control

@onready var gfx: GraphicsSettings = $GraphicsSettings
@onready var api: ApiClient = $ApiClient
@onready var backdrop: TextureRect = $Backdrop
@onready var logo: TextureRect = $SafeArea/UI/TopBar/Logo
@onready var status_label: Label = $SafeArea/UI/TopBar/Status
@onready var tick_label: Label = $SafeArea/UI/TopBar/TickLabel
@onready var cash_label: Label = $SafeArea/UI/TopBar/CashLabel
@onready var login_panel: PanelContainer = $SafeArea/UI/LoginPanel
@onready var name_edit: LineEdit = $SafeArea/UI/LoginPanel/Margin/VBox/NameEdit
@onready var industry_option: OptionButton = $SafeArea/UI/LoginPanel/Margin/VBox/Industry
@onready var url_edit: LineEdit = $SafeArea/UI/LoginPanel/Margin/VBox/UrlEdit
@onready var city_host: Node2D = $GameLayer/CityHost
@onready var city_map: CityMap = $GameLayer/CityHost/CityMap
@onready var ambience: CityAmbience = $GameLayer/CityHost/Ambience
@onready var side: PanelContainer = $SafeArea/UI/SidePanel
@onready var side_title: Label = $SafeArea/UI/SidePanel/Margin/VBox/Title
@onready var side_body: RichTextLabel = $SafeArea/UI/SidePanel/Margin/VBox/Body
@onready var build_option: OptionButton = $SafeArea/UI/SidePanel/Margin/VBox/BuildType
@onready var log_box: RichTextLabel = $SafeArea/UI/BottomPanel/Margin/Log
@onready var rankings_box: ItemList = $SafeArea/UI/RankPanel/Margin/VBox/List
@onready var city_tabs: HBoxContainer = $SafeArea/UI/CityTabs
@onready var settings_panel: SettingsPanel = $SafeArea/UI/SettingsPanel
@onready var game_layer: Node2D = $GameLayer
@onready var bottom_panel: PanelContainer = $SafeArea/UI/BottomPanel
@onready var rank_panel: PanelContainer = $SafeArea/UI/RankPanel

var current_city: String = "ufa"
var selected_parcel: Dictionary = {}
var grade_material: ShaderMaterial

func _ready() -> void:
	DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)
	backdrop.texture = load("res://assets/bg/city_backdrop.jpg")
	logo.texture = load("res://assets/ui/logo_bashcapital.jpg")
	industry_option.clear()
	industry_option.add_item("Молочка (Уфа)", 0)
	industry_option.set_item_metadata(0, "dairy")
	industry_option.add_item("Мебель (Уфа)", 1)
	industry_option.set_item_metadata(1, "furniture")
	industry_option.add_item("Металл (Белорецк)", 2)
	industry_option.set_item_metadata(2, "steel")
	_fill_build_options()
	city_map.tile_clicked.connect(_on_tile_clicked)
	ambience.setup(gfx)
	gfx.quality_changed.connect(_on_quality_changed)
	_on_quality_changed(gfx.quality)
	settings_panel.quality_picked.connect(func(q): gfx.set_quality(q))
	settings_panel.fullscreen_toggled.connect(func(on): gfx.set_fullscreen(on))
	settings_panel.closed.connect(func(): settings_panel.visible = false)
	settings_panel.subscribe_pressed.connect(_on_subscribe)
	settings_panel.hire_deputy_pressed.connect(_on_hire_deputy)
	settings_panel.deputy_tasks_pressed.connect(_on_deputy_tasks)
	side.visible = false
	game_layer.visible = false
	bottom_panel.visible = false
	rank_panel.visible = false
	city_tabs.visible = false
	settings_panel.visible = false

func _fill_build_options() -> void:
	build_option.clear()
	var types := ["farm", "orchard", "dairy_plant", "glass_plant", "quarry", "textile_mill", "board_plant", "furniture_plant", "mine", "steel_mill", "chem_plant", "logging_camp", "warehouse", "garage"]
	for t in types:
		build_option.add_item(t)

func _on_quality_changed(level: int) -> void:
	city_map.detail_level = level
	city_map.queue_redraw()
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

func _on_connect_pressed() -> void:
	api.configure(url_edit.text.strip_edges())
	await api.fetch_catalog()
	var industry: String = str(industry_option.get_item_metadata(industry_option.selected))
	var firm := await api.register_firm(name_edit.text.strip_edges(), industry)
	if firm.has("error"):
		status_label.text = "Ошибка: %s" % firm["error"]
		return
	status_label.text = "Фирма: %s" % firm.get("name", "")
	login_panel.visible = false
	game_layer.visible = true
	bottom_panel.visible = true
	rank_panel.visible = true
	city_tabs.visible = true
	side.visible = true
	current_city = "beloretsk" if industry == "steel" else "ufa"
	await _refresh()

func _refresh() -> void:
	await api.fetch_world()
	_render_world()
	_render_rankings()

func _render_world() -> void:
	var w: Dictionary = api.world
	tick_label.text = "Тик %s · %s.%s" % [w.get("tick", 0), w.get("month", 1), w.get("year", 2026)]
	if api.firm_id != "" and w.has("firms"):
		var f: Dictionary = w["firms"].get(api.firm_id, {})
		var sub := " · PRO" if f.get("subscription_active", false) else ""
		cash_label.text = "Касса: %s ₽ · Фуры: %s%s" % [_money(f.get("cash", 0)), f.get("trucks", 0), sub]
	var cities: Dictionary = w.get("cities", {})
	if cities.has(current_city):
		city_map.set_city_data(cities[current_city], w.get("buildings", {}), w.get("shops", {}))
		side_title.text = _city_name(current_city)
	var logs: Array = w.get("last_tick_log", [])
	log_box.clear()
	for line in logs:
		log_box.append_text(str(line) + "\n")

func _render_rankings() -> void:
	rankings_box.clear()
	var rows: Array = await api.rankings()
	rows.sort_custom(func(a, b): return float(a.get("capital", 0)) > float(b.get("capital", 0)))
	for r in rows:
		var tag := "AI" if r.get("is_ai", false) else "Вы"
		rankings_box.add_item("%s · %s — %s ₽" % [tag, r.get("name", ""), _money(r.get("capital", 0))])

func _on_tile_clicked(x: int, y: int, parcel: Dictionary) -> void:
	selected_parcel = parcel
	selected_parcel["x"] = x
	selected_parcel["y"] = y
	var zone := str(parcel.get("zone", "?"))
	var slots := int(parcel.get("retail_slots", 0))
	side_body.clear()
	side_body.append_text("[b]Участок (%s, %s)[/b]\nЗона: %s\nДорога: %s\nСлоты аренды: %s\nТрафик: %.2f\n" % [
		x, y, zone, "да" if bool(parcel.get("has_road", false)) else "нет (подведём за ваш счёт)", slots, float(parcel.get("traffic", 0))
	])
	var city: Dictionary = api.world.get("cities", {}).get(current_city, {})
	for d in city.get("deposits", []):
		if int(d.get("x", 999)) == x and int(d.get("y", 999)) == y and int(d.get("amount", 0)) > 0:
			side_body.append_text("\n[color=#c2b280]Залежь: %s · %s/%s[/color]\n" % [d.get("item_id"), d.get("amount"), d.get("max_amount")])
			break

func _on_build_pressed() -> void:
	if selected_parcel.is_empty():
		return
	var typ := build_option.get_item_text(build_option.selected)
	var res := await api.action("build", {
		"city_id": current_city,
		"building_type": typ,
		"x": selected_parcel.get("x", 0),
		"y": selected_parcel.get("y", 0),
	})
	if res.has("error"):
		status_label.text = str(res["error"])
	_render_world()

func _on_rent_pressed() -> void:
	if selected_parcel.is_empty():
		return
	var item := "bottled_milk"
	if api.firm.get("industry", "") == "furniture":
		item = "furniture"
	elif api.firm.get("industry", "") == "steel":
		item = "steel_goods"
	var res := await api.action("rent_shop", {
		"city_id": current_city,
		"x": selected_parcel.get("x", 0),
		"y": selected_parcel.get("y", 0),
		"item_id": item,
		"price": 55,
	})
	if res.has("error"):
		status_label.text = str(res["error"])
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

func _on_city_ufa() -> void:
	current_city = "ufa"
	await _refresh()

func _on_city_beloretsk() -> void:
	current_city = "beloretsk"
	await _refresh()

func _city_name(id: String) -> String:
	for c in api.catalog.get("cities", []):
		if str(c.get("id")) == id:
			return str(c.get("name"))
	return id

func _money(v) -> String:
	return str(int(float(v)))
