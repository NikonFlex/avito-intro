package usecase

import (
	"context"

	"avito-intro/internal/entity"

	"github.com/google/uuid"
)

type TeamUsecase interface {
	AddTeam(ctx context.Context, team entity.Team, members []entity.User) (entity.Team, error)
	GetTeam(ctx context.Context, teamName string) (entity.Team, []entity.User, error)
}

type UserUsecase interface {
	SetIsActive(ctx context.Context, userID uuid.UUID, isActive bool) (entity.User, error)
}

type PullRequestUsecase interface {
	CreatePR(ctx context.Context, prID uuid.UUID, prName string, authorID uuid.UUID) (entity.PullRequest, error)
	MergePR(ctx context.Context, prID uuid.UUID) (entity.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID uuid.UUID, oldReviewerID uuid.UUID) (entity.PullRequest, uuid.UUID, error)
	GetUserReviews(ctx context.Context, userID uuid.UUID) ([]entity.PullRequest, error)
}
