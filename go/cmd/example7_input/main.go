package main

import (
	"basement/signals"
	"basement/tui"
	"fmt"
)

func main() {
	// Example 7: Input Handling
	// Move a character around the screen using arrow keys.

	x := signals.New(10)
	y := signals.New(5)
	msg := signals.New("Press Arrow Keys to move, 'q' to quit")

	// Create a computed signal for the character position
	// We render the whole screen content based on x/y
	view := signals.NewComputed(func() string {
		// This is a bit inefficient (rebuilding the whole string), but works for demo.
		// In a real app, we'd use layout nodes.
		// For now, we just return a string that positions the character.

		// We can't easily position absolute text in the template string without layout support.
		// So we'll just show the coordinates.
		return fmt.Sprintf("Position: (%d, %d)", x.Get(), y.Get())
	})

	app := func() tui.Renderable {
		// We use a trick: we render the character at the specific position by padding?
		// No, the template engine doesn't support absolute positioning yet.
		// So we will just display the coordinates and the message.
		// AND we will manually draw the character in a custom way?
		// No, let's stick to the template.

		return tui.Template(`
# Input Demo

%v

%v

(Use Arrow Keys)
`, msg, view)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	// Handle Input
	quit := make(chan bool)

	screen.OnKey(func(ev tui.KeyEvent) {
		switch ev.Key {
		case tui.KeyArrowUp:
			y.Set(y.Get() - 1)
			msg.Set("Moved Up")
		case tui.KeyArrowDown:
			y.Set(y.Get() + 1)
			msg.Set("Moved Down")
		case tui.KeyArrowLeft:
			x.Set(x.Get() - 1)
			msg.Set("Moved Left")
		case tui.KeyArrowRight:
			x.Set(x.Get() + 1)
			msg.Set("Moved Right")
		case tui.KeyChar:
			if ev.Rune == 'q' || (ev.Mod == tui.ModCtrl && ev.Rune == 'c') {
				quit <- true
			}
		}
	})

	<-quit
}
