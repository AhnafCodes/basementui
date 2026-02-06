package signals

import (
	"testing"
)

func TestSignal(t *testing.T) {
	count := New(0)
	if count.Get() != 0 {
		t.Errorf("Expected 0, got %d", count.Get())
	}

	count.Set(1)
	if count.Get() != 1 {
		t.Errorf("Expected 1, got %d", count.Get())
	}
}

func TestEffect(t *testing.T) {
	count := New(0)
	runCount := 0

	CreateEffect(func() {
		_ = count.Get()
		runCount++
	})

	if runCount != 1 {
		t.Errorf("Effect should run immediately. Got %d", runCount)
	}

	count.Set(1)
	if runCount != 2 {
		t.Errorf("Effect should run on update. Got %d", runCount)
	}

	count.Set(2)
	if runCount != 3 {
		t.Errorf("Effect should run on update. Got %d", runCount)
	}
}

func TestComputed(t *testing.T) {
	count := New(1)
	double := NewComputed(func() int {
		return count.Get() * 2
	})

	if double.Get() != 2 {
		t.Errorf("Expected 2, got %d", double.Get())
	}

	count.Set(2)
	if double.Get() != 4 {
		t.Errorf("Expected 4, got %d", double.Get())
	}
}

func TestDependencyTracking(t *testing.T) {
	a := New(1)
	b := New(2)
	sum := 0

	CreateEffect(func() {
		sum = a.Get() + b.Get()
	})

	if sum != 3 {
		t.Errorf("Expected 3, got %d", sum)
	}

	a.Set(2)
	if sum != 4 {
		t.Errorf("Expected 4, got %d", sum)
	}

	b.Set(3)
	if sum != 5 {
		t.Errorf("Expected 5, got %d", sum)
	}
}

func TestBatching(t *testing.T) {
	a := New(1)
	b := New(1)
	runs := 0

	CreateEffect(func() {
		_ = a.Get() + b.Get()
		runs++
	})

	if runs != 1 {
		t.Errorf("Initial run failed")
	}

	// Without batching, this would trigger 2 runs
	Batch(func() {
		a.Set(2)
		b.Set(2)
	})

	if runs != 2 {
		t.Errorf("Expected 2 runs (1 initial + 1 batch), got %d", runs)
	}
}

func TestEffectDispose(t *testing.T) {
	a := New(1)
	runs := 0

	eff := CreateEffect(func() {
		_ = a.Get()
		runs++
	})

	if runs != 1 {
		t.Errorf("Initial run failed")
	}

	a.Set(2)
	if runs != 2 {
		t.Errorf("Update failed")
	}

	eff.Dispose()
	a.Set(3)
	if runs != 2 {
		t.Errorf("Effect should not run after dispose")
	}
}

func TestConditionalDependency(t *testing.T) {
	flag := New(true)
	a := New(10)
	b := New(20)

	var lastVal int

	CreateEffect(func() {
		if flag.Get() {
			lastVal = a.Get()
		} else {
			lastVal = b.Get()
		}
	})

	if lastVal != 10 {
		t.Errorf("Expected 10, got %d", lastVal)
	}

	// Switch branch
	flag.Set(false)
	if lastVal != 20 {
		t.Errorf("Expected 20, got %d", lastVal)
	}

	// Update 'a' - should NOT trigger effect because we are on 'b' branch
	a.Set(30)
	// If it triggered, lastVal would still be 20 (since flag is false),
	// but we can check if effect ran by side effect or just trust the logic.
	// Better: check if 'a' has subscribers? We can't easily access private map.
	// Let's use a counter.

	runs := 0
	CreateEffect(func() {
		runs++
		if flag.Get() {
			_ = a.Get()
		} else {
			_ = b.Get()
		}
	})
	// runs = 1 (initial)

	flag.Set(true) // runs = 2. Now depends on 'a', not 'b'.

	b.Set(99) // Should NOT trigger
	if runs != 2 {
		t.Errorf("Stale dependency 'b' triggered update")
	}

	a.Set(40) // Should trigger
	if runs != 3 {
		t.Errorf("Active dependency 'a' failed to trigger update")
	}
}

func TestComputedLaziness(t *testing.T) {
	a := New(1)
	evals := 0

	c := NewComputed(func() int {
		evals++
		return a.Get()
	})

	if evals != 0 {
		t.Errorf("Computed should be lazy")
	}

	_ = c.Get()
	if evals != 1 {
		t.Errorf("Computed should evaluate on Get")
	}

	_ = c.Get()
	if evals != 1 {
		t.Errorf("Computed should cache value")
	}

	a.Set(2)
	if evals != 1 {
		t.Errorf("Computed should not re-evaluate immediately on dependency change")
	}

	_ = c.Get()
	if evals != 2 {
		t.Errorf("Computed should re-evaluate on Get after dirty")
	}
}

func TestPeek(t *testing.T) {
	a := New(1)
	runs := 0

	CreateEffect(func() {
		runs++
		_ = a.Peek()
	})

	if runs != 1 {
		t.Errorf("Initial run failed")
	}

	a.Set(2)
	if runs != 1 {
		t.Errorf("Peek should not create dependency")
	}
}
