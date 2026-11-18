#!/usr/bin/env python3
from PIL import Image, ImageDraw
import os

OUT = "/workspace/client/assets/map"
WEB = "/workspace/web/assets/map"
os.makedirs(OUT, exist_ok=True)
os.makedirs(WEB, exist_ok=True)
W = H = 128


def shade(rgb, f):
    return tuple(max(0, min(255, int(c * f))) for c in rgb)


def diamond(d, cx, cy, tw, th, fill):
    d.polygon([(cx, cy - th), (cx + tw, cy), (cx, cy + th), (cx - tw, cy)], fill=fill)


def box(d, cx, cy, bw, bd, bh, wall, roof, accent=None):
    left = [(cx - bw, cy), (cx, cy + bd), (cx, cy + bd - bh), (cx - bw, cy - bh)]
    right = [(cx + bw, cy), (cx, cy + bd), (cx, cy + bd - bh), (cx + bw, cy - bh)]
    top = [(cx - bw, cy - bh), (cx, cy - bd - bh), (cx + bw, cy - bh), (cx, cy + bd - bh)]
    d.polygon(left, fill=shade(wall, 0.85) + (255,))
    d.polygon(right, fill=shade(wall, 0.7) + (255,))
    d.polygon(top, fill=roof + (255,))
    if accent:
        for i in range(2):
            d.rectangle(
                [cx - 12 + i * 10, cy - bh + 10, cx - 7 + i * 10, cy - bh + 16],
                fill=accent + (220,),
            )


def base(zone):
    img = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    cx, cy = 64, 72
    cols = {
        "grass": (92, 140, 90),
        "meadow": (110, 150, 95),
        "industrial": (120, 124, 128),
        "center": (86, 130, 120),
        "rural": (120, 148, 88),
        "road": (70, 72, 76),
        "residential": (102, 142, 108),
    }
    top = cols.get(zone, cols["grass"])
    diamond(d, cx + 3, cy + 4, 50, 26, (0, 0, 0, 40))
    diamond(d, cx, cy, 50, 26, top + (255,))
    if zone == "road":
        diamond(d, cx, cy, 36, 16, shade(top, 1.1) + (255,))
        d.line([(cx - 14, cy), (cx + 14, cy)], fill=(230, 210, 120, 230), width=1)
    return img, d, cx, cy


def save(name, img):
    img.save(f"{OUT}/{name}.png")
    img.save(f"{WEB}/{name}.png")


