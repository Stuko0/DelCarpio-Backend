package handlers

import (
	"net/http"

	"delcarpio/backend/internal/postgrest"
)

type AdminProductHandler struct {
	pg *postgrest.Client
}

func NewAdminProductHandler(pg *postgrest.Client) *AdminProductHandler {
	return &AdminProductHandler{pg: pg}
}

func (h *AdminProductHandler) List(w http.ResponseWriter, r *http.Request) {
	var products []map[string]interface{}
	if err := h.pg.List(r.Context(), "products", nil, &products); err != nil {
		jsonError(w, "query failed", 500)
		return
	}
	if products == nil {
		products = []map[string]interface{}{}
	}
	jsonOK(w, products)
}

func (h *AdminProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := decodeJSON(r, &input); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	// Required fields
	if _, ok := input["name"]; !ok {
		jsonError(w, "name is required", 400)
		return
	}
	if _, ok := input["slug"]; !ok {
		jsonError(w, "slug is required", 400)
		return
	}

	// Defaults
	if _, ok := input["visible"]; !ok {
		input["visible"] = true
	}

	var created map[string]interface{}
	if err := h.pg.Create(r.Context(), "products", input, &created); err != nil {
		jsonError(w, "create failed: "+err.Error(), 500)
		return
	}
	jsonOK(w, created, 201)
}

func (h *AdminProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		jsonError(w, "slug required", 400)
		return
	}

	var input map[string]interface{}
	if err := decodeJSON(r, &input); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	// Remove slug from updates (PK)
	delete(input, "slug")
	input["updated"] = "now()"

	filters := postgrest.EqFilter("slug", slug)
	var result []map[string]interface{}
	if err := h.pg.Update(r.Context(), "products", filters, input, &result); err != nil {
		jsonError(w, "update failed: "+err.Error(), 500)
		return
	}

	if len(result) == 0 {
		jsonError(w, "product not found", 404)
		return
	}
	jsonOK(w, result[0])
}

func (h *AdminProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		jsonError(w, "slug required", 400)
		return
	}

	// Soft delete: set visible=false
	filters := postgrest.EqFilter("slug", slug)
	var result []map[string]interface{}
	if err := h.pg.Update(r.Context(), "products", filters, map[string]interface{}{
		"visible": false,
		"updated": "now()",
	}, &result); err != nil {
		jsonError(w, "delete failed: "+err.Error(), 500)
		return
	}

	if len(result) == 0 {
		jsonError(w, "product not found", 404)
		return
	}
	jsonOK(w, map[string]interface{}{"deleted": true, "slug": slug})
}
