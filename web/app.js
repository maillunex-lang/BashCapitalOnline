const state = {
  base: "http://127.0.0.1:8080",
  firmId: localStorage.getItem("bc_firm") || "",
  city: "ufa",
  world: null,
  catalog: null,
  selected: null,
  quality: Number(localStorage.getItem("bc_quality") || 2),
  imgs: {},
  tab: "map",
  manage: null,
};

const $ = (id) => document.getElementById(id);

const BUILDING_SPRITE = {
  farm: "bld_farm", livestock_farm: "bld_livestock", orchard: "bld_orchard",
  logging_camp: "bld_logging", quarry: "bld_quarry", mine: "bld_mine",
  oil_well: "bld_oil", gas_well: "bld_gas", chem_plant: "bld_chem",
  mill: "bld_mill", bakery: "bld_bakery", food_plant: "bld_food",
  dairy_plant: "bld_dairy", textile_mill: "bld_textile", apparel_factory: "bld_apparel",
  board_plant: "bld_board", sawmill: "bld_sawmill", furniture_plant: "bld_furniture",
  glass_plant: "bld_glass", paper_mill: "bld_paper", plastics_plant: "bld_plastics",
  refinery: "bld_refinery", steel_mill: "bld_steel", electronics_plant: "bld_electronics",
  appliance_plant: "bld_appliance", warehouse: "bld_warehouse", garage: "bld_garage",
};
const ZONE_TILE = {
  center: "tile_center", residential: "tile_residential", industrial: "tile_industrial",
  rural: "tile_rural", empty: "tile_grass", commercial: "tile_center",
};
const SHOP_ITEM_BY_INDUSTRY = {
  dairy: "bottled_milk", food: "bread_wheat", furniture: "furniture_wood",
  apparel: "clothing", steel: "goods_steel", oil: "fuel_can_gasoline",
};

function img(name) {
  if (!name) return null;
  if (state.imgs[name]) return state.imgs[name];
  const el = new Image();
  el.src = `./assets/map/${name}.png`;
  state.imgs[name] = el;
  return el;
}

function itemDef(id) {
  return (state.catalog?.items || []).find((x) => x.id === id);
}
function buildingDef(id) {
  return (state.catalog?.buildings || []).find((x) => x.id === id);
}
function shopDef(id) {
  return (state.catalog?.shops || []).find((x) => x.id === id);
}
function itemName(id) {
  return itemDef(id)?.name || id;
}
function buildingName(id) {
  return buildingDef(id)?.name || id;
}
function shopName(id) {
  return shopDef(id)?.name || id;
}
function itemIconSrc(id) {
  const it = itemDef(id);
  const icon = it?.icon || `item_${id}`;
  return `./assets/icons/${icon}.png`;
}
function buildingIconSrc(id) {
  const b = buildingDef(id);
  const icon = b?.icon || `bld_${id}`;
  return `./assets/icons/${icon}.png`;
}
function shopIconSrc(id) {
  const s = shopDef(id);
  const icon = s?.icon || `shop_${id}`;
  return `./assets/ui/${icon}.png`;
}
function rowWithIcon(iconSrc, html) {
  return `<div class="row-main"><img class="ico" src="${iconSrc}" alt="" loading="lazy" onerror="this.style.opacity=.25"/><div class="row-text">${html}</div></div>`;
}

function shopAllows(shopType, item) {
  const def = (state.catalog?.shops || []).find((s) => s.id === shopType);
  if (!def || !item) return false;
  if (def.allowed_items?.length) return def.allowed_items.includes(item.id);
  if (def.allowed_tiers?.length && !def.allowed_tiers.includes(item.tier)) return false;
  return (def.allowed_categories || []).includes(item.category);
}

