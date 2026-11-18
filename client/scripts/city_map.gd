extends Node2D
class_name CityMap

signal tile_clicked(x: int, y: int, parcel: Dictionary)

const TILE_W := 64.0
const TILE_H := 32.0

const NO_TILE := Vector2i(9999, 9999)

var selected := NO_TILE
var hover := NO_TILE
var detail_level := 2:
	set(value):
		if detail_level != value:
			detail_level = value
			_redraw_tiles()

var _city_id := ""
var _parcels: Dictionary = {}
var _buildings: Dictionary = {}
var _shops: Dictionary = {}
var _tiles: Dictionary = {}
var _tile_order: Array[Vector2i] = []
var _shop_radii: Dictionary = {}
var _catchment := Node2D.new()
var _sprites: Dictionary = {}

class ParcelTile extends Node2D:
	var visual: Dictionary = {}

var zone_colors := {
	"center": Color("2F6F6A"),
	"residential": Color("4A7C59"),
	"commercial": Color("C4A35A"),
	"industrial": Color("6B7280"),
	"rural": Color("7FAE6E"),
	"empty": Color("3A4A46"),
}

var _building_icons: Dictionary = {}
var _item_icons: Dictionary = {}

func bind_catalog(catalog: Dictionary) -> void:
	_building_icons.clear()
	_item_icons.clear()
	for entry in catalog.get("buildings", []):
		_building_icons[entry["id"]] = entry
	for entry in catalog.get("items", []):
		_item_icons[entry["id"]] = entry
	_redraw_tiles()

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
		"res://assets/ui/%s.png" % name,
	]:
		if ResourceLoader.exists(path):
			var t: Texture2D = load(path)
			_sprites[name] = t
			return t
	_sprites[name] = null
	return null

func _ready() -> void:
	_catchment.z_index = 1
	_catchment.draw.connect(_draw_catchment)
	add_child(_catchment)

func _coord(entry: Dictionary) -> Vector2i:
	return Vector2i(int(entry["x"]), int(entry["y"]))

func parcel_at(coord: Vector2i) -> Dictionary:
	return _parcels.get(coord, {})

func building_at(coord: Vector2i) -> Dictionary:
	return _buildings.get(coord, {})

func shop_at(coord: Vector2i) -> Dictionary:
	return _shops.get(coord, {})

func set_city_data(city: Dictionary, world_buildings: Dictionary, world_shops: Dictionary, firm_id := "") -> void:
	var city_id: String = city["id"]
	var switch_city := _city_id != city_id
	_city_id = city_id
	if switch_city:
		selected = NO_TILE
		hover = NO_TILE

	var parcels := {}
	var deposits := {}
	var buildings := {}
	var shops := {}
	for parcel in city.get("parcels", []):
		parcels[_coord(parcel)] = parcel
	var resources = city.get("deposits")
	if resources is Array:
		for deposit in resources:
			if deposit["amount"] > 0:
				var coord := _coord(deposit)
				if not deposits.has(coord):
					deposits[coord] = []
				deposits[coord].append([deposit["item_id"], deposit["amount"], deposit["max_amount"]])
	for building in world_buildings.values():
		if building["city_id"] == city_id:
			buildings[_coord(building)] = building
	var shop_radii := {}
	for shop in world_shops.values():
		if shop["city_id"] == city_id:
			var coord := _coord(shop)
			shops[coord] = shop
			shop_radii[coord] = int(shop.get("radius", 2))

	# The server's bounded dirty list can be cleared, and resource quantities
	# can change between ticks without a revision. Diff the rendered fields.
	var layout_changed := switch_city or parcels.size() != _tiles.size()
	var old_catchment_radius: int = _shop_radii.get(selected, 0)
	_shop_radii = shop_radii
	_parcels = parcels
	_buildings = buildings
	_shops = shops
	for coord in _tiles.keys():
		if switch_city or not parcels.has(coord):
			var tile: ParcelTile = _tiles[coord]
			remove_child(tile)
			tile.queue_free()
			_tiles.erase(coord)
			layout_changed = true
	for coord in parcels:
		var parcel: Dictionary = parcels[coord]
		var building: Dictionary = buildings.get(coord, {})
		var shop: Dictionary = shops.get(coord, {})
		var owner := str(building.get("firm_id", shop.get("firm_id", "")))
		var visual := {
			"zone": parcel["zone"],
			"road": parcel["has_road"],
			"retail_slot": int(parcel["retail_slots"]) > 0,
			"building_type": str(building.get("type", "")),
			"rented": not shop.is_empty(),
			"deposits": deposits.get(coord, []),
			"owned": owner == firm_id and not owner.is_empty(),
		}
		var tile: ParcelTile = _tiles.get(coord)
		if tile == null:
			tile = ParcelTile.new()
			tile.draw.connect(_draw_parcel.bind(coord, tile))
			_tiles[coord] = tile
			add_child(tile)
			layout_changed = true
		var center := _iso(coord, float(parcel["elevation"]))
		if center != tile.position:
			tile.position = center
			layout_changed = true
		if tile.visual != visual:
			tile.visual = visual
			tile.queue_redraw()
	if not parcels.has(selected):
		selected = NO_TILE
	if not parcels.has(hover):
		hover = NO_TILE
	if layout_changed:
		_tile_order.assign(_tiles.keys())
		_tile_order.sort_custom(func(a: Vector2i, b: Vector2i):
			var ay: float = _tiles[a].position.y
			var by: float = _tiles[b].position.y
			return a.x < b.x if ay == by else ay < by
		)
		for index in _tile_order.size():
			move_child(_tiles[_tile_order[index]], index)
	if layout_changed or old_catchment_radius != int(_shop_radii.get(selected, 0)):
		_catchment.queue_redraw()

