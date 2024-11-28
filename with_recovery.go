package task

import (
	"errors"
	"sync"
)

type withRecover struct {
	wg sync.WaitGroup

	sem chan token

	errOnce sync.Once
	err     error

	recover recoverFunc

	mu sync.Mutex
}

func WithRecover(recover recoverFunc) *withRecover {
	return &withRecover{
		recover: recover,
		mu:      sync.Mutex{},
	}
}

func (g *withRecover) Do(handler func() error) {
	g.wg.Add(1)
	go func() {
		defer func() {

			g.done()

			if r := recover(); r != nil {
				g.recover(r)

				str, _ := r.(error)

				g.mu.Lock()
				g.err = errors.Join(g.err, str)
				g.mu.Unlock()
			}
		}()

		if err := handler(); err != nil {
			g.errOnce.Do(func() {
				g.mu.Lock()
				g.err = errors.Join(g.err, err)
				g.mu.Unlock()
			})

		}
	}()
}

func (g *withRecover) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
}

func (g *withRecover) Wait() error {
	g.wg.Wait()

	return g.err
}
