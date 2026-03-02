package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

var errTest = errors.New("test error")

func TestBreakerStartsClosed(t *testing.T) {
	b := New(3, time.Second)
	if b.State() != StateClosed {
		t.Errorf("expected StateClosed, got %d", b.State())
	}
}

func TestBreakerOpensAfterMaxFailures(t *testing.T) {
	b := New(3, time.Second)

	for i := 0; i < 3; i++ {
		b.Execute(func() error { return errTest })
	}

	if b.State() != StateOpen {
		t.Errorf("expected StateOpen after 3 failures, got %d", b.State())
	}

	err := b.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreakerResetsOnSuccess(t *testing.T) {
	b := New(3, time.Second)

	// 2 failures then a success should reset count
	b.Execute(func() error { return errTest })
	b.Execute(func() error { return errTest })
	b.Execute(func() error { return nil })

	if b.State() != StateClosed {
		t.Errorf("expected StateClosed after success, got %d", b.State())
	}

	// Should need 3 more failures to open
	b.Execute(func() error { return errTest })
	b.Execute(func() error { return errTest })
	if b.State() != StateClosed {
		t.Errorf("expected StateClosed after only 2 new failures, got %d", b.State())
	}
}

func TestBreakerTransitionsToHalfOpen(t *testing.T) {
	b := New(2, 10*time.Millisecond)

	b.Execute(func() error { return errTest })
	b.Execute(func() error { return errTest })

	if b.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %d", b.State())
	}

	// Wait for reset timeout
	time.Sleep(15 * time.Millisecond)

	// Next call should transition to HalfOpen and succeed
	err := b.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("expected nil error in half-open, got %v", err)
	}

	if b.State() != StateHalfOpen {
		t.Errorf("expected StateHalfOpen after 1 success, got %d", b.State())
	}

	// Second success should close the circuit
	b.Execute(func() error { return nil })
	if b.State() != StateClosed {
		t.Errorf("expected StateClosed after 2 successes, got %d", b.State())
	}
}

func TestBreakerHalfOpenReOpensOnFailure(t *testing.T) {
	b := New(2, 10*time.Millisecond)

	b.Execute(func() error { return errTest })
	b.Execute(func() error { return errTest })

	time.Sleep(15 * time.Millisecond)

	// Fail in half-open should re-open the circuit
	b.Execute(func() error { return errTest })

	if b.State() != StateOpen {
		t.Errorf("expected StateOpen after half-open failure, got %d", b.State())
	}
}

