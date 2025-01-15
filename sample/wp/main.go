package main

import (
	"log"

	"github.com/kuroko-shirai/task"
)

func main() {
	wp := task.WorkerPool(3, func() func(f any, args ...any) {
		return func(p any, args ...any) {
			log.Println("a common handler of panic with arg:", p)
		}
	}())

	for i := 0; i < 10; i++ {
		wp.SubmitJob(
			func(arg int) func() error {
				return func() error {
					log.Printf("job %d started", arg)

					return nil
				}
			}(i),
			func(arg int) func(f any, args ...any) {
				return func(p any, args ...any) {
					log.Println("a custom handler of panic with arg:", p, arg)
				}
			}(i),
		)
	}

	if err := wp.Start(); err != nil {
		log.Println("error", err)
	}
}
