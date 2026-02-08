package main

import (
	"basement/signals"
	"basement/tui"
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
	// Optimization: Return a LayoutNode tree instead of a string to avoid re-parsing Markdown on every frame.
	listView := signals.NewComputed(func() interface{} {
		idx := selectedIndex.Get()

		var nodes []interface{}
		for i, item := range items {
			label := "  " + item
			if i == idx {
				label = "> " + item
			}

			// Create a Box for each item
			// We use manual styling via Box?
			// tui.Box doesn't support styling text directly yet (it takes string).
			// But we can use a string with markup inside Box?
			// If we use markup, it still gets parsed.

			// To be fastest, we should use raw strings if possible, or simple markup.
			// Parsing a small string "#green(...)" is faster than a big block.

			if i == idx {
				nodes = append(nodes, tui.Box("#green("+label+")", false, 0))
			} else {
				nodes = append(nodes, tui.Box(label, false, 0))
			}
		}

		return tui.Col(nodes...)
	})

	app := func() tui.Renderable {
		// We use a static template that just holds the list
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
