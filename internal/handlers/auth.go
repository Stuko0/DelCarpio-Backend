package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

type AuthHandler struct {
	supabaseURL string
	anonKey     string
	client      *http.Client
}

func NewAuthHandler(supabaseURL, anonKey string) *AuthHandler {
	return &AuthHandler{
		supabaseURL: supabaseURL,
		anonKey:     anonKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "cannot read body", 400)
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err == nil {
		if opts, ok := data["options"].(map[string]interface{}); ok {
			if d, ok := opts["data"].(map[string]interface{}); ok {
				d["role"] = "customer"
			}
		}
		modified, err := json.Marshal(data)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(modified))
		}
	}

	h.proxyToSupabase(w, r, h.supabaseURL+"/auth/v1/signup")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.proxyToSupabase(w, r, h.supabaseURL+"/auth/v1/token?grant_type=password")
}

func (h *AuthHandler) proxyToSupabase(w http.ResponseWriter, r *http.Request, target string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "cannot read body", 400)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, bytes.NewReader(body))
	if err != nil {
		jsonError(w, "internal error", 500)
		return
	}
	req.Header.Set("apikey", h.anonKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		jsonError(w, "upstream error", 502)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		jsonError(w, "cannot read response", 502)
		return
	}

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Del("Content-Length")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

type jsonErr struct {
	Error string `json:"error"`
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(jsonErr{Error: msg})
}

func jsonOK(w http.ResponseWriter, data interface{}, status ...int) {
	w.Header().Set("Content-Type", "application/json")
	code := 200
	if len(status) > 0 {
		code = status[0]
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
