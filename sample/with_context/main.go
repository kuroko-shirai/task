package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/kuroko-shirai/task"
)

func worker(ctx context.Context, name string) error {
	log.Println(name, "started")

	if name == "worker-2" {
		panic(errors.New("worker-2 got panic"))
	}

	select {
	case <-ctx.Done():
		log.Printf("worker %s stopped by context\n", name)

		return ctx.Err()
	case <-time.After(2 * time.Second):
		log.Printf("worker %s finished\n", name)

		return nil
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := task.WithContext(
		ctx,
		func(p any, args ...any) {
			log.Println("panic:", p)
		},
	)

	for i := 1; i <= 3; i++ {
		g.Do(
			func(ctx context.Context) func() error {
				return func() error {
					return worker(ctx, fmt.Sprintf("worker-%d", i))
				}
			}(ctx),
		)
	}

	time.AfterFunc(1*time.Second, func() {
		cancel()
	})

	if err := g.Wait(); err != nil {
		log.Println("error:", err)
	}
}
