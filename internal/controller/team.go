package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"avito-intro/internal/entity"
	"avito-intro/internal/repository"
	"avito-intro/internal/usecase"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TeamController struct {
	teamUC usecase.TeamUsecase
	logger *zap.Logger
}

func NewTeamController(teamUC usecase.TeamUsecase, logger *zap.Logger) *TeamController {
	return &TeamController{
		teamUC: teamUC,
		logger: logger,
	}
}

func (c *TeamController) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req TeamDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid request body")
		return
	}

	memberIDs := make([]uuid.UUID, len(req.Members))
	members := make([]entity.User, len(req.Members))
	for i, m := range req.Members {
		user, err := TeamMemberDTOToEntity(m, req.TeamName)
		if err != nil {
			c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "invalid user_id format")
			return
		}
		members[i] = user
		memberIDs[i] = user.UserID
	}

	team := entity.Team{
		TeamName: req.TeamName,
		Members:  memberIDs,
	}

	createdTeam, err := c.teamUC.AddTeam(r.Context(), team, members)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			c.sendError(w, http.StatusBadRequest, ErrorCodeTeamExists, "team_name already exists")
			return
		}
		c.logger.Error("failed to add team", zap.Error(err))
		c.sendError(w, http.StatusInternalServerError, ErrorCodeInvalidInput, "internal server error")
		return
	}

	_, retrievedMembers, err := c.teamUC.GetTeam(r.Context(), createdTeam.TeamName)
	if err != nil {
		c.logger.Error("failed to get team", zap.Error(err))
		c.sendError(w, http.StatusInternalServerError, ErrorCodeInvalidInput, "internal server error")
		return
	}

	response := struct {
		Team TeamDTO `json:"team"`
	}{
		Team: TeamToDTO(createdTeam, retrievedMembers),
	}

	c.sendJSON(w, http.StatusCreated, response)
}

func (c *TeamController) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		c.sendError(w, http.StatusBadRequest, ErrorCodeInvalidInput, "team_name query parameter is required")
		return
	}

	team, members, err := c.teamUC.GetTeam(r.Context(), teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.sendError(w, http.StatusNotFound, ErrorCodeNotFound, "team not found")
			return
		}
		c.logger.Error("failed to get team", zap.Error(err))
		c.sendError(w, http.StatusInternalServerError, ErrorCodeInvalidInput, "internal server error")
		return
	}

	response := TeamToDTO(team, members)
	c.sendJSON(w, http.StatusOK, response)
}

func (c *TeamController) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (c *TeamController) sendError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	c.sendJSON(w, status, resp)
}
