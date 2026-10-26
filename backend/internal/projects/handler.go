package projects

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

type projectResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type createRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    Status `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	var body createRequest
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}

	start, err := parseDate(body.StartDate)
	if err != nil {
		writeError(writer, http.StatusBadRequest, ErrInvalidDates.Error())
		return
	}
	end, err := parseDate(body.EndDate)
	if err != nil {
		writeError(writer, http.StatusBadRequest, ErrInvalidDates.Error())
		return
	}

	project, err := h.service.Create(request.Context(), CreateProject{
		Name:      body.Name,
		StartDate: start,
		EndDate:   end,
		Status:    body.Status,
	})
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, toResponse(project))
}

func (h *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	id, ok := parseProjectID(writer, request)
	if !ok {
		return
	}

	project, err := h.service.Get(request.Context(), id)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, toResponse(project))
}

func toResponse(project Project) projectResponse {
	return projectResponse{
		ID:        project.ID,
		Name:      project.Name,
		StartDate: project.StartDate.UTC().Format("2006-01-02"),
		EndDate:   project.EndDate.UTC().Format("2006-01-02"),
		Status:    project.Status,
		CreatedAt: project.CreatedAt,
	}
}

func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", value)
}

type memberResponse struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type addMemberRequest struct {
	Name string `json:"name"`
}

type membersResponse struct {
	Members []memberResponse `json:"members"`
}

func (h *Handler) AddMember(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseProjectID(writer, request)
	if !ok {
		return
	}

	var body addMemberRequest
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}

	member, err := h.service.AddMember(request.Context(), AddMember{
		ProjectID: projectID,
		Name:      body.Name,
	})
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, toMemberResponse(member))
}

func (h *Handler) ListMembers(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseProjectID(writer, request)
	if !ok {
		return
	}

	members, err := h.service.ListMembers(request.Context(), projectID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}

	body := membersResponse{Members: make([]memberResponse, 0, len(members))}
	for _, member := range members {
		body.Members = append(body.Members, toMemberResponse(member))
	}
	writeJSON(writer, http.StatusOK, body)
}

func parseProjectID(writer http.ResponseWriter, request *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
	if err != nil || id < 1 {
		writeError(writer, http.StatusBadRequest, "invalid project id")
		return 0, false
	}
	return id, true
}

func toMemberResponse(member Member) memberResponse {
	return memberResponse{
		ID:        member.ID,
		ProjectID: member.ProjectID,
		Name:      member.Name,
		CreatedAt: member.CreatedAt,
	}
}

func writeServiceError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateMember):
		writeError(writer, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrMemberNameRequired), errors.Is(err, ErrInvalidDates), errors.Is(err, ErrInvalidStatus):
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
