package effort

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

func (h *Handler) Register(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	storyID, ok := parseID(writer, request, "story_id")
	if !ok {
		return
	}

	decoder := json.NewDecoder(request.Body)
	var body struct {
		MemberID int64           `json:"member_id"`
		Date     string          `json:"date"`
		Activity string          `json:"activity"`
		Hours    json.RawMessage `json:"hours"`
	}
	if err := decoder.Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}

	date, ok := parseDate(writer, body.Date)
	if !ok {
		return
	}
	saved, err := h.service.Register(request.Context(), RegisterInput{
		ProjectID: projectID,
		StoryID:   storyID,
		MemberID:  body.MemberID,
		Date:      date,
		Activity:  body.Activity,
		Hours:     hoursText(body.Hours),
	})
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, entryResponse{
		ID:        saved.ID,
		ProjectID: saved.ProjectID,
		StoryID:   saved.StoryID,
		MemberID:  saved.MemberID,
		Date:      saved.Date.Format("2006-01-02"),
		Activity:  saved.Activity,
		Hours:     saved.Hours,
	})
}

type entryResponse struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"project_id"`
	StoryID   int64  `json:"story_id"`
	MemberID  int64  `json:"member_id"`
	Date      string `json:"date"`
	Activity  string `json:"activity"`
	Hours     Hours  `json:"hours"`
}

func parseID(writer http.ResponseWriter, request *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(request, name), 10, 64)
	if err != nil || id < 1 {
		writeError(writer, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return id, true
}

func parseDate(writer http.ResponseWriter, value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, true
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		writeError(writer, http.StatusBadRequest, "effort date is invalid")
		return time.Time{}, false
	}
	return parsed, true
}

func hoursText(raw json.RawMessage) string {
	text := string(bytes.TrimSpace(raw))
	if text == "" || text == "null" {
		return ""
	}
	return text
}

func writeServiceError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrStoryNotFound), errors.Is(err, ErrMemberNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidID),
		errors.Is(err, ErrDateRequired),
		errors.Is(err, ErrActivityRequired),
		errors.Is(err, ErrActivityTooLong),
		errors.Is(err, ErrHoursRequired),
		errors.Is(err, ErrHoursNotPositive),
		errors.Is(err, ErrInvalidHours):
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
