# BasementUI: Design Specification

## 1. Overview
**BasementUI** is a modern Terminal User Interface (TUI) library for Go, heavily inspired by **uhtml** and **Preact Signals**.

It avoids the complexity and overhead of a Virtual DOM. Instead, it uses a **Tagged Template** approach (adapted for Go) to create reactive UI nodes. It parses templates once, identifies "Holes" (dynamic parts), and binds them directly to **Signals**. When a signal changes, only the specific text or style associated with that signal is updated in the terminal buffer.

## 2. Core Concepts

### 2.1 Signals (The State)
State is managed via Signals.
*   `Signal[T]`: Holds a value.
*   `Computed[T]`: Derived value.
*   `Effect`: Side effect triggered by signal changes.

### 2.2 The "Tag" Function (`tui.Template`)
In Go: `tui.Template("Hello %v", name)`

This function:
1.  Parses the markdown string into an AST.
2.  Identifies placeholders (`%v`).
3.  Maps the arguments (`name`) to those placeholders.
4.  Returns a `Renderable` object containing the AST and arguments.

### 2.3 Rendering
`tui.Render(screen, func() tui.Renderable)`
This function mounts the reactive node to the screen. It executes the function inside an Effect, setting up the initial render and activating signal subscriptions.

## 3. Module Specifications

### 3.1 Module: `signals` (✅ Implemented)
*   `Signal[T]`: Reactive value container with `Get()`, `Set()`, and `Peek()`.
*   `Effect`: Automatic dependency tracking via global `activeEffect`.
*   `Computed[T]`: Derived reactive values.

### 3.2 Module: `tui` (✅ Implemented)
The "Browser Engine" for the terminal.
*   `Cell`: Represents a single grid unit `struct { Char rune; Style basement.Style }`.
*   `Buffer`: A 2D grid of cells.
*   `Screen`: Manages `Front` and `Back` buffers and handles rendering to stdout with **Diffing**.
*   **Input Handling**:
    *   Uses `golang.org/x/term` for cross-platform Raw Mode.
    *   Parses ANSI escape sequences for special keys (Arrows, Home/End, F-keys).
    *   Exposes `OnKey(callback)` API for handling events.
*   **Layout Engine**:
    *   **Flexbox-like Model**: Supports `Row` and `Column` layouts.
    *   **Sizing**: `Fixed(n)`, `Flex(n)`, and `Auto` sizing strategies.
    *   **Box Model**: Supports Borders and Padding.
    *   **Two-Pass Rendering**: `Measure` pass calculates geometry, `Draw` pass renders content.
*   **Scrolling**:
    *   `Screen` supports `ScrollY` offset.
    *   Rendering pipeline applies offset to all nodes.
*   **Syntax Highlighting**:
    *   Optional integration with `chroma` via build tags (`-tags chroma`).
    *   Maps Chroma tokens to ANSI-16 colors for consistent TUI styling.

### 3.3 Module: `basement` (Template Engine) (✅ Implemented)
Parses the markdown string into an AST.
*   `Node`: Represents Text, Style, Block, Hole, List, CodeBlock, HR, or Quote.
*   `Parser`: Regex-based parser supporting Headers, Bold, Underline, Colors, Lists, Code Fences, Blockquotes, and Horizontal Rules.
*   `Style`: Shared style definition.

## 4. Roadmap

1.  **Phase 1: Signals** (✅ Complete)
2.  **Phase 2: DOM/Screen** (✅ Complete)
3.  **Phase 3: Template Parser** (✅ Complete)
4.  **Phase 4: Wiring** (✅ Complete)
5.  **Phase 5: Input & Layout** (✅ Complete)
    *   **Input Handling**:
        *   Raw Mode support via `x/term`.
        *   ANSI Escape Sequence parser.
        *   `KeyEvent` struct and `OnKey` callback.
    *   **Layout**:
        *   Implemented `LayoutNode` with `Measure` and `Draw`.
        *   Added `Row`, `Col`, `Box` primitives.
        *   Integrated with Template engine via Holes.
    *   **Scrolling**:
        *   Implemented vertical scrolling support in `Screen` and `Render`.
    *   **Markdown Extensions**:
        *   Added support for Lists, Code Blocks, Blockquotes, and HRs.
    *   **Syntax Highlighting**:
        *   Added optional Chroma support for code blocks.
