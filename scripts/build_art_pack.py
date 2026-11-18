#!/usr/bin/env python3
"""Build unified BashCapital art pack from generated sheets + procedural fallbacks."""
from __future__ import annotations

import hashlib
import json
import os
import shutil
import zipfile
from pathlib import Path

from PIL import Image, ImageDraw, ImageEnhance, ImageFilter, ImageOps

ROOT = Path("/workspace")
ART = Path("/opt/cursor/artifacts/assets")
CLIENT_MAP = ROOT / "client/assets/map"
CLIENT_UI = ROOT / "client/assets/ui"
CLIENT_BG = ROOT / "client/assets/bg"
CLIENT_ICONS = ROOT / "client/assets/icons"
WEB_MAP = ROOT / "web/assets/map"
WEB_UI = ROOT / "web/assets/ui"
WEB_BG = ROOT / "web/assets/bg"
WEB_ICONS = ROOT / "web/assets/icons"
PACK_DIR = ROOT / "data/packs"
MANIFEST = ROOT / "data/remote_manifest.json"

for p in [CLIENT_MAP, CLIENT_UI, CLIENT_BG, CLIENT_ICONS, WEB_MAP, WEB_UI, WEB_BG, WEB_ICONS, PACK_DIR]:
    p.mkdir(parents=True, exist_ok=True)

W = H = 160


def shade(rgb, f):
    return tuple(max(0, min(255, int(c * f))) for c in rgb)


def save_both(name: str, img: Image.Image, kind: str = "map"):
    if kind == "map":
        a, b = CLIENT_MAP, WEB_MAP
    elif kind == "ui":
        a, b = CLIENT_UI, WEB_UI
    elif kind == "bg":
        a, b = CLIENT_BG, WEB_BG
    else:
        a, b = CLIENT_ICONS, WEB_ICONS
    path = f"{name}.png" if kind != "bg" and not name.endswith((".jpg", ".png")) else name
    if kind == "bg":
        out_name = name if name.endswith((".jpg", ".png")) else f"{name}.jpg"
        rgb = img.convert("RGB")
        rgb.save(a / out_name, quality=90)
        rgb.save(b / out_name, quality=90)
        return
    if not path.endswith(".png"):
        path = f"{name}.png"
    img.save(a / path)
    img.save(b / path)


def remove_bg(img: Image.Image, thresh: int = 245) -> Image.Image:
    img = img.convert("RGBA")
    px = img.load()
    w, h = img.size
    for y in range(h):
        for x in range(w):
            r, g, b, a = px[x, y]
            if r >= thresh and g >= thresh and b >= thresh:
                px[x, y] = (r, g, b, 0)
            elif r > 230 and g > 230 and b > 230 and abs(r - g) < 12 and abs(g - b) < 12:
                px[x, y] = (r, g, b, max(0, 255 - (r - 200) * 4))
    return img


