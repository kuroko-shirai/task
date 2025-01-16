package task

import (
	"sync"
)

type background struct {
	wg sync.WaitGroup
}

func Background() Task {
	return &background{}
}

func (it *background) Do(h HandlerType, rs ...RecoverType) {
	it.wg.Add(1)

	go func() {
		defer it.done()

		h()
	}()
}

func (it *background) done() {
	it.wg.Done()
}

func (it *background) Wait() error {
	it.wg.Wait()

	return nil
}
