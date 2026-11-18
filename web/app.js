const state = {
  base: "http://127.0.0.1:8080",
  firmId: localStorage.getItem("bc_firm") || "",
  city: sessionStorage.getItem("bc_shot_city") || "ufa",
  world: null,
  catalog: null,
  selected: null,
  quality: Number(localStorage.getItem("bc_quality") || 2),
  imgs: {},
  tab: "desk",
  manage: null,
  buildCat: "all",
  pickBuilding: null,
  pickShop: null,
  place: null, // { kind: 'build'|'shop', buildingType?, recipeId?, shopType?, itemId? }
  mapZoom: 1,
};
sessionStorage.removeItem("bc_shot_city");

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
const CAT_LABELS = {
  all: "Все",
  extract: "Добыча",
  factory: "Заводы",
  logistics: "Логистика",
  food: "Еда",
  dairy: "Молочка",
  steel: "Металл",
  oil: "Нефть",
  apparel: "Одежда",
  furniture: "Мебель",
  tech: "Техника",
};

function img(name, folder = "map") {
  if (!name) return null;
  const key = `${folder}:${name}`;
  if (state.imgs[key]) return state.imgs[key];
  const el = new Image();
  el.onload = () => { if (state.tab === "map") drawMap(); };
  el.src = `./assets/${folder}/${name}.png`;
  state.imgs[key] = el;
  return el;
}
function mapImg(name) { return img(name, "map"); }
function assetImg(...names) {
  for (const n of names) {
    if (!n) continue;
    const folder = n.startsWith("item_") || n.startsWith("bld_") ? "icons"
      : n.startsWith("shop_") || n.startsWith("icon_") ? "ui" : "map";
    const el = img(n, folder);
    if (el) return el;
  }
  return null;
}

