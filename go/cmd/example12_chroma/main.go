package main

import (
	"basement/tui"
)

func main() {
	// Example 12: Syntax Highlighting (Optional)
	// Demonstrates code block rendering.
	//
	// To see syntax highlighting, run with:
	//   go run -tags chroma cmd/example12_chroma/main.go
	//
	// Without the tag, code blocks will be rendered as plain dimmed text.

	markdown := `
# Syntax Highlighting Demo

This example shows how code blocks are rendered.
If built with ` + "`-tags chroma`" + `, you should see colors below.

## Go Code

` + "```go" + `
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	// This is a comment
	x := 42
}
` + "```" + `

## Python Code

` + "```python" + `
def hello(name):
    print(f"Hello, {name}!")
    # List comprehension
    squares = [x**2 for x in range(10)]
    return squares
` + "```" + `

## JSON

` + "```json" + `
{
  "name": "BasementUI",
  "version": 1.0,
  "features": ["TUI", "Signals", "Layout"]
}
` + "```" + `

(Press 'q' or Ctrl+C to exit)
`

	app := func() tui.Renderable {
		return tui.Template(markdown)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	tui.Render(screen, app)

	// Handle Input
	quit := make(chan bool)
	screen.OnKey(func(ev tui.KeyEvent) {
		if ev.Rune == 'q' || (ev.Key == tui.KeyChar && ev.Mod == tui.ModCtrl && ev.Rune == 'c') {
			quit <- true
		}
	})
	<-quit
}
