package task

import (
	"context"
	"errors"
	"sync"
)

type withContext struct {
	wg sync.WaitGroup

	sem chan token

	errOnce sync.Once
	err     error

	recover  rT
	canceler cT
}

func WithContext(
	ctx context.Context,
	recover rT,
) (*withContext, context.Context) {
	ctx, canceler := withCancelCause(ctx)

	return &withContext{
		recover:  recover,
		canceler: canceler,
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
