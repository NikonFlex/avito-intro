package usecase

import (
	"context"

	"avito-intro/internal/entity"
	"avito-intro/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var _ UserUsecase = (*UserUsecaseImpl)(nil)

type UserUsecaseImpl struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewUserUsecase(
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *UserUsecaseImpl {
	return &UserUsecaseImpl{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (u *UserUsecaseImpl) SetIsActive(ctx context.Context, userID uuid.UUID, isActive bool) (entity.User, error) {
	u.logger.Info("setting user active status",
		zap.String("user_id", userID.String()),
		zap.Bool("is_active", isActive),
	)

	user, err := u.getUser(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}

	updatedUser := u.updateUserActiveStatus(user, isActive)

	if err := u.saveUser(ctx, &updatedUser); err != nil {
		return entity.User{}, err
	}

	u.logger.Info("user active status updated successfully",
		zap.String("user_id", userID.String()),
		zap.Bool("is_active", isActive),
	)

	return updatedUser, nil
}

func (u *UserUsecaseImpl) getUser(ctx context.Context, userID uuid.UUID) (entity.User, error) {
	user, err := u.userRepo.GetUser(ctx, userID)
	if err != nil {
		u.logger.Error("failed to get user", zap.String("user_id", userID.String()), zap.Error(err))
		return entity.User{}, err
	}
	return *user, nil
}

func (u *UserUsecaseImpl) updateUserActiveStatus(user entity.User, isActive bool) entity.User {
	user.IsActive = isActive
	return user
}

func (u *UserUsecaseImpl) saveUser(ctx context.Context, user *entity.User) error {
	if err := u.userRepo.UpdateUser(ctx, user); err != nil {
		u.logger.Error("failed to update user",
			zap.String("user_id", user.UserID.String()),
			zap.Error(err),
		)
		return err
	}
	return nil
}
