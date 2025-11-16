package entity

import (
	"time"

	"github.com/google/uuid"
)

type PullRequestStatus string

const (
	StatusOpen   PullRequestStatus = "OPEN"
	StatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     uuid.UUID
	PullRequestName   string
	AuthorID          uuid.UUID
	Status            PullRequestStatus
	AssignedReviewers []uuid.UUID
	CreatedAt         time.Time
	MergedAt          *time.Time
}