function itemDef(id) { return (state.catalog?.items || []).find((x) => x.id === id); }
function buildingDef(id) { return (state.catalog?.buildings || []).find((x) => x.id === id); }
function shopDef(id) { return (state.catalog?.shops || []).find((x) => x.id === id); }
function itemName(id) { return itemDef(id)?.name || id; }
function buildingName(id) { return buildingDef(id)?.name || id; }
function shopName(id) { return shopDef(id)?.name || id; }
function itemIconSrc(id) {
  const it = itemDef(id);
  return `./assets/icons/${it?.icon || `item_${id}`}.png`;
}
function buildingIconSrc(id) {
  const b = buildingDef(id);
  return `./assets/icons/${b?.icon || `bld_${id}`}.png`;
}
function shopIconSrc(id) {
  const s = shopDef(id);
  return `./assets/ui/${s?.icon || `shop_${id}`}.png`;
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

function enabledCities() {
  return (state.catalog?.cities || []).filter((c) => c.enabled !== false);
}

function money(v) { return Math.round(Number(v || 0)).toLocaleString("ru-RU"); }

function cityRadius() {
  return state.world?.cities?.[state.city]?.radius || 4;
}

function tileScale() {
  // Fit isometric city into canvas; quality + placement bump size further.
  const canvas = $("map");
  const w = canvas?.width || 1100;
  const h = canvas?.height || 720;
  const r = cityRadius();
  // city spans ~(2r+1) tiles on both diagonals
  const span = 2 * r + 1;
  const fitX = (w * 0.86) / (span * 72 * 0.5);
  const fitY = (h * 0.72) / (span * 36 * 0.5);
  let s = Math.min(fitX, fitY);
  s *= [0.92, 1, 1.08, 1.16][state.quality] || 1.08;
  if (state.place) s *= 1.12;
  return Math.max(1.05, Math.min(2.4, s));
}

function isoMetrics() {
  const s = tileScale();
  return { tw: 72 * s, th: 36 * s, sprite: 0.78 * s, icon: 56 * s };
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

function setTab(tab) {
  state.tab = tab;
  document.querySelectorAll(".tab").forEach((el) => el.classList.toggle("active", el.dataset.tab === tab));
  $("desktopSection").hidden = tab !== "desk";
  $("mapSection").hidden = tab !== "map";
  ["market", "rank", "more"].forEach((id) => {
    const p = $("panel-" + id);
    if (p) p.hidden = tab !== id;
  });
  if (tab === "map") {
    requestAnimationFrame(() => drawMap());
  }
}

function fillBuildsFromCatalog() {
  const sel = $("buildType");
  if (sel) {
    sel.innerHTML = "";
    (state.catalog?.buildings || []).forEach((b) => {
      const o = document.createElement("option");
      o.value = b.id;
      o.textContent = b.name;
      sel.appendChild(o);
    });
  }
  const shopSel = $("shopType");
  if (shopSel) {
    shopSel.innerHTML = "";
    (state.catalog?.shops || []).forEach((s) => {
      const o = document.createElement("option");
      o.value = s.id;
      o.textContent = s.name;
      shopSel.appendChild(o);
    });
  }
  fillCityChips();
  renderBuildCats();
  renderBuildCatalog();
  renderShopCatalog();
}

function buildingCategories() {
  // Curated chips only — raw industry_tag dump (agro/timber/…) cluttered the desk.
  const order = ["all", "extract", "factory", "logistics", "food", "dairy", "furniture", "apparel", "steel", "oil", "tech"];
  const present = new Set(["all"]);
  (state.catalog?.buildings || []).forEach((b) => {
    if (order.includes(b.category)) present.add(b.category);
    if (order.includes(b.industry_tag)) present.add(b.industry_tag);
  });
  return order.filter((id) => present.has(id));
}

function renderBuildCats() {
  const host = $("buildCats");
  if (!host) return;
  host.innerHTML = "";
  buildingCategories().forEach((id) => {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "cat-chip" + (state.buildCat === id ? " active" : "");
    btn.textContent = CAT_LABELS[id] || id;
    btn.onclick = () => {
      state.buildCat = id;
      renderBuildCats();
      renderBuildCatalog();
    };
    host.appendChild(btn);
  });
}

function filteredBuildings() {
  return (state.catalog?.buildings || []).filter((b) => {
    if (state.buildCat === "all") return true;
    return b.category === state.buildCat || b.industry_tag === state.buildCat;
  });
}

function renderBuildCatalog() {
  const host = $("buildCatalog");
  if (!host) return;
  const list = filteredBuildings();
  host.innerHTML = list.map((b) => {
    const active = state.pickBuilding === b.id ? " active" : "";
    const cost = b.build_cost ? `${money(b.build_cost)} ₽` : (b.zone || "");
    return `<button type="button" class="catalog-card${active}" data-b="${b.id}">
      <img src="${buildingIconSrc(b.id)}" alt="" loading="lazy" onerror="this.style.opacity=.3"/>
      <strong>${b.name}</strong>
      <span>${cost}</span>
    </button>`;
  }).join("") || `<p class="muted">Нет зданий в категории</p>`;
  host.querySelectorAll("button[data-b]").forEach((btn) => {
    btn.onclick = () => selectBuilding(btn.dataset.b);
  });
}

function selectBuilding(id) {
  state.pickBuilding = id;
  state.pickShop = null;
  if ($("buildType")) $("buildType").value = id;
  renderBuildCatalog();
  renderShopCatalog();
  const b = buildingDef(id);
  $("buildDetail").hidden = false;
  $("shopDetail").hidden = true;
  $("buildPickIcon").src = buildingIconSrc(id);
  $("buildPickName").textContent = b?.name || id;
  $("buildPickMeta").textContent = [b?.category, b?.zone, b?.build_cost ? `стройка ${money(b.build_cost)} ₽` : ""]
    .filter(Boolean).join(" · ");
  refreshRecipes();
}

function refreshRecipes() {
  const recipeSel = $("recipeType");
  if (!recipeSel) return;
  const typ = state.pickBuilding || $("buildType")?.value;
  recipeSel.innerHTML = "";
  (state.catalog?.recipes || []).filter((r) => r.building_type === typ).forEach((r) => {
    const o = document.createElement("option");
    o.value = r.id;
    o.textContent = `${r.name} → ${itemName(r.output?.item_id)}`;
    recipeSel.appendChild(o);
  });
  if (!recipeSel.options.length) {
    const o = document.createElement("option");
    o.value = "";
    o.textContent = "Без рецепта / добыча";
    recipeSel.appendChild(o);
  }
}

function renderShopCatalog() {
  const host = $("shopCatalog");
  if (!host) return;
  host.innerHTML = (state.catalog?.shops || []).map((s) => {
    const active = state.pickShop === s.id ? " active" : "";
    return `<button type="button" class="catalog-card${active}" data-s="${s.id}">
      <img src="${shopIconSrc(s.id)}" alt="" loading="lazy" onerror="this.style.opacity=.3"/>
      <strong>${s.name}</strong>
      <span>${(s.allowed_categories || []).slice(0, 2).join(", ") || "розница"}</span>
    </button>`;
  }).join("");
  host.querySelectorAll("button[data-s]").forEach((btn) => {
    btn.onclick = () => selectShop(btn.dataset.s);
  });
}

function selectShop(id) {
  state.pickShop = id;
  state.pickBuilding = null;
  if ($("shopType")) $("shopType").value = id;
  renderShopCatalog();
  renderBuildCatalog();
  $("shopDetail").hidden = false;
  $("buildDetail").hidden = true;
  const s = shopDef(id);
  $("shopPickIcon").src = shopIconSrc(id);
  $("shopPickName").textContent = s?.name || id;
  $("shopPickMeta").textContent = (s?.allowed_categories || []).join(", ") || "специализация витрины";
  refreshShopItems();
}

function refreshShopItems() {
  const shopType = state.pickShop || $("shopType")?.value;
  const sel = $("shopItem");
  if (!sel) return;
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
  } else {
    const prefer = SHOP_ITEM_BY_INDUSTRY[$("industry")?.value] || "bread_wheat";
    if ([...sel.options].some((o) => o.value === prefer)) sel.value = prefer;
  }
}

function fillCityChips() {
  const host = $("cityChips");
  if (!host) return;
  host.innerHTML = "";
  enabledCities().forEach((c) => {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "chip" + (state.city === c.id ? " active" : "");
    btn.dataset.city = c.id;
    btn.innerHTML = `<img src="./assets/ui/icon_map.png" alt="" /> ${c.name}`;
    btn.onclick = () => {
      state.city = c.id;
      fillCityChips();
      refresh().catch(() => drawMap());
    };
    host.appendChild(btn);
  });
}

function openCityPicker(kind) {
  const body = $("cityPickBody");
  const cities = enabledCities();
  $("cityPickSub").textContent = kind === "shop"
    ? "Аренда: выберите город с коммерческими слотами"
    : "Стройка: выберите город, затем участок на карте";
  body.innerHTML = cities.map((c) => {
    const pop = state.world?.cities?.[c.id]?.population;
    return `<button type="button" class="city-pick-card" data-city="${c.id}">
      <img src="./assets/ui/icon_map.png" alt="" />
      <div><strong>${c.name}</strong><span class="muted">${pop ? `население ${money(pop)}` : "открытый город"}</span></div>
    </button>`;
  }).join("");
  body.querySelectorAll("button[data-city]").forEach((btn) => {
    btn.onclick = () => {
      state.city = btn.dataset.city;
      $("cityPickSheet").hidden = true;
      fillCityChips();
      enterPlaceMode();
    };
  });
  $("cityPickSheet").hidden = false;
}

function enterPlaceMode() {
  if (!state.place) return;
  setTab("map");
  $("placeBanner").hidden = false;
  $("confirmPlace").hidden = false;
  if (state.place.kind === "build") {
    const name = buildingName(state.place.buildingType);
    $("placeTitle").textContent = `Строим: ${name}`;
    $("placeSub").textContent = "Выберите свободный участок (для добычи — с нужной залежью), затем «Подтвердить»";
    $("confirmPlace").innerHTML = `<img src="./assets/ui/icon_factory.png" alt="" /> Построить здесь`;
  } else {
    const name = shopName(state.place.shopType);
    $("placeTitle").textContent = `Аренда: ${name}`;
    $("placeSub").textContent = "Выберите коммерческий слот (жилой/центр с магазином), затем подтвердите";
    $("confirmPlace").innerHTML = `<img src="./assets/ui/icon_shop.png" alt="" /> Арендовать здесь`;
  }
  state.selected = null;
  $("tileInfo").textContent = "Тапните участок на карте";
  drawMap();
}

function exitPlaceMode() {
  state.place = null;
  $("placeBanner").hidden = true;
  $("confirmPlace").hidden = true;
  drawMap();
}

function startBuildPlace() {
  if (!state.firmId) return alert("Сначала откройте фирму во вкладке кабинета");
  if (!state.pickBuilding) return alert("Выберите здание из каталога");
  state.place = {
    kind: "build",
    buildingType: state.pickBuilding,
    recipeId: $("recipeType").value || "",
  };
  openCityPicker("build");
}

function startShopPlace() {
  if (!state.firmId) return alert("Сначала откройте фирму во вкладке кабинета");
  if (!state.pickShop) return alert("Выберите тип магазина");
  const item = $("shopItem").value;
  if (!item) return alert("Выберите товар для витрины");
  state.place = {
    kind: "shop",
    shopType: state.pickShop,
    itemId: item,
  };
  openCityPicker("shop");
}

function retailPriceForItem(itemId) {
  const corr = state.catalog?.economy?.price_corridors?.[itemId];
  if (corr && Number.isFinite(corr.min) && Number.isFinite(corr.max)) {
    return Math.round((corr.min + corr.max) / 2);
  }
  return 0;
}

function cityLabel(id) {
  return (state.catalog?.cities || []).find((c) => c.id === id)?.name || id;
}

function fillTradeForm() {
  const citySel = $("ordCity");
  const itemSel = $("ordItem");
  if (!citySel || !itemSel || !state.catalog) return;
  const curCity = citySel.value || state.city;
  const curItem = itemSel.value;
  citySel.innerHTML = enabledCities().map((c) =>
    `<option value="${c.id}" ${c.id === curCity ? "selected" : ""}>${c.name}</option>`
  ).join("");
  const finished = (state.catalog.items || []).filter((it) => it.tier === "finished" || it.tier === "intermediate");
  const list = finished.length ? finished : (state.catalog.items || []);
  itemSel.innerHTML = list.slice(0, 200).map((it) =>
    `<option value="${it.id}" ${it.id === curItem ? "selected" : ""}>${it.name}</option>`
  ).join("");
  if (!itemSel.value && list.length) itemSel.value = list[0].id;
  const mid = retailPriceForItem(itemSel.value);
  if (mid > 0 && (!$("ordPrice").dataset.touched || !$("ordPrice").value)) {
    $("ordPrice").value = String(mid);
  }
}

async function acceptSellAsContract(orderId) {
  const o = state.world?.orders?.[orderId];
  if (!o || !state.firmId) return;
  const to = state.city || o.city_id;
  await act("contract", {
    seller_id: o.firm_id,
    item_id: o.item_id,
    city_from: o.city_id,
    city_to: to,
    qty: Math.min(10, o.qty - o.filled),
    price: o.price,
    ticks: 12,
    use_own_fleet: true,
  });
}

async function confirmPlace() {
  if (!state.place || !state.selected) return alert("Сначала выберите участок на карте");
  const p = state.selected;
  if (state.place.kind === "build") {
    await act("build", {
      city_id: state.city,
      building_type: state.place.buildingType,
      recipe_id: state.place.recipeId || "",
      x: p.x, y: p.y,
    });
  } else {
    await act("rent_shop", {
      city_id: state.city, x: p.x, y: p.y,
      shop_type: state.place.shopType,
      item_id: state.place.itemId,
      price: retailPriceForItem(state.place.itemId),
    });
  }
  exitPlaceMode();
  setTab("desk");
}

function renderDeputyTasks(f) {
  const host = $("deputyTasks");
  if (!host) return;
  if (!f.deputy?.hired) {
    host.innerHTML = "<div class='row muted'>Заместитель не нанят</div>";
    return;
  }
  const tasks = f.deputy.tasks || [];
  host.innerHTML = `<h3 style="color:var(--accent-2);font-size:13px">Задания · ${f.deputy.name}</h3>` + (
    tasks.length
      ? tasks.map((t) =>
        `<div class="row"><div><strong>${t.label || t.kind}</strong><br/><span class="muted">${t.active === false ? "выкл" : "активно"}${t.last_note ? " · " + t.last_note : ""}</span></div></div>`
      ).join("")
      : "<div class='row muted'>Задания не выданы</div>"
  );
}

function renderDeskReport() {
  if (!state.world) return;
  const f = state.firmId ? state.world.firms[state.firmId] : null;
  $("deskTitle").textContent = f?.name || "Ваша фирма";
  $("deskHint").textContent = f
    ? "Стройте из каталога, магазины арендуйте здесь, карту открывайте только для участка или осмотра города."
    : "Откройте фирму справа внизу — затем выберите здание и нажмите «Выбрать город и участок».";
  const blds = Object.values(state.world.buildings || {}).filter((b) => b.firm_id === state.firmId);
  const shops = Object.values(state.world.shops || {}).filter((s) => s.firm_id === state.firmId);
  $("repBuildings").textContent = String(blds.length);
  $("repShops").textContent = String(shops.length);
  $("repTrucks").textContent = String(f?.trucks || 0);
  const log = (state.world.last_tick_log || []).slice(-6);
  $("deskLog").innerHTML = log.map((l) => `<div class="row">${l}</div>`).join("") || "<div class='row muted'>Журнал пуст — нажмите «Тик»</div>";
}

async function refresh() {
  state.base = ($("server")?.value || state.base).replace(/\/$/, "");
  state.world = await api("/api/world");
  if (!state.catalog) {
    state.catalog = await api("/api/catalog");
    // Soft-compat for Economy Catalog V2 (dynamic IDs; no _std/_prem clones).
    state.catalog.items = state.catalog.items || [];
    state.catalog.recipes = state.catalog.recipes || [];
    state.catalog.buildings = state.catalog.buildings || [];
    state.catalog.shops = state.catalog.shops || [];
    state.economyCatalogVersion = state.catalog.economy_catalog_version || state.catalog.version || 0;
    fillBuildsFromCatalog();
  }
  const w = state.world;
  $("tick").textContent = w.tick;
  const seasons = ["Кыш", "Яҙ", "Йәй", "Көҙ"];
  const season = seasons[Math.floor(((w.month || 1) - 1) / 3)] || "";
  $("date").textContent = `${String(w.month).padStart(2, "0")}.${w.year} · ${season}`;
  if (state.firmId && w.firms[state.firmId]) {
    const f = w.firms[state.firmId];
    $("cash").textContent = `${money(f.cash)} ₽`;
    $("topSub").textContent = f.name;
    if ($("subStatus")) {
      $("subStatus").textContent = f.subscription_active
        ? `Подписка до тика ${f.subscription_until_tick}${f.deputy?.hired ? " · " + f.deputy.name : ""}`
        : "Подписка не активна (нужна для заместителя)";
    }
    renderDeputyTasks(f);
  } else if ($("deputyTasks")) {
    $("deputyTasks").innerHTML = "<div class='row muted'>Откройте фирму и оформите подписку</div>";
  }
  const cityDef = (state.catalog.cities || []).find((c) => c.id === state.city);
  if ($("cityPill")) $("cityPill").textContent = cityDef?.name || state.city;
  if ($("log")) {
    $("log").innerHTML = (w.last_tick_log || []).map((l) => `<div>${l}</div>`).join("") || "<div>Пока тихо</div>";
  }
  const ranks = await api("/api/rankings");
  ranks.sort((a, b) => b.capital - a.capital);
  if ($("rank")) {
    $("rank").innerHTML = ranks.map((r, i) =>
      `<div>${i + 1}. ${r.is_ai ? "AI" : "●"} <b style="color:#e8f2ee">${r.name}</b> — ${money(r.capital)} ₽</div>`
    ).join("");
  }

  const market = w.city_market || {};
  if ($("market")) {
    $("market").innerHTML = Object.values(market).map((m) =>
      `<div class="row">${rowWithIcon(itemIconSrc(m.item_id), `<strong>${itemName(m.item_id)}</strong><br/>${m.stock} шт · ${money(m.price)} ₽ · q${Number(m.quality).toFixed(1)}`)}
        <button data-buy="${m.item_id}">+10</button></div>`
    ).join("") || "<div class='row'>Нет городского опта</div>";
    $("market").querySelectorAll("button[data-buy]").forEach((btn) => {
      btn.onclick = () => act("buy_city_market", { item_id: btn.dataset.buy, qty: 10 });
    });
  }

  fillTradeForm();
  const orders = Object.values(w.orders || {}).filter((o) => o.qty > o.filled);
  if ($("orders")) {
    $("orders").innerHTML = orders.slice(0, 50).map((o) => {
      const firm = w.firms[o.firm_id];
      const left = o.qty - o.filled;
      const mine = o.firm_id === state.firmId;
      const canContract = !mine && o.side === "sell" && state.firmId;
      return `<div class="row">${rowWithIcon(itemIconSrc(o.item_id), `<strong>${o.side === "sell" ? "Продажа" : "Покупка"}</strong> ${itemName(o.item_id)} ×${left}<br/>${firm?.name || o.firm_id} · ${cityLabel(o.city_id)} · ${money(o.price)} ₽`)}
        ${canContract ? `<button data-contract="${o.id}">В контракт</button>` : (mine ? `<span class="pill">ваш</span>` : "")}</div>`;
    }).join("") || "<div class='row'>Биржа пуста</div>";
    $("orders").querySelectorAll("button[data-contract]").forEach((btn) => {
      btn.onclick = () => acceptSellAsContract(btn.dataset.contract);
    });
  }
  if ($("contracts")) {
    const contracts = Object.values(w.contracts || {}).filter((c) => c.ticks_left > 0);
    $("contracts").innerHTML = contracts.slice(0, 40).map((c) => {
      const seller = w.firms[c.seller_id]?.name || c.seller_id;
      const buyer = w.firms[c.buyer_id]?.name || c.buyer_id;
      const mine = c.seller_id === state.firmId || c.buyer_id === state.firmId;
      return `<div class="row">${rowWithIcon(itemIconSrc(c.item_id), `<strong>${itemName(c.item_id)}</strong> ×${c.qty_per_tick}/тик · ${money(c.price)} ₽<br/>${seller} → ${buyer}<br/><span class="muted">${cityLabel(c.city_from)} → ${cityLabel(c.city_to)} · ещё ${c.ticks_left} тик${mine ? " · ваш" : ""}</span>`)}</div>`;
    }).join("") || "<div class='row'>Нет активных контрактов</div>";
  }
  const ships = Object.values(w.shipments || {});
  if ($("shipments")) {
    $("shipments").innerHTML = "<h3 style='color:var(--accent-2);font-size:13px'>В пути</h3>" + (ships.map((s) =>
      `<div class="row">${rowWithIcon(itemIconSrc(s.item_id), `<strong>${itemName(s.item_id)}</strong> ×${s.qty}<br/>${cityLabel(s.from_city)} → ${cityLabel(s.to_city)} · ${s.ticks_left} тик`)}</div>`
    ).join("") || "<div class='row'>Нет доставок</div>");
  }

  renderDeskReport();
  renderMyAssets();
  if (state.tab === "map") drawMap();
  if (state.manage) reopenManage();
}

function renderMyAssets() {
  if (!$("myBuildings")) return;
  if (!state.firmId || !state.world) {
    $("myBuildings").innerHTML = "<div class='row muted'>Нет заводов</div>";
    $("myShops").innerHTML = "<div class='row muted'>Нет магазинов</div>";
    return;
  }
  const blds = Object.values(state.world.buildings || {}).filter((b) => b.firm_id === state.firmId);
  $("myBuildings").innerHTML = "<h3 style='color:var(--accent-2);font-size:13px'>Мои заводы</h3>" + (blds.map((b) => {
    const inv = Object.entries(b.inventory || {}).filter(([, q]) => q > 0).map(([k, v]) => `${itemName(k)}:${v}`).join(", ");
    const city = (state.catalog?.cities || []).find((c) => c.id === b.city_id)?.name || b.city_id;
    return `<div class="row">${rowWithIcon(buildingIconSrc(b.type), `<strong>${buildingName(b.type)}</strong> · ${city} (${b.x},${b.y})<br/>
      ${b.sell_external === false ? "🔒 себе" : "🌐 наружу"} · ${b.recipe_id || "—"}<br/>
      <span class="muted">${inv || "пусто"}</span>`)}
      <button data-b="${b.id}">Настроить</button></div>`;
  }).join("") || "<div class='row'>Нет заводов</div>");
  $("myBuildings").querySelectorAll("button[data-b]").forEach((btn) => {
    btn.onclick = () => openBuildingManage(btn.dataset.b);
  });

  const shops = Object.values(state.world.shops || {}).filter((s) => s.firm_id === state.firmId);
  $("myShops").innerHTML = "<h3 style='color:var(--accent-2);font-size:13px'>Мои магазины</h3>" + (shops.map((s) =>
    `<div class="row">${rowWithIcon(shopIconSrc(s.shop_type), `<strong>${shopName(s.shop_type)}</strong> → ${itemName(s.item_id)}<br/>
      запас ${s.stock} · ${money(s.price)} ₽ · ${s.stock_source || "own"}`)}
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
  $("manageIcon").src = buildingIconSrc(b.type);
  $("manageTitle").textContent = buildingName(b.type);
  $("manageSub").textContent = `Специализация: только рецепты «${b.type}»`;
  const recipes = (state.catalog.recipes || []).filter((r) => r.building_type === b.type);
  const firms = Object.values(state.world.firms || {}).filter((f) => f.id !== state.firmId);
  $("manageBody").innerHTML = `
    <div class="hint">Как в Capitalism Lab: тип здания фиксирует линейку продуктов.</div>
    <div class="field"><label>Что производить</label>
      <select id="mRecipe">${recipes.map((r) => `<option value="${r.id}" ${r.id === b.recipe_id ? "selected" : ""}>${r.name} → ${itemName(r.output?.item_id)}</option>`).join("")}</select>
    </div>
    <div class="field"><label>Закупка входов</label>
      <select id="mInput">
        <option value="own" ${b.default_input?.mode === "own" ? "selected" : ""}>Со своего склада</option>
        <option value="city" ${b.default_input?.mode === "city" ? "selected" : ""}>У города</option>
        <option value="market" ${b.default_input?.mode === "market" ? "selected" : ""}>С биржи</option>
        <option value="firm" ${b.default_input?.mode === "firm" ? "selected" : ""}>У фирмы</option>
      </select>
    </div>
    <div class="field"><label>Фирма-поставщик</label>
      <select id="mFirm"><option value="">—</option>${firms.map((f) => `<option value="${f.id}" ${b.default_input?.firm_id === f.id ? "selected" : ""}>${f.name}</option>`).join("")}</select>
    </div>
    <label class="toggle"><span>Автозакупка</span><input type="checkbox" id="mAuto" ${b.auto_procure ? "checked" : ""} /></label>
    <label class="toggle"><span>Продажа другим</span><input type="checkbox" id="mSell" ${b.sell_external !== false ? "checked" : ""} /></label>
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
  $("manageIcon").src = shopIconSrc(s.shop_type);
  $("manageTitle").textContent = shopName(s.shop_type);
  $("manageSub").textContent = "Ассортимент ограничен типом магазина";
  const shops = state.catalog.shops || [];
  const itemsFor = (type) => (state.catalog.items || []).filter((it) => shopAllows(type, it));
  const renderItems = (type, selected) => itemsFor(type).map((it) =>
    `<option value="${it.id}" ${it.id === selected ? "selected" : ""}>${it.name}</option>`
  ).join("");
  $("manageBody").innerHTML = `
    <div class="hint">Тип витрины ограничивает товары, как в Capitalism Lab.</div>
    <div class="field"><label>Тип магазина</label>
      <select id="mShopType">${shops.map((x) => `<option value="${x.id}" ${x.id === s.shop_type ? "selected" : ""}>${x.name}</option>`).join("")}</select>
    </div>
    <div class="field"><label>Товар</label>
      <select id="mShopItem">${renderItems(s.shop_type, s.item_id)}</select>
    </div>
    <div class="field"><label>Цена</label><input id="mPrice" type="number" value="${s.price}" /></div>
    <div class="field"><label>Источник</label>
      <select id="mStockSrc">
        <option value="own" ${s.stock_source === "own" ? "selected" : ""}>Склад</option>
        <option value="city" ${s.stock_source === "city" ? "selected" : ""}>Город</option>
        <option value="market" ${s.stock_source === "market" ? "selected" : ""}>Биржа</option>
      </select>
    </div>
    <label class="toggle"><span>Автопополнение</span><input type="checkbox" id="mRestock" ${s.auto_restock !== false ? "checked" : ""} /></label>
    <button id="mSaveShop">Сохранить</button>
    <button class="secondary" id="mStock">Пополнить +10</button>
  `;
  $("mShopType").onchange = () => { $("mShopItem").innerHTML = renderItems($("mShopType").value, ""); };
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
  const { tw, th } = isoMetrics();
  return [(x - y) * tw * 0.5, (x + y) * th * 0.5];
}

function preferMapSprite(el) {
  // Skip tiny placeholder map sprites (~1KB / tiny natural size); prefer catalog icons.
  if (!el?.complete || !el.naturalWidth) return false;
  return el.naturalWidth >= 64 && el.naturalHeight >= 64;
}

function drawMap() {
  const canvas = $("map");
  if (!canvas) return;
  const ctx = canvas.getContext("2d", { alpha: false });
  const w = canvas.width, h = canvas.height;

  // Opaque base — never leave transparent pixels (headless shots / photo bleed)
  ctx.fillStyle = "#122820";
  ctx.fillRect(0, 0, w, h);
  const g = ctx.createRadialGradient(w * 0.5, h * 0.32, 30, w * 0.5, h * 0.48, w * 0.75);
  g.addColorStop(0, "#2d5a4c");
  g.addColorStop(0.55, "#1a3830");
  g.addColorStop(1, "#0e1c18");
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, w, h);

  const city = state.world?.cities?.[state.city];
  if (!city) {
    ctx.fillStyle = "#eef6f2";
    ctx.font = "600 22px Manrope";
    ctx.fillText("Загрузка города…", 40, 60);
    return;
  }

  const { tw, th, sprite, icon } = isoMetrics();
  const ox = w / 2, oy = h * 0.36;
  const colors = {
    center: "#3a7f78", residential: "#4f8a5f", industrial: "#7a828c",
    rural: "#7fad68", empty: "#3f524c", commercial: "#c4a35a",
  };
  const parcels = [...city.parcels].sort((a, b) => (a.x + a.y) - (b.x + b.y));
  const buildings = Object.values(state.world.buildings || {}).filter((b) => b.city_id === state.city);
  const shops = Object.values(state.world.shops || {}).filter((s) => s.city_id === state.city);
  const deposits = city.deposits || [];

  for (const p of parcels) {
    const [ix, iy] = iso(p.x, p.y);
    const cx = ox + ix, cy = oy + iy;
    const selected = state.selected && state.selected.x === p.x && state.selected.y === p.y;
    // Empty industrial pads look like a gray checkerboard at scale — use meadow unless occupied.
    const occupied = buildings.some((b) => b.x === p.x && b.y === p.y)
      || shops.some((s) => s.x === p.x && s.y === p.y);
    let tileName = ZONE_TILE[p.zone] || "tile_grass";
    if (p.zone === "industrial" && !occupied) tileName = "tile_meadow";
    if (p.zone === "empty") tileName = "tile_grass";
    const tile = mapImg(tileName);
    if (preferMapSprite(tile)) {
      const dw = tw * 1.08, dh = th * 2.1;
      ctx.drawImage(tile, cx - dw / 2, cy - dh * 0.55, dw, dh);
    } else {
      drawTile(ctx, cx, cy, colors[p.zone] || colors.empty, selected, tw, th);
    }
    if (selected) {
      ctx.beginPath();
      ctx.moveTo(cx, cy - th / 2);
      ctx.lineTo(cx + tw / 2, cy);
      ctx.lineTo(cx, cy + th / 2);
      ctx.lineTo(cx - tw / 2, cy);
      ctx.closePath();
      ctx.strokeStyle = "#e0c575";
      ctx.lineWidth = Math.max(3, 4 * tileScale() * 0.45);
      ctx.stroke();
    }

    const dep = deposits.find((d) => d.x === p.x && d.y === p.y && d.amount > 0);
    if (dep) {
      const di = assetImg("item_" + dep.item_id, itemDef(dep.item_id)?.icon, "res_" + dep.item_id);
      const size = Math.min(icon * 0.55, tw * 0.55);
      if (di?.complete && di.naturalWidth) {
        // Soft plate under deposit icon
        ctx.fillStyle = "rgba(8,20,16,0.45)";
        ctx.beginPath();
        ctx.ellipse(cx, cy + th * 0.05, size * 0.42, size * 0.22, 0, 0, Math.PI * 2);
        ctx.fill();
        ctx.drawImage(di, cx - size / 2, cy - size * 0.55, size, size);
      }
    }
    if (p.has_road) {
      const road = mapImg("tile_road");
      if (preferMapSprite(road)) {
        const dw = tw * 0.95, dh = th * 1.7;
        ctx.globalAlpha = 0.9;
        ctx.drawImage(road, cx - dw / 2, cy - dh * 0.45, dw, dh);
        ctx.globalAlpha = 1;
      }
    }
    const bld = buildings.find((b) => b.x === p.x && b.y === p.y);
    const shopHere = shops.find((s) => s.x === p.x && s.y === p.y);
    if (bld) {
      const mapSprite = mapImg(BUILDING_SPRITE[bld.type] || ("bld_" + bld.type));
      if (preferMapSprite(mapSprite)) {
        const s = Math.min(tw / 110, th / 40) * 1.35;
        ctx.drawImage(mapSprite, cx - mapSprite.width * s * 0.5, cy - mapSprite.height * s * 0.88, mapSprite.width * s, mapSprite.height * s);
      } else {
        drawBuilding(ctx, cx, cy, p.zone === "industrial" || buildingDef(bld.type)?.category === "factory", tw);
        const badge = assetImg("bld_" + bld.type, buildingDef(bld.type)?.icon);
        const bs = Math.min(icon * 0.38, tw * 0.38);
        if (badge?.complete && badge.naturalWidth) {
          ctx.drawImage(badge, cx - bs / 2, cy - tw * 0.55 - bs * 0.2, bs, bs);
        }
      }
      if (bld.firm_id === state.firmId) {
        ctx.fillStyle = "#e0c575";
        ctx.beginPath();
        ctx.arc(cx + tw * 0.28, cy - th * 1.1, Math.max(4, tw * 0.06), 0, Math.PI * 2);
        ctx.fill();
      }
    } else if (shopHere) {
      const mapShop = mapImg("bld_shop");
      if (preferMapSprite(mapShop)) {
        const s = Math.min(tw / 110, th / 40) * 1.25;
        ctx.drawImage(mapShop, cx - mapShop.width * s * 0.5, cy - mapShop.height * s * 0.88, mapShop.width * s, mapShop.height * s);
      } else {
        drawBuilding(ctx, cx, cy, false, tw);
      }
      const pi = assetImg("item_" + shopHere.item_id, itemDef(shopHere.item_id)?.icon);
      const ps = Math.min(icon * 0.32, tw * 0.32);
      if (pi?.complete && pi.naturalWidth) ctx.drawImage(pi, cx + tw * 0.12, cy - th * 0.2, ps, ps);
    } else if ((p.retail_slots || p.slots || 0) > 0) {
      ctx.fillStyle = "rgba(212,181,106,0.4)";
      ctx.beginPath();
      ctx.arc(cx, cy - th * 0.12, Math.max(6, tw * 0.1), 0, Math.PI * 2);
      ctx.fill();
      ctx.strokeStyle = "rgba(212,181,106,0.8)";
      ctx.lineWidth = 2;
      ctx.stroke();
    }
  }

  ctx.fillStyle = "#0a1613";
  ctx.globalAlpha = 0.72;
  ctx.fillRect(16, h - 52, 320, 36);
  ctx.globalAlpha = 1;
  ctx.fillStyle = "#eef6f2";
  ctx.font = "700 14px Manrope";
  ctx.fillText(`${cityDefName()} · ×${tileScale().toFixed(2)} · тап = участок`, 26, h - 28);
}

function cityDefName() {
  return (state.catalog?.cities || []).find((c) => c.id === state.city)?.name || state.city;
}

function drawTile(ctx, cx, cy, color, selected, tw, th) {
  ctx.beginPath();
  ctx.moveTo(cx, cy - th / 2);
  ctx.lineTo(cx + tw / 2, cy);
  ctx.lineTo(cx, cy + th / 2);
  ctx.lineTo(cx - tw / 2, cy);
  ctx.closePath();
  ctx.fillStyle = selected ? shade(color, 1.25) : color;
  ctx.fill();
}

function drawBuilding(ctx, cx, cy, tall, tw = 72) {
  const h = tall ? tw * 0.55 : tw * 0.35;
  const w = tw * 0.42;
  ctx.fillStyle = tall ? "#9AA3AD" : "#D9C4A0";
  ctx.fillRect(cx - w / 2, cy - h, w, h);
}

function shade(hex, f) {
  const n = parseInt(hex.slice(1), 16);
  return `rgb(${Math.min(255, ((n >> 16) & 255) * f) | 0},${Math.min(255, ((n >> 8) & 255) * f) | 0},${Math.min(255, (n & 255) * f) | 0})`;
}

function pickTile(mx, my) {
  const city = state.world?.cities?.[state.city];
  if (!city) return null;
  const { tw, th } = isoMetrics();
  const ox = $("map").width / 2, oy = $("map").height * 0.36;
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
  let info = `Участок (${p.x},${p.y}) · ${p.zone}`;
  if (dep) info += ` · залежь ${itemName(dep.item_id)} ${dep.amount}/${dep.max_amount}`;
  if (bld) info += ` · ${buildingName(bld.type)}`;
  if (shop) info += ` · ${shopName(shop.shop_type)}: ${itemName(shop.item_id)}`;
  if (state.place) info += " · готово к подтверждению";
  $("tileInfo").textContent = info;
  drawMap();
  if (!state.place) {
    if (bld && bld.firm_id === state.firmId) openBuildingManage(bld.id);
    else if (shop && shop.firm_id === state.firmId) openShopManage(shop.id);
  }
});

$("closeSheet").onclick = () => { $("manageSheet").hidden = true; state.manage = null; };
$("closeCityPick").onclick = () => { $("cityPickSheet").hidden = true; };
$("cancelPlace").onclick = () => { exitPlaceMode(); setTab("desk"); };
$("confirmPlace").onclick = () => confirmPlace().catch((e) => alert(e.message));
$("backDesktop").onclick = () => { exitPlaceMode(); setTab("desk"); };
$("startBuildPlace").onclick = startBuildPlace;
$("startShopPlace").onclick = startShopPlace;
$("goBuild").onclick = () => { $("buildBlock").scrollIntoView({ behavior: "smooth", block: "start" }); };
$("goShop").onclick = () => { $("shopBlock").scrollIntoView({ behavior: "smooth", block: "start" }); };
$("goCity").onclick = () => { exitPlaceMode(); setTab("map"); drawMap(); };

document.querySelectorAll(".tab").forEach((el) => {
  el.onclick = () => {
    if (el.dataset.tab !== "map") exitPlaceMode();
    setTab(el.dataset.tab);
  };
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
    const ind = $("industry").value;
    if (ind === "steel") state.city = "beloretsk";
    else if (ind === "oil") state.city = enabledCities().some((c) => c.id === "sterlitamak") ? "sterlitamak" : "beloretsk";
    else state.city = "ufa";
    fillCityChips();
    setTab("desk");
    await refresh();
  } catch (err) { alert(err.message); }
};

async function act(path, body) {
  try {
    await api("/api/action/" + path, { method: "POST", body: JSON.stringify(body) });
    await refresh();
  } catch (err) { alert(err.message); }
}

$("buyParcel").onclick = () => state.selected && act("buy_parcel", { city_id: state.city, x: state.selected.x, y: state.selected.y });
$("develop").onclick = () => state.selected && act("develop_housing", { city_id: state.city, x: state.selected.x, y: state.selected.y });
$("truck").onclick = () => act("buy_truck", {});
$("refresh").onclick = () => refresh().catch((e) => alert(e.message));
$("tickBtn").onclick = async () => {
  try { await api("/api/tick", { method: "POST", body: "{}" }); await refresh(); }
  catch (e) { alert(e.message); }
};
$("quality").value = String(state.quality);
$("quality").onchange = () => {
  state.quality = Number($("quality").value);
  localStorage.setItem("bc_quality", String(state.quality));
  drawMap();
};
$("subscribe").onclick = () => act("subscribe", { months: 1 });
if ($("subscribePay")) {
  $("subscribePay").onclick = async () => {
    if (!state.firmId) return alert("Сначала откройте фирму");
    try {
      const pay = await api("/api/payment/create", {
        method: "POST",
        body: JSON.stringify({ months: 1, return_url: location.href }),
      });
      if (pay.confirmation_url) {
        if (pay.mode === "mock") {
          await fetch(pay.confirmation_url);
          await refresh();
          alert("Подписка активирована (mock-pay / -dev)");
        } else {
          location.href = pay.confirmation_url;
        }
      } else {
        alert("Нет confirmation_url");
      }
    } catch (e) {
      alert(e.message);
    }
  };
}
$("hireDeputy").onclick = () => act("hire_deputy", { name: $("deputyName").value || "Заместитель" });
$("applyTasks").onclick = () => {
  const tasks = [...document.querySelectorAll(".task:checked")].map((el) => el.value);
  act("set_deputy_tasks", { tasks });
};
if ($("ordPrice")) {
  $("ordPrice").addEventListener("input", () => { $("ordPrice").dataset.touched = "1"; });
}
if ($("ordItem")) {
  $("ordItem").onchange = () => {
    const mid = retailPriceForItem($("ordItem").value);
    if (mid > 0 && !$("ordPrice").dataset.touched) $("ordPrice").value = String(mid);
  };
}
if ($("ordSubmit")) {
  $("ordSubmit").onclick = () => {
    if (!state.firmId) return alert("Сначала откройте фирму");
    const qty = Number($("ordQty").value);
    const price = Number($("ordPrice").value);
    if (!qty || !price) return alert("Укажите количество и цену");
    act("order", {
      side: $("ordSide").value,
      city_id: $("ordCity").value || state.city,
      item_id: $("ordItem").value,
      qty, price,
    });
  };
}

// Deep-link helpers for demos / screenshots: ?tab=map&city=ufa&place=1
(function applyBootQuery() {
  const q = new URLSearchParams(location.search);
  if (q.get("city")) state.city = q.get("city");
  if (q.get("tab") === "map" || q.get("place") === "1") state.tab = "map";
})();

setTab(state.tab || "desk");
refresh().then(() => {
  const q = new URLSearchParams(location.search);
  if (q.get("place") === "1") {
    state.place = {
      kind: "build",
      buildingType: q.get("bld") || "farm",
      recipeId: q.get("recipe") || "",
    };
    enterPlaceMode();
  }
}).catch(() => {});
