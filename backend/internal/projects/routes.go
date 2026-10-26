package projects

import "github.com/go-chi/chi/v5"

func Mount(router chi.Router, service *Service) {
	handler := NewHandler(service)
	router.Post("/projects", handler.Create)
	router.Get("/projects/{id}", handler.Get)
	router.Post("/projects/{id}/members", handler.AddMember)
	router.Get("/projects/{id}/members", handler.ListMembers)
}
