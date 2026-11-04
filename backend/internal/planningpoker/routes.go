package planningpoker

import "github.com/go-chi/chi/v5"

func Mount(router chi.Router, service *Service) {
	handler := NewHandler(service)
	router.Post("/projects/{project_id}/stories/{story_id}/rounds", handler.OpenRound)
	router.Post("/projects/{project_id}/rounds/{round_id}/votes", handler.CastVote)
	router.Post("/projects/{project_id}/rounds/{round_id}/reveal", handler.Reveal)
	router.Get("/projects/{project_id}/rounds/{round_id}", handler.View)
}
