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

	recover  recoverFunc
	canceler cancelerFunc

	mu sync.Mutex
}

func WithContext(
	ctx context.Context,
	recover recoverFunc,
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

func (g *withContext) Wait() error {
	g.wg.Wait()
	if g.canceler != nil {
		g.canceler(g.err)
	}
	return g.err
}

func (g *withContext) Do(handler func() error) {
	if g.sem != nil {
		g.sem <- token{}
	}

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
				if g.canceler != nil {
					g.canceler(g.err)
				}
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
