package main

import (
	"basement/signals"
	"basement/tui"
	"time"
)

func main() {
	// Example 3: Computed Values
	// Shows how to derive state from other signals.
	// 'double' automatically updates whenever 'count' changes.

	count := signals.New(0)
	double := signals.NewComputed(func() int {
		return count.Get() * 2
	})

	app := func() tui.Renderable {
		return tui.Template(`
# Computed Values

Count:  %v
Double: %v

(Double is derived from Count)

(Press 'q' or Ctrl+C to exit)
`, count, double)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

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