function fillBuildsFromCatalog() {
  const sel = $("buildType");
  sel.innerHTML = "";
  (state.catalog?.buildings || []).forEach((b) => {
    const o = document.createElement("option");
    o.value = b.id;
    o.textContent = `${b.name}`;
    sel.appendChild(o);
  });
  const shopSel = $("shopType");
  shopSel.innerHTML = "";
  (state.catalog?.shops || []).forEach((s) => {
    const o = document.createElement("option");
    o.value = s.id;
    o.textContent = s.name;
    shopSel.appendChild(o);
  });
  refreshRecipes();
  refreshShopItems();
}

function refreshRecipes() {
  const recipeSel = $("recipeType");
  const typ = $("buildType").value;
  recipeSel.innerHTML = "";
  (state.catalog?.recipes || []).filter((r) => r.building_type === typ).forEach((r) => {
    const o = document.createElement("option");
    o.value = r.id;
    o.textContent = `${r.name} → ${itemName(r.output?.item_id)}`;
    recipeSel.appendChild(o);
  });
}

function refreshShopItems() {
  const shopType = $("shopType").value;
  const sel = $("shopItem");
  sel.innerHTML = "";
  (state.catalog?.items || []).filter((it) => shopAllows(shopType, it)).forEach((it) => {
    const o = document.createElement("option");
    o.value = it.id;
    o.textContent = `${it.name} (${it.tier})`;
    sel.appendChild(o);
  });
  if (!sel.options.length) {
    const o = document.createElement("option");
    o.value = "";
    o.textContent = "Нет подходящих товаров";
    sel.appendChild(o);
  }
}

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

function setTab(tab) {
  state.tab = tab;
  document.querySelectorAll(".tab").forEach((el) => el.classList.toggle("active", el.dataset.tab === tab));
  $("mapSection").hidden = tab !== "map";
  ["firm", "market", "rank", "more"].forEach((id) => {
    const p = $("panel-" + id);
    if (p) p.hidden = tab !== id;
  });
}

async function refresh() {
  state.base = $("server").value.replace(/\/$/, "");
  state.world = await api("/api/world");
  if (!state.catalog) {
    state.catalog = await api("/api/catalog");
    fillBuildsFromCatalog();
  }
  const w = state.world;
  $("tick").textContent = w.tick;
  $("date").textContent = `${String(w.month).padStart(2, "0")}.${w.year}`;
  if (state.firmId && w.firms[state.firmId]) {
    const f = w.firms[state.firmId];
    $("cash").textContent = `${money(f.cash)} ₽`;
    $("topSub").textContent = f.name;
    $("subStatus").textContent = f.subscription_active
      ? `Подписка до тика ${f.subscription_until_tick}${f.deputy?.hired ? " · " + f.deputy.name : ""}`
      : "Подписка не активна";
  }
  const cityDef = (state.catalog.cities || []).find((c) => c.id === state.city);
  $("cityPill").textContent = cityDef?.name || state.city;
  $("log").innerHTML = (w.last_tick_log || []).map((l) => `<div>${l}</div>`).join("") || "<div>Пока тихо</div>";
  const ranks = await api("/api/rankings");
  ranks.sort((a, b) => b.capital - a.capital);
  $("rank").innerHTML = ranks.map((r, i) =>
    `<div>${i + 1}. ${r.is_ai ? "AI" : "●"} <b style="color:#e8f2ee">${r.name}</b> — ${money(r.capital)} ₽</div>`
  ).join("");

  const market = w.city_market || {};
  $("market").innerHTML = Object.values(market).map((m) =>
    `<div class="row">${rowWithIcon(itemIconSrc(m.item_id), `<strong>${itemName(m.item_id)}</strong><br/>${m.stock} шт · ${money(m.price)} ₽ · q${Number(m.quality).toFixed(1)}`)}
      <button data-buy="${m.item_id}">+10</button></div>`
  ).join("") || "<div class='row'>Нет городского опта</div>";
  $("market").querySelectorAll("button[data-buy]").forEach((btn) => {
    btn.onclick = () => act("buy_city_market", { item_id: btn.dataset.buy, qty: 10 });
  });

  const orders = Object.values(w.orders || {}).filter((o) => o.qty > o.filled);
  $("orders").innerHTML = orders.slice(0, 40).map((o) => {
    const firm = w.firms[o.firm_id];
    return `<div class="row">${rowWithIcon(itemIconSrc(o.item_id), `<strong>${o.side === "sell" ? "Продажа" : "Покупка"}</strong> ${itemName(o.item_id)} ×${o.qty - o.filled}<br/>${firm?.name || o.firm_id} · ${money(o.price)} ₽`)}</div>`;
  }).join("") || "<div class='row'>Биржа пуста</div>";

  const ships = Object.values(w.shipments || {});
  if ($("shipments")) {
    $("shipments").innerHTML = "<h3 style='color:var(--accent-2);font-size:13px'>В пути (мин. 1 тик)</h3>" + (ships.map((s) =>
      `<div class="row">${rowWithIcon(itemIconSrc(s.item_id), `<strong>${itemName(s.item_id)}</strong> ×${s.qty}<br/>${s.from_city} → ${s.to_city} · ${s.ticks_left} тик · дист. ${Math.round(s.distance || 0)}`)}</div>`
    ).join("") || "<div class='row'>Нет доставок</div>");
  }

  renderMyAssets();
  drawMap();
  if (state.manage) reopenManage();
}

