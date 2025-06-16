package repository

import (
	"errors"
	"strconv"
	"sync"

	"github.com/nessibeliyeltay/task-api/internal/model"
)

//go:generate mockgen -source=task.go -destination=mocks/task_mock.go -package=mocks

var ErrTaskNotFound = errors.New("task not found")

type TaskRepositoryInterface interface {
	CreateTask(task *model.Task) (*model.Task, error)
	ListTasks() ([]*model.Task, error)
	GetTask(id string) (*model.Task, error)
	UpdateTask(task *model.Task) (*model.Task, error)
	DeleteTask(id string) error
}

type InMemoryTaskRepository struct {
	tasks  map[string]*model.Task
	mu     sync.RWMutex
	nextID int64
}

func NewTaskRepository() TaskRepositoryInterface {
	return &InMemoryTaskRepository{
		tasks:  make(map[string]*model.Task),
		nextID: 1,
	}
}

func (r *InMemoryTaskRepository) CreateTask(task *model.Task) (*model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = r.getNextID()
	r.tasks[task.ID] = task
	return task, nil
}

func (r *InMemoryTaskRepository) getNextID() string {
	id := r.nextID
	r.nextID++
	return strconv.FormatInt(id, 10)
}

func (r *InMemoryTaskRepository) ListTasks() ([]*model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*model.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *InMemoryTaskRepository) GetTask(id string) (*model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (r *InMemoryTaskRepository) UpdateTask(task *model.Task) (*model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return nil, ErrTaskNotFound
	}

	r.tasks[task.ID] = task
	return task, nil
}

func (r *InMemoryTaskRepository) DeleteTask(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return ErrTaskNotFound
	}

	delete(r.tasks, id)
	return nil
}
