package reports

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Project(w http.ResponseWriter, r *http.Request) { h.serve(w, r, false) }
func (h *Handler) Sprint(w http.ResponseWriter, r *http.Request)  { h.serve(w, r, true) }

func (h *Handler) serve(w http.ResponseWriter, r *http.Request, bySprint bool) {
	projectID, err := parseID(r, "project_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidID.Error())
		return
	}
	var report Report
	if bySprint {
		sprintID, parseErr := parseID(r, "sprint_id")
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, ErrInvalidID.Error())
			return
		}
		report, err = h.service.ForSprint(r.Context(), projectID, sprintID)
	} else {
		report, err = h.service.ForProject(r.Context(), projectID)
	}
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound), errors.Is(err, ErrSprintNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrInvalidID):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func parseID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id < 1 {
		return 0, ErrInvalidID
	}
	return id, nil
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	// Marshal before writing headers so encoding failures cannot publish a partial report.
	body, err := json.Marshal(value)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
