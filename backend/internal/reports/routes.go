package reports

import "github.com/go-chi/chi/v5"

func Mount(router chi.Router, service *Service) {
	handler := NewHandler(service)
	router.Get("/projects/{project_id}/report", handler.Project)
	router.Get("/projects/{project_id}/report.pdf", handler.ProjectPDF)
	router.Get("/projects/{project_id}/sprints/{sprint_id}/report", handler.Sprint)
	router.Get("/projects/{project_id}/sprints/{sprint_id}/report.pdf", handler.SprintPDF)
}
