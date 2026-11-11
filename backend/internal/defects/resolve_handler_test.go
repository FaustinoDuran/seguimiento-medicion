package defects

import (
	"encoding/json"
	"errors"
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

func TestHTTPResolutionErrorsLeaveDefectUnchanged(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{name: "missing sprint", body: `{}`, status: 400},
		{name: "null sprint", body: `{"resolution_sprint_id":null}`, status: 400},
		{name: "zero sprint", body: `{"resolution_sprint_id":0}`, status: 400},
		{name: "negative sprint", body: `{"resolution_sprint_id":-1}`, status: 400},
		{name: "fractional sprint", body: `{"resolution_sprint_id":3.1}`, status: 400},
		{name: "string sprint", body: `{"resolution_sprint_id":"3"}`, status: 400},
		{name: "overflow sprint", body: `{"resolution_sprint_id":9223372036854775808}`, status: 400},
		{name: "empty body", status: 400},
		{name: "null body", body: `null`, status: 400},
		{name: "array", body: `[]`, status: 400},
		{name: "invalid JSON", body: `{`, status: 400},
		{name: "extra JSON", body: `{"resolution_sprint_id":3}{}`, status: 400},
		{name: "extra junk", body: `{"resolution_sprint_id":3}x`, status: 400},
		{name: "invalid project", path: "/projects/x/defects/1/resolve", body: `{"resolution_sprint_id":3}`, status: 400},
		{name: "negative defect", path: "/projects/1/defects/-1/resolve", body: `{"resolution_sprint_id":3}`, status: 400},
		{
			name: "overflow defect", path: "/projects/1/defects/9223372036854775808/resolve",
			body: `{"resolution_sprint_id":3}`, status: 400,
		},
		{name: "unknown defect", path: "/projects/1/defects/99/resolve", body: `{"resolution_sprint_id":3}`, status: 404},
		{name: "foreign defect", path: "/projects/2/defects/1/resolve", body: `{"resolution_sprint_id":3}`, status: 404},
		{name: "unknown sprint", body: `{"resolution_sprint_id":99}`, status: 404},
		{name: "foreign sprint", body: `{"resolution_sprint_id":8}`, status: 404},
		{name: "earlier sprint", body: `{"resolution_sprint_id":9}`, status: 409},
		{name: "earlier tie", body: `{"resolution_sprint_id":2}`, status: 409},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, reader, _, original := resolutionFixture(t)
			router := chi.NewRouter()
			Mount(router, reader)
			MountResolution(router, service)
			path := tc.path
			if path == "" {
				path = "/projects/1/defects/1/resolve"
			}
			response := requestJSON(router, "POST", path, tc.body)
			assertHTTPError(t, response, tc.status)
			get := requestJSON(router, "GET", "/projects/1/defects/1", "")
			if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), original) {
				t.Fatalf("rejected request changed defect: %d %s", get.Code, get.Body.String())
			}
		})
	}
}

func TestHTTPResolutionConflictAndWriteFailure(t *testing.T) {
	service, reader, storage, _ := resolutionFixture(t)
	router := chi.NewRouter()
	Mount(router, reader)
	MountResolution(router, service)
	path := "/projects/1/defects/1/resolve"
	storage.resolveErr = errors.New("password=private host=internal")
	failed := requestJSON(router, "POST", path, `{"resolution_sprint_id":3}`)
	assertHTTPError(t, failed, 500)
	if failed.Body.String() != "{\"error\":\"internal error\"}\n" {
		t.Fatalf("leaked failure: %s", failed.Body.String())
	}
	storage.resolveErr = nil
	first := requestJSON(router, "POST", path, `{"resolution_sprint_id":3}`)
	if first.Code != 200 {
		t.Fatalf("resolve: %d %s", first.Code, first.Body.String())
	}
	conflict := requestJSON(router, "POST", path, `{"resolution_sprint_id":1}`)
	assertHTTPError(t, conflict, 409)
	get := requestJSON(router, "GET", "/projects/1/defects/1", "")
	if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), decodeDefect(t, first)) {
		t.Fatalf("conflict changed resolution: %s", get.Body.String())
	}
}
