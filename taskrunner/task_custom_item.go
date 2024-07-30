package taskrunner

import (
	"sync/atomic"
	"time"
)

type customTaskItem interface {
	startSchedule()
	terminate()
	run()
}

type repeatTaskItem struct {
	id                    TaskItemId
	closure               TaskClosure
	repeatingIntervalInMs int64
	ticker                *time.Ticker
	stopCh                chan bool
	isStopped             int32
	taskRunner            *TaskRunner
}

func (m *repeatTaskItem) startSchedule() {
	go m.run()
}

func (m *repeatTaskItem) terminate() {
	if atomic.LoadInt32(&m.isStopped) == 0 {
		m.stopCh <- true
	}
}

func (m *repeatTaskItem) run() {
	m.ticker = time.NewTicker(time.Duration(m.repeatingIntervalInMs) * time.Millisecond)
	defer m.ticker.Stop()
	for {
		m.taskRunner.addTaskInternal(m.id, m.closure)
		select {
		case <-m.ticker.C:
			continue
		case stop := <-m.stopCh:
			if stop {
				atomic.AddInt32(&m.isStopped, 1)
				break
			}
		}
		if atomic.LoadInt32(&m.isStopped) > 0 {
			break
		}
	}
}

type delayedTaskItem struct {
	id              TaskItemId
	closure         TaskClosure
	delayedTimeInMs int64
	stopCh          chan bool
	isStopped       int32
	taskRunner      *TaskRunner
}

func (m *delayedTaskItem) startSchedule() {
	go m.run()
}

func (m *delayedTaskItem) terminate() {
	if atomic.LoadInt32(&m.isStopped) == 0 {
		m.stopCh <- true
	}
}

func (m *delayedTaskItem) run() {
	ticker := time.NewTicker(time.Duration(m.delayedTimeInMs) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&m.isStopped) > 0 {
				break
			}
			_ = m.taskRunner.addTaskInternal(m.id, m.closure)
			m.taskRunner.RemoveTask(m.id)
			atomic.AddInt32(&m.isStopped, 1)
		case stop := <-m.stopCh:
			if stop {
				atomic.AddInt32(&m.isStopped, 1)
			}
		}
		if atomic.LoadInt32(&m.isStopped) > 0 {
			break
		}
	}
}
