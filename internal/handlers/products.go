package handlers

import (
	"net/http"

	"delcarpio/backend/internal/postgrest"
)

type ProductHandler struct {
	pg *postgrest.Client
}

func NewProductHandler(pg *postgrest.Client) *ProductHandler {
	return &ProductHandler{pg: pg}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := postgrest.ListFilters("*", "visible", "true", "created", "desc", 50)

	var products []map[string]interface{}
	if err := h.pg.List(r.Context(), "products", filters, &products); err != nil {
		jsonError(w, "query failed", 500)
		return
	}

	if products == nil {
		products = []map[string]interface{}{}
	}
	jsonOK(w, products)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	filters := postgrest.ListFilters("*", "slug", slug, "created", "desc", 1)

	var product map[string]interface{}
	if err := h.pg.GetOne(r.Context(), "products", filters, &product); err != nil {
		if err == postgrest.ErrNoRows {
			jsonError(w, "product not found", 404)
			return
		}
		jsonError(w, "query failed", 500)
		return
	}
	jsonOK(w, product)
}