def connected_boxes(img: Image.Image, min_area: int = 800):
    """Find opaque blobs and return bounding boxes sorted reading-order."""
    alpha = img.split()[-1]
    bw = alpha.point(lambda a: 255 if a > 20 else 0)
    w, h = bw.size
    pix = bw.load()
    visited = [[False] * w for _ in range(h)]
    boxes = []
    for y in range(h):
        for x in range(w):
            if visited[y][x] or pix[x, y] == 0:
                visited[y][x] = True
                continue
            # BFS
            stack = [(x, y)]
            visited[y][x] = True
            minx = maxx = x
            miny = maxy = y
            area = 0
            while stack:
                cx, cy = stack.pop()
                area += 1
                minx = min(minx, cx)
                maxx = max(maxx, cx)
                miny = min(miny, cy)
                maxy = max(maxy, cy)
                for nx, ny in ((cx + 1, cy), (cx - 1, cy), (cx, cy + 1), (cx, cy - 1)):
                    if 0 <= nx < w and 0 <= ny < h and not visited[ny][nx]:
                        visited[ny][nx] = True
                        if pix[nx, ny]:
                            stack.append((nx, ny))
            if area >= min_area:
                # pad
                pad = 4
                boxes.append((max(0, minx - pad), max(0, miny - pad), min(w, maxx + pad + 1), min(h, maxy + pad + 1), area))
    boxes.sort(key=lambda b: (b[1] // 80, b[0]))
    return boxes


def fit_square(crop: Image.Image, size: int = W) -> Image.Image:
    out = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    crop = crop.convert("RGBA")
    # trim empty
    bbox = crop.getbbox()
    if bbox:
        crop = crop.crop(bbox)
    scale = min(size / crop.width, size / crop.height) * 0.92
    nw, nh = max(1, int(crop.width * scale)), max(1, int(crop.height * scale))
    crop = crop.resize((nw, nh), Image.Resampling.LANCZOS)
    out.paste(crop, ((size - nw) // 2, (size - nh) // 2), crop)
    return out


def slice_grid(sheet_path: Path, names: list[str], cols: int, rows: int, size: int = W):
    """Slice a known collage by equal grid cells (more reliable than blob detect on JPEG)."""
    if not sheet_path.exists():
        print("missing", sheet_path)
        return 0
    img = remove_bg(Image.open(sheet_path), thresh=248)
    # content bbox
    bbox = img.getbbox()
    if not bbox:
        return 0
    grid = img.crop(bbox)
    cw, ch = grid.width // cols, grid.height // rows
    n = 0
    for i, name in enumerate(names):
        if i >= cols * rows:
            break
        r, c = divmod(i, cols)
        # slight inset to avoid seams
        inset = max(2, cw // 40)
        cell = grid.crop((c * cw + inset, r * ch + inset, (c + 1) * cw - inset, (r + 1) * ch - inset))
        sprite = fit_square(cell, size)
        # skip nearly empty
        a = sprite.split()[-1]
        if sum(1 for p in a.getdata() if p > 20) < 80:
            continue
        save_both(name, sprite, "map")
        n += 1
    print(f"grid-sliced {n}/{len(names)} from {sheet_path.name} ({cols}x{rows})")
    return n


def slice_named(sheet_path: Path, names: list[str], size: int = W, min_area: int = 1200):
    if not sheet_path.exists():
        print("missing", sheet_path)
        return 0
    img = remove_bg(Image.open(sheet_path))
    boxes = connected_boxes(img, min_area=min_area)
    if len(boxes) < max(3, len(names) // 3):
        # blob detect failed (JPEG bleed) — caller should use slice_grid
        print(f"blob-sliced weak {len(boxes)}/{len(names)} from {sheet_path.name}")
    n = 0
    for i, name in enumerate(names):
        if i >= len(boxes):
            break
        x0, y0, x1, y1, _ = boxes[i]
        crop = img.crop((x0, y0, x1, y1))
        sprite = fit_square(crop, size)
        save_both(name, sprite, "map")
        n += 1
    print(f"sliced {n}/{len(names)} from {sheet_path.name}")
    return n


def diamond(d, cx, cy, tw, th, fill):
    d.polygon([(cx, cy - th), (cx + tw, cy), (cx, cy + th), (cx - tw, cy)], fill=fill)


def box3d(d, cx, cy, bw, bd, bh, wall, roof, accent=None, windows=2):
    left = [(cx - bw, cy), (cx, cy + bd), (cx, cy + bd - bh), (cx - bw, cy - bh)]
    right = [(cx + bw, cy), (cx, cy + bd), (cx, cy + bd - bh), (cx + bw, cy - bh)]
    top = [(cx - bw, cy - bh), (cx, cy - bd - bh), (cx + bw, cy - bh), (cx, cy + bd - bh)]
    # soft shadow
    d.polygon([(cx - bw + 4, cy + 4), (cx + 4, cy + bd + 4), (cx + bw + 4, cy + 4)], fill=(0, 0, 0, 45))
    d.polygon(left, fill=shade(wall, 0.88) + (255,))
    d.polygon(right, fill=shade(wall, 0.72) + (255,))
    d.polygon(top, fill=roof + (255,))
    # roof edge highlight
    d.line([(cx - bw, cy - bh), (cx, cy - bd - bh), (cx + bw, cy - bh)], fill=shade(roof, 1.15) + (200,), width=1)
    if accent:
        for i in range(windows):
            wx = cx - 14 + i * 12
            d.rectangle([wx, cy - bh + 10, wx + 6, cy - bh + 18], fill=accent + (230,))
            d.rectangle([wx + 1, cy - bh + 11, wx + 5, cy - bh + 17], fill=shade(accent, 1.2) + (180,))


def procedural_missing():
    """Richer procedural tiles/buildings for any gaps."""
    zones = {
        "grass": (86, 138, 92),
        "meadow": (104, 148, 98),
        "industrial": (118, 122, 128),
        "center": (78, 128, 118),
        "rural": (118, 150, 90),
        "road": (72, 74, 78),
        "residential": (98, 140, 108),
    }
    for z, col in zones.items():
        img = Image.new("RGBA", (W, H), (0, 0, 0, 0))
        d = ImageDraw.Draw(img)
        cx, cy = W // 2, int(H * 0.58)
        diamond(d, cx + 4, cy + 6, 58, 30, (0, 0, 0, 50))
        # extruded tile sides
        side_h = 10
        d.polygon([(cx - 58, cy), (cx, cy + 30), (cx, cy + 30 + side_h), (cx - 58, cy + side_h)], fill=shade(col, 0.55) + (255,))
        d.polygon([(cx + 58, cy), (cx, cy + 30), (cx, cy + 30 + side_h), (cx + 58, cy + side_h)], fill=shade(col, 0.4) + (255,))
        diamond(d, cx, cy, 58, 30, col + (255,))
        if z == "road":
            diamond(d, cx, cy, 40, 18, shade(col, 1.15) + (255,))
            d.line([(cx - 16, cy), (cx + 16, cy)], fill=(230, 205, 110, 230), width=2)
        elif z in ("grass", "meadow", "rural", "residential"):
            for ox, oy in ((-18, -4), (12, 6), (-6, 10), (20, -8)):
                d.ellipse([cx + ox, cy + oy, cx + ox + 4, cy + oy + 3], fill=shade(col, 1.2) + (180,))
        path = CLIENT_MAP / f"tile_{z}.png"
        if not path.exists() or path.stat().st_size < 5000:
            save_both(f"tile_{z}", img, "map")

    builders = {
        "bld_house": ((214, 196, 170), (55, 130, 125)),
        "bld_apartment": ((170, 176, 182), (70, 120, 125)),
        "bld_shop": ((235, 215, 185), (200, 150, 70)),
        "bld_dairy": ((235, 235, 240), (90, 150, 150)),
        "bld_farm": ((210, 180, 140), (150, 70, 60)),
        "bld_furniture": ((170, 130, 90), (55, 125, 120)),
        "bld_steel": ((110, 115, 120), (70, 80, 85)),
        "bld_bakery": ((235, 215, 195), (170, 90, 70)),
        "bld_oil": ((90, 95, 100), (60, 70, 75)),
        "bld_mine": ((120, 105, 90), (80, 85, 90)),
        "bld_warehouse": ((150, 155, 160), (60, 125, 120)),
        "bld_orchard": ((90, 60, 40), (55, 130, 80)),
        "bld_logging": ((140, 100, 70), (55, 120, 80)),
        "bld_chem": ((120, 140, 160), (70, 130, 140)),
        "bld_refinery": ((110, 110, 120), (60, 125, 120)),
        "bld_electronics": ((90, 110, 130), (60, 140, 160)),
        "bld_textile": ((200, 190, 210), (120, 90, 140)),
        "bld_apparel": ((180, 160, 190), (120, 90, 140)),
        "bld_food": ((240, 230, 210), (100, 140, 90)),
        "bld_mill": ((190, 175, 150), (120, 110, 90)),
        "bld_glass": ((160, 180, 190), (100, 150, 160)),
        "bld_paper": ((210, 210, 205), (90, 130, 130)),
        "bld_plastics": ((140, 170, 180), (70, 140, 150)),
        "bld_board": ((160, 140, 110), (100, 120, 90)),
        "bld_sawmill": ((150, 120, 80), (90, 110, 70)),
        "bld_appliance": ((120, 130, 145), (70, 130, 140)),
        "bld_garage": ((130, 135, 140), (80, 90, 95)),
        "bld_livestock": ((200, 170, 130), (120, 80, 50)),
        "bld_quarry": ((190, 175, 140), (150, 140, 110)),
        "bld_gas": ((170, 200, 220), (90, 130, 140)),
    }
    for name, (wall, roof) in builders.items():
        path = CLIENT_MAP / f"{name}.png"
        # refresh procedural always as secondary; keep sliced if larger/detail exists
        img = Image.new("RGBA", (W, H), (0, 0, 0, 0))
        d = ImageDraw.Draw(img)
        cx, cy = W // 2, int(H * 0.62)
        zone = "industrial"
        if "farm" in name or "orchard" in name or "livestock" in name or "logging" in name:
            zone = "rural"
        elif "house" in name or "shop" in name:
            zone = "residential"
        col = zones.get(zone, zones["industrial"])
        diamond(d, cx + 3, cy + 5, 54, 28, (0, 0, 0, 40))
        diamond(d, cx, cy, 54, 28, col + (255,))
        if name == "bld_oil":
            d.ellipse([cx - 28, cy + 8, cx + 28, cy + 28], fill=(35, 35, 38, 200))
            d.rectangle([cx - 4, cy - 10, cx + 4, cy + 18], fill=(55, 55, 60, 255))
            d.line([(cx, cy - 28), (cx + 28, cy - 42)], fill=(45, 45, 50, 255), width=5)
            d.ellipse([cx + 22, cy - 48, cx + 34, cy - 36], fill=(70, 70, 75, 255))
        elif name == "bld_orchard":
            for ox in (-22, 0, 22):
                d.rectangle([cx + ox - 2, cy - 2, cx + ox + 2, cy + 16], fill=(90, 60, 40, 255))
                d.ellipse([cx + ox - 14, cy - 28, cx + ox + 14, cy], fill=(50, 125, 75, 255))
                d.ellipse([cx + ox - 6, cy - 18, cx + ox - 2, cy - 14], fill=(180, 50, 70, 255))
        elif name == "bld_mine":
            d.ellipse([cx - 36, cy - 4, cx + 36, cy + 28], fill=(110, 90, 70, 255))
            d.polygon([(cx - 18, cy + 8), (cx, cy - 28), (cx + 18, cy + 8)], fill=(90, 90, 95, 255))
            d.rectangle([cx - 10, cy - 2, cx + 10, cy + 16], fill=(30, 30, 35, 255))
        elif name == "bld_gas":
            d.ellipse([cx - 26, cy - 20, cx + 26, cy + 24], fill=(170, 205, 225, 235))
            d.rectangle([cx - 6, cy - 36, cx + 6, cy - 8], fill=(90, 100, 110, 255))
        else:
            box3d(d, cx, cy - 4, 34, 18, 34 if "apartment" in name or "refinery" in name else 26, wall, roof, (180, 220, 235), 3 if "apartment" in name else 2)
            if name == "bld_dairy":
                d.ellipse([cx + 18, cy - 36, cx + 42, cy - 10], fill=(220, 225, 230, 255))
                d.rectangle([cx + 20, cy - 18, cx + 40, cy + 10], fill=(205, 210, 215, 255))
            if name == "bld_shop":
                d.polygon([(cx - 28, cy - 20), (cx + 28, cy - 20), (cx + 24, cy - 12), (cx - 24, cy - 12)], fill=(200, 80, 70, 255))
            if name == "bld_steel" or name == "bld_chem" or name == "bld_refinery":
                d.rectangle([cx + 22, cy - 50, cx + 30, cy - 8], fill=shade(wall, 0.7) + (255,))
                d.ellipse([cx + 18, cy - 58, cx + 34, cy - 46], fill=(180, 180, 185, 160))
        # only overwrite if no richer sliced sprite
        if not path.exists() or Image.open(path).size[0] < 140:
            save_both(name, img, "map")


def import_hero_assets():
    mapping = [
        (ART / "logo_bashcapital_gen.jpg", "logo", "ui"),
        (ART / "panel_header_gen.jpg", "panel_header", "ui"),
        (ART / "city_backdrop_gen.jpg", "city_backdrop", "bg"),
        (ART / "ui_icons_sheet_gen.jpg", "ui_icons_sheet", "icons"),
        (ART / "map_style_reference_gen.jpg", "map_style_reference", "icons"),
        (ART / "buildings_sheet.jpg", "buildings_sheet", "icons"),
        (ART / "items_sheet.jpg", "items_sheet", "icons"),
    ]
    for src, name, kind in mapping:
        if not src.exists():
            continue
        im = Image.open(src)
        if kind == "ui" and name == "logo":
            # square logo mark for HUD
            im = ImageOps.fit(im.convert("RGB"), (512, 512), Image.Resampling.LANCZOS)
            save_both("logo_bashcapital.jpg", im, "bg")  # misuse: write via custom
            im.save(CLIENT_UI / "logo_bashcapital.jpg", quality=92)
            im.save(WEB_UI / "logo_bashcapital.jpg", quality=92)
            # also png for godot crispness
            im.convert("RGBA").save(CLIENT_UI / "logo_bashcapital.png")
            im.convert("RGBA").save(WEB_UI / "logo_bashcapital.png")
        elif kind == "ui" and name == "panel_header":
            im = ImageOps.fit(im.convert("RGB"), (1024, 256), Image.Resampling.LANCZOS)
            im.save(CLIENT_UI / "panel_header.jpg", quality=90)
            im.save(WEB_UI / "panel_header.jpg", quality=90)
            im.convert("RGBA").save(CLIENT_UI / "panel_header.png")
            im.convert("RGBA").save(WEB_UI / "panel_header.png")
        elif kind == "bg":
            im = ImageOps.fit(im.convert("RGB"), (1600, 900), Image.Resampling.LANCZOS)
            im.save(CLIENT_BG / "city_backdrop.jpg", quality=90)
            im.save(WEB_BG / "city_backdrop.jpg", quality=90)
        else:
            im.convert("RGB").save(CLIENT_ICONS / f"{name}.jpg", quality=90)
            im.convert("RGB").save(WEB_ICONS / f"{name}.jpg", quality=90)
        print("imported", src.name)


def slice_ui_icons():
    sheet = ART / "ui_icons_sheet_gen.jpg"
    if not sheet.exists():
        return
    img = remove_bg(Image.open(sheet), thresh=250)
    # 2x4 or 3x3 grid assumption — crop equal cells from content bbox
    bbox = img.getbbox()
    if not bbox:
        return
    grid = img.crop(bbox)
    # detect 2 rows x 4 cols roughly
    names = [
        "icon_factory", "icon_shop", "icon_market", "icon_firm",
        "icon_map", "icon_rank", "icon_settings", "icon_fuel",
    ]
    cols, rows = 4, 2
    cw, ch = grid.width // cols, grid.height // rows
    for i, name in enumerate(names):
        r, c = divmod(i, cols) if False else (i // cols, i % cols)
        cell = grid.crop((c * cw, r * ch, (c + 1) * cw, (r + 1) * ch))
        icon = fit_square(cell, 96)
        # boost contrast slightly
        enhancer = ImageEnhance.Contrast(icon)
        icon = enhancer.enhance(1.08)
        save_both(name, icon, "ui")
    # shop type icons from same palette procedural badges
    shop_colors = {
        "shop_grocery": (70, 150, 110),
        "shop_bakery": (210, 150, 80),
        "shop_furniture": (150, 110, 70),
        "shop_apparel": (140, 100, 160),
        "shop_electronics": (70, 130, 180),
        "shop_fuel": (200, 90, 60),
        "shop_hardware": (120, 130, 140),
        "shop_department": (90, 140, 150),
    }
    for name, col in shop_colors.items():
        img = Image.new("RGBA", (96, 96), (0, 0, 0, 0))
        d = ImageDraw.Draw(img)
        d.rounded_rectangle([8, 8, 88, 88], radius=18, fill=(12, 26, 24, 230))
        d.rounded_rectangle([14, 14, 82, 82], radius=14, fill=col + (255,))
        d.rounded_rectangle([28, 34, 68, 70], radius=8, fill=(12, 26, 24, 200))
        d.rectangle([34, 28, 62, 36], fill=(210, 179, 106, 255))
        save_both(name, img, "ui")
    print("ui icons ready")


def pack_zips():
    packs = {
        "base_art": [CLIENT_UI, CLIENT_BG, CLIENT_ICONS],
        "city_tiles": [CLIENT_MAP],
    }
    manifest_packs = []
    for pack_id, dirs in packs.items():
        zpath = PACK_DIR / f"{pack_id}_v2.zip"
        with zipfile.ZipFile(zpath, "w", zipfile.ZIP_DEFLATED) as zf:
            for d in dirs:
                for f in sorted(d.rglob("*")):
                    if f.is_file() and f.suffix.lower() in {".png", ".jpg", ".jpeg", ".webp"}:
                        zf.write(f, f"{d.name}/{f.name}")
        data = zpath.read_bytes()
        digest = hashlib.sha256(data).hexdigest()
        manifest_packs.append({
            "id": pack_id,
            "version": 2,
            "url": f"/api/packs/{pack_id}_v2.zip",
            "sha256": digest,
            "required": True,
            "size_bytes": len(data),
        })
        print(pack_id, "bytes", len(data), "sha", digest[:12])
    apk = {
        "version_code": 2,
        "version_name": "0.2.0",
        "url": "/api/updates/apk/bashcapital-0.2.0.apk",
        "sha256": "pending-release",
        "changelog": "Art pack v2 · remote assets · distance logistics · manage panels",
    }
    manifest = {
        "app_version": "0.2.0",
        "min_client_version": "0.1.0",
        "apk": apk,
        "packs": manifest_packs,
        "style": {
            "palette": {"bg": "#0c1a18", "accent": "#3d9b84", "gold": "#d2b36a", "text": "#e8f2ee"},
            "map": "isometric teal-roof Bashkortostan industry",
        },
    }
    MANIFEST.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    # sync web copies of map already done via save_both
    print("manifest written")


def main():
    import_hero_assets()
    slice_ui_icons()
    # city buildings from rich sheets (3 rows × 6 / 6 / 2 — pad names)
    city_names = [
        "tile_grass", "tile_rural", "tile_road", "bld_house", "bld_apartment", "bld_shop",
        "bld_dairy", "bld_food", "bld_logging", "bld_mine", "bld_oil", "bld_livestock",
        "bld_warehouse", "bld_quarry",
    ]
    # 14 assets on a ~6-wide collage → use 6x3 grid and take first 14
    slice_grid(ART / "iso_city_tileset.jpg", city_names, cols=6, rows=3, size=160)
    ground_names = ["tile_meadow", "tile_steppe", "tile_industrial", "tile_water", "tile_snow"]
    slice_grid(ART / "iso_ground_tiles.jpg", ground_names, cols=3, rows=2, size=160)
    # map center from meadow/industrial aliases
    if (CLIENT_MAP / "tile_meadow.png").exists():
        shutil.copy(CLIENT_MAP / "tile_meadow.png", CLIENT_MAP / "tile_center.png")
        shutil.copy(CLIENT_MAP / "tile_meadow.png", WEB_MAP / "tile_center.png")
    if (CLIENT_MAP / "tile_industrial.png").exists():
        shutil.copy(CLIENT_MAP / "tile_industrial.png", WEB_MAP / "tile_industrial.png")
    if (CLIENT_MAP / "tile_grass.png").exists():
        shutil.copy(CLIENT_MAP / "tile_grass.png", CLIENT_MAP / "tile_residential.png")
        shutil.copy(CLIENT_MAP / "tile_grass.png", WEB_MAP / "tile_residential.png")
    # resources sheet
    res_names = [
        "res_wheat", "res_wood", "res_iron_ore", "res_oil", "res_milk",
        "res_cherry", "res_sand", "res_coal", "res_cattle", "res_apple",
        "res_potato", "res_flax", "res_cotton", "res_quartz", "res_gas",
        "res_copper_ore", "res_chemicals", "res_honey",
    ]
    slice_grid(ART / "iso_resources_decor.jpg", res_names, cols=6, rows=3, size=96)
    procedural_missing()
    # sync any client map files missing on web
    for f in CLIENT_MAP.glob("*.png"):
        dest = WEB_MAP / f.name
        if not dest.exists() or dest.stat().st_mtime < f.stat().st_mtime:
            shutil.copy(f, dest)
    pack_zips()
    print("art pack build complete")


if __name__ == "__main__":
    main()
