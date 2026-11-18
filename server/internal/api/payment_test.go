package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bashcapital/server/internal/model"
	"bashcapital/server/internal/sim"
	"bashcapital/server/internal/store"
)

func testEngine() *sim.Engine {
	return &sim.Engine{
		Catalog: &model.Catalog{Economy: model.EconomyConfig{}},
		World: &model.World{
			Firms: map[string]*model.Firm{}, Buildings: map[string]*model.BuildingInst{},
			Shops: map[string]*model.ShopLease{}, Cities: map[string]*model.CityState{},
			Orders: map[string]*model.MarketOrder{}, Contracts: map[string]*model.Contract{},
			Shipments: map[string]*model.Shipment{}, Events: []string{},
		},
	}
}

func TestMockPayGatedWithoutDev(t *testing.T) {
	os.Unsetenv("ALLOW_MOCK_PAY")
	dir := t.TempDir()
	db, err := store.OpenDB(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	eng := testEngine()
	eng.World.Firms["f1"] = &model.Firm{ID: "f1", Name: "T", Cash: 1e6, BrandAwareness: map[string]float64{}}
	srv := New(eng, 0, true, "tok", db)
	srv.Dev = false

	pay := store.Payment{
		ID: "p1", FirmID: "f1", Months: 1, Amount: 490, Currency: "RUB",
		Status: "pending", Provider: "yookassa", CreatedAt: 1, UpdatedAt: 1,
	}
	if err := db.SavePayment(pay); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/payment/mock-pay?id=p1", nil)
	rr := httptest.NewRecorder()
	srv.handlePaymentMockPay(rr, req)
	if rr.Code != 403 {
		t.Fatalf("expected 403 without -dev, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestMockPayWorksInDev(t *testing.T) {
	os.Unsetenv("ALLOW_MOCK_PAY")
	dir := t.TempDir()
	db, err := store.OpenDB(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	eng := testEngine()
	eng.World.Firms["f1"] = &model.Firm{ID: "f1", Name: "T", Cash: 1e6, BrandAwareness: map[string]float64{}, Researched: map[string]bool{}}
	srv := New(eng, 0, true, "tok", db)
	srv.Dev = true
	_ = db.SavePayment(store.Payment{
		ID: "p2", FirmID: "f1", Months: 1, Amount: 490, Currency: "RUB",
		Status: "pending", Provider: "yookassa", CreatedAt: 1, UpdatedAt: 1,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/payment/mock-pay?id=p2", nil)
	rr := httptest.NewRecorder()
	srv.handlePaymentMockPay(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d %s", rr.Code, rr.Body.String())
	}
	if !eng.World.Firms["f1"].SubscriptionActive {
		t.Fatal("subscription should activate after mock-pay")
	}
}

func TestWebhookRequiresSecretOutsideDev(t *testing.T) {
	os.Unsetenv("YOOKASSA_WEBHOOK_SECRET")
	os.Unsetenv("YOOKASSA_SECRET")
	os.Unsetenv("YOOKASSA_SHOP_ID")
	os.Unsetenv("YOOKASSA_SECRET_KEY")
	os.Unsetenv("ALLOW_MOCK_PAY")
	dir := t.TempDir()
	db, err := store.OpenDB(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	eng := testEngine()
	srv := New(eng, 0, true, "tok", db)
	srv.Dev = false

	req := httptest.NewRequest(http.MethodPost, "/api/payment/webhook/yookassa", strings.NewReader(`{"payment_id":"x","status":"succeeded"}`))
	rr := httptest.NewRecorder()
	srv.handleYooKassaWebhook(rr, req)
	if rr.Code != 503 {
		t.Fatalf("expected 503 without secret, got %d %s", rr.Code, rr.Body.String())
	}
}

func TestWebhookWithSecretActivates(t *testing.T) {
	t.Setenv("YOOKASSA_WEBHOOK_SECRET", "whsec")
	os.Unsetenv("YOOKASSA_SHOP_ID")
	os.Unsetenv("YOOKASSA_SECRET_KEY")
	dir := t.TempDir()
	db, err := store.OpenDB(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	eng := testEngine()
	eng.World.Firms["f1"] = &model.Firm{ID: "f1", Name: "T", Cash: 1e6, BrandAwareness: map[string]float64{}, Researched: map[string]bool{}}
	srv := New(eng, 0, true, "tok", db)
	srv.Dev = false
	_ = db.SavePayment(store.Payment{
		ID: "p3", FirmID: "f1", Months: 2, Amount: 980, Currency: "RUB",
		Status: "pending", Provider: "yookassa", CreatedAt: 1, UpdatedAt: 1,
	})

	body := `{"event":"payment.succeeded","object":{"id":"yk_1","status":"succeeded","metadata":{"payment_id":"p3","firm_id":"f1","months":2}}}`
	req := httptest.NewRequest(http.MethodPost, "/api/payment/webhook/yookassa", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer whsec")
	rr := httptest.NewRecorder()
	srv.handleYooKassaWebhook(rr, req)
	if rr.Code != 200 {
		t.Fatalf("got %d %s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["status"] != "succeeded" {
		t.Fatalf("unexpected: %+v", out)
	}
	if !eng.World.Firms["f1"].SubscriptionActive {
		t.Fatal("expected subscription active")
	}
}

func TestCreatePaymentMockInDev(t *testing.T) {
	os.Unsetenv("YOOKASSA_SHOP_ID")
	os.Unsetenv("YOOKASSA_SECRET_KEY")
	dir := t.TempDir()
	db, err := store.OpenDB(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	eng := testEngine()
	eng.World.Firms["f1"] = &model.Firm{ID: "f1", Name: "T", Cash: 1e6}
	srv := New(eng, 0, true, "tok", db)
	srv.Dev = true

	req := httptest.NewRequest(http.MethodPost, "/api/payment/create", strings.NewReader(`{"months":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Firm-Id", "f1")
	rr := httptest.NewRecorder()
	srv.handlePaymentCreate(rr, req)
	if rr.Code != 200 {
		t.Fatalf("got %d %s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["mode"] != "mock" {
		t.Fatalf("expected mock mode, got %+v", out)
	}
	url, _ := out["confirmation_url"].(string)
	if !strings.Contains(url, "/api/payment/mock-pay?id=") {
		t.Fatalf("bad confirmation url: %s", url)
	}
}