def main():
    for z in ["grass", "meadow", "industrial", "center", "rural", "road", "residential"]:
        img, d, cx, cy = base(z)
        save(f"tile_{z}", img)

    builders = {
        "bld_house": ("residential", lambda d, cx, cy: box(d, cx, cy - 2, 28, 16, 26, (214, 196, 170), (70, 120, 100), (170, 210, 230))),
        "bld_apartment": ("center", lambda d, cx, cy: box(d, cx, cy, 34, 18, 42, (180, 186, 190), (90, 110, 120), (190, 220, 235))),
        "bld_shop": ("residential", lambda d, cx, cy: (
            box(d, cx, cy - 2, 30, 16, 22, (230, 210, 180), (190, 150, 70), (255, 240, 180)),
            d.polygon([(46, 60), (82, 60), (78, 66), (50, 66)], fill=(200, 80, 70, 255)),
        )),
        "bld_dairy": ("industrial", lambda d, cx, cy: (
            box(d, 58, cy, 26, 14, 24, (230, 230, 235), (150, 170, 185), (180, 220, 235)),
            d.ellipse([68, 40, 88, 58], fill=(210, 215, 220, 255)),
            d.rectangle([68, 50, 88, 88], fill=(200, 205, 210, 255)),
            d.ellipse([86, 40, 106, 58], fill=(210, 215, 220, 255)),
            d.rectangle([86, 50, 106, 88], fill=(200, 205, 210, 255)),
        )),
        "bld_glass": ("industrial", lambda d, cx, cy: (
            box(d, cx, cy, 32, 16, 28, (160, 180, 190), (120, 150, 160), (140, 210, 230)),
            d.rectangle([cx + 8, cy - 42, cx + 14, cy - 28], fill=(90, 90, 95, 255)),
        )),
        "bld_furniture": ("industrial", lambda d, cx, cy: box(d, cx, cy, 34, 18, 26, (170, 130, 90), (120, 90, 60), (240, 220, 180))),
        "bld_mine": ("industrial", lambda d, cx, cy: (
            d.ellipse([30, 70, 98, 100], fill=(110, 90, 70, 255)),
            d.polygon([(44, 78), (64, 50), (84, 78)], fill=(90, 90, 95, 255)),
            d.rectangle([54, 70, 74, 88], fill=(40, 40, 45, 255)),
        )),
        "bld_oil": ("industrial", lambda d, cx, cy: (
            d.rectangle([58, 70, 70, 92], fill=(50, 50, 55, 255)),
            d.line([(64, 55), (64, 70)], fill=(40, 40, 45, 255), width=4),
            d.line([(64, 55), (88, 42)], fill=(40, 40, 45, 255), width=4),
            d.ellipse([40, 84, 88, 104], fill=(30, 30, 30, 180)),
        )),
        "bld_logging": ("rural", lambda d, cx, cy: (
            box(d, 54, cy + 4, 18, 12, 16, (140, 100, 70), (90, 70, 50)),
            d.ellipse([70, 80, 110, 92], fill=(120, 85, 55, 255)),
            d.rectangle([34, 60, 38, 88], fill=(90, 60, 40, 255)),
            d.ellipse([22, 40, 50, 70], fill=(50, 110, 70, 255)),
        )),
        "bld_quarry": ("industrial", lambda d, cx, cy: (
            d.polygon([(30, 70), (64, 50), (98, 70), (64, 95)], fill=(190, 175, 140, 255)),
            d.polygon([(40, 72), (64, 58), (88, 72), (64, 86)], fill=(160, 145, 115, 255)),
        )),
        "bld_warehouse": ("industrial", lambda d, cx, cy: box(d, cx, cy + 2, 38, 18, 22, (150, 155, 160), (100, 110, 120))),
        "bld_textile": ("industrial", lambda d, cx, cy: box(d, cx, cy, 30, 16, 24, (200, 190, 210), (130, 110, 140), (230, 220, 240))),
        "bld_steel": ("industrial", lambda d, cx, cy: (
            box(d, 54, cy + 2, 24, 14, 30, (100, 105, 110), (70, 75, 80)),
            d.rectangle([78, 70, 100, 96], fill=(90, 70, 60, 255)),
        )),
        "bld_farm": ("rural", lambda d, cx, cy: (
            box(d, 50, cy + 4, 18, 12, 14, (210, 180, 140), (150, 70, 60)),
            d.polygon([(70, 78), (100, 64), (110, 78), (80, 92)], fill=(160, 70, 60, 255)),
        )),
        "bld_orchard": ("meadow", lambda d, cx, cy: (
            d.rectangle([38, 70, 42, 92], fill=(90, 60, 40, 255)),
            d.ellipse([26, 48, 54, 78], fill=(55, 120, 70, 255)),
            d.rectangle([62, 70, 66, 92], fill=(90, 60, 40, 255)),
            d.ellipse([50, 48, 78, 78], fill=(55, 120, 70, 255)),
            d.rectangle([86, 70, 90, 92], fill=(90, 60, 40, 255)),
            d.ellipse([74, 48, 102, 78], fill=(55, 120, 70, 255)),
        )),
        "bld_chem": ("industrial", lambda d, cx, cy: (
            box(d, 58, cy, 24, 14, 26, (120, 140, 160), (80, 100, 120), (120, 200, 220)),
            d.ellipse([78, 48, 102, 72], fill=(70, 130, 160, 255)),
        )),
        "bld_board": ("industrial", lambda d, cx, cy: box(d, cx, cy, 34, 16, 22, (160, 140, 110), (120, 100, 80))),
        "bld_garage": ("industrial", lambda d, cx, cy: (
            box(d, cx, cy + 2, 36, 16, 18, (130, 135, 140), (90, 95, 100)),
            d.rectangle([48, 72, 80, 88], fill=(40, 40, 45, 255)),
        )),
        "bld_bakery": ("industrial", lambda d, cx, cy: box(d, cx, cy, 28, 16, 24, (230, 210, 190), (170, 90, 70), (255, 230, 200))),
        "bld_apparel": ("industrial", lambda d, cx, cy: box(d, cx, cy, 30, 16, 26, (180, 160, 190), (120, 90, 140), (230, 210, 240))),
        "bld_electronics": ("industrial", lambda d, cx, cy: box(d, cx, cy, 32, 16, 28, (90, 110, 130), (60, 80, 100), (120, 200, 255))),
        "bld_food": ("industrial", lambda d, cx, cy: box(d, cx, cy, 30, 16, 24, (240, 230, 210), (100, 140, 90), (200, 230, 180))),
        "bld_plastics": ("industrial", lambda d, cx, cy: box(d, cx, cy, 32, 16, 26, (140, 170, 180), (90, 130, 140), (160, 220, 230))),
        "bld_refinery": ("industrial", lambda d, cx, cy: (
            box(d, 50, cy, 22, 12, 34, (110, 110, 120), (70, 70, 80)),
            d.ellipse([78, 40, 108, 70], fill=(80, 90, 100, 255)),
            d.rectangle([84, 65, 102, 96], fill=(90, 95, 105, 255)),
        )),
        "bld_mill": ("industrial", lambda d, cx, cy: box(d, cx, cy, 30, 16, 30, (190, 175, 150), (130, 110, 90), (230, 220, 200))),
        "bld_livestock": ("rural", lambda d, cx, cy: (
            box(d, 50, cy + 4, 20, 12, 14, (200, 170, 130), (120, 80, 50)),
            d.ellipse([78, 78, 98, 94], fill=(170, 140, 100, 255)),
        )),
        "bld_sawmill": ("industrial", lambda d, cx, cy: (
            box(d, cx, cy, 30, 16, 22, (150, 120, 80), (100, 80, 55)),
            d.ellipse([78, 78, 110, 96], fill=(120, 90, 55, 255)),
        )),
        "bld_paper": ("industrial", lambda d, cx, cy: box(d, cx, cy, 34, 16, 26, (210, 210, 205), (140, 145, 150), (250, 250, 245))),
        "bld_appliance": ("industrial", lambda d, cx, cy: box(d, cx, cy, 32, 16, 28, (120, 130, 145), (80, 90, 105), (180, 210, 230))),
        "bld_gas": ("industrial", lambda d, cx, cy: (
            d.ellipse([40, 55, 88, 100], fill=(170, 200, 220, 230)),
            d.rectangle([58, 48, 70, 70], fill=(90, 100, 110, 255)),
        )),
    }

    for name, (zone, fn) in builders.items():
        img, d, cx, cy = base(zone)
        fn(d, cx, cy)
        save(name, img)

    piles = {
        "iron_ore": ((110, 90, 70), (140, 110, 90)),
        "oil": ((30, 30, 30), (60, 60, 60)),
        "sand": ((200, 180, 130), (230, 210, 160)),
        "quartz": ((160, 190, 210), (220, 235, 245)),
        "wood": ((100, 70, 45), (130, 95, 60)),
        "flax": ((100, 140, 70), (140, 180, 90)),
        "cherry": ((140, 40, 70), (180, 70, 100)),
        "milk": ((230, 225, 200), (250, 248, 230)),
        "chemicals": ((70, 110, 150), (100, 150, 190)),
        "wheat": ((210, 180, 90), (230, 200, 120)),
        "cattle": ((150, 120, 90), (180, 150, 120)),
        "coal": ((40, 40, 45), (70, 70, 75)),
        "cotton": ((230, 230, 220), (250, 250, 245)),
        "apple": ((180, 50, 50), (80, 140, 70)),
        "potato": ((160, 120, 70), (190, 150, 100)),
        "gas": ((180, 210, 230), (220, 235, 245)),
        "copper_ore": ((160, 90, 50), (190, 120, 70)),
    }
    for name, (c1, c2) in piles.items():
        img = Image.new("RGBA", (64, 64), (0, 0, 0, 0))
        d = ImageDraw.Draw(img)
        d.ellipse([10, 28, 54, 54], fill=shade(c1, 0.7) + (200,))
        d.ellipse([16, 18, 48, 46], fill=c1 + (255,))
        d.ellipse([22, 14, 42, 34], fill=c2 + (255,))
        save(f"res_{name}", img)

    print("sprites:", len(os.listdir(OUT)))


if __name__ == "__main__":
    main()
