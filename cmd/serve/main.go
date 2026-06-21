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
	"delcarpio/backend/internal/db"
	"delcarpio/backend/internal/handlers"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(corsMiddleware)

	productHandler := handlers.NewProductHandler(pool)
	recipeHandler := handlers.NewRecipeHandler(pool)
	orderHandler := handlers.NewOrderHandler(pool)
	authHandler := handlers.NewAuthHandler(cfg.SupabaseURL, cfg.SupabaseAnonKey)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/api/products", productHandler.List)
	r.Get("/api/products/{slug}", productHandler.Get)
	r.Get("/api/recipes", recipeHandler.List)
	r.Get("/api/recipes/{slug}", recipeHandler.Get)

	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.SupabaseJWTSecret))
		r.Post("/api/orders", orderHandler.Create)
		r.Get("/api/orders", orderHandler.List)
	})

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := []string{
			"http://localhost:4321",
			"http://localhost:5173",
			"https://delcarpio.vercel.app",
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
