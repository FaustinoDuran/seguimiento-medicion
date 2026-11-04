package planningpoker

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

func (h *Handler) OpenRound(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	storyID, ok := parseID(writer, request, "story_id")
	if !ok {
		return
	}
	round, err := h.service.OpenRound(request.Context(), projectID, storyID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, roundResponse{
		ID:        round.ID,
		ProjectID: round.ProjectID,
		StoryID:   round.StoryID,
		Status:    round.Status,
		Votes:     []voteResponse{},
	})
}

func (h *Handler) CastVote(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	roundID, ok := parseID(writer, request, "round_id")
	if !ok {
		return
	}
	var body struct {
		MemberID    int64  `json:"member_id"`
		StoryPoints *int64 `json:"story_points"`
	}
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	points := int64(0)
	if body.StoryPoints != nil {
		points = *body.StoryPoints
	}
	vote, err := h.service.CastVote(request.Context(), CastVote{
		ProjectID:   projectID,
		RoundID:     roundID,
		MemberID:    body.MemberID,
		StoryPoints: points,
	})
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, voteResponse{
		MemberID:    vote.MemberID,
		StoryPoints: &vote.StoryPoints,
	})
}

func (h *Handler) Reveal(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	roundID, ok := parseID(writer, request, "round_id")
	if !ok {
		return
	}
	var body struct {
		MemberID int64 `json:"member_id"`
	}
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	if body.MemberID < 1 {
		writeError(writer, http.StatusBadRequest, "member id is required")
		return
	}
	view, err := h.service.Reveal(request.Context(), projectID, roundID, body.MemberID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, roundResponse{
		ID:        view.ID,
		ProjectID: view.ProjectID,
		StoryID:   view.StoryID,
		Status:    view.Status,
		Votes:     voteResponses(view.Votes),
	})
}

func voteResponses(votes []VisibleVote) []voteResponse {
	body := make([]voteResponse, 0, len(votes))
	for _, vote := range votes {
		body = append(body, voteResponse{MemberID: vote.MemberID, StoryPoints: vote.StoryPoints})
	}
	return body
}

func (h *Handler) View(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	roundID, ok := parseID(writer, request, "round_id")
	if !ok {
		return
	}
	memberID, err := strconv.ParseInt(request.URL.Query().Get("member_id"), 10, 64)
	if err != nil || memberID < 1 {
		writeError(writer, http.StatusBadRequest, "member id is required")
		return
	}
	view, err := h.service.View(request.Context(), projectID, roundID, memberID)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	body := roundResponse{
		ID:        view.ID,
		ProjectID: view.ProjectID,
		StoryID:   view.StoryID,
		Status:    view.Status,
		Votes:     voteResponses(view.Votes),
	}
	writeJSON(writer, http.StatusOK, body)
}

type roundResponse struct {
	ID        int64          `json:"id"`
	ProjectID int64          `json:"project_id"`
	StoryID   int64          `json:"story_id"`
	Status    RoundStatus    `json:"status"`
	Votes     []voteResponse `json:"votes"`
}

type voteResponse struct {
	MemberID    int64  `json:"member_id"`
	StoryPoints *int64 `json:"story_points,omitempty"`
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
	case errors.Is(err, ErrStoryNotFound), errors.Is(err, ErrMemberNotFound), errors.Is(err, ErrRoundNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrOpenRoundExists), errors.Is(err, ErrRoundClosed):
		writeError(writer, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidStoryPoints):
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
