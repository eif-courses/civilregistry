package web

import (
	"context"

	"github.com/eif-courses/civilregistry/internal/repository"
	"go.uber.org/zap"
)

type Service struct {
	repo   *repository.Queries
	logger *zap.SugaredLogger
}

func NewService(repo *repository.Queries, logger *zap.SugaredLogger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetPublicPosts(ctx context.Context) ([]repository.Post, error) {
	s.logger.Info("Fetching posts for web display")

	posts, err := s.repo.GetPublicPosts(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get posts for web: %v", err)
		return nil, err
	}

	return posts, nil
}
