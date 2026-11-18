extends Node2D
class_name CityAmbience

## Декоративный слой: деревья и автомобили. Зависит от GraphicsSettings.

var gfx: GraphicsSettings
var trees: Array[Node2D] = []
var cars: Array[Dictionary] = []
var _time := 0.0

func setup(settings: GraphicsSettings) -> void:
	gfx = settings
	gfx.quality_changed.connect(func(_q): _rebuild())
	_rebuild()

func _rebuild() -> void:
	for t in trees:
		t.queue_free()
	trees.clear()
	cars.clear()
	if gfx == null:
		return
	# trees around map
	var tree_count := 0
	if gfx.quality >= GraphicsSettings.QUALITY_MEDIUM:
		tree_count = 8
	if gfx.quality >= GraphicsSettings.QUALITY_HIGH:
		tree_count = 16
	if gfx.quality >= GraphicsSettings.QUALITY_ULTRA:
		tree_count = 28
	for i in tree_count:
		var t := Node2D.new()
		t.position = Vector2(randf_range(-420, 420), randf_range(-260, 300))
		t.set_meta("phase", randf() * TAU)
		t.set_meta("base_x", t.position.x)
		add_child(t)
		trees.append(t)
	var density := gfx.traffic_density()
	var car_n := int(12 * density)
	for i in car_n:
		cars.append({
			"pos": Vector2(randf_range(-380, 380), randf_range(-40, 40) + (i % 3) * 18.0),
			"speed": randf_range(28, 70) * (1.0 if randf() > 0.5 else -1.0),
			"color": Color(randf_range(0.55, 0.95), randf_range(0.45, 0.75), randf_range(0.35, 0.55)),
		})
	queue_redraw()

func _process(delta: float) -> void:
	if gfx == null:
		return
	_time += delta
	if gfx.enable_tree_sway():
		for t in trees:
			var phase: float = t.get_meta("phase")
			var base: float = t.get_meta("base_x")
			t.position.x = base + sin(_time * 1.5 + phase) * 3.0
	for c in cars:
		c.pos.x += c.speed * delta
		if c.pos.x > 460:
			c.pos.x = -460
		elif c.pos.x < -460:
			c.pos.x = 460
	if gfx.enable_cars() or gfx.enable_tree_sway():
		queue_redraw()

func _draw() -> void:
	if gfx == null:
		return
	for t in trees:
		_draw_tree(t.position, gfx.enable_tree_sway())
	if gfx.enable_cars():
		for c in cars:
			_draw_car(c.pos, c.color, c.speed < 0)

func _draw_tree(p: Vector2, sway: bool) -> void:
	var lean := sin(_time * 1.5 + p.x * 0.01) * (3.0 if sway else 0.0)
	draw_line(p, p + Vector2(lean, -22), Color("5C4033"), 3.0)
	draw_circle(p + Vector2(lean, -28), 12, Color("3F6B4A"))
	draw_circle(p + Vector2(lean - 6, -22), 8, Color("4F7D58"))

func _draw_car(p: Vector2, col: Color, flip: bool) -> void:
	var dir := -1.0 if flip else 1.0
	var body := Rect2(p + Vector2(-10 * dir, -5), Vector2(20 * dir, 8))
	draw_rect(body, col)
	draw_circle(p + Vector2(-6 * dir, 4), 2.5, Color(0.1, 0.1, 0.1))
	draw_circle(p + Vector2(6 * dir, 4), 2.5, Color(0.1, 0.1, 0.1))
	if gfx.enable_shaders():
		draw_rect(Rect2(p + Vector2(-4 * dir, -8), Vector2(8 * dir, 4)), Color(0.7, 0.85, 0.95, 0.7))
