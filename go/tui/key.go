package tui

// Key represents a special key or a character
type Key int

const (
	KeyNull Key = iota
	KeyEnter
	KeyBackspace
	KeyTab
	KeyEsc
	KeySpace

	// Cursor movement
	KeyArrowUp
	KeyArrowDown
	KeyArrowRight
	KeyArrowLeft

	// Navigation
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDown
	KeyDelete
	KeyInsert

	// Function keys
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// Char represents a regular rune key
	KeyChar
)

// Mod represents modifier keys (Ctrl, Alt, Shift)
type Mod int

const (
	ModNone  Mod = 0
	ModCtrl  Mod = 1 << 0
	ModAlt   Mod = 1 << 1
	ModShift Mod = 1 << 2
)

// KeyEvent represents a keyboard event
type KeyEvent struct {
	Key  Key
	Rune rune
	Mod  Mod
}
