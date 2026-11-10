package effort

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const maxActivityLength = 200

var (
	ErrInvalidID        = errors.New("project, story and member ids must be positive")
	ErrDateRequired     = errors.New("effort date is required")
	ErrActivityRequired = errors.New("activity is required")
	ErrActivityTooLong  = errors.New("activity must be at most 200 characters")
	ErrHoursRequired    = errors.New("hours are required")
	ErrHoursNotPositive = errors.New("hours must be greater than zero")
	ErrInvalidHours     = errors.New("hours are invalid")
	ErrStoryNotFound    = errors.New("story not found")
	ErrMemberNotFound   = errors.New("member not found")
	hoursToken          = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,7})(\.[0-9]{1,2})?$`)
)

type Service struct {
	repo    Repository
	stories StoryLookup
	members MemberLookup
	now     func() time.Time
}

func NewService(repo Repository, stories StoryLookup, members MemberLookup) *Service {
	return &Service{repo: repo, stories: stories, members: members, now: time.Now}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (Entry, error) {
	if input.ProjectID < 1 || input.StoryID < 1 || input.MemberID < 1 {
		return Entry{}, ErrInvalidID
	}
	if input.Date.IsZero() {
		return Entry{}, ErrDateRequired
	}
	activity := strings.TrimSpace(input.Activity)
	if activity == "" {
		return Entry{}, ErrActivityRequired
	}
	if utf8.RuneCountInString(activity) > maxActivityLength {
		return Entry{}, ErrActivityTooLong
	}
	hours, err := parseHours(input.Hours)
	if err != nil {
		return Entry{}, err
	}
	if err := s.stories.Find(ctx, input.ProjectID, input.StoryID); err != nil {
		return Entry{}, err
	}
	if err := s.members.Find(ctx, input.ProjectID, input.MemberID); err != nil {
		return Entry{}, err
	}
	return s.repo.Save(ctx, Entry{
		ProjectID: input.ProjectID,
		StoryID:   input.StoryID,
		MemberID:  input.MemberID,
		Date:      calendarDate(input.Date),
		Activity:  activity,
		Hours:     hours,
		CreatedAt: s.now().UTC(),
	})
}

func parseHours(raw string) (Hours, error) {
	text := strings.TrimSpace(raw)
	if text == "" || text == "null" {
		return 0, ErrHoursRequired
	}
	if !hoursToken.MatchString(text) {
		return 0, ErrInvalidHours
	}
	negative := strings.HasPrefix(text, "-")
	if negative {
		text = text[1:]
	}
	parts := strings.Split(text, ".")
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, ErrInvalidHours
	}
	var frac int64
	if len(parts) == 2 {
		fracText := parts[1]
		if len(fracText) == 1 {
			fracText += "0"
		}
		frac, err = strconv.ParseInt(fracText, 10, 64)
		if err != nil {
			return 0, ErrInvalidHours
		}
	}
	hundredths := whole*100 + frac
	if negative {
		hundredths = -hundredths
	}
	if hundredths <= 0 {
		return 0, ErrHoursNotPositive
	}
	return Hours(hundredths), nil
}

func calendarDate(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
