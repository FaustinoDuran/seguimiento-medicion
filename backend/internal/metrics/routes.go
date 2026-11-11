package metrics

import "github.com/go-chi/chi/v5"

func Mount(router chi.Router, service *Service) {
	handler := NewHandler(service)
	router.Get("/projects/{project_id}/metrics", handler.Project)
	router.Get("/sprints/{sprint_id}/metrics", handler.Sprint)
}
