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
signature `func(recovery any)`. For example, let's log the
panic message.

```go
g := task.WithRecover(
	func(recovery any) {
		log.Println("panic:", recovery)
	},
)
```

Next, we can schedule the asynchronous execution of the
task. Note that within the `Do` method, a function closure
is featured.

```go
g.Do(func() func() error {
	return func() error {
		...
		return nil
	}
}())
```

This allows us to pass arbitrary argument lists. For
example, we can pass a channel to return some state
from the handler.

```go
ch := make(chan chTask, 1)
g.Do(func(out chan<- chTask) func() error {
	return func() error {
		out <- chTask{
			State: 1,
			Err:   nil,
		}

		return nil
	}
}(ch))
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

Пользователь может передавать аргументы в обработчик,
указавая сигнатуру функтора `h` внутри метода `Do`.

```go
arg := 1
g.Do(func(arg int) func() error {
	return func() error {
		...
	}
}(arg))
```

Также пользователь может вызывать не предустановленные
обработчики паники, если у вас есть появится такая
необходимость. Это может оказаться полезным, если нужно
перехватить конкретный момент выполнения запроса, либо
залогировать аргументы, при которых функтор `h` дает сбой.

```go
g.Do(func() func() error {
	return func() error {
		...
	}
}(), func(recovery any) {
	log.Println("a custom handler of panic:", recovery)
})
```

### With Context

Также как и в пакете `errgroup` вы можете определять
самостоятельную работу с контекстом. Например, мы можем
определить функцию `worker(context.Context, string) error`,
которая запускает панику для одного обработчика, а также
обрабатывает состояния контекста через `select`.

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

Как вы можете помнить, пакет `errgroup` не дает возможности
сохранить состояние всех ошибок, завершая работу группы при
наличии проблемы. Пакет `task` позволяет сохраняет состояния
ошибок обработчиков и даже при возникновении паники
позволяет пользователю определить сценарии восстановления
системы без завершения работы микросервиса. Чтобы подключить
контекст создайте группу задач с сценарием восстановления.

```go
ctx, cancel := context.WithCancel(context.Background())
g, ctx := task.WithContext(
	ctx,
	func(recovery any) {
		log.Println("panic:", recovery)
	},
)
```

Затем
