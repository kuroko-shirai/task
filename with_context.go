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
) (Task, context.Context) {
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

func (it *withContext) Do(h hT, rs ...rT) {
	cr := it.recover
	if rs != nil {
		cr = rs[0]
	}

	if it.sem != nil {
		it.sem <- token{}
	}

	it.wg.Add(1)
	go func() {
		defer func() {

			it.done()

			if r := recover(); r != nil {
				cr(r)

				str, _ := r.(error)
				it.lock(func() {
					it.err = errors.Join(it.err, str)
				})
			}
		}()

		if err := h(); err != nil {
			it.lock(func() {
				it.err = errors.Join(it.err, err)
			})
		}
	}()
}

func (it *withContext) done() {
	if it.sem != nil {
		<-it.sem
	}

	it.wg.Done()
}

func (it *withContext) Wait() error {
	it.wg.Wait()
	if it.canceler != nil {
		it.canceler(it.err)
	}

	return it.err
}

func (it *withContext) lock(f func()) {
	it.mu.Lock()
	defer it.mu.Unlock()
	f()
}
