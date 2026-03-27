package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prastuvwxyz/pilot/internal/config"
	"github.com/prastuvwxyz/pilot/internal/handler/agents"
	"github.com/prastuvwxyz/pilot/internal/handler/auth"
	"github.com/prastuvwxyz/pilot/internal/handler/dashboard"
	"github.com/prastuvwxyz/pilot/internal/handler/health"
	"github.com/prastuvwxyz/pilot/internal/handler/kanban"
	"github.com/prastuvwxyz/pilot/internal/handler/middleware"
	"github.com/prastuvwxyz/pilot/internal/store"
	"github.com/prastuvwxyz/pilot/internal/watcher"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// Logger
	logLevel := slog.LevelInfo
	if cfg.IsDevelopment() {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Stores
	taskStore := store.NewTaskStore(cfg.Paths.EngineeringTasks)
	gitStore := store.NewGitStore(cfg.Paths.PrasMemory)

	// Handlers
	authHandler := auth.NewHandler(cfg)
	dashboardHandler := dashboard.NewHandler(taskStore, cfg.Paths.PrasMemory)
	kanbanHandler := kanban.NewHandler(taskStore, gitStore)
	healthHandler := health.NewHandler()
	agentsHandler := agents.NewHandler(cfg.Paths.PrasMemory)

	// Router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Static("/static", "./web/static")

	// Public routes
	r.GET("/healthz", healthHandler.Check)
	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.HandleLogin)
	r.GET("/logout", authHandler.HandleLogout)
	r.GET("/", func(c *gin.Context) {
		if _, err := c.Cookie("sid"); err == nil {
			c.Redirect(http.StatusFound, "/dashboard")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.JWTAuth(cfg.Pilot.JWTSecret))
	{
		protected.GET("/dashboard", dashboardHandler.Show)
		protected.GET("/kanban", kanbanHandler.ShowKanban)
		protected.POST("/tasks", kanbanHandler.CreateTask)
		protected.GET("/tasks/:id", kanbanHandler.GetTaskDetail)
		protected.POST("/tasks/:id/edit", kanbanHandler.UpdateTask)
		protected.DELETE("/tasks/:id", kanbanHandler.DeleteTask)
		protected.PUT("/tasks/:id/approve", kanbanHandler.ApproveTask)
		protected.PUT("/tasks/:id/status", kanbanHandler.MoveTask)
		protected.GET("/agents", agentsHandler.List)
		protected.GET("/agents/:id", agentsHandler.Detail)
		protected.GET("/agents/:id/files/:filename", agentsHandler.GetFile)
		protected.POST("/agents/:id/files/:filename", agentsHandler.SaveFile)
	}

	// File watcher (goroutine)
	w, err := watcher.New(watcher.Config{
		EngineeringTasksPath: cfg.Paths.EngineeringTasks,
		LeadChannel:          cfg.OpenClaw.EngineeringLeadChannel,
		DevChannel:           cfg.OpenClaw.EngineeringDevChannel,
	})
	if err != nil {
		slog.Error("failed to create watcher", "err", err)
		os.Exit(1)
	}
	go w.Start()
	defer w.Close()

	// HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		slog.Error("server error", "err", err)
	case sig := <-quit:
		slog.Info("shutting down", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	}
	slog.Info("server stopped")
}
