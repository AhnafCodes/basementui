package basement

// Style represents the visual style of a cell
type Style struct {
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
	Strike    bool // New field
	Reverse   bool
	Blink     bool
	Color     string // ANSI color code
	BgColor   string // ANSI background color code
}

// GetColorCode returns the ANSI escape code for a given color name
func GetColorCode(name string) string {
	switch name {
	case "black":   return "\x1b[30m"
	case "red":     return "\x1b[31m"
	case "green":   return "\x1b[32m"
	case "blue":    return "\x1b[34m"
	case "magenta": return "\x1b[35m"
	case "cyan":    return "\x1b[36m"
	case "white":   return "\x1b[37m"
	case "yellow":  return "\x1b[33m"
	case "grey":    return "\x1b[90m"
	default:        return ""
	}
}
