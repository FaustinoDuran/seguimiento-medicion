package bdd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cucumber/godog"
)

type world struct {
	router    http.Handler
	projectID int64
	status    int
	body      []byte
}

type metricsReport struct {
	ProjectID                int64       `json:"project_id"`
	SprintID                 *int64      `json:"sprint_id"`
	PlannedStoryPoints       int64       `json:"planned_story_points"`
	CompletedStoryPoints     int64       `json:"completed_story_points"`
	Velocity                 json.Number `json:"velocity"`
	EstimatedHours           json.Number `json:"estimated_hours"`
	ActualHours              json.Number `json:"actual_hours"`
	Deviation                int64       `json:"deviation"`
	CompletedStoryPercentage json.Number `json:"completed_story_percentage"`
	DefectsDetected          int64       `json:"defects_detected"`
	DefectsResolved          int64       `json:"defects_resolved"`
}

func (w *world) reset() {
	w.projectID = 0
	w.status = 0
	w.body = nil
}

func (w *world) register(ctx *godog.ScenarioContext) {
	ctx.Step(`^un proyecto con historias estimadas, un Sprint, horas cargadas y un defecto$`, w.projectWithMeasurementData)
	ctx.Step(`^un proyecto cuyo único Sprint sigue abierto$`, w.projectWithOnlyOpenSprint)
	ctx.Step(`^consulta las métricas del proyecto$`, w.queryProjectMetrics)
	ctx.Step(`^consulta métricas con un id de proyecto 0$`, w.queryProjectZero)
	ctx.Step(`^el backend devuelve story points, velocidad, horas, desviación, porcentaje y defectos$`, w.metricsMatchTheFixture)
	ctx.Step(`^la velocidad del proyecto es 0$`, w.velocityIsZero)
	ctx.Step(`^la respuesta es 400$`, w.responseIsBadRequest)
}

func (w *world) projectWithMeasurementData() error {
	projectID, err := w.create("/projects", `{
		"name":"Métricas BDD","start_date":"2026-10-26","end_date":"2026-11-23"
	}`)
	if err != nil {
		return err
	}
	w.projectID = projectID
	memberID, err := w.create(fmt.Sprintf("/projects/%d/members", projectID), `{"name":"Faustino"}`)
	if err != nil {
		return err
	}
	storyID, err := w.create(fmt.Sprintf("/projects/%d/stories", projectID), `{
		"identifier":"HU-BDD","title":"Medir el proyecto","story_points":8
	}`)
	if err != nil {
		return err
	}
	sprintID, err := w.create(fmt.Sprintf("/projects/%d/sprints", projectID), `{"goal":"Cerrar la medición"}`)
	if err != nil {
		return err
	}
	if err := w.postStatus(fmt.Sprintf("/sprints/%d/stories", sprintID), fmt.Sprintf(`{"story_id":%d}`, storyID), http.StatusCreated); err != nil {
		return err
	}
	if err := w.postStatus(fmt.Sprintf("/sprints/%d/stories/%d/complete", sprintID, storyID), "", http.StatusOK); err != nil {
		return err
	}
	if err := w.postStatus(fmt.Sprintf("/projects/%d/stories/%d/effort", projectID, storyID), fmt.Sprintf(`{
		"member_id":%d,"date":"2026-11-10","activity":"medir","hours":1.5
	}`, memberID), http.StatusCreated); err != nil {
		return err
	}
	return w.postStatus(fmt.Sprintf("/projects/%d/stories/%d/defects", projectID, storyID), fmt.Sprintf(`{
		"description":"el informe sale vacío","severity":"high","detection_sprint_id":%d
	}`, sprintID), http.StatusCreated)
}

func (w *world) projectWithOnlyOpenSprint() error {
	projectID, err := w.create("/projects", `{
		"name":"Sin cerrar","start_date":"2026-10-26","end_date":"2026-11-23"
	}`)
	if err != nil {
		return err
	}
	w.projectID = projectID
	_, err = w.create(fmt.Sprintf("/projects/%d/sprints", projectID), `{"goal":"Sprint todavía abierto"}`)
	return err
}

func (w *world) queryProjectMetrics() error {
	return w.get(fmt.Sprintf("/projects/%d/metrics", w.projectID))
}

func (w *world) queryProjectZero() error {
	return w.get("/projects/0/metrics")
}

func (w *world) metricsMatchTheFixture() error {
	report, err := w.report()
	if err != nil {
		return err
	}
	if w.status != http.StatusOK {
		return fmt.Errorf("estado %d: %s", w.status, w.body)
	}
	if report.SprintID != nil || report.ProjectID != w.projectID {
		return fmt.Errorf("alcance: %+v", report)
	}
	if report.PlannedStoryPoints != 8 || report.CompletedStoryPoints != 8 || report.Deviation != 0 {
		return fmt.Errorf("story points: %+v", report)
	}
	if report.Velocity.String() != "0" || report.EstimatedHours.String() != "0" || report.ActualHours.String() != "1.5" {
		return fmt.Errorf("velocidad y horas: %+v", report)
	}
	if report.CompletedStoryPercentage.String() != "100" || report.DefectsDetected != 1 || report.DefectsResolved != 0 {
		return fmt.Errorf("porcentaje y defectos: %+v", report)
	}
	return nil
}

func (w *world) velocityIsZero() error {
	report, err := w.report()
	if err != nil {
		return err
	}
	if w.status != http.StatusOK {
		return fmt.Errorf("estado %d: %s", w.status, w.body)
	}
	if report.Velocity.String() != "0" {
		return fmt.Errorf("velocidad %s", report.Velocity)
	}
	return nil
}

func (w *world) responseIsBadRequest() error {
	if w.status != http.StatusBadRequest {
		return fmt.Errorf("estado %d: %s", w.status, w.body)
	}
	return nil
}

func (w *world) report() (metricsReport, error) {
	var report metricsReport
	if err := json.Unmarshal(w.body, &report); err != nil {
		return report, fmt.Errorf("json: %w (%s)", err, w.body)
	}
	return report, nil
}

func (w *world) create(path, body string) (int64, error) {
	if err := w.postStatus(path, body, http.StatusCreated); err != nil {
		return 0, err
	}
	var payload struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(w.body, &payload); err != nil || payload.ID == 0 {
		return 0, fmt.Errorf("id de %s: %s", path, w.body)
	}
	return payload.ID, nil
}

func (w *world) postStatus(path, body string, want int) error {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	w.router.ServeHTTP(recorder, request)
	w.status = recorder.Code
	w.body = recorder.Body.Bytes()
	if recorder.Code != want {
		return fmt.Errorf("%s: %d %s", path, recorder.Code, recorder.Body.String())
	}
	return nil
}

func (w *world) get(path string) error {
	recorder := httptest.NewRecorder()
	w.router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	w.status = recorder.Code
	payload, err := io.ReadAll(recorder.Body)
	if err != nil {
		return err
	}
	w.body = payload
	return nil
}
