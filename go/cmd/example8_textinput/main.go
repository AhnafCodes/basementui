package main

import (
	"basement/signals"
	"basement/tui"
)

func main() {
	// Example 8: Text Input
	// A simple text field implementation.
	// Demonstrates combining input handling with reactive state to build controls.

	input := signals.New("")

	// Computed view for the input field
	field := signals.NewComputed(func() string {
		txt := input.Get()

		// Add a static cursor at the end
		cursorChar := "#white(! !)" // Reverse space (block cursor)

		return "#blue(> )" + txt + cursorChar
	})

	app := func() tui.Renderable {
		return tui.Template(`
# Text Input Demo

Type something below:

%v

(Press 'Esc' or Ctrl+C to quit)
`, field)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	// Handle Input
	quit := make(chan bool)

	screen.OnKey(func(ev tui.KeyEvent) {
		current := input.Get()

		switch ev.Key {
		case tui.KeyChar:
			// Handle Ctrl+C
			if ev.Mod == tui.ModCtrl && ev.Rune == 'c' {
				quit <- true
				return
			}
			input.Set(current + string(ev.Rune))
		case tui.KeySpace:
			input.Set(current + " ")
		case tui.KeyBackspace:
			// Note: This is a simple byte-slicing backspace.
			// For full unicode support, we should convert to []rune, slice, then back to string.
			if len(current) > 0 {
				input.Set(current[:len(current)-1])
			}
		case tui.KeyEsc:
			quit <- true
		}
	})

	<-quit
}
