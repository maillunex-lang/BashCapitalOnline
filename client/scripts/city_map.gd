extends Node2D
class_name CityMap

signal tile_clicked(x: int, y: int, parcel: Dictionary)

const TILE_W := 64.0
const TILE_H := 32.0

var parcels: Array = []
var buildings: Dictionary = {}
var shops: Dictionary = {}
var deposits: Array = []
var selected: Vector2i = Vector2i(9999, 9999)
var hover: Vector2i = Vector2i(9999, 9999)
var detail_level: int = 2
var _sprites: Dictionary = {} # path -> Texture2D
var map_revision: int = -1
var dirty_only: Array = [] # optional dirty tiles [{x,y}] for incremental redraw

var zone_colors := {
	"center": Color("2F6F6A"),
	"residential": Color("4A7C59"),
	"commercial": Color("C4A35A"),
	"industrial": Color("6B7280"),
	"rural": Color("7FAE6E"),
	"empty": Color("3A4A46"),
}

var building_sprite := {
	"farm": "bld_farm",
	"livestock_farm": "bld_livestock",
	"orchard": "bld_orchard",
	"logging_camp": "bld_logging",
	"quarry": "bld_quarry",
	"mine": "bld_mine",
	"oil_well": "bld_oil",
	"gas_well": "bld_gas",
	"chem_plant": "bld_chem",
	"mill": "bld_mill",
	"bakery": "bld_bakery",
	"food_plant": "bld_food",
	"dairy_plant": "bld_dairy",
	"textile_mill": "bld_textile",
	"apparel_factory": "bld_apparel",
	"board_plant": "bld_board",
	"sawmill": "bld_sawmill",
	"furniture_plant": "bld_furniture",
	"glass_plant": "bld_glass",
	"paper_mill": "bld_paper",
	"plastics_plant": "bld_plastics",
	"refinery": "bld_refinery",
	"steel_mill": "bld_steel",
	"electronics_plant": "bld_electronics",
	"appliance_plant": "bld_appliance",
	"warehouse": "bld_warehouse",
	"garage": "bld_garage",
}

var zone_tile := {
	"center": "tile_center",
	"residential": "tile_residential",
	"industrial": "tile_industrial",
	"rural": "tile_rural",
	"empty": "tile_grass",
	"commercial": "tile_center",
}

func _tex(name: String) -> Texture2D:
	if name == "":
		return null
	if _sprites.has(name):
		return _sprites[name]
	for path in [
		"res://assets/map/%s.png" % name,
		"res://assets/icons/%s.png" % name,
		"res://assets/ui/%s.png" % name,
	]:
		if ResourceLoader.exists(path):
			var t: Texture2D = load(path)
			_sprites[name] = t
			return t
	_sprites[name] = null
	return null

func set_city_data(city: Dictionary, world_buildings: Dictionary, world_shops: Dictionary, revision: int = -1, dirty: Array = []) -> void:
	parcels = city.get("parcels", [])
	deposits = city.get("deposits", [])
	buildings = {}
	for id in world_buildings.keys():
		var b: Dictionary = world_buildings[id]
		if str(b.get("city_id", "")) == str(city.get("id", "")):
			buildings["%s,%s" % [b.get("x"), b.get("y")]] = b
	shops = {}
	for id in world_shops.keys():
		var s: Dictionary = world_shops[id]
		if str(s.get("city_id", "")) == str(city.get("id", "")):
			shops["%s,%s" % [s.get("x"), s.get("y")]] = s
	# RM-MAP-DIRTY: if only a few tiles changed, still full redraw for now but track revision.
	map_revision = revision
	dirty_only = dirty
	queue_redraw()

