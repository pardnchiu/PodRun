package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/pardnchiu/PodRun/internal/database"
	"github.com/pardnchiu/PodRun/internal/handler"
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("failed to load .env",
			slog.String("error", err.Error()))
	}
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		if _, err := os.Stat("/.dockerenv"); err == nil {
			dbPath = "/data/database.db"
		} else {
			home, _ := os.UserHomeDir()
			dbPath = filepath.Join(home, ".podrun", "database.db")
			// 確保目錄存在
			dir := filepath.Dir(dbPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}
	}
	db, err := database.NewSQLite(dbPath)
	if err != nil {
		log.Fatalf("[x] failed to create database: %v", err)
	}

	// # NOT THIS PROJECT POINT, REMOVE IT FOR NOW
	// if emails := os.Getenv("ALLOW_EMAILS"); emails != "" {
	// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()

	// 	if err := db.ResetUsers(ctx); err != nil {
	// 		log.Fatalf("[x] failed to reset users: %v", err)
	// 	}

	// 	for e := range strings.SplitSeq(emails, ",") {
	// 		e = strings.TrimSpace(e)
	// 		if e == "" {
	// 			continue
	// 		}
	// 		if err := db.UpsertUser(ctx, &model.User{Email: e}); err != nil {
	// 			log.Fatalf("[x] failed to insert email %s: %v\n", e, err)
	// 		}
	// 	}
	// }

	if err := handler.NewRoutes(db); err != nil {
		log.Fatalf("[x] failed to initialize http: %v", err)
	}
}
