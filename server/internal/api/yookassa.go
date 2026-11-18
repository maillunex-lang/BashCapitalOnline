package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// YooKassa credentials from env. Live create requires both shop id and secret key.
func yooKassaCreds() (shopID, secret string, live bool) {
	shopID = strings.TrimSpace(os.Getenv("YOOKASSA_SHOP_ID"))
	secret = strings.TrimSpace(os.Getenv("YOOKASSA_SECRET_KEY"))
	if secret == "" {
		// Back-compat: some deploys used YOOKASSA_SECRET as API key only when SHOP_ID is set.
		// Prefer YOOKASSA_SECRET_KEY for API; YOOKASSA_SECRET remains webhook gate.
	}
	live = shopID != "" && secret != ""
	return shopID, secret, live
}

// YooKassaLive reports whether live Payment create credentials are configured.
func YooKassaLive() (shopID string, live bool) {
	shop, _, ok := yooKassaCreds()
	return shop, ok
}

func yooKassaWebhookSecret() string {
	if v := strings.TrimSpace(os.Getenv("YOOKASSA_WEBHOOK_SECRET")); v != "" {
		return v
	}
	// Prefer dedicated webhook gate; fall back to API secret if set.
	if v := strings.TrimSpace(os.Getenv("YOOKASSA_SECRET")); v != "" {
		return v
	}
	_, secret, _ := yooKassaCreds()
	return secret
}

func mockPayAllowed(dev bool) bool {
	if os.Getenv("ALLOW_MOCK_PAY") == "1" {
		return true
	}
	return dev
}

type yooAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type yooCreateReq struct {
	Amount       yooAmount         `json:"amount"`
	Capture      bool              `json:"capture"`
	Confirmation map[string]string `json:"confirmation"`
	Description  string            `json:"description"`
	Metadata     map[string]string `json:"metadata"`
}

type yooPayment struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Paid         bool   `json:"paid"`
	Amount       yooAmount `json:"amount"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
	Metadata map[string]string `json:"metadata"`
}

type yooAPI struct {
	base   string
	shopID string
	secret string
	http   *http.Client
}

func newYooAPI() *yooAPI {
	shop, secret, _ := yooKassaCreds()
	base := strings.TrimSpace(os.Getenv("YOOKASSA_API_BASE"))
	if base == "" {
		base = "https://api.yookassa.ru/v3"
	}
	return &yooAPI{
		base: strings.TrimRight(base, "/"), shopID: shop, secret: secret,
		http: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *yooAPI) authHeader() string {
	tok := base64.StdEncoding.EncodeToString([]byte(c.shopID + ":" + c.secret))
	return "Basic " + tok
}

func (c *yooAPI) createPayment(idempotenceKey string, amount float64, currency, returnURL, description string, meta map[string]string) (*yooPayment, error) {
	body := yooCreateReq{
		Amount:       yooAmount{Value: fmt.Sprintf("%.2f", amount), Currency: currency},
		Capture:      true,
		Confirmation: map[string]string{"type": "redirect", "return_url": returnURL},
		Description:  description,
		Metadata:     meta,
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.base+"/payments", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Idempotence-Key", idempotenceKey)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("yookassa create %d: %s", res.StatusCode, truncate(string(b), 300))
	}
	var out yooPayment
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out.ID == "" {
		return nil, fmt.Errorf("yookassa: empty payment id")
	}
	return &out, nil
}

func (c *yooAPI) getPayment(providerID string) (*yooPayment, error) {
	req, err := http.NewRequest(http.MethodGet, c.base+"/payments/"+providerID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("yookassa get %d: %s", res.StatusCode, truncate(string(b), 300))
	}
	var out yooPayment
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func webhookAuthorized(r *http.Request, expected string) bool {
	if expected == "" {
		return false
	}
	if r.Header.Get("X-Yookassa-Signature") == expected {
		return true
	}
	if r.Header.Get("Authorization") == "Bearer "+expected {
		return true
	}
	// Optional Basic shop:secret style
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Basic ") {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
		if err == nil && string(raw) == expected {
			return true
		}
		shop, secret, ok := yooKassaCreds()
		if ok && string(raw) == shop+":"+secret {
			return true
		}
	}
	return false
}