func _draw() -> void:
	draw_circle(Vector2.ZERO, 520, Color(0.18, 0.28, 0.24, 0.35))

func _draw_parcel(coord: Vector2i, tile: ParcelTile) -> void:
	var zone: String = tile.visual.zone
	var col: Color = zone_colors.get(zone, zone_colors["empty"])
	if coord == selected:
		col = col.lightened(0.25)
	elif coord == hover:
		col = col.lightened(0.12)
	var tile_tex := _tex(str(zone_tile.get(zone, "tile_grass")))
	if tile_tex:
		var size := tile_tex.get_size()
		var tint := Color(1, 1, 1, 0.96)
		if coord == selected:
			tint = Color(1.18, 1.18, 1.12)
		elif coord == hover:
			tint = Color(1.08, 1.08, 1.06)
		tile.draw_texture_rect(tile_tex, Rect2(-size * Vector2(0.5, 0.62), size), false, tint)
	else:
		_draw_tile(tile, Vector2.ZERO, col)
	if tile.visual.road:
		var road := _tex("tile_road")
		if road:
			var size := road.get_size()
			tile.draw_texture_rect(road, Rect2(-size * Vector2(0.5, 0.55), size * 0.55), false, Color(1, 1, 1, 0.9))
		else:
			_draw_road(tile, Vector2.ZERO)
	for deposit in tile.visual.deposits:
		_draw_deposit(tile, Vector2.ZERO, deposit[0], float(deposit[1]) / maxf(1.0, float(deposit[2])))
	var building_type: String = tile.visual.building_type
	var rented: bool = tile.visual.rented
	if not building_type.is_empty():
		_draw_building_sprite(tile, Vector2.ZERO, building_type, zone)
	elif tile.visual.retail_slot or rented:
		var shop_tex := _tex("bld_shop")
		if shop_tex:
			var size := shop_tex.get_size()
			tile.draw_texture_rect(shop_tex, Rect2(-size * Vector2(0.5, 0.72), size * 0.72), false)
		_draw_shop_marker(tile, Vector2.ZERO, rented)
	if tile.visual.owned:
		tile.draw_circle(Vector2(18, -22), 4, Color("E0C575"))

func _iso(coord: Vector2i, elevation: float) -> Vector2:
	return Vector2((coord.x - coord.y) * TILE_W * 0.5, (coord.x + coord.y) * TILE_H * 0.5 - elevation * 10.0)

func _draw_tile(tile: ParcelTile, center: Vector2, col: Color) -> void:
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
		tile.draw_colored_polygon(shadow, Color(0, 0, 0, 0.12 if detail_level < 3 else 0.2))
	tile.draw_colored_polygon(pts, col)
	if detail_level >= 1:
		tile.draw_polyline(pts + PackedVector2Array([pts[0]]), Color(1, 1, 1, 0.12), 1.0, true)

func _draw_building_sprite(tile: ParcelTile, center: Vector2, typ: String, zone: String) -> void:
	var entry: Dictionary = _building_icons.get(typ, {})
	if entry.is_empty():
		_draw_building(tile, center, zone)
		return
	var tex := CatalogUI.icon(entry, "buildings")
	var s := tex.get_size()
	var scale := 0.78 if detail_level < 3 else 0.86
	tile.draw_texture_rect(tex, Rect2(center - Vector2(s.x * scale * 0.5, s.y * scale * 0.78), s * scale), false)

