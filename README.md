# go-taskflow
A task flow in Go, you can use it to do and undo.

## Installation
```shell
$ go get github.com/xgfone/go-taskflow
```

## Example
```go
package main

import (
    "fmt"

    "github.com/xgfone/go-taskflow"
)

type Task string

func (self Task) Do() bool {
    fmt.Println("Do " + self)
    return true
}

func (self Task) Undo() bool {
    fmt.Println("Undo " + self)
    return true
}

func do1() bool {
    fmt.Println("Do1")
    return true
}

func undo1() bool {
    fmt.Println("Undo1")
    return true
}

func do2() bool {
    fmt.Println("Do2")
    return false
}

func undo2() bool {
    fmt.Println("Undo2")
    return false
}

func do3() bool {
    fmt.Println("Do3")
    return true
}

func undo3() bool {
    fmt.Println("Undo3")
    return false
}

func main() {
    flow := taskflow.NewLineFlow()
    flow.Retry = 1

    task1 := taskflow.Tasker{DoF: do1, UndoF: undo1}
    task2 := taskflow.Tasker{DoF: do2, UndoF: undo2}
    task3 := taskflow.Tasker{DoF: do3, UndoF: undo3}
    var task Task = "task"

    flow.Add(task1).Add(task).Add(task2).Add(task3)
    if err := flow.Execute(); err != nil {
        if err == taskflow.DoError {
            fmt.Println("DoError")
        } else if err == taskflow.UndoError {
            fmt.Println("UndoError")
        }
    } else {
        fmt.Println("Successfully")
    }
}
// # Output:
// Do1
// Do task
// Do2
// Do2
// Undo task
// Undo1
// DoError
```