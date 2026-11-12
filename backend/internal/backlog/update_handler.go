package backlog

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MountUpdates(router chi.Router, service *UpdateService) {
	router.Patch("/projects/{project_id}/stories/{id}", func(writer http.ResponseWriter, request *http.Request) {
		projectID, ok := parseID(writer, request, "project_id")
		if !ok {
			return
		}
		id, ok := parseID(writer, request, "id")
		if !ok {
			return
		}
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		var input *UpdateStory
		if err := decoder.Decode(&input); err != nil || input == nil {
			writeError(writer, http.StatusBadRequest, "invalid json")
			return
		}
		var extra json.RawMessage
		if err := decoder.Decode(&extra); err != io.EOF {
			writeError(writer, http.StatusBadRequest, "expected a single json object")
			return
		}
		input.ProjectID, input.ID = projectID, id
		story, err := service.Update(request.Context(), *input)
		switch {
		case errors.Is(err, ErrInvalidUpdateID), errors.Is(err, ErrNoStoryChanges), errors.Is(err, ErrInvalidUpdateStatus):
			writeError(writer, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrUpdateSprintClosed), errors.Is(err, ErrUpdateNeedsOpenSprint),
			errors.Is(err, ErrStoryUnestimated), errors.Is(err, ErrCompletedStoryEstimate):
			writeError(writer, http.StatusConflict, err.Error())
		case err != nil:
			writeServiceError(writer, err)
		default:
			writeJSON(writer, http.StatusOK, story)
		}
	})
}
