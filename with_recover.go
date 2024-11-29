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
}

func WithRecover(recover recoverFunc) *withRecover {
	return &withRecover{
		recover: recover,
	}
}

func (g *withRecover) Do(h func() error, s ...func(f any, args ...any)) {
	cr := g.recover
	if s != nil {
		cr = s[0]
	}

	g.wg.Add(1)
	go func() {
		defer func() {

			g.done()

			if r := recover(); r != nil {
				cr(r)

				g.errOnce.Do(func() {
					str, _ := r.(error)
					g.err = errors.Join(g.err, str)
				})
			}
		}()

		if err := h(); err != nil {
			g.errOnce.Do(func() {
				g.err = errors.Join(g.err, err)
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
