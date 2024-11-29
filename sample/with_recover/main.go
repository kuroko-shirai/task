package main

import (
	"errors"
	"log"

	"github.com/kuroko-shirai/task"
)

type chTask struct {
	State int
	Err   error
}

func main() {
	g := task.WithRecover(
		func(recovery any, args ...any) {
			log.Println("a common handler of panic:", recovery)
		},
	)

	ch := make(chan chTask, 1)
	g.Do(func(out chan<- chTask) func() error {
		return func() error {
			log.Println("worker-1 started")

			out <- chTask{
				State: 1,
				Err:   nil,
			}

			return nil
		}
	}(ch))

	arg2 := 1
	g.Do(
		func(arg int) func() error {
			return func() error {
				log.Println("worker-2 started")

				panic(errors.New("worker-2 got panic"))
			}
		}(arg2),
		func(arg int) func(f any, args ...any) {
			return func(p any, args ...any) {
				log.Println("a custom handler of panic with arg:", p, arg)
			}
		}(arg2),
	)

	g.Do(func() func() error {
		return func() error {
			log.Println("worker-3 started")

			panic(errors.New("worker-3 got panic"))
		}
	}())

	arg := 1
	g.Do(func(arg int) func() error {
		return func() error {
			log.Println("worker-4 started with arg =", arg)

			return errors.New("worker-4 got error")
		}
	}(arg))

	if err := g.Wait(); err != nil {
		log.Println("something wrong in first batch:\n", err.Error())
	}

	log.Println("ch:", <-ch)
}
