package effort

import "github.com/go-chi/chi/v5"

func Mount(router chi.Router, service *Service) {
	handler := NewHandler(service)
	router.Post("/projects/{project_id}/stories/{story_id}/effort", handler.Register)
}
