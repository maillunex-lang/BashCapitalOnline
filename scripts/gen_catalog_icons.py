#!/usr/bin/env python3
"""
Generate beautiful icons for every catalog item, building and shop.
Uses AI archetype sheets + distinctive procedural silhouettes for full coverage.
"""
from __future__ import annotations

import hashlib
import json
import math
import re
import shutil
from pathlib import Path

from PIL import Image, ImageDraw, ImageEnhance, ImageFilter, ImageFont, ImageOps

ROOT = Path(__file__).resolve().parent.parent
ART = Path("/opt/cursor/artifacts/assets")
DATA = ROOT / "data"
CLIENT_ICONS = ROOT / "client/assets/icons"
WEB_ICONS = ROOT / "web/assets/icons"
CLIENT_UI = ROOT / "client/assets/ui"
WEB_UI = ROOT / "web/assets/ui"
CLIENT_MAP = ROOT / "client/assets/map"
WEB_MAP = ROOT / "web/assets/map"

SIZE = 96

CAT_PALETTE = {
    "food": ((210, 150, 70), (120, 70, 35)),
    "dairy": ((230, 235, 240), (90, 140, 160)),
    "grain": ((220, 185, 90), (140, 100, 40)),
    "veg": ((90, 160, 80), (40, 90, 45)),
    "fruit": ((200, 70, 80), (120, 40, 50)),
    "livestock": ((180, 140, 100), (100, 70, 45)),
    "fiber": ((190, 185, 160), (110, 100, 80)),
    "textile": ((170, 140, 190), (90, 70, 120)),
    "apparel": ((150, 110, 160), (80, 55, 100)),
    "timber": ((150, 110, 70), (80, 55, 35)),
    "home": ((160, 130, 95), (90, 70, 50)),
    "mineral": ((170, 165, 150), (90, 90, 85)),
    "ore": ((120, 90, 70), (60, 45, 35)),
    "metal": ((140, 150, 160), (70, 80, 90)),
    "fuel": ((200, 110, 50), (110, 50, 25)),
    "oil": ((50, 50, 55), (20, 20, 25)),
    "oilseed": ((200, 170, 60), (120, 100, 30)),
    "chem": ((100, 160, 170), (50, 90, 100)),
    "packaging": ((180, 190, 195), (90, 100, 110)),
    "tech": ((70, 130, 160), (35, 70, 95)),
}

TIER_RING = {
    "raw": (120, 160, 110),
    "intermediate": (200, 170, 80),
    "finished": (210, 179, 106),
}


def shade(rgb, f):
    return tuple(max(0, min(255, int(c * f))) for c in rgb)


def hash_color(key: str, sat=0.35, light=0.55):
    h = int(hashlib.md5(key.encode()).hexdigest()[:6], 16)
    # golden-angle-ish hue from hash
    hue = (h % 360) / 360.0
    # HSL-ish to RGB
    c = (1 - abs(2 * light - 1)) * sat
    x = c * (1 - abs((hue * 6) % 2 - 1))
    m = light - c / 2
    i = int(hue * 6) % 6
    rgb = [(c, x, 0), (x, c, 0), (0, c, x), (0, x, c), (x, 0, c), (c, 0, x)][i]
    return tuple(int((v + m) * 255) for v in rgb)


def save_icon(name: str, img: Image.Image, kind: str = "icons"):
    if kind == "ui":
        a, b = CLIENT_UI, WEB_UI
    elif kind == "map":
        a, b = CLIENT_MAP, WEB_MAP
    else:
        a, b = CLIENT_ICONS, WEB_ICONS
    a.mkdir(parents=True, exist_ok=True)
    b.mkdir(parents=True, exist_ok=True)
    path = f"{name}.png"
    img.save(a / path)
    img.save(b / path)


def remove_bg(img: Image.Image, thresh: int = 246) -> Image.Image:
    """Remove sheet background via flood-fill from edges (preserves white products like milk)."""
    img = img.convert("RGBA")
    w, h = img.size
    px = img.load()

    def is_bg(x, y):
        r, g, b, a = px[x, y]
        if a == 0:
            return True
        if r >= thresh and g >= thresh and b >= thresh:
            return True
        # soft cream paper backgrounds on AI sheets
        if r > 232 and g > 225 and b > 205 and abs(r - g) < 28 and abs(g - b) < 35:
            return True
        return False

    visited = [[False] * w for _ in range(h)]
    stack = []
    for x in range(w):
        stack.append((x, 0))
        stack.append((x, h - 1))
    for y in range(h):
        stack.append((0, y))
        stack.append((w - 1, y))
    while stack:
        x, y = stack.pop()
        if x < 0 or y < 0 or x >= w or y >= h or visited[y][x]:
            continue
        visited[y][x] = True
        if not is_bg(x, y):
            continue
        r, g, b, _ = px[x, y]
        px[x, y] = (r, g, b, 0)
        stack.extend(((x + 1, y), (x - 1, y), (x, y + 1), (x, y - 1)))
    return img


