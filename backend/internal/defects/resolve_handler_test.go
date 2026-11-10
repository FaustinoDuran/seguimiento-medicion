package defects

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHTTPResolveAndConsultDefect(t *testing.T) {
	service, reader, _, _ := resolutionFixture(t)
	router := chi.NewRouter()
	Mount(router, reader)
	MountResolution(router, service)
	open := requestJSON(router, "GET", "/projects/1/defects/1", "")
	var openBody map[string]json.RawMessage
	if err := json.Unmarshal(open.Body.Bytes(), &openBody); err != nil {
		t.Fatal(err)
	}
	if string(openBody["resolution_sprint_id"]) != "null" {
		t.Fatalf("open resolution: %s", open.Body.String())
	}
	resolved := requestJSON(router, "POST", "/projects/1/defects/1/resolve", `{"resolution_sprint_id":3}`)
	if resolved.Code != 200 {
		t.Fatalf("resolve: %d %s", resolved.Code, resolved.Body.String())
	}
	defect := decodeDefect(t, resolved)
	if defect.Status != "resolved" || defect.ResolutionSprintID == nil || *defect.ResolutionSprintID != 3 {
		t.Fatalf("resolution: %+v", defect)
	}
	for range 2 {
		get := requestJSON(router, "GET", "/projects/1/defects/1", "")
		if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), defect) {
			t.Fatalf("get: %d %s", get.Code, get.Body.String())
		}
		retry := requestJSON(router, "POST", "/projects/1/defects/1/resolve", `{"resolution_sprint_id":3}`)
		if retry.Code != 200 || !reflect.DeepEqual(decodeDefect(t, retry), defect) {
			t.Fatalf("retry: %d %s", retry.Code, retry.Body.String())
		}
	}
}
