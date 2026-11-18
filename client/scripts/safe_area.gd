extends MarginContainer
class_name SafeAreaRoot

## Учитывает вырез камеры / notch / home indicator.
## Вешает отступы на корневой UI-контейнер.

func _ready() -> void:
	_apply()
	get_tree().root.size_changed.connect(_apply)
	# mobile/display cutouts update
	if DisplayServer.has_signal("window_safe_area_changed"):
		pass

func _notification(what: int) -> void:
	if what == NOTIFICATION_WM_WINDOW_FOCUS_IN or what == NOTIFICATION_RESIZED:
		_apply()

func _apply() -> void:
	var wa: Rect2i = DisplayServer.get_display_safe_area()
	var win: Vector2i = DisplayServer.window_get_size()
	if win.x <= 0 or win.y <= 0:
		return
	# Convert absolute safe rect into margins relative to window
	var left := maxi(wa.position.x, 0)
	var top := maxi(wa.position.y, 0)
	var right := maxi(win.x - (wa.position.x + wa.size.x), 0)
	var bottom := maxi(win.y - (wa.position.y + wa.size.y), 0)
	# Extra padding under typical camera punch-hole
	top = maxi(top, 12)
	left = maxi(left, 8)
	right = maxi(right, 8)
	bottom = maxi(bottom, 8)
	add_theme_constant_override("margin_left", left)
	add_theme_constant_override("margin_top", top)
	add_theme_constant_override("margin_right", right)
	add_theme_constant_override("margin_bottom", bottom)
