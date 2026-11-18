extends Node
class_name GraphicsSettings

## Качество графики: 0 low, 1 medium, 2 high, 3 ultra
signal quality_changed(level: int)

const QUALITY_LOW := 0
const QUALITY_MEDIUM := 1
const QUALITY_HIGH := 2
const QUALITY_ULTRA := 3

var quality: int = QUALITY_HIGH
var fullscreen: bool = true

func _ready() -> void:
	if ConfigFile.new().load("user://settings.cfg") == OK:
		pass
	_load()

func _load() -> void:
	var cfg := ConfigFile.new()
	if cfg.load("user://settings.cfg") == OK:
		quality = int(cfg.get_value("graphics", "quality", QUALITY_HIGH))
		fullscreen = bool(cfg.get_value("display", "fullscreen", true))
	apply()

func save() -> void:
	var cfg := ConfigFile.new()
	cfg.set_value("graphics", "quality", quality)
	cfg.set_value("display", "fullscreen", fullscreen)
	cfg.save("user://settings.cfg")

func set_quality(level: int) -> void:
	quality = clampi(level, QUALITY_LOW, QUALITY_ULTRA)
	apply()
	save()
	quality_changed.emit(quality)

func set_fullscreen(on: bool) -> void:
	fullscreen = on
	apply_display()
	save()

func apply() -> void:
	apply_display()
	match quality:
		QUALITY_LOW:
			Engine.max_fps = 30
			get_tree().root.content_scale_factor = 0.85
			RenderingServer.viewport_set_msaa_2d(get_viewport().get_viewport_rid(), RenderingServer.VIEWPORT_MSAA_DISABLED)
		QUALITY_MEDIUM:
			Engine.max_fps = 60
			get_tree().root.content_scale_factor = 1.0
			RenderingServer.viewport_set_msaa_2d(get_viewport().get_viewport_rid(), RenderingServer.VIEWPORT_MSAA_2X)
		QUALITY_HIGH:
			Engine.max_fps = 120
			get_tree().root.content_scale_factor = 1.0
			RenderingServer.viewport_set_msaa_2d(get_viewport().get_viewport_rid(), RenderingServer.VIEWPORT_MSAA_4X)
		QUALITY_ULTRA:
			Engine.max_fps = 0
			get_tree().root.content_scale_factor = 1.0
			RenderingServer.viewport_set_msaa_2d(get_viewport().get_viewport_rid(), RenderingServer.VIEWPORT_MSAA_8X)

func apply_display() -> void:
	if fullscreen:
		DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)
	else:
		DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_WINDOWED)

func quality_name() -> String:
	match quality:
		QUALITY_LOW: return "Низкое"
		QUALITY_MEDIUM: return "Среднее"
		QUALITY_HIGH: return "Высокое"
		_: return "Ультра"

func enable_cars() -> bool:
	return quality >= QUALITY_HIGH

func enable_tree_sway() -> bool:
	return quality >= QUALITY_ULTRA

func enable_shaders() -> bool:
	return quality >= QUALITY_ULTRA

func enable_soft_shadows() -> bool:
	return quality >= QUALITY_MEDIUM

func traffic_density() -> float:
	match quality:
		QUALITY_LOW: return 0.0
		QUALITY_MEDIUM: return 0.15
		QUALITY_HIGH: return 0.45
		_: return 1.0
