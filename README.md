# 📋 TASK

## Description

This package allows grouping goroutines and tracking their
errors, even in the event of a panic, enabling you to
prevent service crashes by handling exceptions and
monitoring errors in parallel-running goroutines.

## Usage Instructions

### With Recover

To create a task group, it is enough to define a function
for recovering from a panic. This can be a simple log
message or a more complex operation to roll back to some
state. Note that the recovery function must have the
signature `func(r any, args ...any)`. For example, let's log the
panic message.

```go
g := task.WithRecover(
	func(p any, args ...any) {
		log.Println("panic:", p)
	},
)
```

Next, we can schedule the asynchronous execution of the
task. Note that within the `Do` method, a function closure
is featured.

```go
g.Do(
	func() func() error {
		return func() error {
			...
			return nil
		}
	}(),
)
```

This allows us to pass arbitrary argument lists. For
example, we can pass a channel to return some state
from the handler.

```go
ch := make(chan chTask, 1)
g.Do(
	func(out chan<- chTask) func() error {
		return func() error {
			...
			out <- chTask{
				State: 1,
			}
			...
			return nil
		}
	}(ch),
)
...
g.Wait()
res := <-ch
```

This is safe and enables easy aggregation of the results of
the asynchronous execution of a group of tasks.

To wait for the execution of a group of tasks, it is
sufficient to call the `Wait` method. Note that in the event
of a panic, the service will not crash, but rather the
situation that occurred during the panic will be recorded in
the list of errors. When we request the list of errors at
the end of the `Wait` method, the panic message will be
included among them.

```go
if err := g.Wait(); err != nil {
	...
}
```

The user can pass arguments to the handler by specifying the
signature of the functor `h` inside the `Do` method.

```go
arg := 1
g.Do(
	func(arg int) func() error {
		return func() error {
			log.Println(arg)
			...
			return nil
		}
	}(arg),
)
```

Additionally, the user can invoke non-standard panic
handlers if the need arises. This can be useful if you need
to intercept a specific moment of the request execution, or
log the arguments under which the functor `h` fails.

```go
arg := 1
g.Do(
	func(arg int) func() error {
		return func() error {
			log.Println("worker-2 started")

			panic(errors.New("worker-2 got panic"))
		}
	}(arg),
	func(arg int) func(p any, args ...any) {
		return func(p any, args ...any) {
			log.Println("a custom handler of panic with arg:", p, arg)
		}
	}(arg),
)
```

### With Context

Just like in the `errgroup` package, you can define your own
context handling. For example, we can define a function
`worker(context.Context, string) error` that triggers a
panic for one handler, and also handles context states
via select.

```go
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
```

As you may recall, the `errgroup` package does not provide
the ability to save the state of all errors, terminating the
group's work in case of a problem. The `task` package saves
the error states of handlers and even in the event of a
panic, allows the user to define system recovery scenarios
without terminating the microservice. To connect the
context, create a task group with a recovery scenario.

```go
ctx, cancel := context.WithCancel(context.Background())
g, ctx := task.WithContext(
	ctx,
	func(p any, args ...any) {
		log.Println("panic:", p)
	},
)
```

Then the user can add to the task group.

```go
for i := 1; i <= 3; i++ {
	g.Do(
		func(ctx context.Context) func() error {
			return func() error {
				return worker(ctx, fmt.Sprintf("worker-%d", i))
			}
		}(ctx),
	)
}
```

To create a situation where the context is cancelled, we
will place a pause before waiting for the task group to
finish.

```go
time.AfterFunc(1*time.Second, func() {
	cancel()
})

if err := g.Wait(); err != nil {
	log.Println("error:", err)
}
```

When executing the code, we will get the following messages,
demonstrating the panic handling and context cancellation
for a pair of handlers.

```
2024/11/29 14:06:03 worker-2 started
2024/11/29 14:06:03 panic: worker-2 got panic
2024/11/29 14:06:03 worker-3 started
2024/11/29 14:06:03 worker-1 started
2024/11/29 14:06:04 worker worker-3 stopped by context
2024/11/29 14:06:04 worker worker-1 stopped by context
2024/11/29 14:06:04 error: worker-2 got panic
```
