package mcp

import (
	"context"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
)

type ProjectReader interface {
	List(context.Context) ([]projects.Project, error)
	Get(context.Context, int64) (projects.Project, error)
}

type StoryReader interface {
	List(context.Context, int64) ([]backlog.Story, error)
	Get(context.Context, int64, int64) (backlog.Story, error)
}

type SprintReader interface {
	List(context.Context, int64) ([]sprints.Sprint, error)
	Get(context.Context, int64) (sprints.Sprint, error)
	ListStories(context.Context, int64) ([]sprints.AssignedStory, error)
}

type MetricsReader interface {
	ForProject(context.Context, int64) (metrics.Report, error)
	ForSprint(context.Context, int64) (metrics.Report, error)
}

type ReportReader interface {
	ForProject(context.Context, int64) (reports.Report, error)
	ForSprint(context.Context, int64, int64) (reports.Report, error)
}

// Readers limits the adapter to the query methods of the existing services.
type Readers struct {
	Projects ProjectReader
	Stories  StoryReader
	Sprints  SprintReader
	Metrics  MetricsReader
	Reports  ReportReader
}

var (
	_ ProjectReader = (*projects.Service)(nil)
	_ StoryReader   = (*backlog.Service)(nil)
	_ SprintReader  = (*sprints.Service)(nil)
	_ MetricsReader = (*metrics.Service)(nil)
	_ ReportReader  = (*reports.Service)(nil)
)
