package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"bashcapital/server/internal/model"
)

// DB wraps SQLite persistence for world, accounts, sessions, payments.
type DB struct {
	sql *sql.DB
	path string
}

func OpenDB(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	d := &DB{sql: sqlDB, path: path}
	if err := d.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Close() error {
	if d == nil || d.sql == nil {
		return nil
	}
	return d.sql.Close()
}

func (d *DB) migrate() error {
	_, err := d.sql.Exec(`
CREATE TABLE IF NOT EXISTS meta (k TEXT PRIMARY KEY, v TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS world_save (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  payload BLOB NOT NULL,
  updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS accounts (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  pass_hash TEXT NOT NULL,
  firm_id TEXT,
  created_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  firm_id TEXT,
  expires_at INTEGER NOT NULL,
  FOREIGN KEY(account_id) REFERENCES accounts(id)
);
CREATE TABLE IF NOT EXISTS payments (
  id TEXT PRIMARY KEY,
  firm_id TEXT NOT NULL,
  account_id TEXT,
  months INTEGER NOT NULL,
  amount REAL NOT NULL,
  currency TEXT NOT NULL,
  status TEXT NOT NULL,
  provider TEXT NOT NULL,
  confirmation_url TEXT,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL,
  provider_payload TEXT
);
`)
	return err
}

func (d *DB) SaveWorld(w *model.World) error {
	if d == nil {
		return nil
	}
	b, err := json.Marshal(w)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(
		`INSERT INTO world_save(id, payload, updated_at) VALUES(1, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET payload=excluded.payload, updated_at=excluded.updated_at`,
		b, time.Now().Unix(),
	)
	return err
}

func (d *DB) LoadWorld() (*model.World, error) {
	if d == nil {
		return nil, sql.ErrNoRows
	}
	var payload []byte
	err := d.sql.QueryRow(`SELECT payload FROM world_save WHERE id=1`).Scan(&payload)
	if err != nil {
		return nil, err
	}
	var w model.World
	if err := json.Unmarshal(payload, &w); err != nil {
		return nil, err
	}
	if w.Cities == nil {
		w.Cities = map[string]*model.CityState{}
	}
	if w.Firms == nil {
		w.Firms = map[string]*model.Firm{}
	}
	if w.Buildings == nil {
		w.Buildings = map[string]*model.BuildingInst{}
	}
	if w.Shops == nil {
		w.Shops = map[string]*model.ShopLease{}
	}
	if w.Orders == nil {
		w.Orders = map[string]*model.MarketOrder{}
	}
	if w.Contracts == nil {
		w.Contracts = map[string]*model.Contract{}
	}
	if w.Shipments == nil {
		w.Shipments = map[string]*model.Shipment{}
	}
	if w.CityMarket == nil {
		w.CityMarket = map[string]model.CityMarketStock{}
	}
	return &w, nil
}

func (d *DB) HasWorld() bool {
	if d == nil {
		return false
	}
	var n int
	_ = d.sql.QueryRow(`SELECT COUNT(1) FROM world_save WHERE id=1`).Scan(&n)
	return n > 0
}

type Account struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	PassHash  string `json:"-"`
	FirmID    string `json:"firm_id,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

type Session struct {
	Token     string `json:"token"`
	AccountID string `json:"account_id"`
	FirmID    string `json:"firm_id,omitempty"`
	ExpiresAt int64  `json:"expires_at"`
}

type Payment struct {
	ID               string  `json:"id"`
	FirmID           string  `json:"firm_id"`
	AccountID        string  `json:"account_id,omitempty"`
	Months           int     `json:"months"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"` // pending|succeeded|canceled
	Provider         string  `json:"provider"`
	ConfirmationURL  string  `json:"confirmation_url,omitempty"`
	CreatedAt        int64   `json:"created_at"`
	UpdatedAt        int64   `json:"updated_at"`
	ProviderPayload  string  `json:"provider_payload,omitempty"`
}