func TestBreakerSuccessfulCallPassesThrough(t *testing.T) {
	b := New(3, time.Second)
	called := false

	err := b.Execute(func() error {
		called = true
		return nil
	})

	if !called {
		t.Error("function was not called")
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// State.String()
// ---------------------------------------------------------------------------

func TestStateString_Closed(t *testing.T) {
	if StateClosed.String() != "closed" {
		t.Errorf("StateClosed.String() = %q, want %q", StateClosed.String(), "closed")
	}
}

func TestStateString_Open(t *testing.T) {
	if StateOpen.String() != "open" {
		t.Errorf("StateOpen.String() = %q, want %q", StateOpen.String(), "open")
	}
}

func TestStateString_HalfOpen(t *testing.T) {
	if StateHalfOpen.String() != "half-open" {
		t.Errorf("StateHalfOpen.String() = %q, want %q", StateHalfOpen.String(), "half-open")
	}
}

func TestStateString_Unknown(t *testing.T) {
	unknown := State(99)
	if unknown.String() != "unknown" {
		t.Errorf("State(99).String() = %q, want %q", unknown.String(), "unknown")
	}
}

func TestStateString_AllValues(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(-1), "unknown"},
		{State(42), "unknown"},
		{State(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// New() constructor
// ---------------------------------------------------------------------------

func TestNew_InitializesFields(t *testing.T) {
	b := New(5, 10*time.Second)
	if b == nil {
		t.Fatal("New() returned nil")
	}
	if b.State() != StateClosed {
		t.Errorf("initial state = %v, want StateClosed", b.State())
	}
	if b.maxFailures != 5 {
		t.Errorf("maxFailures = %d, want 5", b.maxFailures)
	}
	if b.resetTimeout != 10*time.Second {
		t.Errorf("resetTimeout = %v, want 10s", b.resetTimeout)
	}
	if b.failures != 0 {
		t.Errorf("failures = %d, want 0", b.failures)
	}
	if b.successes != 0 {
		t.Errorf("successes = %d, want 0", b.successes)
	}
}

// ---------------------------------------------------------------------------
// Execute edge cases
// ---------------------------------------------------------------------------

func TestBreakerExecute_ReturnsOriginalError(t *testing.T) {
	b := New(10, time.Second)
	err := b.Execute(func() error { return errTest })
	if !errors.Is(err, errTest) {
		t.Errorf("expected errTest, got %v", err)
	}
}

func TestBreakerExecute_OpenReturnsErrCircuitOpen(t *testing.T) {
	b := New(1, time.Minute) // long timeout so it stays open
	b.Execute(func() error { return errTest })

	if b.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %v", b.State())
	}

	err := b.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreakerExecute_MultipleSuccessesStayClosed(t *testing.T) {
	b := New(3, time.Second)
	for i := 0; i < 100; i++ {
		err := b.Execute(func() error { return nil })
		if err != nil {
			t.Fatalf("unexpected error on success %d: %v", i, err)
		}
	}
	if b.State() != StateClosed {
		t.Errorf("expected StateClosed after many successes, got %v", b.State())
	}
}

func TestBreakerExecute_FailureCountResetOnSuccess(t *testing.T) {
	b := New(3, time.Second)

	// 2 failures
	b.Execute(func() error { return errTest })
	b.Execute(func() error { return errTest })

	// 1 success resets the count
	b.Execute(func() error { return nil })

	// 2 more failures should NOT open (need 3 consecutive)
	b.Execute(func() error { return errTest })
	b.Execute(func() error { return errTest })

	if b.State() != StateClosed {
		t.Errorf("expected StateClosed, got %v", b.State())
	}
}

func TestBreakerExecute_HalfOpenFailureReopens(t *testing.T) {
	b := New(1, 10*time.Millisecond)

	// Open the breaker
	b.Execute(func() error { return errTest })
	if b.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %v", b.State())
	}

	// Wait for reset timeout
	time.Sleep(15 * time.Millisecond)

	// Fail in half-open -> should reopen
	b.Execute(func() error { return errTest })
	if b.State() != StateOpen {
		t.Errorf("expected StateOpen after half-open failure, got %v", b.State())
	}
}

func TestBreakerExecute_HalfOpenTwoSuccessesClose(t *testing.T) {
	b := New(1, 10*time.Millisecond)

	b.Execute(func() error { return errTest })
	time.Sleep(15 * time.Millisecond)

	// First success: half-open
	b.Execute(func() error { return nil })
	if b.State() != StateHalfOpen {
		t.Fatalf("expected StateHalfOpen after 1 success, got %v", b.State())
	}

	// Second success: closed
	b.Execute(func() error { return nil })
	if b.State() != StateClosed {
		t.Errorf("expected StateClosed after 2 successes in half-open, got %v", b.State())
	}
}

func TestBreakerState_IsConcurrencySafe(t *testing.T) {
	b := New(100, time.Second)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			b.Execute(func() error { return nil })
			_ = b.State()
		}
	}()

	for i := 0; i < 100; i++ {
		b.Execute(func() error { return errTest })
		_ = b.State()
	}

	<-done
	// If no race condition panic, the test passes
}

func TestErrCircuitOpen_Message(t *testing.T) {
	if ErrCircuitOpen.Error() != "circuit breaker is open" {
		t.Errorf("ErrCircuitOpen.Error() = %q, want %q", ErrCircuitOpen.Error(), "circuit breaker is open")
	}
}
