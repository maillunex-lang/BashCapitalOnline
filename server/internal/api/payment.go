package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
		Months    int    `json:"months"`
		ReturnURL string `json:"return_url"`
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
	returnURL := body.ReturnURL
	if returnURL == "" {
		returnURL = os.Getenv("PAYMENT_RETURN_URL")
	}
	if returnURL == "" {
		returnURL = base + "/api/payment/status?id=" + pay.ID
	}

	_, _, live := yooKassaCreds()
	mode := "mock"
	if live {
		api := newYooAPI()
		meta := map[string]string{
			"payment_id": pay.ID,
			"firm_id":    firmID,
			"months":     strconv.Itoa(body.Months),
		}
		yp, err := api.createPayment(pay.ID, pay.Amount, pay.Currency, returnURL,
			fmt.Sprintf("BashCapital подписка %d мес. · %s", body.Months, firm.Name), meta)
		if err != nil {
			http.Error(w, "yookassa: "+err.Error(), 502)
			return
		}
		pay.ConfirmationURL = yp.Confirmation.ConfirmationURL
		payload, _ := json.Marshal(map[string]any{"yookassa_id": yp.ID, "status": yp.Status})
		pay.ProviderPayload = string(payload)
		mode = "live"
	} else {
		if !mockPayAllowed(s.Dev) {
			http.Error(w, "payments require YOOKASSA_SHOP_ID + YOOKASSA_SECRET_KEY (or run -dev / ALLOW_MOCK_PAY=1)", 503)
			return
		}
		pay.ConfirmationURL = fmt.Sprintf("%s/api/payment/mock-pay?id=%s", base, pay.ID)
		payload, _ := json.Marshal(map[string]any{"mode": "mock"})
		pay.ProviderPayload = string(payload)
	}

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
		"mode":             mode,
		"note":             "Live: webhook POST /api/payment/webhook/yookassa with YOOKASSA_WEBHOOK_SECRET / API secret. Mock only in -dev or ALLOW_MOCK_PAY=1.",
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

// handleYooKassaWebhook — YooKassa notification.
// Non-dev without webhook secret: reject. With live creds: re-fetch payment status from API.
func (s *Server) handleYooKassaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", 405)
		return
	}
	secret := yooKassaWebhookSecret()
	_, _, live := yooKassaCreds()
	if secret != "" {
		if !webhookAuthorized(r, secret) {
			http.Error(w, "invalid signature", 401)
			return
		}
	} else if !s.Dev && !mockPayAllowed(false) {
		http.Error(w, "webhook secret required (YOOKASSA_WEBHOOK_SECRET or YOOKASSA_SECRET)", 503)
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
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	payID := body.Object.Metadata.PaymentID
	if payID == "" {
		payID = body.PaymentID
	}
	providerID := body.Object.ID
	status := body.Object.Status
	if status == "" {
		status = body.Status
	}
	if status == "" && (body.Event == "payment.succeeded" || body.Event == "") {
		status = "succeeded"
	}

	// Live path: trust API status over notification body.
	if live && providerID != "" {
		yp, err := newYooAPI().getPayment(providerID)
		if err != nil {
			http.Error(w, "yookassa verify: "+err.Error(), 502)
			return
		}
		status = yp.Status
		if payID == "" && yp.Metadata != nil {
			payID = yp.Metadata["payment_id"]
		}
	}
	if payID == "" {
		http.Error(w, "payment_id required", 400)
		return
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
	if p.Status == "succeeded" {
		writeJSON(w, map[string]any{"ok": true, "status": "succeeded", "idempotent": true})
		return
	}
	if status == "succeeded" || status == "waiting_for_capture" {
		p.Status = "succeeded"
		p.UpdatedAt = time.Now().Unix()
		if providerID != "" {
			payload, _ := json.Marshal(map[string]any{"yookassa_id": providerID, "status": status})
			p.ProviderPayload = string(payload)
		}
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
// Gated: only -dev or ALLOW_MOCK_PAY=1.
func (s *Server) handlePaymentMockPay(w http.ResponseWriter, r *http.Request) {
	if !mockPayAllowed(s.Dev) {
		http.Error(w, "mock-pay disabled (set ALLOW_MOCK_PAY=1 or run with -dev)", 403)
		return
	}
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
	writeJSON(w, map[string]any{"ok": true, "status": "succeeded", "firm_id": p.FirmID, "mode": "mock"})
}
