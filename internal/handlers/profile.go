package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"delcarpio/backend/internal/auth"
	"delcarpio/backend/internal/postgrest"
)

type ProfileHandler struct {
	pg *postgrest.Client
}

func NewProfileHandler(pg *postgrest.Client) *ProfileHandler {
	return &ProfileHandler{pg: pg}
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		jsonError(w, "unauthorized", 401)
		return
	}

	filters := postgrest.EqFilter("user_id", userID)
	var user map[string]interface{}
	err := h.pg.GetOne(r.Context(), "users", filters, &user)
	if err == postgrest.ErrNoRows {
		jsonOK(w, map[string]interface{}{
			"user_id": userID,
			"name":    "",
			"email":   "",
			"phone":   "",
			"role":    "customer",
		})
		return
	}
	if err != nil {
		jsonError(w, "query failed", 500)
		return
	}
	jsonOK(w, user)
}

func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		jsonError(w, "unauthorized", 401)
		return
	}

	var updates map[string]interface{}
	if err := decodeJSON(r, &updates); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	// Extract caller role from JWT
	callerRole := getRoleFromJWT(r)
	allowed := map[string]bool{"name": true, "email": true, "phone": true, "avatar_url": true}
	if callerRole == "admin" || callerRole == "owner" {
		allowed["role"] = true
	}

	clean := make(map[string]interface{})
	for k, v := range updates {
		if allowed[k] {
			clean[k] = v
		}
	}
	if len(clean) == 0 {
		jsonError(w, "no valid fields to update", 400)
		return
	}
	clean["updated"] = time.Now().UTC().Format(time.RFC3339)

	filters := postgrest.EqFilter("user_id", userID)
	var existing map[string]interface{}
	err := h.pg.GetOne(r.Context(), "users", filters, &existing)

	if err == nil {
		var result []map[string]interface{}
		if err := h.pg.Update(r.Context(), "users", filters, clean, &result); err != nil {
			jsonError(w, "update failed: "+err.Error(), 500)
			return
		}
		if len(result) > 0 {
			jsonOK(w, result[0])
		} else {
			jsonOK(w, clean)
		}
	} else {
		clean["user_id"] = userID
		var created map[string]interface{}
		if err := h.pg.Create(r.Context(), "users", clean, &created); err != nil {
			jsonError(w, "create failed: "+err.Error(), 500)
			return
		}
		jsonOK(w, created, 201)
	}
}

func getRoleFromJWT(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse without verification just to read claims
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return ""
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	meta, _ := claims["user_metadata"].(map[string]interface{})
	role, _ := meta["role"].(string)
	return role
}