function renderMyAssets() {
  if (!state.firmId || !state.world) return;
  const blds = Object.values(state.world.buildings || {}).filter((b) => b.firm_id === state.firmId);
  $("myBuildings").innerHTML = "<h3 style='color:var(--accent-2);font-size:13px'>Мои заводы</h3>" + (blds.map((b) => {
    const inv = Object.entries(b.inventory || {}).filter(([, q]) => q > 0).map(([k, v]) => `${itemName(k)}:${v}`).join(", ");
    return `<div class="row">${rowWithIcon(buildingIconSrc(b.type), `<strong>${buildingName(b.type)}</strong> (${b.x},${b.y}) · ${b.recipe_id}<br/>
      ${b.sell_external === false ? "🔒 только себе" : "🌐 можно на сторону"} · вход: ${b.default_input?.mode || "own"}<br/>
      <span class="muted">${inv || "пусто"}</span>`)}
      <button data-b="${b.id}">Настроить</button></div>`;
  }).join("") || "<div class='row'>Нет заводов</div>");
  $("myBuildings").querySelectorAll("button[data-b]").forEach((btn) => {
    btn.onclick = () => openBuildingManage(btn.dataset.b);
  });

  const shops = Object.values(state.world.shops || {}).filter((s) => s.firm_id === state.firmId);
  $("myShops").innerHTML = "<h3 style='color:var(--accent-2);font-size:13px'>Мои магазины</h3>" + (shops.map((s) =>
    `<div class="row">${rowWithIcon(shopIconSrc(s.shop_type), `<strong>${shopName(s.shop_type)}</strong> → ${itemName(s.item_id)}<br/>
      запас ${s.stock} · ${money(s.price)} ₽ · источник ${s.stock_source || "own"}`)}
      <button data-s="${s.id}">Настроить</button></div>`
  ).join("") || "<div class='row'>Нет магазинов</div>");
  $("myShops").querySelectorAll("button[data-s]").forEach((btn) => {
    btn.onclick = () => openShopManage(btn.dataset.s);
  });
}

