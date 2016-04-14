package taskflow

import "errors"

var (
	DoError   = errors.New("Failed to Do it, and Undo")
	UndoError = errors.New("Failed to Do and Undo it")
)

type LineFlow struct {
	Retry uint

	// Private fields
	tasks []Task
}

func NewLineFlow() *LineFlow {
	flow := &LineFlow{Retry: 0}
	flow.tasks = make([]Task, 0)
	return flow
}

func (self *LineFlow) Add(task Task) *LineFlow {
	self.tasks = append(self.tasks, task)
	return self
}

func (self LineFlow) Len() int {
	return len(self.tasks)
}

func (self LineFlow) Cap() int {
	return cap(self.tasks)
}

func (self LineFlow) GetTasks() []Task {
	return self.tasks
}

func (self LineFlow) GetTask(index int) Task {
	if index < 0 || index > len(self.tasks) {
		return nil
	}
	return self.tasks[index]
}

func (self LineFlow) do(task Task) bool {
	var i uint
	for i = 0; i <= self.Retry; i += 1 {
		if ok := task.Do(); ok {
			return true
		}
	}
	return false
}

func (self LineFlow) undo(task Task) bool {
	var i uint
	for i = 0; i <= self.Retry; i += 1 {
		if ok := task.Undo(); ok {
			return true
		}
		i += 1
	}
	return false
}

func (self LineFlow) Execute() error {
	tasks := make([]Task, 0)
	undo := false
	for _, task := range self.tasks {
		if ok := self.do(task); ok {
			tasks = append(tasks, task)
		} else {
			undo = true
			break
		}
	}
	if undo {
		for _len := len(tasks) - 1; _len >= 0; _len -= 1 {
			if _ok := self.undo(tasks[_len]); !_ok {
				return UndoError
			}

		}
		return DoError
	}
	return nil
}
