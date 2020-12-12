// Copyright 2020 xgfone
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
	ctxs  map[string]interface{}
	name  string
	tasks []Task
	index int
	aundo func()
	bundo func()
	ado   func()
	bdo   func()

	undoAll bool
	fail    bool
}

// NewLineFlow returns a new line flow, which executes the task in turn.
func NewLineFlow(name string) *LineFlow { return &LineFlow{name: name, index: -1} }

func (f *LineFlow) String() string {
	return fmt.Sprintf("LineFlow(name=%s)", f.name)
}

// Name returns the name of line flow.
func (f *LineFlow) Name() string { return f.name }

// Tasks returns all the tasks.
func (f *LineFlow) Tasks() []Task { return f.tasks }

// DoneTasks returns all the done tasks.
func (f *LineFlow) DoneTasks() []Task { return f.tasks[:f.index] }

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
func (f *LineFlow) GetCtx(key string) (value interface{}) { return f.ctxs[key] }

// Add adds the task or flow into the line flow.
func (f *LineFlow) Add(tasks ...Task) Flow {
	if f.index > -1 {
		panic("LineFlow: the tasks have been executed")
	}
	f.tasks = append(f.tasks, tasks...)
	return f
}

// SetUndoAllTasks sets whether to undo all the tasks or not
// if the task has implemented the interface UndoAll.
func (f *LineFlow) SetUndoAllTasks(b bool) *LineFlow { f.undoAll = b; return f }

// UndoFailedTask sets whether to undo the failed task or not.
func (f *LineFlow) UndoFailedTask(undo bool) *LineFlow { f.fail = undo; return f }

// BeforeDo executes the do function before executing the Do method.
func (f *LineFlow) BeforeDo(do func()) *LineFlow { f.bdo = do; return f }

// BeforeUndo executes the undo function before executing the Undo method.
func (f *LineFlow) BeforeUndo(undo func()) *LineFlow { f.bundo = undo; return f }

// AfterDo executes the do function after executing the Do method.
func (f *LineFlow) AfterDo(do func()) *LineFlow { f.ado = do; return f }

// AfterUndo executes the undo function after executing the Undo method.
func (f *LineFlow) AfterUndo(undo func()) *LineFlow { f.aundo = undo; return f }

// Do does the tasks, which undoes them if a certain task fails.
func (f *LineFlow) Do(c context.Context) (err error) {
	if f.bdo != nil {
		f.bdo()
	}
	if f.ado != nil {
		defer f.ado()
	}

	f.index = 0
	for i, end := 0, len(f.tasks); i < end; i++ {
		task := f.tasks[i]
		if err = task.Do(c); err != nil {
			if f.fail {
				f.index++
			}

			errs := f.undo(c)
			if len(errs) == 0 {
				return NewFlowError(f.name, NewTaskError(task.Name(), err, nil))
			}

			tname := task.Name()
			tes := make(TaskErrors, 0, len(errs)+1)
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
	if f.bundo != nil {
		f.bundo()
	}
	if f.aundo != nil {
		defer f.aundo()
	}

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
