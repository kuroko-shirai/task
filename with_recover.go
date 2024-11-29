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

	recover rT
}

func WithRecover(recover rT) *withRecover {
	return &withRecover{
		recover: recover,
	}
}

func (g *withRecover) Do(h hT, rs ...rT) {
	cr := g.recover
	if rs != nil {
		cr = rs[0]
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
