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

type UserController struct {
	userUC usecase.UserUsecase
	prUC   usecase.PullRequestUsecase
	logger *zap.Logger
}

func NewUserController(userUC usecase.UserUsecase, prUC usecase.PullRequestUsecase, logger *zap.Logger) *UserController {
	return &UserController{
		userUC: userUC,
		prUC:   prUC,
		logger: logger,
	}
}

func (c *UserController) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid user_id format")
		return
	}

	user, err := c.userUC.SetIsActive(r.Context(), userID, req.IsActive)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.sendError(w, http.StatusNotFound, ErrorCodeNotFound, "user not found")
			return
		}
		c.logger.Error("failed to set user active status", zap.Error(err))
		c.sendError(w, http.StatusInternalServerError, ErrorCodeInvalidInput, "internal server error")
		return
	}

	response := struct {
		User UserDTO `json:"user"`
	}{
		User: UserToDTO(user),
	}

	c.sendJSON(w, http.StatusOK, response)
}

func (c *UserController) GetReview(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "user_id query parameter is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid user_id format")
		return
	}

	prs, err := c.prUC.GetUserReviews(r.Context(), userID)
	if err != nil {
		c.logger.Error("failed to get user reviews", zap.Error(err))
		c.sendError(w, http.StatusInternalServerError, ErrorCodeInvalidInput, "internal server error")
		return
	}

	prDTOs := make([]PullRequestShortDTO, len(prs))
	for i, pr := range prs {
		prDTOs[i] = PullRequestToShortDTO(pr)
	}

	response := struct {
		UserID       string                `json:"user_id"`
		PullRequests []PullRequestShortDTO `json:"pull_requests"`
	}{
		UserID:       userIDStr,
		PullRequests: prDTOs,
	}

	c.sendJSON(w, http.StatusOK, response)
}

func (c *UserController) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (c *UserController) sendError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	c.sendJSON(w, status, resp)
}
