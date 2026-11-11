package reports

import "context"

// Repository loads all project data from one consistent read.
type Repository interface {
	LoadProject(context.Context, int64) (Source, error)
}
