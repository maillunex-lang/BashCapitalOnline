package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"bashcapital/server/internal/model"
	"bashcapital/server/internal/sim"
	"bashcapital/server/internal/store"
)

type Server struct {
	Engine     *sim.Engine
	DB         *store.DB
	Mu         sync.RWMutex
	Upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	tickEvery  time.Duration
	manualOnly bool
	AdminToken string
	Dev        bool // -dev: allows mock-pay without ALLOW_MOCK_PAY
}

func New(engine *sim.Engine, tickEvery time.Duration, manualOnly bool, adminToken string, db *store.DB) *Server {
	return &Server{
		Engine: engine,
		DB:     db,
		Upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		clients:    map[*websocket.Conn]bool{},
		tickEvery:  tickEvery,
		manualOnly: manualOnly,
		AdminToken: adminToken,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/catalog", s.handleCatalog)
	mux.HandleFunc("/api/world", s.handleWorld)
	mux.HandleFunc("/api/rankings", s.handleRankings)
	mux.HandleFunc("/api/register", s.handleRegister)
	mux.HandleFunc("/api/deputy/catalog", s.handleDeputyCatalog)
	mux.HandleFunc("/api/action/", s.handleAction)
	mux.HandleFunc("/api/tick", s.handleManualTick)
	mux.HandleFunc("/api/route", s.handleRouteQuote)
	mux.HandleFunc("/api/auth/register", s.handleAuthRegister)
	mux.HandleFunc("/api/auth/login", s.handleAuthLogin)
	mux.HandleFunc("/api/auth/me", s.handleAuthMe)
	mux.HandleFunc("/api/payment/create", s.handlePaymentCreate)
	mux.HandleFunc("/api/payment/status", s.handlePaymentGet)
	mux.HandleFunc("/api/payment/webhook/yookassa", s.handleYooKassaWebhook)
	mux.HandleFunc("/api/payment/mock-pay", s.handlePaymentMockPay)
	mux.HandleFunc("/api/admin/status", s.handleAdminStatus)
	mux.HandleFunc("/api/admin/tick", s.handleAdminTick)
	mux.HandleFunc("/api/admin/pause", s.handleAdminPause)
	mux.HandleFunc("/api/admin/economy", s.handleAdminEconomy)
	mux.HandleFunc("/api/admin/cities", s.handleAdminCities)
	mux.HandleFunc("/api/admin/recipe", s.handleAdminRecipe)
	mux.HandleFunc("/api/updates/manifest", s.handleUpdatesManifest)
	mux.HandleFunc("/api/packs/", s.handlePackDownload)
	mux.HandleFunc("/api/map/ack", s.handleMapAck)
	mux.HandleFunc("/ws", s.handleWS)
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Firm-Id, X-Admin-Token, X-Session-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleRouteQuote(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		http.Error(w, "from&to required", 400)
		return
	}
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	ticks, cost := s.Engine.RouteQuote(from, to)
	dist := s.Engine.CityDistance(from, to)
	writeJSON(w, map[string]any{
		"from": from, "to": to,
		"ticks": ticks, "cost_per_unit": cost, "distance": dist,
		"min_ticks": 1,
	})
}

func (s *Server) StartTicker() {
	if s.manualOnly || s.tickEvery <= 0 {
		return
	}
	go func() {
		t := time.NewTicker(s.tickEvery)
		for range t.C {
			s.Mu.Lock()
			if !s.Engine.Paused {
				s.Engine.Tick()
			}
			tick := s.Engine.World.Tick
			s.Mu.Unlock()
			s.broadcast(map[string]any{"type": "tick", "tick": tick})
		}
	}()
}

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	c := s.Engine.Catalog
	writeJSON(w, map[string]any{
		"economy_catalog_version": c.Version,
		"items": c.Items, "recipes": c.Recipes, "buildings": c.Buildings, "shops": c.Shops,
		"cities": c.Cities, "routes": c.Routes,
		"economy": c.Economy, "ai_personas": c.AIPersonas, "rnd_tree": c.RnDTree,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ver := 1
	if s.Engine != nil && s.Engine.Catalog != nil && s.Engine.Catalog.Version > 0 {
		ver = s.Engine.Catalog.Version
	}
	tick := 0
	if s.Engine != nil && s.Engine.World != nil {
		tick = s.Engine.World.Tick
	}
	writeJSON(w, map[string]any{"ok": true, "tick": tick, "economy_catalog_version": ver})
}

func (s *Server) handleWorld(w http.ResponseWriter, r *http.Request) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	writeJSON(w, s.Engine.World)
}

func (s *Server) handleRankings(w http.ResponseWriter, r *http.Request) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	type row struct {
		ID      string  `json:"id"`
		Name    string  `json:"name"`
		IsAI    bool    `json:"is_ai"`
		Capital float64 `json:"capital"`
		Cash    float64 `json:"cash"`
	}
	var rows []row
	for _, f := range s.Engine.World.Firms {
		rows = append(rows, row{f.ID, f.Name, f.IsAI, f.CapitalSnapshot, f.Cash})
	}
	writeJSON(w, rows)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	var body struct {
		Name     string `json:"name"`
		Industry string `json:"industry"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	acc, _, authErr := s.sessionFromRequest(r)
	if authErr != nil && (r.Header.Get("X-Session-Token") != "" || r.URL.Query().Get("session") != "") {
		http.Error(w, "недействительная сессия", http.StatusUnauthorized)
		return
	}
	if acc != nil && acc.FirmID != "" {
		http.Error(w, "у аккаунта уже есть фирма", http.StatusConflict)
		return
	}
	eventsBefore := len(s.Engine.World.Events)
	firm, err := s.Engine.CreatePlayer(body.Name, body.Industry)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if acc != nil && s.DB != nil {
		firm.AccountID = acc.ID
		if err := s.DB.SaveWorldAndLinkFirm(s.Engine.World, acc.ID, firm.ID); err != nil {
			delete(s.Engine.World.Firms, firm.ID)
			s.Engine.World.Events = s.Engine.World.Events[:eventsBefore]
			http.Error(w, "не удалось привязать фирму к аккаунту", http.StatusInternalServerError)
			return
		}
	} else {
		s.Engine.Persist() // survive restart before first tick
	}
	writeJSON(w, firm)
}

func (s *Server) handleDeputyCatalog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.Engine.DeputyCatalog())
}

func (s *Server) handleManualTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	s.Mu.Lock()
	s.Engine.Tick()
	tick := s.Engine.World.Tick
	log := s.Engine.World.LastTickLog
	s.Mu.Unlock()
	s.broadcast(map[string]any{"type": "tick", "tick": tick})
	writeJSON(w, map[string]any{"tick": tick, "log": log})
}

func (s *Server) handleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	firmID := s.resolveFirmID(r)
	if firmID == "" {
		http.Error(w, "X-Firm-Id or X-Session-Token required", 401)
		return
	}
	action := r.URL.Path[len("/api/action/"):]
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if _, ok := s.Engine.World.Firms[firmID]; !ok {
		http.Error(w, "unknown firm", 401)
		return
	}
	var (
		result any
		err    error
	)
	switch action {
	case "build":
		result, err = s.Engine.Build(firmID, str(body["city_id"]), str(body["building_type"]), str(body["recipe_id"]), intNum(body["x"]), intNum(body["y"]))
	case "set_recipe":
		err = s.Engine.SetRecipe(firmID, str(body["building_id"]), str(body["recipe_id"]))
	case "configure_building":
		inputs, _ := body["input_sources"].(map[string]any)
		result, err = s.Engine.ConfigureBuilding(
			firmID, str(body["building_id"]), str(body["recipe_id"]),
			boolNum(body["auto_procure"]), str(body["default_input"]), str(body["firm_id"]),
			boolDefaultTrue(body["sell_external"]), inputs,
		)
	case "rent_shop":
		result, err = s.Engine.RentShop(firmID, str(body["city_id"]), intNum(body["x"]), intNum(body["y"]), str(body["shop_type"]), str(body["item_id"]), floatNum(body["price"]))
	case "configure_shop":
		result, err = s.Engine.ConfigureShop(firmID, str(body["shop_id"]), str(body["shop_type"]), str(body["item_id"]), floatNum(body["price"]), str(body["stock_source"]), boolDefaultTrue(body["auto_restock"]))
	case "transfer_stock":
		result, err = s.Engine.TransferStock(firmID, str(body["from_building_id"]), str(body["to_building_id"]), str(body["item_id"]), intNum(body["qty"]))
	case "stock_shop":
		err = s.Engine.StockShop(firmID, str(body["shop_id"]), intNum(body["qty"]))
	case "order":
		result, err = s.Engine.PlaceOrder(firmID, str(body["city_id"]), str(body["item_id"]), str(body["side"]), intNum(body["qty"]), floatNum(body["price"]))
	case "contract":
		result, err = s.Engine.CreateContract(str(body["seller_id"]), firmID, str(body["item_id"]), str(body["city_from"]), str(body["city_to"]), intNum(body["qty"]), floatNum(body["price"]), intNum(body["ticks"]), boolNum(body["use_own_fleet"]))
	case "research":
		err = s.Engine.Research(firmID, str(body["recipe_id"]))
	case "buy_truck":
		err = s.Engine.BuyTruck(firmID)
	case "credit":
		err = s.Engine.TakeCredit(firmID, floatNum(body["amount"]))
	case "buy_parcel":
		err = s.Engine.BuyParcel(firmID, str(body["city_id"]), intNum(body["x"]), intNum(body["y"]))
	case "develop_housing":
		err = s.Engine.DevelopHousing(firmID, str(body["city_id"]), intNum(body["x"]), intNum(body["y"]))
	case "advertise":
		err = s.Engine.Advertise(firmID, str(body["category"]), floatNum(body["budget"]))
	case "subscribe":
		err = s.Engine.ActivateSubscription(firmID, intNum(body["months"]))
		result = s.Engine.World.Firms[firmID]
	case "hire_deputy":
		err = s.Engine.HireDeputy(firmID, str(body["name"]))
		result = s.Engine.World.Firms[firmID]
	case "set_deputy_tasks":
		err = s.Engine.SetDeputyTasks(firmID, strList(body["tasks"]))
		result = s.Engine.World.Firms[firmID]
	case "buy_city_market":
		err = s.Engine.BuyFromCity(firmID, str(body["item_id"]), intNum(body["qty"]))
	case "chat":
		result, err = s.Engine.PostChat(firmID, str(body["channel"]), str(body["text"]))
	default:
		http.Error(w, "unknown action", 404)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Engine.Persist() // persist mutations immediately (build/rent/configure/…)
	if result == nil {
		result = map[string]any{"ok": true}
	}
	writeJSON(w, result)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.Mu.Lock()
	s.clients[conn] = true
	s.Mu.Unlock()
	defer func() {
		s.Mu.Lock()
		delete(s.clients, conn)
		s.Mu.Unlock()
		conn.Close()
	}()
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (s *Server) broadcast(msg any) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	for c := range s.clients {
		_ = c.WriteJSON(msg)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func str(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func intNum(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	default:
		return 0
	}
}

func floatNum(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	default:
		return 0
	}
}

func boolNum(v any) bool {
	b, _ := v.(bool)
	return b
}

func boolDefaultTrue(v any) bool {
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != "false" && t != "0"
	default:
		return true
	}
}

func strList(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, x := range arr {
		if s, ok := x.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Ensure model import used for docs/link
var _ = model.World{}
