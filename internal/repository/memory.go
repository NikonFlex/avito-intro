package repository

import (
	"context"
	"errors"
	"sync"

	"avito-intro/internal/entity"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

var (
	_ UserRepository        = (*MemoryRepository)(nil)
	_ TeamRepository        = (*MemoryRepository)(nil)
	_ PullRequestRepository = (*MemoryRepository)(nil)
)

type MemoryRepository struct {
	mu           sync.RWMutex
	users        map[uuid.UUID]*entity.User
	teams        map[string]*entity.Team
	pullRequests map[uuid.UUID]*entity.PullRequest
	logger       *zap.Logger
}

func NewMemoryRepository(logger *zap.Logger) *MemoryRepository {
	return &MemoryRepository{
		users:        make(map[uuid.UUID]*entity.User),
		teams:        make(map[string]*entity.Team),
		pullRequests: make(map[uuid.UUID]*entity.PullRequest),
		logger:       logger,
	}
}

// UserRepository implementation

func (r *MemoryRepository) CreateUser(ctx context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.UserID]; exists {
		r.logger.Warn("user already exists", zap.String("user_id", user.UserID.String()))
		return ErrAlreadyExists
	}

	r.logger.Info("creating user",
		zap.String("user_id", user.UserID.String()),
		zap.String("username", user.Username),
		zap.String("team_name", user.TeamName),
		zap.Bool("is_active", user.IsActive),
	)

	r.users[user.UserID] = user
	return nil
}

func (r *MemoryRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.UserID]; !exists {
		r.logger.Warn("user not found for update", zap.String("user_id", user.UserID.String()))
		return ErrNotFound
	}

	r.logger.Info("updating user",
		zap.String("user_id", user.UserID.String()),
		zap.String("username", user.Username),
		zap.String("team_name", user.TeamName),
		zap.Bool("is_active", user.IsActive),
	)

	r.users[user.UserID] = user
	return nil
}

func (r *MemoryRepository) GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		r.logger.Warn("user not found", zap.String("user_id", userID.String()))
		return nil, ErrNotFound
	}

	r.logger.Debug("user retrieved", zap.String("user_id", userID.String()))
	return user, nil
}

func (r *MemoryRepository) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.users[userID]
	return exists, nil
}

func (r *MemoryRepository) GetUsersByTeam(ctx context.Context, teamName string) ([]*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []*entity.User
	for _, user := range r.users {
		if user.TeamName == teamName {
			users = append(users, user)
		}
	}

	r.logger.Debug("users retrieved by team",
		zap.String("team_name", teamName),
		zap.Int("count", len(users)),
	)
	return users, nil
}

func (r *MemoryRepository) GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*entity.User, 0, len(userIDs))
	for _, id := range userIDs {
		if user, exists := r.users[id]; exists {
			users = append(users, user)
		}
	}

	r.logger.Debug("users retrieved by IDs",
		zap.Int("requested", len(userIDs)),
		zap.Int("found", len(users)),
	)
	return users, nil
}

// TeamRepository implementation

func (r *MemoryRepository) CreateTeam(ctx context.Context, team *entity.Team) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.teams[team.TeamName]; exists {
		r.logger.Warn("team already exists", zap.String("team_name", team.TeamName))
		return ErrAlreadyExists
	}

	r.logger.Info("creating team",
		zap.String("team_name", team.TeamName),
		zap.Int("members_count", len(team.Members)),
	)

	r.teams[team.TeamName] = team
	return nil
}

func (r *MemoryRepository) GetTeam(ctx context.Context, teamName string) (*entity.Team, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	team, exists := r.teams[teamName]
	if !exists {
		r.logger.Warn("team not found", zap.String("team_name", teamName))
		return nil, ErrNotFound
	}

	r.logger.Debug("team retrieved", zap.String("team_name", teamName))
	return team, nil
}

func (r *MemoryRepository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.teams[teamName]
	return exists, nil
}

// PullRequestRepository implementation

func (r *MemoryRepository) CreatePullRequest(ctx context.Context, pr *entity.PullRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.pullRequests[pr.PullRequestID]; exists {
		r.logger.Warn("pull request already exists", zap.String("pr_id", pr.PullRequestID.String()))
		return ErrAlreadyExists
	}

	r.logger.Info("creating pull request",
		zap.String("pr_id", pr.PullRequestID.String()),
		zap.String("pr_name", pr.PullRequestName),
		zap.String("author_id", pr.AuthorID.String()),
		zap.Int("reviewers_count", len(pr.AssignedReviewers)),
	)

	r.pullRequests[pr.PullRequestID] = pr
	return nil
}

func (r *MemoryRepository) GetPullRequest(ctx context.Context, prID uuid.UUID) (*entity.PullRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pr, exists := r.pullRequests[prID]
	if !exists {
		r.logger.Warn("pull request not found", zap.String("pr_id", prID.String()))
		return nil, ErrNotFound
	}

	r.logger.Debug("pull request retrieved", zap.String("pr_id", prID.String()))
	return pr, nil
}

func (r *MemoryRepository) UpdatePullRequest(ctx context.Context, pr *entity.PullRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.pullRequests[pr.PullRequestID]; !exists {
		r.logger.Warn("pull request not found for update", zap.String("pr_id", pr.PullRequestID.String()))
		return ErrNotFound
	}

	r.logger.Info("updating pull request",
		zap.String("pr_id", pr.PullRequestID.String()),
		zap.String("status", string(pr.Status)),
	)

	r.pullRequests[pr.PullRequestID] = pr
	return nil
}

func (r *MemoryRepository) GetPullRequestsByReviewer(ctx context.Context, userID uuid.UUID) ([]*entity.PullRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var prs []*entity.PullRequest
	for _, pr := range r.pullRequests {
		for _, reviewerID := range pr.AssignedReviewers {
			if reviewerID == userID {
				prs = append(prs, pr)
				break
			}
		}
	}

	r.logger.Debug("pull requests retrieved by reviewer",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(prs)),
	)
	return prs, nil
}

func (r *MemoryRepository) PRExists(ctx context.Context, prID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.pullRequests[prID]
	return exists, nil
}
