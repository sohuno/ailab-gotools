package taskrunner

import (
	"k8s.io/klog/v2"
)

type taskItem struct {
	id         TaskItemId
	closure    TaskClosure
	isRunning  bool
	taskRunner *TaskRunner
}

func (m *taskItem) run() {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("PanicHappenedInTaskItem r:%+v", r)
			m.taskRunner.removeTask(m.id)
		}
	}()
	m.closure.Run()
	m.taskRunner.removeTask(m.id)
}
