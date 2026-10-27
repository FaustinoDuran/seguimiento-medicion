package sprints

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type sprintResponse struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Goal      string    `json:"goal"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type createRequest struct {
	Goal string `json:"goal"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "invalid project id")
	if !ok {
		return
	}

	var body createRequest
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}

	sprint, err := h.service.Create(request.Context(), CreateSprint{
		ProjectID: projectID,
		Goal:      body.Goal,
	})
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, toResponse(sprint))
}

func (h *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	id, ok := parseID(writer, request, "invalid sprint id")
	if !ok {
		return
	}

	sprint, err := h.service.Get(request.Context(), id)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, toResponse(sprint))
}

func toResponse(sprint Sprint) sprintResponse {
	return sprintResponse{
		ID:        sprint.ID,
		ProjectID: sprint.ProjectID,
		Goal:      sprint.Goal,
		Status:    sprint.Status,
		CreatedAt: sprint.CreatedAt,
	}
}

func parseID(writer http.ResponseWriter, request *http.Request, message string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
	if err != nil || id < 1 {
		writeError(writer, http.StatusBadRequest, message)
		return 0, false
	}
	return id, true
}

func writeServiceError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrProjectNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrGoalRequired):
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
	writeJSON(writer, status, errorResponse{Error: message})
}
