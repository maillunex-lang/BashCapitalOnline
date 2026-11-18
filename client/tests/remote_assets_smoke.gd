extends SceneTree

func _initialize() -> void:
	_run.call_deferred()

func _run() -> void:
	var args := OS.get_cmdline_user_args()
	if args.size() != 1:
		printerr("Pass the URL of the V2 server after --")
		quit(2)
		return
	var api := ApiClient.new()
	root.add_child(api)
	api.configure(args[0])
	var assets := RemoteAssets.new()
	root.add_child(assets)
	assets.cache_dir = "user://smoke_asset_packs"
	assets.setup(api)
	assets.packs_ready.connect(CatalogUI.clear_texture_cache)
	await assets.check_and_fetch()
	if not _check(RemoteAssets.texture("catalog_v2/item_wheat.png") != null, "catalog archive supplies wheat art"):
		return
	if not _check(CatalogUI.texture("map/tile_grass.png") is ImageTexture, "map uses mounted textures"):
		return

	var catalog_pack := assets.cache_dir.path_join("catalog_v2_v1.zip")
	var valid_hash := FileAccess.get_sha256(catalog_pack)
	var writer := FileAccess.open(catalog_pack, FileAccess.WRITE)
	writer.store_string("bad cache")
	writer.close()
	await assets.check_and_fetch()
	if not _check(FileAccess.get_sha256(catalog_pack) == valid_hash, "corrupt cached archive was replaced"):
		return

	var manifest := await api.fetch_manifest()
	var bad_pack: Dictionary = manifest["packs"][0].duplicate()
	bad_pack["version"] = 999
	bad_pack["sha256"] = "0".repeat(64)
	await assets._ensure_pack(bad_pack)
	if not _check(not FileAccess.file_exists(assets.cache_dir.path_join("catalog_v2_v999.zip")), "hash mismatch was rejected"):
		return

	api.configure("http://127.0.0.1:1")
	await assets.check_and_fetch()
	if not _check(RemoteAssets.texture("map/tile_grass.png") == null, "offline fallback unmounts the archive"):
		return
	if not _check(CatalogUI.texture("map/tile_grass.png") != null, "bundled texture remains readable"):
		return
	print("Remote asset smoke: mounted catalog and city art, repaired corrupt cache, rejected bad hash, and fell back to bundled art.")
	quit(0)

func _check(ok: bool, label: String) -> bool:
	if not ok:
		printerr("FAIL: " + label)
		quit(1)
	return ok
