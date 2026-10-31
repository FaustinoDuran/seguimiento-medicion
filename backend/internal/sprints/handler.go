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

type assignRequest struct {
	StoryID int64 `json:"story_id"`
}

type assignedStoryResponse struct {
	SprintID   int64  `json:"sprint_id"`
	StoryID    int64  `json:"story_id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Status     string `json:"status"`
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

func (h *Handler) Assign(writer http.ResponseWriter, request *http.Request) {
	sprintID, ok := parseID(writer, request, "invalid sprint id")
	if !ok {
		return
	}
	var body assignRequest
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	if body.StoryID < 1 {
		writeError(writer, http.StatusBadRequest, "story id is required")
		return
	}
	assigned, err := h.service.Assign(request.Context(), sprintID, body.StoryID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, toAssignedStory(assigned))
}

func (h *Handler) Close(writer http.ResponseWriter, request *http.Request) {
	id, ok := parseID(writer, request, "invalid sprint id")
	if !ok {
		return
	}

	sprint, err := h.service.Close(request.Context(), id)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, toResponse(sprint))
}

func (h *Handler) ListStories(writer http.ResponseWriter, request *http.Request) {
	sprintID, ok := parseID(writer, request, "invalid sprint id")
	if !ok {
		return
	}
	stories, err := h.service.ListStories(request.Context(), sprintID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	response := make([]assignedStoryResponse, 0, len(stories))
	for _, story := range stories {
		response = append(response, toAssignedStory(story))
	}
	writeJSON(writer, http.StatusOK, response)
}

func (h *Handler) Complete(writer http.ResponseWriter, request *http.Request) {
	sprintID, ok := parseID(writer, request, "invalid sprint id")
	if !ok {
		return
	}
	storyID, err := strconv.ParseInt(chi.URLParam(request, "story_id"), 10, 64)
	if err != nil || storyID < 1 {
		writeError(writer, http.StatusBadRequest, "invalid story id")
		return
	}
	story, err := h.service.Complete(request.Context(), sprintID, storyID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, toAssignedStory(story))
}

func toAssignedStory(story AssignedStory) assignedStoryResponse {
	return assignedStoryResponse{
		SprintID:   story.SprintID,
		StoryID:    story.StoryID,
		Identifier: story.Identifier,
		Title:      story.Title,
		Status:     story.Status,
	}
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
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrProjectNotFound), errors.Is(err, ErrStoryNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrGoalRequired):
		writeError(writer, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrSprintClosed), errors.Is(err, ErrStoryAlreadyAssigned), errors.Is(err, ErrStoryNotAssigned):
		writeError(writer, http.StatusConflict, err.Error())
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
