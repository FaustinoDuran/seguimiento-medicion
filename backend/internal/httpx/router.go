package httpx

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter() chi.Router {
	router := chi.NewRouter()
	router.Get("/health", healthHandler)
	return router
}
