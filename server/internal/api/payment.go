package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"bashcapital/server/internal/store"
)

const subPricePerMonth = 490.0

func (s *Server) handlePaymentCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	firmID := s.resolveFirmID(r)
	if firmID == "" {
		http.Error(w, "X-Firm-Id or session required", 401)
		return
	}
	var body struct {
		Months int `json:"months"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Months <= 0 {
		body.Months = 1
	}
	if body.Months > 12 {
		body.Months = 12
	}
	s.Mu.RLock()
	firm := s.Engine.World.Firms[firmID]
	s.Mu.RUnlock()
	if firm == nil || firm.IsAI {
		http.Error(w, "firm not found", 404)
		return
	}
	accID := ""
	if acc, _, err := s.sessionFromRequest(r); err == nil {
		accID = acc.ID
	}
	now := time.Now().Unix()
	pay := store.Payment{
		ID: uuid.NewString(), FirmID: firmID, AccountID: accID,
		Months: body.Months, Amount: subPricePerMonth * float64(body.Months),
		Currency: "RUB", Status: "pending", Provider: "yookassa",
		CreatedAt: now, UpdatedAt: now,
	}
	base := os.Getenv("PUBLIC_API_URL")
	if base == "" {
		base = "http://127.0.0.1:8080"
	}
	// Stub confirmation URL — in production replace with YooKassa Payment create response.
	pay.ConfirmationURL = fmt.Sprintf("%s/api/payment/mock-pay?id=%s", base, pay.ID)
	if s.DB != nil {
		if err := s.DB.SavePayment(pay); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	writeJSON(w, map[string]any{
		"payment_id":       pay.ID,
		"amount":           pay.Amount,
		"currency":         pay.Currency,
		"months":           pay.Months,
		"status":           pay.Status,
		"confirmation_url": pay.ConfirmationURL,
		"provider":         "yookassa",
		"note":             "Webhook: POST /api/payment/webhook/yookassa — set YOOKASSA_SECRET when going live",
	})
}

func (s *Server) handlePaymentGet(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "payments unavailable", 503)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", 400)
		return
	}
	p, err := s.DB.PaymentByID(id)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	writeJSON(w, p)
}

// handleYooKassaWebhook accepts YooKassa-like notification payloads.
// Without YOOKASSA_SECRET it accepts stub events for local/dev.
func (s *Server) handleYooKassaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	secret := os.Getenv("YOOKASSA_SECRET")
	if secret != "" && r.Header.Get("X-Yookassa-Signature") != secret && r.Header.Get("Authorization") != "Bearer "+secret {
		http.Error(w, "invalid signature", 401)
		return
	}
	var body struct {
		Event  string `json:"event"`
		Object struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Metadata struct {
				PaymentID string `json:"payment_id"`
				FirmID    string `json:"firm_id"`
				Months    int    `json:"months"`
			} `json:"metadata"`
		} `json:"object"`
		// Local stub fields
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
	}
	raw, _ := json.Marshal(map[string]any{})
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	_ = raw
	payID := body.Object.Metadata.PaymentID
	if payID == "" {
		payID = body.PaymentID
	}
	if payID == "" {
		payID = body.Object.ID
	}
	status := body.Object.Status
	if status == "" {
		status = body.Status
	}
	if status == "" && (body.Event == "payment.succeeded" || body.Event == "") {
		status = "succeeded"
	}
	if s.DB == nil {
		http.Error(w, "db required", 503)
		return
	}
	p, err := s.DB.PaymentByID(payID)
	if err != nil {
		http.Error(w, "payment not found", 404)
		return
	}
	if status == "succeeded" || status == "waiting_for_capture" {
		p.Status = "succeeded"
		p.UpdatedAt = time.Now().Unix()
		_ = s.DB.SavePayment(*p)
		s.Mu.Lock()
		_ = s.Engine.ActivateSubscriptionPaid(p.FirmID, p.Months)
		s.Mu.Unlock()
		writeJSON(w, map[string]any{"ok": true, "status": "succeeded"})
		return
	}
	if status == "canceled" {
		p.Status = "canceled"
		p.UpdatedAt = time.Now().Unix()
		_ = s.DB.SavePayment(*p)
	}
	writeJSON(w, map[string]any{"ok": true, "status": p.Status})
}

// Dev helper: GET mock-pay completes pending payment without YooKassa.
func (s *Server) handlePaymentMockPay(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" || s.DB == nil {
		http.Error(w, "id required", 400)
		return
	}
	p, err := s.DB.PaymentByID(id)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	if p.Status != "pending" {
		writeJSON(w, map[string]any{"ok": true, "status": p.Status})
		return
	}
	p.Status = "succeeded"
	p.UpdatedAt = time.Now().Unix()
	_ = s.DB.SavePayment(*p)
	s.Mu.Lock()
	_ = s.Engine.ActivateSubscriptionPaid(p.FirmID, p.Months)
	s.Mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "status": "succeeded", "firm_id": p.FirmID})
}
