package main

import (
	"basement/signals"
	"basement/tui"
	"time"
)

func main() {
	// Example 6: Conditional Rendering
	// Demonstrates how to change the UI structure based on state.
	// We use a Computed signal to return different text/styles based on a condition.

	status := signals.New("loading")

	view := signals.NewComputed(func() string {
		s := status.Get()
		if s == "loading" {
			return "#yellow(Loading data...)"
		} else if s == "success" {
			return "#green(Data loaded successfully!)"
		} else {
			return "#red(Error loading data.)"
		}
	})

	app := func() tui.Renderable {
		return tui.Template(`
# Status Monitor

Status: %v

(Press 'q' or Ctrl+C to exit)
`, view)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	go func() {
		time.Sleep(2 * time.Second)
		status.Set("success")
		time.Sleep(2 * time.Second)
		status.Set("error")
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
