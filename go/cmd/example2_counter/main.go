package main

import (
	"basement/signals"
	"basement/tui"
	"time"
)

func main() {
	// Example 2: Reactive Counter
	// Introduces Signals for state management.
	// The UI updates automatically when the signal changes.

	count := signals.New(0)

	app := func() tui.Renderable {
		return tui.Template(`
# Reactive Counter

Current count: **%v**

The value above updates automatically every second.

(Press 'q' or Ctrl+C to exit)
`, count)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	// Update state in a background goroutine
	go func() {
		for {
			time.Sleep(1 * time.Second)
			count.Set(count.Get() + 1)
		}
	}()

	// Wait for exit signal
	quit := make(chan bool)
	screen.OnKey(func(ev tui.KeyEvent) {
		if ev.Rune == 'q' || (ev.Key == tui.KeyChar && ev.Mod == tui.ModCtrl && ev.Rune == 'c') {
			quit <- true
		}
	})
	<-quit
}
