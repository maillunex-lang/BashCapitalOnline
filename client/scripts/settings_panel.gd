extends PanelContainer
class_name SettingsPanel

signal closed
signal quality_picked(level: int)
signal fullscreen_toggled(on: bool)
signal subscribe_pressed(months: int)
signal hire_deputy_pressed(deputy_name: String)
signal deputy_tasks_pressed(kinds: PackedStringArray)

@onready var quality: OptionButton = $Margin/VBox/Quality
@onready var fullscreen: CheckButton = $Margin/VBox/Fullscreen
@onready var sub_status: Label = $Margin/VBox/SubStatus
@onready var deputy_name: LineEdit = $Margin/VBox/DeputyName
@onready var task_list: VBoxContainer = $Margin/VBox/Tasks

var task_checks: Dictionary = {}

func _ready() -> void:
	quality.clear()
	quality.add_item("Низкое", 0)
	quality.add_item("Среднее", 1)
	quality.add_item("Высокое", 2)
	quality.add_item("Ультра", 3)
	for kind in [
		["procure_inputs", "Поиск поставщиков"],
		["restock", "Пополнение магазинов"],
		["price_retail", "Цены в рознице"],
		["sell_excess", "Сбыт излишков"],
		["research", "R&D"],
		["auto_contracts", "Автоконтракты"],
	]:
		var cb := CheckBox.new()
		cb.text = kind[1]
		cb.set_meta("kind", kind[0])
		task_list.add_child(cb)
		task_checks[kind[0]] = cb

func bind_state(q: int, fs: bool, subscribed: bool, until_tick: int, deputy: Dictionary) -> void:
	quality.select(q)
	fullscreen.button_pressed = fs
	if subscribed:
		sub_status.text = "Подписка активна до тика %s" % until_tick
	else:
		sub_status.text = "Подписка не активна — заместитель недоступен"
	if deputy.has("name"):
		deputy_name.text = str(deputy.get("name", "Заместитель"))
	var active_kinds: Array = []
	for t in deputy.get("tasks", []):
		if t.get("active", false):
			active_kinds.append(str(t.get("kind")))
	for kind in task_checks.keys():
		task_checks[kind].button_pressed = kind in active_kinds

func _on_quality_item_selected(index: int) -> void:
	quality_picked.emit(index)

func _on_fullscreen_toggled(on: bool) -> void:
	fullscreen_toggled.emit(on)

func _on_close() -> void:
	closed.emit()

func _on_subscribe() -> void:
	subscribe_pressed.emit(1)

func _on_hire() -> void:
	hire_deputy_pressed.emit(deputy_name.text.strip_edges())

func _on_apply_tasks() -> void:
	var kinds: PackedStringArray = []
	for kind in task_checks.keys():
		if task_checks[kind].button_pressed:
			kinds.append(kind)
	deputy_tasks_pressed.emit(kinds)
