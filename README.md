# go-taskflow [![Build Status](https://github.com/xgfone/go-taskflow/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/go-taskflow/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-taskflow)](https://pkg.go.dev/github.com/xgfone/go-taskflow) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-taskflow/master/LICENSE)

A task flow supporting `Go1.7+`, you can use it to do and undo the tasks.

## Installation
```shell
$ go get -u github.com/xgfone/go-taskflow
```

## Example
```go
package main

import (
	"context"
	"fmt"

	"github.com/xgfone/go-taskflow"
)

func logf(msg string, args ...interface{}) error {
	fmt.Printf(msg+"\n", args...)
	return nil
}

func do(n string) taskflow.TaskFunc {
	return func(context.Context) error { return logf("do the task '%s'", n) }
}

func undo(n string) taskflow.TaskFunc {
	return func(context.Context) error { return logf("undo the task '%s'", n) }
}

func failDo(n string) taskflow.TaskFunc {
	return func(context.Context) error {
		logf("do the task '%s'", n)
		return fmt.Errorf("failure")
	}
}

func newTask(n string) taskflow.Task {
	return taskflow.NewTask(n, do(n), undo(n))
}

func newFailTask(n string) taskflow.Task {
	return taskflow.NewTask(n, failDo(n), undo(n))
}

func main() {
	flow1 := taskflow.NewLineFlow("lineflow1")
	flow1.
		AddTasks(
			newTask("task1"),
			newTask("task2"),
			newTask("task3"),
		)

	flow2 := taskflow.NewLineFlow("lineflow2")
	flow2.
		AddTasks(
			newTask("task4"),
			newFailTask("task5"),
			newTask("task6"),
		)

	flow3 := taskflow.NewLineFlow("lineflow3")
	flow3.AddTask("task7", do("task7"), undo("task7")) // Use task functions
	flow3.
		AddTasks(
			flow1,
			newTask("task8"),
			flow2,
			newTask("task9"),
		)

	err := flow3.Do(context.Background())
	fmt.Println(err)

	// Output:
	// do the task 'task7'
	// do the task 'task1'
	// do the task 'task2'
	// do the task 'task3'
	// do the task 'task8'
	// do the task 'task4'
	// do the task 'task5'
	// undo the task 'task4'
	// undo the task 'task8'
	// undo the task 'task3'
	// undo the task 'task2'
	// undo the task 'task1'
	// undo the task 'task7'
	// FlowError(name=lineflow3, errs=[FlowError(name=lineflow2, errs=[TaskError(name=task5, doerr=failure)])])
}
```
