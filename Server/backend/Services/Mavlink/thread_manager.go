package MavlinkService

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"MavlinkProject/Server/backend/Shared/Drones"
)

type WorkerPool struct {
	workers       int
	taskQueue     chan func()
	stopChan      chan struct{}
	wg            sync.WaitGroup
	mu            sync.RWMutex
	activeWorkers int32
	stats         WorkerStats
}

type WorkerStats struct {
	TasksCompleted int64
	TasksFailed    int64
	TasksPending   int64
}

type ThreadManager struct {
	mu          sync.RWMutex
	workerPools map[string]*WorkerPool
	globalPool  *WorkerPool
	ctx         context.Context
	cancel      context.CancelFunc

	maxWorkers int
	autoScale  bool

	droneWorkers map[string]context.CancelFunc
}

func NewThreadManager(maxWorkers int) *ThreadManager {
	if maxWorkers == 0 {
		maxWorkers = runtime.NumCPU() * 2
	}

	ctx, cancel := context.WithCancel(context.Background())

	tm := &ThreadManager{
		ctx:          ctx,
		cancel:       cancel,
		workerPools:  make(map[string]*WorkerPool),
		droneWorkers: make(map[string]context.CancelFunc),
		maxWorkers:   maxWorkers,
		autoScale:    true,
	}

	tm.globalPool = tm.createWorkerPool("global", runtime.NumCPU())

	return tm
}

func (tm *ThreadManager) createWorkerPool(name string, workers int) *WorkerPool {
	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan func(), 1000),
		stopChan:  make(chan struct{}),
	}

	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.stopChan:
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			atomic.AddInt64(&p.stats.TasksPending, -1)
			task()
			atomic.AddInt64(&p.stats.TasksCompleted, 1)
		}
	}
}

func (p *WorkerPool) Submit(task func()) bool {
	select {
	case p.taskQueue <- task:
		atomic.AddInt64(&p.stats.TasksPending, 1)
		return true
	default:
		return false
	}
}

func (p *WorkerPool) SubmitWithTimeout(task func(), timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case p.taskQueue <- task:
		atomic.AddInt64(&p.stats.TasksPending, 1)
		return true
	case <-ctx.Done():
		atomic.AddInt64(&p.stats.TasksFailed, 1)
		return false
	}
}

func (p *WorkerPool) Stop() {
	close(p.stopChan)
	p.wg.Wait()
}

func (p *WorkerPool) GetStats() WorkerStats {
	return WorkerStats{
		TasksCompleted: atomic.LoadInt64(&p.stats.TasksCompleted),
		TasksFailed:    atomic.LoadInt64(&p.stats.TasksFailed),
		TasksPending:   atomic.LoadInt64(&p.stats.TasksPending),
	}
}

func (tm *ThreadManager) SubmitGlobalTask(task func()) bool {
	return tm.globalPool.Submit(task)
}

func (tm *ThreadManager) SubmitDroneTask(droneID string, task func()) bool {
	tm.mu.RLock()
	pool, exists := tm.workerPools[droneID]
	tm.mu.RUnlock()

	if !exists {
		tm.mu.Lock()
		pool = tm.createWorkerPool(droneID, runtime.NumCPU())
		tm.workerPools[droneID] = pool
		tm.mu.Unlock()
	}

	return pool.Submit(task)
}

func (tm *ThreadManager) CreateDroneWorker(droneID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.droneWorkers[droneID]; exists {
		return fmt.Errorf("无人机工作线程已存在: %s", droneID)
	}

	droneCtx, cancel := context.WithCancel(tm.ctx)
	tm.droneWorkers[droneID] = cancel

	pool := tm.createWorkerPool(droneID, runtime.NumCPU())
	tm.workerPools[droneID] = pool

	go tm.runDroneWorker(droneID, droneCtx)

	return nil
}

