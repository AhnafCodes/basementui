package main

import (
	"basement/tui"
)

func main() {
	// Example 1: Hello World
	// Shows the most basic usage: rendering static text with markdown styling.

	app := func() tui.Renderable {
		return tui.Template(`
# Hello, BasementUI!

This is a **static** example.
You can use *italics*, __underline__, and even #green(colors)!

(Press any key to exit)
`)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	// Wait for any key to exit
	quit := make(chan bool)
	screen.OnKey(func(ev tui.KeyEvent) {
		quit <- true
	})
	<-quit
}
