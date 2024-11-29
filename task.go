package task

type token struct{}

type handlerFunc func() error

type recoverFunc func(f any, args ...any)

type cancelerFunc func(error)
