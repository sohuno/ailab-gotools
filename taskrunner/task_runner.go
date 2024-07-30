package taskrunner

import (
	"sync"
	"sync/atomic"

	"github.com/panjf2000/ants"
	"k8s.io/klog/v2"
)

// https://pkg.go.dev/github.com/panjf2000/ants@v1.2.0

type TaskRunner struct {
	name          string
	taskMap       map[TaskItemId]*taskItem
	customTaskMap map[TaskItemId]customTaskItem

	taskClosureNextId int64
	pool              *ants.Pool
	eventCh           *TaskEventChannel
	mutex             sync.Mutex
}

func NewTaskRunner(name string, size int) *TaskRunner {
	pool, _ := ants.NewPool(size)
	return &TaskRunner{
		name:              name,
		taskMap:           make(map[TaskItemId]*taskItem, 0),
		customTaskMap:     make(map[TaskItemId]customTaskItem, 0),
		taskClosureNextId: 0,
		pool:              pool,
		eventCh:           NewTaskEventChannel(),
	}
}

func (m *TaskRunner) Startup() {
	go m.scheduleOneTask()
}

func (m *TaskRunner) Shutdown() {
	for _, task := range m.customTaskMap {
		task.terminate()
	}
}

func (m *TaskRunner) AddTask(closure TaskClosure) TaskItemId {
	id := m.getUniqueTaskId()
	return m.addTaskInternal(id, closure)
}

func (m *TaskRunner) addTaskInternal(id TaskItemId, closure TaskClosure) TaskItemId {
	m.mutex.Lock()
	if _, found := m.taskMap[id]; found {
		m.mutex.Unlock()
		return id
	}
	m.taskMap[id] = &taskItem{
		id:         id,
		closure:    closure,
		isRunning:  false,
		taskRunner: m,
	}
	m.mutex.Unlock()
	m.eventCh.SendCh <- id
	return 0
}

func (m *TaskRunner) AddRepeatingTask(closure TaskClosure, repeatingIntervalInMs int64) TaskItemId {
	id := m.getUniqueTaskId()
	repeatingTask := &repeatTaskItem{
		id:                    id,
		closure:               closure,
		repeatingIntervalInMs: repeatingIntervalInMs,
		stopCh:                make(chan bool, 1),
		taskRunner:            m,
	}
	m.mutex.Lock()
	m.customTaskMap[id] = repeatingTask
	m.mutex.Unlock()
	repeatingTask.startSchedule()
	return id
}

func (m *TaskRunner) AddDelayedTask(closure TaskClosure, delayedTimeInMs int64) TaskItemId {
	id := m.getUniqueTaskId()
	delayedTask := &delayedTaskItem{
		id:              id,
		closure:         closure,
		delayedTimeInMs: delayedTimeInMs,
		stopCh:          make(chan bool, 1),
		taskRunner:      m,
	}
	m.mutex.Lock()
	m.customTaskMap[id] = delayedTask
	m.mutex.Unlock()
	delayedTask.startSchedule()
	return id
}

func (m *TaskRunner) RemoveTask(id TaskItemId) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if task, found := m.customTaskMap[id]; found {
		go task.terminate()
		delete(m.customTaskMap, id)
	}
}

func (m *TaskRunner) getUniqueTaskId() TaskItemId {
	return TaskItemId(atomic.AddInt64(&m.taskClosureNextId, 1))
}

func (m *TaskRunner) getTask() *taskItem {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.taskMap) == 0 {
		return nil
	}
	for _, task := range m.taskMap {
		if !task.isRunning {
			task.isRunning = true
			return task
		}
	}
	return nil
}

func (m *TaskRunner) removeTask(taskId TaskItemId) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.taskMap, taskId)
}

func (m *TaskRunner) scheduleOneTask() {
	for {
		select {
		case _, ok := <-m.eventCh.RecvCh:
			if ok {
				task := m.getTask()
				if task != nil {
					m.pool.Submit(task.run)
				} else {
					klog.Errorf("Task in TaskRunner(%s) is nil", m.name)
				}
			} else {
				klog.Errorf("RecvCh in TaskRunner(%s) is closed", m.name)
			}
		}
	}
}
