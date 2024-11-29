package task

import (
	"errors"
	"sync"
)

type withRecover struct {
	wg sync.WaitGroup

	sem chan token

	err error

	recover rT

	mu sync.Mutex
}

func WithRecover(recover rT) *withRecover {
	return &withRecover{
		recover: recover,
		mu:      sync.Mutex{},
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

				str, _ := r.(error)
				g.lock(func() {
					g.err = errors.Join(g.err, str)
				})
			}
		}()

		if err := h(); err != nil {
			g.lock(func() {
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

	g.mu.Lock()
	err := g.err
	g.mu.Unlock()

	return err
}

func (g *withRecover) lock(f func()) {
	g.mu.Lock()
	defer g.mu.Unlock()
	f()
}
