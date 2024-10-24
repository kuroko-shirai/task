package task

type Task interface {
	Do()
}

type HandleFunc func()

type RecoverFunc func(f any)

type WithRecover struct {
	Task

	Handle  HandleFunc
	Recover RecoverFunc
}

func New(recover RecoverFunc, handle HandleFunc) Task {
	return &WithRecover{
		Recover: recover,
		Handle:  handle,
	}
}

func (task *WithRecover) Do() {
	go func(args ...any) {
		defer func() {
			if r := recover(); r != nil {
				task.Recover(r)
			}
		}()

		task.Handle()
	}()
}
