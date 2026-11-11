package defects

import (
	"encoding/json"
	"io"
	"net/http"
)

type ResolutionHandler struct {
	service *ResolutionService
}

func NewResolutionHandler(service *ResolutionService) *ResolutionHandler {
	return &ResolutionHandler{service: service}
}

func (h *ResolutionHandler) Resolve(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := parseID(writer, request, "project_id")
	if !ok {
		return
	}
	id, ok := parseID(writer, request, "id")
	if !ok {
		return
	}
	var body struct {
		ResolutionSprintID int64 `json:"resolution_sprint_id"`
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
	defect, err := h.service.Resolve(request.Context(), ResolveInput{
		ProjectID: projectID, DefectID: id, ResolutionSprintID: body.ResolutionSprintID,
	})
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, defect)
}
