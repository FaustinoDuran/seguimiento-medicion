package defects

import "github.com/go-chi/chi/v5"

func Mount(router chi.Router, service *Service) {
	handler := NewHandler(service)
	router.Post("/projects/{project_id}/stories/{story_id}/defects", handler.Register)
	router.Get("/projects/{project_id}/defects/{id}", handler.Get)
}

func MountResolution(router chi.Router, service *ResolutionService) {
	handler := NewResolutionHandler(service)
	router.Post("/projects/{project_id}/defects/{id}/resolve", handler.Resolve)
}
