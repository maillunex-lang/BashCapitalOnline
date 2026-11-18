package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"bashcapital/server/internal/store"
)

func hashPassword(pw, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + pw))
	return hex.EncodeToString(sum[:])
}

func newToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Server) handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	if s.DB == nil {
		http.Error(w, "auth unavailable", 503)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	user := strings.TrimSpace(body.Username)
	if len(user) < 3 || len(body.Password) < 4 {
		http.Error(w, "username>=3, password>=4", 400)
		return
	}
	salt := newToken()[:16]
	acc := store.Account{
		ID: uuid.NewString(), Username: user,
		PassHash:  salt + "$" + hashPassword(body.Password, salt),
		CreatedAt: time.Now().Unix(),
	}
	if err := s.DB.CreateAccount(acc); err != nil {
		http.Error(w, "username taken", 409)
		return
	}
	sess := store.Session{
		Token: newToken(), AccountID: acc.ID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	if err := s.DB.CreateSession(sess); err != nil {
		http.Error(w, "не удалось открыть сессию", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"account_id": acc.ID, "username": acc.Username, "session_token": sess.Token,
	})
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	if s.DB == nil {
		http.Error(w, "auth unavailable", 503)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	acc, err := s.DB.AccountByUsername(strings.TrimSpace(body.Username))
	if err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	parts := strings.SplitN(acc.PassHash, "$", 2)
	if len(parts) != 2 || hashPassword(body.Password, parts[0]) != parts[1] {
		http.Error(w, "invalid credentials", 401)
		return
	}
	sess := store.Session{
		Token: newToken(), AccountID: acc.ID, FirmID: acc.FirmID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	if err := s.DB.CreateSession(sess); err != nil {
		http.Error(w, "не удалось открыть сессию", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"account_id": acc.ID, "username": acc.Username, "firm_id": acc.FirmID,
		"session_token": sess.Token,
	})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	acc, _, err := s.sessionFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	writeJSON(w, map[string]any{
		"account_id": acc.ID, "username": acc.Username, "firm_id": acc.FirmID,
	})
}

func (s *Server) sessionFromRequest(r *http.Request) (*store.Account, *store.Session, error) {
	if s.DB == nil {
		return nil, nil, sql.ErrConnDone
	}
	tok := r.Header.Get("X-Session-Token")
	if tok == "" {
		tok = r.URL.Query().Get("session")
	}
	if tok == "" {
		return nil, nil, sql.ErrNoRows
	}
	sess, err := s.DB.SessionByToken(tok)
	if err != nil {
		return nil, nil, err
	}
	acc, err := s.DB.AccountByID(sess.AccountID)
	if err != nil {
		return nil, nil, err
	}
	return acc, sess, nil
}

func (s *Server) resolveFirmID(r *http.Request) string {
	if r.Header.Get("X-Session-Token") != "" || r.URL.Query().Get("session") != "" {
		acc, _, err := s.sessionFromRequest(r)
		if err != nil {
			return ""
		}
		return acc.FirmID
	}
	firmID := r.Header.Get("X-Firm-Id")
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	firm := s.Engine.World.Firms[firmID]
	if firm == nil || firm.IsAI || firm.AccountID != "" {
		return ""
	}
	return firmID
}
