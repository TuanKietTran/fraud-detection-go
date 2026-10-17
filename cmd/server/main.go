package main

import (
	"log"
	"net/http"

	"github.com/example/fraud-detection/internal/api"
	"github.com/example/fraud-detection/internal/query"
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

	log.Println("Fraud Detection API listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
