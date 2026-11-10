package effort

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRegisterStoresEffort(t *testing.T) {
	service, repo := testService()

	saved, err := service.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if saved.ID == 0 || len(repo.entries) != 1 {
		t.Fatalf("saved: %+v entries: %d", saved, len(repo.entries))
	}
	if saved.ProjectID != 1 || saved.StoryID != 4 || saved.MemberID != 7 {
		t.Fatalf("ids: %+v", saved)
	}
	if !saved.Date.Equal(time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("date: %s", saved.Date)
	}
	if saved.Activity != "implementar el alta" || saved.Hours.String() != "1.5" {
		t.Fatalf("effort: %+v", saved)
	}
}

func TestRegisterAppendsAnotherEntry(t *testing.T) {
	service, repo := testService()
	ctx := context.Background()
	first, err := service.Register(ctx, validInput())
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	secondInput := validInput()
	secondInput.Activity = "revisar el alta"
	second, err := service.Register(ctx, secondInput)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if first.ID == second.ID || len(repo.entries) != 2 {
		t.Fatalf("ids: %d %d entries: %d", first.ID, second.ID, len(repo.entries))
	}
}

func TestRegisterNormalizesHours(t *testing.T) {
	cases := []struct {
		hours string
		want  string
	}{
		{hours: "1", want: "1"},
		{hours: "1.5", want: "1.5"},
		{hours: "1.50", want: "1.5"},
		{hours: "0.01", want: "0.01"},
	}
	for _, tc := range cases {
		t.Run(tc.hours, func(t *testing.T) {
			service, repo := testService()
			input := validInput()
			input.Hours = tc.hours
			saved, err := service.Register(context.Background(), input)
			if err != nil {
				t.Fatalf("register: %v", err)
			}
			if saved.Hours.String() != tc.want || len(repo.entries) != 1 {
				t.Fatalf("hours: got %s entries %d", saved.Hours, len(repo.entries))
			}
		})
	}
}

func TestRegisterRejectsInvalidHours(t *testing.T) {
	cases := []struct {
		name  string
		hours string
		want  error
	}{
		{name: "empty", hours: "", want: ErrHoursRequired},
		{name: "blank", hours: "   ", want: ErrHoursRequired},
		{name: "null", hours: "null", want: ErrHoursRequired},
		{name: "zero", hours: "0", want: ErrHoursNotPositive},
		{name: "zero with decimals", hours: "0.00", want: ErrHoursNotPositive},
		{name: "negative", hours: "-1", want: ErrHoursNotPositive},
		{name: "negative fraction", hours: "-0.5", want: ErrHoursNotPositive},
		{name: "three decimals", hours: "1.234", want: ErrInvalidHours},
		{name: "above storage limit", hours: "100000000", want: ErrInvalidHours},
		{name: "text", hours: `"1.5"`, want: ErrInvalidHours},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, repo := testService()
			input := validInput()
			input.Hours = tc.hours
			_, err := service.Register(context.Background(), input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error: got %v want %v", err, tc.want)
			}
			if len(repo.entries) != 0 {
				t.Fatalf("persisted %d entries", len(repo.entries))
			}
		})
	}
}

func TestRegisterRejectsInvalidActivityAndDate(t *testing.T) {
	cases := []struct {
		name  string
		input RegisterInput
		want  error
	}{
		{name: "empty activity", input: withActivity("   "), want: ErrActivityRequired},
		{name: "activity too long", input: withActivity(strings.Repeat("a", 201)), want: ErrActivityTooLong},
		{name: "missing date", input: withDate(time.Time{}), want: ErrDateRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, repo := testService()
			_, err := service.Register(context.Background(), tc.input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error: got %v want %v", err, tc.want)
			}
			if len(repo.entries) != 0 {
				t.Fatalf("persisted %d entries", len(repo.entries))
			}
		})
	}
}

