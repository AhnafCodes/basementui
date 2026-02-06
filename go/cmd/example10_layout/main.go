package main

import (
	"basement/signals"
	"basement/tui"
)

func main() {
	// Example 10: Responsive Layout
	// Demonstrates the Flexbox-like layout system.
	// We build a dashboard with a Sidebar (Fixed) and Main Content (Flex).

	menuItems := []string{"Dashboard", "Settings", "Logs", "Exit"}
	selectedIndex := signals.New(0)

	// Sidebar Component
	sidebar := signals.NewComputed(func() interface{} {
		idx := selectedIndex.Get()

		// Build menu items dynamically
		var menuNodes []interface{}
		menuNodes = append(menuNodes, tui.Box("MENU", false, 0))
		menuNodes = append(menuNodes, tui.Box("-------", false, 0))

		for i, item := range menuItems {
			label := item
			if i == idx {
				label = "> " + item
			}
			menuNodes = append(menuNodes, tui.Box(label, false, 0))
		}

		return tui.Box(
			tui.Col(menuNodes...),
			true, 1, // Border, Padding
		).WithWidth(tui.Fixed(20)).WithHeight(tui.Flex(1))
	})

	// Main Content Component
	content := signals.NewComputed(func() interface{} {
		idx := selectedIndex.Get()
		selectedItem := menuItems[idx]

		return tui.Box(
			tui.Col(
				tui.Box("# "+selectedItem, false, 0),
				tui.Box("Welcome to the admin panel.", false, 0),
				tui.Box("", false, 0),
				tui.Row(
					tui.Box("Stat 1: 100%", true, 1).WithWidth(tui.Flex(1)),
					tui.Box("Stat 2: OK", true, 1).WithWidth(tui.Flex(1)),
				),
			),
			true, 1,
		).WithWidth(tui.Flex(1)).WithHeight(tui.Flex(1))
	})

	// Root Layout
	layout := func() tui.Renderable {
		// We use a hole to inject the layout tree
		return tui.Template("%v",
			tui.Row(
				sidebar,
				content,
			).WithWidth(tui.Flex(1)).WithHeight(tui.Flex(1)),
		)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, layout)

	// Handle Input
	quit := make(chan bool)
	screen.OnKey(func(ev tui.KeyEvent) {
		if ev.Rune == 'q' || (ev.Key == tui.KeyChar && ev.Mod == tui.ModCtrl && ev.Rune == 'c') {
			quit <- true
		}

		idx := selectedIndex.Get()
		if ev.Key == tui.KeyArrowUp {
			if idx > 0 {
				selectedIndex.Set(idx - 1)
			}
		} else if ev.Key == tui.KeyArrowDown {
			if idx < len(menuItems)-1 {
				selectedIndex.Set(idx + 1)
			}
		} else if ev.Key == tui.KeyEnter {
			if menuItems[idx] == "Exit" {
				quit <- true
			}
		}
	})
	<-quit
}
