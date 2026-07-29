package handlers

import (
	"net/http"

	"delcarpio/backend/internal/auth"
	"delcarpio/backend/internal/postgrest"
)

type AddressHandler struct {
	pg *postgrest.Client
}

func NewAddressHandler(pg *postgrest.Client) *AddressHandler {
	return &AddressHandler{pg: pg}
}

func (h *AddressHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	filters := postgrest.EqFilter("user_id", userID)
	var addresses []map[string]interface{}
	if err := h.pg.List(r.Context(), "addresses", filters, &addresses); err != nil {
		// Table might not exist yet — return empty list instead of 500
		addresses = []map[string]interface{}{}
	}
	if addresses == nil {
		addresses = []map[string]interface{}{}
	}
	jsonOK(w, addresses)
}

func (h *AddressHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	var input map[string]interface{}
	if err := decodeJSON(r, &input); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	required := []string{"name", "address", "city"}
	for _, field := range required {
		if _, ok := input[field]; !ok {
			jsonError(w, field+" is required", 400)
			return
		}
	}

	input["user_id"] = userID

	var created map[string]interface{}
	if err := h.pg.Create(r.Context(), "addresses", input, &created); err != nil {
		jsonError(w, "La funcionalidad de direcciones no está disponible. Ejecuta la migración SQL en Supabase.", 400)
		return
	}
	jsonOK(w, created, 201)
}

func (h *AddressHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", 400)
		return
	}

	// Verify ownership
	ownerFilter := postgrest.EqFilter("id", id)
	var existing map[string]interface{}
	if err := h.pg.GetOne(r.Context(), "addresses", ownerFilter, &existing); err != nil {
		if err == postgrest.ErrNoRows {
			jsonError(w, "address not found", 404)
			return
		}
		jsonError(w, "query failed", 500)
		return
	}
	if owner, _ := existing["user_id"].(string); owner != userID {
		jsonError(w, "forbidden", 403)
		return
	}

	var input map[string]interface{}
	if err := decodeJSON(r, &input); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}
	input["updated"] = "now()"

	var result []map[string]interface{}
	if err := h.pg.Update(r.Context(), "addresses", ownerFilter, input, &result); err != nil {
		jsonError(w, "update failed: "+err.Error(), 500)
		return
	}
	if len(result) == 0 {
		jsonError(w, "address not found", 404)
		return
	}
	jsonOK(w, result[0])
}

func (h *AddressHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", 400)
		return
	}

	ownerFilter := postgrest.EqFilter("id", id)
	var existing map[string]interface{}
	if err := h.pg.GetOne(r.Context(), "addresses", ownerFilter, &existing); err != nil {
		if err == postgrest.ErrNoRows {
			jsonError(w, "address not found", 404)
			return
		}
		jsonError(w, "query failed", 500)
		return
	}
	if owner, _ := existing["user_id"].(string); owner != userID {
		jsonError(w, "forbidden", 403)
		return
	}

	var result []map[string]interface{}
	if err := h.pg.Update(r.Context(), "addresses", ownerFilter, map[string]interface{}{
		"deleted": true,
		"updated": "now()",
	}, &result); err != nil {
		jsonError(w, "delete failed: "+err.Error(), 500)
		return
	}
	jsonOK(w, map[string]interface{}{"deleted": true, "id": id})
}
