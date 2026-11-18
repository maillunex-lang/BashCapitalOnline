"""Run a Chromium map smoke against the serving web UI and a fresh V2 server."""
import argparse
import shutil

from playwright.sync_api import sync_playwright

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--web", default="http://127.0.0.1:8090")
parser.add_argument("--api", default="http://127.0.0.1:8080")
args = parser.parse_args()

with sync_playwright() as playwright:
    browser = playwright.chromium.launch(executable_path=shutil.which("chromium"), args=["--no-sandbox"])
    page = browser.new_page(viewport={"width": 1280, "height": 900})
    errors = []
    page.on("pageerror", lambda error: errors.append(str(error)))
    page.route("http://127.0.0.1:8080/**", lambda route: route.continue_(
        url=route.request.url.replace("http://127.0.0.1:8080", args.api.rstrip("/"), 1)
    ))
    page.goto(args.web.rstrip("/") + "/?tab=map&city=ufa")
    page.wait_for_function("state.world?.cities?.ufa && paintedMap?.tiles.size === 41")
    page.wait_for_load_state("networkidle")
    page.evaluate("() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)))")
    evidence = page.evaluate("""() => {
      const original = paintMap;
      const counts = [];
      paintMap = (ctx, w, h, city, tiles, ...args) => {
        const rows = [...tiles];
        counts.push(rows.length);
        original(ctx, w, h, city, rows, ...args);
      };
      function check(condition, message) {
        if (!condition) throw new Error(message);
      }
      function verifyRepaint(label) {
        counts.length = 0;
        drawMap();
        check(counts.length === 1 && counts[0] > 0 && counts[0] < 41, label + ": expected a partial draw: " + counts);
        const patches = counts[0];
        const canvas = $("map"), ctx = canvas.getContext("2d");
        const incremental = ctx.getImageData(0, 0, canvas.width, canvas.height).data;
        paintedMap = null;
        drawMap();
        const full = ctx.getImageData(0, 0, canvas.width, canvas.height).data;
        // Chromium can round antialiased channels differently near a clip edge.
        let total = 0, max = 0;
        for (let i = 0; i < full.length; i++) {
          const delta = Math.abs(full[i] - incremental[i]);
          if (delta) {
            total++; max = Math.max(max, delta);
          }
        }
        check(max <= 24 && total <= full.length * 0.001, label + ": pixel difference " + JSON.stringify({total, max}));
        return patches;
      }
      counts.length = 0;
      drawMap();
      check(counts.length === 0, "unchanged snapshot should not paint");
      const city = state.world.cities.ufa;
      const selected = city.parcels.find(p => p.x === 0 && p.y === 0);
      state.selected = selected;
      const selection = verifyRepaint("selection");
      selected.has_road = !selected.has_road;
      const road = verifyRepaint("road");
      const distance = (b) => Math.abs(b.x) + Math.abs(b.y);
      const occupied = Object.values(state.world.buildings).filter(b => b.city_id === "ufa")
        .sort((a, b) => distance(b) - distance(a))[0];
      delete state.world.buildings[occupied.id];
      const removal = verifyRepaint("overlapping building removal");

      counts.length = 0;
      state.city = "sterlitamak";
      state.selected = null;
      drawMap();
      check(counts[0] === 25, "switching cities should paint the new city");
      state.city = "beloretsk";
      drawMap();
      check(counts[1] === 25, "Beloretsk should paint its own parcels");
      return { selection, road, removal, citySizes: counts };
    }""")
    assert not errors, errors
    print(evidence)
    browser.close()
