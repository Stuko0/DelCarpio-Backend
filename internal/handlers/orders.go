package handlers

import (
	"net/http"

	"delcarpio/backend/internal/auth"
	"delcarpio/backend/internal/postgrest"
)

type OrderHandler struct {
	pg *postgrest.Client
}

func NewOrderHandler(pg *postgrest.Client) *OrderHandler {
	return &OrderHandler{pg: pg}
}

type createOrderRequest struct {
	Items []map[string]interface{} `json:"items"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		jsonError(w, "unauthorized", 401)
		return
	}

	var req createOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	total := calculateTotal(req.Items)

	payload := map[string]interface{}{
		"user_id": userID,
		"items":   req.Items,
		"status":  "pending",
		"total":   total,
	}

	var created map[string]interface{}
	if err := h.pg.Create(r.Context(), "orders", payload, &created); err != nil {
		jsonError(w, "create failed", 500)
		return
	}

	jsonOK(w, created, 201)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		jsonError(w, "unauthorized", 401)
		return
	}

	filters := postgrest.ListFilters("*", "user_id", userID, "created", "desc", 50)

	var orders []map[string]interface{}
	if err := h.pg.List(r.Context(), "orders", filters, &orders); err != nil {
		jsonError(w, "query failed", 500)
		return
	}

	if orders == nil {
		orders = []map[string]interface{}{}
	}
	jsonOK(w, orders)
}

func calculateTotal(items []map[string]interface{}) float64 {
	var total float64
	for _, item := range items {
		price, _ := item["price"].(float64)
		qty, _ := item["quantity"].(float64)
		if qty == 0 {
			qty = 1
		}
		total += price * qty
	}
	return total
}
