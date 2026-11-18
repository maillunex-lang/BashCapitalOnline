extends MarginContainer
class_name SafeAreaRoot

func _ready() -> void:
	_apply()
	get_viewport().size_changed.connect(_apply)

func _notification(what: int) -> void:
	if what == NOTIFICATION_WM_WINDOW_FOCUS_IN and is_node_ready():
		_apply()

func _apply() -> void:
	var margins := Vector4(8, 12, 8, 8)
	if OS.has_feature("mobile"):
		var safe := DisplayServer.get_display_safe_area()
		var window := DisplayServer.window_get_size()
		if safe.has_area() and window.x > 0 and window.y > 0:
			var scale := get_viewport().get_stretch_transform().get_scale()
			var safe_end := safe.end
			margins.x = maxf(margins.x, ceilf(safe.position.x / scale.x))
			margins.y = maxf(margins.y, ceilf(safe.position.y / scale.y))
			margins.z = maxf(margins.z, ceilf((window.x - safe_end.x) / scale.x))
			margins.w = maxf(margins.w, ceilf((window.y - safe_end.y) / scale.y))
	add_theme_constant_override("margin_left", int(margins.x))
	add_theme_constant_override("margin_top", int(margins.y))
	add_theme_constant_override("margin_right", int(margins.z))
	add_theme_constant_override("margin_bottom", int(margins.w))
