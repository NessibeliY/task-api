package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nessibeliyeltay/task-api/config"
	"github.com/nessibeliyeltay/task-api/internal/handler"
	"github.com/nessibeliyeltay/task-api/internal/repository"
	"github.com/nessibeliyeltay/task-api/internal/service"
	"github.com/nessibeliyeltay/task-api/pkg/logger"
)

func main() {
	cfg := config.New()

	log := logger.New(cfg.Logger.ToLoggerConfig())

	repo := repository.NewTaskRepository()
	service := service.NewTaskService(repo, log)
	handler := handler.NewTaskHandler(service, log)

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		log.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	})

	handler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("Starting server", zap.String("port", fmt.Sprintf("%d", cfg.Server.Port)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down task service")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := service.Shutdown(ctx); err != nil {
		log.Error("Error shutting down task service", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Error shutting down server", err)
	}

	log.Info("Server stopped")
}
