package usecase

import (
	"context"

	"avito-intro/internal/entity"
	"avito-intro/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var _ TeamUsecase = (*TeamUsecaseImpl)(nil)

type TeamUsecaseImpl struct {
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	logger   *zap.Logger
}

func NewTeamUsecase(
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
	logger *zap.Logger,
) *TeamUsecaseImpl {
	return &TeamUsecaseImpl{
		userRepo: userRepo,
		teamRepo: teamRepo,
		logger:   logger,
	}
}

func (u *TeamUsecaseImpl) AddTeam(ctx context.Context, team entity.Team, members []entity.User) (entity.Team, error) {
	u.logger.Info("adding team",
		zap.String("team_name", team.TeamName),
		zap.Int("members_count", len(members)),
	)

	if err := u.checkTeamNotExists(ctx, team.TeamName); err != nil {
		return entity.Team{}, err
	}

	if err := u.createOrUpdateMembers(ctx, members); err != nil {
		return entity.Team{}, err
	}

	if err := u.createTeam(ctx, &team); err != nil {
		return entity.Team{}, err
	}

	u.logger.Info("team created successfully", zap.String("team_name", team.TeamName))
	return team, nil
}

func (u *TeamUsecaseImpl) GetTeam(ctx context.Context, teamName string) (entity.Team, []entity.User, error) {
	u.logger.Debug("getting team", zap.String("team_name", teamName))

	team, err := u.getTeamByName(ctx, teamName)
	if err != nil {
		return entity.Team{}, nil, err
	}

	users, err := u.getTeamMembers(ctx, team.Members)
	if err != nil {
		return entity.Team{}, nil, err
	}

	u.logger.Debug("team retrieved successfully",
		zap.String("team_name", teamName),
		zap.Int("members_count", len(users)),
	)

	return team, users, nil
}

func (u *TeamUsecaseImpl) checkTeamNotExists(ctx context.Context, teamName string) error {
	exists, err := u.teamRepo.TeamExists(ctx, teamName)
	if err != nil {
		u.logger.Error("failed to check team existence", zap.Error(err))
		return err
	}

	if exists {
		u.logger.Warn("team already exists", zap.String("team_name", teamName))
		return repository.ErrAlreadyExists
	}

	return nil
}

func (u *TeamUsecaseImpl) createOrUpdateMembers(ctx context.Context, members []entity.User) error {
	for _, member := range members {
		exists, err := u.userRepo.UserExists(ctx, member.UserID)
		if err != nil {
			u.logger.Error("failed to check user existence",
				zap.String("user_id", member.UserID.String()),
				zap.Error(err),
			)
			return err
		}

		if exists {
			if err := u.userRepo.UpdateUser(ctx, &member); err != nil {
				u.logger.Error("failed to update user",
					zap.String("user_id", member.UserID.String()),
					zap.Error(err),
				)
				return err
			}
			continue
		}

		if err := u.userRepo.CreateUser(ctx, &member); err != nil {
			u.logger.Error("failed to create user",
				zap.String("user_id", member.UserID.String()),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func (u *TeamUsecaseImpl) createTeam(ctx context.Context, team *entity.Team) error {
	if err := u.teamRepo.CreateTeam(ctx, team); err != nil {
		u.logger.Error("failed to create team", zap.Error(err))
		return err
	}
	return nil
}

func (u *TeamUsecaseImpl) getTeamByName(ctx context.Context, teamName string) (entity.Team, error) {
	team, err := u.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		u.logger.Error("failed to get team", zap.Error(err))
		return entity.Team{}, err
	}
	return *team, nil
}

func (u *TeamUsecaseImpl) getTeamMembers(ctx context.Context, memberIDs []uuid.UUID) ([]entity.User, error) {
	users, err := u.userRepo.GetUsersByIDs(ctx, memberIDs)
	if err != nil {
		u.logger.Error("failed to get team members", zap.Error(err))
		return nil, err
	}

	result := make([]entity.User, len(users))
	for i, user := range users {
		result[i] = *user
	}

	return result, nil
}
