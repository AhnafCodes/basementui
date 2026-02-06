//go:build !chroma

package tui

import "basement/basement"

// Highlight returns a list of styled spans for the given code and language.
// This default implementation returns a single span with Dim style.
func Highlight(code, lang string) []Span {
	return []Span{
		{Text: code, Style: basement.Style{Dim: true}},
	}
}
