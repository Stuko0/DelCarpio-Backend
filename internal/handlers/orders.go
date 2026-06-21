package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"delcarpio/backend/internal/auth"
)

type OrderHandler struct {
	db *pgxpool.Pool
}

func NewOrderHandler(db *pgxpool.Pool) *OrderHandler {
	return &OrderHandler{db: db}
}

type createOrderRequest struct {
	Items []map[string]interface{} `json:"items"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		jsonError(w, "unauthorized", 401)
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	total := calculateTotal(req.Items)

	itemsJSON, err := json.Marshal(req.Items)
	if err != nil {
		jsonError(w, "invalid items", 400)
		return
	}

	rows, err := h.db.Query(r.Context(),
		`INSERT INTO orders (user_id, items, status, total)
		 VALUES ($1, $2, 'pending', $3)
		 RETURNING id, user_id, items, status, total, created`,
		userID, itemsJSON, total)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	created, err := pgx.CollectOneRow(rows, pgx.RowToMap)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, created, 201)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		jsonError(w, "unauthorized", 401)
		return
	}

	rows, err := h.db.Query(r.Context(),
		`SELECT id, user_id, items, status, total, created
		 FROM orders
		 WHERE user_id = $1
		 ORDER BY created DESC
		 LIMIT 50`, userID)
	if err != nil {
		jsonError(w, "query failed", 500)
		return
	}
	defer rows.Close()

	orders, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		jsonError(w, "collect failed", 500)
		return
	}

	if orders == nil {
		orders = []map[string]interface{}{}
	}

	jsonOK(w, orders)
}

func calculateTotal(items []map[string]interface{}) float64 {
	var total float64
	for _, item := range items {
		price, _ := item["price"].(float64)
		qty, _ := item["quantity"].(float64)
		if qty == 0 {
			qty = 1
		}
		total += price * qty
	}
	return total
}
