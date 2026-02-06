package signals

import (
	"reflect"
	"sync"
)

// Getter is a type-erased interface for Signals and Computeds
type Getter interface {
	GetValue() interface{}
}

// Dependency represents something that can be depended on (Signal, Computed)
type Dependency interface {
	subscribe(s Subscriber)
	unsubscribe(s Subscriber)
}

// Subscriber represents something that depends on others (Effect, Computed)
type Subscriber interface {
	onDependencyUpdated()
	addDependency(d Dependency)
}

// Global State
var (
	activeSubscriber Subscriber
	activeMu         sync.Mutex

	batchDepth int
	batchQueue map[Subscriber]struct{}
	batchMu    sync.Mutex
)

// Batch executes the given function in a batch.
// Updates are flushed only after the outermost batch finishes.
func Batch(fn func()) {
	batchMu.Lock()
	batchDepth++
	batchMu.Unlock()

	defer func() {
		batchMu.Lock()
		batchDepth--
		if batchDepth == 0 && len(batchQueue) > 0 {
			queue := batchQueue
			batchQueue = nil
			batchMu.Unlock()

			for sub := range queue {
				sub.onDependencyUpdated()
			}
		} else {
			batchMu.Unlock()
		}
	}()

	fn()
}

// Signal represents a reactive value
type Signal[T any] struct {
	value       T
	subscribers map[Subscriber]struct{}
	mu          sync.RWMutex
}

// New creates a new Signal with an initial value
func New[T any](val T) *Signal[T] {
	return &Signal[T]{
		value:       val,
		subscribers: make(map[Subscriber]struct{}),
	}
}

func (s *Signal[T]) subscribe(sub Subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribers[sub] = struct{}{}
}

func (s *Signal[T]) unsubscribe(sub Subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subscribers, sub)
}

func (s *Signal[T]) GetValue() interface{} {
	return s.Get()
}

func (s *Signal[T]) Get() T {
	// Dependency Tracking
	activeMu.Lock()
	current := activeSubscriber
	activeMu.Unlock()

	if current != nil {
		current.addDependency(s)
		s.subscribe(current)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

func (s *Signal[T]) Peek() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

func (s *Signal[T]) Set(val T) {
	s.mu.Lock()
	if reflect.DeepEqual(s.value, val) {
		s.mu.Unlock()
		return
	}
	s.value = val

	// Snapshot subscribers
	subs := make([]Subscriber, 0, len(s.subscribers))
	for sub := range s.subscribers {
		subs = append(subs, sub)
	}
	s.mu.Unlock()

	// Notify
	for _, sub := range subs {
		sub.onDependencyUpdated()
	}
}

// Computed represents a derived value
type Computed[T any] struct {
	fn           func() T
	value        T
	dirty        bool
	dependencies map[Dependency]struct{}
	subscribers  map[Subscriber]struct{}
	mu           sync.Mutex
}

func NewComputed[T any](fn func() T) *Computed[T] {
	return &Computed[T]{
		fn:           fn,
		dirty:        true, // Starts dirty so it evaluates on first Get
		dependencies: make(map[Dependency]struct{}),
		subscribers:  make(map[Subscriber]struct{}),
	}
}

func (c *Computed[T]) subscribe(sub Subscriber) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribers[sub] = struct{}{}
}

func (c *Computed[T]) unsubscribe(sub Subscriber) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribers, sub)
}

func (c *Computed[T]) addDependency(d Dependency) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dependencies[d] = struct{}{}
}

func (c *Computed[T]) onDependencyUpdated() {
	c.mu.Lock()
	if c.dirty {
		c.mu.Unlock()
		return
	}
	c.dirty = true

	// Notify subscribers
	subs := make([]Subscriber, 0, len(c.subscribers))
	for sub := range c.subscribers {
		subs = append(subs, sub)
	}
	c.mu.Unlock()

	for _, sub := range subs {
		sub.onDependencyUpdated()
	}
}

func (c *Computed[T]) GetValue() interface{} {
	return c.Get()
}

func (c *Computed[T]) Get() T {
	// 1. Check if we are being tracked
	activeMu.Lock()
	current := activeSubscriber
	activeMu.Unlock()

	if current != nil {
		current.addDependency(c)
		c.subscribe(current)
	}

	c.mu.Lock()
	if c.dirty {
		// Re-evaluate
		// Cleanup old dependencies first
		for dep := range c.dependencies {
			dep.unsubscribe(c)
		}
		c.dependencies = make(map[Dependency]struct{})

		// Set as active subscriber
		activeMu.Lock()
		prev := activeSubscriber
		activeSubscriber = c
		activeMu.Unlock()

		// Run function
		// We need to unlock c.mu while running fn to avoid deadlocks if fn accesses other signals
		c.mu.Unlock()
		val := c.fn()
		c.mu.Lock()

		c.value = val
		c.dirty = false

		// Restore active subscriber
		activeMu.Lock()
		activeSubscriber = prev
		activeMu.Unlock()
	}
	defer c.mu.Unlock()
	return c.value
}

// Effect represents a side effect
type Effect struct {
	fn           func()
	dependencies map[Dependency]struct{}
	mu           sync.Mutex
	disposed     bool
}

func CreateEffect(fn func()) *Effect {
	e := &Effect{
		fn:           fn,
		dependencies: make(map[Dependency]struct{}),
	}
	e.Run()
	return e
}

func (e *Effect) addDependency(d Dependency) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.dependencies[d] = struct{}{}
}

func (e *Effect) onDependencyUpdated() {
	// Check batching
	batchMu.Lock()
	if batchDepth > 0 {
		if batchQueue == nil {
			batchQueue = make(map[Subscriber]struct{})
		}
		batchQueue[e] = struct{}{}
		batchMu.Unlock()
		return
	}
	batchMu.Unlock()

	e.Run()
}

func (e *Effect) Run() {
	e.mu.Lock()
	if e.disposed {
		e.mu.Unlock()
		return
	}

	// Cleanup old dependencies
	// We need to track new dependencies during this run
	// So we unsubscribe from everything first?
	// Or we diff?
	// Simpler to unsubscribe all, then re-subscribe as we run.
	// This is slightly inefficient but correct.
	// Optimization: Double buffering dependencies.

	oldDeps := e.dependencies
	e.dependencies = make(map[Dependency]struct{})
	e.mu.Unlock()

	for dep := range oldDeps {
		dep.unsubscribe(e)
	}

	// Set active
	activeMu.Lock()
	prev := activeSubscriber
	activeSubscriber = e
	activeMu.Unlock()

	// Run
	e.fn()

	// Restore
	activeMu.Lock()
	activeSubscriber = prev
	activeMu.Unlock()
}

func (e *Effect) Dispose() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.disposed {
		return
	}
	e.disposed = true
	for dep := range e.dependencies {
		dep.unsubscribe(e)
	}
	e.dependencies = nil
}
