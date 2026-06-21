package handlers

import (
	"net/http"

	"delcarpio/backend/internal/postgrest"
)

type RecipeHandler struct {
	pg *postgrest.Client
}

func NewRecipeHandler(pg *postgrest.Client) *RecipeHandler {
	return &RecipeHandler{pg: pg}
}

func (h *RecipeHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := postgrest.ListFilters("*", "published", "true", "created", "desc", 50)

	var recipes []map[string]interface{}
	if err := h.pg.List(r.Context(), "recipes", filters, &recipes); err != nil {
		jsonError(w, "query failed", 500)
		return
	}

	if recipes == nil {
		recipes = []map[string]interface{}{}
	}
	jsonOK(w, recipes)
}

func (h *RecipeHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	filters := postgrest.ListFilters("*", "slug", slug, "created", "desc", 1)

	var recipe map[string]interface{}
	if err := h.pg.GetOne(r.Context(), "recipes", filters, &recipe); err != nil {
		if err == postgrest.ErrNoRows {
			jsonError(w, "recipe not found", 404)
			return
		}
		jsonError(w, "query failed", 500)
		return
	}
	jsonOK(w, recipe)
}
