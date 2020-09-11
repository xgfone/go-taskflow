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
	"fmt"
	"strings"
)

// FlowError represents the flow error.
type FlowError struct {
	Name string
	TaskErrors
}

// NewFlowError returns a new FlowError.
func NewFlowError(name string, errs ...TaskError) FlowError {
	return FlowError{Name: name, TaskErrors: errs}
}

func (e FlowError) Unwrap() error {
	if len(e.TaskErrors) != 0 {
		return e.TaskErrors
	}
	return nil
}

func (e FlowError) Error() string {
	if len(e.TaskErrors) == 0 {
		return fmt.Sprintf("FlowError(name=%s)", e.Name)
	}

	return fmt.Sprintf("FlowError(name=%s, errs=[%s])", e.Name, e.TaskErrors.Error())
}

// TaskError is used to represent the task error.
type TaskError interface {
	error

	Name() string
	DoErr() error
	UndoErr() error
}

// NewTaskError returns a new TaskError.
func NewTaskError(name string, doErr, undoErr error) TaskError {
	return taskError{name: name, doerr: doErr, undoerr: undoErr}
}

type taskError struct {
	name    string
	doerr   error
	undoerr error
}

func (e taskError) Name() string   { return e.name }
func (e taskError) DoErr() error   { return e.doerr }
func (e taskError) UndoErr() error { return e.undoerr }
func (e taskError) Unwrap() error {
	if e.doerr != nil {
		return e.doerr
	}
	return e.undoerr
}
func (e taskError) Error() string {
	if e.doerr == nil {
		if e.undoerr == nil {
			return fmt.Sprintf("TaskError(name=%s)", e.name)
		}

		if fe, ok := e.undoerr.(FlowError); ok && fe.Name == e.name {
			return fe.Error()
		}

		return fmt.Sprintf("TaskError(name=%s, undoerr=%s)", e.name, e.undoerr.Error())
	}

	if e.undoerr == nil {
		if fe, ok := e.doerr.(FlowError); ok && fe.Name == e.name {
			return fe.Error()
		}

		return fmt.Sprintf("TaskError(name=%s, doerr=%s)", e.name, e.doerr.Error())
	}

	return fmt.Sprintf("TaskError(name=%s, doerr=%s, undoerr=%s)",
		e.name, e.doerr.Error(), e.undoerr.Error())
}

// TaskErrors is a set of TaskError.
type TaskErrors []TaskError

func (es TaskErrors) Error() string {
	ss := make([]string, len(es))
	for i, e := range es {
		ss[i] = e.Error()
	}
	return strings.Join(ss, ", ")
}

// Append is equal to es.AppendTaskError(NewTaskError(name, doErr, undoErr)).
func (es *TaskErrors) Append(name string, doErr, undoErr error) {
	es.AppendTaskError(NewTaskError(name, doErr, undoErr))
}

// AppendTaskError appends the task error.
func (es *TaskErrors) AppendTaskError(e TaskError) { *es = append(*es, e) }
