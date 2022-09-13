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
)

var _ Flow = &LineFlow{}

// LineFlow is used to execute the tasks in turn.
type LineFlow struct {
	name  string
	tasks []Task
	errf  func(err error)

	ctxs  map[string]interface{}
	index int

	undoAll  bool
	undofail bool
}

// NewLineFlow returns a new line flow, which executes the task in turn.
func NewLineFlow(name string) *LineFlow { return &LineFlow{name: name, index: -1} }

func (f *LineFlow) String() string { return fmt.Sprintf("LineFlow(name=%s)", f.name) }

// Name returns the name of line flow.
func (f *LineFlow) Name() string { return f.name }

// Tasks returns all the tasks.
func (f *LineFlow) Tasks() []Task { return f.tasks }

// DoneTasks returns all the done tasks.
func (f *LineFlow) DoneTasks() []Task { return f.tasks[:f.index] }

// SetUndoAllTasks sets whether to undo all the tasks or not
// if the task has implemented the interface UndoAll.
func (f *LineFlow) SetUndoAllTasks(b bool) { f.undoAll = b }

// SetUndoFailedTask sets whether to undo the failed task or not.
func (f *LineFlow) SetUndoFailedTask(undo bool) { f.undofail = undo }

// SetCtx adds the key-value context to allow that the subsequent tasks access
// it, which will override it if the key has existed.
func (f *LineFlow) SetCtx(key string, value interface{}) {
	if f.ctxs == nil {
		f.ctxs = map[string]interface{}{key: value}
	} else {
		f.ctxs[key] = value
	}
}

// GetCtx returns the value of the context named key.
//
// Return nil if the key does not exist.
func (f *LineFlow) GetCtx(key string) interface{} { return f.ctxs[key] }

// AddTasks adds the task or flow into the line flow.
func (f *LineFlow) AddTasks(tasks ...Task) {
	if f.index > -1 {
		panic("LineFlow: the tasks have been executed")
	}
	f.tasks = append(f.tasks, tasks...)
}

// AddTask adds the task with the task name and do/undo function.
func (f *LineFlow) AddTask(name string, do TaskFunc, undo ...TaskFunc) {
	f.AddTasks(NewTask(name, do, undo...))
}

// SetErrorHandler sets the handler to handle it if there is an error.
func (f *LineFlow) SetErrorHandler(handle func(err error)) {
	f.errf = handle
}

// Do does the tasks, which undoes them if a certain task fails.
func (f *LineFlow) Do(c context.Context) (err error) {
	if err = f.do(c); err != nil && f.errf != nil {
		f.errf(err)
	}
	return
}

func (f *LineFlow) do(c context.Context) (err error) {
	f.index = 0
	for i, end := 0, len(f.tasks); i < end; i++ {
		task := f.tasks[i]
		if err = task.Do(c); err != nil {
			if f.undofail {
				f.index++
			}

			errs := f.undo(c)
			if len(errs) == 0 {
				return NewFlowError(f.name, NewTaskError(task.Name(), err, nil))
			}

			tname := task.Name()
			tes := make(TaskErrors, 1, len(errs)+1)
			tes[0] = NewTaskError(tname, err, nil)
			for _, e := range errs {
				if ename := e.Name(); ename == tname {
					tes[0] = NewTaskError(tname, err, e.UndoErr())
				} else {
					tes = append(tes, e)
				}
			}

			return NewFlowError(f.name, tes...)
		}

		f.index++
	}

	return
}

// Undo undoes the done tasks.
func (f *LineFlow) Undo(c context.Context) error {
	if tes := f.undo(c); len(tes) != 0 {
		return NewFlowError(f.name, tes...)
	}
	return nil
}

func (f *LineFlow) undo(c context.Context) (tes TaskErrors) {
	for index := f.index - 1; index >= 0; index-- {
		task := f.tasks[index]

		var err error
		if ta, ok := task.(UndoAll); ok && f.undoAll {
			err = ta.UndoAll(c)
		} else {
			err = task.Undo(c)
		}

		if err != nil {
			tes.Append(task.Name(), nil, err)
		}
	}
	return
}
