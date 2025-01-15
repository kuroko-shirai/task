package task

import (
	"errors"
	"sync"
)

type job struct {
	h HandlerType
}

type worker struct {
	t  Task
	js []job
}

type workerPool struct {
	ws []worker
	wg *sync.WaitGroup

	idx int
	num int
}

func newWorker(r RecoverType) *worker {
	return &worker{
		t: WithRecover(r),
	}
}

func WorkerPool(num int, rs ...RecoverType) *workerPool {
	ws := make([]worker, 0, num)

	var r RecoverType
	if len(rs) > 0 {
		r = rs[0]
	}

	for range num {
		ws = append(ws, *newWorker(r))
	}

	return &workerPool{
		ws:  ws,
		wg:  &sync.WaitGroup{},
		idx: 0,
		num: num,
	}
}

func (it *workerPool) SubmitJob(h HandlerType, rs ...RecoverType) {
	if it.idx == it.num {
		it.idx = 0
	}

	if len(rs) > 0 {
		it.ws[it.idx].t = WithRecover(rs[0])
	}

	it.ws[it.idx].js = append(it.ws[it.idx].js, job{
		h: h,
	})

	it.idx++
}

func (it *workerPool) Start() error {
	var es error

	for _, w := range it.ws {
		for _, job := range w.js {
			w.t.Do(job.h)
		}
		if err := w.t.Wait(); err != nil {
			es = errors.Join(es, err)
		}
	}

	return es
}
