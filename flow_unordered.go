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

// UnorderedFlow is used to execute the tasks in unordered turn.
type UnorderedFlow struct {
	name  string
	tasks []Task
	aundo func()
	bundo func()
	ado   func()
	bdo   func()

	undoAll    bool
	concurrent bool
}

// NewUnorderedFlow returns a new UnorderedFlow.
func NewUnorderedFlow(name string) *UnorderedFlow {
	return &UnorderedFlow{name: name}
}

func (f *UnorderedFlow) String() string {
	return fmt.Sprintf("UnorderedFlow(name=%s)", f.name)
}

// Name returns the name of line flow.
func (f *UnorderedFlow) Name() string { return f.name }

// Tasks returns all the tasks.
func (f *UnorderedFlow) Tasks() []Task { return f.tasks }

// Add adds the task or flow into the line flow.
func (f *UnorderedFlow) Add(tasks ...Task) {
	f.tasks = append(f.tasks, tasks...)
}

// SetUndoAllTasks sets whether to undo all the tasks or not
// if the task has implemented the interface UndoAll.
func (f *UnorderedFlow) SetUndoAllTasks(b bool) *UnorderedFlow { f.undoAll = b; return f }

// Concurrent sets whether to do the tasks concurrently.
func (f *UnorderedFlow) Concurrent(b bool) *UnorderedFlow { f.concurrent = b; return f }

// BeforeDo executes the do function before executing the Do method.
func (f *UnorderedFlow) BeforeDo(do func()) *UnorderedFlow { f.bdo = do; return f }

// BeforeUndo executes the undo function before executing the Undo method.
func (f *UnorderedFlow) BeforeUndo(undo func()) *UnorderedFlow { f.bundo = undo; return f }

// AfterDo executes the do function after executing the Do method.
func (f *UnorderedFlow) AfterDo(do func()) *UnorderedFlow { f.ado = do; return f }

// AfterUndo executes the undo function after executing the Undo method.
func (f *UnorderedFlow) AfterUndo(undo func()) *UnorderedFlow { f.aundo = undo; return f }

// Do does the tasks, which undoes them if a certain task fails.
func (f *UnorderedFlow) Do(c context.Context) (err error) {
	if f.bdo != nil {
		f.bdo()
	}
	if f.ado != nil {
		defer f.ado()
	}

	_len := len(f.tasks)
	resultch := make(chan taskResult, _len)
	for i := 0; i < _len; i++ {
		if f.concurrent {
			go f.runTask(c, f.tasks[i], resultch)
		} else {
			f.runTask(c, f.tasks[i], resultch)
		}
	}

	results := make([]taskResult, _len)
	for i := 0; i < _len; i++ {
		results[i] = <-resultch
	}

	var tes TaskErrors
	for _, r := range results {
		if r.DoErr != nil {
			tes.Append(r.Task.Name(), r.DoErr, r.UndoErr)
		}
	}

	if len(tes) != 0 {
		err = NewFlowError(f.name, tes...)
	}

	return
}

type taskResult struct {
	Task    Task
	DoErr   error
	UndoErr error
}

func (f *UnorderedFlow) runTask(c context.Context, task Task, r chan<- taskResult) {
	result := taskResult{Task: task}
	defer func() {
		r <- result
	}()

	if result.DoErr = task.Do(c); result.DoErr != nil {
		if ta, ok := task.(UndoAll); ok && f.undoAll {
			result.UndoErr = ta.UndoAll(c)
		} else {
			result.UndoErr = task.Undo(c)
		}

	}
}

// Undo does nothing. If undoing all the tasks, please use UndoAll.
func (f *UnorderedFlow) Undo(c context.Context) error { return nil }

// UndoAll undoes all the tasks.
func (f *UnorderedFlow) UndoAll(c context.Context) error {
	if tes := f.undo(c, f.tasks); len(tes) != 0 {
		return NewFlowError(f.name, tes...)
	}
	return nil
}

func (f *UnorderedFlow) undo(c context.Context, tasks []Task) (tes TaskErrors) {
	if !f.undoAll {
		return
	}

	if f.bundo != nil {
		f.bundo()
	}
	if f.aundo != nil {
		defer f.aundo()
	}

	for _, task := range tasks {
		var err error
		if ta, ok := task.(UndoAll); ok {
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
