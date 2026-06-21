package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	directURL := os.Getenv("DATABASE_DIRECT_URL")
	if directURL == "" {
		log.Fatal("DATABASE_DIRECT_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, directURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	migrationsDir := "supabase/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("read migrations dir: %v", err)
	}

	var names []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".sql" {
			names = append(names, f.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		sql, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			log.Fatalf("read %s: %v", name, err)
		}

		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			log.Fatalf("apply %s: %v", name, err)
		}
		fmt.Printf("Applied: %s\n", name)
	}
}
