package taskrunner

import (
	"container/list"
)

type TaskEventChannel struct {
	SendCh chan TaskItemId
	RecvCh chan TaskItemId
	queue  *list.List
}

func NewTaskEventChannel() *TaskEventChannel {
	eventCh := &TaskEventChannel{
		SendCh: make(chan TaskItemId, 1),
		RecvCh: make(chan TaskItemId, 1),
		queue:  list.New(),
	}
	go eventCh.handleChannel()
	return eventCh
}

func (m *TaskEventChannel) handleChannel() {
	for {
		if front := m.queue.Front(); front == nil {
			if m.SendCh == nil {
				close(m.RecvCh)
				return
			}
			value, ok := <-m.SendCh
			if !ok {
				close(m.RecvCh)
				return
			}
			m.queue.PushBack(value)
		} else {
			select {
			case m.RecvCh <- front.Value.(TaskItemId):
				m.queue.Remove(front)
			case value, ok := <-m.SendCh:
				if ok {
					m.queue.PushBack(value)
				} else {
					m.RecvCh = nil
				}
			}
		}
	}
}
