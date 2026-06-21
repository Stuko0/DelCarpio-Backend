package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductHandler struct {
	db *pgxpool.Pool
}

func NewProductHandler(db *pgxpool.Pool) *ProductHandler {
	return &ProductHandler{db: db}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		`SELECT id, name, slug, visible, created, updated
		 FROM products
		 WHERE visible = true
		 ORDER BY created DESC
		 LIMIT 50`)
	if err != nil {
		jsonError(w, "query failed", 500)
		return
	}
	defer rows.Close()

	products, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		jsonError(w, "collect failed", 500)
		return
	}

	if products == nil {
		products = []map[string]interface{}{}
	}

	jsonOK(w, products)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	row, err := h.db.Query(r.Context(),
		`SELECT id, name, slug, visible, created, updated
		 FROM products
		 WHERE slug = $1
		 LIMIT 1`, slug)
	if err != nil {
		jsonError(w, "query failed", 500)
		return
	}
	defer row.Close()

	product, err := pgx.CollectOneRow(row, pgx.RowToMap)
	if err != nil {
		if err == pgx.ErrNoRows {
			jsonError(w, "product not found", 404)
			return
		}
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, product)
}
