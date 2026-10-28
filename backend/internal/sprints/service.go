package sprints

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrGoalRequired    = errors.New("sprint goal is required")
	ErrProjectNotFound = errors.New("project not found")
	ErrNotFound        = errors.New("sprint not found")
	ErrStoryNotFound   = errors.New("story not found")
)

type Service struct {
	repo     Repository
	projects ProjectLookup
	stories  StoryLookup
	now      func() time.Time
}

func NewService(repo Repository, projects ProjectLookup, stories StoryLookup) *Service {
	return &Service{repo: repo, projects: projects, stories: stories, now: time.Now}
}

func (s *Service) Create(ctx context.Context, input CreateSprint) (Sprint, error) {
	goal := strings.TrimSpace(input.Goal)
	if goal == "" {
		return Sprint{}, ErrGoalRequired
	}

	if err := s.projects.Exists(ctx, input.ProjectID); err != nil {
		return Sprint{}, err
	}

	return s.repo.Save(ctx, Sprint{
		ProjectID: input.ProjectID,
		Goal:      goal,
		Status:    StatusOpen,
		CreatedAt: s.now().UTC(),
	})
}

func (s *Service) Get(ctx context.Context, id int64) (Sprint, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Assign(ctx context.Context, sprintID, storyID int64) (AssignedStory, error) {
	sprint, err := s.repo.FindByID(ctx, sprintID)
	if err != nil {
		return AssignedStory{}, err
	}
	story, err := s.stories.FindInProject(ctx, sprint.ProjectID, storyID)
	if err != nil {
		return AssignedStory{}, err
	}
	assignment := Assignment{SprintID: sprint.ID, StoryID: story.ID}
	if err := s.repo.SaveAssignment(ctx, assignment); err != nil {
		return AssignedStory{}, err
	}
	return assignedStory(sprint.ID, story), nil
}

func (s *Service) ListStories(ctx context.Context, sprintID int64) ([]AssignedStory, error) {
	sprint, err := s.repo.FindByID(ctx, sprintID)
	if err != nil {
		return nil, err
	}
	assignments, err := s.repo.ListAssignments(ctx, sprint.ID)
	if err != nil {
		return nil, err
	}
	listed := make([]AssignedStory, 0, len(assignments))
	for _, assignment := range assignments {
		story, err := s.stories.FindInProject(ctx, sprint.ProjectID, assignment.StoryID)
		if err != nil {
			return nil, err
		}
		listed = append(listed, assignedStory(sprint.ID, story))
	}
	return listed, nil
}

func assignedStory(sprintID int64, story StoryRef) AssignedStory {
	return AssignedStory{
		SprintID:   sprintID,
		StoryID:    story.ID,
		Identifier: story.Identifier,
		Title:      story.Title,
	}
}
