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

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", 400)
		return
	}

	filters := postgrest.EqFilter("id", id)
	var order map[string]interface{}
	if err := h.pg.GetOne(r.Context(), "orders", filters, &order); err != nil {
		if err == postgrest.ErrNoRows {
			jsonError(w, "order not found", 404)
			return
		}
		jsonError(w, "query failed", 500)
		return
	}

	// Verify ownership
	ownerID, _ := order["user_id"].(string)
	if ownerID != userID {
		jsonError(w, "forbidden", 403)
		return
	}

	jsonOK(w, order)
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", 400)
		return
	}

	filters := postgrest.EqFilter("id", id)
	var order map[string]interface{}
	if err := h.pg.GetOne(r.Context(), "orders", filters, &order); err != nil {
		if err == postgrest.ErrNoRows {
			jsonError(w, "order not found", 404)
			return
		}
		jsonError(w, "query failed", 500)
		return
	}

	ownerID, _ := order["user_id"].(string)
	if ownerID != userID {
		jsonError(w, "forbidden", 403)
		return
	}

	status, _ := order["status"].(string)
	if status != "pending" {
		jsonError(w, "only pending orders can be cancelled", 400)
		return
	}

	var result []map[string]interface{}
	if err := h.pg.Update(r.Context(), "orders", filters, map[string]interface{}{
		"status":  "cancelled",
		"updated": "now()",
	}, &result); err != nil {
		jsonError(w, "cancel failed: "+err.Error(), 500)
		return
	}

	if len(result) > 0 {
		jsonOK(w, result[0])
	} else {
		jsonOK(w, map[string]interface{}{"status": "cancelled"})
	}
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
