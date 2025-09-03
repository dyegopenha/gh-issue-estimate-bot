package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dyegopenha/gh-issue-estimate-bot/internal/webhook"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("WEBHOOK_SECRET is required")
	}

	logger := log.New(os.Stdout, "[issue-estimate-bot] ", log.LstdFlags|log.Lshortfile)

	mux := http.NewServeMux()
	h := webhook.NewHandler(logger, secret)
	h.Register(mux)

	addr := ":" + port
	logger.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatalf("server error: %v", err)
	}
}
