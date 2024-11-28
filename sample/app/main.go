package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/kuroko-shirai/task"
)

type chTask struct {
	State int
	Err   error
}

func worker(ctx context.Context, name string) error {
	fmt.Printf("%s started\n", name)
	time.Sleep(2 * time.Second)
	if name == "second group" {
		return fmt.Errorf("second group failed")
	}
	fmt.Printf("%s finished\n", name)
	return nil
}

func main() {
	{
		g := task.WithRecover(
			func(recovery any) {
				log.Println("panic:", recovery)
			},
		)

		ch := make(chan chTask, 1)
		g.Do(func(ctx context.Context, out chan<- chTask) func() error {
			return func() error {
				log.Println("first group")

				out <- chTask{
					State: 1,
					Err:   nil,
				}

				return nil
			}
		}(context.Background(), ch))

		g.Do(func() func() error {
			return func() error {
				log.Println("second group")
				panic(errors.New("panic in second group"))
				return nil
			}
		}())

		g.Do(func(ctx context.Context) func() error {
			return func() error {
				log.Println("third group")
				time.Sleep(1000 * time.Millisecond)
				return errors.New("error in third group")
			}
		}(context.Background()))

		if err := g.Wait(); err != nil {
			log.Println("something wrong in first batch", err)
		}

		fmt.Println(<-ch)
	}

	{

		g, ctx := task.WithContext(
			context.Background(),
			func(recovery any) {
				log.Println("panic:", recovery)
			},
		)

		g.Do(func() func() error {
			return func() error {
				return worker(ctx, "first group")
			}
		}())

		g.Do(func() func() error {
			return func() error {
				return worker(ctx, "second group")
			}
		}())

		g.Do(func() func() error {
			return func() error {
				return worker(ctx, "third group")
			}
		}())

		if err := g.Wait(); err != nil {
			log.Println("something wrong in second batch", err)
		}
	}
}
