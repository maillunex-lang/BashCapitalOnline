const state = {
  base: "http://127.0.0.1:8080",
  firmId: localStorage.getItem("bc_firm") || "",
  city: "ufa",
  world: null,
  catalog: null,
  selected: null,
  quality: 2,
};

const $ = (id) => document.getElementById(id);

const builds = [
  "farm","orchard","dairy_plant","glass_plant","quarry","textile_mill","board_plant",
  "furniture_plant","mine","steel_mill","chem_plant","logging_camp","warehouse","garage"
];
builds.forEach((b) => {
  const o = document.createElement("option");
  o.value = b; o.textContent = b;
  $("buildType").appendChild(o);
});

async function api(path, opts = {}) {
  const headers = { "Content-Type": "application/json", ...(opts.headers || {}) };
  if (state.firmId) headers["X-Firm-Id"] = state.firmId;
  const res = await fetch(state.base + path, { ...opts, headers });
  const text = await res.text();
  let data;
  try { data = JSON.parse(text); } catch { data = { error: text }; }
  if (!res.ok) throw new Error(data.error || text || res.statusText);
  return data;
}

function money(v) { return Math.round(Number(v || 0)).toLocaleString("ru-RU"); }

async function refresh() {
  state.base = $("server").value.replace(/\/$/, "");
  state.world = await api("/api/world");
  if (!state.catalog) state.catalog = await api("/api/catalog");
  const w = state.world;
  $("tick").textContent = w.tick;
  $("date").textContent = `${String(w.month).padStart(2,"0")}.${w.year}`;
  if (state.firmId && w.firms[state.firmId]) {
    const f = w.firms[state.firmId];
    $("cash").textContent = `${money(f.cash)} ₽${f.subscription_active ? " · PRO" : ""}`;
    $("subStatus").textContent = f.subscription_active
      ? `Подписка до тика ${f.subscription_until_tick}${f.deputy?.hired ? " · заместитель «" + f.deputy.name + "»" : ""}`
      : "Подписка не активна — заместитель недоступен";
  }
  $("cityPill").textContent = state.city === "ufa" ? "Уфа" : "Белорецк";
  $("log").innerHTML = (w.last_tick_log || []).map((l) => `<div>${l}</div>`).join("") || "<div>Пока тихо</div>";
  const ranks = await api("/api/rankings");
  ranks.sort((a,b) => b.capital - a.capital);
  $("rank").innerHTML = ranks.map((r,i) =>
    `<div>${i+1}. ${r.is_ai ? "AI" : "●"} <b style="color:#e8f2ee">${r.name}</b> — ${money(r.capital)} ₽</div>`
  ).join("");
  drawMap();
}

function iso(x, y) {
  const tw = 54, th = 27;
  return [(x - y) * tw * 0.5, (x + y) * th * 0.5];
}

