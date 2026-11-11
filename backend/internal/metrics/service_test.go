package metrics

import (
	"context"
	"errors"
	"testing"
)

func TestProjectMetricsFromCatalog(t *testing.T) {
	service := testService(sampleCatalog())

	report, err := service.ForProject(context.Background(), 1)
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertReport(t, report, Report{
		ProjectID:                1,
		PlannedStoryPoints:       13,
		CompletedStoryPoints:     8,
		Velocity:                 amount("8"),
		EstimatedHours:           amount("0"),
		ActualHours:              amount("4.5"),
		Deviation:                5,
		CompletedStoryPercentage: amount("50"),
		DefectsDetected:          1,
		DefectsResolved:          0,
	})
}

func TestSprintMetricsIgnoreTheOtherSprint(t *testing.T) {
	service := testService(sampleCatalog())

	report, err := service.ForSprint(context.Background(), 10)
	if err != nil {
		t.Fatalf("sprint: %v", err)
	}
	sprintID := int64(10)
	assertReport(t, report, Report{
		ProjectID:                1,
		SprintID:                 &sprintID,
		PlannedStoryPoints:       8,
		CompletedStoryPoints:     8,
		Velocity:                 amount("8"),
		EstimatedHours:           amount("0"),
		ActualHours:              amount("4.5"),
		Deviation:                0,
		CompletedStoryPercentage: amount("100"),
		DefectsDetected:          1,
		DefectsResolved:          0,
	})
}

func TestUnestimatedStoryDoesNotAddPlannedPoints(t *testing.T) {
	catalog := sampleCatalog()
	catalog.stories = append(catalog.stories, Story{
		ID: 3, ProjectID: 1, Status: "todo",
	})
	catalog.assignments = append(catalog.assignments, Assignment{SprintID: 10, StoryID: 3})
	service := testService(catalog)

	report, err := service.ForSprint(context.Background(), 10)
	if err != nil {
		t.Fatalf("sprint: %v", err)
	}
	if report.PlannedStoryPoints != 8 {
		t.Fatalf("planned: %d", report.PlannedStoryPoints)
	}
	if report.CompletedStoryPercentage.String() != "50" {
		t.Fatalf("percentage: %s", report.CompletedStoryPercentage)
	}
}

func TestProjectVelocityAveragesClosedSprints(t *testing.T) {
	eight, five := int64(8), int64(5)
	catalog := &memoryCatalog{
		projects: map[int64]struct{}{1: {}},
		sprints: []Sprint{
			{ID: 10, ProjectID: 1, Status: "closed"},
			{ID: 11, ProjectID: 1, Status: "closed"},
		},
		stories: []Story{
			{ID: 1, ProjectID: 1, Status: "done", StoryPoints: &eight},
			{ID: 2, ProjectID: 1, Status: "done", StoryPoints: &five},
		},
		assignments: []Assignment{
			{SprintID: 10, StoryID: 1},
			{SprintID: 11, StoryID: 2},
		},
	}
	service := testService(catalog)

	report, err := service.ForProject(context.Background(), 1)
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if report.Velocity.String() != "6.5" {
		t.Fatalf("velocity: %s", report.Velocity)
	}
}

func TestProjectVelocityWithoutClosedSprintsIsZero(t *testing.T) {
	catalog := sampleCatalog()
	catalog.sprints[0].Status = "open"
	catalog.sprints[1].Status = "open"
	service := testService(catalog)

	report, err := service.ForProject(context.Background(), 1)
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if report.Velocity.String() != "0" {
		t.Fatalf("velocity: %s", report.Velocity)
	}
}

func TestActualHoursSumEffortInScope(t *testing.T) {
	service := testService(sampleCatalog())

	report, err := service.ForSprint(context.Background(), 10)
	if err != nil {
		t.Fatalf("sprint: %v", err)
	}
	if report.EstimatedHours.String() != "0" || report.ActualHours.String() != "4.5" {
		t.Fatalf("hours: estimated %s actual %s", report.EstimatedHours, report.ActualHours)
	}
}

