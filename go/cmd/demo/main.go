package main

import (
	"basement/signals"
	"basement/tui"
	"time"
)

func main() {
	// 1. Create State
	count := signals.New(0)

	// 2. Define View
	app := func() tui.Renderable {
		return tui.Template(`
# Counter App
Current count: **%v**

(Press 'q' or Ctrl+C to exit)
`, count)
	}

	// 3. Mount to Screen
	screen := tui.NewScreen()
	defer screen.Close()

	// Initial Render
	tui.Render(screen, app)

	// Simulate updates
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
