package civil

import (
	"context"
	"fmt"

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
	s.logger.Info("Fetching public posts")

	posts, err := s.repo.GetPublicPosts(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get public posts: %v", err)
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}

	s.logger.Infof("Retrieved %d public posts", len(posts))
	return posts, nil
}

func (s *Service) CreatePost(ctx context.Context, title, body string) (*repository.Post, error) {
	s.logger.Infof("Creating new post: %s", title)

	if title == "" || body == "" {
		return nil, fmt.Errorf("title and body are required")
	}

	post, err := s.repo.CreatePost(ctx, repository.CreatePostParams{
		Title: title,
		Body:  body,
	})
	if err != nil {
		s.logger.Errorf("Failed to create post: %v", err)
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	s.logger.Infof("Created post with ID: %s", post.ID)
	return &post, nil
}

func (s *Service) HealthCheck(ctx context.Context) error {
	s.logger.Info("Performing health check")
	return nil
}
