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

func WithRecover(recover rT) Task {
	return &withRecover{
		recover: recover,
		mu:      sync.Mutex{},
	}
}

func (it *withRecover) Do(h hT, rs ...rT) {
	cr := it.recover
	if rs != nil {
		cr = rs[0]
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

func (it *withRecover) done() {
	if it.sem != nil {
		<-it.sem
	}
	it.wg.Done()
}

func (it *withRecover) Wait() error {
	it.wg.Wait()

	it.mu.Lock()
	err := it.err
	it.mu.Unlock()

	return err
}

func (it *withRecover) lock(f func()) {
	it.mu.Lock()
	defer it.mu.Unlock()
	f()
}
