package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/nessibeliyeltay/task-api/internal/dto"
	"github.com/nessibeliyeltay/task-api/internal/model"
	"github.com/nessibeliyeltay/task-api/internal/repository"
	"github.com/nessibeliyeltay/task-api/internal/repository/mocks"
	"github.com/nessibeliyeltay/task-api/pkg/logger"
)

func setupTestLogger() *logger.Logger {
	config := logger.DefaultConfig()
	config.LogToFile = false
	config.LogToStdout = true
	return logger.New(config)
}

func TestCreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepositoryInterface(ctrl)
	mockLogger := setupTestLogger()
	service := NewTaskService(mockRepo, mockLogger)
	defer service.Shutdown(context.Background())

	service.SetProcessingDelay(100 * time.Millisecond)
	service.SetWorkerCount(1) // Для простоты теста

	ctx := context.Background()
	req := dto.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
	}

	mockRepo.EXPECT().
		CreateTask(gomock.Any()).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			assert.Equal(t, "Test Task", task.Title)
			assert.Equal(t, "Test Description", task.Description)
			assert.Equal(t, model.StatusPending, task.Status)
			return task, nil
		})

	// Два UpdateTask вызова: один для processing, один для completed
	mockRepo.EXPECT().
		UpdateTask(gomock.Any()).
		Times(2).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			return task, nil
		})

	createdTask, err := service.CreateTask(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "Test Task", createdTask.Title)
	assert.Equal(t, "Test Description", createdTask.Description)
	assert.Equal(t, model.StatusPending, createdTask.Status)

	// Ждём завершения обработки
	time.Sleep(300 * time.Millisecond)
}

func TestGetTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepositoryInterface(ctrl)
	mockLogger := setupTestLogger()
	service := NewTaskService(mockRepo, mockLogger)
	service.SetProcessingDelay(100 * time.Millisecond)
	service.SetWorkerCount(5)

	taskID := "1"

	// Тест успешного получения задачи
	t.Run("success", func(t *testing.T) {
		expectedTask := &model.Task{
			ID:          taskID,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.StatusCompleted,
		}

		mockRepo.EXPECT().
			GetTask(taskID).
			Return(expectedTask, nil)

		task, err := service.GetTask(taskID)
		assert.NoError(t, err)
		assert.Equal(t, expectedTask, task)
	})

	// Тест случая, когда задача не найдена
	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetTask(taskID).
			Return(nil, repository.ErrTaskNotFound)

		task, err := service.GetTask(taskID)
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.ErrorIs(t, err, repository.ErrTaskNotFound)
	})

	// Даем время на завершение всех горутин
	time.Sleep(200 * time.Millisecond)
}

func TestListTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepositoryInterface(ctrl)
	mockLogger := setupTestLogger()
	service := NewTaskService(mockRepo, mockLogger)
	service.SetProcessingDelay(100 * time.Millisecond)
	service.SetWorkerCount(5)

	// Тест успешного получения списка задач
	t.Run("success", func(t *testing.T) {
		expectedTasks := []*model.Task{
			{
				ID:          "1",
				Title:       "Task 1",
				Description: "Description 1",
				Status:      model.StatusCompleted,
			},
			{
				ID:          "2",
				Title:       "Task 2",
				Description: "Description 2",
				Status:      model.StatusProcessing,
			},
		}

		mockRepo.EXPECT().
			ListTasks().
			Return(expectedTasks, nil)

		tasks, err := service.ListTasks()
		assert.NoError(t, err)
		assert.Equal(t, expectedTasks, tasks)
	})

	// Тест пустого списка
	t.Run("empty list", func(t *testing.T) {
		mockRepo.EXPECT().
			ListTasks().
			Return([]*model.Task{}, nil)

		tasks, err := service.ListTasks()
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	// Даем время на завершение всех горутин
	time.Sleep(200 * time.Millisecond)
}

func TestDeleteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepositoryInterface(ctrl)
	mockLogger := setupTestLogger()
	service := NewTaskService(mockRepo, mockLogger)
	service.SetProcessingDelay(100 * time.Millisecond)
	service.SetWorkerCount(5)

	taskID := "1"

	// Тест успешного удаления задачи
	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteTask(taskID).
			Return(nil)

		err := service.DeleteTask(taskID)
		assert.NoError(t, err)
	})

	// Тест случая, когда задача не найдена
	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteTask(taskID).
			Return(repository.ErrTaskNotFound)

		err := service.DeleteTask(taskID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrTaskNotFound)
	})

	// Даем время на завершение всех горутин
	time.Sleep(200 * time.Millisecond)
}

func TestWorkerPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepositoryInterface(ctrl)
	mockLogger := setupTestLogger()
	service := NewTaskService(mockRepo, mockLogger)
	service.SetProcessingDelay(100 * time.Millisecond)
	service.SetWorkerCount(2)

	ctx := context.Background()
	req := dto.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
	}

	// Ожидаем создание задачи
	mockRepo.EXPECT().
		CreateTask(gomock.Any()).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			assert.Equal(t, model.StatusPending, task.Status)
			return task, nil
		})

	// Ожидаем обновление статуса на processing
	mockRepo.EXPECT().
		UpdateTask(gomock.Any()).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			assert.Equal(t, model.StatusProcessing, task.Status)
			assert.NotNil(t, task.StartedAt)
			return task, nil
		})

	// Ожидаем обновление статуса на completed
	mockRepo.EXPECT().
		UpdateTask(gomock.Any()).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			assert.Equal(t, model.StatusCompleted, task.Status)
			assert.NotNil(t, task.CompletedAt)
			assert.NotEmpty(t, task.DurationStr)
			return task, nil
		})

	// Ожидаем получение задачи после обработки
	mockRepo.EXPECT().
		GetTask(gomock.Any()).
		DoAndReturn(func(id string) (*model.Task, error) {
			return &model.Task{
				ID:          id,
				Title:       "Test Task",
				Description: "Test Description",
				Status:      model.StatusCompleted,
				DurationStr: "100ms",
			}, nil
		})

	createdTask, err := service.CreateTask(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, createdTask)

	// Даем время на обработку задачи
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что задача была обработана
	retrievedTask, err := service.GetTask(createdTask.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusCompleted, retrievedTask.Status)
	assert.NotEmpty(t, retrievedTask.DurationStr)

	// Даем время на завершение всех горутин
	time.Sleep(200 * time.Millisecond)
}

func TestGracefulShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepositoryInterface(ctrl)
	mockLogger := setupTestLogger()
	service := NewTaskService(mockRepo, mockLogger)
	service.SetProcessingDelay(100 * time.Millisecond)
	service.SetWorkerCount(2)

	ctx := context.Background()
	req := dto.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
	}

	// Ожидаем создание задачи
	mockRepo.EXPECT().
		CreateTask(gomock.Any()).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			assert.Equal(t, model.StatusPending, task.Status)
			return task, nil
		})

	// Ожидаем обновление статуса на processing
	mockRepo.EXPECT().
		UpdateTask(gomock.Any()).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			assert.Equal(t, model.StatusProcessing, task.Status)
			return task, nil
		})

	createdTask, err := service.CreateTask(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, createdTask)

	// Даем время на начало обработки
	time.Sleep(50 * time.Millisecond)

	// Запускаем graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	err = service.Shutdown(shutdownCtx)
	assert.NoError(t, err)

	// Даем время на завершение всех горутин
	time.Sleep(200 * time.Millisecond)
}

func TestShutdownTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepositoryInterface(ctrl)
	mockLogger := setupTestLogger()
	service := NewTaskService(mockRepo, mockLogger)
	service.SetProcessingDelay(0) // убираем искусственную задержку
	service.SetWorkerCount(1)

	ctx := context.Background()
	req := dto.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
	}

	startedCh := make(chan struct{})
	blockProcessingCh := make(chan struct{})

	// Ожидаем создание задачи
	mockRepo.EXPECT().
		CreateTask(gomock.Any()).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			assert.Equal(t, model.StatusPending, task.Status)
			return task, nil
		})

	// Ожидаем обновление статуса на processing и блокируем обработку
	mockRepo.EXPECT().
		UpdateTask(gomock.Any()).
		DoAndReturn(func(task *model.Task) (*model.Task, error) {
			assert.Equal(t, model.StatusProcessing, task.Status)
			// сигнализируем, что обработка началась
			close(startedCh)
			// блокируем обработку
			<-blockProcessingCh
			return task, nil
		})

	// Создаем задачу
	createdTask, err := service.CreateTask(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, createdTask)

	// Ждем, пока обработка начнется
	<-startedCh

	// Запускаем Shutdown с очень коротким таймаутом
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	err = service.Shutdown(shutdownCtx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Разблокируем обработку
	close(blockProcessingCh)

	// Даём немного времени воркеру завершиться (необязательно, если Shutdown ждёт)
	time.Sleep(100 * time.Millisecond)
}