func TestDeviationAndCompletionPercentage(t *testing.T) {
	eight, five := int64(8), int64(5)
	catalog := sampleCatalog()
	catalog.stories = []Story{
		{ID: 1, ProjectID: 1, Status: "done", StoryPoints: &eight},
		{ID: 2, ProjectID: 1, Status: "todo", StoryPoints: &five},
	}
	catalog.assignments = []Assignment{
		{SprintID: 10, StoryID: 1},
		{SprintID: 10, StoryID: 2},
	}
	catalog.effort = []Effort{{ProjectID: 1, StoryID: 1, Hours: amount("1.5")}}
	service := testService(catalog)

	report, err := service.ForSprint(context.Background(), 10)
	if err != nil {
		t.Fatalf("sprint: %v", err)
	}
	if report.Deviation != 5 {
		t.Fatalf("deviation: %d", report.Deviation)
	}
	if report.CompletedStoryPercentage.String() != "50" {
		t.Fatalf("percentage: %s", report.CompletedStoryPercentage)
	}
}

func TestDefectCounts(t *testing.T) {
	resolvedSprint := int64(11)
	catalog := sampleCatalog()
	catalog.defects = []Defect{
		{ProjectID: 1, Status: "open", DetectionSprintID: 10},
		{ProjectID: 1, Status: "resolved", DetectionSprintID: 10, ResolutionSprintID: &resolvedSprint},
	}
	service := testService(catalog)

	sprint, err := service.ForSprint(context.Background(), 10)
	if err != nil {
		t.Fatalf("sprint: %v", err)
	}
	if sprint.DefectsDetected != 2 || sprint.DefectsResolved != 0 {
		t.Fatalf("sprint defects: %+v", sprint)
	}
	project, err := service.ForProject(context.Background(), 1)
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if project.DefectsDetected != 2 || project.DefectsResolved != 1 {
		t.Fatalf("project defects: %+v", project)
	}
}

func TestProjectMetricsIgnoreAnotherProject(t *testing.T) {
	catalog := sampleCatalog()
	five := int64(5)
	catalog.projects[2] = struct{}{}
	catalog.stories = append(catalog.stories, Story{ID: 9, ProjectID: 2, Status: "done", StoryPoints: &five})
	catalog.effort = append(catalog.effort, Effort{ProjectID: 2, StoryID: 9, Hours: amount("9")})
	catalog.defects = append(catalog.defects, Defect{ProjectID: 2, Status: "open", DetectionSprintID: 99})
	service := testService(catalog)

	report, err := service.ForProject(context.Background(), 1)
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if report.ActualHours.String() != "4.5" || report.DefectsDetected != 1 {
		t.Fatalf("leaked: %+v", report)
	}
}

func TestRejectsInvalidAndMissingIDs(t *testing.T) {
	service := testService(sampleCatalog())
	if _, err := service.ForProject(context.Background(), 0); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("zero project: %v", err)
	}
	if _, err := service.ForProject(context.Background(), 99); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("missing project: %v", err)
	}
	if _, err := service.ForSprint(context.Background(), 0); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("zero sprint: %v", err)
	}
	if _, err := service.ForSprint(context.Background(), 99); !errors.Is(err, ErrSprintNotFound) {
		t.Fatalf("missing sprint: %v", err)
	}
}

func TestEmptyProjectReturnsZeros(t *testing.T) {
	service := testService(&memoryCatalog{projects: map[int64]struct{}{1: {}}})
	report, err := service.ForProject(context.Background(), 1)
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	assertReport(t, report, Report{
		ProjectID:                1,
		Velocity:                 amount("0"),
		EstimatedHours:           amount("0"),
		ActualHours:              amount("0"),
		CompletedStoryPercentage: amount("0"),
	})
}

