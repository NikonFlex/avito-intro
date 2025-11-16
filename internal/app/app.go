package app

import (
	"context"
	"net/http"

	"avito-intro/config"
	"avito-intro/internal/controller"
	"avito-intro/internal/repository"
	"avito-intro/internal/usecase"

	"go.uber.org/zap"
)

type App struct {
	server *http.Server
	logger *zap.Logger
	config *config.Config
}

func New(cfg *config.Config, logger *zap.Logger) *App {
	repo := repository.NewMemoryRepository(logger)

	teamUC := usecase.NewTeamUsecase(repo, repo, logger)
	userUC := usecase.NewUserUsecase(repo, logger)
	prUC := usecase.NewPullRequestUsecase(repo, repo, logger)

	teamController := controller.NewTeamController(teamUC, logger)
	userController := controller.NewUserController(userUC, prUC, logger)
	prController := controller.NewPullRequestController(prUC, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /team/add", teamController.AddTeam)
	mux.HandleFunc("GET /team/get", teamController.GetTeam)

	mux.HandleFunc("POST /users/setIsActive", userController.SetIsActive)
	mux.HandleFunc("GET /users/getReview", userController.GetReview)

	mux.HandleFunc("POST /pullRequest/create", prController.CreatePR)
	mux.HandleFunc("POST /pullRequest/merge", prController.MergePR)
	mux.HandleFunc("POST /pullRequest/reassign", prController.ReassignReviewer)

	server := &http.Server{
		Addr:         cfg.ServerAddr(),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &App{
		server: server,
		logger: logger,
		config: cfg,
	}
}

func (a *App) Run() error {
	a.logger.Info("Server starting", zap.String("addr", a.server.Addr))
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Server shutting down...")
	return a.server.Shutdown(ctx)
}
