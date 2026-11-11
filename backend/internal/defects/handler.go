package defects

import (
	"encoding/json"
	"errors"
	"io"
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
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	storyID, ok := parseID(writer, request, "story_id")
	if !ok {
		return
	}
	var body struct {
		Description       string          `json:"description"`
		Severity          string          `json:"severity"`
		Status            json.RawMessage `json:"status"`
		DetectionSprintID int64           `json:"detection_sprint_id"`
	}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(writer, http.StatusBadRequest, "expected a single json object")
		return
	}
	var status string
	if len(body.Status) > 0 {
		if err := json.Unmarshal(body.Status, &status); err != nil {
			writeError(writer, http.StatusBadRequest, "invalid json")
			return
		}
	}
	defect, err := h.service.Register(request.Context(), RegisterInput{
		ProjectID: projectID, StoryID: storyID,
		Description: body.Description, Severity: body.Severity,
		Status: status, StatusProvided: len(body.Status) > 0,
		DetectionSprintID: body.DetectionSprintID,
	})
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, defect)
}

func (h *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	id, ok := parseID(writer, request, "id")
	if !ok {
		return
	}
	defect, err := h.service.Get(request.Context(), projectID, id)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, defect)
}

func (h *Handler) List(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	defects, err := h.service.List(request.Context(), projectID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, struct {
		Defects []Defect `json:"defects"`
	}{Defects: defects})
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
	case errors.Is(err, ErrResolutionBeforeDetection), errors.Is(err, ErrResolutionConflict):
		writeError(writer, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrProjectNotFound),
		errors.Is(err, ErrStoryNotFound), errors.Is(err, ErrSprintNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidID), errors.Is(err, ErrDescriptionRequired),
		errors.Is(err, ErrDescriptionTooLong), errors.Is(err, ErrInvalidSeverity), errors.Is(err, ErrInvalidStatus):
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
