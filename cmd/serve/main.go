package main

import (
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"delcarpio/backend/internal/handlers"
	"delcarpio/backend/internal/hooks"
)

func main() {
	app := pocketbase.New()

	// ── migrations ──
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// ── hooks on serve ──
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// CORS — allow Astro frontend
		origins := getEnv("PB_ALLOWED_ORIGINS", "http://localhost:4321,http://localhost:5173")
		se.Router.Bind(apis.CORS(apis.CORSConfig{
			AllowOrigins: strings.Split(origins, ","),
		}))

		// Register module handlers
		handlers.RegisterProducts(se, app)
		handlers.RegisterRecipes(se, app)
		handlers.RegisterOrders(se, app)

		return se.Next()
	})

	// ── lifecycle hooks ──
	hooks.RegisterProductHooks(app)
	hooks.RegisterOrderHooks(app)

	port := getEnv("PORT", "8090")
	os.Args = []string{os.Args[0], "serve", "--http=0.0.0.0:" + port}

	log.Printf("Del Carpio backend starting on :%s …", port)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
