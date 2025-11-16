package controller

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type PullRequestDTO struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type ErrorCode string

const (
	ErrorCodeTeamExists   ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists     ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged     ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned  ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate  ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
)

type ErrorResponse struct {
	Error struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
	} `json:"error"`
}
