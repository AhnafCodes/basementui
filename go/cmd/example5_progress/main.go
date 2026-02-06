package main

import (
	"basement/signals"
	"basement/tui"
	"strings"
	"time"
)

func main() {
	// Example 5: Progress Bar
	// Shows how to build a custom component (progress bar) using Computed values.

	progress := signals.New(0)

	// Create a computed signal that returns the visual representation of the bar
	bar := signals.NewComputed(func() string {
		p := progress.Get()
		width := 20
		filled := int(float64(p) / 100.0 * float64(width))
		if filled > width {
			filled = width
		}
		empty := width - filled

		return "[" + strings.Repeat("#", filled) + strings.Repeat("-", empty) + "]"
	})

	app := func() tui.Renderable {
		return tui.Template(`
# Task Progress

Downloading...
%v  **%v%%**

(Press 'q' or Ctrl+C to exit)
`, bar, progress)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	go func() {
		for i := 0; i <= 100; i++ {
			time.Sleep(50 * time.Millisecond)
			progress.Set(i)
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
