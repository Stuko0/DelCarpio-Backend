package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"delcarpio/backend/internal/auth"
	"delcarpio/backend/internal/config"
	"delcarpio/backend/internal/handlers"
	"delcarpio/backend/internal/postgrest"
	stripeclient "delcarpio/backend/internal/stripe"
)

func main() {
	cfg := config.Load()

	if cfg.SupabaseURL == "" || cfg.SupabaseServiceRole == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY are required")
	}

	pg := postgrest.New(cfg.SupabaseURL, cfg.SupabaseServiceRole)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(corsMiddleware)

	productHandler := handlers.NewProductHandler(pg)
	recipeHandler := handlers.NewRecipeHandler(pg)
	orderHandler := handlers.NewOrderHandler(pg)
	authHandler := handlers.NewAuthHandler(cfg.SupabaseURL, cfg.SupabaseAnonKey)
	profileHandler := handlers.NewProfileHandler(pg)
	addressHandler := handlers.NewAddressHandler(pg)

	var stripeClient *stripeclient.Client
	if cfg.StripeSecretKey != "" {
		stripeClient = stripeclient.New(cfg.StripeSecretKey)
	}
	paymentHandler := handlers.NewPaymentHandler(pg, stripeClient, handlers.PaymentConfig{
		BaseURL: getBaseURL(),
	})
	adminProductHandler := handlers.NewAdminProductHandler(pg)
	adminOrderHandler := handlers.NewAdminOrderHandler(pg)
	adminRecipeHandler := handlers.NewAdminRecipeHandler(pg)
	adminStatsHandler := handlers.NewAdminStatsHandler(pg)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/api/products", productHandler.List)
	r.Get("/api/products/{slug}", productHandler.Get)
	r.Get("/api/recipes", recipeHandler.List)
	r.Get("/api/recipes/{slug}", recipeHandler.Get)

	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/payments/webhook", paymentHandler.Webhook)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.SupabaseJWTSecret))
		r.Post("/api/orders", orderHandler.Create)
		r.Get("/api/orders", orderHandler.List)
		r.Get("/api/orders/{id}", orderHandler.Get)
		r.Post("/api/orders/{id}/cancel", orderHandler.Cancel)
		r.Get("/api/profile", profileHandler.Get)
		r.Put("/api/profile", profileHandler.Update)
		r.Get("/api/profile/addresses", addressHandler.List)
		r.Post("/api/profile/addresses", addressHandler.Create)
		r.Put("/api/profile/addresses/{id}", addressHandler.Update)
		r.Delete("/api/profile/addresses/{id}", addressHandler.Delete)
		r.Post("/api/payments/create-checkout", paymentHandler.CreateCheckout)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.SupabaseJWTSecret))
		r.Use(auth.AdminMiddleware)
		r.Get("/api/admin/products", adminProductHandler.List)
		r.Post("/api/admin/products", adminProductHandler.Create)
		r.Put("/api/admin/products/{slug}", adminProductHandler.Update)
		r.Delete("/api/admin/products/{slug}", adminProductHandler.Delete)
		r.Get("/api/admin/orders", adminOrderHandler.List)
		r.Get("/api/admin/orders/{id}", adminOrderHandler.Get)
		r.Put("/api/admin/orders/{id}/status", adminOrderHandler.UpdateStatus)
		r.Get("/api/admin/recipes", adminRecipeHandler.List)
		r.Post("/api/admin/recipes", adminRecipeHandler.Create)
		r.Put("/api/admin/recipes/{slug}", adminRecipeHandler.Update)
		r.Delete("/api/admin/recipes/{slug}", adminRecipeHandler.Delete)
		r.Get("/api/admin/stats", adminStatsHandler.Stats)
	})

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Del Carpio backend starting on :%s …", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := []string{
			"http://localhost:4321",
			"http://localhost:5173",
			"https://delcarpio.stuko.dev",
		}
		allow := false
		for _, a := range allowed {
			if origin == a {
				allow = true
				break
			}
		}
		if allow {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getBaseURL() string {
	if url := os.Getenv("BASE_URL"); url != "" {
		return url
	}
	return "https://delcarpio-backend.onrender.com"
}