func (tm *ThreadManager) runDroneWorker(droneID string, ctx context.Context) {
	tm.mu.RLock()
	pool := tm.workerPools[droneID]
	tm.mu.RUnlock()

	for {
		select {
		case <-ctx.Done():
			if pool != nil {
				pool.Stop()
			}
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func (tm *ThreadManager) StopDroneWorker(droneID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	cancel, exists := tm.droneWorkers[droneID]
	if !exists {
		return fmt.Errorf("无人机工作线程不存在: %s", droneID)
	}

	cancel()
	delete(tm.droneWorkers, droneID)

	if pool, exists := tm.workerPools[droneID]; exists {
		pool.Stop()
		delete(tm.workerPools, droneID)
	}

	return nil
}

func (tm *ThreadManager) Stop() {
	tm.cancel()

	tm.mu.RLock()
	for _, cancel := range tm.droneWorkers {
		cancel()
	}
	tm.mu.RUnlock()

	time.Sleep(100 * time.Millisecond)

	tm.mu.Lock()
	for _, pool := range tm.workerPools {
		pool.Stop()
	}
	tm.workerPools = make(map[string]*WorkerPool)
	tm.droneWorkers = make(map[string]context.CancelFunc)
	tm.mu.Unlock()

	if tm.globalPool != nil {
		tm.globalPool.Stop()
	}
}

func (tm *ThreadManager) GetWorkerCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.workerPools)
}

func (tm *ThreadManager) GetGlobalStats() WorkerStats {
	if tm.globalPool != nil {
		return tm.globalPool.GetStats()
	}
	return WorkerStats{}
}

func (tm *ThreadManager) GetDroneStats(droneID string) (WorkerStats, error) {
	tm.mu.RLock()
	pool, exists := tm.workerPools[droneID]
	tm.mu.RUnlock()

	if !exists {
		return WorkerStats{}, fmt.Errorf("无人机工作线程不存在: %s", droneID)
	}

	return pool.GetStats(), nil
}

type ConcurrentDroneManager struct {
	mu        sync.RWMutex
	drones    map[string]*Drones.Drone
	threadMgr *ThreadManager

	channelDrones map[chan *Drones.Drone]bool

	stopChan chan struct{}
	wg       sync.WaitGroup

	eventHandlers map[string]EventHandler
}

type EventHandler func(drone *Drones.Drone, eventType string)

func NewConcurrentDroneManager(maxWorkers int) *ConcurrentDroneManager {
	cdm := &ConcurrentDroneManager{
		drones:        make(map[string]*Drones.Drone),
		threadMgr:     NewThreadManager(maxWorkers),
		channelDrones: make(map[chan *Drones.Drone]bool),
		stopChan:      make(chan struct{}),
		eventHandlers: make(map[string]EventHandler),
	}

	cdm.startEventProcessor()

	return cdm
}

func (cdm *ConcurrentDroneManager) AddDrone(drone *Drones.Drone) error {
	cdm.mu.Lock()
	defer cdm.mu.Unlock()

	if _, exists := cdm.drones[drone.GetID()]; exists {
		return fmt.Errorf("无人机已存在: %s", drone.GetID())
	}

	cdm.drones[drone.GetID()] = drone

	if err := cdm.threadMgr.CreateDroneWorker(drone.GetID()); err != nil {
		return err
	}

	drone.RegisterCallback(drone.GetID(), func(event Drones.DroneEvent) {
		cdm.handleDroneEvent(drone, event)
	})

	return nil
}

func (cdm *ConcurrentDroneManager) RemoveDrone(droneID string) error {
	cdm.mu.Lock()
	defer cdm.mu.Unlock()

	if _, exists := cdm.drones[droneID]; !exists {
		return fmt.Errorf("无人机不存在: %s", droneID)
	}

	delete(cdm.drones, droneID)

	return cdm.threadMgr.StopDroneWorker(droneID)
}

func (cdm *ConcurrentDroneManager) GetDrone(droneID string) (*Drones.Drone, error) {
	cdm.mu.RLock()
	defer cdm.mu.RUnlock()

	drone, exists := cdm.drones[droneID]
	if !exists {
		return nil, fmt.Errorf("无人机不存在: %s", droneID)
	}

	return drone, nil
}

func (cdm *ConcurrentDroneManager) GetAllDrones() []*Drones.Drone {
	cdm.mu.RLock()
	defer cdm.mu.RUnlock()

	drones := make([]*Drones.Drone, 0, len(cdm.drones))
	for _, drone := range cdm.drones {
		drones = append(drones, drone)
	}

	return drones
}

func (cdm *ConcurrentDroneManager) SubmitDroneTask(droneID string, task func()) bool {
	return cdm.threadMgr.SubmitDroneTask(droneID, task)
}

func (cdm *ConcurrentDroneManager) SubmitGlobalTask(task func()) bool {
	return cdm.threadMgr.SubmitGlobalTask(task)
}

func (cdm *ConcurrentDroneManager) RegisterEventHandler(id string, handler EventHandler) {
	cdm.mu.Lock()
	defer cdm.mu.Unlock()

	cdm.eventHandlers[id] = handler
}

func (cdm *ConcurrentDroneManager) UnregisterEventHandler(id string) {
	cdm.mu.Lock()
	defer cdm.mu.Unlock()

	delete(cdm.eventHandlers, id)
}

func (cdm *ConcurrentDroneManager) handleDroneEvent(drone *Drones.Drone, event Drones.DroneEvent) {
	cdm.mu.RLock()
	handlers := make([]EventHandler, 0, len(cdm.eventHandlers))
	for _, handler := range cdm.eventHandlers {
		handlers = append(handlers, handler)
	}
	cdm.mu.RUnlock()

	for _, handler := range handlers {
		handler(drone, event.Type)
	}
}

func (cdm *ConcurrentDroneManager) startEventProcessor() {
	cdm.wg.Add(1)
	go func() {
		defer cdm.wg.Done()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-cdm.stopChan:
				return
			case <-ticker.C:
				cdm.checkDroneHealth()
			}
		}
	}()
}

func (cdm *ConcurrentDroneManager) checkDroneHealth() {
	cdm.mu.RLock()
	drones := make([]*Drones.Drone, 0, len(cdm.drones))
	for _, drone := range cdm.drones {
		drones = append(drones, drone)
	}
	cdm.mu.RUnlock()

	for _, drone := range drones {
		droneID := drone.GetID()
		cdm.threadMgr.SubmitDroneTask(droneID, func() {
			if drone.IsTimedOut(10 * time.Second) {
				drone.SetConnected(false)
				drone.SetStatus(Drones.StatusDisconnected)
			}
		})
	}
}

func (cdm *ConcurrentDroneManager) Stop() {
	close(cdm.stopChan)
	cdm.wg.Wait()
	cdm.threadMgr.Stop()
}

func (cdm *ConcurrentDroneManager) Subscribe() chan *Drones.Drone {
	ch := make(chan *Drones.Drone, 100)

	cdm.mu.Lock()
	cdm.channelDrones[ch] = true
	cdm.mu.Unlock()

	return ch
}

func (cdm *ConcurrentDroneManager) Unsubscribe(ch chan *Drones.Drone) {
	cdm.mu.Lock()
	delete(cdm.channelDrones, ch)
	cdm.mu.Unlock()
	close(ch)
}

func (cdm *ConcurrentDroneManager) BroadcastDrone(drone *Drones.Drone) {
	cdm.mu.RLock()
	channels := make([]chan *Drones.Drone, 0, len(cdm.channelDrones))
	for ch := range cdm.channelDrones {
		channels = append(channels, ch)
	}
	cdm.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- drone:
		default:
		}
	}
}
