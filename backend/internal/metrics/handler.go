package metrics

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type reportResponse struct {
	ProjectID                int64  `json:"project_id"`
	SprintID                 *int64 `json:"sprint_id"`
	PlannedStoryPoints       int64  `json:"planned_story_points"`
	CompletedStoryPoints     int64  `json:"completed_story_points"`
	Velocity                 Amount `json:"velocity"`
	EstimatedHours           Amount `json:"estimated_hours"`
	ActualHours              Amount `json:"actual_hours"`
	Deviation                int64  `json:"deviation"`
	CompletedStoryPercentage Amount `json:"completed_story_percentage"`
	DefectsDetected          int64  `json:"defects_detected"`
	DefectsResolved          int64  `json:"defects_resolved"`
}

func (h *Handler) Project(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	report, err := h.service.ForProject(request.Context(), projectID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, toResponse(report))
}

func (h *Handler) Sprint(writer http.ResponseWriter, request *http.Request) {
	sprintID, ok := parseID(writer, request, "sprint_id")
	if !ok {
		return
	}
	report, err := h.service.ForSprint(request.Context(), sprintID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, toResponse(report))
}

func toResponse(report Report) reportResponse {
	return reportResponse{
		ProjectID:                report.ProjectID,
		SprintID:                 report.SprintID,
		PlannedStoryPoints:       report.PlannedStoryPoints,
		CompletedStoryPoints:     report.CompletedStoryPoints,
		Velocity:                 report.Velocity,
		EstimatedHours:           report.EstimatedHours,
		ActualHours:              report.ActualHours,
		Deviation:                report.Deviation,
		CompletedStoryPercentage: report.CompletedStoryPercentage,
		DefectsDetected:          report.DefectsDetected,
		DefectsResolved:          report.DefectsResolved,
	}
}

func parseID(writer http.ResponseWriter, request *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(request, name), 10, 64)
	if err != nil || id < 1 {
		writeError(writer, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return id, true
}

func writeServiceError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrProjectNotFound), errors.Is(err, ErrSprintNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidID):
		writeError(writer, http.StatusBadRequest, err.Error())
	default:
		writeError(writer, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(body)
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, struct {
		Error string `json:"error"`
	}{Error: message})
}
