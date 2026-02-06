package main

import (
	"basement/signals"
	"basement/tui"
	"fmt"
)

func main() {
	// Example 9: Interactive List
	// A navigable menu using Up/Down keys.
	// Demonstrates dynamic styling based on selection state.

	items := []string{
		"Option 1: Start Server",
		"Option 2: Deploy to Production",
		"Option 3: View Logs",
		"Option 4: Settings",
		"Option 5: Exit",
	}

	selectedIndex := signals.New(0)

	// Computed view for the list
	listView := signals.NewComputed(func() string {
		idx := selectedIndex.Get()
		var out string

		for i, item := range items {
			if i == idx {
				// Highlight selected item
				out += fmt.Sprintf("#green(> %s)\n", item)
			} else {
				out += fmt.Sprintf("  %s\n", item)
			}
		}
		return out
	})

	app := func() tui.Renderable {
		return tui.Template(`
# Main Menu

%v

(Use Up/Down to navigate, Enter to select, 'q' or Ctrl+C to quit)
`, listView)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	// Handle Input
	quit := make(chan bool)

	screen.OnKey(func(ev tui.KeyEvent) {
		idx := selectedIndex.Get()

		switch ev.Key {
		case tui.KeyArrowUp:
			if idx > 0 {
				selectedIndex.Set(idx - 1)
			}
		case tui.KeyArrowDown:
			if idx < len(items)-1 {
				selectedIndex.Set(idx + 1)
			}
		case tui.KeyEnter:
			// Action on select
			if idx == len(items)-1 { // Exit
				quit <- true
			}
		case tui.KeyChar:
			if ev.Rune == 'q' || (ev.Mod == tui.ModCtrl && ev.Rune == 'c') {
				quit <- true
			}
		}
	})

	<-quit
}
