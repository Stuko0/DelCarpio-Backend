package handlers

import (
	"context"
	"net/http"

	"delcarpio/backend/internal/postgrest"
)

type AdminStatsHandler struct {
	pg *postgrest.Client
}

func NewAdminStatsHandler(pg *postgrest.Client) *AdminStatsHandler {
	return &AdminStatsHandler{pg: pg}
}

func (h *AdminStatsHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"products": h.count(r.Context(), "products"),
		"orders":   h.count(r.Context(), "orders"),
		"profiles": h.count(r.Context(), "profiles"),
		"recipes":  h.count(r.Context(), "recipes"),
	}
	jsonOK(w, stats)
}

func (h *AdminStatsHandler) count(ctx context.Context, table string) int {
	var result []map[string]interface{}
	f := postgrest.ListFilters("count", "id", "*", "id", "asc", 1)
	if err := h.pg.List(ctx, table, f, &result); err != nil {
		return 0
	}
	return len(result)
}
