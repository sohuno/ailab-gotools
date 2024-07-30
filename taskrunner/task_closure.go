package taskrunner

type TaskItemId int64

type TaskClosure interface {
	Run()
}
