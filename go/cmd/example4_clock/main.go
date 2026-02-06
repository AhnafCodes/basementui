package main

import (
	"basement/signals"
	"basement/tui"
	"time"
)

func main() {
	// Example 4: Digital Clock
	// Demonstrates a real-time update scenario.

	now := signals.New(time.Now().Format("15:04:05"))

	app := func() tui.Renderable {
		return tui.Template(`
# Digital Clock

The current time is:
#cyan(%v)

(Press 'q' or Ctrl+C to exit)
`, now)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			now.Set(time.Now().Format("15:04:05"))
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
