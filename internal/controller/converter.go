package controller

import (
	"time"

	"avito-intro/internal/entity"

	"github.com/google/uuid"
)

func UserToDTO(user entity.User) UserDTO {
	return UserDTO{
		UserID:   user.UserID.String(),
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func TeamMemberToDTO(user entity.User) TeamMemberDTO {
	return TeamMemberDTO{
		UserID:   user.UserID.String(),
		Username: user.Username,
		IsActive: user.IsActive,
	}
}

func TeamToDTO(team entity.Team, members []entity.User) TeamDTO {
	memberDTOs := make([]TeamMemberDTO, len(members))
	for i, member := range members {
		memberDTOs[i] = TeamMemberToDTO(member)
	}

	return TeamDTO{
		TeamName: team.TeamName,
		Members:  memberDTOs,
	}
}

func PullRequestToDTO(pr entity.PullRequest) PullRequestDTO {
	reviewerIDs := make([]string, len(pr.AssignedReviewers))
	for i, id := range pr.AssignedReviewers {
		reviewerIDs[i] = id.String()
	}

	return PullRequestDTO{
		PullRequestID:     pr.PullRequestID.String(),
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID.String(),
		Status:            string(pr.Status),
		AssignedReviewers: reviewerIDs,
		CreatedAt:         formatTimePtr(&pr.CreatedAt),
		MergedAt:          formatTimePtr(pr.MergedAt),
	}
}

func PullRequestToShortDTO(pr entity.PullRequest) PullRequestShortDTO {
	return PullRequestShortDTO{
		PullRequestID:   pr.PullRequestID.String(),
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID.String(),
		Status:          string(pr.Status),
	}
}

func TeamMemberDTOToEntity(dto TeamMemberDTO, teamName string) (entity.User, error) {
	userID, err := uuid.Parse(dto.UserID)
	if err != nil {
		return entity.User{}, err
	}

	return entity.User{
		UserID:   userID,
		Username: dto.Username,
		TeamName: teamName,
		IsActive: dto.IsActive,
	}, nil
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
