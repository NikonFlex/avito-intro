package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"avito-intro/internal/repository"
	"avito-intro/internal/usecase"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PullRequestController struct {
	prUC   usecase.PullRequestUsecase
	logger *zap.Logger
}

func NewPullRequestController(prUC usecase.PullRequestUsecase, logger *zap.Logger) *PullRequestController {
	return &PullRequestController{
		prUC:   prUC,
		logger: logger,
	}
}

func (c *PullRequestController) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid request body")
		return
	}

	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid pull_request_id format")
		return
	}

	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid author_id format")
		return
	}

	pr, err := c.prUC.CreatePR(r.Context(), prID, req.PullRequestName, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			c.sendError(w, http.StatusConflict, ErrorCodePRExists, "PR id already exists")
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			c.sendError(w, http.StatusNotFound, ErrorCodeNotFound, "author or team not found")
			return
		}
		c.logger.Error("failed to create PR", zap.Error(err))
		c.sendError(w, http.StatusInternalServerError, ErrorCodeInvalidInput, "internal server error")
		return
	}

	response := struct {
		PR PullRequestDTO `json:"pr"`
	}{
		PR: PullRequestToDTO(pr),
	}

	c.sendJSON(w, http.StatusCreated, response)
}

func (c *PullRequestController) MergePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid request body")
		return
	}

	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid pull_request_id format")
		return
	}

	pr, err := c.prUC.MergePR(r.Context(), prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.sendError(w, http.StatusNotFound, ErrorCodeNotFound, "PR not found")
			return
		}
		c.logger.Error("failed to merge PR", zap.Error(err))
		c.sendError(w, http.StatusInternalServerError, ErrorCodeInvalidInput, "internal server error")
		return
	}

	response := struct {
		PR PullRequestDTO `json:"pr"`
	}{
		PR: PullRequestToDTO(pr),
	}

	c.sendJSON(w, http.StatusOK, response)
}

func (c *PullRequestController) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid request body")
		return
	}

	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid pull_request_id format")
		return
	}

	oldReviewerID, err := uuid.Parse(req.OldUserID)
	if err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid old_user_id format")
		return
	}

	pr, newReviewerID, err := c.prUC.ReassignReviewer(r.Context(), prID, oldReviewerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.sendError(w, http.StatusNotFound, ErrorCodeNotFound, "PR or user not found")
			return
		}
		if errors.Is(err, usecase.ErrPRMerged) {
			c.sendError(w, http.StatusConflict, ErrorCodePRMerged, "cannot reassign on merged PR")
			return
		}
		if errors.Is(err, usecase.ErrNotAssigned) {
			c.sendError(w, http.StatusConflict, ErrorCodeNotAssigned, "reviewer is not assigned to this PR")
			return
		}
		if errors.Is(err, usecase.ErrNoCandidate) {
			c.sendError(w, http.StatusConflict, ErrorCodeNoCandidate, "no active replacement candidate in team")
			return
		}
		c.logger.Error("failed to reassign reviewer", zap.Error(err))
		c.sendError(w, http.StatusInternalServerError, ErrorCodeInvalidInput, "internal server error")
		return
	}

	response := struct {
		PR         PullRequestDTO `json:"pr"`
		ReplacedBy string         `json:"replaced_by"`
	}{
		PR:         PullRequestToDTO(pr),
		ReplacedBy: newReviewerID.String(),
	}

	c.sendJSON(w, http.StatusOK, response)
}

func (c *PullRequestController) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (c *PullRequestController) sendError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	c.sendJSON(w, status, resp)
}