function openBuildingManage(id) {
  const b = state.world.buildings[id];
  if (!b) return;
  state.manage = { kind: "building", id };
  $("manageSheet").hidden = false;
  $("manageIcon").src = "./assets/ui/icon_factory.png";
  $("manageTitle").textContent = buildingName(b.type);
  $("manageSub").textContent = `Специализация: только рецепты «${b.type}»`;
  const recipes = (state.catalog.recipes || []).filter((r) => r.building_type === b.type);
  const firms = Object.values(state.world.firms || {}).filter((f) => f.id !== state.firmId);
  $("manageBody").innerHTML = `
    <div class="hint">Как в Capitalism Lab: мебельный завод не печёт хлеб. Выберите продукт, откуда брать входы и можно ли продавать другим игрокам.</div>
    <div class="field"><label>Что производить</label>
      <select id="mRecipe">${recipes.map((r) => `<option value="${r.id}" ${r.id === b.recipe_id ? "selected" : ""}>${r.name} → ${itemName(r.output?.item_id)}</option>`).join("")}</select>
    </div>
    <div class="field"><label>Закупка входов по умолчанию</label>
      <select id="mInput">
        <option value="own" ${b.default_input?.mode === "own" ? "selected" : ""}>Со своего склада</option>
        <option value="city" ${b.default_input?.mode === "city" ? "selected" : ""}>У города (low quality)</option>
        <option value="market" ${b.default_input?.mode === "market" ? "selected" : ""}>У других игроков (биржа)</option>
        <option value="firm" ${b.default_input?.mode === "firm" ? "selected" : ""}>У конкретной фирмы</option>
      </select>
    </div>
    <div class="field"><label>Фирма-поставщик</label>
      <select id="mFirm"><option value="">—</option>${firms.map((f) => `<option value="${f.id}" ${b.default_input?.firm_id === f.id ? "selected" : ""}>${f.name}</option>`).join("")}</select>
    </div>
    <label class="toggle"><span>Автозакупка перед производством</span><input type="checkbox" id="mAuto" ${b.auto_procure ? "checked" : ""} /></label>
    <label class="toggle"><span>Разрешить продажу другим игрокам</span><input type="checkbox" id="mSell" ${b.sell_external !== false ? "checked" : ""} /></label>
    <p class="muted">Если выключить продажу — завод отдаёт товар только вашим магазинам/заводам.</p>
    <button id="mSave">Сохранить</button>
  `;
  $("mSave").onclick = async () => {
    await act("configure_building", {
      building_id: id,
      recipe_id: $("mRecipe").value,
      default_input: $("mInput").value,
      firm_id: $("mFirm").value,
      auto_procure: $("mAuto").checked,
      sell_external: $("mSell").checked,
    });
    $("manageSheet").hidden = true;
    state.manage = null;
  };
}

function openShopManage(id) {
  const s = state.world.shops[id];
  if (!s) return;
  state.manage = { kind: "shop", id };
  $("manageSheet").hidden = false;
  $("manageIcon").src = "./assets/ui/icon_shop.png";
  $("manageTitle").textContent = shopName(s.shop_type);
  $("manageSub").textContent = "Ассортимент ограничен типом магазина";
  const shops = state.catalog.shops || [];
  const itemsFor = (type) => (state.catalog.items || []).filter((it) => shopAllows(type, it));
  const renderItems = (type, selected) => itemsFor(type).map((it) =>
    `<option value="${it.id}" ${it.id === selected ? "selected" : ""}>${it.name}</option>`
  ).join("");
  $("manageBody").innerHTML = `
    <div class="hint">Продуктовый не продаст телефон или автомобиль — тип витрины как в Capitalism Lab.</div>
    <div class="field"><label>Тип магазина</label>
      <select id="mShopType">${shops.map((x) => `<option value="${x.id}" ${x.id === s.shop_type ? "selected" : ""}>${x.name}</option>`).join("")}</select>
    </div>
    <div class="field"><label>Товар на витрине</label>
      <select id="mShopItem">${renderItems(s.shop_type, s.item_id)}</select>
    </div>
    <div class="field"><label>Розничная цена</label><input id="mPrice" type="number" value="${s.price}" /></div>
    <div class="field"><label>Откуда пополнять</label>
      <select id="mStockSrc">
        <option value="own" ${s.stock_source === "own" ? "selected" : ""}>Со своего склада</option>
        <option value="city" ${s.stock_source === "city" ? "selected" : ""}>У города</option>
        <option value="market" ${s.stock_source === "market" ? "selected" : ""}>С биржи</option>
      </select>
    </div>
    <label class="toggle"><span>Автопополнение</span><input type="checkbox" id="mRestock" ${s.auto_restock !== false ? "checked" : ""} /></label>
    <button id="mSaveShop">Сохранить</button>
    <button class="secondary" id="mStock">Пополнить +10 со склада</button>
  `;
  $("mShopType").onchange = () => {
    $("mShopItem").innerHTML = renderItems($("mShopType").value, "");
  };
  $("mSaveShop").onclick = async () => {
    await act("configure_shop", {
      shop_id: id,
      shop_type: $("mShopType").value,
      item_id: $("mShopItem").value,
      price: Number($("mPrice").value),
      stock_source: $("mStockSrc").value,
      auto_restock: $("mRestock").checked,
    });
    $("manageSheet").hidden = true;
    state.manage = null;
  };
  $("mStock").onclick = () => act("stock_shop", { shop_id: id, qty: 10 });
}

