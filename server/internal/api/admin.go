package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"bashcapital/server/internal/model"
)

func (s *Server) requireAdmin(r *http.Request) bool {
	if s.AdminToken == "" {
		// unset token: allow in local/dev with header "dev" or empty for GET status only is risky —
		// require explicit token always; default "bashcapital-dev" set in main if env empty.
		return false
	}
	tok := r.Header.Get("X-Admin-Token")
	if tok == "" {
		tok = r.URL.Query().Get("token")
	}
	return tok == s.AdminToken
}

func (s *Server) handleAdminStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(r) {
		http.Error(w, "admin token required", 401)
		return
	}
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	writeJSON(w, s.Engine.AdminStatus())
}

func (s *Server) handleAdminTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !s.requireAdmin(r) {
		http.Error(w, "POST + admin token", 401)
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

func (s *Server) handleAdminPause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !s.requireAdmin(r) {
		http.Error(w, "POST + admin token", 401)
		return
	}
	var body struct {
		Paused bool `json:"paused"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	s.Mu.Lock()
	s.Engine.SetPaused(body.Paused)
	paused := s.Engine.Paused
	s.Mu.Unlock()
	writeJSON(w, map[string]any{"paused": paused})
}

func (s *Server) handleAdminEconomy(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(r) {
		http.Error(w, "admin token required", 401)
		return
	}
	if r.Method == http.MethodGet {
		s.Mu.RLock()
		defer s.Mu.RUnlock()
		writeJSON(w, s.Engine.Catalog.Economy)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "GET|POST", 405)
		return
	}
	var body struct {
		PriceCorridors map[string]model.PriceCorridor `json:"price_corridors"`
		CityMarket     []model.CityMarketOffer        `json:"city_market"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Mu.Lock()
	err := s.Engine.UpdateEconomy(body.PriceCorridors, body.CityMarket)
	s.Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleAdminCities(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(r) {
		http.Error(w, "admin token required", 401)
		return
	}
	if r.Method == http.MethodGet {
		s.Mu.RLock()
		defer s.Mu.RUnlock()
		writeJSON(w, s.Engine.Catalog.Cities)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "GET|POST", 405)
		return
	}
	var body struct {
		CityID  string `json:"city_id"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Mu.Lock()
	err := s.Engine.SetCityEnabled(body.CityID, body.Enabled)
	s.Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleAdminRecipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !s.requireAdmin(r) {
		http.Error(w, "POST + admin token", 401)
		return
	}
	var body struct {
		RecipeID     string `json:"recipe_id"`
		ResearchCost int    `json:"research_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Mu.Lock()
	err := s.Engine.UpdateRecipeResearch(body.RecipeID, body.ResearchCost)
	s.Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleUpdatesManifest(w http.ResponseWriter, r *http.Request) {
	path := "remote_manifest.json"
	if s.Engine.DataDir != "" {
		path = s.Engine.DataDir + "/remote_manifest.json"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		// fallback relative
		b, err = os.ReadFile("../data/remote_manifest.json")
		if err != nil {
			http.Error(w, "manifest missing", 404)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(b)
}

func (s *Server) handlePackDownload(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/api/packs/"):]
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
		http.Error(w, "bad pack name", 400)
		return
	}
	path := name
	if s.Engine.DataDir != "" {
		path = s.Engine.DataDir + "/packs/" + name
	}
	b, err := os.ReadFile(path)
	if err != nil {
		b, err = os.ReadFile("../data/packs/" + name)
		if err != nil {
			http.Error(w, "pack missing", 404)
			return
		}
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	_, _ = w.Write(b)
}

func (s *Server) handleMapAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	s.Mu.Lock()
	s.Engine.ClearDirty()
	rev := s.Engine.World.MapRevision
	s.Mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "map_revision": rev})
}
