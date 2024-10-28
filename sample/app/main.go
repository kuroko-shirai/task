package main

import (
	"context"
	"log"
	"sync"

	"github.com/kuroko-shirai/task"
)

type ChCacheGetItemsOnOff struct {
	State int
	Err   error
}

func main() {
	var wg sync.WaitGroup

	ch := make(chan ChCacheGetItemsOnOff, 1)

	wg.Add(1)

	newTask := task.New(
		func(recovery any) {
			log.Printf("Panic in the workflow process! %!w", recovery)
		},
		func(ctx context.Context, out chan<- ChCacheGetItemsOnOff) func() {
			return func() {
				defer wg.Done()

				out <- ChCacheGetItemsOnOff{
					State: 1,
					Err:   nil,
				}
			}
		}(context.Background(), ch),
	)
	newTask.Do()

	wg.Wait()

	close(ch)

	result := <-ch

	log.Printf("state: %d", result.State)
}
