package handlers

import (
	"net/http"

	"delcarpio/backend/internal/postgrest"
)

type AdminRecipeHandler struct {
	pg *postgrest.Client
}

func NewAdminRecipeHandler(pg *postgrest.Client) *AdminRecipeHandler {
	return &AdminRecipeHandler{pg: pg}
}

func (h *AdminRecipeHandler) List(w http.ResponseWriter, r *http.Request) {
	var recipes []map[string]interface{}
	if err := h.pg.List(r.Context(), "recipes", nil, &recipes); err != nil {
		jsonError(w, "query failed", 500)
		return
	}
	if recipes == nil {
		recipes = []map[string]interface{}{}
	}
	jsonOK(w, recipes)
}

func (h *AdminRecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := decodeJSON(r, &input); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	if _, ok := input["title"]; !ok {
		jsonError(w, "title is required", 400)
		return
	}
	if _, ok := input["slug"]; !ok {
		jsonError(w, "slug is required", 400)
		return
	}

	if _, ok := input["published"]; !ok {
		input["published"] = false
	}

	var created map[string]interface{}
	if err := h.pg.Create(r.Context(), "recipes", input, &created); err != nil {
		jsonError(w, "create failed: "+err.Error(), 500)
		return
	}
	jsonOK(w, created, 201)
}

func (h *AdminRecipeHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	delete(input, "slug")
	input["updated"] = "now()"

	filters := postgrest.EqFilter("slug", slug)
	var result []map[string]interface{}
	if err := h.pg.Update(r.Context(), "recipes", filters, input, &result); err != nil {
		jsonError(w, "update failed: "+err.Error(), 500)
		return
	}

	if len(result) == 0 {
		jsonError(w, "recipe not found", 404)
		return
	}
	jsonOK(w, result[0])
}

func (h *AdminRecipeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		jsonError(w, "slug required", 400)
		return
	}

	// Soft delete: set published=false
	filters := postgrest.EqFilter("slug", slug)
	var result []map[string]interface{}
	if err := h.pg.Update(r.Context(), "recipes", filters, map[string]interface{}{
		"published": false,
		"updated":   "now()",
	}, &result); err != nil {
		jsonError(w, "delete failed: "+err.Error(), 500)
		return
	}

	if len(result) == 0 {
		jsonError(w, "recipe not found", 404)
		return
	}
	jsonOK(w, map[string]interface{}{"deleted": true, "slug": slug})
}
