package service

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

	"go.uber.org/zap"

	"github.com/nessibeliyeltay/task-api/internal/dto"
	"github.com/nessibeliyeltay/task-api/internal/model"
	"github.com/nessibeliyeltay/task-api/internal/repository"
	"github.com/nessibeliyeltay/task-api/pkg/logger"
)

var ErrInvalidTaskID = errors.New("invalid task ID format")

type TaskServiceInterface interface {
	CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*model.Task, error)
	ListTasks() ([]*model.Task, error)
	GetTask(id string) (*model.Task, error)
	DeleteTask(id string) error
	Shutdown(ctx context.Context) error
}

type TaskService struct {
	repo            repository.TaskRepositoryInterface
	logger          *logger.Logger
	processingDelay time.Duration
	workerCount     int
	taskQueue       chan *model.Task
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	shutdownChan    chan struct{}
}

func NewTaskService(repo repository.TaskRepositoryInterface, logger *logger.Logger) *TaskService {
	ctx, cancel := context.WithCancel(context.Background())
	service := &TaskService{
		repo:            repo,
		logger:          logger,
		processingDelay: 2 * time.Minute, // Default processing time
		workerCount:     5,               // Default number of workers
		taskQueue:       make(chan *model.Task, 100),
		ctx:             ctx,
		cancel:          cancel,
		shutdownChan:    make(chan struct{}),
	}

	service.startWorkers(ctx)

	return service
}

// SetProcessingDelay sets the processing delay for testing purposes
func (s *TaskService) SetProcessingDelay(delay time.Duration) {
	s.processingDelay = delay
}

// SetWorkerCount sets the number of concurrent workers
func (s *TaskService) SetWorkerCount(count int) {
	if count < 1 {
		count = 1
	}
	s.workerCount = count
}

// startWorkers starts the worker pool
func (s *TaskService) startWorkers(ctx context.Context) {
	for i := 0; i < s.workerCount; i++ {
		workerID := i
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.logger.Info("Worker started", zap.Int("worker_id", workerID))

			for {
				select {
				case <-ctx.Done():
					s.logger.Info("Worker stopping due to context cancellation", zap.Int("worker_id", workerID))
					return
				case <-s.shutdownChan:
					s.logger.Info("Worker stopping due to shutdown signal", zap.Int("worker_id", workerID))
					return
				default:
					select {
					case <-ctx.Done():
						s.logger.Info("Worker stopping due to context cancellation", zap.Int("worker_id", workerID))
						return
					case <-s.shutdownChan:
						s.logger.Info("Worker stopping due to shutdown signal", zap.Int("worker_id", workerID))
						return
					case task := <-s.taskQueue:
						if task == nil {
							continue
						}

						task.UpdateStatus(model.StatusProcessing)
						if _, err := s.repo.UpdateTask(task); err != nil {
							s.logger.Error("Failed to update task status", err,
								zap.String("task_id", task.ID))
							continue
						}

						s.logger.Info("Task processing started",
							zap.String("task_id", task.ID),
							zap.Time("started_at", *task.StartedAt),
							zap.String("duration", task.DurationStr))

						// Simulate processing with context
						select {
						case <-ctx.Done():
							s.logger.Info("Task processing cancelled", zap.String("task_id", task.ID))
							return
						case <-s.shutdownChan:
							s.logger.Info("Task processing cancelled due to shutdown", zap.String("task_id", task.ID))
							return
						case <-time.After(s.processingDelay):
							// Continue processing
						}

						// Complete task
						task.UpdateStatus(model.StatusCompleted)
						task.Result = "Task completed successfully"
						if _, err := s.repo.UpdateTask(task); err != nil {
							s.logger.Error("Failed to update task status", err,
								zap.String("task_id", task.ID))
							continue
						}

						s.logger.Info("Task completed",
							zap.String("task_id", task.ID),
							zap.String("result", task.Result),
							zap.String("duration", task.DurationStr))
					}
				}
			}
		}()
	}
}

func (s *TaskService) CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*model.Task, error) {
	task := model.NewTask(req.Title, req.Description)

	task, err := s.repo.CreateTask(task)
	if err != nil {
		return nil, errors.Wrap(err, "create task")
	}

	s.logger.Info("Task created",
		zap.String("task_id", task.ID),
		zap.String("status", string(task.Status)))

	select {
	case s.taskQueue <- task:
		s.logger.Info("Task queued for processing",
			zap.String("task_id", task.ID))
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "task queue is full")
	default:
		s.logger.Warn("Task queue is full, task will be processed when space is available",
			zap.String("task_id", task.ID))
		s.taskQueue <- task
	}

	return task, nil
}

func (s *TaskService) ListTasks() ([]*model.Task, error) {
	return s.repo.ListTasks() //nolint:wrapcheck
}

func (s *TaskService) GetTask(id string) (*model.Task, error) {
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return nil, ErrInvalidTaskID
	}

	task, err := s.repo.GetTask(id)
	if err != nil {
		return nil, errors.Wrap(err, "get task")
	}

	s.logger.Info("Task retrieved",
		zap.String("task_id", task.ID),
		zap.String("status", string(task.Status)))

	return task, nil
}

func (s *TaskService) DeleteTask(id string) error {
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return ErrInvalidTaskID
	}

	err := s.repo.DeleteTask(id)
	if err != nil {
		return errors.Wrap(err, "delete task")
	}

	s.logger.Info("Task deleted",
		zap.String("task_id", id))

	return nil
}

// Shutdown gracefully shuts down the task service
func (s *TaskService) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down task service")
	close(s.shutdownChan)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "ctx done")
	case <-done:
		return nil
	}
}
