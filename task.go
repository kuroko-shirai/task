package task

type token struct{}

type handlerFunc func() error

type recoverFunc func(f any)

type cancelerFunc func(error)
