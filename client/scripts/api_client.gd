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

func _ready() -> void:
	pass

func configure(url: String) -> void:
	base_url = url.rstrip("/")

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
	return res if res is Dictionary else {}

func auth_login(username: String, password: String) -> Dictionary:
	var res = await _post("/api/auth/login", {"username": username, "password": password})
	if res is Dictionary and res.has("session_token"):
		session_token = str(res["session_token"])
		if res.get("firm_id", "") != "":
			firm_id = str(res["firm_id"])
	return res if res is Dictionary else {}

func route_quote(from_city: String, to_city: String) -> Dictionary:
	var res = await _get("/api/route?from=%s&to=%s" % [from_city, to_city])
	return res if res is Dictionary else {}

func create_payment(months: int = 1) -> Dictionary:
	var res = await _post("/api/payment/create", {"months": months})
	return res if res is Dictionary else {}

func fetch_catalog() -> void:
	var res = await _get("/api/catalog")
	if res is Dictionary:
		catalog = res

func fetch_world() -> void:
	var res = await _get("/api/world")
	if res is Dictionary:
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
	var res = await _get("/api/rankings")
	return res if res is Array else []

func _get(path: String):
	var http := HTTPRequest.new()
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
	connected_changed.emit(response_code >= 200 and response_code < 300)
	if response_code >= 400:
		return {"error": text, "code": response_code}
	var data = JSON.parse_string(text)
	return data
