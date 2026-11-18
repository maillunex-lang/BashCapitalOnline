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

function tileScale() {
  // Larger when placing or on high quality
  const q = [0.85, 1, 1.15, 1.3][state.quality] || 1.15;
  const placeBoost = state.place ? 1.18 : 1.05;
  return q * placeBoost;
}

function isoMetrics() {
  const s = tileScale();
  return { tw: 72 * s, th: 36 * s, sprite: 0.72 * s };
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

  const orders = Object.values(w.orders || {}).filter((o) => o.qty > o.filled);
  if ($("orders")) {
    $("orders").innerHTML = orders.slice(0, 40).map((o) => {
      const firm = w.firms[o.firm_id];
      return `<div class="row">${rowWithIcon(itemIconSrc(o.item_id), `<strong>${o.side === "sell" ? "Продажа" : "Покупка"}</strong> ${itemName(o.item_id)} ×${o.qty - o.filled}<br/>${firm?.name || o.firm_id} · ${money(o.price)} ₽`)}</div>`;
    }).join("") || "<div class='row'>Биржа пуста</div>";
  }
  const ships = Object.values(w.shipments || {});
  if ($("shipments")) {
    $("shipments").innerHTML = "<h3 style='color:var(--accent-2);font-size:13px'>В пути</h3>" + (ships.map((s) =>
      `<div class="row">${rowWithIcon(itemIconSrc(s.item_id), `<strong>${itemName(s.item_id)}</strong> ×${s.qty}<br/>${s.from_city} → ${s.to_city} · ${s.ticks_left} тик`)}</div>`
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

function drawMap() {
  const canvas = $("map");
  if (!canvas || canvas.offsetParent === null && state.tab !== "map") {
    // still draw when visible
  }
  const ctx = canvas.getContext("2d");
  const w = canvas.width, h = canvas.height;
  ctx.clearRect(0, 0, w, h);

  // atmospheric ground
  const g = ctx.createRadialGradient(w * 0.5, h * 0.35, 40, w * 0.5, h * 0.5, w * 0.7);
  g.addColorStop(0, "rgba(70, 120, 100, 0.45)");
  g.addColorStop(0.55, "rgba(28, 55, 48, 0.55)");
  g.addColorStop(1, "rgba(12, 22, 20, 0.75)");
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, w, h);

  const city = state.world?.cities?.[state.city];
  if (!city) {
    ctx.fillStyle = "rgba(238,246,242,0.7)";
    ctx.font = "600 22px Manrope";
    ctx.fillText("Загрузка города…", 40, 60);
    return;
  }

  const { tw, th, sprite } = isoMetrics();
  const ox = w / 2, oy = h * 0.38;
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
    const tile = mapImg(ZONE_TILE[p.zone] || "tile_grass");
    if (tile?.complete && tile.naturalWidth) {
      const s = sprite * 0.95;
      ctx.globalAlpha = selected ? 1 : 0.96;
      ctx.drawImage(tile, cx - tile.width * s * 0.5, cy - tile.height * s * 0.62, tile.width * s, tile.height * s);
      ctx.globalAlpha = 1;
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
      ctx.strokeStyle = "rgba(212,181,106,0.95)";
      ctx.lineWidth = 3;
      ctx.stroke();
    }

    const dep = deposits.find((d) => d.x === p.x && d.y === p.y && d.amount > 0);
    if (dep) {
      const di = assetImg("res_" + dep.item_id, "item_" + dep.item_id);
      const size = 40 * tileScale();
      if (di?.complete && di.naturalWidth) ctx.drawImage(di, cx - size / 2, cy - size * 0.15, size, size);
    }
    if (p.has_road) {
      const road = mapImg("tile_road");
      if (road?.complete && road.naturalWidth) {
        const rs = 42 * tileScale();
        ctx.globalAlpha = 0.9;
        ctx.drawImage(road, cx - rs / 2, cy - rs * 0.4, rs, rs * 0.75);
        ctx.globalAlpha = 1;
      }
    }
    const bld = buildings.find((b) => b.x === p.x && b.y === p.y);
    const shopHere = shops.find((s) => s.x === p.x && s.y === p.y);
    if (bld) {
      const bi = assetImg(BUILDING_SPRITE[bld.type] || "", "bld_" + bld.type, buildingDef(bld.type)?.icon);
      if (bi?.complete && bi.naturalWidth) {
        const s = sprite;
        ctx.drawImage(bi, cx - bi.width * s * 0.5, cy - bi.height * s * 0.82, bi.width * s, bi.height * s);
      } else drawBuilding(ctx, cx, cy, p.zone === "industrial");
      // firm badge
      if (bld.firm_id === state.firmId) {
        ctx.fillStyle = "rgba(212,181,106,0.95)";
        ctx.beginPath();
        ctx.arc(cx + 18 * tileScale(), cy - 28 * tileScale(), 5, 0, Math.PI * 2);
        ctx.fill();
      }
    } else if (shopHere) {
      const si = assetImg(shopDef(shopHere.shop_type)?.icon, "shop_" + shopHere.shop_type, "bld_shop");
      if (si?.complete && si.naturalWidth) {
        const s = sprite * 0.92;
        ctx.drawImage(si, cx - si.width * s * 0.5, cy - si.height * s * 0.78, si.width * s, si.height * s);
      }
      const pi = assetImg("item_" + shopHere.item_id, itemDef(shopHere.item_id)?.icon);
      const ps = 28 * tileScale();
      if (pi?.complete && pi.naturalWidth) ctx.drawImage(pi, cx + 8, cy - 4, ps, ps);
    } else if ((p.retail_slots || p.slots || 0) > 0) {
      const shopImg = mapImg("bld_shop");
      if (shopImg?.complete && shopImg.naturalWidth) {
        const s = sprite * 0.7;
        ctx.globalAlpha = 0.55;
        ctx.drawImage(shopImg, cx - shopImg.width * s * 0.5, cy - shopImg.height * s * 0.75, shopImg.width * s, shopImg.height * s);
        ctx.globalAlpha = 1;
      }
    }
  }

  // legend strip
  ctx.fillStyle = "rgba(7,20,18,0.55)";
  ctx.fillRect(16, h - 48, 280, 32);
  ctx.fillStyle = "rgba(238,246,242,0.85)";
  ctx.font = "600 13px Manrope";
  ctx.fillText(`${cityDefName()} · масштаб ×${tileScale().toFixed(2)}`, 26, h - 28);
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

function drawBuilding(ctx, cx, cy, tall) {
  const h = tall ? 36 : 22;
  ctx.fillStyle = tall ? "#9AA3AD" : "#D9C4A0";
  ctx.fillRect(cx - 16, cy - h, 32, h);
}

function shade(hex, f) {
  const n = parseInt(hex.slice(1), 16);
  return `rgb(${Math.min(255, ((n >> 16) & 255) * f) | 0},${Math.min(255, ((n >> 8) & 255) * f) | 0},${Math.min(255, (n & 255) * f) | 0})`;
}

function pickTile(mx, my) {
  const city = state.world?.cities?.[state.city];
  if (!city) return null;
  const { tw, th } = isoMetrics();
  const ox = $("map").width / 2, oy = $("map").height * 0.38;
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
$("hireDeputy").onclick = () => act("hire_deputy", { name: $("deputyName").value || "Заместитель" });
$("applyTasks").onclick = () => {
  const tasks = [...document.querySelectorAll(".task:checked")].map((el) => el.value);
  act("set_deputy_tasks", { tasks });
};

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