func _draw() -> void:
	draw_circle(Vector2.ZERO, 520, Color(0.18, 0.28, 0.24, 0.35))
	var sorted := parcels.duplicate()
	sorted.sort_custom(func(a, b): return (a.x + a.y) < (b.x + b.y))
	for p in sorted:
		var x: int = int(p.get("x", 0))
		var y: int = int(p.get("y", 0))
		var zone: String = str(p.get("zone", "empty"))
		var col: Color = zone_colors.get(zone, zone_colors["empty"])
		if Vector2i(x, y) == selected:
			col = col.lightened(0.25)
		elif Vector2i(x, y) == hover:
			col = col.lightened(0.12)
		var center := _iso(x, y)
		var tile_tex := _tex(str(zone_tile.get(zone, "tile_grass")))
		if tile_tex:
			var ts := tile_tex.get_size()
			var tint := Color(1.18, 1.18, 1.12, 1.0) if Vector2i(x, y) == selected else Color(1, 1, 1, 0.96)
			draw_texture_rect(tile_tex, Rect2(center - Vector2(ts.x * 0.5, ts.y * 0.62), ts), false, tint)
		else:
			_draw_tile(center, col)
		if bool(p.get("has_road", false)):
			var road := _tex("tile_road")
			if road:
				var rs := road.get_size()
				draw_texture_rect(road, Rect2(center - Vector2(rs.x * 0.5, rs.y * 0.55), rs * 0.55), false, Color(1, 1, 1, 0.9))
			else:
				_draw_road(center)
		for d in deposits:
			if int(d.get("x", 999)) == x and int(d.get("y", 999)) == y and int(d.get("amount", 0)) > 0:
				_draw_deposit(center, str(d.get("item_id", "")), float(d.get("amount", 0)) / maxf(1.0, float(d.get("max_amount", 1))))
		var key := "%s,%s" % [x, y]
		if buildings.has(key):
			var b: Dictionary = buildings[key]
			_draw_building_sprite(center, str(b.get("type", "")), zone)
		elif int(p.get("retail_slots", 0)) > 0:
			var shop_tex := _tex("bld_shop")
			if shop_tex:
				var ss := shop_tex.get_size()
				draw_texture_rect(shop_tex, Rect2(center - Vector2(ss.x * 0.5, ss.y * 0.72), ss * 0.72), false)
			else:
				_draw_shop_marker(center, shops.has(key))
			if shops.has(key) and Vector2i(x, y) == selected:
				var r: int = int(shops[key].get("radius", 2))
				_draw_catchment(x, y, r)

func _iso(x: int, y: int) -> Vector2:
	var elev := 0.0
	for p in parcels:
		if int(p.get("x", 999)) == x and int(p.get("y", 999)) == y:
			elev = float(p.get("elevation", 0))
			break
	# map_3d lite: elevation shifts iso Y (Urals relief)
	return Vector2((x - y) * TILE_W * 0.5, (x + y) * TILE_H * 0.5 - elev * 10.0)

func _draw_tile(center: Vector2, col: Color) -> void:
	var pts := PackedVector2Array([
		center + Vector2(0, -TILE_H * 0.5),
		center + Vector2(TILE_W * 0.5, 0),
		center + Vector2(0, TILE_H * 0.5),
		center + Vector2(-TILE_W * 0.5, 0),
	])
	if detail_level >= 1:
		var shadow := PackedVector2Array()
		for p in pts:
			shadow.append(p + Vector2(3, 4))
		draw_colored_polygon(shadow, Color(0, 0, 0, 0.12 if detail_level < 3 else 0.2))
	draw_colored_polygon(pts, col)
	if detail_level >= 1:
		draw_polyline(pts + PackedVector2Array([pts[0]]), Color(1, 1, 1, 0.12), 1.0, true)

func _draw_building_sprite(center: Vector2, typ: String, zone: String) -> void:
	var name := str(building_sprite.get(typ, ""))
	var tex := _tex(name) if name != "" else null
	if tex == null:
		tex = _tex("bld_" + typ)
	if tex == null:
		_draw_building(center, zone)
		return
	var s := tex.get_size()
	var scale := 0.78 if detail_level < 3 else 0.86
	draw_texture_rect(tex, Rect2(center - Vector2(s.x * scale * 0.5, s.y * scale * 0.78), s * scale), false)

func _draw_building(center: Vector2, zone: String) -> void:
	var h := 28.0 if zone == "industrial" else 18.0
	if detail_level >= 3:
		h += 6.0
	var top := center + Vector2(0, -h)
	var base_col := Color("D9C4A0") if zone != "industrial" else Color("9AA3AD")
	var wall := base_col.darkened(0.15)
	var b0 := center + Vector2(-14, 0)
	var b1 := center + Vector2(14, 0)
	var b2 := center + Vector2(0, 10)
	var t0 := top + Vector2(-14, 0)
	var t1 := top + Vector2(14, 0)
	var t2 := top + Vector2(0, 10)
	if detail_level >= 1:
		draw_colored_polygon(PackedVector2Array([b0 + Vector2(4, 6), b2 + Vector2(4, 6), b1 + Vector2(4, 6)]), Color(0, 0, 0, 0.18))
	draw_colored_polygon(PackedVector2Array([b0, b2, t2, t0]), wall)
	draw_colored_polygon(PackedVector2Array([b1, b2, t2, t1]), wall.darkened(0.1))
	draw_colored_polygon(PackedVector2Array([t0, t1, t2]), base_col.lightened(0.1))
	if detail_level >= 2:
		draw_rect(Rect2(center + Vector2(-4, -h + 6), Vector2(4, 4)), Color(0.75, 0.9, 1.0, 0.55))