function reopenManage() {
  if (state.manage?.kind === "building") openBuildingManage(state.manage.id);
  if (state.manage?.kind === "shop") openShopManage(state.manage.id);
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
  const parcels = [...city.parcels].sort((a, b) => (a.x + a.y) - (b.x + b.y));
  const buildings = Object.values(state.world.buildings || {}).filter((b) => b.city_id === state.city);
  const shops = Object.values(state.world.shops || {}).filter((s) => s.city_id === state.city);
  const deposits = city.deposits || [];
  for (const p of parcels) {
    const [ix, iy] = iso(p.x, p.y);
    const cx = ox + ix, cy = oy + iy;
    const selected = state.selected && state.selected.x === p.x && state.selected.y === p.y;
    const tile = img(ZONE_TILE[p.zone] || "tile_grass");
    if (tile?.complete && tile.naturalWidth) {
      const s = 0.55;
      ctx.globalAlpha = selected ? 1 : 0.95;
      ctx.drawImage(tile, cx - tile.width * s * 0.5, cy - tile.height * s * 0.62, tile.width * s, tile.height * s);
      ctx.globalAlpha = 1;
    } else {
      drawTile(ctx, cx, cy, colors[p.zone] || colors.empty, selected);
    }
    const dep = deposits.find((d) => d.x === p.x && d.y === p.y && d.amount > 0);
    if (dep) {
      const di = img("res_" + dep.item_id);
      if (di?.complete && di.naturalWidth) ctx.drawImage(di, cx - 14, cy - 4, 28, 28);
    }
    if (p.has_road) {
      const road = img("tile_road");
      if (road?.complete && road.naturalWidth) {
        ctx.globalAlpha = 0.85;
        ctx.drawImage(road, cx - 18, cy - 14, 36, 28);
        ctx.globalAlpha = 1;
      }
    }
    const bld = buildings.find((b) => b.x === p.x && b.y === p.y);
    if (bld) {
      const bi = img(BUILDING_SPRITE[bld.type] || "");
      if (bi?.complete && bi.naturalWidth) {
        const s = 0.48;
        ctx.drawImage(bi, cx - bi.width * s * 0.5, cy - bi.height * s * 0.78, bi.width * s, bi.height * s);
      } else drawBuilding(ctx, cx, cy, p.zone === "industrial");
      if (selected && bld.firm_id === state.firmId) {
        // quick open hint
      }
    } else if ((p.retail_slots || 0) > 0) {
      const shopImg = img("bld_shop");
      if (shopImg?.complete && shopImg.naturalWidth) {
        const s = 0.42;
        ctx.drawImage(shopImg, cx - shopImg.width * s * 0.5, cy - shopImg.height * s * 0.75, shopImg.width * s, shopImg.height * s);
      }
    }
  }
}

function drawTile(ctx, cx, cy, color, selected) {
  const tw = 54, th = 27;
  ctx.beginPath();
  ctx.moveTo(cx, cy - th / 2);
  ctx.lineTo(cx + tw / 2, cy);
  ctx.lineTo(cx, cy + th / 2);
  ctx.lineTo(cx - tw / 2, cy);
  ctx.closePath();
  ctx.fillStyle = selected ? shade(color, 1.25) : color;
  ctx.fill();
}