func assertReport(t *testing.T, got, want Report) {
	t.Helper()
	if got.ProjectID != want.ProjectID {
		t.Fatalf("project id: got %d want %d", got.ProjectID, want.ProjectID)
	}
	if (got.SprintID == nil) != (want.SprintID == nil) || (got.SprintID != nil && *got.SprintID != *want.SprintID) {
		t.Fatalf("sprint id: got %v want %v", got.SprintID, want.SprintID)
	}
	if got.PlannedStoryPoints != want.PlannedStoryPoints || got.CompletedStoryPoints != want.CompletedStoryPoints {
		t.Fatalf("points: planned %d/%d completed %d/%d", got.PlannedStoryPoints, want.PlannedStoryPoints, got.CompletedStoryPoints, want.CompletedStoryPoints)
	}
	if got.Velocity.String() != want.Velocity.String() {
		t.Fatalf("velocity: %s vs %s", got.Velocity, want.Velocity)
	}
	if got.EstimatedHours.String() != want.EstimatedHours.String() || got.ActualHours.String() != want.ActualHours.String() {
		t.Fatalf("hours: %s/%s vs %s/%s", got.EstimatedHours, got.ActualHours, want.EstimatedHours, want.ActualHours)
	}
	if got.Deviation != want.Deviation {
		t.Fatalf("deviation: %d vs %d", got.Deviation, want.Deviation)
	}
	if got.CompletedStoryPercentage.String() != want.CompletedStoryPercentage.String() {
		t.Fatalf("percentage: %s vs %s", got.CompletedStoryPercentage, want.CompletedStoryPercentage)
	}
	if got.DefectsDetected != want.DefectsDetected || got.DefectsResolved != want.DefectsResolved {
		t.Fatalf("defects: %d/%d vs %d/%d", got.DefectsDetected, got.DefectsResolved, want.DefectsDetected, want.DefectsResolved)
	}
}

func sampleCatalog() *memoryCatalog {
	eight, five := int64(8), int64(5)
	return &memoryCatalog{
		projects: map[int64]struct{}{1: {}},
		sprints: []Sprint{
			{ID: 10, ProjectID: 1, Status: "closed"},
			{ID: 11, ProjectID: 1, Status: "open"},
		},
		stories: []Story{
			{ID: 1, ProjectID: 1, Status: "done", StoryPoints: &eight},
			{ID: 2, ProjectID: 1, Status: "todo", StoryPoints: &five},
		},
		assignments: []Assignment{
			{SprintID: 10, StoryID: 1},
			{SprintID: 11, StoryID: 2},
		},
		effort: []Effort{
			{ProjectID: 1, StoryID: 1, Hours: amount("1.5")},
			{ProjectID: 1, StoryID: 1, Hours: amount("3")},
		},
		defects: []Defect{
			{ProjectID: 1, Status: "open", DetectionSprintID: 10},
		},
	}
}

type memoryCatalog struct {
	projects    map[int64]struct{}
	sprints     []Sprint
	stories     []Story
	assignments []Assignment
	effort      []Effort
	defects     []Defect
	err         error
}

func (c *memoryCatalog) ProjectExists(_ context.Context, projectID int64) error {
	if c.err != nil {
		return c.err
	}
	if _, ok := c.projects[projectID]; !ok {
		return ErrProjectNotFound
	}
	return nil
}

func (c *memoryCatalog) Load(_ context.Context, projectID int64) (Snapshot, error) {
	if c.err != nil {
		return Snapshot{}, c.err
	}
	if _, ok := c.projects[projectID]; !ok {
		return Snapshot{}, ErrProjectNotFound
	}
	snap := Snapshot{}
	for _, sprint := range c.sprints {
		if sprint.ProjectID == projectID {
			snap.Sprints = append(snap.Sprints, sprint)
		}
	}
	for _, story := range c.stories {
		if story.ProjectID == projectID {
			snap.Stories = append(snap.Stories, story)
		}
	}
	sprintIDs := map[int64]struct{}{}
	for _, sprint := range snap.Sprints {
		sprintIDs[sprint.ID] = struct{}{}
	}
	for _, assignment := range c.assignments {
		if _, ok := sprintIDs[assignment.SprintID]; ok {
			snap.Assignments = append(snap.Assignments, assignment)
		}
	}
	for _, entry := range c.effort {
		if entry.ProjectID == projectID {
			snap.Effort = append(snap.Effort, entry)
		}
	}
	for _, defect := range c.defects {
		if defect.ProjectID == projectID {
			snap.Defects = append(snap.Defects, defect)
		}
	}
	return snap, nil
}

func (c *memoryCatalog) Sprint(_ context.Context, sprintID int64) (Sprint, error) {
	if c.err != nil {
		return Sprint{}, c.err
	}
	for _, sprint := range c.sprints {
		if sprint.ID == sprintID {
			return sprint, nil
		}
	}
	return Sprint{}, ErrSprintNotFound
}

func testService(catalog *memoryCatalog) *Service {
	return NewService(catalog)
}

func amount(text string) Amount {
	value, err := parseAmount(text)
	if err != nil {
		panic(err)
	}
	return value
}
