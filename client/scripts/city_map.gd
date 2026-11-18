extends Node2D
class_name CityMap

signal tile_clicked(x: int, y: int, parcel: Dictionary)

const TILE_W := 64.0
const TILE_H := 32.0

const GRASS := preload("res://assets/city_kenney/grass.png")
const PAVING := preload("res://assets/city_kenney/paving.png")
const SPRITES := {
	"factory": preload("res://assets/city_kenney/factory.png"),
	"refinery": preload("res://assets/city_kenney/refinery.png"),
	"logistics": preload("res://assets/city_kenney/warehouse.png"),
	"extract": preload("res://assets/city_kenney/extract.png"),
	"farm": preload("res://assets/city_kenney/farm.png"),
	"shop": preload("res://assets/city_kenney/shop.png"),
	"tree": preload("res://assets/city_kenney/tree.png"),
}

const ROAD_DIRECTIONS := [Vector2i.RIGHT, Vector2i.DOWN, Vector2i.LEFT, Vector2i.UP]

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

var _view := Rect2()
var _frame := Rect2()
var _zoom := 1.0
var _pan := Vector2.ZERO
var _touches: Dictionary = {}
var _press_origin := Vector2.ZERO
var _dragging := false
var _mouse_down := false

class ParcelTile extends Node2D:
	var visual: Dictionary = {}

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
		_zoom = 1.12
		_pan = Vector2.ZERO
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
		var road_links := PackedVector2Array()
		if parcel["has_road"]:
			for direction in ROAD_DIRECTIONS:
				var neighbor := parcel_at(coord + direction)
				if neighbor.get("has_road", false):
					var rise := float(neighbor["elevation"]) - float(parcel["elevation"])
					road_links.append(_iso(direction, rise) * 0.5)
		var visual := {
			"zone": parcel["zone"],
			"road": parcel["has_road"],
			"road_links": road_links,
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

func visual_bounds() -> Rect2:
	var bounds := Rect2()
	for tile in _tiles.values():
		var footprint := Rect2(tile.position - Vector2(34, 49), Vector2(68, 81))
		bounds = footprint if not bounds.has_area() else bounds.merge(footprint)
	return bounds

func _draw_parcel(coord: Vector2i, tile: ParcelTile) -> void:
	var zone: String = tile.visual.zone
	var highlight := Color.WHITE
	if coord == selected:
		highlight = Color(1.22, 1.17, 0.85)
	elif coord == hover:
		highlight = Color(1.12, 1.10, 1.02)
	var terrain := GRASS if zone in ["rural", "residential", "empty"] else PAVING
	if zone == "industrial":
		highlight *= Color("#adb8b8")
	var terrain_size := terrain.get_size() * (TILE_W / terrain.get_width())
	tile.draw_texture_rect(terrain, Rect2(Vector2(-TILE_W * 0.5, -TILE_H * 0.5), terrain_size), false, highlight)
	if coord == selected:
		var outline := PackedVector2Array([Vector2(0, -16), Vector2(32, 0), Vector2(0, 16), Vector2(-32, 0), Vector2(0, -16)])
		tile.draw_polyline(outline, Color("#eed481"), 1.5, true)
	if tile.visual.road:
		_draw_road(tile, tile.visual.road_links)
	for deposit in tile.visual.deposits:
		_draw_deposit(tile, Vector2.ZERO, deposit[0], float(deposit[1]) / maxf(1.0, float(deposit[2])))
	var building_type: String = tile.visual.building_type
	var rented: bool = tile.visual.rented
	if not building_type.is_empty():
		var entry: Dictionary = _building_icons.get(building_type, {})
		var sprite := str(entry.get("category", "factory"))
		if sprite == "extract" and zone in ["rural", "residential"]:
			sprite = "farm"
		elif "refiner" in building_type:
			sprite = "refinery"
		_draw_map_sprite(tile, SPRITES.get(sprite, SPRITES["factory"]), Vector2(5, 8), Vector2(46, 48))
	elif rented:
		_draw_map_sprite(tile, SPRITES["shop"], Vector2(5, 8), Vector2(37, 44))
	elif tile.visual.retail_slot:
		tile.draw_circle(Vector2(7, -2), 3, Color("#f5e6c8"))
	elif detail_level > 0 and zone in ["rural", "empty"] and not tile.visual.road and tile.visual.deposits.is_empty() and posmod(coord.x * 13 + coord.y * 7, 5) < 2:
		_draw_map_sprite(tile, SPRITES["tree"], Vector2(8, 4), Vector2(16, 27))
	if tile.visual.owned:
		tile.draw_circle(Vector2(18, -22), 4, Color("E0C575"))

func _iso(coord: Vector2i, elevation: float) -> Vector2:
	return Vector2((coord.x - coord.y) * TILE_W * 0.5, (coord.x + coord.y) * TILE_H * 0.5 - elevation * 10.0)

func _draw_map_sprite(tile: ParcelTile, texture: Texture2D, base: Vector2, limit: Vector2) -> void:
	var original := texture.get_size()
	var size := original * minf(limit.x / original.x, limit.y / original.y)
	tile.draw_texture_rect(texture, Rect2(base - Vector2(size.x * 0.5, size.y), size), false)

func _draw_road(tile: ParcelTile, links: PackedVector2Array) -> void:
	var asphalt := Color("#424951")
	tile.draw_circle(Vector2.ZERO, 3.5, asphalt)
	for end in links:
		tile.draw_line(Vector2.ZERO, end, asphalt, 7.0, true)
		tile.draw_line(end * 0.28, end * 0.77, Color("#b6a16f"), 0.8, true)

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
	if _item_icons.has(item_id):
		var tex := CatalogUI.icon(_item_icons[item_id], "items")
		if tex != null:
			tile.draw_texture_rect(tex, Rect2(center + Vector2(-20, -3), Vector2(14, 14)), false)
			tile.draw_arc(center + Vector2(-13, 4), 8, -PI/2, -PI/2 + TAU * clampf(fill_ratio, 0.05, 1.0), 12, Color(1, 1, 1, 0.65), 1.0)
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

func frame_map(area: Rect2) -> void:
	_view = area
	_frame = visual_bounds()
	_apply_camera()

func reset_camera() -> void:
	_zoom = 1.0
	_pan = Vector2.ZERO
	_apply_camera()

func zoom_at(factor: float, focus: Vector2) -> void:
	var previous := _zoom
	_zoom = clampf(_zoom * factor, 0.7, 4.0)
	_pan = (focus - _view.get_center()) * (1.0 - _zoom / previous) + _pan * (_zoom / previous)
	_apply_camera()

func zoom_center(factor: float) -> void:
	zoom_at(factor, _view.get_center())

func _apply_camera() -> void:
	if not _view.has_area() or not _frame.has_area():
		return
	var fit := minf(_view.size.x / _frame.size.x, _view.size.y / _frame.size.y)
	var magnification := fit * _zoom
	var travel := ((_frame.size * magnification + _view.size) * 0.5 - Vector2(80, 80)).max(Vector2.ZERO)
	_pan = _pan.clamp(-travel, travel)
	var host := get_parent() as Node2D
	host.scale = Vector2.ONE * magnification
	host.position = _view.get_center() + _pan - _frame.get_center() * magnification

func _pan_by(delta: Vector2) -> void:
	_pan += delta
	_apply_camera()

func _drag_from(origin: Vector2, current: Vector2, relative: Vector2) -> void:
	if _dragging:
		_pan_by(relative)
	elif current.distance_to(origin) >= 10.0:
		_dragging = true
		_pan_by(current - origin)

func _input(event: InputEvent) -> void:
	# Complete a map gesture even when the pointer leaves the map over a control.
	if _mouse_down and (event is InputEventMouseMotion or (event is InputEventMouseButton and not event.pressed)):
		_unhandled_input(event)
	elif (event is InputEventScreenDrag or (event is InputEventScreenTouch and not event.pressed)) and _touches.has(event.index):
		_unhandled_input(event)

func _unhandled_input(event: InputEvent) -> void:
	if not is_visible_in_tree() or event.device == InputEvent.DEVICE_ID_EMULATION:
		return
	if event is InputEventScreenTouch:
		if event.pressed:
			if not _view.has_point(event.position):
				return
			if _touches.is_empty():
				_press_origin = event.position
				_dragging = false
			else:
				_dragging = true
			_touches[event.index] = event.position
		elif _touches.has(event.index):
			_touches.erase(event.index)
			if not _dragging and not event.canceled:
				_select_at(event.position)
		else:
			return
	elif event is InputEventScreenDrag and _touches.has(event.index):
		var last: Vector2 = _touches[event.index]
		_touches[event.index] = event.position
		if _touches.size() == 1:
			_drag_from(_press_origin, event.position, event.position - last)
		elif _touches.size() == 2:
			var other: Vector2
			for finger in _touches:
				if finger != event.index:
					other = _touches[finger]
			var distance := other.distance_to(last)
			if distance >= 8.0:
				zoom_at(other.distance_to(event.position) / distance, (last + other) * 0.5)
			_pan_by((event.position - last) * 0.5)
	elif event is InputEventMagnifyGesture and _view.has_point(event.position):
		zoom_at(event.factor, event.position)
	elif event is InputEventMouseButton:
		if event.pressed and _view.has_point(event.position):
			match event.button_index:
				MOUSE_BUTTON_WHEEL_UP: zoom_at(1.18, event.position)
				MOUSE_BUTTON_WHEEL_DOWN: zoom_at(1.0 / 1.18, event.position)
				MOUSE_BUTTON_LEFT:
					_press_origin = event.position
					_mouse_down = true
					_dragging = false
				_: return
		elif not event.pressed and event.button_index == MOUSE_BUTTON_LEFT and _mouse_down:
			_mouse_down = false
			if not _dragging:
				_select_at(event.position)
		else:
			return
	elif event is InputEventMouseMotion:
		if _mouse_down:
			_drag_from(_press_origin, event.position, event.relative)
		else:
			var tile := _tile_at(event.position)
			if tile != hover:
				_redraw_at(hover)
				hover = tile
				_redraw_at(hover)
			return
	else:
		return
	get_viewport().set_input_as_handled()

func _tile_at(point: Vector2) -> Vector2i:
	if not _view.has_point(point):
		return NO_TILE
	return _pick(get_global_transform_with_canvas().affine_inverse() * point)

func _select_at(point: Vector2) -> void:
	var tile := _tile_at(point)
	if tile == NO_TILE:
		return
	_redraw_at(selected)
	selected = tile
	_redraw_at(selected)
	_catchment.queue_redraw()
	tile_clicked.emit(tile.x, tile.y, parcel_at(tile))

func _pick(local: Vector2) -> Vector2i:
	for index in range(_tile_order.size() - 1, -1, -1):
		var coord := _tile_order[index]
		var tile: ParcelTile = _tiles[coord]
		var offset := (local - tile.position).abs()
		if offset.x / (TILE_W * 0.5) + offset.y / (TILE_H * 0.5) <= 1.0:
			return coord
	return NO_TILE
