package signals

import (
	"sync"
)

// Getter is a type-erased interface for Signals
type Getter interface {
	GetValue() interface{}
}

// Subscriber is an interface for anything that depends on a signal
type Subscriber interface {
	OnUpdate()
}

// Signal represents a reactive value
type Signal[T any] struct {
	value       T
	subscribers []Subscriber
	mu          sync.RWMutex
}

// New creates a new Signal with an initial value
func New[T any](val T) *Signal[T] {
	return &Signal[T]{
		value: val,
	}
}

// GetValue implements the Getter interface
func (s *Signal[T]) GetValue() interface{} {
	return s.Get()
}

// Get returns the current value and tracks dependency if called within an Effect
func (s *Signal[T]) Get() T {
	// We need to be careful with locking order.
	// First, capture the active effect if any.
	// Accessing the global activeEffect is technically a race if multiple goroutines
	// are running effects. For this MVP, we assume UI effects run on the main thread.
	effect := activeEffect

	if effect != nil {
		s.subscribe(effect)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Peek returns the current value without tracking dependency
func (s *Signal[T]) Peek() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set updates the value and notifies subscribers
func (s *Signal[T]) Set(val T) {
	s.mu.Lock()

	// Fast equality check using interface comparison.
	// This uses == for comparable types (int, string, pointers) which is O(1).
	// For non-comparable types (structs with slices, linked lists), the recover
	// skips the check and always propagates — safe and avoids the catastrophic
	// cost of reflect.DeepEqual on cyclic structures (e.g. doubly-linked LayoutNodes).
	if fastEqual(s.value, val) {
		s.mu.Unlock()
		return
	}

	s.value = val
	// Copy subscribers to avoid holding lock during notification
	subs := make([]Subscriber, len(s.subscribers))
	copy(subs, s.subscribers)
	s.mu.Unlock()

	for _, sub := range subs {
		sub.OnUpdate()
	}
}

// fastEqual compares two values using interface == (pointer/value equality).
// Returns false for non-comparable types instead of panicking.
func fastEqual[T any](a, b T) bool {
	defer func() { recover() }()
	return any(a) == any(b)
}

func (s *Signal[T]) subscribe(sub Subscriber) {
	s.mu.Lock() // Upgrade to Write Lock
	defer s.mu.Unlock()

	// Check if already subscribed to avoid duplicates
	for _, existing := range s.subscribers {
		if existing == sub {
			return
		}
	}
	s.subscribers = append(s.subscribers, sub)
}

// Effect represents a side effect that runs when signals change
type Effect struct {
	fn func()
}

// OnUpdate implements the Subscriber interface
func (e *Effect) OnUpdate() {
	e.Run()
}

// Run executes the effect function while tracking dependencies
func (e *Effect) Run() {
	// Note: This global variable approach is not goroutine-safe.
	// Effects should ideally be run on a single UI thread.
	prevEffect := activeEffect
	activeEffect = e
	defer func() { activeEffect = prevEffect }()

	e.fn()
}

var activeEffect *Effect

// CreateEffect creates and runs a new effect
func CreateEffect(fn func()) *Effect {
	e := &Effect{fn: fn}
	e.Run()
	return e
}

// Computed represents a value derived from other signals
type Computed[T any] struct {
	sig *Signal[T]
	fn  func() T
}

// NewComputed creates a new Computed value
func NewComputed[T any](fn func() T) *Computed[T] {
	c := &Computed[T]{
		fn: fn,
	}
	// Create an internal signal to hold the result
	var zero T
	c.sig = New(zero)

	// Create an effect that updates the internal signal whenever dependencies change
	CreateEffect(func() {
		c.sig.Set(c.fn())
	})

	return c
}

// Get returns the computed value (and tracks dependency on the internal signal)
func (c *Computed[T]) Get() T {
	return c.sig.Get()
}

// GetValue implements the Getter interface for Computed
func (c *Computed[T]) GetValue() interface{} {
	return c.Get()
}
