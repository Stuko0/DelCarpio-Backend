package stripe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	secretKey string
	http      *http.Client
}

type SessionRequest struct {
	Mode                string          `json:"mode"`
	SuccessURL          string          `json:"success_url"`
	CancelURL           string          `json:"cancel_url"`
	LineItems           []LineItem      `json:"line_items"`
	Metadata            map[string]string `json:"metadata"`
	PaymentIntentData   *PaymentIntentData `json:"payment_intent_data,omitempty"`
}

type LineItem struct {
	PriceData PriceData `json:"price_data"`
	Quantity  int       `json:"quantity"`
}

type PriceData struct {
	Currency    string            `json:"currency"`
	ProductData ProductData      `json:"product_data"`
	UnitAmount  int64             `json:"unit_amount"` // in cents
}

type ProductData struct {
	Name string `json:"name"`
}

type PaymentIntentData struct {
	Description string `json:"description"`
}

type SessionResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Status   string `json:"status"`
}

type Event struct {
	Type string          `json:"type"`
	Data EventData       `json:"data"`
}

type EventData struct {
	Object EventObject   `json:"object"`
}

type EventObject struct {
	ID              string            `json:"id"`
	Status          string            `json:"status"`
	AmountTotal     int64             `json:"amount_total"`
	PaymentStatus   string            `json:"payment_status"`
	Metadata        map[string]string `json:"metadata"`
	ClientReference string            `json:"client_reference_id"`
}

func New(secretKey string) *Client {
	return &Client{
		secretKey: secretKey,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) CreateCheckoutSession(req *SessionRequest) (*SessionResponse, error) {
	data := url.Values{}
	data.Set("mode", req.Mode)
	data.Set("success_url", req.SuccessURL)
	data.Set("cancel_url", req.CancelURL)

	for k, v := range req.Metadata {
		data.Set("metadata["+k+"]", v)
	}

	for i, item := range req.LineItems {
		prefix := fmt.Sprintf("line_items[%d]", i)
		data.Set(prefix+"[quantity]", fmt.Sprintf("%d", item.Quantity))
		data.Set(prefix+"[price_data][currency]", item.PriceData.Currency)
		data.Set(prefix+"[price_data][product_data][name]", item.PriceData.ProductData.Name)
		data.Set(prefix+"[price_data][unit_amount]", fmt.Sprintf("%d", item.PriceData.UnitAmount))
	}

	r, err := http.NewRequest(http.MethodPost, "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	r.Header.Set("Authorization", "Bearer "+c.secretKey)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(r)
	if err != nil {
		return nil, fmt.Errorf("stripe: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("stripe %s: %s", resp.Status, string(respBody))
	}

	var result SessionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &result, nil
}

func (c *Client) GetSession(sessionID string) (*SessionResponse, error) {
	r, err := http.NewRequest(http.MethodGet, "https://api.stripe.com/v1/checkout/sessions/"+sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	r.Header.Set("Authorization", "Bearer "+c.secretKey)

	resp, err := c.http.Do(r)
	if err != nil {
		return nil, fmt.Errorf("stripe: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("stripe %s: %s", resp.Status, string(respBody))
	}

	var result SessionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &result, nil
}

func (c *Client) ConstructEvent(payload []byte, signature string) (*Event, error) {
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &event, nil
}
