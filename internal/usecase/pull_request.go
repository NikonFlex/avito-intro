package usecase

import (
	"context"
	"errors"
	"math/rand"
	"slices"
	"time"

	"avito-intro/internal/entity"
	"avito-intro/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrPRMerged    = errors.New("PR is already merged")
	ErrNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate = errors.New("no active replacement candidate in team")
)

var _ PullRequestUsecase = (*PullRequestUsecaseImpl)(nil)

type PullRequestUsecaseImpl struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
	logger   *zap.Logger
}

func NewPullRequestUsecase(
	userRepo repository.UserRepository,
	prRepo repository.PullRequestRepository,
	logger *zap.Logger,
) *PullRequestUsecaseImpl {
	return &PullRequestUsecaseImpl{
		userRepo: userRepo,
		prRepo:   prRepo,
		logger:   logger,
	}
}

func (u *PullRequestUsecaseImpl) CreatePR(ctx context.Context, prID uuid.UUID, prName string, authorID uuid.UUID) (entity.PullRequest, error) {
	u.logger.Info("creating pull request",
		zap.String("pr_id", prID.String()),
		zap.String("pr_name", prName),
		zap.String("author_id", authorID.String()),
	)

	if err := u.checkPRNotExists(ctx, prID); err != nil {
		return entity.PullRequest{}, err
	}

	author, err := u.getAuthor(ctx, authorID)
	if err != nil {
		return entity.PullRequest{}, err
	}

	reviewers, err := u.assignReviewers(ctx, author)
	if err != nil {
		return entity.PullRequest{}, err
	}

	pr := entity.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            entity.StatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now(),
		MergedAt:          nil,
	}

	if err := u.prRepo.CreatePullRequest(ctx, &pr); err != nil {
		u.logger.Error("failed to create PR", zap.Error(err))
		return entity.PullRequest{}, err
	}

	u.logger.Info("pull request created successfully",
		zap.String("pr_id", prID.String()),
		zap.Int("reviewers_count", len(reviewers)),
	)

	return pr, nil
}

func (u *PullRequestUsecaseImpl) MergePR(ctx context.Context, prID uuid.UUID) (entity.PullRequest, error) {
	u.logger.Info("merging pull request", zap.String("pr_id", prID.String()))

	pr, err := u.getPR(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, err
	}

	if pr.Status == entity.StatusMerged {
		u.logger.Info("PR already merged", zap.String("pr_id", prID.String()))
		return pr, nil
	}

	pr.Status = entity.StatusMerged
	now := time.Now()
	pr.MergedAt = &now

	if err := u.prRepo.UpdatePullRequest(ctx, &pr); err != nil {
		u.logger.Error("failed to update PR", zap.Error(err))
		return entity.PullRequest{}, err
	}

	u.logger.Info("pull request merged successfully", zap.String("pr_id", prID.String()))
	return pr, nil
}

func (u *PullRequestUsecaseImpl) ReassignReviewer(ctx context.Context, prID uuid.UUID, oldReviewerID uuid.UUID) (entity.PullRequest, uuid.UUID, error) {
	u.logger.Info("reassigning reviewer",
		zap.String("pr_id", prID.String()),
		zap.String("old_reviewer_id", oldReviewerID.String()),
	)

	pr, err := u.getPR(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, uuid.Nil, err
	}

	if err := u.checkPRNotMerged(pr); err != nil {
		return entity.PullRequest{}, uuid.Nil, err
	}

	if err := u.checkReviewerAssigned(pr, oldReviewerID); err != nil {
		return entity.PullRequest{}, uuid.Nil, err
	}

	oldReviewer, err := u.getUser(ctx, oldReviewerID)
	if err != nil {
		return entity.PullRequest{}, uuid.Nil, err
	}

	newReviewer, err := u.findReplacementReviewer(ctx, oldReviewer.TeamName, pr.AuthorID, pr.AssignedReviewers)
	if err != nil {
		return entity.PullRequest{}, uuid.Nil, err
	}

	u.replaceReviewer(&pr, oldReviewerID, newReviewer.UserID)

	if err := u.prRepo.UpdatePullRequest(ctx, &pr); err != nil {
		u.logger.Error("failed to update PR", zap.Error(err))
		return entity.PullRequest{}, uuid.Nil, err
	}

	u.logger.Info("reviewer reassigned successfully",
		zap.String("pr_id", prID.String()),
		zap.String("new_reviewer_id", newReviewer.UserID.String()),
	)

	return pr, newReviewer.UserID, nil
}

func (u *PullRequestUsecaseImpl) GetUserReviews(ctx context.Context, userID uuid.UUID) ([]entity.PullRequest, error) {
	u.logger.Debug("getting user reviews", zap.String("user_id", userID.String()))

	prs, err := u.prRepo.GetPullRequestsByReviewer(ctx, userID)
	if err != nil {
		u.logger.Error("failed to get PRs by reviewer", zap.Error(err))
		return nil, err
	}

	result := make([]entity.PullRequest, len(prs))
	for i, pr := range prs {
		result[i] = *pr
	}

	u.logger.Debug("user reviews retrieved",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(result)),
	)

	return result, nil
}

