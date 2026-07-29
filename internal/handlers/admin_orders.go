package handlers

import (
	"net/http"

	"delcarpio/backend/internal/postgrest"
)

type AdminOrderHandler struct {
	pg *postgrest.Client
}

func NewAdminOrderHandler(pg *postgrest.Client) *AdminOrderHandler {
	return &AdminOrderHandler{pg: pg}
}

func (h *AdminOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := postgrest.ListFilters("*", "id", "*", "created", "desc", 100)
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

func (h *AdminOrderHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	jsonOK(w, order)
}

func (h *AdminOrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", 400)
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &input); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	validStatuses := map[string]bool{
		"pending": true, "paid": true, "processing": true,
		"shipped": true, "delivered": true, "cancelled": true,
	}
	if !validStatuses[input.Status] {
		jsonError(w, "invalid status", 400)
		return
	}

	filters := postgrest.EqFilter("id", id)
	var result []map[string]interface{}
	if err := h.pg.Update(r.Context(), "orders", filters, map[string]interface{}{
		"status":  input.Status,
		"updated": "now()",
	}, &result); err != nil {
		jsonError(w, "update failed: "+err.Error(), 500)
		return
	}

	if len(result) == 0 {
		jsonError(w, "order not found", 404)
		return
	}
	jsonOK(w, result[0])
}