function drawBuilding(ctx, cx, cy, tall) {
  const h = tall ? 28 : 16;
  ctx.fillStyle = tall ? "#9AA3AD" : "#D9C4A0";
  ctx.fillRect(cx - 12, cy - h, 24, h);
}

function shade(hex, f) {
  const n = parseInt(hex.slice(1), 16);
  return `rgb(${Math.min(255, ((n >> 16) & 255) * f) | 0},${Math.min(255, ((n >> 8) & 255) * f) | 0},${Math.min(255, (n & 255) * f) | 0})`;
}

function pickTile(mx, my) {
  const city = state.world?.cities?.[state.city];
  if (!city) return null;
  const ox = $("map").width / 2, oy = $("map").height * 0.42;
  const tw = 54, th = 27;
  const a = (mx - ox) / (tw * 0.5), b = (my - oy) / (th * 0.5);
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
  const bld = Object.values(state.world?.buildings || {}).find((b) => b.city_id === state.city && b.x === p.x && b.y === p.y);
  const shop = Object.values(state.world?.shops || {}).find((s) => s.city_id === state.city && s.x === p.x && s.y === p.y);
  let info = `Участок (${p.x},${p.y}) · ${p.zone}` + (dep ? ` · залежь ${itemName(dep.item_id)} ${dep.amount}/${dep.max_amount}` : "");
  if (bld) info += ` · ${buildingName(bld.type)}`;
  if (shop) info += ` · ${shopName(shop.shop_type)}: ${itemName(shop.item_id)}`;
  $("tileInfo").textContent = info;
  drawMap();
  if (bld && bld.firm_id === state.firmId) openBuildingManage(bld.id);
  else if (shop && shop.firm_id === state.firmId) openShopManage(shop.id);
});

$("closeSheet").onclick = () => { $("manageSheet").hidden = true; state.manage = null; };
$("buildType").onchange = () => refreshRecipes();
$("shopType").onchange = () => refreshShopItems();

document.querySelectorAll(".tab").forEach((el) => {
  el.onclick = () => setTab(el.dataset.tab);
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
    state.city = ["steel", "oil"].includes($("industry").value) ? "beloretsk" : "ufa";
    setTab("map");
    await refresh();
  } catch (err) { alert(err.message); }
};

async function act(path, body) {
  try {
    await api("/api/action/" + path, { method: "POST", body: JSON.stringify(body) });
    await refresh();
  } catch (err) { alert(err.message); }
}

$("build").onclick = () => {
  if (!state.selected) return alert("Выберите участок");
  act("build", {
    city_id: state.city,
    building_type: $("buildType").value,
    recipe_id: $("recipeType").value || "",
    x: state.selected.x,
    y: state.selected.y,
  });
};
$("buyParcel").onclick = () => state.selected && act("buy_parcel", { city_id: state.city, x: state.selected.x, y: state.selected.y });
$("develop").onclick = () => state.selected && act("develop_housing", { city_id: state.city, x: state.selected.x, y: state.selected.y });
$("rent").onclick = () => {
  if (!state.selected) return;
  const item = $("shopItem").value || SHOP_ITEM_BY_INDUSTRY[$("industry").value] || "bread";
  act("rent_shop", {
    city_id: state.city, x: state.selected.x, y: state.selected.y,
    shop_type: $("shopType").value, item_id: item, price: 55,
  });
};
$("truck").onclick = () => act("buy_truck", {});
$("refresh").onclick = () => refresh().catch((e) => alert(e.message));
$("tickBtn").onclick = async () => {
  try { await api("/api/tick", { method: "POST", body: "{}" }); await refresh(); }
  catch (e) { alert(e.message); }
};
$("ufa").onclick = () => { state.city = "ufa"; refresh(); };
$("bel").onclick = () => { state.city = "beloretsk"; refresh(); };
$("quality").value = String(state.quality);
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

setTab("map");
refresh().catch(() => {});
