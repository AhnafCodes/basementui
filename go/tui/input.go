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

	// We need a way to interrupt the blocking read.
	// Since we can't easily interrupt ReadByte without closing Stdin (which is bad),
	// we will rely on the fact that when the program exits or Close() is called,
	// we might just let this goroutine die with the process or handle it if we can.
	// However, for a library, it's better to be clean.
	// One common trick is to use a non-blocking read with a select, but that requires syscalls.
	// Given the constraints, we'll stick to the blocking read but check 'done' where possible.
	// A better approach for production would be to use a separate goroutine for reading that sends to an internal channel,
	// and then select on that internal channel and 'done'.

	inputData := make(chan byte)
	go func() {
		for {
			b, err := reader.ReadByte()
			if err != nil {
				close(inputData)
				return
			}
			inputData <- b
		}
	}()

	for {
		select {
		case <-done:
			close(ch)
			return
		case b, ok := <-inputData:
			if !ok {
				close(ch)
				return
			}
			processByte(b, reader, ch)
		}
	}
}

func processByte(b byte, reader *bufio.Reader, ch chan<- KeyEvent) {
	if b == 0x1b { // ESC
		// Check if there are more bytes available immediately
		if reader.Buffered() == 0 {
			// Wait a tiny bit to see if it's an escape sequence
			time.Sleep(10 * time.Millisecond)
			if reader.Buffered() == 0 {
				ch <- KeyEvent{Key: KeyEsc}
				return
			}
		}

		// Peek next byte
		next, _ := reader.Peek(1)
		if next[0] == '[' {
			reader.ReadByte() // Consume '['
			parseCSI(reader, ch)
		} else if next[0] == 'O' { // SS3 (e.g. F1-F4)
			reader.ReadByte() // Consume 'O'
			parseSS3(reader, ch)
		} else {
			// Alt + Key
			nextByte, _ := reader.ReadByte()
			ch <- KeyEvent{Key: KeyChar, Rune: rune(nextByte), Mod: ModAlt}
		}
	} else {
		// Regular character or Ctrl key
		if b <= 0x1f {
			// Ctrl + Key
			// 0x01 = Ctrl+A, ..., 0x1a = Ctrl+Z
			// Special handling for Enter, Tab, Backspace
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
}

func parseCSI(reader *bufio.Reader, ch chan<- KeyEvent) {
	// We consumed ESC [
	// Now reading parameter bytes (0-9:;<=>?)
	// Then intermediate bytes (!"#$%&'()*+,-./)
	// Then final byte (@A-Z[\]^_`a-z{|}~)

	// Simplified parser for common keys
	b, _ := reader.ReadByte()

	switch b {
	case 'A': ch <- KeyEvent{Key: KeyArrowUp}
	case 'B': ch <- KeyEvent{Key: KeyArrowDown}
	case 'C': ch <- KeyEvent{Key: KeyArrowRight}
	case 'D': ch <- KeyEvent{Key: KeyArrowLeft}
	case 'H': ch <- KeyEvent{Key: KeyHome}
	case 'F': ch <- KeyEvent{Key: KeyEnd}
	case '1', '2', '3', '4', '5', '6':
		// Extended keys: Home, Insert, Delete, End, PgUp, PgDown
		// They usually end with ~
		// e.g. 1~ (Home), 2~ (Insert), 3~ (Delete), 4~ (End), 5~ (PgUp), 6~ (PgDown)
		// Also F5-F12 are 15~, 17~...

		// We need to read until ~
		param := []byte{b}
		for {
			next, err := reader.Peek(1)
			if err != nil || next[0] < '0' || next[0] > '9' {
				break
			}
			// It's a digit
			d, _ := reader.ReadByte()
			param = append(param, d)
		}

		// Consume ~ if present
		if next, _ := reader.Peek(1); next[0] == '~' {
			reader.ReadByte()
		}

		s := string(param)
		switch s {
		case "1": ch <- KeyEvent{Key: KeyHome}
		case "2": ch <- KeyEvent{Key: KeyInsert}
		case "3": ch <- KeyEvent{Key: KeyDelete}
		case "4": ch <- KeyEvent{Key: KeyEnd}
		case "5": ch <- KeyEvent{Key: KeyPgUp}
		case "6": ch <- KeyEvent{Key: KeyPgDown}
		case "15": ch <- KeyEvent{Key: KeyF5}
		case "17": ch <- KeyEvent{Key: KeyF6}
		case "18": ch <- KeyEvent{Key: KeyF7}
		case "19": ch <- KeyEvent{Key: KeyF8}
		case "20": ch <- KeyEvent{Key: KeyF9}
		case "21": ch <- KeyEvent{Key: KeyF10}
		case "23": ch <- KeyEvent{Key: KeyF11}
		case "24": ch <- KeyEvent{Key: KeyF12}
		}
	}
}

func parseSS3(reader *bufio.Reader, ch chan<- KeyEvent) {
	// We consumed ESC O
	b, _ := reader.ReadByte()
	switch b {
	case 'P': ch <- KeyEvent{Key: KeyF1}
	case 'Q': ch <- KeyEvent{Key: KeyF2}
	case 'R': ch <- KeyEvent{Key: KeyF3}
	case 'S': ch <- KeyEvent{Key: KeyF4}
	}
}
