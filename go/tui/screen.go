package tui

import (
	"bufio"
	"basement/basement"
	"fmt"
	"os"
	"sync"

	"golang.org/x/term"
)

// Cell represents a single character on the screen
type Cell struct {
	Char  rune
	Style basement.Style
}

// Buffer represents a 2D grid of cells
type Buffer struct {
	Width  int
	Height int
	Cells  []Cell
}

// NewBuffer creates a new buffer of the given size
func NewBuffer(width, height int) *Buffer {
	return &Buffer{
		Width:  width,
		Height: height,
		Cells:  make([]Cell, width*height),
	}
}

// Set writes a rune and style to a specific coordinate
func (b *Buffer) Set(x, y int, ch rune, style basement.Style) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}
	b.Cells[y*b.Width+x] = Cell{Char: ch, Style: style}
}

// Get returns the cell at the given coordinate
func (b *Buffer) Get(x, y int) Cell {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return Cell{}
	}
	return b.Cells[y*b.Width+x]
}

// Resize resizes the buffer, preserving content where possible
func (b *Buffer) Resize(width, height int) {
	newCells := make([]Cell, width*height)

	minH := height
	if b.Height < minH {
		minH = b.Height
	}
	minW := width
	if b.Width < minW {
		minW = b.Width
	}

	for y := 0; y < minH; y++ {
		for x := 0; x < minW; x++ {
			newCells[y*width+x] = b.Cells[y*b.Width+x]
		}
	}

	b.Width = width
	b.Height = height
	b.Cells = newCells
}

// Screen manages the terminal display
type Screen struct {
	Front *Buffer // What is currently on screen
	Back  *Buffer // What we are drawing to
	mu    sync.Mutex
	out   *bufio.Writer

	// Input handling
	inputChan <-chan KeyEvent
	doneChan  chan struct{}
	oldState  *State

	// Scrolling
	ScrollY int
}

// NewScreen initializes a new screen
func NewScreen() *Screen {
	// Try to get actual terminal size
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w, h = 80, 24 // Fallback
	}

	s := &Screen{
		Front:    NewBuffer(w, h),
		Back:     NewBuffer(w, h),
		out:      bufio.NewWriter(os.Stdout),
		doneChan: make(chan struct{}),
	}

	// Enable raw mode
	oldState, err := enableRawMode(os.Stdin)
	if err == nil {
		s.oldState = oldState
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Failed to enable raw mode: %v\n", err)
	}

	// Start input loop
	s.inputChan = StartInput(s.doneChan)

	// Hide cursor initially
	s.out.WriteString("\x1b[?25l")
	s.out.Flush()

	return s
}

// Close restores the terminal state
func (s *Screen) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Signal input loop to stop
	close(s.doneChan)

	// Show cursor
	s.out.WriteString("\x1b[?25h")

	// Move cursor to bottom (simple heuristic)
	fmt.Fprintf(s.out, "\x1b[%dH", s.Back.Height+1)
	s.out.Flush()

	// Restore terminal mode
	if s.oldState != nil {
		disableRawMode(os.Stdin, s.oldState)
	}
}

// OnKey registers a callback for key events
func (s *Screen) OnKey(fn func(KeyEvent)) {
	go func() {
		for ev := range s.inputChan {
			fn(ev)
		}
	}()
}

// Clear clears the back buffer
func (s *Screen) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.Back.Cells {
		s.Back.Cells[i] = Cell{Char: ' '}
	}
}

// Render flushes the back buffer to the terminal
func (s *Screen) Render() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// We track the current cursor position to optimize movement
	curX, curY := -1, -1

	for y := 0; y < s.Back.Height; y++ {
		for x := 0; x < s.Back.Width; x++ {
			idx := y*s.Back.Width + x
			backCell := s.Back.Cells[idx]
			frontCell := s.Front.Cells[idx]

			// Diffing: Only draw if changed
			if backCell != frontCell {
				// Move cursor if not already there
				if curX != x || curY != y {
					// ANSI: Row;ColH (1-based)
					fmt.Fprintf(s.out, "\x1b[%d;%dH", y+1, x+1)
					curX, curY = x, y
				}

				// Apply styles
				s.writeStyle(backCell.Style)

				// Write char
				if backCell.Char == 0 {
					s.out.WriteRune(' ')
				} else {
					s.out.WriteRune(backCell.Char)
				}

				// Reset style
				s.out.WriteString("\x1b[0m")

				// Update cursor pos (it moved forward by 1)
				curX++

				// Update front buffer
				s.Front.Cells[idx] = backCell
			}
		}
	}

	s.out.Flush()
}

func (s *Screen) writeStyle(st basement.Style) {
	if st.Bold {
		s.out.WriteString("\x1b[1m")
	}
	if st.Dim {
		s.out.WriteString("\x1b[2m")
	}
	if st.Underline {
		s.out.WriteString("\x1b[4m")
	}
	if st.Reverse {
		s.out.WriteString("\x1b[7m")
	}
	if st.Blink {
		s.out.WriteString("\x1b[5m")
	}
	if st.Color != "" {
		s.out.WriteString(st.Color)
	}
	if st.BgColor != "" {
		s.out.WriteString(st.BgColor)
	}
}

// DrawText draws a string to the back buffer at x, y
func (s *Screen) DrawText(x, y int, text string, style basement.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()

	col := x
	for _, r := range text {
		if r == '\n' {
			y++
			col = x
			continue
		}
		s.Back.Set(col, y, r, style)
		col++
	}
}
