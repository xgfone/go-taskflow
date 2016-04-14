package taskflow

type Task interface {
	Do() bool
	Undo() bool
}

type DoFunc func() bool
type UndoFunc func() bool

type Tasker struct {
	DoF   DoFunc
	UndoF UndoFunc
}

func (self Tasker) Do() bool {
	return self.DoF()
}

func (self Tasker) Undo() bool {
	return self.UndoF()
}
