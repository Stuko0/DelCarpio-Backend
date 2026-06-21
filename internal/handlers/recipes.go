package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeHandler struct {
	db *pgxpool.Pool
}

func NewRecipeHandler(db *pgxpool.Pool) *RecipeHandler {
	return &RecipeHandler{db: db}
}

func (h *RecipeHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		`SELECT id, title, slug, published, content_markdown, created, updated
		 FROM recipes
		 WHERE published = true
		 ORDER BY created DESC
		 LIMIT 50`)
	if err != nil {
		jsonError(w, "query failed", 500)
		return
	}
	defer rows.Close()

	recipes, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		jsonError(w, "collect failed", 500)
		return
	}

	if recipes == nil {
		recipes = []map[string]interface{}{}
	}

	jsonOK(w, recipes)
}

func (h *RecipeHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, title, slug, published, content_markdown, created, updated
		 FROM recipes
		 WHERE slug = $1
		 LIMIT 1`, slug)
	if err != nil {
		jsonError(w, "query failed", 500)
		return
	}
	defer rows.Close()

	recipe, err := pgx.CollectOneRow(rows, pgx.RowToMap)
	if err != nil {
		if err == pgx.ErrNoRows {
			jsonError(w, "recipe not found", 404)
			return
		}
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, recipe)
}
