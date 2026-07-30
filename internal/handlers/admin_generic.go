package handlers

import (
	"net/http"

	"delcarpio/backend/internal/postgrest"
)

type AdminGenericHandler struct {
	pg *postgrest.Client
}

func NewAdminGenericHandler(pg *postgrest.Client) *AdminGenericHandler {
	return &AdminGenericHandler{pg: pg}
}

// ListAll returns all rows from a table (admin)
func (h *AdminGenericHandler) ListAll(table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rows []map[string]interface{}
		if err := h.pg.List(r.Context(), table, nil, &rows); err != nil {
			rows = []map[string]interface{}{}
		}
		jsonOK(w, rows)
	}
}

// GetOne returns a single row by id
func (h *AdminGenericHandler) GetOne(table, idField string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			jsonError(w, "id required", 400)
			return
		}
		filters := postgrest.EqFilter(idField, id)
		var row map[string]interface{}
		if err := h.pg.GetOne(r.Context(), table, filters, &row); err != nil {
			jsonError(w, "not found", 404)
			return
		}
		jsonOK(w, row)
	}
}

// Create inserts a new row
func (h *AdminGenericHandler) Create(table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		if err := decodeJSON(r, &input); err != nil {
			jsonError(w, "invalid body", 400)
			return
		}
		var created map[string]interface{}
		if err := h.pg.Create(r.Context(), table, input, &created); err != nil {
			jsonError(w, "create failed: "+err.Error(), 500)
			return
		}
		jsonOK(w, created, 201)
	}
}

// Update modifies an existing row by id
func (h *AdminGenericHandler) Update(table, idField string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			jsonError(w, "id required", 400)
			return
		}
		var input map[string]interface{}
		if err := decodeJSON(r, &input); err != nil {
			jsonError(w, "invalid body", 400)
			return
		}
		input["updated"] = "now()"
		filters := postgrest.EqFilter(idField, id)
		var result []map[string]interface{}
		if err := h.pg.Update(r.Context(), table, filters, input, &result); err != nil {
			jsonError(w, "update failed: "+err.Error(), 500)
			return
		}
		if len(result) == 0 {
			jsonError(w, "not found", 404)
			return
		}
		jsonOK(w, result[0])
	}
}

// Delete soft-deletes by setting visible=false
func (h *AdminGenericHandler) Delete(table, idField string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			jsonError(w, "id required", 400)
			return
		}
		filters := postgrest.EqFilter(idField, id)
		var result []map[string]interface{}
		if err := h.pg.Update(r.Context(), table, filters, map[string]interface{}{
			"visible": false,
			"updated": "now()",
		}, &result); err != nil {
			jsonError(w, "delete failed: "+err.Error(), 500)
			return
		}
		if len(result) == 0 {
			jsonError(w, "not found", 404)
			return
		}
		jsonOK(w, map[string]interface{}{"deleted": true, "id": id})
	}
}
