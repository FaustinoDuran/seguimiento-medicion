package defects

import "context"

type ResolutionService struct {
	repo ResolutionRepository
}

func NewResolutionService(repo ResolutionRepository) *ResolutionService {
	return &ResolutionService{repo: repo}
}

func (s *ResolutionService) Resolve(ctx context.Context, input ResolveInput) (Defect, error) {
	return s.repo.Resolve(ctx, input)
}
