# BasementUI Developer Guide

Welcome to BasementUI! This guide provides a step-by-step introduction to building Terminal User Interfaces (TUIs) using our reactive, signal-based architecture.

## Table of Contents
1.  [Philosophy](#philosophy)
2.  [Quick Start](#quick-start)
3.  [Core Concepts](#core-concepts)
    *   [Signals & State](#signals--state)
    *   [Templates & Views](#templates--views)
    *   [Input Handling](#input-handling)
    *   [Layout System](#layout-system)
4.  [Advanced Topics](#advanced-topics)
    *   [Computed Values](#computed-values)
    *   [Custom Components](#custom-components)
    *   [Scrolling](#scrolling)
    *   [Syntax Highlighting](#syntax-highlighting)

---

## Philosophy

BasementUI is different from other TUI libraries (like Bubble Tea or tview). It is inspired by **SolidJS** and **Preact Signals**.

*   **No Virtual DOM**: We don't diff trees. We diff the screen buffer.
*   **Fine-Grained Reactivity**: When a variable changes, only the parts of the UI that use it are updated.
*   **Immediate Mode Feel**: You write your view as a function that returns a template, but it runs reactively.

---

## Quick Start

Every BasementUI app follows this pattern:
1.  **Setup Screen**: Initialize the terminal.
2.  **Define State**: Create signals.
3.  **Define View**: Create a function that returns a `Renderable`.
4.  **Render**: Mount the view to the screen.
5.  **Loop**: Handle input and block until exit.

### Minimal Example

```go
package main

import (
    "basement/signals"
    "basement/tui"
)

func main() {
    // 1. Setup
    screen := tui.NewScreen()
    defer screen.Close() // Important! Restores terminal on exit.

    // 2. State
    count := signals.New(0)

    // 3. View
    app := func() tui.Renderable {
        return tui.Template("Count: **%v**\n(Press 'q' to exit)", count)
    }

    // 4. Render
    tui.Render(screen, app)

    // 5. Input Loop
    quit := make(chan bool)
    screen.OnKey(func(ev tui.KeyEvent) {
        if ev.Rune == 'q' {
            quit <- true
        }
        if ev.Key == tui.KeyArrowUp {
            count.Set(count.Get() + 1)
        }
    })
    <-quit
}
```

---

## Core Concepts

### Signals & State

State is managed via `signals.New(value)`.
*   **Get()**: Read the value (and subscribe to updates).
*   **Set(value)**: Update the value (and notify subscribers).
*   **Peek()**: Read without subscribing.

**Example:** See `go/cmd/example2_counter/main.go`

```go
count := signals.New(0)
count.Set(count.Get() + 1)
```

### Templates & Views

Views are defined using Markdown-like syntax. You can use `**bold**`, `__underline__`, `~~strike~~`, and `#color(text)`.
Dynamic data is injected using `%v` placeholders (Holes).

**Example:** See `go/cmd/example6_conditional/main.go`

```go
status := signals.New("Loading...")
tui.Template("Status: #yellow(%v)", status)
```

### Input Handling

BasementUI puts the terminal in **Raw Mode**. This means:
1.  You must handle `Ctrl+C` manually if you want it to quit.
2.  You receive key events immediately (no Enter needed).

Use `screen.OnKey` to register a handler.

**Example:** See `go/cmd/example7_input/main.go`

```go
screen.OnKey(func(ev tui.KeyEvent) {
    if ev.Key == tui.KeyArrowLeft {
        // Move left
    }
    if ev.Rune == 'q' || (ev.Key == tui.KeyChar && ev.Mod == tui.ModCtrl && ev.Rune == 'c') {
        // Quit
    }
})
```

### Layout System

For complex UIs, use the Flexbox-like layout engine instead of raw strings.
*   **Row**: Horizontal stack.
*   **Col**: Vertical stack.
*   **Box**: Container with Border and Padding.
*   **Size**: `Fixed(n)`, `Flex(n)`, `Auto()`.

**Example:** See `go/cmd/example10_layout/main.go`

```go
sidebar := tui.Box("Menu", true, 1).WithWidth(tui.Fixed(20))
content := tui.Box("Content", true, 1).WithWidth(tui.Flex(1))

layout := tui.Row(sidebar, content)
```

---

## Advanced Topics

### Computed Values

Use `signals.NewComputed` to derive state from other signals. The computation is lazy and only runs when dependencies change.

**Example:** See `go/cmd/example3_computed/main.go`

```go
count := signals.New(1)
double := signals.NewComputed(func() int {
    return count.Get() * 2
})
```

### Custom Components

You can build reusable components by returning `*LayoutNode` or `Renderable`.

**Example:** See `go/cmd/example5_progress/main.go`

```go
func ProgressBar(progress *signals.Signal[int]) *signals.Computed[string] {
    return signals.NewComputed(func() string {
        p := progress.Get()
        // ... generate bar string ...
        return "[" + bar + "]"
    })
}
```

### Scrolling

To handle content larger than the screen, bind a signal to `screen.ScrollY`.

**Example:** See `go/cmd/example11_markdown/main.go`

```go
scrollY := signals.New(0)

wrappedApp := func() tui.Renderable {
    screen.ScrollY = scrollY.Get() // Sync state
    return app()
}
```

### Syntax Highlighting

BasementUI supports syntax highlighting via [Chroma](https://github.com/alecthomas/chroma). This is an optional dependency.

To enable it, build with `-tags chroma`.

**Example:** See `go/cmd/example12_chroma/main.go`

```bash
go run -tags chroma cmd/example12_chroma/main.go
```

---

## Troubleshooting

*   **Cursor is gone?** If your app crashes, the cursor might remain hidden. Run `reset` in your terminal.
*   **Input not working?** Ensure you are handling the correct `KeyEvent`. Debug by printing `ev.Key` and `ev.Rune` to a log file.
*   **Layout looks wrong?** Check if you are mixing `Auto` and `Flex` correctly. `Auto` takes the size of its content; `Flex` takes remaining space.
