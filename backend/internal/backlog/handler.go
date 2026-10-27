package backlog

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	decoder := json.NewDecoder(request.Body)
	var input CreateStory
	if err := decoder.Decode(&input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(writer, http.StatusBadRequest, "expected a single json object")
		return
	}
	input.ProjectID = projectID
	story, err := h.service.Create(request.Context(), input)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, story)
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
	story, err := h.service.Get(request.Context(), projectID, id)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, story)
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
	case errors.Is(err, ErrNotFound), errors.Is(err, projects.ErrNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateIdentifier), errors.Is(err, ErrProjectClosed):
		writeError(writer, http.StatusConflict, err.Error())
	case errors.Is(err, ErrIdentifierRequired), errors.Is(err, ErrTitleRequired),
		errors.Is(err, ErrInvalidPriority), errors.Is(err, ErrInvalidStatus),
		errors.Is(err, ErrInvalidStoryPoints), errors.Is(err, ErrInvalidCriterion):
		writeError(writer, http.StatusBadRequest, err.Error())
	default:
		writeError(writer, http.StatusInternalServerError, "internal error")
	}
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, struct {
		Error string `json:"error"`
	}{Error: message})
}

func writeJSON(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(body)
}
