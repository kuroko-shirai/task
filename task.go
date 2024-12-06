package task

type Task interface {
	Do(h hT, rs ...rT)
	Wait() error
}

type token struct{}

// hT - type used by handler
type hT func() error

// rt - type used by recover
type rT func(f any, args ...any)

// cT - type used by canceler
type cT func(error)
