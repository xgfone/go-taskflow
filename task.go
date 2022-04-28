// Copyright 2020~2022 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package taskflow

import (
	"context"
	"fmt"
	"time"
)

// Task represents a task.
type Task interface {
	fmt.Stringer
	Name() string
	Undo(context.Context) error
	Do(context.Context) error
}

// UndoAll is used to undo all the tasks.
type UndoAll interface {
	UndoAll(context.Context) error
}

// TaskFunc is the function to execute a task.
type TaskFunc func(context.Context) error

// NewTask returns a new Task.
func NewTask(name string, do TaskFunc, undo ...TaskFunc) Task {
	if name == "" {
		panic("the task name must not be empty")
	} else if do == nil {
		panic("the task do function must not be nil")
	}

	undof := undoNothing
	if len(undo) != 0 && undo[0] != nil {
		undof = undo[0]
	}

	return baseTask{name: name, do: do, undo: undof}
}

type baseTask struct {
	name string
	undo TaskFunc
	do   TaskFunc
}

func undoNothing(context.Context) error         { return nil }
func (t baseTask) String() string               { return fmt.Sprintf("Task(name=%s)", t.name) }
func (t baseTask) Name() string                 { return t.name }
func (t baseTask) Undo(c context.Context) error { return t.undo(c) }
func (t baseTask) Do(c context.Context) error   { return t.do(c) }

// LogTaskFunc returns a wrap function to create the LogTask.
func LogTaskFunc(logf func(fmt string, args ...interface{})) func(Task) Task {
	return func(t Task) Task { return NewLogTask(t, logf) }
}

// NewLogTask wraps the task t and prints the log before executing the task.
func NewLogTask(task Task, logf func(fmt string, args ...interface{})) Task {
	return logTask{Task: task, logf: logf}
}

type logTask struct {
	logf func(string, ...interface{})
	Task
}

func (t logTask) Do(c context.Context) error {
	t.logf("doing the task named '%s'", t.Name())
	return t.Task.Do(c)
}

func (t logTask) Undo(c context.Context) error {
	t.logf("undoing the task named '%s'", t.Name())
	return t.Task.Undo(c)
}

// NewRetryTask returns a new Task, which will wrap the task t and retry it
// if the task fails.
func NewRetryTask(t Task, retryNum int, retryInterval time.Duration) Task {
	return retryTask{Task: t, number: retryNum, interval: retryInterval}
}

type retryTask struct {
	interval time.Duration
	number   int
	Task
}

func (t retryTask) Do(c context.Context) error   { return t.retry(c, t.Task.Do) }
func (t retryTask) Undo(c context.Context) error { return t.retry(c, t.Task.Undo) }
func (t retryTask) retry(c context.Context, f TaskFunc) (err error) {
	if err = f(c); err == nil {
		return
	}

	for num := t.number; num > 0; num-- {
		select {
		case <-c.Done():
			return
		default:
		}

		if err = f(c); err == nil {
			return
		}

		if t.interval > 0 {
			timer := time.NewTimer(t.interval)
			select {
			case <-timer.C:
			case <-c.Done():
				timer.Stop()
				return
			}
		}
	}

	return
}
