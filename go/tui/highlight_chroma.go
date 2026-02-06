//go:build chroma

package tui

import (
	"basement/basement"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

// Highlight returns a list of styled spans for the given code and language using Chroma.
func Highlight(code, lang string) []Span {
	// 1. Get Lexer
	var lexer chroma.Lexer
	if lang != "" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// 2. Get Style (Monokai is a safe default for dark terminals)
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// 3. Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// Fallback on error
		return []Span{{Text: code, Style: basement.Style{Dim: true}}}
	}

	// 4. Map Tokens to Spans
	var spans []Span
	for _, token := range iterator.Tokens() {
		entry := style.Get(token.Type)

		// Map Chroma style to Basement style (ANSI 16 colors)
		// This is a simplified mapping.
		bs := basement.Style{}

		if entry.Bold == chroma.Yes {
			bs.Bold = true
		}
		if entry.Underline == chroma.Yes {
			bs.Underline = true
		}
		if entry.Italic == chroma.Yes {
			// Basement doesn't support Italic yet, maybe Dim?
			// bs.Dim = true
		}

		// Color Mapping
		// We need to map RGB to ANSI color names (black, red, green, etc.)
		// Since we don't have a full RGB->ANSI converter, we'll use heuristics based on token type
		// or try to approximate if Chroma gives us a color.

		// Better approach for TUI: Map Token Types directly to ANSI colors
		// instead of relying on the RGB values from the Chroma style.
		// This ensures it looks good in the terminal.

		switch token.Type.Category() {
		case chroma.Keyword:
			bs.Color = "\x1b[35m" // Magenta
			bs.Bold = true
		case chroma.Name:
			bs.Color = "\x1b[37m" // White
		case chroma.LiteralString:
			bs.Color = "\x1b[32m" // Green
		case chroma.LiteralNumber:
			bs.Color = "\x1b[36m" // Cyan
		case chroma.Comment:
			bs.Color = "\x1b[90m" // Grey (Bright Black)
			bs.Dim = true
		case chroma.Operator:
			bs.Color = "\x1b[37m" // White
		case chroma.Punctuation:
			bs.Color = "\x1b[37m" // White
		default:
			// Keep default
		}

		spans = append(spans, Span{Text: token.Value, Style: bs})
	}

	return spans
}
