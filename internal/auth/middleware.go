package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

type userResponse struct {
	ID string `json:"id"`
}

func Middleware(supabaseURL, anonKey string) func(http.Handler) http.Handler {
	client := &http.Client{Timeout: 10 * time.Second}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token against Supabase Auth API
			userID, err := validateWithSupabase(client, supabaseURL, anonKey, token)
			if err != nil || userID == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validateWithSupabase(client *http.Client, supabaseURL, anonKey, token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, supabaseURL+"/auth/v1/user", nil)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", anonKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase %s: %s", resp.Status, string(body))
	}

	var user userResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return user.ID, nil
}

func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(userIDKey).(string)
	return id
}

func AdminMiddleware(supabaseURL, anonKey string) func(http.Handler) http.Handler {
	client := &http.Client{Timeout: 10 * time.Second}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate and get full user metadata
			userID, role, err := validateWithRole(client, supabaseURL, anonKey, token)
			if err != nil || userID == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			if role != "admin" && role != "owner" {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type userResponseFull struct {
	ID           string                 `json:"id"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
}

func validateWithRole(client *http.Client, supabaseURL, anonKey, token string) (string, string, error) {
	// First try: validate via Supabase Auth API
	userID, role, err := validateWithSupabaseFull(client, supabaseURL, anonKey, token)
	if err == nil && userID != "" {
		return userID, role, nil
	}
	return "", "", fmt.Errorf("unauthorized")
}

func validateWithSupabaseFull(client *http.Client, supabaseURL, anonKey, token string) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, supabaseURL+"/auth/v1/user", nil)
	if err != nil {
		return "", "", fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", anonKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("supabase %s", resp.Status)
	}

	var user userResponseFull
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", "", fmt.Errorf("decode: %w", err)
	}

	role, _ := user.UserMetadata["role"].(string)
	return user.ID, role, nil
}
