package task

import (
	"sync"
)

type Job struct {
	h hT
}

type Worker struct {
	t  Task
	js []Job
}

type WorkerPool struct {
	ws []Worker
	wg *sync.WaitGroup

	idx int
	num int
}

func NewWorker(r rT) *Worker {
	return &Worker{
		t: WithRecover(r),
	}
}

func NewWorkerPool(num int, rs ...rT) *WorkerPool {
	ws := make([]Worker, 0, num)

	var r rT
	if len(rs) > 0 {
		r = rs[0]
	}

	for range num {
		ws = append(ws, *NewWorker(r))
	}

	return &WorkerPool{
		ws:  ws,
		wg:  &sync.WaitGroup{},
		idx: 0,
		num: num,
	}
}

func (it *WorkerPool) SubmitJob(h hT, rs ...rT) {
	if it.idx == it.num {
		it.idx = 0
	}

	if len(rs) > 0 {
		it.ws[it.idx].t = WithRecover(rs[0])
	}

	it.ws[it.idx].js = append(it.ws[it.idx].js, Job{
		h: h,
	})

	it.idx++
}

func (it *WorkerPool) Start() error {
	for _, w := range it.ws {
		for _, job := range w.js {
			w.t.Do(job.h)
		}
		w.t.Wait()
	}

	return nil
}