func _draw_shop_marker(center: Vector2, rented: bool) -> void:
	var c := Color("E2B14B") if rented else Color("F5E6C8")
	draw_circle(center + Vector2(0, -6), 5, c)
	draw_arc(center + Vector2(0, -6), 5, 0, TAU, 16, Color(0, 0, 0, 0.25), 1.0)

func _draw_road(center: Vector2) -> void:
	var strip := PackedVector2Array([
		center + Vector2(0, -6),
		center + Vector2(18, 0),
		center + Vector2(0, 6),
		center + Vector2(-18, 0),
	])
	draw_colored_polygon(strip, Color(0.35, 0.36, 0.38, 0.85))
	draw_line(center + Vector2(-10, 0), center + Vector2(10, 0), Color(0.85, 0.75, 0.35, 0.7), 1.0)

func _draw_catchment(cx: int, cy: int, radius: int) -> void:
	for p in parcels:
		var x: int = int(p.x)
		var y: int = int(p.y)
		if abs(x - cx) + abs(y - cy) <= radius:
			var c := _iso(x, y)
			draw_circle(c, 4, Color(0.85, 0.7, 0.2, 0.2))

func _draw_deposit(center: Vector2, item_id: String, fill_ratio: float) -> void:
	var tex := _tex("res_" + item_id)
	if tex == null:
		tex = _tex("item_" + item_id)
	if tex:
		var s := tex.get_size()
		draw_texture_rect(tex, Rect2(center + Vector2(-s.x * 0.28, -s.y * 0.15), s * 0.55), false)
		draw_arc(center + Vector2(0, 8), 8, -PI/2, -PI/2 + TAU * clampf(fill_ratio, 0.05, 1.0), 16, Color(1, 1, 1, 0.55), 2.0)
		return
	var col := Color("8B7355")
	match item_id:
		"oil": col = Color("2C2C2C")
		"iron_ore": col = Color("6E5A4B")
		"wood": col = Color("3E6B3A")
		"sand": col = Color("C2B280")
		"quartz": col = Color("A8C0D0")
		"milk": col = Color("E8E4D0")
		"cherry": col = Color("8B1E3F")
		"flax": col = Color("6B8F3C")
		"chemicals": col = Color("4A6FA5")
		"wheat": col = Color("D2B45A")
		"coal": col = Color("2A2A2E")
		"cattle": col = Color("A07850")
		"apple": col = Color("B04040")
		"potato": col = Color("A07846")
		"gas": col = Color("A8C8E0")
		"copper_ore": col = Color("B06A3A")
		"cotton": col = Color("E8E8E0")
	draw_circle(center + Vector2(0, 2), 7, col)
	draw_arc(center + Vector2(0, 2), 7, -PI/2, -PI/2 + TAU * clampf(fill_ratio, 0.05, 1.0), 16, Color(1, 1, 1, 0.55), 2.0)

func maxf(a: float, b: float) -> float:
	return a if a > b else b

func _input(event: InputEvent) -> void:
	if event is InputEventMouseMotion:
		var local := to_local((event as InputEventMouseMotion).position)
		hover = _pick(local)
		queue_redraw()
	elif event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		var local := to_local((event as InputEventMouseButton).position)
		var tile := _pick(local)
		if tile.x == 9999:
			return
		selected = tile
		var parcel := {}
		for p in parcels:
			if int(p.x) == tile.x and int(p.y) == tile.y:
				parcel = p
				break
		tile_clicked.emit(tile.x, tile.y, parcel)
		queue_redraw()

func _pick(local: Vector2) -> Vector2i:
	var a := local.x / (TILE_W * 0.5)
	var b := local.y / (TILE_H * 0.5)
	var x := int(round((a + b) * 0.5))
	var y := int(round((b - a) * 0.5))
	for p in parcels:
		if int(p.x) == x and int(p.y) == y:
			return Vector2i(x, y)
	return Vector2i(9999, 9999)
