package sprints

import "github.com/go-chi/chi/v5"

func Mount(router chi.Router, service *Service) {
	handler := NewHandler(service)
	router.Post("/projects/{id}/sprints", handler.Create)
	router.Get("/sprints/{id}", handler.Get)
	router.Post("/sprints/{id}/stories", handler.Assign)
	router.Get("/sprints/{id}/stories", handler.ListStories)
}
