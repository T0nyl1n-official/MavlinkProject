package mavlink

import (
	"fmt"
	"sync"
	"time"

	"MavlinkProject/Server/backend/Shared/Drones"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusPaused   TaskStatus = "paused"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed   TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type TaskPriority int

const (
	PriorityLow    TaskPriority = 1
	PriorityNormal TaskPriority = 5
	PriorityHigh   TaskPriority = 10
	PriorityCritical TaskPriority = 20
)

type Waypoint struct {
	ID          int     `json:"id"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Altitude    float64 `json:"altitude"`
	Speed       float64 `json:"speed"`
	HoldTime    int     `json:"hold_time"`
	AcceptanceRadius int `json:"acceptance_radius"`
	YawAngle    float64 `json:"yaw_angle"`
}

type Mission struct {
	ID          string     `json:"id"`
	Name        string    `json:"name"`
	Waypoints   []Waypoint `json:"waypoints"`
	CurrentIndex int       `json:"current_index"`
	Status      TaskStatus `json:"status"`
	Priority    TaskPriority `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Task struct {
	ID          string     `json:"id"`
	Name        string    `json:"name"`
	Mission     *Mission  `json:"mission,omitempty"`
	DroneID     string    `json:"drone_id"`
	Status      TaskStatus `json:"status"`
	Priority    TaskPriority `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Progress    float64   `json:"progress"`
	Error       string    `json:"error,omitempty"`
	callback    TaskCallback
}

type TaskCallback func(task *Task)

type Scheduler struct {
	mu          sync.RWMutex
	tasks       map[string]*Task
	taskQueue   []*Task
	droneTasks  map[string][]*Task
	
	started     bool
	stopChan    chan bool
	
	taskExecutors map[string]TaskExecutor
}

type TaskExecutor func(task *Task) error

func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make(map[string]*Task),
		taskQueue: make([]*Task, 0),
		droneTasks: make(map[string][]*Task),
		taskExecutors: make(map[string]TaskExecutor),
		stopChan: make(chan bool),
	}
}

func (s *Scheduler) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("任务已存在: %s", task.ID)
	}
	
	task.Status = TaskStatusPending
	task.CreatedAt = time.Now()
	
	s.tasks[task.ID] = task
	s.taskQueue = append(s.taskQueue, task)
	
	s.droneTasks[task.DroneID] = append(s.droneTasks[task.DroneID], task)
	
	s.sortTaskQueue()
	
	return nil
}

func (s *Scheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	if task.Status == TaskStatusRunning {
		return fmt.Errorf("无法删除正在运行的任务: %s", taskID)
	}
	
	delete(s.tasks, taskID)
	
	for i, t := range s.taskQueue {
		if t.ID == taskID {
			s.taskQueue = append(s.taskQueue[:i], s.taskQueue[i+1:]...)
			break
		}
	}
	
	if tasks, ok := s.droneTasks[task.DroneID]; ok {
		for i, t := range tasks {
			if t.ID == taskID {
				s.droneTasks[task.DroneID] = append(tasks[:i], tasks[i+1:]...)
				break
			}
		}
	}
	
	return nil
}

func (s *Scheduler) GetTask(taskID string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	
	return task, nil
}

func (s *Scheduler) GetAllTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	
	return tasks
}

func (s *Scheduler) GetTasksByDrone(droneID string) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tasks, ok := s.droneTasks[droneID]
	if !ok {
		return nil
	}
	
	result := make([]*Task, len(tasks))
	copy(result, tasks)
	
	return result
}

func (s *Scheduler) GetTasksByStatus(status TaskStatus) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var result []*Task
	for _, task := range s.tasks {
		if task.Status == status {
			result = append(result, task)
		}
	}
	
	return result
}

func (s *Scheduler) StartTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	if task.Status != TaskStatusPending && task.Status != TaskStatusPaused {
		return fmt.Errorf("任务当前状态无法启动: %s", task.Status)
	}
	
	now := time.Now()
	task.Status = TaskStatusRunning
	task.StartedAt = &now
	
	if executor, ok := s.taskExecutors[task.Name]; ok {
		go func() {
			err := executor(task)
			s.mu.Lock()
			defer s.mu.Unlock()
			if err != nil {
				task.Status = TaskStatusFailed
				task.Error = err.Error()
			} else {
				task.Status = TaskStatusCompleted
				completedAt := time.Now()
				task.CompletedAt = &completedAt
				task.Progress = 100.0
			}
			if task.callback != nil {
				task.callback(task)
			}
		}()
	}
	
	return nil
}

func (s *Scheduler) PauseTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	if task.Status != TaskStatusRunning {
		return fmt.Errorf("任务当前状态无法暂停: %s", task.Status)
	}
	
	task.Status = TaskStatusPaused
	
	return nil
}

func (s *Scheduler) CancelTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	if task.Status == TaskStatusCompleted || task.Status == TaskStatusCancelled {
		return fmt.Errorf("任务已完成或已取消，无法取消: %s", taskID)
	}
	
	task.Status = TaskStatusCancelled
	now := time.Now()
	task.CompletedAt = &now
	
	return nil
}

func (s *Scheduler) RegisterExecutor(name string, executor TaskExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.taskExecutors[name] = executor
}

func (s *Scheduler) SetTaskCallback(taskID string, callback TaskCallback) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	task.callback = callback
	return nil
}

func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.started {
		return fmt.Errorf("调度器已经在运行")
	}
	
	s.started = true
	go s.processQueue()
	
	return nil
}

func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.started {
		return fmt.Errorf("调度器未在运行")
	}
	
	s.started = false
	close(s.stopChan)
	s.stopChan = make(chan bool)
	
	return nil
}

func (s *Scheduler) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

func (s *Scheduler) processQueue() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.processNextTask()
		}
	}
}

func (s *Scheduler) processNextTask() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if len(s.taskQueue) == 0 {
		return
	}
	
	for _, task := range s.taskQueue {
		if task.Status == TaskStatusPending {
			task.Status = TaskStatusRunning
			now := time.Now()
			task.StartedAt = &now
			
			if executor, ok := s.taskExecutors[task.Name]; ok {
				go func(t *Task) {
					err := executor(t)
					s.mu.Lock()
					defer s.mu.Unlock()
					if err != nil {
						t.Status = TaskStatusFailed
						t.Error = err.Error()
					} else {
						t.Status = TaskStatusCompleted
						completedAt := time.Now()
						t.CompletedAt = &completedAt
						t.Progress = 100.0
					}
					if t.callback != nil {
						t.callback(t)
					}
				}(task)
			}
			break
		}
	}
}

func (s *Scheduler) sortTaskQueue() {
	for i := 0; i < len(s.taskQueue)-1; i++ {
		for j := i + 1; j < len(s.taskQueue); j++ {
			if s.taskQueue[i].Priority < s.taskQueue[j].Priority {
				s.taskQueue[i], s.taskQueue[j] = s.taskQueue[j], s.taskQueue[i]
			}
		}
	}
}

func (s *Scheduler) GetPendingTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	count := 0
	for _, task := range s.tasks {
		if task.Status == TaskStatusPending {
			count++
		}
	}
	
	return count
}

func (s *Scheduler) GetRunningTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	count := 0
	for _, task := range s.tasks {
		if task.Status == TaskStatusRunning {
			count++
		}
	}
	
	return count
}

func (s *Scheduler) CreateMission(id, name string, waypoints []Waypoint, priority TaskPriority) *Mission {
	return &Mission{
		ID:          id,
		Name:        name,
		Waypoints:   waypoints,
		CurrentIndex: 0,
		Status:      TaskStatusPending,
		Priority:   priority,
		CreatedAt:  time.Now(),
	}
}

func (s *Scheduler) CreateTask(id, name, droneID string, mission *Mission, priority TaskPriority) *Task {
	return &Task{
		ID:       id,
		Name:     name,
		Mission:  mission,
		DroneID:  droneID,
		Status:   TaskStatusPending,
		Priority: priority,
		CreatedAt: time.Now(),
		Progress: 0.0,
	}
}
