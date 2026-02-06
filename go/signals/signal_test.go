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
