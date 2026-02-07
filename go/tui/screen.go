package tui

import (
	"bufio"
	"basement/basement"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

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
		copy(newCells[y*width:y*width+minW], b.Cells[y*b.Width:y*b.Width+minW])
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

	// Capabilities
	supportsItalic bool
	supportsStrike bool

	// Resize handling
	resizeCh chan os.Signal
	OnResize func(w, h int)

	// Pre-allocated blank row for fast clear
	blankRow []Cell

	// Reusable buffer for cursor positioning escape sequences
	posBuf []byte
}

// NewScreen initializes a new screen
func NewScreen() *Screen {
	// Try to get actual terminal size
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w, h = 80, 24 // Fallback
	}

	// Pre-allocate blank row for fast clear
	blankRow := make([]Cell, w)
	for i := range blankRow {
		blankRow[i] = Cell{Char: ' '}
	}

	s := &Screen{
		Front:    NewBuffer(w, h),
		Back:     NewBuffer(w, h),
		out:      bufio.NewWriterSize(os.Stdout, 64*1024), // 64KB write buffer
		doneChan: make(chan struct{}),
		blankRow: blankRow,
		posBuf:   make([]byte, 0, 32),
	}

	// Check for capabilities
	termEnv := os.Getenv("TERM")
	if strings.Contains(termEnv, "xterm") ||
	   strings.Contains(termEnv, "truecolor") ||
	   strings.Contains(termEnv, "alacritty") ||
	   strings.Contains(termEnv, "kitty") ||
	   strings.Contains(termEnv, "screen") ||
	   strings.Contains(termEnv, "tmux") {
		s.supportsItalic = true
		s.supportsStrike = true // Most modern terms support both
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

	// Start SIGWINCH listener for terminal resize
	s.resizeCh = make(chan os.Signal, 1)
	signal.Notify(s.resizeCh, syscall.SIGWINCH)
	go s.handleResize()

	// Hide cursor initially
	s.out.WriteString("\x1b[?25l")
	s.out.Flush()

	return s
}

// Close restores the terminal state
func (s *Screen) Close() {
	// Stop resize signal before acquiring lock
	signal.Stop(s.resizeCh)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Signal input loop and resize handler to stop
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

// handleResize listens for SIGWINCH and resizes buffers
func (s *Screen) handleResize() {
	for {
		select {
		case <-s.doneChan:
			return
		case <-s.resizeCh:
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				continue
			}
			s.mu.Lock()
			s.Front.Resize(w, h)
			s.Back.Resize(w, h)
			// Update blank row for new width
			s.blankRow = make([]Cell, w)
			for i := range s.blankRow {
				s.blankRow[i] = Cell{Char: ' '}
			}
			// Force full redraw by invalidating front buffer
			for i := range s.Front.Cells {
				s.Front.Cells[i] = Cell{}
			}
			s.mu.Unlock()
			if s.OnResize != nil {
				s.OnResize(w, h)
			}
		}
	}
}

// Clear clears the back buffer
func (s *Screen) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearBackBuf()
}

// clearBackBuf clears the back buffer without locking (for internal use)
func (s *Screen) clearBackBuf() {
	w := s.Back.Width
	h := s.Back.Height
	cells := s.Back.Cells
	if len(s.blankRow) != w {
		s.blankRow = make([]Cell, w)
		for i := range s.blankRow {
			s.blankRow[i] = Cell{Char: ' '}
		}
	}
	for y := 0; y < h; y++ {
		copy(cells[y*w:(y+1)*w], s.blankRow)
	}
}

// Render flushes the back buffer to the terminal
func (s *Screen) Render() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.renderUnlocked()
}

// Frame executes draw under a single lock: clear, draw, diff+flush.
// Use drawTextUnlocked inside the draw callback.
func (s *Screen) Frame(draw func()) {
	s.mu.Lock()

	// Clear
	s.clearBackBuf()

	// Draw to back buffer
	draw()

	// Diff and flush
	s.renderUnlocked()

	s.mu.Unlock()
}

func (s *Screen) renderUnlocked() {
	w := s.Back.Width
	h := s.Back.Height
	backCells := s.Back.Cells
	frontCells := s.Front.Cells

	curX, curY := -1, -1
	var lastStyle basement.Style
	styleActive := false

	for y := 0; y < h; y++ {
		rowOff := y * w
		for x := 0; x < w; x++ {
			idx := rowOff + x
			backCell := backCells[idx]

			if backCell != frontCells[idx] {
				// Move cursor if needed
				if curX != x || curY != y {
					s.writeCursorPos(y+1, x+1)
					curX, curY = x, y
				}

				// Only emit style escapes when style changes
				if !styleActive || backCell.Style != lastStyle {
					if styleActive {
						s.out.WriteString("\x1b[0m")
					}
					s.writeStyle(backCell.Style)
					lastStyle = backCell.Style
					styleActive = true
				}

				ch := backCell.Char
				if ch == 0 {
					ch = ' '
				}
				s.out.WriteRune(ch)
				curX++

				frontCells[idx] = backCell
			}
		}
	}

	// Reset style once at end
	if styleActive {
		s.out.WriteString("\x1b[0m")
	}

	s.out.Flush()
}

// writeCursorPos writes ANSI cursor position without fmt.Fprintf overhead
func (s *Screen) writeCursorPos(row, col int) {
	s.posBuf = s.posBuf[:0]
	s.posBuf = append(s.posBuf, '\x1b', '[')
	s.posBuf = strconv.AppendInt(s.posBuf, int64(row), 10)
	s.posBuf = append(s.posBuf, ';')
	s.posBuf = strconv.AppendInt(s.posBuf, int64(col), 10)
	s.posBuf = append(s.posBuf, 'H')
	s.out.Write(s.posBuf)
}

func (s *Screen) writeStyle(st basement.Style) {
	if st.Bold {
		s.out.WriteString("\x1b[1m")
	}
	if st.Dim {
		s.out.WriteString("\x1b[2m")
	}
	if st.Italic {
		if s.supportsItalic {
			s.out.WriteString("\x1b[3m")
		} else {
			s.out.WriteString("\x1b[2m") // Fallback to Dim
		}
	}
	if st.Underline {
		s.out.WriteString("\x1b[4m")
	}
	if st.Strike {
		if s.supportsStrike {
			s.out.WriteString("\x1b[9m")
		}
		// No fallback for strike
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
	s.drawTextUnlocked(x, y, text, style)
}

// drawTextUnlocked is the lock-free version for use within Frame()
func (s *Screen) drawTextUnlocked(x, y int, text string, style basement.Style) {
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
