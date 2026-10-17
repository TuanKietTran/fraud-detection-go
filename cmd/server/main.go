package main

import (
	"log"
	"net/http"
	"os"

	"github.com/TuanKietTran/fraud-detection-go/internal/api"
	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	scoreH := query.NewScoreEventHandler()
	handlers := api.NewHandlers(scoreH)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Mount("/v1", handlers.Routes())

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8103"
	}
	log.Println("Fraud Detection API listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