function drawMap() {
  const canvas = $("map");
  const ctx = canvas.getContext("2d");
  const w = canvas.width, h = canvas.height;
  ctx.clearRect(0, 0, w, h);
  const g = ctx.createLinearGradient(0, 0, 0, h);
  g.addColorStop(0, "rgba(45,90,80,.35)");
  g.addColorStop(1, "rgba(20,35,32,.55)");
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, w, h);

  const city = state.world?.cities?.[state.city];
  if (!city) return;
  const ox = w / 2, oy = h * 0.42;
  const colors = {
    center: "#2F6F6A", residential: "#4A7C59", industrial: "#6B7280",
    rural: "#7FAE6E", empty: "#3A4A46", commercial: "#C4A35A",
  };
  const parcels = [...city.parcels].sort((a,b) => (a.x+a.y)-(b.x+b.y));
  const buildings = Object.values(state.world.buildings || {}).filter((b) => b.city_id === state.city);
  const shops = Object.values(state.world.shops || {}).filter((s) => s.city_id === state.city);
  const deposits = city.deposits || [];
  const depColor = {
    oil: "#2C2C2C", iron_ore: "#6E5A4B", wood: "#3E6B3A", sand: "#C2B280",
    quartz: "#A8C0D0", milk: "#E8E4D0", cherry: "#8B1E3F", flax: "#6B8F3C", chemicals: "#4A6FA5",
  };

  // ambience by quality
  const q = Number(state.quality || 2);
  if (q >= 1) {
    const n = q >= 3 ? 22 : q >= 2 ? 12 : 6;
    for (let i = 0; i < n; i++) {
      const tx = ox + ((i * 97) % 700) - 350;
      const ty = oy + ((i * 53) % 320) - 140;
      const lean = q >= 3 ? Math.sin(Date.now() / 700 + i) * 3 : 0;
      ctx.strokeStyle = "#5C4033";
      ctx.lineWidth = 3;
      ctx.beginPath();
      ctx.moveTo(tx, ty);
      ctx.lineTo(tx + lean, ty - 22);
      ctx.stroke();
      ctx.fillStyle = "#3F6B4A";
      ctx.beginPath();
      ctx.arc(tx + lean, ty - 28, 11, 0, Math.PI * 2);
      ctx.fill();
    }
  }
  if (q >= 2) {
    const cars = q >= 3 ? 10 : 4;
    for (let i = 0; i < cars; i++) {
      const t = (Date.now() / 30 + i * 80) % 900;
      const cx = ox - 450 + t;
      const cy = oy + 90 + (i % 3) * 14;
      ctx.fillStyle = i % 2 ? "#c4a35a" : "#6aa7c4";
      ctx.fillRect(cx, cy, 18, 7);
    }
  }

  for (const p of parcels) {
    const [ix, iy] = iso(p.x, p.y);
    const cx = ox + ix, cy = oy + iy;
    const selected = state.selected && state.selected.x === p.x && state.selected.y === p.y;
    drawTile(ctx, cx, cy, colors[p.zone] || colors.empty, selected);
    const dep = deposits.find((d) => d.x === p.x && d.y === p.y && d.amount > 0);
    if (dep) {
      ctx.fillStyle = depColor[dep.item_id] || "#8B7355";
      ctx.beginPath();
      ctx.arc(cx, cy + 2, 6, 0, Math.PI * 2);
      ctx.fill();
      ctx.strokeStyle = "rgba(255,255,255,.5)";
      ctx.beginPath();
      ctx.arc(cx, cy + 2, 6, -Math.PI/2, -Math.PI/2 + Math.PI * 2 * (dep.amount / Math.max(1, dep.max_amount)));
      ctx.stroke();
    }
    if (p.has_road) {
      ctx.fillStyle = "rgba(70,72,76,.85)";
      ctx.beginPath();
      ctx.moveTo(cx, cy - 5);
      ctx.lineTo(cx + 16, cy);
      ctx.lineTo(cx, cy + 5);
      ctx.lineTo(cx - 16, cy);
      ctx.closePath();
      ctx.fill();
      ctx.strokeStyle = "rgba(210,180,80,.7)";
      ctx.beginPath();
      ctx.moveTo(cx - 8, cy);
      ctx.lineTo(cx + 8, cy);
      ctx.stroke();
    }
    if (buildings.find((b) => b.x === p.x && b.y === p.y)) drawBuilding(ctx, cx, cy, p.zone === "industrial");
    else if ((p.retail_slots || 0) > 0) {
      const shop = shops.find((s) => s.x === p.x && s.y === p.y);
      const rented = !!shop;
      ctx.beginPath();
      ctx.arc(cx, cy - 6, 5, 0, Math.PI * 2);
      ctx.fillStyle = rented ? "#E2B14B" : "#F5E6C8";
      ctx.fill();
      if (selected && shop) {
        const r = shop.radius || 2;
        for (const q of city.parcels) {
          if (Math.abs(q.x - p.x) + Math.abs(q.y - p.y) <= r) {
            const [qx, qy] = iso(q.x, q.y);
            ctx.fillStyle = "rgba(226,177,75,.15)";
            ctx.beginPath();
            ctx.arc(ox + qx, oy + qy, 5, 0, Math.PI * 2);
            ctx.fill();
          }
        }
      }
    }
  }
  if (q >= 3) {
    // fake ultra vignette
    const vg = ctx.createRadialGradient(ox, oy, 80, ox, oy, 420);
    vg.addColorStop(0, "rgba(0,0,0,0)");
    vg.addColorStop(1, "rgba(0,0,0,0.28)");
    ctx.fillStyle = vg;
    ctx.fillRect(0, 0, w, h);
  }
}

function drawTile(ctx, cx, cy, color, selected) {
  const tw = 54, th = 27;
  ctx.beginPath();
  ctx.moveTo(cx, cy - th/2);
  ctx.lineTo(cx + tw/2, cy);
  ctx.lineTo(cx, cy + th/2);
  ctx.lineTo(cx - tw/2, cy);
  ctx.closePath();
  ctx.fillStyle = selected ? shade(color, 1.25) : color;
  ctx.fill();
  ctx.strokeStyle = "rgba(255,255,255,.12)";
  ctx.stroke();
}

function drawBuilding(ctx, cx, cy, tall) {
  const h = tall ? 28 : 16;
  ctx.fillStyle = tall ? "#9AA3AD" : "#D9C4A0";
  ctx.beginPath();
  ctx.moveTo(cx - 12, cy);
  ctx.lineTo(cx, cy + 8);
  ctx.lineTo(cx, cy + 8 - h);
  ctx.lineTo(cx - 12, cy - h);
  ctx.closePath();
  ctx.fill();
  ctx.fillStyle = tall ? "#7E8792" : "#C4AE8A";
  ctx.beginPath();
  ctx.moveTo(cx + 12, cy);
  ctx.lineTo(cx, cy + 8);
  ctx.lineTo(cx, cy + 8 - h);
  ctx.lineTo(cx + 12, cy - h);
  ctx.closePath();
  ctx.fill();
  ctx.fillStyle = "#E8D7B8";
  ctx.beginPath();
  ctx.moveTo(cx - 12, cy - h);
  ctx.lineTo(cx, cy - h + 8);
  ctx.lineTo(cx + 12, cy - h);
  ctx.lineTo(cx, cy - h - 8);
  ctx.closePath();
  ctx.fill();
}

