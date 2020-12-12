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

func ExampleUnorderedFlow() {
	logf := func(msg string, args ...interface{}) { fmt.Printf(msg+"\n", args...) }
	do := func(n string) TaskFunc {
		return func(c context.Context) error { logf("do the task '%s'", n); return nil }
	}
	undo := func(n string) TaskFunc {
		return func(c context.Context) error { logf("undo the task '%s'", n); return nil }
	}
	failDo := func(n string) TaskFunc {
		return func(c context.Context) error {
			logf("do the task '%s'", n)
			return fmt.Errorf("failure")
		}
	}
	newTask := func(n string) Task { return NewTask(n, do(n), undo(n)) }
	newFailTask := func(n string) Task { return NewTask(n, failDo(n), undo(n)) }

	flow1 := NewUnorderedFlow("unorderedflow1")
	flow1.
		BeforeDo(func() { logf("do the task '%s'", flow1.Name()) }).
		AddTasks(
			newTask("task1"),
			newTask("task2"),
			newTask("task3"),
		)

	flow2 := NewUnorderedFlow("unorderedflow2")
	flow2.
		BeforeDo(func() { logf("do the task '%s'", flow2.Name()) }).
		AddTasks(
			newTask("task4"),
			newFailTask("task5"),
			newTask("task6"),
		)

	flow3 := NewUnorderedFlow("unorderedflow3")
	flow3.
		BeforeDo(func() { logf("do the task '%s'", flow3.Name()) }).
		AddTasks(
			newTask("task7"),
			flow1,
			newTask("task8"),
			flow2,
			newTask("task9"),
		)

	err := flow3.Do(context.Background())
	fmt.Println(err)

	// Output:
	// do the task 'unorderedflow3'
	// do the task 'task7'
	// do the task 'unorderedflow1'
	// do the task 'task1'
	// do the task 'task2'
	// do the task 'task3'
	// do the task 'task8'
	// do the task 'unorderedflow2'
	// do the task 'task4'
	// do the task 'task5'
	// undo the task 'task5'
	// do the task 'task6'
	// do the task 'task9'
	// FlowError(name=unorderedflow3, errs=[FlowError(name=unorderedflow2, errs=[TaskError(name=task5, doerr=failure)])])
}
