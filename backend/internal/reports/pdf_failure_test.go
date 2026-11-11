package reports

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestPDFResponseDoesNotPublishAFailedRender(t *testing.T) {
	previous := renderReportPDF
	renderReportPDF = func(Report) ([]byte, error) {
		return []byte("%PDF-partial"), errors.New("render failed")
	}
	t.Cleanup(func() { renderReportPDF = previous })

	router := chi.NewRouter()
	Mount(router, NewService(pdfSourceRepository{source: Source{
		Project: Project{ID: 1, Name: "Seguimiento", Status: "active"},
	}}))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/projects/1/report.pdf", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	if response.Body.String() != `{"error":"internal error"}` {
		t.Fatal(response.Body.String())
	}
}

type pdfSourceRepository struct{ source Source }

func (r pdfSourceRepository) LoadProject(context.Context, int64) (Source, error) {
	return r.source, nil
}
