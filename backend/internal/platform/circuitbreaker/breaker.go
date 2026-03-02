package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Breaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	successes    int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
}

func New(maxFailures int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (b *Breaker) Execute(fn func() error) error {
	b.mu.Lock()
	if b.state == StateOpen {
		if time.Since(b.lastFailure) > b.resetTimeout {
			b.state = StateHalfOpen
			b.successes = 0
		} else {
			b.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	b.mu.Unlock()

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.failures++
		b.lastFailure = time.Now()
		if b.failures >= b.maxFailures {
			b.state = StateOpen
		}
		return err
	}

	if b.state == StateHalfOpen {
		b.successes++
		if b.successes >= 2 {
			b.state = StateClosed
			b.failures = 0
			b.successes = 0
		}
	} else {
		b.failures = 0
	}

	return nil
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// String returns a human-readable name for a circuit breaker state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
