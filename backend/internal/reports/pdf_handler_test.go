package reports_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/go-chi/chi/v5"
)

func TestProjectPDFDownloadLeavesJSONReportUntouched(t *testing.T) {
	source := reports.Source{
		Project: reports.Project{ID: 1, Name: "Seguimiento", Status: "active"},
		Stories: []reports.Story{{
			ID: 1, ProjectID: 1, Identifier: "HU-A", Title: "Exportar informe", Status: "todo",
		}},
	}
	router := mountedMemory(source, nil)

	jsonResponse := getRaw(router, "/projects/1/report")
	if jsonResponse.Code != http.StatusOK || !strings.Contains(jsonResponse.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("json report: %d %s", jsonResponse.Code, jsonResponse.Header().Get("Content-Type"))
	}
	if !strings.Contains(jsonResponse.Body.String(), `"name":"Seguimiento"`) {
		t.Fatalf("json body: %s", jsonResponse.Body.String())
	}

	response := getRaw(router, "/projects/1/report.pdf")
	if response.Code != http.StatusOK {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Header().Get("Content-Type"), "application/pdf") {
		t.Fatal(response.Header().Get("Content-Type"))
	}
	disposition := response.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "attachment") || !strings.Contains(disposition, "reporte-proyecto-1.pdf") {
		t.Fatal(disposition)
	}
	assertPDFText(t, response.Body.Bytes(), "Seguimiento", "HU-A", "Sin estimar")
}

func TestSprintPDFDownloadMatchesThatScope(t *testing.T) {
	source := reports.Source{
		Project: reports.Project{ID: 4, Name: "Seguimiento", Status: "active"},
		Sprints: []reports.Sprint{
			{ID: 2, ProjectID: 4, Goal: "Entregar el informe", Status: "closed"},
			{ID: 3, ProjectID: 4, Goal: "Selector interno", Status: "open"},
		},
		Stories: []reports.Story{
			{ID: 1, ProjectID: 4, Identifier: "HU-A", Title: "Exportar informe", Status: "done", SprintID: id(2), StoryPoints: id(5)},
			{ID: 2, ProjectID: 4, Identifier: "HU-B", Title: "Fuera del sprint", Status: "todo", SprintID: id(3)},
		},
	}
	response := getRaw(mountedMemory(source, nil), "/projects/4/sprints/2/report.pdf")
	if response.Code != http.StatusOK {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	disposition := response.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "reporte-proyecto-4-sprint-2.pdf") {
		t.Fatal(disposition)
	}
	text := assertPDFText(t, response.Body.Bytes(), "Entregar el informe", "Cerrado", "HU-A", "Completada")
	if strings.Contains(text, "HU-B") || strings.Contains(text, "Selector interno") || strings.Contains(text, "Todo el proyecto") {
		t.Fatalf("sprint pdf includes another scope\n%s", text)
	}
}

func TestPDFDownloadRejectsMissingAndInvalidScopes(t *testing.T) {
	source := reports.Source{Project: reports.Project{ID: 1, Name: "Seguimiento", Status: "active"}}
	cases := []struct {
		path string
		repo memoryRepository
		want int
	}{
		{path: "/projects/0/report.pdf", want: http.StatusBadRequest},
		{path: "/projects/abc/report.pdf", want: http.StatusBadRequest},
		{path: "/projects/1/sprints/9/report.pdf", repo: memoryRepository{source: source}, want: http.StatusNotFound},
		{path: "/projects/1/report.pdf", repo: memoryRepository{err: reports.ErrProjectNotFound}, want: http.StatusNotFound},
		{path: "/projects/1/report.pdf", repo: memoryRepository{err: errors.New("postgres password secret")}, want: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			repo := tc.repo
			if repo.source.Project.ID == 0 && repo.err == nil {
				repo.source = source
			}
			response := getRaw(mountedMemory(repo.source, repo.err), tc.path)
			if response.Code != tc.want {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Header().Get("Content-Type"), "application/pdf") || bytes.HasPrefix(response.Body.Bytes(), []byte("%PDF")) {
				t.Fatal("error response returned a pdf")
			}
			if !strings.Contains(response.Header().Get("Content-Type"), "application/json") || !strings.Contains(response.Body.String(), `"error"`) {
				t.Fatalf("json error: %s %s", response.Header().Get("Content-Type"), response.Body.String())
			}
			if strings.Contains(response.Body.String(), "password") {
				t.Fatal("internal details exposed")
			}
		})
	}
}

func mountedMemory(source reports.Source, err error) http.Handler {
	router := chi.NewRouter()
	reports.Mount(router, reports.NewService(memoryRepository{source: source, err: err}))
	return router
}

func getRaw(router http.Handler, path string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	return response
}

func assertPDFText(t *testing.T, file []byte, wants ...string) string {
	t.Helper()
	if !bytes.HasPrefix(file, []byte("%PDF")) {
		t.Fatalf("expected a PDF, got %q", file)
	}
	text := pdfPlainText(t, file)
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Errorf("PDF text missing %q\n%s", want, text)
		}
	}
	return text
}
