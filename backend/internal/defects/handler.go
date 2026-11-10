package defects

import (
	"encoding/json"
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

func (h *Handler) Register(writer http.ResponseWriter, request *http.Request) {
	projectID, _ := strconv.ParseInt(chi.URLParam(request, "project_id"), 10, 64)
	storyID, _ := strconv.ParseInt(chi.URLParam(request, "story_id"), 10, 64)
	var body struct {
		Description       string `json:"description"`
		Severity          string `json:"severity"`
		Status            string `json:"status"`
		DetectionSprintID int64  `json:"detection_sprint_id"`
	}
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	defect, err := h.service.Register(request.Context(), RegisterInput{
		ProjectID: projectID, StoryID: storyID,
		Description: body.Description, Severity: body.Severity,
		Status: body.Status, DetectionSprintID: body.DetectionSprintID,
	})
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(writer, http.StatusCreated, defect)
}

func (h *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	projectID, _ := strconv.ParseInt(chi.URLParam(request, "project_id"), 10, 64)
	id, _ := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
	defect, err := h.service.Get(request.Context(), projectID, id)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(writer, http.StatusOK, defect)
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
