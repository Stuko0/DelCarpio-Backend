package handlers

import (
	"io"
	"net/http"

	"delcarpio/backend/internal/auth"
	"delcarpio/backend/internal/postgrest"
	stripeclient "delcarpio/backend/internal/stripe"
)

type PaymentHandler struct {
	pg    *postgrest.Client
	st    *stripeclient.Client
	cfg   PaymentConfig
}

type PaymentConfig struct {
	PublicKey string
	BaseURL   string
}

func NewPaymentHandler(pg *postgrest.Client, st *stripeclient.Client, cfg PaymentConfig) *PaymentHandler {
	return &PaymentHandler{pg: pg, st: st, cfg: cfg}
}

func (h *PaymentHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		jsonError(w, "unauthorized", 401)
		return
	}

	var req struct {
		OrderID string                   `json:"order_id"`
		Items   []stripeclient.LineItem `json:"items"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}
	if req.OrderID == "" || len(req.Items) == 0 {
		jsonError(w, "order_id and items required", 400)
		return
	}

	// Verify order exists and belongs to user
	filters := postgrest.EqFilter("id", req.OrderID)
	var order map[string]interface{}
	if err := h.pg.GetOne(r.Context(), "orders", filters, &order); err != nil {
		jsonError(w, "order not found", 404)
		return
	}
	if owner, _ := order["user_id"].(string); owner != userID {
		jsonError(w, "forbidden", 403)
		return
	}

	sessionReq := &stripeclient.SessionRequest{
		Mode:       "payment",
		SuccessURL: h.cfg.BaseURL + "/pedido/confirmado?id=" + req.OrderID,
		CancelURL:  h.cfg.BaseURL + "/checkout?status=cancelled",
		LineItems:  req.Items,
		Metadata: map[string]string{
			"order_id": req.OrderID,
			"user_id":  userID,
		},
	}

	session, err := h.st.CreateCheckoutSession(sessionReq)
	if err != nil {
		jsonError(w, "stripe error: "+err.Error(), 502)
		return
	}

	jsonOK(w, map[string]interface{}{
		"session_id": session.ID,
		"url":        session.URL,
	})
}

func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", 400)
		return
	}

	event, err := h.st.ConstructEvent(body, r.Header.Get("Stripe-Signature"))
	if err != nil {
		http.Error(w, "invalid signature", 400)
		return
	}

	if event.Type != "checkout.session.completed" {
		jsonOK(w, map[string]string{"status": "ignored"})
		return
	}

	orderID := event.Data.Object.Metadata["order_id"]
	if orderID == "" {
		jsonOK(w, map[string]string{"status": "no_order_id"})
		return
	}

	// Map Stripe payment status to order status
	var orderStatus string
	switch event.Data.Object.PaymentStatus {
	case "paid":
		orderStatus = "paid"
	case "no_payment_required":
		orderStatus = "processing"
	case "unpaid":
		orderStatus = "pending"
	default:
		orderStatus = "paid"
	}

	orderFilters := postgrest.EqFilter("id", orderID)
	h.pg.Update(r.Context(), "orders", orderFilters, map[string]interface{}{
		"status":  orderStatus,
		"updated": "now()",
	}, nil)

	jsonOK(w, map[string]string{"status": "ok"})
}