func TestRegisterTrimsActivityAndKeepsTwoHundredCharacters(t *testing.T) {
	service, _ := testService()
	input := validInput()
	input.Activity = "  " + strings.Repeat("á", 200) + "  "
	saved, err := service.Register(context.Background(), input)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if saved.Activity != strings.Repeat("á", 200) {
		t.Fatalf("activity length: %d", len([]rune(saved.Activity)))
	}
}

func TestRegisterRejectsUnknownStoryOrMember(t *testing.T) {
	cases := []struct {
		name  string
		input RegisterInput
		want  error
	}{
		{name: "missing story", input: withStory(99), want: ErrStoryNotFound},
		{name: "story in another project", input: withProject(2), want: ErrStoryNotFound},
		{name: "missing member", input: withMember(99), want: ErrMemberNotFound},
		{name: "member in another project", input: withMember(8), want: ErrMemberNotFound},
		{name: "non positive ids", input: withMember(0), want: ErrInvalidID},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, repo := testService()
			_, err := service.Register(context.Background(), tc.input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error: got %v want %v", err, tc.want)
			}
			if len(repo.entries) != 0 {
				t.Fatalf("persisted %d entries", len(repo.entries))
			}
		})
	}
}

func TestRegisterChecksFormatBeforeLookups(t *testing.T) {
	service, repo := testService()
	input := withStory(99)
	input.Hours = "-1"
	_, err := service.Register(context.Background(), input)
	if !errors.Is(err, ErrHoursNotPositive) {
		t.Fatalf("error: got %v", err)
	}
	if len(repo.entries) != 0 {
		t.Fatalf("persisted %d entries", len(repo.entries))
	}
}

func TestRegisterDoesNotSaveWhenPersistenceFails(t *testing.T) {
	service, repo := testService()
	repo.err = errors.New("db down")
	_, err := service.Register(context.Background(), validInput())
	if err == nil || errors.Is(err, ErrHoursRequired) {
		t.Fatalf("error: %v", err)
	}
	if len(repo.entries) != 0 {
		t.Fatalf("persisted %d entries", len(repo.entries))
	}
}

func validInput() RegisterInput {
	return RegisterInput{
		ProjectID: 1,
		StoryID:   4,
		MemberID:  7,
		Date:      time.Date(2026, 11, 10, 15, 4, 0, 0, time.UTC),
		Activity:  "  implementar el alta  ",
		Hours:     "1.50",
	}
}

func withActivity(activity string) RegisterInput {
	input := validInput()
	input.Activity = activity
	return input
}

func withDate(date time.Time) RegisterInput {
	input := validInput()
	input.Date = date
	return input
}

func withStory(storyID int64) RegisterInput {
	input := validInput()
	input.StoryID = storyID
	return input
}

func withProject(projectID int64) RegisterInput {
	input := validInput()
	input.ProjectID = projectID
	return input
}

func withMember(memberID int64) RegisterInput {
	input := validInput()
	input.MemberID = memberID
	return input
}

type memoryRepository struct {
	nextID  int64
	entries []Entry
	err     error
}

func (r *memoryRepository) Save(_ context.Context, entry Entry) (Entry, error) {
	if r.err != nil {
		return Entry{}, r.err
	}
	r.nextID++
	entry.ID = r.nextID
	r.entries = append(r.entries, entry)
	return entry, nil
}

type fakeStories struct{}

func (fakeStories) Find(_ context.Context, projectID, storyID int64) error {
	if projectID == 1 && storyID == 4 {
		return nil
	}
	return ErrStoryNotFound
}

type fakeMembers struct{}

func (fakeMembers) Find(_ context.Context, projectID, memberID int64) error {
	if projectID == 1 && memberID == 7 {
		return nil
	}
	return ErrMemberNotFound
}

func testService() (*Service, *memoryRepository) {
	repo := &memoryRepository{}
	return NewService(repo, fakeStories{}, fakeMembers{}), repo
}
