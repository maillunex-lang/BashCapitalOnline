extends Node
class_name RemoteAssets

signal packs_ready
signal update_available(info: Dictionary)
signal progress(msg: String)

const MAX_PACK_SIZE := 16 * 1024 * 1024
const PACK_PROBES := {"catalog_v2": "catalog_v2/icons.json", "city_tiles": "map/tile_grass.png"}

static var _mounted: Dictionary = {}

var api: ApiClient
var cache_dir := "user://asset_packs"

func setup(client: ApiClient) -> void:
	api = client
	DirAccess.make_dir_recursive_absolute(cache_dir)

func check_and_fetch() -> void:
	if api == null:
		return
	for archive in _mounted.values():
		archive.close()
	_mounted.clear()
	progress.emit("Проверка манифеста ассетов…")
	var manifest := await api.fetch_manifest()
	if manifest.is_empty() or manifest.has("error"):
		progress.emit("Манифест недоступен · используются встроенные ассеты")
		packs_ready.emit()
		return

	var apk: Dictionary = manifest.get("apk", {})
	var local_apk_code := int(ProjectSettings.get_setting("application/config/android_version_code", 0))
	if OS.has_feature("android") and not OS.has_feature("android_preview"):
		if _valid_sha(str(apk.get("sha256", ""))) and int(apk.get("version_code", 0)) > local_apk_code:
			update_available.emit(apk)
	for pack in manifest.get("packs", []):
		if pack is Dictionary:
			await _ensure_pack(pack)
	packs_ready.emit()

func _ensure_pack(pack: Dictionary) -> void:
	var id := str(pack.get("id", ""))
	var version := int(pack.get("version", 0))
	var expected := str(pack.get("sha256", "")).to_lower()
	var size := int(pack.get("size_bytes", 0))
	var path := str(pack.get("url", ""))
	if not PACK_PROBES.has(id) or version <= 0 or not _valid_sha(expected):
		return
	if size <= 0 or size > MAX_PACK_SIZE or not path.begins_with("/api/packs/"):
		return

	var zip_path := "%s/%s_v%d.zip" % [cache_dir, id, version]
	if FileAccess.file_exists(zip_path) and FileAccess.get_sha256(zip_path) == expected:
		if _mount(id, zip_path):
			return

	progress.emit("Загрузка пакета %s…" % id)
	var http := HTTPRequest.new()
	http.timeout = 12.0
	http.body_size_limit = size
	add_child(http)
	if http.request(api.base_url + path) != OK:
		http.queue_free()
		return
	var result: Array = await http.request_completed
	http.queue_free()
	var body: PackedByteArray = result[3]
	if result[0] != HTTPRequest.RESULT_SUCCESS or result[1] != 200 or body.size() != size:
		progress.emit("Пакет %s недоступен · используются встроенные ассеты" % id)
		return
	var hash := HashingContext.new()
	hash.start(HashingContext.HASH_SHA256)
	hash.update(body)
	if hash.finish().hex_encode() != expected:
		progress.emit("Пакет %s не прошёл проверку · используются встроенные ассеты" % id)
		return

	var file := FileAccess.open(zip_path, FileAccess.WRITE)
	if file == null:
		return
	file.store_buffer(body)
	var saved := file.get_error() == OK
	file.close()
	if saved and _mount(id, zip_path):
		progress.emit("Пакет %s подключён" % id)

func _mount(id: String, path: String) -> bool:
	var archive := ZIPReader.new()
	if archive.open(path) != OK:
		return false
	if not archive.file_exists(PACK_PROBES[id]):
		archive.close()
		return false
	if _mounted.has(id):
		_mounted[id].close()
	_mounted[id] = archive
	return true

static func texture(path: String) -> Texture2D:
	if not path.ends_with(".png"):
		return null
	var packs := ["city_tiles", "catalog_v2"] if path.begins_with("map/") else ["catalog_v2"]
	for id in packs:
		var archive: ZIPReader = _mounted.get(id)
		if archive == null or not archive.file_exists(path):
			continue
		var png := archive.read_file(path)
		var img := Image.new()
		if img.load_png_from_buffer(png) == OK:
			return ImageTexture.create_from_image(img)
	return null

static func _valid_sha(value: String) -> bool:
	return value.length() == 64 and value.is_valid_hex_number()