func (u *PullRequestUsecaseImpl) checkPRNotExists(ctx context.Context, prID uuid.UUID) error {
	exists, err := u.prRepo.PRExists(ctx, prID)
	if err != nil {
		u.logger.Error("failed to check PR existence", zap.Error(err))
		return err
	}

	if exists {
		u.logger.Warn("PR already exists", zap.String("pr_id", prID.String()))
		return repository.ErrAlreadyExists
	}

	return nil
}

func (u *PullRequestUsecaseImpl) getAuthor(ctx context.Context, authorID uuid.UUID) (entity.User, error) {
	author, err := u.userRepo.GetUser(ctx, authorID)
	if err != nil {
		u.logger.Error("failed to get author", zap.String("author_id", authorID.String()), zap.Error(err))
		return entity.User{}, err
	}
	return *author, nil
}

func (u *PullRequestUsecaseImpl) assignReviewers(ctx context.Context, author entity.User) ([]uuid.UUID, error) {
	teamMembers, err := u.userRepo.GetUsersByTeam(ctx, author.TeamName)
	if err != nil {
		u.logger.Error("failed to get team members", zap.Error(err))
		return nil, err
	}

	candidates := u.filterActiveCandidates(teamMembers, author.UserID)
	reviewers := u.selectRandomReviewers(candidates, 2)

	u.logger.Info("reviewers assigned",
		zap.Int("candidates", len(candidates)),
		zap.Int("selected", len(reviewers)),
	)

	return reviewers, nil
}

func (u *PullRequestUsecaseImpl) filterActiveCandidates(teamMembers []*entity.User, authorID uuid.UUID) []entity.User {
	var candidates []entity.User
	for _, member := range teamMembers {
		if member.UserID != authorID && member.IsActive {
			candidates = append(candidates, *member)
		}
	}
	return candidates
}

func (u *PullRequestUsecaseImpl) selectRandomReviewers(candidates []entity.User, maxCount int) []uuid.UUID {
	count := min(len(candidates), maxCount)
	if count == 0 {
		return []uuid.UUID{}
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	reviewers := make([]uuid.UUID, count)
	for i := range count {
		reviewers[i] = candidates[i].UserID
	}

	return reviewers
}

func (u *PullRequestUsecaseImpl) getPR(ctx context.Context, prID uuid.UUID) (entity.PullRequest, error) {
	pr, err := u.prRepo.GetPullRequest(ctx, prID)
	if err != nil {
		u.logger.Error("failed to get PR", zap.String("pr_id", prID.String()), zap.Error(err))
		return entity.PullRequest{}, err
	}
	return *pr, nil
}

func (u *PullRequestUsecaseImpl) getUser(ctx context.Context, userID uuid.UUID) (entity.User, error) {
	user, err := u.userRepo.GetUser(ctx, userID)
	if err != nil {
		u.logger.Error("failed to get user", zap.String("user_id", userID.String()), zap.Error(err))
		return entity.User{}, err
	}
	return *user, nil
}

func (u *PullRequestUsecaseImpl) checkPRNotMerged(pr entity.PullRequest) error {
	if pr.Status == entity.StatusMerged {
		u.logger.Warn("cannot reassign on merged PR", zap.String("pr_id", pr.PullRequestID.String()))
		return ErrPRMerged
	}
	return nil
}

func (u *PullRequestUsecaseImpl) checkReviewerAssigned(pr entity.PullRequest, reviewerID uuid.UUID) error {
	if slices.Contains(pr.AssignedReviewers, reviewerID) {
		return nil
	}

	u.logger.Warn("reviewer not assigned to PR",
		zap.String("pr_id", pr.PullRequestID.String()),
		zap.String("reviewer_id", reviewerID.String()),
	)
	return ErrNotAssigned
}

func (u *PullRequestUsecaseImpl) findReplacementReviewer(ctx context.Context, teamName string, authorID uuid.UUID, currentReviewers []uuid.UUID) (entity.User, error) {
	teamMembers, err := u.userRepo.GetUsersByTeam(ctx, teamName)
	if err != nil {
		u.logger.Error("failed to get team members", zap.Error(err))
		return entity.User{}, err
	}

	candidates := u.filterReplacementCandidates(teamMembers, authorID, currentReviewers)
	if len(candidates) == 0 {
		u.logger.Warn("no replacement candidates available")
		return entity.User{}, ErrNoCandidate
	}

	selected := candidates[rand.Intn(len(candidates))]
	return selected, nil
}

func (u *PullRequestUsecaseImpl) filterReplacementCandidates(teamMembers []*entity.User, authorID uuid.UUID, currentReviewers []uuid.UUID) []entity.User {
	var candidates []entity.User
	for _, member := range teamMembers {
		if !member.IsActive {
			continue
		}
		if member.UserID == authorID {
			continue
		}
		if u.isAlreadyReviewer(member.UserID, currentReviewers) {
			continue
		}
		candidates = append(candidates, *member)
	}
	return candidates
}

func (u *PullRequestUsecaseImpl) isAlreadyReviewer(userID uuid.UUID, reviewers []uuid.UUID) bool {
	return slices.Contains(reviewers, userID)
}

func (u *PullRequestUsecaseImpl) replaceReviewer(pr *entity.PullRequest, oldReviewerID, newReviewerID uuid.UUID) {
	for i, id := range pr.AssignedReviewers {
		if id == oldReviewerID {
			pr.AssignedReviewers[i] = newReviewerID
			return
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
