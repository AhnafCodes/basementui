package tui

import (
	"basement/basement"
	"testing"
)

func TestBuffer(t *testing.T) {
	b := NewBuffer(10, 5)
	if len(b.Cells) != 50 {
		t.Errorf("Expected 50 cells, got %d", len(b.Cells))
	}

	b.Set(0, 0, 'a', basement.Style{Bold: true})
	cell := b.Get(0, 0)
	if cell.Char != 'a' || !cell.Style.Bold {
		t.Errorf("Set/Get failed")
	}
}

func TestBufferResize(t *testing.T) {
	b := NewBuffer(10, 10)
	b.Set(0, 0, 'x', basement.Style{})

	b.Resize(5, 5)
	if b.Width != 5 || b.Height != 5 {
		t.Errorf("Resize failed")
	}
	if b.Get(0, 0).Char != 'x' {
		t.Errorf("Resize should preserve content")
	}
}

func TestScreen(t *testing.T) {
	s := NewScreen()
	s.Clear()
	s.DrawText(0, 0, "Hello", basement.Style{Bold: true})

	cell := s.Back.Get(0, 0)
	if cell.Char != 'H' || !cell.Style.Bold {
		t.Errorf("DrawText failed")
	}
}