function shade(hex, f) {
  const n = parseInt(hex.slice(1), 16);
  let r = Math.min(255, ((n >> 16) & 255) * f);
  let g = Math.min(255, ((n >> 8) & 255) * f);
  let b = Math.min(255, (n & 255) * f);
  return `rgb(${r|0},${g|0},${b|0})`;
}

function pickTile(mx, my) {
  const city = state.world?.cities?.[state.city];
  if (!city) return null;
  const ox = $("map").width / 2, oy = $("map").height * 0.42;
  const tw = 54, th = 27;
  const lx = mx - ox, ly = my - oy;
  const a = lx / (tw * 0.5), b = ly / (th * 0.5);
  const x = Math.round((a + b) * 0.5);
  const y = Math.round((b - a) * 0.5);
  return city.parcels.find((p) => p.x === x && p.y === y) || null;
}

$("map").addEventListener("click", (e) => {
  const rect = $("map").getBoundingClientRect();
  const scaleX = $("map").width / rect.width;
  const scaleY = $("map").height / rect.height;
  const p = pickTile((e.clientX - rect.left) * scaleX, (e.clientY - rect.top) * scaleY);
  if (!p) return;
  state.selected = p;
  const dep = (state.world?.cities?.[state.city]?.deposits || []).find((d) => d.x === p.x && d.y === p.y && d.amount > 0);
  $("tileInfo").textContent = `Участок (${p.x},${p.y}) · зона ${p.zone} · дорога ${p.has_road ? "да" : "нет"} · слоты ${p.retail_slots || 0}`
    + (dep ? ` · залежь ${dep.item_id} ${dep.amount}/${dep.max_amount}` : "");
  drawMap();
});

$("register").onclick = async () => {
  try {
    state.base = $("server").value.replace(/\/$/, "");
    const firm = await api("/api/register", {
      method: "POST",
      body: JSON.stringify({ name: $("name").value, industry: $("industry").value }),
    });
    state.firmId = firm.id;
    localStorage.setItem("bc_firm", firm.id);
    state.city = $("industry").value === "steel" ? "beloretsk" : "ufa";
    await refresh();
  } catch (err) {
    alert(err.message);
  }
};

async function act(path, body) {
  try {
    await api("/api/action/" + path, { method: "POST", body: JSON.stringify(body) });
    await refresh();
  } catch (err) {
    alert(err.message);
  }
}

$("build").onclick = () => {
  if (!state.selected) return alert("Выберите участок");
  act("build", {
    city_id: state.city,
    building_type: $("buildType").value,
    x: state.selected.x,
    y: state.selected.y,
  });
};
$("buyParcel").onclick = () => state.selected && act("buy_parcel", { city_id: state.city, x: state.selected.x, y: state.selected.y });
$("develop").onclick = () => state.selected && act("develop_housing", { city_id: state.city, x: state.selected.x, y: state.selected.y });
$("rent").onclick = () => {
  if (!state.selected) return;
  const item = $("industry").value === "furniture" ? "furniture" : ($("industry").value === "steel" ? "steel_goods" : "bottled_milk");
  act("rent_shop", { city_id: state.city, x: state.selected.x, y: state.selected.y, item_id: item, price: 55 });
};
$("truck").onclick = () => act("buy_truck", {});
$("refresh").onclick = () => refresh().catch((e) => alert(e.message));
$("tickBtn").onclick = async () => {
  try {
    await api("/api/tick", { method: "POST", body: "{}" });
    await refresh();
  } catch (e) { alert(e.message); }
};
$("ufa").onclick = () => { state.city = "ufa"; refresh(); };
$("bel").onclick = () => { state.city = "beloretsk"; refresh(); };
$("quality").onchange = () => {
  state.quality = Number($("quality").value);
  localStorage.setItem("bc_quality", String(state.quality));
  drawMap();
};
$("subscribe").onclick = () => act("subscribe", { months: 1 });
$("hireDeputy").onclick = () => act("hire_deputy", { name: $("deputyName").value || "Заместитель" });
$("applyTasks").onclick = () => {
  const tasks = [...document.querySelectorAll(".task:checked")].map((el) => el.value);
  act("set_deputy_tasks", { tasks });
};

state.quality = Number(localStorage.getItem("bc_quality") || 2);
if ($("quality")) $("quality").value = String(state.quality);
setInterval(() => { if (state.quality >= 2 && state.world) drawMap(); }, 120);

refresh().catch(() => {
  $("log").innerHTML = "<div>Сервер не запущен. Запустите server с флагом -dev.</div>";
});
