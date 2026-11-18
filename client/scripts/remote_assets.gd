extends Node
class_name RemoteAssets

signal packs_ready
signal update_available(info: Dictionary)
signal progress(msg: String)

var api: ApiClient
var manifest: Dictionary = {}
var cache_dir: String = "user://asset_packs"

func setup(client: ApiClient) -> void:
	api = client
	DirAccess.make_dir_recursive_absolute(cache_dir)

func check_and_fetch() -> void:
	if api == null:
		return
	progress.emit("Проверка манифеста ассетов…")
	var res: Dictionary = await api.fetch_manifest()
	if not res.is_empty() and not res.has("error"):
		manifest = res
	else:
		progress.emit("Манифест недоступен — локальные ассеты")
		packs_ready.emit()
		return
	var apk: Dictionary = manifest.get("apk", {})
	if int(apk.get("version_code", 0)) > _local_apk_code():
		update_available.emit(apk)
	for pack in manifest.get("packs", []):
		await _ensure_pack(pack)
	packs_ready.emit()

func _local_apk_code() -> int:
	# Godot export sets this via feature tag / config later; MVP constant.
	return 1

func _ensure_pack(pack: Dictionary) -> void:
	var id := str(pack.get("id", ""))
	var ver := int(pack.get("version", 1))
	var marker := "%s/%s_v%d.ok" % [cache_dir, id, ver]
	if FileAccess.file_exists(marker):
		return
	progress.emit("Загрузка пака %s…" % id)
	var url_path := str(pack.get("url", ""))
	if url_path == "":
		return
	# Relative pack URLs are served by game server.
	var http := HTTPRequest.new()
	add_child(http)
	var err := http.request(api.base_url + url_path)
	if err != OK:
		http.queue_free()
		return
	var result = await http.request_completed
	http.queue_free()
	var code: int = result[1]
	var body: PackedByteArray = result[3]
	if code < 200 or code >= 300:
		progress.emit("Пак %s не скачан (%s)" % [id, code])
		return
	var zip_path := "%s/%s_v%d.zip" % [cache_dir, id, ver]
	var f := FileAccess.open(zip_path, FileAccess.WRITE)
	if f:
		f.store_buffer(body)
		f.close()
	# Mark done; full unzip into user:// can be added when CDN is live.
	var ok := FileAccess.open(marker, FileAccess.WRITE)
	if ok:
		ok.store_string(str(pack.get("sha256", "")))
		ok.close()
	progress.emit("Пак %s готов" % id)