func _draw_building(tile: ParcelTile, center: Vector2, zone: String) -> void:
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
		tile.draw_colored_polygon(PackedVector2Array([b0 + Vector2(4, 6), b2 + Vector2(4, 6), b1 + Vector2(4, 6)]), Color(0, 0, 0, 0.18))
	tile.draw_colored_polygon(PackedVector2Array([b0, b2, t2, t0]), wall)
	tile.draw_colored_polygon(PackedVector2Array([b1, b2, t2, t1]), wall.darkened(0.1))
	tile.draw_colored_polygon(PackedVector2Array([t0, t1, t2]), base_col.lightened(0.1))
	if detail_level >= 2:
		tile.draw_rect(Rect2(center + Vector2(-4, -h + 6), Vector2(4, 4)), Color(0.75, 0.9, 1.0, 0.55))

func _draw_shop_marker(tile: ParcelTile, center: Vector2, rented: bool) -> void:
	var c := Color("E2B14B") if rented else Color("F5E6C8")
	tile.draw_circle(center + Vector2(0, -6), 5, c)
	tile.draw_arc(center + Vector2(0, -6), 5, 0, TAU, 16, Color(0, 0, 0, 0.25), 1.0)

func _draw_road(tile: ParcelTile, center: Vector2) -> void:
	var strip := PackedVector2Array([
		center + Vector2(0, -6),
		center + Vector2(18, 0),
		center + Vector2(0, 6),
		center + Vector2(-18, 0),
	])
	tile.draw_colored_polygon(strip, Color(0.35, 0.36, 0.38, 0.85))
	tile.draw_line(center + Vector2(-10, 0), center + Vector2(10, 0), Color(0.85, 0.75, 0.35, 0.7), 1.0)

func _draw_catchment() -> void:
	var radius: int = _shop_radii.get(selected, 0)
	if radius <= 0:
		return
	for y in range(-radius, radius + 1):
		for x in range(-radius + absi(y), radius - absi(y) + 1):
			var tile: ParcelTile = _tiles.get(selected + Vector2i(x, y))
			if tile:
				_catchment.draw_circle(tile.position, 4, Color(0.85, 0.7, 0.2, 0.2))

func _draw_deposit(tile: ParcelTile, center: Vector2, item_id: String, fill_ratio: float) -> void:
	var tex := _tex("res_" + item_id)
	if tex == null and _item_icons.has(item_id):
		tex = CatalogUI.icon(_item_icons[item_id], "items")
	if tex:
		var s := tex.get_size()
		tile.draw_texture_rect(tex, Rect2(center + Vector2(-s.x * 0.28, -s.y * 0.15), s * 0.55), false)
		tile.draw_arc(center + Vector2(0, 8), 8, -PI/2, -PI/2 + TAU * clampf(fill_ratio, 0.05, 1.0), 16, Color(1, 1, 1, 0.55), 2.0)
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
	tile.draw_circle(center + Vector2(0, 2), 7, col)
	tile.draw_arc(center + Vector2(0, 2), 7, -PI/2, -PI/2 + TAU * clampf(fill_ratio, 0.05, 1.0), 16, Color(1, 1, 1, 0.55), 2.0)

func _redraw_tiles() -> void:
	for tile in _tiles.values():
		tile.queue_redraw()

func _redraw_at(coord: Vector2i) -> void:
	var tile: ParcelTile = _tiles.get(coord)
	if tile:
		tile.queue_redraw()

func _unhandled_input(event: InputEvent) -> void:
	if not is_visible_in_tree():
		return
	if event is InputEventMouseMotion:
		var tile := _pick(get_global_transform_with_canvas().affine_inverse() * event.position)
		if tile != hover:
			_redraw_at(hover)
			hover = tile
			_redraw_at(hover)
	elif event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		var tile := _pick(get_global_transform_with_canvas().affine_inverse() * event.position)
		if tile == NO_TILE:
			return
		_redraw_at(selected)
		selected = tile
		_redraw_at(selected)
		_catchment.queue_redraw()
		tile_clicked.emit(tile.x, tile.y, parcel_at(tile))
		get_viewport().set_input_as_handled()

func _pick(local: Vector2) -> Vector2i:
	for index in range(_tile_order.size() - 1, -1, -1):
		var coord := _tile_order[index]
		var tile: ParcelTile = _tiles[coord]
		var offset := (local - tile.position).abs()
		if offset.x / (TILE_W * 0.5) + offset.y / (TILE_H * 0.5) <= 1.0:
			return coord
	return NO_TILE
