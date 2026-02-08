package tui

import (
	"bufio"
	"os"
	"time"
)

// StartInput starts the input loop and returns a channel of key events
func StartInput(done <-chan struct{}) <-chan KeyEvent {
	ch := make(chan KeyEvent)
	go inputLoop(ch, done)
	return ch
}

func inputLoop(ch chan<- KeyEvent, done <-chan struct{}) {
	reader := bufio.NewReader(os.Stdin)

	// Single goroutine reads raw bytes from stdin.
	// This is the ONLY goroutine that touches the reader,
	// eliminating data races on the bufio.Reader.
	rawCh := make(chan byte, 128)
	go func() {
		for {
			b, err := reader.ReadByte()
			if err != nil {
				close(rawCh)
				return
			}
			rawCh <- b
		}
	}()

	for {
		select {
		case <-done:
			close(ch)
			return
		case b, ok := <-rawCh:
			if !ok {
				close(ch)
				return
			}
			if b == 0x1b {
				processEsc(rawCh, ch)
			} else {
				processChar(b, ch)
			}
		}
	}
}

// processEsc handles ESC byte and potential escape sequences.
// Reads additional bytes from rawCh (not from the reader) to avoid races.
func processEsc(rawCh <-chan byte, ch chan<- KeyEvent) {
	// Wait a short time for follow-up bytes to distinguish bare ESC from sequences
	select {
	case next, ok := <-rawCh:
		if !ok {
			ch <- KeyEvent{Key: KeyEsc}
			return
		}
		if next == '[' {
			parseCSI(rawCh, ch)
		} else if next == 'O' {
			parseSS3(rawCh, ch)
		} else {
			// Alt + Key
			ch <- KeyEvent{Key: KeyChar, Rune: rune(next), Mod: ModAlt}
		}
	case <-time.After(10 * time.Millisecond):
		ch <- KeyEvent{Key: KeyEsc}
	}
}

// processChar handles a regular (non-ESC) byte
func processChar(b byte, ch chan<- KeyEvent) {
	if b <= 0x1f {
		// Ctrl + Key or special keys
		switch b {
		case 0x0d: // Enter
			ch <- KeyEvent{Key: KeyEnter}
		case 0x09: // Tab
			ch <- KeyEvent{Key: KeyTab}
		case 0x08: // Backspace (BS)
			ch <- KeyEvent{Key: KeyBackspace}
		case 0x03: // Ctrl+C
			ch <- KeyEvent{Key: KeyChar, Rune: 'c', Mod: ModCtrl}
		default:
			ch <- KeyEvent{Key: KeyChar, Rune: rune(b + 0x60), Mod: ModCtrl}
		}
	} else if b == 0x7f {
		ch <- KeyEvent{Key: KeyBackspace}
	} else {
		ch <- KeyEvent{Key: KeyChar, Rune: rune(b)}
	}
}

// readByteTimeout reads one byte from the channel with a timeout.
// Returns (0, false) if closed or timed out.
func readByteTimeout(rawCh <-chan byte, timeout time.Duration) (byte, bool) {
	select {
	case b, ok := <-rawCh:
		return b, ok
	case <-time.After(timeout):
		return 0, false
	}
}

// csiTimeout is the max time to wait for subsequent bytes within a CSI sequence.
const csiTimeout = 50 * time.Millisecond

func parseCSI(rawCh <-chan byte, ch chan<- KeyEvent) {
	// We consumed ESC [
	// Read all parameter bytes and the final byte.
	// CSI format: ESC [ <params> <final>
	//   params = digits and semicolons (0x30-0x3F)
	//   final  = 0x40-0x7E (letter or ~)
	var params []byte

	for {
		b, ok := readByteTimeout(rawCh, csiTimeout)
		if !ok {
			return
		}
		if b >= 0x40 && b <= 0x7E {
			// Final byte — interpret the sequence
			dispatchCSI(params, b, ch)
			return
		}
		// Parameter or intermediate byte — accumulate
		params = append(params, b)
	}
}

func dispatchCSI(params []byte, final byte, ch chan<- KeyEvent) {
	p := string(params)

	switch final {
	case 'A':
		ch <- KeyEvent{Key: KeyArrowUp}
	case 'B':
		ch <- KeyEvent{Key: KeyArrowDown}
	case 'C':
		ch <- KeyEvent{Key: KeyArrowRight}
	case 'D':
		ch <- KeyEvent{Key: KeyArrowLeft}
	case 'H':
		ch <- KeyEvent{Key: KeyHome}
	case 'F':
		ch <- KeyEvent{Key: KeyEnd}
	case '~':
		// Tilde-terminated: the first param encodes the key
		// Strip modifier after semicolon (e.g. "3;5" → "3")
		key := p
		if i := indexOf(p, ';'); i >= 0 {
			key = p[:i]
		}
		switch key {
		case "1":
			ch <- KeyEvent{Key: KeyHome}
		case "2":
			ch <- KeyEvent{Key: KeyInsert}
		case "3":
			ch <- KeyEvent{Key: KeyDelete}
		case "4":
			ch <- KeyEvent{Key: KeyEnd}
		case "5":
			ch <- KeyEvent{Key: KeyPgUp}
		case "6":
			ch <- KeyEvent{Key: KeyPgDown}
		case "15":
			ch <- KeyEvent{Key: KeyF5}
		case "17":
			ch <- KeyEvent{Key: KeyF6}
		case "18":
			ch <- KeyEvent{Key: KeyF7}
		case "19":
			ch <- KeyEvent{Key: KeyF8}
		case "20":
			ch <- KeyEvent{Key: KeyF9}
		case "21":
			ch <- KeyEvent{Key: KeyF10}
		case "23":
			ch <- KeyEvent{Key: KeyF11}
		case "24":
			ch <- KeyEvent{Key: KeyF12}
		}
	}
}

// indexOf returns the index of the first occurrence of sep in s, or -1.
func indexOf(s string, sep byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return i
		}
	}
	return -1
}

func parseSS3(rawCh <-chan byte, ch chan<- KeyEvent) {
	// We consumed ESC O
	b, ok := readByteTimeout(rawCh, csiTimeout)
	if !ok {
		return
	}
	switch b {
	// Arrow keys (Application Cursor Keys mode)
	case 'A':
		ch <- KeyEvent{Key: KeyArrowUp}
	case 'B':
		ch <- KeyEvent{Key: KeyArrowDown}
	case 'C':
		ch <- KeyEvent{Key: KeyArrowRight}
	case 'D':
		ch <- KeyEvent{Key: KeyArrowLeft}
	// Function keys
	case 'P':
		ch <- KeyEvent{Key: KeyF1}
	case 'Q':
		ch <- KeyEvent{Key: KeyF2}
	case 'R':
		ch <- KeyEvent{Key: KeyF3}
	case 'S':
		ch <- KeyEvent{Key: KeyF4}
	// Keypad keys (some terminals)
	case 'H':
		ch <- KeyEvent{Key: KeyHome}
	case 'F':
		ch <- KeyEvent{Key: KeyEnd}
	}
}