def fit_square(crop: Image.Image, size: int = SIZE) -> Image.Image:
    out = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    crop = crop.convert("RGBA")
    bbox = crop.getbbox()
    if bbox:
        crop = crop.crop(bbox)
    scale = min(size / max(1, crop.width), size / max(1, crop.height)) * 0.88
    nw, nh = max(1, int(crop.width * scale)), max(1, int(crop.height * scale))
    crop = crop.resize((nw, nh), Image.Resampling.LANCZOS)
    out.paste(crop, ((size - nw) // 2, (size - nh) // 2), crop)
    return out


def slice_grid(sheet: Path, names: list[str], cols: int, rows: int, size: int = SIZE) -> dict[str, Image.Image]:
    out: dict[str, Image.Image] = {}
    if not sheet.exists():
        print("missing sheet", sheet)
        return out
    img = remove_bg(Image.open(sheet), thresh=248)
    bbox = img.getbbox()
    if not bbox:
        return out
    grid = img.crop(bbox)
    cw, ch = grid.width // cols, grid.height // rows
    for i, name in enumerate(names):
        if i >= cols * rows:
            break
        r, c = divmod(i, cols)
        inset = max(3, cw // 35)
        cell = grid.crop((c * cw + inset, r * ch + inset, (c + 1) * cw - inset, (r + 1) * ch - inset))
        sprite = fit_square(cell, size)
        alpha = sprite.split()[-1]
        if sum(1 for p in alpha.getdata() if p > 20) < 60:
            continue
        out[name] = sprite
    print(f"sliced {len(out)}/{len(names)} from {sheet.name}")
    return out


def plate(size, fill, rim=None):
    img = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    # soft shadow
    d.ellipse([10, 14, size - 6, size - 2], fill=(0, 0, 0, 55))
    # body
    d.ellipse([6, 6, size - 10, size - 10], fill=fill + (255,))
    # inner rim
    rim = rim or shade(fill, 1.15)
    d.ellipse([14, 14, size - 18, size - 18], outline=rim + (200,), width=2)
    # highlight
    d.arc([16, 12, size // 2 + 8, size // 2], 200, 340, fill=(255, 255, 255, 90), width=3)
    return img


def draw_wheat(d, cx, cy, col):
    for i, ang in enumerate((-28, -10, 10, 28)):
        rad = math.radians(ang)
        x2 = cx + math.sin(rad) * 28
        y2 = cy - 30 + abs(ang) * 0.15
        d.line([(cx, cy + 18), (x2, y2)], fill=shade(col, 0.85) + (255,), width=3)
        for k in range(5):
            t = k / 4
            px = cx + (x2 - cx) * t
            py = cy + 18 + (y2 - cy - 18) * t
            d.ellipse([px - 3, py - 4, px + 3, py + 2], fill=shade(col, 1.05 - k * 0.05) + (255,))


def draw_bottle(d, cx, cy, body, liquid=None):
    liquid = liquid or shade(body, 0.9)
    d.rounded_rectangle([cx - 14, cy - 8, cx + 14, cy + 28], radius=6, fill=body + (230,))
    d.rectangle([cx - 6, cy - 26, cx + 6, cy - 6], fill=shade(body, 0.85) + (255,))
    d.rectangle([cx - 8, cy - 30, cx + 8, cy - 24], fill=(210, 200, 180, 255))
    d.rounded_rectangle([cx - 11, cy - 2, cx + 11, cy + 22], radius=4, fill=liquid + (200,))
    d.line([(cx - 8, cy), (cx - 8, cy + 16)], fill=(255, 255, 255, 100), width=2)


def draw_bread(d, cx, cy, col):
    d.ellipse([cx - 28, cy - 8, cx + 28, cy + 22], fill=shade(col, 0.75) + (255,))
    d.ellipse([cx - 26, cy - 18, cx + 26, cy + 10], fill=col + (255,))
    for ox in (-12, 0, 12):
        d.arc([cx + ox - 8, cy - 14, cx + ox + 8, cy], 200, 340, fill=shade(col, 0.6) + (255,), width=2)


def draw_cheese(d, cx, cy, col):
    d.pieslice([cx - 26, cy - 18, cx + 26, cy + 26], 200, 90, fill=col + (255,))
    d.pieslice([cx - 26, cy - 18, cx + 26, cy + 26], 200, 90, outline=shade(col, 0.7) + (255,))
    d.ellipse([cx - 6, cy - 2, cx + 2, cy + 6], fill=shade(col, 0.85) + (255,))
    d.ellipse([cx + 6, cy + 6, cx + 14, cy + 14], fill=shade(col, 0.85) + (255,))


def draw_jar(d, cx, cy, col, lid=(140, 90, 50)):
    d.rounded_rectangle([cx - 16, cy - 6, cx + 16, cy + 26], radius=5, fill=col + (230,))
    d.rectangle([cx - 18, cy - 14, cx + 18, cy - 4], fill=lid + (255,))
    d.ellipse([cx - 10, cy + 2, cx + 10, cy + 16], fill=shade(col, 1.1) + (160,))


def draw_fruit(d, cx, cy, col, leaf=(60, 130, 70)):
    d.ellipse([cx - 18, cy - 12, cx + 18, cy + 22], fill=col + (255,))
    d.ellipse([cx - 10, cy - 8, cx, cy + 4], fill=shade(col, 1.2) + (120,))
    d.line([(cx, cy - 12), (cx + 2, cy - 22)], fill=(90, 60, 40, 255), width=2)
    d.ellipse([cx + 2, cy - 26, cx + 14, cy - 14], fill=leaf + (255,))


def draw_veg(d, cx, cy, kind, col):
    if kind == "potato":
        d.ellipse([cx - 22, cy - 12, cx + 22, cy + 18], fill=col + (255,))
        for ox, oy in ((-8, -2), (6, 4), (0, 8)):
            d.ellipse([cx + ox, cy + oy, cx + ox + 3, cy + oy + 3], fill=shade(col, 0.6) + (255,))
    elif kind == "carrot":
        d.polygon([(cx, cy + 26), (cx - 10, cy - 6), (cx + 10, cy - 6)], fill=col + (255,))
        for i in range(3):
            d.line([(cx - 4 + i * 4, cy - 6), (cx - 6 + i * 5, cy - 22)], fill=(60, 140, 70, 255), width=2)
    else:
        d.ellipse([cx - 16, cy - 10, cx + 16, cy + 18], fill=col + (255,))
        d.ellipse([cx - 6, cy - 20, cx + 10, cy - 6], fill=(60, 130, 70, 255))


def draw_ore(d, cx, cy, col):
    pts = [(cx - 20, cy + 10), (cx - 8, cy - 18), (cx + 14, cy - 14), (cx + 22, cy + 8), (cx + 4, cy + 20), (cx - 16, cy + 18)]
    d.polygon(pts, fill=col + (255,))
    d.polygon([(cx - 4, cy - 10), (cx + 8, cy - 6), (cx + 2, cy + 6)], fill=shade(col, 1.25) + (180,))


def draw_barrel(d, cx, cy, col):
    d.rounded_rectangle([cx - 18, cy - 20, cx + 18, cy + 24], radius=8, fill=col + (255,))
    for y in (-10, 0, 12):
        d.rectangle([cx - 18, cy + y, cx + 18, cy + y + 3], fill=shade(col, 0.7) + (255,))
    d.ellipse([cx - 8, cy - 8, cx + 8, cy + 4], fill=shade(col, 1.2) + (100,))


def draw_canister(d, cx, cy, col):
    d.rounded_rectangle([cx - 18, cy - 10, cx + 18, cy + 26], radius=4, fill=col + (255,))
    d.rectangle([cx - 6, cy - 22, cx + 10, cy - 8], fill=shade(col, 0.85) + (255,))
    d.rectangle([cx + 8, cy - 18, cx + 20, cy - 10], fill=shade(col, 0.7) + (255,))
    d.ellipse([cx - 8, cy + 2, cx + 8, cy + 14], fill=(30, 30, 30, 200))


def draw_planks(d, cx, cy, col):
    for i, oy in enumerate((-10, 0, 10)):
        d.rounded_rectangle([cx - 26 + i * 2, cy + oy - 4, cx + 22 + i * 2, cy + oy + 8], radius=2, fill=shade(col, 1 - i * 0.08) + (255,))
        d.line([(cx - 20 + i * 2, cy + oy + 2), (cx + 16 + i * 2, cy + oy + 2)], fill=shade(col, 0.7) + (180,), width=1)


def draw_fabric(d, cx, cy, col):
    d.polygon([(cx - 24, cy + 8), (cx - 10, cy - 22), (cx + 24, cy - 10), (cx + 12, cy + 22)], fill=col + (255,))
    d.polygon([(cx - 10, cy - 22), (cx + 6, cy - 26), (cx + 24, cy - 10), (cx + 10, cy - 6)], fill=shade(col, 1.15) + (255,))
    d.line([(cx - 8, cy - 4), (cx + 10, cy + 8)], fill=shade(col, 0.7) + (160,), width=1)


def draw_clothing(d, cx, cy, col):
    d.polygon([(cx - 20, cy - 8), (cx - 8, cy - 20), (cx + 8, cy - 20), (cx + 20, cy - 8),
               (cx + 16, cy + 4), (cx + 10, cy + 4), (cx + 10, cy + 22), (cx - 10, cy + 22),
               (cx - 10, cy + 4), (cx - 16, cy + 4)], fill=col + (255,))
    d.arc([cx - 8, cy - 22, cx + 8, cy - 10], 0, 180, fill=shade(col, 0.7) + (255,), width=2)


def draw_furniture(d, cx, cy, col):
    d.rectangle([cx - 20, cy - 4, cx + 20, cy + 6], fill=col + (255,))
    d.rectangle([cx - 18, cy - 18, cx + 18, cy - 4], fill=shade(col, 1.1) + (255,))
    d.rectangle([cx - 16, cy + 6, cx - 10, cy + 24], fill=shade(col, 0.8) + (255,))
    d.rectangle([cx + 10, cy + 6, cx + 16, cy + 24], fill=shade(col, 0.8) + (255,))


def draw_animal(d, cx, cy, col):
    d.ellipse([cx - 20, cy - 6, cx + 16, cy + 22], fill=col + (255,))
    d.ellipse([cx + 8, cy - 18, cx + 28, cy + 2], fill=shade(col, 1.05) + (255,))
    d.ellipse([cx + 20, cy - 14, cx + 24, cy - 10], fill=(30, 30, 30, 255))
    d.rectangle([cx - 14, cy + 18, cx - 8, cy + 28], fill=shade(col, 0.7) + (255,))
    d.rectangle([cx + 4, cy + 18, cx + 10, cy + 28], fill=shade(col, 0.7) + (255,))


def draw_circuit(d, cx, cy, col):
    d.rounded_rectangle([cx - 22, cy - 18, cx + 22, cy + 18], radius=4, fill=col + (255,))
    for y in (-8, 0, 8):
        d.line([(cx - 16, cy + y), (cx + 16, cy + y)], fill=(80, 200, 160, 220), width=2)
    d.rectangle([cx - 6, cy - 6, cx + 6, cy + 6], fill=(40, 40, 50, 255))
    d.rectangle([cx - 18, cy - 14, cx - 10, cy - 6], fill=(200, 160, 60, 255))


def draw_box(d, cx, cy, col):
    d.polygon([(cx - 20, cy), (cx, cy + 12), (cx + 20, cy), (cx, cy - 12)], fill=shade(col, 1.1) + (255,))
    d.polygon([(cx - 20, cy), (cx, cy + 12), (cx, cy + 28), (cx - 20, cy + 16)], fill=col + (255,))
    d.polygon([(cx + 20, cy), (cx, cy + 12), (cx, cy + 28), (cx + 20, cy + 16)], fill=shade(col, 0.75) + (255,))


def draw_beam(d, cx, cy, col):
    d.polygon([(cx - 28, cy + 8), (cx - 20, cy - 8), (cx + 28, cy - 16), (cx + 20, cy + 2)], fill=col + (255,))
    d.polygon([(cx - 28, cy + 8), (cx + 20, cy + 2), (cx + 20, cy + 10), (cx - 28, cy + 16)], fill=shade(col, 0.7) + (255,))


def draw_default(d, cx, cy, col, label):
    d.rounded_rectangle([cx - 22, cy - 18, cx + 22, cy + 22], radius=10, fill=col + (255,))
    d.rounded_rectangle([cx - 16, cy - 12, cx + 16, cy + 16], radius=6, fill=shade(col, 1.15) + (180,))
    try:
        font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 22)
    except Exception:
        font = ImageFont.load_default()
    ch = (label[:1] or "?").upper()
    d.text((cx, cy - 2), ch, fill=(255, 255, 255, 230), font=font, anchor="mm")


def base_id(iid: str) -> str:
    iid = re.sub(r"_(std|prem)$", "", iid)
    iid = re.sub(r"_(sterli|ufa|beloretsk|salavat|oktyabrsky)$", "", iid)
    return iid


def parse_flags(iid: str):
    grade = None
    region = None
    if iid.endswith("_prem"):
        grade = "prem"
    elif iid.endswith("_std"):
        grade = "std"
    for r in ("sterli", "ufa", "beloretsk", "salavat", "oktyabrsky"):
        if f"_{r}" in iid or iid.endswith(r):
            region = r
    return grade, region


def pick_drawer(iid: str, category: str):
    b = base_id(iid)
    if b.startswith(("bread_", "pastry_", "chakchak")) or b in ("bread",):
        return "bread"
    if b.startswith(("bottled_", "juice_", "koumiss", "yogurt")) or "milk" in b and category == "dairy":
        return "bottle"
    if b.startswith(("cheese_", "qurut")):
        return "cheese"
    if b.startswith("honey") or b.startswith("jam_") or b.startswith(("snack_", "canned_", "sausage_")):
        return "jar"
    if category == "fruit" or b in ("apple", "cherry", "pear", "plum", "raspberry", "strawberry", "currant"):
        return "fruit"
    if category == "veg" or b in ("potato", "carrot", "cabbage", "onion", "tomato", "cucumber", "pepper", "sugar_beet"):
        return "veg"
    if category == "grain" or b.startswith("flour_") or b in ("wheat", "rye", "barley", "oats", "buckwheat", "animal_feed"):
        return "wheat"
    if category in ("ore", "mineral") or b.endswith("_ore"):
        return "ore"
    if category in ("fuel", "oil") or b in ("oil", "gasoline", "diesel", "kerosene", "fuel_oil", "gas", "coal"):
        return "barrel" if "fuel_can" not in b else "canister"
    if b.startswith("fuel_can"):
        return "canister"
    if category == "timber" or b.startswith(("planks_", "board_", "wood", "pine", "birch", "linden", "firewood")):
        return "planks"
    if category in ("textile", "fiber") or b.startswith("fabric_"):
        return "fabric"
    if category == "apparel" or b.startswith("clothing") or b in ("boots", "shoes", "hat", "felt_boots"):
        return "clothing"
    if category == "home" or b.startswith("furniture_") or b in ("appliance", "brick", "book", "toy"):
        return "furniture" if "furniture" in b or b == "appliance" else "box"
    if category == "livestock" or b in ("cattle", "pig", "sheep", "goat", "horse", "poultry", "bee"):
        return "animal"
    if category == "tech" or b in ("circuit", "electronics", "wire", "research_notes"):
        return "circuit"
    if category == "metal" or b.startswith(("steel", "goods_", "beam_")) or b in ("copper", "aluminum", "zinc", "gold"):
        return "beam"
    if category == "packaging" or b in ("bottle", "packaging", "paper", "plastic"):
        return "box" if b != "bottle" else "bottle"
    if category == "chem":
        return "jar"
    if category == "dairy":
        return "bottle"
    if category == "oilseed":
        return "fruit"
    return "default"


def render_item_icon(item: dict, archetypes: dict[str, Image.Image]) -> Image.Image:
    iid = item["id"]
    cat = item.get("category", "food")
    tier = item.get("tier", "finished")
    fill, accent = CAT_PALETTE.get(cat, ((140, 150, 150), (70, 80, 80)))
    # subtle per-id hue shift
    jitter = hash_color(base_id(iid), sat=0.12, light=0.5)
    fill = tuple(int(fill[i] * 0.82 + jitter[i] * 0.18) for i in range(3))

    # Prefer AI archetype when mapped
    arch_key = archetype_for(iid, cat)
    if arch_key and arch_key in archetypes:
        base = fit_square(archetypes[arch_key], SIZE)
        img = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
        d0 = ImageDraw.Draw(img)
        d0.ellipse([8, 12, SIZE - 4, SIZE - 2], fill=(0, 0, 0, 45))
        img = Image.alpha_composite(img, base)
        # grade/region tint overlay kept subtle via badges below
    else:
        img = plate(SIZE, shade(fill, 0.65), accent)
        d = ImageDraw.Draw(img)
        cx, cy = SIZE // 2, SIZE // 2 - 2
        kind = pick_drawer(iid, cat)
        col = accent if cat in ("dairy", "mineral", "packaging") else fill
        col = shade(col, 1.05)
        if kind == "bread":
            draw_bread(d, cx, cy, (210, 160, 90))
        elif kind == "bottle":
            liquid = (240, 245, 250) if "milk" in iid else (fill if cat != "dairy" else (245, 248, 250))
            if "juice" in iid:
                liquid = fill
            if "oil" in iid:
                liquid = (210, 180, 60)
            draw_bottle(d, cx, cy, (200, 220, 230), liquid)
        elif kind == "cheese":
            draw_cheese(d, cx, cy, (230, 200, 90))
        elif kind == "jar":
            draw_jar(d, cx, cy, fill)
        elif kind == "fruit":
            draw_fruit(d, cx, cy, fill)
        elif kind == "veg":
            draw_veg(d, cx, cy, "potato" if "potato" in iid else ("carrot" if "carrot" in iid else "veg"), fill)
        elif kind == "wheat":
            draw_wheat(d, cx, cy + 4, (220, 185, 80))
        elif kind == "ore":
            draw_ore(d, cx, cy, fill)
        elif kind == "barrel":
            draw_barrel(d, cx, cy, (45, 45, 50) if cat == "oil" else fill)
        elif kind == "canister":
            draw_canister(d, cx, cy, (200, 120, 50))
        elif kind == "planks":
            draw_planks(d, cx, cy, (150, 110, 70))
        elif kind == "fabric":
            draw_fabric(d, cx, cy, fill)
        elif kind == "clothing":
            draw_clothing(d, cx, cy, fill)
        elif kind == "furniture":
            draw_furniture(d, cx, cy, (150, 110, 70))
        elif kind == "animal":
            draw_animal(d, cx, cy, fill)
        elif kind == "circuit":
            draw_circuit(d, cx, cy, (40, 80, 100))
        elif kind == "box":
            draw_box(d, cx, cy, fill)
        elif kind == "beam":
            draw_beam(d, cx, cy, (150, 155, 165))
        else:
            draw_default(d, cx, cy, fill, item.get("name", iid))

    # tier ring
    d = ImageDraw.Draw(img)
    ring = TIER_RING.get(tier, (180, 180, 180))
    d.ellipse([3, 3, SIZE - 4, SIZE - 4], outline=ring + (210,), width=2)

    grade, region = parse_flags(iid)
    if grade == "prem":
        d.polygon([(SIZE - 22, 8), (SIZE - 8, 8), (SIZE - 8, 22)], fill=(210, 179, 106, 255))
    elif grade == "std":
        d.ellipse([SIZE - 22, 8, SIZE - 8, 22], fill=(160, 175, 190, 255))
    if region:
        d.rounded_rectangle([8, SIZE - 20, 36, SIZE - 8], radius=4, fill=(61, 155, 132, 230))

    img = ImageEnhance.Contrast(img).enhance(1.06)
    img = ImageEnhance.Color(img).enhance(1.05)
    return img


def archetype_for(iid: str, cat: str) -> str | None:
    b = base_id(iid)
    mapping = {
        "wheat": "arch_wheat", "rye": "arch_wheat", "barley": "arch_wheat", "oats": "arch_wheat",
        "buckwheat": "arch_wheat", "flour_wheat": "arch_wheat", "flour_rye": "arch_wheat",
        "bread_wheat": "arch_bread", "bread_rye": "arch_bread", "bread_barley": "arch_bread",
        "bread_oats": "arch_bread", "bread_buckwheat": "arch_bread",
        "milk": "arch_milk", "bottled_milk": "arch_milk", "goat_milk": "arch_milk",
        "sheep_milk": "arch_milk", "bottled_goat_milk": "arch_milk", "bottled_sheep_milk": "arch_milk",
        "cheese_milk": "arch_cheese", "cheese_goat_milk": "arch_cheese", "qurut": "arch_cheese",
        "honey": "arch_honey", "honey_std": "arch_honey", "honey_prem": "arch_honey",
        "apple": "arch_apple", "cherry": "arch_cherry", "pear": "arch_apple", "plum": "arch_cherry",
        "potato": "arch_potato", "carrot": "arch_potato", "cabbage": "arch_potato",
        "sunflower": "arch_sunflower", "rapeseed": "arch_sunflower",
        "wood": "arch_planks", "pine": "arch_planks", "birch": "arch_planks",
        "planks_wood": "arch_planks", "board_wood": "arch_planks", "furniture_wood": "arch_planks",
        "flax": "arch_fabric", "fabric_flax": "arch_fabric", "cotton": "arch_fabric", "fabric_cotton": "arch_fabric",
        "clothing": "arch_clothing", "clothing_winter": "arch_clothing", "clothing_work": "arch_clothing",
        "iron_ore": "arch_iron", "copper_ore": "arch_iron", "gold_ore": "arch_iron", "zinc_ore": "arch_iron",
        "coal": "arch_coal", "oil": "arch_oil", "steel": "arch_steel", "beam_steel": "arch_steel",
        "gasoline": "arch_fuelcan", "fuel_can_gasoline": "arch_fuelcan", "diesel": "arch_fuelcan",
        "fuel_can_diesel": "arch_fuelcan", "plastic": "arch_plastic", "bottle": "arch_glass",
        "packaging": "arch_box", "paper": "arch_box", "wire": "arch_wire", "circuit": "arch_circuit",
        "electronics": "arch_circuit", "leather": "arch_leather", "boots": "arch_boots", "shoes": "arch_boots",
        "cattle": "arch_cattle", "pig": "arch_cattle", "sheep": "arch_cattle",
    }
    if b in mapping:
        return mapping[b]
    # prefix fallbacks
    for prefix, key in (
        ("bread_", "arch_bread"), ("bottled_", "arch_milk"), ("cheese_", "arch_cheese"),
        ("flour_", "arch_wheat"), ("jam_", "arch_honey"), ("juice_", "arch_apple"),
        ("furniture_", "arch_planks"), ("fabric_", "arch_fabric"), ("planks_", "arch_planks"),
        ("board_", "arch_planks"), ("fuel_can_", "arch_fuelcan"), ("goods_", "arch_steel"),
    ):
        if b.startswith(prefix):
            return key
    if cat == "fruit":
        return "arch_apple"
    if cat == "grain":
        return "arch_wheat"
    if cat == "ore":
        return "arch_iron"
    return None


def render_building_icon(bdef: dict, archetypes: dict[str, Image.Image]) -> Image.Image:
    bid = bdef["id"]
    cat = bdef.get("category", "factory")
    ind = bdef.get("industry_tag", "")
    colors = {
        "extract": ((120, 130, 120), (70, 120, 100)),
        "factory": ((150, 155, 165), (55, 130, 125)),
        "logistics": ((140, 145, 150), (90, 140, 150)),
        "shop": ((220, 200, 160), (200, 140, 70)),
    }
    wall, roof = colors.get(cat, colors["factory"])
    if "farm" in bid or "orchard" in bid or "apiary" in bid or "livestock" in bid or "poultry" in bid:
        wall, roof = (200, 170, 130), (140, 70, 55)
    if "oil" in bid or "refinery" in bid or "chem" in bid:
        wall, roof = (100, 105, 115), (60, 125, 120)
    if "steel" in bid or "mine" in bid:
        wall, roof = (110, 115, 120), (70, 80, 85)
    if "dairy" in bid or "bakery" in bid or "food" in bid:
        wall, roof = (230, 225, 215), (90, 150, 150)

    arch_map = {
        "dairy_plant": "bld_arch_dairy", "bakery": "bld_arch_bakery", "bakery_industrial": "bld_arch_bakery",
        "steel_mill": "bld_arch_steel", "refinery": "bld_arch_refinery", "farm": "bld_arch_farm",
        "livestock_farm": "bld_arch_livestock", "furniture_plant": "bld_arch_furniture",
        "textile_mill": "bld_arch_textile", "mine": "bld_arch_mine", "electronics_plant": "bld_arch_electronics",
        "oil_well": "bld_arch_refinery", "chem_plant": "bld_arch_refinery", "food_plant": "bld_arch_bakery",
    }
    key = arch_map.get(bid)
    if key and key in archetypes:
        return archetypes[key].copy()

    img = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    # ground diamond
    cx, cy = SIZE // 2, int(SIZE * 0.68)
    d.polygon([(cx, cy - 22), (cx + 36, cy), (cx, cy + 22), (cx - 36, cy)], fill=(70, 120, 90, 220))
    d.polygon([(cx + 3, cy + 3), (cx + 36, cy), (cx, cy + 22)], fill=(0, 0, 0, 40))
    # isometric box
    bw, bd, bh = 26, 14, 28
    left = [(cx - bw, cy - 8), (cx, cy + bd - 8), (cx, cy + bd - 8 - bh), (cx - bw, cy - 8 - bh)]
    right = [(cx + bw, cy - 8), (cx, cy + bd - 8), (cx, cy + bd - 8 - bh), (cx + bw, cy - 8 - bh)]
    top = [(cx - bw, cy - 8 - bh), (cx, cy - bd - 8 - bh), (cx + bw, cy - 8 - bh), (cx, cy + bd - 8 - bh)]
    d.polygon(left, fill=shade(wall, 0.88) + (255,))
    d.polygon(right, fill=shade(wall, 0.7) + (255,))
    d.polygon(top, fill=roof + (255,))
    # windows
    for i in range(2):
        wx = cx - 14 + i * 12
        d.rectangle([wx, cy - bh + 2, wx + 6, cy - bh + 12], fill=(180, 220, 230, 230))
    # industry accent
    if "oil" in bid or "refinery" in bid:
        d.rectangle([cx + 18, cy - bh - 18, cx + 24, cy - 10], fill=shade(wall, 0.6) + (255,))
    if "mine" in bid:
        d.ellipse([cx - 20, cy - 6, cx + 10, cy + 14], fill=(40, 40, 45, 220))
    if "farm" in bid:
        d.polygon([(cx - 8, cy - bh - 10), (cx + 20, cy - bh), (cx - 8, cy - bh + 4)], fill=(160, 70, 55, 255))
    img = ImageEnhance.Contrast(img).enhance(1.05)
    return img


def render_shop_icon(sdef: dict, archetypes: dict[str, Image.Image]) -> Image.Image:
    sid = sdef["id"]
    shop_arch = {
        "grocery": "shop_arch_grocery", "bakery_shop": "shop_arch_bakery", "dairy_shop": "shop_arch_dairy",
        "furniture_shop": "shop_arch_furniture", "apparel_shop": "shop_arch_apparel",
        "electronics_shop": "shop_arch_electronics", "fuel_station": "shop_arch_fuel",
        "hardware": "shop_arch_hardware", "hypermarket": "shop_arch_hyper", "butcher": "shop_arch_butcher",
        "confectionery_shop": "shop_arch_candy", "pharmacy_adjacent": "shop_arch_pharmacy",
        "toy_shop": "shop_arch_toy", "bookstore": "shop_arch_book", "bashkir_souvenir": "shop_arch_souvenir",
        "department": "shop_arch_hyper", "wholesale_club": "shop_arch_hyper",
        "fashion_boutique": "shop_arch_apparel", "workwear_shop": "shop_arch_apparel",
        "home_decor": "shop_arch_furniture", "appliance_shop": "shop_arch_electronics",
        "auto_parts": "shop_arch_hardware", "beverage_shop": "shop_arch_grocery",
        "building_market": "shop_arch_hardware", "fish_shop": "shop_arch_butcher",
        "organic_shop": "shop_arch_grocery",
    }
    key = shop_arch.get(sid)
    if key and key in archetypes:
        return archetypes[key].copy()

    colors = {
        "grocery": (70, 150, 110), "bakery_shop": (210, 150, 80), "dairy_shop": (180, 210, 220),
        "furniture_shop": (150, 110, 70), "apparel_shop": (140, 100, 160), "electronics_shop": (70, 130, 180),
        "fuel_station": (200, 90, 60), "hardware": (120, 130, 140), "department": (90, 140, 150),
    }
    col = colors.get(sid, hash_color(sid, sat=0.4, light=0.45))
    img = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    d.rounded_rectangle([6, 6, SIZE - 6, SIZE - 6], radius=18, fill=(12, 26, 24, 235))
    d.rounded_rectangle([12, 12, SIZE - 12, SIZE - 12], radius=14, fill=col + (255,))
    d.rounded_rectangle([28, 34, 68, 72], radius=8, fill=(12, 26, 24, 210))
    d.rectangle([34, 26, 62, 36], fill=(210, 179, 106, 255))
    d.rectangle([40, 48, 56, 54], fill=(210, 179, 106, 200))
    return img


def load_archetypes() -> dict[str, Image.Image]:
    archetypes: dict[str, Image.Image] = {}
    agro_names = [
        "arch_wheat", "arch_bread", "arch_milk", "arch_cheese",
        "arch_honey", "arch_apple", "arch_cherry", "arch_potato",
        "arch_sunflower", "arch_planks", "arch_fabric", "arch_clothing",
    ]
    archetypes.update(slice_grid(ART / "items_agro_food_sheet.jpg", agro_names, 4, 3))

    ind_names = [
        "arch_iron", "arch_coal", "arch_oil", "arch_steel",
        "arch_fuelcan", "arch_plastic", "arch_glass", "arch_box",
        "arch_wire", "arch_circuit", "arch_leather", "arch_boots",
    ]
    archetypes.update(slice_grid(ART / "items_industry_sheet.jpg", ind_names, 4, 3))

    # cattle from older resource sheet if present
    res = slice_grid(
        ART / "iso_resources_decor.jpg",
        ["res_wheat", "res_wood", "res_iron_ore", "res_oil", "res_milk",
         "res_cherry", "res_sand", "res_coal", "res_cattle", "res_apple",
         "res_potato", "res_flax", "res_cotton", "res_quartz", "res_gas",
         "res_copper_ore", "res_chemicals", "res_honey"],
        6, 3, size=SIZE,
    )
    if "res_cattle" in res:
        archetypes["arch_cattle"] = res["res_cattle"]
    # also alias useful res_* into arch when missing
    aliases = {
        "res_wheat": "arch_wheat", "res_wood": "arch_planks", "res_iron_ore": "arch_iron",
        "res_oil": "arch_oil", "res_milk": "arch_milk", "res_cherry": "arch_cherry",
        "res_coal": "arch_coal", "res_apple": "arch_apple", "res_potato": "arch_potato",
        "res_flax": "arch_fabric", "res_honey": "arch_honey",
    }
    for src, dst in aliases.items():
        if dst not in archetypes and src in res:
            archetypes[dst] = res[src]

    bld_names = [
        "bld_arch_dairy", "bld_arch_bakery", "bld_arch_steel", "bld_arch_refinery",
        "bld_arch_farm", "bld_arch_livestock", "bld_arch_furniture", "bld_arch_textile",
        "bld_arch_shop", "bld_arch_fuelshop", "bld_arch_electronics", "bld_arch_mine",
    ]
    archetypes.update(slice_grid(ART / "buildings_icons_sheet.jpg", bld_names, 4, 3))

    shop_names = [
        "shop_arch_grocery", "shop_arch_bakery", "shop_arch_dairy", "shop_arch_furniture", "shop_arch_apparel",
        "shop_arch_electronics", "shop_arch_fuel", "shop_arch_hardware", "shop_arch_hyper",
        "shop_arch_butcher", "shop_arch_candy", "shop_arch_pharmacy", "shop_arch_toy",
        "shop_arch_book", "shop_arch_souvenir",
    ]
    archetypes.update(slice_grid(ART / "shops_icons_sheet.jpg", shop_names, 5, 3))
    return archetypes


def write_atlas(icons: dict[str, Image.Image], path: Path, cols: int = 32):
    if not icons:
        return
    ids = sorted(icons.keys())
    rows = math.ceil(len(ids) / cols)
    atlas = Image.new("RGBA", (cols * SIZE, rows * SIZE), (0, 0, 0, 0))
    for i, iid in enumerate(ids):
        r, c = divmod(i, cols)
        atlas.paste(icons[iid], (c * SIZE, r * SIZE), icons[iid])
    atlas.save(path)
    print("atlas", path.name, atlas.size, "cells", len(ids))


def main():
    for p in (CLIENT_ICONS, WEB_ICONS, CLIENT_UI, WEB_UI):
        p.mkdir(parents=True, exist_ok=True)

    items = json.loads((DATA / "items.json").read_text(encoding="utf-8"))
    buildings = json.loads((DATA / "buildings.json").read_text(encoding="utf-8"))
    shops = json.loads((DATA / "shops.json").read_text(encoding="utf-8"))

    archetypes = load_archetypes()
    print("archetypes", len(archetypes))

    item_icons: dict[str, Image.Image] = {}
    for it in items:
        icon = render_item_icon(it, archetypes)
        name = f"item_{it['id']}"
        save_icon(name, icon, "icons")
        it["icon"] = name
        item_icons[it["id"]] = icon
        # also map deposit sprites for raw resources onto map when missing
        if it.get("tier") == "raw":
            res_name = f"res_{it['id']}"
            if not (CLIENT_MAP / f"{res_name}.png").exists():
                save_icon(res_name, fit_square(icon, 96), "map")

    bld_icons: dict[str, Image.Image] = {}
    for b in buildings:
        icon = render_building_icon(b, archetypes)
        name = f"bld_{b['id']}"
        save_icon(name, icon, "icons")
        b["icon"] = name
        bld_icons[b["id"]] = icon

    for s in shops:
        icon = render_shop_icon(s, archetypes)
        name = s.get("icon") or f"shop_{s['id']}"
        if not name.startswith("shop_"):
            name = f"shop_{s['id']}"
        save_icon(name, icon, "ui")
        s["icon"] = name

    # copy style sheets into icons gallery
    for sheet_name in (
        "items_agro_food_sheet.jpg", "items_industry_sheet.jpg",
        "buildings_icons_sheet.jpg", "shops_icons_sheet.jpg",
    ):
        src = ART / sheet_name
        if src.exists():
            shutil.copy(src, CLIENT_ICONS / sheet_name)
            shutil.copy(src, WEB_ICONS / sheet_name)

    write_atlas(item_icons, CLIENT_ICONS / "items_atlas.png", cols=36)
    write_atlas(item_icons, WEB_ICONS / "items_atlas.png", cols=36)
    write_atlas(bld_icons, CLIENT_ICONS / "buildings_atlas.png", cols=12)
    write_atlas(bld_icons, WEB_ICONS / "buildings_atlas.png", cols=12)

    # also keep jpg preview sheets for More tab
    if item_icons:
        preview = Image.new("RGB", (1280, 720), (12, 26, 24))
        ids = list(item_icons.keys())[:96]
        cols = 12
        for i, iid in enumerate(ids):
            r, c = divmod(i, cols)
            thumb = item_icons[iid].resize((96, 96), Image.Resampling.LANCZOS)
            preview.paste(thumb.convert("RGB"), (40 + c * 100, 40 + r * 100), thumb)
        preview.save(CLIENT_ICONS / "items_sheet.jpg", quality=90)
        preview.save(WEB_ICONS / "items_sheet.jpg", quality=90)

    if bld_icons:
        preview = Image.new("RGB", (1280, 720), (12, 26, 24))
        ids = list(bld_icons.keys())[:48]
        cols = 8
        for i, iid in enumerate(ids):
            r, c = divmod(i, cols)
            thumb = bld_icons[iid].resize((96, 96), Image.Resampling.LANCZOS)
            preview.paste(thumb.convert("RGB"), (40 + c * 150, 40 + r * 140), thumb)
        preview.save(CLIENT_ICONS / "buildings_sheet.jpg", quality=90)
        preview.save(WEB_ICONS / "buildings_sheet.jpg", quality=90)

    (DATA / "items.json").write_text(json.dumps(items, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    (DATA / "buildings.json").write_text(json.dumps(buildings, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    (DATA / "shops.json").write_text(json.dumps(shops, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    stats = {
        "items": len(items),
        "buildings": len(buildings),
        "shops": len(shops),
        "archetypes": len(archetypes),
        "item_icons": len(item_icons),
        "building_icons": len(bld_icons),
    }
    (DATA / "icon_stats.json").write_text(json.dumps(stats, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(stats, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