func (d *DB) CreateAccount(a Account) error {
	_, err := d.sql.Exec(
		`INSERT INTO accounts(id, username, pass_hash, firm_id, created_at) VALUES(?,?,?,?,?)`,
		a.ID, a.Username, a.PassHash, nullStr(a.FirmID), a.CreatedAt,
	)
	return err
}

func (d *DB) AccountByUsername(username string) (*Account, error) {
	var a Account
	var firm sql.NullString
	err := d.sql.QueryRow(
		`SELECT id, username, pass_hash, firm_id, created_at FROM accounts WHERE username=?`, username,
	).Scan(&a.ID, &a.Username, &a.PassHash, &firm, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	if firm.Valid {
		a.FirmID = firm.String
	}
	return &a, nil
}

func (d *DB) AccountByID(id string) (*Account, error) {
	var a Account
	var firm sql.NullString
	err := d.sql.QueryRow(
		`SELECT id, username, pass_hash, firm_id, created_at FROM accounts WHERE id=?`, id,
	).Scan(&a.ID, &a.Username, &a.PassHash, &firm, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	if firm.Valid {
		a.FirmID = firm.String
	}
	return &a, nil
}

func (d *DB) LinkFirm(accountID, firmID string) error {
	_, err := d.sql.Exec(`UPDATE accounts SET firm_id=? WHERE id=?`, firmID, accountID)
	return err
}

func (d *DB) CreateSession(s Session) error {
	_, err := d.sql.Exec(
		`INSERT INTO sessions(token, account_id, firm_id, expires_at) VALUES(?,?,?,?)`,
		s.Token, s.AccountID, nullStr(s.FirmID), s.ExpiresAt,
	)
	return err
}

func (d *DB) SessionByToken(token string) (*Session, error) {
	var s Session
	var firm sql.NullString
	err := d.sql.QueryRow(
		`SELECT token, account_id, firm_id, expires_at FROM sessions WHERE token=?`, token,
	).Scan(&s.Token, &s.AccountID, &firm, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	if firm.Valid {
		s.FirmID = firm.String
	}
	if time.Now().Unix() > s.ExpiresAt {
		return nil, fmt.Errorf("session expired")
	}
	return &s, nil
}

func (d *DB) UpdateSessionFirm(token, firmID string) error {
	_, err := d.sql.Exec(`UPDATE sessions SET firm_id=? WHERE token=?`, firmID, token)
	return err
}

func (d *DB) SavePayment(p Payment) error {
	_, err := d.sql.Exec(
		`INSERT INTO payments(id, firm_id, account_id, months, amount, currency, status, provider, confirmation_url, created_at, updated_at, provider_payload)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
		 ON CONFLICT(id) DO UPDATE SET
		   status=excluded.status, updated_at=excluded.updated_at, provider_payload=excluded.provider_payload,
		   confirmation_url=excluded.confirmation_url`,
		p.ID, p.FirmID, nullStr(p.AccountID), p.Months, p.Amount, p.Currency, p.Status, p.Provider,
		nullStr(p.ConfirmationURL), p.CreatedAt, p.UpdatedAt, nullStr(p.ProviderPayload),
	)
	return err
}

func (d *DB) PaymentByID(id string) (*Payment, error) {
	var p Payment
	var acc, url, payload sql.NullString
	err := d.sql.QueryRow(
		`SELECT id, firm_id, account_id, months, amount, currency, status, provider, confirmation_url, created_at, updated_at, provider_payload
		 FROM payments WHERE id=?`, id,
	).Scan(&p.ID, &p.FirmID, &acc, &p.Months, &p.Amount, &p.Currency, &p.Status, &p.Provider, &url, &p.CreatedAt, &p.UpdatedAt, &payload)
	if err != nil {
		return nil, err
	}
	if acc.Valid {
		p.AccountID = acc.String
	}
	if url.Valid {
		p.ConfirmationURL = url.String
	}
	if payload.Valid {
		p.ProviderPayload = payload.String
	}
	return &p, nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func DefaultDBPath(dataDir string) string {
	return filepath.Join(dataDir, "saves", "bashcapital.db")
}
