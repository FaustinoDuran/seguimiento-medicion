package planningpoker

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type estimateHandler struct{ service *EstimationService }

func MountEstimates(router chi.Router, service *EstimationService) {
	handler := &estimateHandler{service: service}
	router.Post("/projects/{project_id}/rounds/{round_id}/estimate", handler.Record)
}

func (h *estimateHandler) Record(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	roundID, ok := parseID(writer, request, "round_id")
	if !ok {
		return
	}
	var body struct {
		MemberID    int64 `json:"member_id"`
		StoryPoints int64 `json:"story_points"`
	}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(writer, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.Record(request.Context(), RecordEstimate{
		ProjectID: projectID, RoundID: roundID, MemberID: body.MemberID, StoryPoints: body.StoryPoints,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEstimateID):
			writeError(writer, http.StatusBadRequest, err.Error())
		case estimateConflict(err):
			writeError(writer, http.StatusConflict, err.Error())
		default:
			writeServiceError(writer, err)
		}
		return
	}
	writeJSON(writer, http.StatusOK, result)
}

func estimateConflict(err error) bool {
	conflicts := []error{ErrRoundNotRevealed, ErrNoVotes, ErrOutdatedRound, ErrStoryCompleted, ErrStorySprintClosed}
	for _, conflict := range conflicts {
		if errors.Is(err, conflict) {
			return true
		}
	}
	return false
}
