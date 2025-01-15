package task

type Task interface {
	Do(h HandlerType, rs ...RecoverType)
	Wait() error
}

type token struct{}

// HandlerType - type used by handler
type HandlerType func() error

// RecoverType - type used by recover
type RecoverType func(f any, args ...any)

// CancellerType - type used by canceler
type CancellerType func(error)
