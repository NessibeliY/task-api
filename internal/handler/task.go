package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nessibeliyeltay/task-api/internal/dto"
	"github.com/nessibeliyeltay/task-api/internal/repository"
	"github.com/nessibeliyeltay/task-api/internal/service"
	"github.com/nessibeliyeltay/task-api/pkg/logger"
)

type TaskHandler struct {
	service service.TaskServiceInterface
	logger  *logger.Logger
}

func NewTaskHandler(service service.TaskServiceInterface, logger *logger.Logger) *TaskHandler {
	return &TaskHandler{
		service: service,
		logger:  logger,
	}
}

func (h *TaskHandler) RegisterRoutes(router *gin.Engine) {
	tasks := router.Group("/api/v1/tasks")
	{
		tasks.POST("", h.CreateTask)
		tasks.GET("", h.ListTasks)
		tasks.GET("/:id", h.GetTask)
		tasks.DELETE("/:id", h.DeleteTask)
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	task, err := h.service.CreateTask(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create task", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, dto.NewTaskResponse(task))
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	tasks, err := h.service.ListTasks()
	if err != nil {
		h.logger.Error("Failed to list tasks", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
		return
	}

	response := make([]*dto.TaskResponse, len(tasks))
	for i, task := range tasks {
		response[i] = dto.NewTaskResponse(task)
	}

	c.JSON(http.StatusOK, response)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Info("Invalid request: missing task ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	task, err := h.service.GetTask(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTaskID):
			h.logger.Info("Invalid task ID format", zap.String("task_id", id))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		case errors.Is(err, repository.ErrTaskNotFound):
			h.logger.Info("Task not found", zap.String("task_id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		default:
			h.logger.Error("Failed to get task", err, zap.String("task_id", id))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.NewTaskResponse(task))
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Info("Invalid request: missing task ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	if err := h.service.DeleteTask(id); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTaskID):
			h.logger.Info("Invalid task ID format", zap.String("task_id", id))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		case errors.Is(err, repository.ErrTaskNotFound):
			h.logger.Info("Task not found", zap.String("task_id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		default:
			h.logger.Error("Failed to delete task", err, zap.String("task_id", id))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
