package repository

import (
	"context"

	"avito-intro/internal/entity"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	UpdateUser(ctx context.Context, user *entity.User) error
	GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error)
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
	GetUsersByTeam(ctx context.Context, teamName string) ([]*entity.User, error)
	GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]*entity.User, error)
}

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *entity.Team) error
	GetTeam(ctx context.Context, teamName string) (*entity.Team, error)
	TeamExists(ctx context.Context, teamName string) (bool, error)
}

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr *entity.PullRequest) error
	GetPullRequest(ctx context.Context, prID uuid.UUID) (*entity.PullRequest, error)
	UpdatePullRequest(ctx context.Context, pr *entity.PullRequest) error
	GetPullRequestsByReviewer(ctx context.Context, userID uuid.UUID) ([]*entity.PullRequest, error)
	PRExists(ctx context.Context, prID uuid.UUID) (bool, error)
}
