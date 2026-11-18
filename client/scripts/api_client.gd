extends Node
class_name ApiClient

signal world_updated(world: Dictionary)
signal connected_changed(ok: bool)

var base_url: String = "http://127.0.0.1:8080"
var firm_id: String = ""
var session_token: String = ""
var firm: Dictionary = {}
var catalog: Dictionary = {}
var world: Dictionary = {}
var _world_requested := 0
var _world_received := 0

func _ready() -> void:
	pass

func configure(url: String) -> void:
	var next_url := url.rstrip("/")
	if next_url != base_url:
		session_token = ""
		firm_id = ""
		firm = {}
		world = {}
	base_url = next_url

func _headers(extra: Dictionary = {}) -> PackedStringArray:
	var h: PackedStringArray = ["Content-Type: application/json"]
	if firm_id != "":
		h.append("X-Firm-Id: %s" % firm_id)
	if session_token != "":
		h.append("X-Session-Token: %s" % session_token)
	return h

func auth_register(username: String, password: String) -> Dictionary:
	var res = await _post("/api/auth/register", {"username": username, "password": password})
	if res is Dictionary and res.has("session_token"):
		session_token = str(res["session_token"])
		firm_id = ""
	return res if res is Dictionary else {}

func auth_login(username: String, password: String) -> Dictionary:
	var res = await _post("/api/auth/login", {"username": username, "password": password})
	if res is Dictionary and res.has("session_token"):
		session_token = str(res["session_token"])
		firm_id = str(res.get("firm_id", ""))
	return res if res is Dictionary else {}

func route_quote(from_city: String, to_city: String) -> Dictionary:
	var res = await _http_get("/api/route?from=%s&to=%s" % [from_city, to_city])
	return res if res is Dictionary else {}

func create_payment(months: int = 1) -> Dictionary:
	var res = await _post("/api/payment/create", {"months": months})
	return res if res is Dictionary else {}

func fetch_manifest() -> Dictionary:
	var res = await _http_get("/api/updates/manifest")
	return res if res is Dictionary else {}

func fetch_catalog() -> bool:
	var res = await _http_get("/api/catalog")
	if not res is Dictionary or not res.has("economy_catalog_version"):
		return false
	for kind in ["items", "recipes", "buildings", "shops", "cities"]:
		if not res.get(kind) is Array:
			return false
	catalog = res
	return true

func fetch_world() -> void:
	_world_requested += 1
	var request := _world_requested
	var url := base_url
	var res = await _http_get("/api/world")
	if url == base_url and request > _world_received and res is Dictionary and res.get("cities") is Dictionary:
		_world_received = request
		world = res
		world_updated.emit(world)

func register_firm(firm_name: String, industry: String) -> Dictionary:
	var res = await _post("/api/register", {"name": firm_name, "industry": industry})
	if res is Dictionary and res.has("id"):
		firm = res
		firm_id = str(res["id"])
	return res if res is Dictionary else {}

func action(path: String, body: Dictionary) -> Dictionary:
	var res = await _post("/api/action/%s" % path, body)
	await fetch_world()
	return res if res is Dictionary else {"error": str(res)}

func manual_tick() -> Dictionary:
	var res = await _post("/api/tick", {})
	await fetch_world()
	return res if res is Dictionary else {}

func rankings() -> Array:
	var res = await _http_get("/api/rankings")
	return res if res is Array else []

func _http_get(path: String):
	var http := HTTPRequest.new()
	http.timeout = 15.0
	add_child(http)
	var err := http.request(base_url + path)
	if err != OK:
		http.queue_free()
		connected_changed.emit(false)
		return null
	var result = await http.request_completed
	http.queue_free()
	return _parse(result)

func _post(path: String, body: Dictionary):
	var http := HTTPRequest.new()
	http.timeout = 15.0
	add_child(http)
	var err := http.request(base_url + path, _headers(), HTTPClient.METHOD_POST, JSON.stringify(body))
	if err != OK:
		http.queue_free()
		connected_changed.emit(false)
		return null
	var result = await http.request_completed
	http.queue_free()
	return _parse(result)

func _parse(result: Array):
	var response_code: int = result[1]
	var body: PackedByteArray = result[3]
	var text := body.get_string_from_utf8()
	var ok := response_code >= 200 and response_code < 300
	connected_changed.emit(ok)
	if response_code == 0:
		return null
	if not ok:
		return {"error": text, "code": response_code}
	var data = JSON.parse_string(text)
	return data
