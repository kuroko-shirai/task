package task

import (
	"context"
	"errors"
	"sync"
)

type withContext struct {
	wg sync.WaitGroup

	sem chan token

	err error

	recover  rT
	canceler cT

	mu sync.Mutex
}

func WithContext(
	ctx context.Context,
	recover rT,
) (*withContext, context.Context) {
	ctx, canceler := withCancelCause(ctx)

	return &withContext{
		recover:  recover,
		canceler: canceler,
		mu:       sync.Mutex{},
	}, ctx
}

func withCancelCause(
	parent context.Context,
) (context.Context, func(error)) {
	return context.WithCancelCause(parent)
}

func (g *withContext) Do(h hT, rs ...rT) {
	cr := g.recover
	if rs != nil {
		cr = rs[0]
	}

	if g.sem != nil {
		g.sem <- token{}
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

func (g *withContext) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
}

func (g *withContext) Wait() error {
	g.wg.Wait()
	if g.canceler != nil {
		g.canceler(g.err)
	}

	return g.err
}

func (g *withContext) lock(f func()) {
	g.mu.Lock()
	defer g.mu.Unlock()
	f()
}
