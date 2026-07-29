package handlers

import (
	"net/http"
	"strconv"

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

	// Apply query params
	q := r.URL.Query()

	// Category filter
	if cat := q.Get("category"); cat != "" {
		filters.Set("category", "eq."+cat)
	}

	// Search by name (PostgREST ILIKE)
	if search := q.Get("q"); search != "" {
		filters.Set("name", "ilike.*"+search+"*")
	}

	// Price range
	if minP := q.Get("min_price"); minP != "" {
		filters.Set("price", "gte."+minP)
	}
	if maxP := q.Get("max_price"); maxP != "" {
		filters.Set("price", "lte."+maxP)
	}

	// Sort
	if sort := q.Get("sort"); sort != "" {
		switch sort {
		case "price_asc":
			filters.Set("order", "price.asc")
		case "price_desc":
			filters.Set("order", "price.desc")
		case "name_asc":
			filters.Set("order", "name.asc")
		case "name_desc":
			filters.Set("order", "name.desc")
		case "newest":
			filters.Set("order", "created.desc")
		}
	}

	// Pagination
	if page := q.Get("page"); page != "" {
		if limit := q.Get("limit"); limit != "" {
			p, _ := strconv.Atoi(page)
			l, _ := strconv.Atoi(limit)
			if p < 1 {
				p = 1
			}
			if l < 1 || l > 100 {
				l = 50
			}
			filters.Set("offset", strconv.Itoa((p-1)*l))
			filters.Set("limit", strconv.Itoa(l))
		}
	}

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
