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

// Flow is used to execute the task by the certain relationship.
type Flow interface {
	Task

	// Add adds a given Task/Flow into this flow.
	Add(tasks ...Task)
}

// FlowBuilder is used to build the Flow.
type FlowBuilder struct {
	fail       bool
	undoAll    bool
	concurrent bool

	aundo func()
	bundo func()
	ado   func()
	bdo   func()
}

// NewFlowBuilder returns a new FlowBuilder.
func NewFlowBuilder() FlowBuilder {
	return FlowBuilder{}
}

// Concurrent sets whether to do the tasks concurrently for UnorderedFlow.
func (b FlowBuilder) Concurrent(t bool) FlowBuilder { b.concurrent = t; return b }

// UndoFailedTask sets whether to undo the failed task or not for LineFlow.
func (b FlowBuilder) UndoFailedTask(t bool) FlowBuilder { b.fail = t; return b }

// SetUndoAllTasks sets whether to undo all the tasks or not
// if the task has implemented the interface UndoAll.
func (b FlowBuilder) SetUndoAllTasks(t bool) FlowBuilder { b.undoAll = t; return b }

// BeforeDo executes the do function before executing the Do method.
func (b FlowBuilder) BeforeDo(do func()) FlowBuilder { b.bdo = do; return b }

// BeforeUndo executes the undo function before executing the Undo method.
func (b FlowBuilder) BeforeUndo(undo func()) FlowBuilder { b.bundo = undo; return b }

// AfterDo executes the do function after executing the Do method.
func (b FlowBuilder) AfterDo(do func()) FlowBuilder { b.ado = do; return b }

// AfterUndo executes the undo function after executing the Undo method.
func (b FlowBuilder) AfterUndo(undo func()) FlowBuilder { b.aundo = undo; return b }

// LineFlow creates a new LineFlow.
func (b FlowBuilder) LineFlow(name string) *LineFlow {
	return NewLineFlow(name).
		BeforeDo(b.bdo).
		BeforeUndo(b.bundo).
		AfterDo(b.ado).
		AfterUndo(b.aundo).
		SetUndoAllTasks(b.undoAll).
		UndoFailedTask(b.fail)
}

// UnorderedFlow creates a new UnorderedFlow.
func (b FlowBuilder) UnorderedFlow(name string) *UnorderedFlow {
	return NewUnorderedFlow(name).
		BeforeDo(b.bdo).
		BeforeUndo(b.bundo).
		AfterDo(b.ado).
		AfterUndo(b.aundo).
		SetUndoAllTasks(b.undoAll).
		Concurrent(b.concurrent)
}
